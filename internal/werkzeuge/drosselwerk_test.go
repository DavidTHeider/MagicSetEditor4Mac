// **Jeder Scryfall-Klient nimmt seine Drossel aus dem Werk** —
// Befund 2 des Zweitblicks über Phase V, 2026-08-24.
//
// # Warum es diesen Wächter gibt
//
// Der Befund war nicht, dass eine Drossel falsch gerechnet hätte — sie rechnete
// richtig. Der Befund war, dass es **vier** davon gab, jede mit eigenem
// Zustand, und dass der erste Aufruf einer Drossel absichtlich nicht wartet.
// Vier Drosseln, die einander nicht kennen, halten ein Budget nicht ein.
//
// **Der nächste Klient macht denselben Fehler**, wenn niemand nachzählt: das
// Werk anzuhängen ist eine Zeile, die man vergessen kann, und die Pipeline
// bliebe grün.
//
// # Was er nicht prüft
//
// **Ob das Werk je Prozess eines ist.** Er zählt Aufrufstellen, nicht
// Instanzen. Wer ein zweites `NeuesDrosselwerk` in `internal/app` anlegte und
// beide benutzte, käme durch — dagegen hilft nur, dass es genau ein
// `scryfallwerk` gibt, und das prüft der Lauf darunter.
package werkzeuge_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/werkzeuge"
)

// gelesen liest eine Datei aus dem Projekt.
func gelesenAus(t *testing.T, teile ...string) string {
	t.Helper()

	pfad := filepath.Join(append([]string{"..", ".."}, teile...)...)

	roh, fehler := os.ReadFile(pfad) //nolint:gosec // ein fester Pfad im Repositorium
	if fehler != nil {
		t.Fatalf("%s lesen: %v", pfad, fehler)
	}

	return string(roh)
}

// bauStellen findet jeden Aufruf von `scryfall.NeuerKlient`.
var bauStellen = regexp.MustCompile(`scryfall\.NeuerKlient\(`)

func TestJederScryfallklientNimmtSeineDrosselAusDemWerk(t *testing.T) {
	t.Parallel()

	dateien, fehler := filepath.Glob(filepath.Join("..", "app", "*.go"))
	if fehler != nil {
		t.Fatalf("suchen: %v", fehler)
	}

	// **Es heißt `drosselwerk` und nicht mehr `scryfallwerk`.** Seit #1100
	// teilt der Freigabezugang dasselbe Werk — je Gegenstelle eine Drossel,
	// über `Drosselwerk.Fuer`. Ein zweites Werk hieße zwei Zustände über
	// derselben Frage; der Wächter darüber zählt sie.
	gebaut, amWerk := 0, 0

	for _, pfad := range dateien {
		if strings.HasSuffix(pfad, "_test.go") {
			continue
		}

		// **Ohne Kommentare** (#1346):
		// hier stand `os.ReadFile`, und damit zählte ein `MitDrosselwerk(
		// drosselwerk)` im Kommentar mit. Der Nachlauf über Phase 4.4 hat es
		// nachgestellt — `MitDrosselwerk(nil)` im Code, die erwartete Form
		// darüber im Text, und der Lauf blieb grün.
		inhalt := werkzeuge.GoOhneKommentare(t, pfad)
		gebaut += len(bauStellen.FindAllString(inhalt, -1))
		amWerk += strings.Count(inhalt, "MitDrosselwerk(drosselwerk)")
	}

	if gebaut == 0 {
		t.Fatal("kein einziger Klient gefunden — der Wächter sucht am falschen Ort")
	}

	if amWerk != gebaut {
		t.Errorf("%d Klienten gebaut, %d nehmen ihre Drossel aus dem Werk — "+
			"das Budget von Scryfall gilt für Host und IP, nicht je Instanz",
			gebaut, amWerk)
	}
}

// **Und es gibt genau ein Werk.**
//
// Zwei Werke wären zwei Budgets. Der Lauf darüber sähe es nicht — er zählt
// Aufrufstellen.
func TestEsGibtGenauEinDrosselwerk(t *testing.T) {
	t.Parallel()

	dateien, fehler := filepath.Glob(filepath.Join("..", "app", "*.go"))
	if fehler != nil {
		t.Fatalf("suchen: %v", fehler)
	}

	werke := 0

	for _, pfad := range dateien {
		if strings.HasSuffix(pfad, "_test.go") {
			continue
		}

		roh, lesefehler := os.ReadFile(pfad) //nolint:gosec // ein fester Pfad im Repositorium
		if lesefehler != nil {
			t.Fatalf("%s lesen: %v", pfad, lesefehler)
		}

		werke += strings.Count(string(roh), "netz.NeuesDrosselwerk(")
	}

	if werke != 1 {
		t.Errorf("%d Drosselwerke in internal/app, erwartet genau eines", werke)
	}
}

// **Die Zusage in `DRITTANBIETER.md` hat einen Beleg.**
//
// Dort stand „429 sperrt 30 Sekunden", und der Code hielt es nicht: das 429
// wurde in einen Meldungstext umgerechnet und sonst nichts. Eine Zusage in
// einem Dokument, die der Code nicht hält, ist schlimmer als keine.
func TestDieSperrzusageHatEinenBeleg(t *testing.T) {
	t.Parallel()

	if !strings.Contains(gelesenAus(t, "DRITTANBIETER.md"), "sperrt **30 Sekunden**") {
		t.Error("DRITTANBIETER.md nennt die Sperrzeit nicht mehr")
	}

	// **Und `internal/netz` setzt sie.**
	hole := gelesenAus(t, "internal", "netz", "hole.go")

	for _, teil := range []string{"SperreOhneAngabe", "sperreSetzen(drossel"} {
		if !strings.Contains(hole, teil) {
			t.Errorf("internal/netz/hole.go führt %q nicht — dann sperrt ein 429 "+
				"nichts, und die Zusage im Dokument ist unbelegt", teil)
		}
	}
}

// **Und die Sperre gilt für das ganze Werk** —
// #1037.
//
// # Was dieser Wächter vorher nicht hielt
//
// Er suchte **Textfragmente** und einen Aufruf von `sperreSetzen`. Beides war
// da — und die Zusage aus `DRITTANBIETER.md` trotzdem verletzt: ein 429 sperrte
// nur die Drossel des gescheiterten Aufrufs, und die zweite Endpunktklasse lief
// weiter. Ein Wächter, der Wörter zählt, hält keine Wirkung (§10).
//
// **Was er jetzt hält**: die Wirkung, an derselben Gestalt, die der Betrieb
// benutzt — zwei Klassen aus einem Werk.
//
// **Was er nicht hält**: dass der Klient wirklich dieses Werk nimmt. Das prüft
// der Lauf darüber, an den Aufrufen von `MitDrosselwerk`.
func TestDieSperreGiltFuerBeideEndpunktklassen(t *testing.T) {
	t.Parallel()

	uhr := time.Date(2026, time.August, 24, 12, 0, 0, 0, time.UTC)
	werk := netz.NeuesDrosselwerk(func() time.Time { return uhr }, func(time.Duration) {})

	karten := werk.Fuer(scryfall.DrosselKarten, 600*time.Millisecond)
	allgemein := werk.Fuer(scryfall.DrosselAllgemein, 125*time.Millisecond)

	karten.SperreFuer(30 * time.Second)

	if !allgemein.Gesperrt() {
		t.Error("ein 429 auf einem Kartenweg lässt die allgemeine Klasse " +
			"weiterlaufen — DRITTANBIETER.md sagt das Gegenteil zu")
	}
}
