// **Ordnen und Zeigen in einem Stand** — Befund 13 des Zweitblicks über
// Phase 5.2 (#1435).
//
// **Kein `t.Parallel()`**, aus demselben Grund wie in `sammlung_test.go`.
package storage_test

import (
	"fmt"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Das Fenster liest zweimal und gibt, was die Wahl nennt.**
func TestDasFensterLiestBeidesInEinemStand(t *testing.T) {
	db := frischeDatenbank(t)
	speicher := storage.NeuerSammlungsspeicher(db)

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

	var gesehen int

	fenster, fehler := speicher.Fenster(probekontext(t),
		func(ordnung []*domain.Exemplar) []int64 {
			gesehen = len(ordnung)

			// **Rückwärts** — die Reihenfolge kommt von hier.
			return []int64{kennungen[2], kennungen[0]}
		})
	if fehler != nil {
		t.Fatalf("das Fenster: %v", fehler)
	}

	if gesehen != 3 {
		t.Errorf("die Wahl sah %d Zeilen", gesehen)
	}

	if len(fenster) != 2 || fenster[0].Name != "Karte 2" || fenster[1].Name != "Karte 0" {
		t.Fatalf("das Fenster ist %v", namen(fenster))
	}

	// **Und die Zeilen sind vollständig** — der Unterschied zur Ordnungsliste.
	if fenster[0].Zustand != domain.ZustandNearMint {
		t.Errorf("der Zustand fehlt: %+v", fenster[0])
	}
}

// **Ohne Profil, ohne Bestand: beides meldet sich.**
func TestDasFensterMeldetSeineFehlerwege(t *testing.T) {
	db := frischeDatenbank(t)
	speicher := storage.NeuerSammlungsspeicher(db)

	waehleNichts := func([]*domain.Exemplar) []int64 { return nil }

	if _, fehler := speicher.Fenster(t.Context(), waehleNichts); fehler == nil {
		t.Error("ohne Profil kam es durch")
	}

	if fehler := db.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	if _, fehler := speicher.Fenster(probekontext(t), waehleNichts); fehler == nil {
		t.Error("der geschlossene Bestand gab ein Fenster her")
	}
}

// **Und ein Fehler beim Holen der Zeilen kommt heraus.**
//
// **Über eine entfernte Tafel**: die Ordnung steht dann noch, und erst der
// zweite Griff scheitert — genau der Weg, der sonst unbetreten bliebe.
func TestDasFensterMeldetEinenFehlerBeimHolen(t *testing.T) {
	db := frischeDatenbank(t)
	speicher := storage.NeuerSammlungsspeicher(db)

	if _, fehler := speicher.Anlegen(probekontext(t), probeexemplar()); fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	if _, fehler := db.ExecContext(probekontext(t),
		`DROP TABLE sammlung_herkunft`); fehler != nil {
		t.Fatalf("die Tafel entfernen: %v", fehler)
	}

	_, fehler := speicher.Fenster(probekontext(t),
		func(ordnung []*domain.Exemplar) []int64 {
			kennungen := make([]int64, 0, len(ordnung))
			for _, eines := range ordnung {
				kennungen = append(kennungen, eines.ID)
			}

			return kennungen
		})

	if fehler == nil {
		t.Error("es kam durch")
	}
}

// **Und ein Fehler beim Ordnen kommt heraus** — der erste Griff im Geschäft.
func TestDasFensterMeldetEinenFehlerBeimOrdnen(t *testing.T) {
	db := frischeDatenbank(t)
	speicher := storage.NeuerSammlungsspeicher(db)

	if _, fehler := db.ExecContext(probekontext(t), `DROP TABLE sammlung`); fehler != nil {
		t.Fatalf("die Tafel entfernen: %v", fehler)
	}

	if _, fehler := speicher.Fenster(probekontext(t),
		func([]*domain.Exemplar) []int64 { return nil }); fehler == nil {
		t.Error("es kam durch")
	}
}
