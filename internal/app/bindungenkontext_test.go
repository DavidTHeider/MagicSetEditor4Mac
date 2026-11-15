// Der Kontext der Bindungen.
//
// **Abgetrennt von `bindungen_test.go`**, weil die Datei über die Zeilengrenze
// aus AGENTS.md §2 gewachsen war. Der Schnitt folgt der Frage: dort steht, dass
// **jede** Bindung ihre Anwendung erreicht, hier, **womit** sie sie ruft.
package app_test

import (
	"context"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

func TestDerKontextDerBindungen(t *testing.T) {
	probe := &suchprobe{}
	anwendung := app.NeueApp("")
	anwendung.MitSucher(probe)

	bindungen := app.NeueBindungen(anwendung)

	// Ohne Fenster: der Aufruf geht trotzdem durch.
	if antwort := bindungen.Suchfelder(t.Context(), ""); antwort.Fehler != nil {
		t.Errorf("ohne Fenster: %+v", antwort.Fehler)
	}

	anwendung.Bereit(context.WithValue(probekontext(t), fensterschluessel{}, "da"))

	if antwort := bindungen.Suchfelder(t.Context(), ""); antwort.Fehler != nil {
		t.Errorf("mit Fenster: %+v", antwort.Fehler)
	}

	if probe.kontext == nil || probe.kontext.Value(fensterschluessel{}) != "da" {
		t.Errorf("der Kontext des Fensters kam nicht an: %v", probe.kontext)
	}
}

type fensterschluessel struct{}

// standprobe gibt einen festen Datenstand.
type standprobe struct{}

func (standprobe) Datenstand(context.Context) (int64, error) { return 42, nil }

// **Je Ruf eine eigene Vorgangskennung** — und nicht die des Fensters
// (BK-01, #2109). `MitNeuemVorgang` würfelt nur, wenn noch keine da ist; am
// Kontext aus [app.App.Bereit] war immer schon eine, und so trug bis dahin
// jeder Ruf des Fensters dieselbe.
func TestJederRufDesFenstersHatSeineEigeneVorgangskennung(t *testing.T) {
	probe := &suchprobe{}
	anwendung := app.NeueApp("")
	anwendung.MitSucher(probe)
	anwendung.Bereit(probekontext(t))

	bindungen := app.NeueBindungen(anwendung)
	kennungen := map[string]bool{}

	for range 3 {
		if antwort := bindungen.Suchfelder(t.Context(), ""); antwort.Fehler != nil {
			t.Fatalf("Suchfelder: %+v", antwort.Fehler)
		}

		kennung := string(app.VorgangAus(probe.kontext))
		if kennung == "" {
			t.Fatal("der Ruf trug keine Vorgangskennung")
		}

		kennungen[kennung] = true
	}

	if len(kennungen) != 3 {
		t.Errorf("drei Rufe, %d Kennungen: %v", len(kennungen), kennungen)
	}

	// **Eine mitgegebene Kennung bleibt** — wer sie von außen setzt, will
	// sie wiederfinden.
	bindungen.Suchfelder(app.MitVorgang(t.Context(), "abcd1234"), "")

	if kennung := app.VorgangAus(probe.kontext); kennung != "abcd1234" {
		t.Errorf("mitgegeben war abcd1234, angekommen ist %q", kennung)
	}
}
