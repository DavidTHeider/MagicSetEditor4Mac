package domain_test

// **Die Kennungen eines Kontos** — Werkstück 6 von Phase 4.12
// (#1351).

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Eine Adresse braucht ein @ und etwas auf beiden Seiten.**
//
// # Warum so wenig geprüft wird
//
// **Weil mehr falsch wäre.** Die gültigen Adressen sind in RFC 5322 so weit
// gefasst, dass jede kürzere Regel echte Adressen abweist — und wer eine
// abgewiesen bekommt, die sein Postfach hat, hält das Programm für kaputt.
// Was hier steht, fängt den **Tippfehler** und nicht die Grammatik.
func TestEineAdresseBrauchtEinAtUndBeideSeiten(t *testing.T) {
	t.Parallel()

	gut := []string{
		"wer@example.com",
		"wer+turnier@example.co.uk",
		"w@e.d",
		"seltsam!aber#erlaubt@example.org",
	}

	for _, adresse := range gut {
		t.Run(adresse, func(t *testing.T) {
			t.Parallel()

			if err := domain.PruefeEmail(adresse); err != nil {
				t.Errorf("abgewiesen: %v", err)
			}
		})
	}

	schlecht := map[string]string{
		"leer":            "",
		"ohne At":         "wer.example.com",
		"nichts davor":    "@example.com",
		"nichts dahinter": "wer@",
		"zwei At":         "wer@da@example.com",
		"Leerzeichen":     "wer @example.com",
		"zu lang":         strings.Repeat("x", 300) + "@example.com",
	}

	for name, adresse := range schlecht {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if err := domain.PruefeEmail(adresse); !errors.Is(err, domain.ErrValidation) {
				t.Errorf("Fehler %v, erwartet war ErrValidation", err)
			}
		})
	}
}

// **Eine Nummer trägt Ziffern und darf ein Plus vornedran haben.**
//
// **In E.164-Gestalt**, aber ohne Länderprüfung: welche Vorwahlen es gibt,
// ändert sich, und eine Liste im Programm wäre am Tag ihrer Aufnahme veraltet.
func TestEineNummerTraegtZiffernUndDarfEinPlusHaben(t *testing.T) {
	t.Parallel()

	gut := []string{"+4915112345678", "+12025550123", "015112345678"}

	for _, nummer := range gut {
		t.Run(nummer, func(t *testing.T) {
			t.Parallel()

			if err := domain.PruefeTelefon(nummer); err != nil {
				t.Errorf("abgewiesen: %v", err)
			}
		})
	}

	schlecht := map[string]string{
		"leer":            "",
		"Buchstaben":      "+49 ruf mich an",
		"zu kurz":         "+49",
		"zu lang":         "+" + strings.Repeat("1", 20),
		"Plus in der Mit": "049+151",
	}

	for name, nummer := range schlecht {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if err := domain.PruefeTelefon(nummer); !errors.Is(err, domain.ErrValidation) {
				t.Errorf("Fehler %v, erwartet war ErrValidation", err)
			}
		})
	}
}

// **Leerraum am Rand wird abgewiesen, nicht beschnitten.**
//
// Dieselbe Regel wie beim Anzeigenamen (Werkstück 1): zwei Adressen mit
// demselben sichtbaren Text ließen sich nicht unterscheiden.
func TestLeerraumAmRandWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	if err := domain.PruefeEmail(" wer@example.com"); !errors.Is(
		err, domain.ErrValidation) {
		t.Errorf("E-Mail: %v", err)
	}

	if err := domain.PruefeTelefon("+4915112345678 "); !errors.Is(
		err, domain.ErrValidation) {
		t.Errorf("Telefon: %v", err)
	}
}

// **Ein Nutzer mit krummen Kennungen wird abgewiesen.**
//
// **Und mit leeren nicht**: leer heißt „hat er nicht", und das ist der
// Normalfall — der erste Nutzer hat weder Adresse noch Nummer.
func TestEinNutzerMitKrummenKennungenWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	faelle := map[string]domain.Nutzer{
		"krumme E-Mail":    {Name: "Wer", Email: "keine Adresse"},
		"krumme Zweitmail": {Name: "Wer", Zweitemail: "auch keine"},
		"krumme Nummer":    {Name: "Wer", Telefon: "ruf mich an"},
	}

	for name, nutzer := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if err := nutzer.Validate(); !errors.Is(err, domain.ErrValidation) {
				t.Errorf("Fehler %v, erwartet war ErrValidation", err)
			}
		})
	}

	// **Und der Normalfall geht durch.**
	if err := (domain.Nutzer{Name: "Wer"}).Validate(); err != nil {
		t.Errorf("der Nutzer ohne Kennungen wurde abgewiesen: %v", err)
	}

	if err := (domain.Nutzer{
		Name: "Wer", Email: "wer@example.com", Telefon: "+4915112345678",
	}).Validate(); err != nil {
		t.Errorf("der Nutzer mit Kennungen wurde abgewiesen: %v", err)
	}
}

// **Die Meldung nennt das Feld** (§6).
//
// **Sonst stünde bei zwei Adressen zweimal derselbe Satz**, und niemand wüsste,
// welche gemeint ist.
func TestDieMeldungNenntDasFeld(t *testing.T) {
	t.Parallel()

	fehler := (domain.Nutzer{Name: "Wer", Zweitemail: "keine Adresse"}).Validate()
	if fehler == nil {
		t.Fatal("die krumme Zweitadresse kam durch")
	}

	if !strings.Contains(fehler.Error(), "zweitemail") {
		t.Errorf("die Meldung nennt das Feld nicht: %v", fehler)
	}
}
