package app

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
)

// **Die Übersetzung steht im Composition Root und wird dort geprüft.**
//
// Die Fassade nimmt ein Wort entgegen; was daraus wird, entscheidet sich hier.
// Ein Test über die Fassade prüfte das Doppel, nicht die Umrechnung — genau die
// Lücke, vor der AGENTS.md §3 warnt („eine Vorrichtung, die großzügiger ist als
// der Betrieb").
func TestBildartAus(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		wort     string
		erwartet scryfall.Bildart
		gueltig  bool
	}{
		"klein":              {"klein", scryfall.BildKlein, true},
		"normal":             {"normal", scryfall.BildNormal, true},
		"ausschnitt":         {"ausschnitt", scryfall.BildAusschnitt, true},
		"unbekanntes Wort":   {"riesig", "", false},
		"leeres Wort":        {"", "", false},
		"englischer Name":    {"art_crop", "", false},
		"falsche Schreibung": {"Normal", "", false},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			bekommen, gueltig := bildartAus(fall.wort)
			if bekommen != fall.erwartet || gueltig != fall.gueltig {
				t.Errorf("bildartAus(%q) = %q, %v — erwartet %q, %v",
					fall.wort, bekommen, gueltig, fall.erwartet, fall.gueltig)
			}
		})
	}
}

// **Und der Adapter weist ab, bevor er das Netz anfasst.**
func TestBildwegWeistEineUnbekannteArtAb(t *testing.T) {
	t.Parallel()

	// Ein `nil`-Weg genügt: wird er gerufen, stürzt der Test ab — und das ist
	// die Aussage. Die Wache steht **vor** dem Weg.
	_, fehler := bildweg{weg: nil}.Holen(probekontext(t), "Lightning Bolt", "riesig")
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("Fehler = %v, erwartet ErrValidation", fehler)
	}
}

// **AlsFehler macht aus einem Befund wieder einen `error`.**
//
// # Der Grund, aus dem es kein `Error()` auf [Fehler] ist
//
// Ein `nil *Fehler` in einer `error`-Schnittstelle ist **nicht** nil. Der
// erste Fall hier belegt, dass diese Falle nicht zuschnappt.
func TestAlsFehler(t *testing.T) {
	t.Parallel()

	t.Run("nil bleibt nil", func(t *testing.T) {
		t.Parallel()

		if fehler := AlsFehler(nil); fehler != nil {
			t.Errorf("aus nil wurde %v", fehler)
		}
	})

	t.Run("führt die Gattung mit", func(t *testing.T) {
		t.Parallel()

		faelle := map[Fehlercode]error{
			CodeValidierung:   domain.ErrValidation,
			CodeNichtGefunden: domain.ErrNotFound,
			CodeKonflikt:      domain.ErrConflict,
			CodeBelegt:        domain.ErrBelegt,
		}

		for code, gattung := range faelle {
			bekommen := AlsFehler(&Fehler{Code: code, Meldung: "etwas"})
			if !errors.Is(bekommen, gattung) {
				t.Errorf("%q wurde zu %v, erwartet %v", code, bekommen, gattung)
			}
		}
	})

	t.Run("ein unbekannter Code bleibt ein schlichter Fehler", func(t *testing.T) {
		t.Parallel()

		bekommen := AlsFehler(&Fehler{Code: CodeUnbekannt, Meldung: "etwas ging schief"})
		if bekommen == nil || bekommen.Error() != "etwas ging schief" {
			t.Errorf("bekommen = %v", bekommen)
		}
	})

	t.Run("der Schlüssel steht dabei, wenn die Meldung nichts sagt", func(t *testing.T) {
		t.Parallel()

		bekommen := AlsFehler(&Fehler{
			Code: CodeUnbekannt, Schluessel: "unbekannt",
			Meldung: "Es ist ein unerwarteter Fehler aufgetreten.",
		})

		if bekommen == nil || !strings.Contains(bekommen.Error(), "unbekannt") {
			t.Errorf("bekommen = %v — ohne den Schlüssel benennt der Satz nichts", bekommen)
		}
	})
}
