package storage

// **Zwei Prozesse legen denselben neuen Bestand an** — etwa das Fenster und
// der Dienst beim ersten Start. Beide lesen die Buchführung, bevor einer die
// Schreibsperre hat (`_txlock=immediate`); der zweite wartet, bis der erste
// fertig ist, und darf danach nicht mehr mit dem leeren Stand von vorher
// arbeiten. Bis 0.7.0 wandte er dann `0001` noch einmal an und scheiterte an
// einer Tabelle, die es schon gab; mit der Grundlinie hielte er den frischen
// Bestand für einen fremden (Vermerk 0189).

import (
	"errors"
	"path/filepath"
	"sync"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Gestellt, damit der Fall jedes Mal eintritt**: der zweite hat leer
// gelesen, dann legt der erste an.
func TestEinZweiterProzessArbeitetMitDerBuchfuehrungUnterDerSperre(t *testing.T) {
	db := leererBestand(t)

	if fehler := buchfuehrungAnlegen(t.Context(), db); fehler != nil {
		t.Fatal(fehler)
	}

	veraltet, fehler := angewendeteMitSummen(t.Context(), db)
	if fehler != nil || len(veraltet) != 0 {
		t.Fatalf("vorher: %v, %v", veraltet, fehler)
	}

	if fehler := Migriere(t.Context(), db, nil); fehler != nil {
		t.Fatalf("der erste: %v", fehler)
	}

	umfang, summen, _ := grundlinienumfang(grundlinie) //nolint:errcheck // eingebettet

	if fehler := anwendenAlle(t.Context(), db, migrationen, danach(t),
		PruefeBestand, grundlinie, umfang, summen); fehler != nil {
		t.Errorf("der zweite: %v", fehler)
	}
}

// **Und echt, mit zwei Verbindungen auf einer Datei** — so oft, dass der
// Wettlauf mehr als einmal eintritt.
func TestZweiProzesseLegenEinenBestandZugleichAn(t *testing.T) {
	for lauf := range 8 {
		// **Ein Pfad und keine `file:`-Adresse**, wie die Anwendung öffnet:
		// nur dann nimmt `Oeffnen` die Nachbardatei (`sperre.go`).
		ort := filepath.Join(t.TempDir(), "bestand.db")

		var (
			warten sync.WaitGroup
			fehler [2]error
		)

		for i := range fehler {
			warten.Add(1)

			go func() {
				defer warten.Done()

				db, oeffnungsfehler := Oeffnen(t.Context(), ort)
				if oeffnungsfehler != nil {
					fehler[i] = oeffnungsfehler

					return
				}

				defer func() { _ = db.Close() }() //nolint:errcheck // Aufräumen

				fehler[i] = Migriere(t.Context(), db, nil)
			}()
		}

		warten.Wait()

		for i, f := range fehler {
			if f != nil {
				t.Errorf("Lauf %d, Prozess %d: %v", lauf, i, f)
			}
		}
	}
}

// **Und was unter der Sperre steht, wird auch geprüft** (Zweitblick über die
// Folgen vor 0.8.0, Befund 3): hat eine neuere Fassung zwischen Lesen und
// Sperre `0098` verbucht, wird diese Fassung abgewiesen — und kehrt nicht
// still zurück, als stünde der Bestand auf ihrem Stand.
func TestUnterDerSperreWirdAuchGeprueft(t *testing.T) {
	db := leererBestand(t)

	if fehler := Migriere(t.Context(), db, nil); fehler != nil {
		t.Fatal(fehler)
	}

	neuere, _ := kuenftigeWanderung(t, "neu", "CREATE TABLE neu (x);")
	if fehler := MigriereMit(t.Context(), db, neuere, nil); fehler != nil {
		t.Fatalf("die neuere Fassung: %v", fehler)
	}

	umfang, summen, _ := grundlinienumfang(grundlinie) //nolint:errcheck // eingebettet

	var neuer *domain.NeuereFassung
	if fehler := anwendenAlle(t.Context(), db, migrationen, danach(t),
		PruefeBestand, grundlinie, umfang, summen); !errors.As(fehler, &neuer) {
		t.Errorf("der Fehler ist %v, erwartet eine neuere Fassung", fehler)
	}
}
