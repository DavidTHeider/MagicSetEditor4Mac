package domain_test

// **Der zweite Faktor in der Domäne** — Werkstück 7 von Phase 4.12
// (#1351).

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ein tauglicher Faktor kommt durch.**
func TestEinTauglicherFaktorKommtDurch(t *testing.T) {
	t.Parallel()

	for _, art := range domain.Faktorarten() {
		if fehler := (domain.Faktor{
			NutzerID: 1, Art: art, Bezeichnung: "iPhone",
		}).Validate(); fehler != nil {
			t.Errorf("%s: %v", art, fehler)
		}
	}
}

// **Und ein untauglicher nicht — jeder mit seinem Feld.**
//
// **Das Feld ist die Hälfte der Auskunft** (§6): „ungültig" ohne Feld ließe
// offen, was zu ändern ist.
func TestEinUntauglicherFaktorNenntSeinFeld(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		faktor domain.Faktor
		feld   string
	}{
		"ohne Nutzer": {
			faktor: domain.Faktor{Art: domain.FaktorTOTP},
			feld:   "nutzerID",
		},
		"unbekannte Art": {
			faktor: domain.Faktor{NutzerID: 1, Art: "sms"},
			feld:   "art",
		},
		"leere Art": {
			faktor: domain.Faktor{NutzerID: 1},
			feld:   "art",
		},
		"Bezeichnung mit Rand": {
			faktor: domain.Faktor{
				NutzerID: 1, Art: domain.FaktorTOTP, Bezeichnung: "iPhone ",
			},
			feld: "bezeichnung",
		},
		"zu lange Bezeichnung": {
			faktor: domain.Faktor{
				NutzerID: 1, Art: domain.FaktorTOTP,
				Bezeichnung: strings.Repeat("x", domain.Hoechstbezeichnung+1),
			},
			feld: "bezeichnung",
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			fehler := fall.faktor.Validate()
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

// **Die Bezeichnung wird in Zeichen gezählt und nicht in Byte.**
//
// **Sonst wäre ein Name mit Umlauten kürzer erlaubt als einer ohne** — und
// niemand sähe, warum.
func TestDieBezeichnungWirdInZeichenGezaehlt(t *testing.T) {
	t.Parallel()

	if fehler := (domain.Faktor{
		NutzerID: 1, Art: domain.FaktorTOTP,
		Bezeichnung: strings.Repeat("ä", domain.Hoechstbezeichnung),
	}).Validate(); fehler != nil {
		t.Errorf("hundert Umlaute wurden abgewiesen: %v", fehler)
	}
}

// **Ein fehlender Faktor ist ein eigener Fehler und keine Abweisung.**
//
// **Weil die Fläche etwas anderes tun muss** (§8): „noch einmal tippen" ist
// etwas anderes als „frag nach dem Code" — und die **Art** reist mit, damit
// sie das richtige Feld anbietet.
func TestEinFehlenderFaktorIstEinEigenerFehler(t *testing.T) {
	t.Parallel()

	fehler := domain.FehlenderFaktor(domain.FaktorBiometrie)

	if !errors.Is(fehler, domain.ErrZweiterFaktor) {
		t.Fatalf("gemeldet wurde: %v", fehler)
	}

	// **Und ausdrücklich keine Validierung**: sonst führe die Fläche
	// denselben Weg wie bei einem Tippfehler.
	if errors.Is(fehler, domain.ErrValidation) {
		t.Error("er meldet sich als Validierung")
	}

	var feld *domain.FieldError
	if !errors.As(fehler, &feld) ||
		feld.Werte["art"] != string(domain.FaktorBiometrie) {
		t.Errorf("gemeldet wurde %+v", fehler)
	}
}

// **Ein Feldbefund mit `nichtGefunden` meldet sich als [domain.ErrNotFound].**
//
// # Warum das eigens geprüft wird
//
// **Weil es dieselbe Klasse ist wie beim Konflikt** — und derselbe Fehler:
// ein Feldbefund mit `konflikt` meldete sich einst als Validierung, und
// `errors.Is(err, ErrConflict)` war rot. Gefunden am Export
// personenbezogener Daten (Werkstück 8, #1351).
func TestEinFeldbefundMitNichtGefundenMeldetSichSo(t *testing.T) {
	t.Parallel()

	fehler := domain.NewFieldError("nutzerID", domain.SchluesselNichtGefunden,
		"Diesen Nutzer führt der Bestand nicht.")

	if !errors.Is(fehler, domain.ErrNotFound) {
		t.Fatalf("gemeldet wurde: %v", fehler)
	}

	// **Und ausdrücklich nicht als Validierung**: sonst zeigte die Fläche
	// dasselbe Formular noch einmal, statt zu sagen, dass es das nicht gibt.
	if errors.Is(fehler, domain.ErrValidation) {
		t.Error("er meldet sich als Validierung")
	}
}
