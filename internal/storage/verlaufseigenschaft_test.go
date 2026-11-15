package storage_test

// **Alles zurück ist der Anfang, alles wieder ist das Ende** — C05
// (#2230).
//
// Die übrigen Fälle des Verlaufs prüfen je einen Schritt, den jemand
// ausgedacht hat. Hier entstehen die Folgen aus einem Samen: Werte setzen,
// neue Felder anlegen, Felder entfernen, mehrmals dasselbe Feld — auch mit
// Umlauten, Zeichen außerhalb der Grundebene und Zeilenumbrüchen. Für jede
// Folge gilt:
//
//  1. Jede Rücknahme gibt es, und nach allen steht die Karte da wie vorher —
//     Werte **und** Reihenfolge der Felder.
//  2. Eine Rücknahme darüber hinaus gibt es nicht und ändert nichts.
//  3. Nach allen Wiederherstellungen steht die Karte da wie nach der Folge.
//
// **Was er nicht hält**: Folgen über mehrere Karten oder Sets und Schritte,
// die keine Feldänderung sind (Anlegen, Löschen, Tauschen) — die prüfen die
// Fälle daneben einzeln.

import (
	"database/sql"
	"fmt"
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// kartenstand ist eine Karte als Folge von „Feld=Wert" in ihrer Reihenfolge.
func kartenstand(t *testing.T, db *sql.DB, karteID int64) []string {
	t.Helper()

	zeilen, err := db.QueryContext(probekontext(t),
		`SELECT feld, wert FROM card_values WHERE card_id = ? ORDER BY reihenfolge, feld`, karteID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	defer func() { _ = zeilen.Close() }() //nolint:errcheck // gelesen ist gelesen

	var stand []string

	for zeilen.Next() {
		var feld, wert string
		if fehler := zeilen.Scan(&feld, &wert); fehler != nil {
			t.Fatalf("scannen: %v", fehler)
		}

		stand = append(stand, feld+"="+wert)
	}

	if fehler := zeilen.Err(); fehler != nil {
		t.Fatalf("Zeilen: %v", fehler)
	}

	return stand
}

var (
	verlaufsfelder = []string{"name", "text", "farbe", "neu", "zweites neues"}
	verlaufswerte  = []string{"Feuerball", "Drei Schaden", "rot", "Äußerst «grün»", "漢字 🃏", "zwei\nZeilen", "1"}
)

func TestAllesZurueckIstDerAnfangUndAllesWiederDasEnde(t *testing.T) {
	for samen := range uint64(40) {
		t.Run(fmt.Sprintf("Samen %d", samen), func(t *testing.T) {
			zufall := rand.New(rand.NewPCG(samen, 2230)) //nolint:gosec // eine Folge für einen Test, kein Geheimnis
			db := frischeDatenbank(t)
			karteID := dreifeldrigeKarte(t, db)
			setID := setVon(t, db, karteID)
			schreiber := storage.NeuerKartenschreiber(db)
			verlauf := storage.NeuerVerlaufsleser(db)

			anfang := kartenstand(t, db, karteID)
			schritte := 1 + zufall.IntN(10)
			folge := make([]string, 0, schritte)

			for range schritte {
				feld := verlaufsfelder[zufall.IntN(len(verlaufsfelder))]
				// **Entfernt wird nur, was da ist** — sonst gäbe es keinen
				// Schritt, und die Zahl der Rücknahmen stimmte nicht.
				if zufall.IntN(4) == 0 && slices.ContainsFunc(kartenstand(t, db, karteID), func(z string) bool {
					return len(z) > len(feld) && z[:len(feld)+1] == feld+"="
				}) {
					if _, _, _, err := schreiber.FeldEntfernen(probekontext(t), karteID, feld); err != nil {
						t.Fatalf("entfernen %q: %v", feld, err)
					}

					folge = append(folge, "−"+feld)

					continue
				}

				wert := verlaufswerte[zufall.IntN(len(verlaufswerte))]
				if _, _, err := schreiber.FeldSetzenMitVorwert(probekontext(t), karteID, feld, wert); err != nil {
					t.Fatalf("setzen %q: %v", feld, err)
				}

				folge = append(folge, feld+"="+wert)
			}

			ende := kartenstand(t, db, karteID)

			for schritt := range schritte {
				if _, gab, err := verlauf.Zuruecknehmen(probekontext(t), setID); err != nil || !gab {
					t.Fatalf("Rücknahme %d von %d (Folge %q): %v, gab=%v", schritt+1, schritte, folge, err, gab)
				}
			}

			if nachher := kartenstand(t, db, karteID); !slices.Equal(nachher, anfang) {
				t.Fatalf("nach allen Rücknahmen %q, vorher %q (Folge %q)", nachher, anfang, folge)
			}

			if _, gab, err := verlauf.Zuruecknehmen(probekontext(t), setID); err != nil || gab {
				t.Fatalf("eine Rücknahme zu viel: %v, gab=%v", err, gab)
			}

			for schritt := range schritte {
				if _, gab, err := verlauf.Wiederherstellen(probekontext(t), setID); err != nil || !gab {
					t.Fatalf("Wiederherstellung %d von %d (Folge %q): %v, gab=%v", schritt+1, schritte, folge, err, gab)
				}
			}

			if wieder := kartenstand(t, db, karteID); !slices.Equal(wieder, ende) {
				t.Errorf("nach allen Wiederherstellungen %q, nach der Folge %q (Folge %q)", wieder, ende, folge)
			}
		})
	}
}
