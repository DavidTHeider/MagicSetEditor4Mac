// **Was SQLite selbst über den Bestand sagt** — Befund 2 des Zweitblicks über
// Phase 5.2 (#1435).
//
// **Eine eigene Datei**, weil `datenpflege_test.go` an der Zeilengrenze steht
// (§2) — und weil diese beiden Läufe eine eigene Frage stellen: nicht „ist ein
// bekannter Rest übrig", sondern „ist die Datei heil".
package storage_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Eine verwaiste Zeile fällt auf — und wird nicht gelöscht.**
//
// Befund 2 des Zweitblicks über Phase 5.2: die Pflege sah nur nach
// **bekannten** Resten; was SQLite selbst über seine Datei sagt, stand nicht
// darin. Das Issue nennt „Integrität/Fremdschlüssel" zuerst — und sagt im
// selben Atemzug: „echte Korruption nicht durch kosmetisches Löschen
// verdecken". Deshalb ist der Befund eine **Auskunft**.
func TestEineVerwaisteZeileFaelltAlsFremdschluesselAuf(t *testing.T) {
	db := frischeDatenbank(t)

	// **Die Prüfung wird für diesen Zug abgeschaltet** — anders entstünde die
	// verwaiste Zeile gar nicht, und der Befund prüfte einen Fall, den es im
	// Betrieb nicht gibt. Genau darum geht es: er entsteht **außerhalb** des
	// Betriebs, durch eine Wanderung, ein fremdes Werkzeug, eine Datei von
	// Hand.
	waiseLegen(t, db, `INSERT INTO deckplaetze (deck_id, bereich, menge, printing)
	                        VALUES (9999, 'hauptdeck', 1, 'probe-druck')`)

	befunde, fehler := storage.Pflegeanalyse(probekontext(t), db, pflegezeit)
	if fehler != nil {
		t.Fatalf("die Analyse: %v", fehler)
	}

	gefunden := false

	for _, befund := range befunde {
		if befund.Art != "fremdschluessel" {
			continue
		}

		gefunden = true

		if befund.Klasse != domain.PflegeNurInformieren {
			t.Errorf("die Klasse ist %q — hier wird nichts gelöscht", befund.Klasse)
		}
	}

	if !gefunden {
		t.Errorf("die verwaiste Zeile fiel nicht auf: %+v", befunde)
	}
}

// **Und ein heiler Bestand meldet weder das eine noch das andere.**
func TestEinHeilerBestandMeldetKeineIntegritaetsbefunde(t *testing.T) {
	befunde, fehler := storage.Pflegeanalyse(probekontext(t), frischeDatenbank(t),
		pflegezeit)
	if fehler != nil {
		t.Fatalf("die Analyse: %v", fehler)
	}

	for _, befund := range befunde {
		if befund.Art == "integritaet" || befund.Art == "fremdschluessel" {
			t.Errorf("ein heiler Bestand meldet %q: %+v", befund.Art, befund)
		}
	}
}
