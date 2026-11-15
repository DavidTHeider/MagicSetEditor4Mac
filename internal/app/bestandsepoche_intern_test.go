package app

// **Eine Meldung trägt die Epoche des Bestands, in dem sie entstand** —
// Befund 1 des nachgeholten Zweitblicks über die Folgepakete (J05, #2269, zu
// #2292).
//
// Die Sperre aus [App.DateiOeffnen] ordnet Doppelklick und Tausch **in Go**.
// Im Fenster ordnet sie nichts: Wails stellt ein Ereignis in eine Schlange,
// die für sich abgearbeitet wird (`application/events.go`, Zeile 171,
// v3.0.0-beta.27), und die Antwort einer Bindung kann es überholen. Kam die
// Kennung aus dem alten Bestand nach der Antwort des Neuanfangs an, schlug
// das Fenster sie im neuen auf. **Die Epoche macht sie erkennbar**: jeder
// Tausch zählt sie hoch, die Meldung trägt ihre, die Antwort des Tauschs die
// neue — und das Fenster verwirft, was älter ist.

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"
)

func TestEinTauschZaehltDieEpocheHoch(t *testing.T) {
	a := gestartet(t)

	vorher := a.lauf.epoche.Load()

	if fehler := a.imTausch(func() error { return nil }); fehler != nil {
		t.Fatalf("Tausch: %v", fehler)
	}

	if nachher := a.lauf.epoche.Load(); nachher != vorher+1 {
		t.Errorf("Epoche %d nach dem Tausch, vorher %d", nachher, vorher)
	}

	// **Auch ein gescheiterter Tausch zählt** — er hat den Bestand
	// geschlossen und wieder geöffnet, und was vorher galt, ist nicht mehr
	// sicher dasselbe.
	_ = a.imTausch(func() error { return errors.New("gescheitert") }) //nolint:errcheck // gefragt ist die Epoche

	if nachher := a.lauf.epoche.Load(); nachher != vorher+2 {
		t.Errorf("Epoche %d nach dem gescheiterten Tausch, erwartet %d", nachher, vorher+2)
	}
}

func TestEinDoppelklickMeldetSeineEpoche(t *testing.T) {
	a := gestartet(t)

	epochen := make(chan any, 2)
	a.MitMeldung(func(_ context.Context, ereignis string, werte ...any) {
		if ereignis == EreignisDateiGeoeffnet && len(werte) == 2 {
			epochen <- werte[1]
		}
	})
	a.Bereit(t.Context())

	_ = a.imTausch(func() error { return nil }) //nolint:errcheck // nur hochzählen

	a.DateiOeffnen(filepath.Join("..", "..", "testdaten", "probe.mse-set"))

	select {
	case epoche := <-epochen:
		if epoche != a.lauf.epoche.Load() {
			t.Errorf("gemeldete Epoche %v, die des Bestands ist %d", epoche, a.lauf.epoche.Load())
		}
	case <-time.After(10 * time.Second):
		t.Fatal("keine Meldung mit Epoche")
	}
}

// **Beide Antworten nennen die Epoche, auch eine gescheiterte** (Befund 2 des
// Nachlaufs in J05): ein Rückweg, der scheiterte, kann einen anderen Bestand
// hinterlassen, und das Fenster muss es erfahren.
func TestDieAntwortenNennenDieEpoche(t *testing.T) {
	a := gestartet(t)
	a.lauf.epoche.Store(7)

	bilanz := a.Zuruecksetzen(probekontext(t), "layout", "", "", "", "", "")
	if bilanz.Bestandsepoche != 7 {
		t.Errorf("die Rücksetzung nennt %d, die Epoche ist 7", bilanz.Bestandsepoche)
	}

	// Eine Sicherung, die es nicht gibt: der Tausch läuft an und scheitert —
	// und zählt, wie jeder Tausch.
	antwort := a.SicherungZurueckspielen(probekontext(t), filepath.Join(t.TempDir(), "fehlt.msesicherung"), "")
	if antwort.Fehler == nil {
		t.Fatal("eine fehlende Sicherung kam durch")
	}

	if antwort.Bestandsepoche != a.lauf.epoche.Load() || antwort.Bestandsepoche < 7 {
		t.Errorf("das gescheiterte Zurückspielen nennt %d, die Epoche ist %d",
			antwort.Bestandsepoche, a.lauf.epoche.Load())
	}
}
