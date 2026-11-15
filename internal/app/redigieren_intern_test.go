package app

// **Was aus einem Text heraus muss** — Werkstück O10 von Phase 5.2
// (#1434).

import (
	"os"
	"strings"
	"testing"
)

func TestRedigiereNimmtNachweiseUndLangeZufallsfolgen(t *testing.T) {
	zuhause, fehler := os.UserHomeDir()
	if fehler != nil {
		t.Fatalf("Zuhause: %v", fehler)
	}

	faelle := []struct {
		name   string
		roh    string
		bleibt string
		weg    string
	}{
		{
			name:   "der Benutzerpfad wird zur Tilde",
			roh:    "ort=" + zuhause + "/Sets/probe.mse-set",
			bleibt: "probe.mse-set",
			weg:    zuhause,
		},
		{
			name:   "ein Bearer-Nachweis fällt",
			roh:    `grund="Authorization: Bearer geheimnis-des-hauses" abgewiesen`,
			bleibt: "abgewiesen",
			weg:    "geheimnis-des-hauses",
		},
		{
			name:   "eine lange Folge aus Buchstaben und Ziffern fällt",
			roh:    "grund=das Token m5e0Token4711abcdefghijklmnopqrs ist unbekannt",
			bleibt: "ist unbekannt",
			weg:    "m5e0Token4711abcdefghijklmnopqrs",
		},
		{
			name: "ein langes deutsches Wort bleibt",
			// **Die Gegenprobe zur Regel**: ohne die Prüfung auf Ziffern
			// **und** Buchstaben verschwände hier ein Bezeichner, den die
			// Diagnose braucht.
			roh:    "stelle=Bestandsverschluesselungsstand",
			bleibt: "Bestandsverschluesselungsstand",
			weg:    Geschwaerzt,
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			gesehen := redigiere(fall.roh)

			if !strings.Contains(gesehen, fall.bleibt) {
				t.Errorf("%q fehlt in %q", fall.bleibt, gesehen)
			}

			if strings.Contains(gesehen, fall.weg) {
				t.Errorf("%q steht noch in %q", fall.weg, gesehen)
			}
		})
	}
}
