package domain_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// einVermerk ist ein Vermerk, der jede Prüfung besteht.
func einVermerk() domain.Sicherungsvermerk {
	return domain.Sicherungsvermerk{
		Gesichert:  time.Date(2026, 10, 14, 16, 30, 0, 0, time.UTC),
		Anlass:     domain.SicherungVonHand,
		Flaeche:    domain.FlaecheFenster,
		Datei:      "bestand-2026-10-14.db",
		Groesse:    4096,
		Pruefsumme: strings.Repeat("ab", 32),
	}
}

func TestEinSicherungsvermerkBestehtDiePruefung(t *testing.T) {
	t.Parallel()

	for _, anlass := range []domain.Sicherungsanlass{
		domain.SicherungVonHand, domain.SicherungNachPlan, domain.SicherungVorWanderung,
	} {
		for _, flaeche := range []domain.Flaeche{
			domain.FlaecheFenster, domain.FlaecheCLI, domain.FlaecheMCP, domain.FlaecheREST,
		} {
			vermerk := einVermerk()
			vermerk.Anlass, vermerk.Flaeche = anlass, flaeche

			if fehler := vermerk.Pruefe(); fehler != nil {
				t.Errorf("%s über %s: %v", anlass, flaeche, fehler)
			}
		}
	}
}

func TestEinSicherungsvermerkWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		aendern func(*domain.Sicherungsvermerk)
		feld    string
	}{
		"ohne Zeitpunkt":         {func(v *domain.Sicherungsvermerk) { v.Gesichert = time.Time{} }, "gesichert"},
		"ein unbekannter Anlass": {func(v *domain.Sicherungsvermerk) { v.Anlass = "nebenbei" }, "anlass"},
		"eine unbekannte Fläche": {func(v *domain.Sicherungsvermerk) { v.Flaeche = "fax" }, "flaeche"},
		"ohne Dateinamen":        {func(v *domain.Sicherungsvermerk) { v.Datei = "" }, "datei"},
		// **Nur der Name, kein Ordner**: der Bestand wird weitergegeben, und ein
		// Pfad trüge den Namen des Benutzerkontos mit.
		"ein Pfad statt eines Namens": {func(v *domain.Sicherungsvermerk) { v.Datei = "/Users/jemand/bestand.db" }, "datei"},
		"leer":                        {func(v *domain.Sicherungsvermerk) { v.Groesse = 0 }, "groesse"},
		"eine kurze Prüfsumme":        {func(v *domain.Sicherungsvermerk) { v.Pruefsumme = "abc" }, "pruefsumme"},
		"eine Prüfsumme in Großbuchstaben": {func(v *domain.Sicherungsvermerk) {
			v.Pruefsumme = strings.Repeat("AB", 32)
		}, "pruefsumme"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			vermerk := einVermerk()
			fall.aendern(&vermerk)

			fehler := vermerk.Pruefe()
			if !errors.Is(fehler, domain.ErrValidation) {
				t.Fatalf("ergibt %v, erwartet eine Abweisung", fehler)
			}

			var feldfehler *domain.FieldError
			if !errors.As(fehler, &feldfehler) || feldfehler.Field != fall.feld {
				t.Errorf("abgewiesen an %+v, erwartet am Feld %q", feldfehler, fall.feld)
			}
		})
	}
}
