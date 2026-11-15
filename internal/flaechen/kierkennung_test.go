// **Die Kartenerkennung über jede Fläche ohne Fenster** — Werkstück A13.3 von
// Phase 5.3
// (#1452).
//
// # Was hier gemessen wird
//
// **Dass eine Datei, die es nicht gibt, überall dieselbe fachliche Absage
// bekommt — und dass sie sie bekommt, bevor irgendetwas geschieht.** Das
// Lesen der Datei ist die billigste Prüfung des ganzen Wegs; wer sie hinter
// den Anbieter legte, buchte einen Kostenposten für ein Bild, das niemand
// finden kann.
//
// **Und dass eine Datei, die kein Bild ist, nicht hinausgeht.** Die Klasse
// ist eine andere: die Datei liegt da, nur ist sie keine, die ein Anbieter
// entgegennähme — und ihn das herausfinden zu lassen hieße, dafür zu zahlen.
//
// # Und was er nicht deckt
//
// **Eine Erkennung, die wirklich entsteht.** Sie braucht ein Modell mit
// belegter Sichtfähigkeit **und** eine Drucksuche bei Scryfall; beides steht
// hinter `netzlauf` und nicht in der Pipeline (dieselbe Grenze wie bei
// `kieinschaetzung_test.go`). Was hier steht, ist die **Gleichheit der
// Flächen** — die Zusage aus §16.
package flaechen_test

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Eine Bilddatei, die es nicht gibt, wird an jeder Fläche gleich
// abgewiesen.**
func TestEinFehlendesKartenfotoUeberJedeFlaeche(t *testing.T) {
	umgebung := append(eigenerBestand(t), "HOME="+t.TempDir())

	an := starteDienst(t, umgebung)

	fehlt := filepath.Join(t.TempDir(), "gibtsnicht.png")

	abweisungGleich(t, "ein fehlendes Kartenfoto", abweisung{
		klasse: klasseNichtGefunden,
		kern:   "gibtsnicht.png",
		fenster: func() *app.Fehler {
			return fensterAus(t, umgebung).KIKarteErkennen(t.Context(), app.KIErkennungsbezug{
				Pfad: fehlt, Anbieter: "ollama", Alias: "gemma4:8b",
			}).Fehler
		},
		cli: func() (string, int) {
			aus, fehler, code := kommandozeile(t, umgebung,
				"ki-erkennung", fehlt, "ollama", "gemma4:8b")

			return aus + fehler, code
		},
		mcp: func() (string, bool) {
			return mcpRuf(t, umgebung, "schreiben.ki.erkennung",
				map[string]any{
					"pfad": fehlt, "anbieter": "ollama", "alias": "gemma4:8b",
				})
		},
		rest: func() (int, string) {
			return an.ruf(t, http.MethodPost, "/ki/erkennung",
				`{"pfad":`+alsJSONText(t, fehlt)+
					`,"anbieter":"ollama","alias":"gemma4:8b"}`)
		},
	})
}

// **Und eine Datei, die kein Bild ist, ebenso.**
//
// **Eine andere Klasse und deshalb ein eigener Lauf**: hier liegt die Datei
// da. Wer nur den ersten Fall prüfte, ließe offen, ob der zweite überhaupt
// geprüft wird — oder ob ihn der Anbieter abweist, nachdem er abgerechnet
// hat.
func TestEineDateiOhneBildUeberJedeFlaeche(t *testing.T) {
	umgebung := append(eigenerBestand(t), "HOME="+t.TempDir())

	an := starteDienst(t, umgebung)

	keinBild := filepath.Join(t.TempDir(), "notiz.png")
	if fehler := os.WriteFile(keinBild, []byte("kein Bild"), 0o600); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	abweisungGleich(t, "eine Datei, die kein Bild ist", abweisung{
		klasse: klasseValidierung,
		kern:   "notiz.png",
		fenster: func() *app.Fehler {
			return fensterAus(t, umgebung).KIKarteErkennen(t.Context(), app.KIErkennungsbezug{
				Pfad: keinBild, Anbieter: "ollama", Alias: "gemma4:8b",
			}).Fehler
		},
		cli: func() (string, int) {
			aus, fehler, code := kommandozeile(t, umgebung,
				"ki-erkennung", keinBild, "ollama", "gemma4:8b")

			return aus + fehler, code
		},
		mcp: func() (string, bool) {
			return mcpRuf(t, umgebung, "schreiben.ki.erkennung",
				map[string]any{
					"pfad": keinBild, "anbieter": "ollama", "alias": "gemma4:8b",
				})
		},
		rest: func() (int, string) {
			return an.ruf(t, http.MethodPost, "/ki/erkennung",
				`{"pfad":`+alsJSONText(t, keinBild)+
					`,"anbieter":"ollama","alias":"gemma4:8b"}`)
		},
	})
}
