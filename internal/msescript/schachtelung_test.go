package msescript_test

// **Eine Schachtelung hat eine Grenze, und sie greift vor der Auswertung** —
// BK-08 (#2134,
// Vermerk 0177).
//
// Der Zerteiler stieg rekursiv hinab, ohne zu zählen: ein paar hundert
// Kilobyte Klammern aus einem fremden Paket brachten den Stapel zum Überlaufen
// — ein `fatal error`, den kein `recover` fängt. Jetzt endet jede Klasse der
// Rekursion an derselben Grenze mit einem Fehler an der Stelle, an der sie
// überschritten wurde; auch über eingebettete Textskripte hinweg, die einen
// eigenen Zerteiler bekommen.

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// geschachtelt baut je Rekursionsklasse einen Ausdruck der Tiefe `n`.
func geschachtelt(n int) map[string]string {
	text := "1"
	for range n {
		text = `"{` + text + `}"`
	}

	return map[string]string{
		"Klammern":           strings.Repeat("(", n) + "1" + strings.Repeat(")", n),
		"Vorzeichen":         strings.Repeat("-", n) + "1",
		"Listen":             strings.Repeat("[", n) + strings.Repeat("]", n),
		"Aufrufe":            strings.Repeat("f(", n) + "1" + strings.Repeat(")", n),
		"Zuweisungen":        strings.Repeat("a := ", n) + "1",
		"Bedingungen":        strings.Repeat("if 1 then ", n) + "1",
		"Blöcke":             strings.Repeat("{", n) + "1" + strings.Repeat("}", n),
		"eingebettete Texte": text,
	}
}

func TestEineZuTiefeSchachtelungIstEinFehlerAnIhrerStelle(t *testing.T) {
	t.Parallel()

	for klasse, quelle := range geschachtelt(msescript.HoechsteSchachtelung + 10) {
		t.Run(klasse, func(t *testing.T) {
			t.Parallel()

			_, fehler := msescript.Zerteile(quelle)

			var befund *msescript.Fehler
			if !errors.As(fehler, &befund) || befund.Schluessel != msescript.SchluesselZuTiefGeschachtelt {
				t.Fatalf("erwartet war der Befund zu tiefer Schachtelung, gekommen ist %v", fehler)
			}

			if befund.Zeile < 1 || befund.Spalte < 1 {
				t.Errorf("ohne Stelle: %+v", befund)
			}
		})
	}
}

func TestEineSchachtelungUnterDerGrenzeGehtDurch(t *testing.T) {
	t.Parallel()

	// **Ein Drittel der Grenze** — mehr, als ein Spielpaket je braucht:
	// `magic.mse-game/script` schachtelt höchstens 13 Klammern tief.
	for klasse, quelle := range geschachtelt(msescript.HoechsteSchachtelung / 3) {
		t.Run(klasse, func(t *testing.T) {
			t.Parallel()

			if _, fehler := msescript.Zerteile(quelle); fehler != nil {
				var befund *msescript.Fehler
				if errors.As(fehler, &befund) && befund.Schluessel == msescript.SchluesselZuTiefGeschachtelt {
					t.Fatalf("unter der Grenze abgewiesen: %v", fehler)
				}
			}
		})
	}
}

// **Eine Million Klammern endet mit einem Fehler und nicht mit dem Prozess.**
func TestEineExtremeSchachtelungEndetOhneAbsturz(t *testing.T) {
	t.Parallel()

	n := 1_000_000

	if _, fehler := msescript.Zerteile(strings.Repeat("(", n) + "1" + strings.Repeat(")", n)); fehler == nil {
		t.Error("eine Million Klammern wurden angenommen")
	}
}

// **Auch durch den zweiten Eingang** — das Textskript (Zweitblick über E,
// Befund 1). `AuswertenTextskript` verwarf den Fehler des Zerteilers, weil er
// vor der Grenze keinen haben konnte, und wertete einen leeren Knoten aus:
// ein `panic` statt des Befunds.
func TestEineZuTiefeSchachtelungImTextskriptIstEinFehler(t *testing.T) {
	t.Parallel()

	for klasse, quelle := range geschachtelt(msescript.HoechsteSchachtelung + 10) {
		t.Run(klasse, func(t *testing.T) {
			t.Parallel()

			_, fehler := msescript.AuswertenTextskript(t.Context(),
				"davor{"+quelle+"}danach", msescript.Bibliothek())

			var befund *msescript.Fehler
			if !errors.As(fehler, &befund) || befund.Schluessel != msescript.SchluesselZuTiefGeschachtelt {
				t.Fatalf("erwartet war der Befund zu tiefer Schachtelung, gekommen ist %v", fehler)
			}
		})
	}
}
