package domain_test

// **Die Angaben eines Kontos** — Werkstück 8 von Phase 4.12
// (#1351).

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ein Nutzer mit allen Angaben kommt durch.**
func TestEinNutzerMitAllenAngabenKommtDurch(t *testing.T) {
	t.Parallel()

	if fehler := (domain.Nutzer{
		Name:         "Ich",
		Anrede:       "Frau",
		Namenszusatz: "von der Heide",
		Vorname:      "Ada",
		Weiterenamen: "Augusta",
		Nachname:     "Lovelace",
		Slogan:       "rechnet gern",
		Kurzebio:     "Kurz.",
		Langebio:     "Lang.",
	}).Validate(); fehler != nil {
		t.Errorf("abgewiesen: %v", fehler)
	}
}

// **Und einer ohne ebenso** — jede Angabe ist freiwillig.
//
// **Die Pflicht trägt allein der Anzeigename.** Wer sich anmeldet, muss
// sagen, wie er heißen will, und sonst nichts.
func TestEinNutzerOhneAngabenKommtDurch(t *testing.T) {
	t.Parallel()

	if fehler := (domain.Nutzer{Name: "Ich"}).Validate(); fehler != nil {
		t.Errorf("abgewiesen: %v", fehler)
	}
}

// **Jede Angabe hat ihre Grenze — und die Meldung nennt das Feld** (§6).
func TestJedeAngabeNenntIhrFeld(t *testing.T) {
	t.Parallel()

	zulang := strings.Repeat("x", domain.Hoechstlangebio+1)

	faelle := map[string]struct {
		nutzer domain.Nutzer
		feld   string
	}{
		"Anrede mit Rand": {
			nutzer: domain.Nutzer{Name: "Ich", Anrede: "Frau "},
			feld:   "anrede",
		},
		"Namenszusatz zu lang": {
			nutzer: domain.Nutzer{Name: "Ich", Namenszusatz: zulang},
			feld:   "namenszusatz",
		},
		"Vorname zu lang": {
			nutzer: domain.Nutzer{Name: "Ich", Vorname: zulang},
			feld:   "vorname",
		},
		"weitere Namen zu lang": {
			nutzer: domain.Nutzer{Name: "Ich", Weiterenamen: zulang},
			feld:   "weiterenamen",
		},
		"Nachname zu lang": {
			nutzer: domain.Nutzer{Name: "Ich", Nachname: zulang},
			feld:   "nachname",
		},
		"Slogan zu lang": {
			nutzer: domain.Nutzer{Name: "Ich", Slogan: zulang},
			feld:   "slogan",
		},
		"kurze Biografie zu lang": {
			nutzer: domain.Nutzer{Name: "Ich", Kurzebio: zulang},
			feld:   "kurzebio",
		},
		"lange Biografie zu lang": {
			nutzer: domain.Nutzer{Name: "Ich", Langebio: zulang},
			feld:   "langebio",
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			fehler := fall.nutzer.Validate()
			if !errors.Is(fehler, domain.ErrValidation) {
				t.Fatalf("angenommen wurde: %v", fehler)
			}

			var feld *domain.FieldError
			if !errors.As(fehler, &feld) || feld.Field != fall.feld {
				t.Errorf("gemeldet wurde %v", fehler)
			}
		})
	}
}

// **Die Grenzen zählen Zeichen und nicht Byte.**
//
// **Sonst wäre ein Name mit Umlauten kürzer erlaubt als einer ohne** — und
// niemand sähe, warum.
func TestDieGrenzenZaehlenZeichen(t *testing.T) {
	t.Parallel()

	if fehler := (domain.Nutzer{
		Name: "Ich", Vorname: strings.Repeat("ä", domain.Hoechstnamensteil),
	}).Validate(); fehler != nil {
		t.Errorf("hundert Umlaute wurden abgewiesen: %v", fehler)
	}
}

// **Die Anrede ist freier Text.**
//
// # Warum das eine Zusage ist und keine Nachlässigkeit
//
// **Weil jede Liste jemanden auslässt.** Eine Anwendung, die zwingt, sich für
// einen falschen Eintrag zu entscheiden, hat den Menschen falsch angeredet —
// genau das, was das Feld verhindern soll.
func TestDieAnredeIstFreierText(t *testing.T) {
	t.Parallel()

	for _, anrede := range []string{"Frau", "Herr", "Mx.", "Dr.", "—", "они"} {
		if fehler := (domain.Nutzer{Name: "Ich", Anrede: anrede}).Validate(); fehler != nil {
			t.Errorf("%q wurde abgewiesen: %v", anrede, fehler)
		}
	}
}
