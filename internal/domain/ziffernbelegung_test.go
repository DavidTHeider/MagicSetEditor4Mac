package domain_test

// **Die Belegung von ⌘0 bis ⌘9 ist einstellbar** — UX-02.AK-03
// (#1674):
// *„bestehende Kürzel werden erhalten oder mit dokumentierter Migration und
// konfigurierbarer Belegung ersetzt."* Die Migration steht seit #1753 in
// `CHANGELOG.md`; die Belegung fehlte.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestPruefeZiffernbelegung(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name string
		wert string
		gut  bool
	}{
		{"leer heißt die Vorgabe", "", true},
		{"eine Ziffer", "1=start", true},
		{"mehrere, in beliebiger Folge", "3=karten,1=start,0=handeln", true},
		{"eine Ziffer ohne Seite räumt sie", "1=start,2=", true},
		{"dieselbe Ziffer zweimal", "1=start,1=sets", false},
		{"ein Buchstabe statt einer Ziffer", "a=start", false},
		{"zwei Ziffern", "10=start", false},
		{"ohne Gleichheitszeichen", "1start", false},
		{"eine Seite mit Großbuchstaben", "1=Start", false},
		{"eine Seite mit Zeichen", "1=start;2", false},
		{"ein leerer Eintrag", "1=start,,2=sets", false},

		// **Eine Seite, die es nicht gibt** — Befund G13 des Zweitblicks über
		// Phase 6.3 (#1665): MCP, Kommandozeile und REST nahmen `3=kartne`
		// an, das Fenster fiel still auf die Vorgabe und löschte den Eintrag
		// bei der nächsten Änderung.
		{"eine Seite, die es nicht gibt", "3=kartne", false},
		{"eine Seite aus dem Setbereich", "7=druckbogen", true},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			fehler := domain.PruefeZiffernbelegung(fall.wert)
			if fall.gut && fehler != nil {
				t.Fatalf("%q abgewiesen: %v", fall.wert, fehler)
			}

			if !fall.gut && !errors.Is(fehler, domain.ErrValidation) {
				t.Fatalf("%q angenommen oder falsch abgewiesen: %v", fall.wert, fehler)
			}
		})
	}
}

// **Eine Einstellung wie jede andere**: sie steht in der Reihe, gilt für den
// Menschen in jedem seiner Profile und wirkt sofort — und über
// `PruefeEinstellung` erreicht sie dieselbe Prüfung.
func TestZiffernbelegungIstEineEinstellung(t *testing.T) {
	t.Parallel()

	ebene, bekannt := domain.EbeneVon(domain.EinstellungZiffernbelegung)
	if !bekannt || ebene != domain.EbeneNutzer {
		t.Fatalf("Ebene %q (%v), erwartet nutzer", ebene, bekannt)
	}

	if wirkung := domain.EinstellungsWirkung(domain.EinstellungZiffernbelegung); wirkung != domain.WirkungSofort {
		t.Fatalf("Wirkung %q, erwartet sofort", wirkung)
	}

	if fehler := domain.PruefeEinstellung(domain.EinstellungZiffernbelegung, "1=start,1=sets"); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("doppelte Ziffer angenommen: %v", fehler)
	}
}

// **Die Seiten sind die des Fensters** — aufgezählt, sortiert und ohne
// Doppel. Dass es dieselben sind wie in `navigation.ts`, hält
// `internal/werkzeuge/fensterseiten_test.go`.
func TestFensterseitenSindEineMenge(t *testing.T) {
	t.Parallel()

	seiten := domain.Fensterseiten()
	if len(seiten) == 0 {
		t.Fatal("keine Seiten")
	}

	for i := 1; i < len(seiten); i++ {
		if seiten[i-1] >= seiten[i] {
			t.Errorf("nicht sortiert oder doppelt: %q vor %q", seiten[i-1], seiten[i])
		}
	}

	seiten[0] = "verändert"
	if domain.Fensterseiten()[0] == "verändert" {
		t.Error("die Liste ist keine Kopie")
	}
}
