package domain_test

// **Vorgaben, die auf Zeilen zeigen** — Werkstück 15 von Phase 4.12
// (#1351),
// Punkte 4 bis 7 der Kommentierung.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die fünf Arten stehen fest, und zwei davon halten mehrere Werte.**
func TestDieVorgabeartenSindFuenf(t *testing.T) {
	t.Parallel()

	arten := domain.Vorgabearten()
	if len(arten) != 5 {
		t.Fatalf("es sind %d Arten: %v", len(arten), arten)
	}

	mehrwertig := map[domain.Vorgabeart]bool{
		domain.VorgabeFormat:   false,
		domain.VorgabeSpiel:    false,
		domain.VorgabeVorlage:  false,
		domain.VorgabeDeck:     true,
		domain.VorgabeLagerort: true,
	}

	for _, art := range arten {
		soll, bekannt := mehrwertig[art]
		if !bekannt {
			t.Errorf("die Art %q steht in keiner Erwartung", art)

			continue
		}

		if art.Mehrwertig() != soll {
			t.Errorf("%q: mehrwertig = %v, erwartet %v",
				art, art.Mehrwertig(), soll)
		}
	}
}

// **Eine Art, die es nicht gibt, wird abgewiesen.**
func TestEineUnbekannteVorgabeartWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	faelle := map[string]string{
		"erfunden":    "kartenrücken",
		"leer":        "",
		"mit Rändern": " format ",
		"groß":        "Format",
	}

	for name, art := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			fehler := domain.Vorgabeart(art).Pruefe()
			if !errors.Is(fehler, domain.ErrValidation) {
				t.Errorf("%q kam durch: %v", art, fehler)
			}
		})
	}

	// **Der Beleg, dass die Prüfung nicht alles abweist** (§10).
	if fehler := domain.VorgabeFormat.Pruefe(); fehler != nil {
		t.Errorf("die echte Art wurde abgewiesen: %v", fehler)
	}
}

// **Eine einwertige Art nimmt höchstens einen Wert.**
func TestEineEinwertigeArtNimmtNurEinen(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		art     domain.Vorgabeart
		werte   []string
		glueckt bool
	}{
		"Format, keiner":     {domain.VorgabeFormat, nil, true},
		"Format, einer":      {domain.VorgabeFormat, []string{"7"}, true},
		"Format, zwei":       {domain.VorgabeFormat, []string{"7", "8"}, false},
		"Decks, keines":      {domain.VorgabeDeck, nil, true},
		"Decks, drei":        {domain.VorgabeDeck, []string{"1", "2", "3"}, true},
		"Decks, doppelt":     {domain.VorgabeDeck, []string{"1", "1"}, false},
		"Lagerorte, leerer":  {domain.VorgabeLagerort, []string{""}, false},
		"Format, nur Ränder": {domain.VorgabeFormat, []string{"  "}, false},

		// **Und die Art selbst wird geprüft** — sonst käme eine erfundene
		// durch, solange ihre Werte in Ordnung sind. Der Fall steht hier und
		// nicht nur bei [domain.Vorgabeart.Pruefe]: dass `Validate` sie
		// **ruft**, ist eine eigene Zusage.
		"erfundene Art": {"kartenrücken", []string{"1"}, false},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			vorgabe := domain.Profilvorgabe{Art: fall.art, Werte: fall.werte}

			fehler := vorgabe.Validate()
			if fall.glueckt != (fehler == nil) {
				t.Fatalf("glückt = %v, Fehler = %v", fall.glueckt, fehler)
			}
		})
	}
}

// **Ein Wert mit Rändern wird abgewiesen und nicht beschnitten.**
//
// **Dieselbe Regel wie beim Anzeigenamen**: zwei Vorgaben mit demselben
// sichtbaren Wert ließen sich nicht unterscheiden, und wer sich vertippt,
// bekäme eine Vorgabe, die er nicht gesetzt hat.
func TestEinWertMitRaendernWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	vorgabe := domain.Profilvorgabe{
		Art: domain.VorgabeDeck, Werte: []string{"1", " 2"},
	}

	fehler := vorgabe.Validate()
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("der Wert mit Rand kam durch: %v", fehler)
	}

	// **Und der Befund nennt das Feld** (§8).
	var feld *domain.FieldError
	if !errors.As(fehler, &feld) || feld.Field != "werte" {
		t.Errorf("der Befund nennt kein Feld: %v", fehler)
	}
}
