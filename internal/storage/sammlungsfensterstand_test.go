// **Ein Stand und nicht zwei** — Hinweis 13 des Nachlaufs zum Zweitblick über
// Phase 5.2 (#1435).
//
// # Warum es diesen Lauf gibt
//
// `Fenster` hält die Zusage, dass Ordnung und Zeilen **dieselbe** Sammlung
// sehen. Der Lauf nebenan las drei unveränderte Zeilen; wurde das Geschäft
// durch zwei unabhängige Abfragen ersetzt, blieb er grün — und mit ihm die
// ganze Storage-Suite. Eine Zusage über einen Stand, die niemand verletzen
// kann, ist keine geprüfte Zusage (§10).
//
// **Hier schreibt jemand dazwischen.** `waehle` läuft zwischen den beiden
// Abfragen, und dieser Lauf benutzt genau das: aus dem Rückruf heraus legt
// eine **zweite Verbindung** ein Exemplar an. Wer danach noch dieselbe
// Sammlung sieht, hat einen Stand; wer die neue Zeile sieht, hat zwei.
//
// **Auf der Platte und mit zwei Verbindungen**: ein Bestand im Arbeitsspeicher
// gehört einer Verbindung, und ein zweiter Schreiber auf derselben wäre kein
// zweiter.
//
// **Kein `t.Parallel()`**, aus demselben Grund wie in `sammlung_test.go`.
package storage_test

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestEinFremderSchreiberAendertDasLaufendeFensterNicht(t *testing.T) {
	ort := filepath.Join(t.TempDir(), "bestand.db")

	erste := geoeffnet(t, ort)
	if err := storage.Migriere(probekontext(t), erste, nil); err != nil {
		t.Fatalf("migrieren: %v", err)
	}

	speicher := storage.NeuerSammlungsspeicher(erste)
	kennungen := dreiExemplare(t, speicher)

	// **Die zweite Verbindung ist wirklich eine zweite** — dieselbe Datei,
	// ein eigener Zugang, ein eigener Stand.
	zweite := geoeffnet(t, ort)
	fremder := storage.NeuerSammlungsspeicher(zweite)

	var sahDieOrdnung, sahDieWahl int

	fenster, fehler := speicher.Fenster(probekontext(t),
		func(ordnung []*domain.Exemplar) []int64 {
			sahDieOrdnung = len(ordnung)

			dazwischenSchreiben(t, fremder)

			// **Alles, was die Ordnung genannt hat** — käme die neue Zeile
			// durch, stünde sie im Ergebnis, obwohl sie hier nie gewählt wurde.
			gewaehlt := make([]int64, 0, len(ordnung))
			for _, eines := range ordnung {
				gewaehlt = append(gewaehlt, eines.ID)
			}

			// **Und die neue dazu** — so würde sie wirklich sichtbar, wenn der
			// zweite Griff einen anderen Stand läse.
			sahDieWahl = len(gewaehlt)

			return append(gewaehlt, kennungen[len(kennungen)-1]+1)
		})
	if fehler != nil {
		t.Fatalf("das Fenster: %v", fehler)
	}

	if sahDieOrdnung != 3 || sahDieWahl != 3 {
		t.Fatalf("die Ordnung sah %d, die Wahl %d Zeilen", sahDieOrdnung, sahDieWahl)
	}

	if len(fenster) != 3 {
		t.Errorf("das Fenster zeigt %d Zeilen: %v — der fremde Schreiber kam durch",
			len(fenster), namen(fenster))
	}

	// **Die Gegenprobe**: ein **neues** Fenster sieht ihn sehr wohl.
	//
	// Ohne sie sagte der Lauf nur, dass die vierte Zeile nirgends auftaucht —
	// auch dann, wenn der fremde Schreiber gar nichts geschrieben hätte.
	danach, fehler := speicher.Fenster(probekontext(t),
		func(ordnung []*domain.Exemplar) []int64 {
			gewaehlt := make([]int64, 0, len(ordnung))
			for _, eines := range ordnung {
				gewaehlt = append(gewaehlt, eines.ID)
			}

			return gewaehlt
		})
	if fehler != nil {
		t.Fatalf("das zweite Fenster: %v", fehler)
	}

	if len(danach) != 4 {
		t.Errorf("danach zeigt das Fenster %d Zeilen: %v — geschrieben wurde nichts",
			len(danach), namen(danach))
	}
}

// dreiExemplare legt drei unterscheidbare Exemplare an.
func dreiExemplare(t *testing.T, speicher *storage.Sammlungsspeicher) []int64 {
	t.Helper()

	kennungen := make([]int64, 0, 3)

	for zaehler := range 3 {
		exemplar := probeexemplar()
		exemplar.Printing = fmt.Sprintf("00000000-0000-4000-8000-%012d", zaehler)
		exemplar.Name = fmt.Sprintf("Karte %d", zaehler)

		id, fehler := speicher.Anlegen(probekontext(t), exemplar)
		if fehler != nil {
			t.Fatalf("anlegen: %v", fehler)
		}

		kennungen = append(kennungen, id)
	}

	return kennungen
}

// dazwischenSchreiben legt aus dem Rückruf heraus ein viertes Exemplar an.
func dazwischenSchreiben(t *testing.T, fremder *storage.Sammlungsspeicher) {
	t.Helper()

	exemplar := probeexemplar()
	exemplar.Printing = "00000000-0000-4000-8000-000000000009"
	exemplar.Name = "Karte dazwischen"

	if _, fehler := fremder.Anlegen(probekontext(t), exemplar); fehler != nil {
		t.Fatalf("der fremde Schreiber: %v", fehler)
	}
}
