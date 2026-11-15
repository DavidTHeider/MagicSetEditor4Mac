package domain_test

import (
	"regexp"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestVergleichscode(t *testing.T) {
	t.Parallel()

	anna := "12D3KooWAnnaAnnaAnnaAnnaAnnaAnnaAnnaAnnaAnnaAnnaAnna"
	bert := "12D3KooWBertBertBertBertBertBertBertBertBertBertBert"
	carl := "12D3KooWCarlCarlCarlCarlCarlCarlCarlCarlCarlCarlCarl"

	t.Run("beide Seiten sehen dieselben Ziffern", func(t *testing.T) {
		t.Parallel()

		if domain.Vergleichscode(anna, bert) != domain.Vergleichscode(bert, anna) {
			t.Fatal("die Reihenfolge der Kennungen verändert den Code")
		}
	})

	t.Run("sechs Gruppen zu fünf Ziffern", func(t *testing.T) {
		t.Parallel()

		gestalt := regexp.MustCompile(`^\d{5}( \d{5}){5}$`)
		if code := domain.Vergleichscode(anna, bert); !gestalt.MatchString(code) {
			t.Fatalf("Gestalt: %q", code)
		}
	})

	t.Run("ein anderer Schlüssel gibt einen anderen Code", func(t *testing.T) {
		t.Parallel()

		if domain.Vergleichscode(anna, bert) == domain.Vergleichscode(anna, carl) {
			t.Fatal("zwei Gegenüber, ein Code")
		}
	})

	t.Run("jede Hälfte gehört einer Seite", func(t *testing.T) {
		t.Parallel()

		// **Die Gestalt aus Signal**: jede Hälfte hängt nur an einem
		// Schlüssel — wer dasselbe Gegenüber behält, behält seine Hälfte.
		ab, ac := domain.Vergleichscode(anna, bert), domain.Vergleichscode(anna, carl)
		haelften := func(code string) [2]string { return [2]string{code[:17], code[18:]} }

		a, b := haelften(ab), haelften(ac)
		if a[0] != b[0] && a[0] != b[1] && a[1] != b[0] && a[1] != b[1] {
			t.Fatalf("keine gemeinsame Hälfte: %q gegen %q", ab, ac)
		}
	})

	t.Run("festgehalten gegen eine frühere Rechnung", func(t *testing.T) {
		t.Parallel()

		// **Der Wert ist unabhängig gerechnet** — mit Pythons `hashlib`
		// nach der Beschreibung im Kopf von `vergleichscode.go`, nicht mit
		// diesem Code. Ändert sich die Ableitung, sehen zwei Fassungen
		// verschiedene Ziffern für dasselbe Paar, und jeder Vergleich
		// scheiterte.
		const erwartet = "49311 67386 01529 95503 63193 56021"
		if code := domain.Vergleichscode(anna, bert); code != erwartet {
			t.Fatalf("Vergleichscode = %q", code)
		}
	})
}

func TestVergleichscodePasst(t *testing.T) {
	t.Parallel()

	anna := "12D3KooWAnnaAnnaAnnaAnnaAnnaAnnaAnnaAnnaAnnaAnnaAnna"
	bert := "12D3KooWBertBertBertBertBertBertBertBertBertBertBert"
	code := domain.Vergleichscode(anna, bert)

	faelle := []struct {
		name    string
		eingabe string
		passt   bool
	}{
		{"wörtlich", code, true},
		{"ohne Leerzeichen", regexp.MustCompile(` `).ReplaceAllString(code, ""), true},
		{"mit Bindestrichen", regexp.MustCompile(` `).ReplaceAllString(code, "-"), true},
		{"eine Ziffer anders", andereZiffer(code), false},
		{"leer", "", false},
		{"nur die Hälfte", code[:17], false},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			if got := domain.VergleichscodePasst(anna, bert, fall.eingabe); got != fall.passt {
				t.Fatalf("VergleichscodePasst(%q) = %v", fall.eingabe, got)
			}
		})
	}
}

// andereZiffer tauscht die erste Ziffer gegen eine sicher andere.
func andereZiffer(code string) string {
	if code[0] == '0' {
		return "1" + code[1:]
	}

	return "0" + code[1:]
}
