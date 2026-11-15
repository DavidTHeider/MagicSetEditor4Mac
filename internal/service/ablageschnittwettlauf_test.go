package service_test

// **Was zwischen Lesen und Löschen passieren kann** —
// #1282,
// Befund 7.
//
// **Eine eigene Datei**, weil `ablageruecknahme_test.go` mit ihr über 400
// Zeilen ginge (§2).

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ausschneiden verliert keine Änderung, die dazwischenkommt** —
// #1282,
// Befund 7.
//
// # Das Fenster hält die beiden nicht auseinander
//
// Ausschneiden nimmt den Vorgangsschlüssel `ablage-<setID>`, ein Feldschreiben
// `<karteID>/<feld>` — zwei verschiedene Schlüssel laufen nebeneinander. Und
// `ablegen` geht nach **draußen**, an die Zwischenablage des Systems; die
// Lücke dazwischen ist keine Millisekunde, sondern ein Systemaufruf.
//
// Gelöscht wurde danach allein nach Kennung. Der neue Stand war fort, ohne je
// in der Ablage gestanden zu haben.
func TestAusschneidenVerliertKeineAenderungDazwischen(t *testing.T) {
	t.Parallel()

	doppel := zweiSets()
	ablagen := kartenablage(doppel).MitLoescher(&ablageloescher{})

	// **Die Änderung kommt, während abgelegt wird** — genau in der Lücke.
	inhalt, fehler := ablagen.Ausschneiden(t.Context(), 1, []int64{11},
		func(string) error {
			doppel.karten[1][0].Values = []domain.Value{
				{Field: "name", Text: "Erste"},
				{Field: "rule text", Text: "Fliegend und flink"},
			}

			return nil
		})

	if !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("Fehler war %v — will ErrConflict", fehler)
	}

	// **Der abgelegte Inhalt ist der von vorhin** — er steht in der
	// Zwischenablage und soll benannt werden.
	if !strings.Contains(inhalt, "Fliegend\n") {
		t.Errorf("der abgelegte Inhalt ist nicht der gelesene:\n%s", inhalt)
	}

	// **Und die Karte steht noch da.**
	if len(doppel.karten[1]) != 2 {
		t.Errorf("%d Karten im Set — die Änderung ist mit der Karte "+
			"verschwunden", len(doppel.karten[1]))
	}
}

// **Und wenn das Nachlesen beim Ausschneiden scheitert, wird nicht gelöscht** —
// #1282,
// Befund 7.
//
// Der Vergleich vor dem Löschen ist die Zusage; kommt er nicht zustande, gilt
// dasselbe wie bei einer Abweichung: die Karte bleibt stehen.
func TestEinGescheitertesNachlesenLoeschtNicht(t *testing.T) {
	t.Parallel()

	doppel := zweiSets()
	loescher := &ablageloescher{}
	ablagen := kartenablage(doppel).MitLoescher(loescher)

	_, fehler := ablagen.Ausschneiden(t.Context(), 1, []int64{11},
		func(string) error {
			doppel.kartenfehler = errors.New("der Bestand antwortet nicht")

			return nil
		})

	if fehler == nil {
		t.Fatal("das Ausschneiden hat den Lesefehler nicht gemeldet")
	}

	if len(loescher.geloescht) != 0 {
		t.Errorf("es wurde gelöscht: %v — ohne Vergleich wird nicht gelöscht",
			loescher.geloescht)
	}
}
