// Was eine Bewegung des Verlaufs zurückmeldet.
//
// **Nicht nur „es gab etwas".** Eine Rücknahme stellt genau **einen** Datensatz
// wieder her (AGENTS.md §7); wer nur erfährt, dass etwas geschehen ist, lädt den
// ganzen Stand nach — und überschreibt damit alles, woran gerade jemand tippt.
package storage_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestEineBewegungNenntIhreKarte(t *testing.T) {
	db, setID, karteID := bestandMitKarte(t)

	if _, _, err := storage.NeuerKartenschreiber(db).
		FeldSetzenMitVorwert(probekontext(t), karteID, "name", "Eisball"); err != nil {
		t.Fatalf("setzen: %v", err)
	}

	leser := storage.NeuerVerlaufsleser(db)

	genommen, gab, err := leser.Zuruecknehmen(probekontext(t), setID)
	if err != nil || !gab {
		t.Fatalf("zurücknehmen: %v / %v", gab, err)
	}

	if genommen.KarteID != karteID || genommen.Feld != "name" {
		t.Errorf("%+v — erwartet war Karte %d, Feld name", genommen, karteID)
	}

	// **Und die Gegenrichtung ebenso.** Sie geht über eine andere Abfrage.
	gebracht, gab, err := leser.Wiederherstellen(probekontext(t), setID)
	if err != nil || !gab {
		t.Fatalf("wiederherstellen: %v / %v", gab, err)
	}

	if gebracht.KarteID != karteID || gebracht.Feld != "name" {
		t.Errorf("%+v — erwartet war Karte %d, Feld name", gebracht, karteID)
	}
}

// **Ein leerer Verlauf nennt keine Karte** — und ist kein Fehler.
func TestEineBewegungOhneEintragNenntNichts(t *testing.T) {
	db, setID, _ := bestandMitKarte(t)

	eintrag, gab, err := storage.NeuerVerlaufsleser(db).Zuruecknehmen(probekontext(t), setID)
	if err != nil || gab {
		t.Fatalf("zurücknehmen: %v / %v", gab, err)
	}

	if eintrag.KarteID != 0 {
		t.Errorf("%+v — erwartet war ein leerer Eintrag", eintrag)
	}
}
