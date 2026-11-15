package msescript_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **Wie MSE zwei Ordnungsschlüssel vergleicht** — `smart_compare`
// (#2006).
//
// Die Fälle sind dieselben, die `domain.Vergleiche` hält; ein Lauf in
// `internal/service` (`ordnungsgleich_test.go`) fährt beide Portierungen über
// dieselben Paare und über gefuzzte.
func TestKlugVergleicheWieMSE(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name          string
		links, rechts string
		erwartet      int
	}{
		{"gleiche Texte", "Karte", "Karte", 0},
		{"Zahl im Text als Zahl", "Karte 2", "Karte 10", -1},
		{"die längere Zahl ist die größere", "10/99", "2/99", 1},
		{"gleich lange Zahlen nach der ersten Abweichung", "12/99", "13/99", -1},
		{"gleich lange Zahlen, die größere zuerst gefragt", "13", "12", 1},
		{"eine Zahl, die weiterläuft, steht dahinter", "12", "123", -1},
		{"eine Zahl, die weiterläuft, von rechts", "123", "12", 1},
		{"gleiche Zahl, dann Text", "1a", "1b", -1},
		{"Zahl gegen Ziffernfolge mit Text dahinter", "12x", "123", -1},
		{"groß und klein sind gleich", "apfel", "Apfel", 0},
		{"Umlaut beim Grundbuchstaben", "Ärger", "Banane", -1},
		{"Umlaut hinter dem Grundbuchstaben", "apfel", "Ärger", -1},
		{"Latin Extended A", "Łódź", "Lublin", -1},
		{"zerlegter Buchstabe zwischen Aegis und Afar", "Æther", "Afar", -1},
		{"zerlegter Buchstabe hinter Aegis", "Aegis", "Æther", -1},
		{"zerlegter Buchstabe rechts", "Afar", "Æther", 1},
		{"zwei zerlegte Buchstaben", "Æ", "Œ", -1},
		{"links zu Ende mitten in der Zerlegung — MSE gibt 1", "A", "Æ", 1},
		{"rechts zu Ende mitten in der Zerlegung — MSE gibt −1", "Æ", "A", -1},
		{"Zerlegung, links ohne zweiten Buchstaben", "ae", "Æ", 0},
		{"ij als zwei Buchstaben", "ĳ", "ik", -1},
		{"ǅ ordnet nach seiner Kleinschreibung", "ǅ", "dy", 1},
		{"dz zerfällt in groß und klein gleich", "ǅ", "ǆ", 0},
		{"Steuerzeichen nach dem Codepunkt", "a\x01", "a\x02", -1},
		{"Steuerzeichen rechts", "a\x02", "a\x01", 1},
		{"ƀ ist b", "ƀ", "b", 0},
		{"Ḁ ist a", "Ḁ", "a", 0},
		{"ein Zeichen ohne Grundbuchstaben fällt in die Kleinschreibung", "ẹ", "z", 1},
		{"Leerzeichen im Latin-1-Block", "×", "x", 1},
		{"ſ liegt hinter der Tabelle und ordnet nach sich selbst", "ſ", "s", 1},
		{"kürzerer Text zuerst", "Kart", "Karte", -1},
		{"längerer Text danach", "Karte", "Kart", 1},
		{"beide enden in derselben Zahl", "Karte 7", "Karte 7", 0},
		{"Ende in einer Zahl, die links kürzer ist", "a1", "a1b", -1},
		{"Ende in einer Zahl, die rechts kürzer ist", "a1b", "a1", 1},
		{"Ende in verschiedenen Zahlen, links kleiner", "a2", "a3", -1},
		{"Ende in verschiedenen Zahlen, links größer", "a3", "a2", 1},
		{"Ende in einer Zahl, links kürzer, verschieden", "a3", "a2b", 1},
		{"Ende in einer Zahl, rechts kürzer, verschieden", "a2b", "a3", -1},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			if ergebnis := msescript.KlugVergleiche(fall.links, fall.rechts); ergebnis != fall.erwartet {
				t.Errorf("KlugVergleiche(%q, %q) = %d, erwartet %d",
					fall.links, fall.rechts, ergebnis, fall.erwartet)
			}
		})
	}
}
