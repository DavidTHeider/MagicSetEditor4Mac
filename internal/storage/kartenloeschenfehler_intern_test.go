package storage

import (
	"database/sql"
	"testing"
)

// **Die Fehlerwege des Zuges über mehrere Karten** — Befund 4 des Zweitblicks
// über Phase 3.7
// (#1168).
//
// Paketintern und über den Stolpertreiber, aus demselben Grund wie beim
// Löschen einer einzelnen Karte nebenan: mit einer funktionierenden Datenbank
// sind sie nicht erreichbar, weil sie funktioniert.
//
// **Und sie zählen hier doppelt.** Die Zusage dieses Weges ist „ganz oder gar
// nicht"; jeder dieser Fehler muss die Transaktion zurückrollen, statt einen
// halben Schnitt zu hinterlassen.
func TestDerZugMeldetSeineFehlerwege(t *testing.T) {
	faelle := map[string]string{
		"die Transaktion beginnt nicht": "",
		"das Lesen der Karte scheitert": "stolper-fragen",
		"der Verlauf scheitert":         "stolper-karte-verlauf",
		"das Wegschreiben scheitert":    "stolper-karte-schreiben",
		"der Commit scheitert":          "stolper-karte-commit",
	}

	for name, treiber := range faelle {
		t.Run(name, func(t *testing.T) {
			zweiterSchreibzaehler = 0

			db := geschlosseneDatenbank(t)

			if treiber != "" {
				geoeffnet, err := sql.Open(treiber, "")
				if err != nil {
					t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
				}

				t.Cleanup(func() { _ = db.Close() })

				db = geoeffnet
			}

			if err := NeuerKartenschreiber(db).KartenLoeschen(probekontext(t), []int64{1, 2}); err == nil {
				t.Errorf("%s: kein Fehler", name)
			}
		})
	}
}

// **Eine leere Liste ist kein Fehler und kein Zug.**
//
// Sie kommt vor: wer ⌘X ohne Haken und ohne aufgeschlagene Karte drückt,
// schneidet nichts aus. Eine Transaktion dafür zu öffnen wäre Arbeit für
// nichts — und ein Fehler wäre eine Meldung über etwas, das niemand versucht
// hat (§8).
func TestEineLeereListeIstKeinZug(t *testing.T) {
	t.Parallel()

	if err := NeuerKartenschreiber(geschlosseneDatenbank(t)).
		KartenLoeschen(probekontext(t), nil); err != nil {
		t.Errorf("Fehler bei leerer Liste: %v", err)
	}
}
