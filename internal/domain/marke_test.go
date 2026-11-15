package domain_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Eine Marke ist ein Schlüssel und kein Satz** —
// #1208.
func TestEineMarkeWirdInNormalformGebracht(t *testing.T) {
	for _, fall := range []struct {
		name, roh, will string
	}{
		{name: "Ränder fallen weg", roh: "  fertig  ", will: "fertig"},
		{name: "kleingeschrieben", roh: "Fertig", will: "fertig"},
		{name: "auch mit Umlaut", roh: "PRÜFEN", will: "prüfen"},
		{name: "innere Leere wird eine", roh: "noch  zu   tun", will: "noch zu tun"},
		{name: "Zeilenumbrüche werden Leerzeichen", roh: "a\nb", will: "a b"},
	} {
		t.Run(fall.name, func(t *testing.T) {
			if haben := domain.Markennormalform(fall.roh); haben != fall.will {
				t.Errorf("Markennormalform(%q) = %q, erwartet %q", fall.roh, haben, fall.will)
			}
		})
	}
}

// **Was keine Marke sein kann, wird abgewiesen — mit Feldbezug** (§8).
func TestEineUnbrauchbareMarkeWirdAbgewiesen(t *testing.T) {
	for _, fall := range []struct {
		name, roh string
	}{
		{name: "leer", roh: "   "},
		{name: "nur Umbrüche", roh: "\n\t"},
		{name: "zu lang", roh: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
	} {
		t.Run(fall.name, func(t *testing.T) {
			_, err := domain.PruefeMarke(fall.roh)
			if !errors.Is(err, domain.ErrValidation) {
				t.Fatalf("PruefeMarke(%q) = %v, erwartet ErrValidation", fall.roh, err)
			}

			var feld *domain.FieldError
			if !errors.As(err, &feld) || feld.Field != "marke" {
				t.Errorf("kein Feldbezug `marke`: %v", err)
			}
		})
	}
}

// **Eine gültige Marke kommt in Normalform zurück.**
func TestEineGueltigeMarkeKommtInNormalform(t *testing.T) {
	marke, err := domain.PruefeMarke("  Noch Zu Tun ")
	if err != nil {
		t.Fatalf("PruefeMarke: %v", err)
	}

	if marke != "noch zu tun" {
		t.Errorf("Marke = %q", marke)
	}
}

// **Die Grenze ist gezählt und nicht geraten.**
func TestDieMarkengrenzeGiltInZeichenUndNichtInBytes(t *testing.T) {
	// 40 Umlaute sind 80 Bytes — und genau 40 Zeichen.
	lang := ""
	for range domain.Markengrenze {
		lang += "ä"
	}

	if _, err := domain.PruefeMarke(lang); err != nil {
		t.Fatalf("%d Zeichen wurden abgewiesen: %v", domain.Markengrenze, err)
	}
}
