package service_test

// **Wann ein Zugriffstoken nicht gilt** — Werkstück 10 von Phase 4.12
// (#1351).
//
// **Die Fehlerwege stehen eigens hier**, weil sie eine eigene Frage
// beantworten: nicht „was kann ein Token", sondern „was sagt die Anwendung
// dem, der keines hat" (§2).

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// abgewiesen ist der eine Satz, den jeder Fehlschlag bekommt.
const abgewiesen = "Validierung fehlgeschlagen: dieses Token gilt nicht"

func TestJederFehlschlagKlingtGleich(t *testing.T) {
	t.Parallel()

	// **Das ist die Zusage, für die es diese Datei gibt.** Jede Unterscheidung
	// wäre eine Auskunft an den, der rät: „diese Kennung gibt es nicht" sagte
	// ihm, welche es gibt; „abgelaufen" sagte ihm, dass er beinahe richtig
	// lag. Geprüft wird deshalb nicht, *dass* abgewiesen wird, sondern dass
	// **derselbe** Satz herauskommt.
	verwaltung := service.NeueTokenverwaltung(neuesTokendoppel())

	_, echt, fehler := verwaltung.Anlegen(context.Background(), einToken())
	if fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	faelle := []struct {
		name  string
		gabe  string
		grund string
	}{
		{"leer", "", "gar nichts"},
		{"ohne Marke", "1_ABCDEFGH", "zwei Teile statt drei"},
		{"fremde Marke", "npm_1_ABCDEFGH", "eine Marke, die nicht unsere ist"},
		{"Kennung ist Text", "mse_eins_ABCDEFGH", "keine Zahl an zweiter Stelle"},
		{"Kennung ist null", "mse_0_ABCDEFGH", "eine Kennung, die es nie gibt"},
		{"Kennung ist negativ", "mse_-1_ABCDEFGH", "dito, mit Vorzeichen"},
		{"Kennung unbekannt", "mse_99_ABCDEFGH", "eine Zeile, die es nicht gibt"},
		{"Geheimnis falsch", "mse_1_ABCDEFGH", "die richtige Kennung, das falsche Wort"},
		{"Geheimnis leer", "mse_1_", "die richtige Kennung, gar kein Wort"},
		{"Marke allein", "mse", "ein Wort ohne Trenner"},
		{"vier Teile", "mse_1_A_B", "ein Trenner zu viel im Geheimnis"},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			_, fehler := verwaltung.Pruefe(
				context.Background(), fall.gabe, domain.ScopeLesen)
			if fehler == nil {
				t.Fatalf("%q (%s) wurde angenommen", fall.gabe, fall.grund)
			}

			if fehler.Error() != abgewiesen {
				t.Errorf("Meldung = %q, erwartet %q", fehler, abgewiesen)
			}

			if !errors.Is(fehler, domain.ErrValidation) {
				t.Errorf("Fehler = %v, erwartet ErrValidation", fehler)
			}
		})
	}

	t.Run("und das echte gilt", func(t *testing.T) {
		// **Der Beleg, dass die Fälle darüber überhaupt etwas messen.** Ohne
		// ihn hieße eine Reihe von Ablehnungen auch dann alles, wenn `Pruefe`
		// grundsätzlich ablehnte (§10).
		if _, fehler := verwaltung.Pruefe(
			context.Background(), echt, domain.ScopeLesen); fehler != nil {
			t.Fatalf("das echte Token wurde abgewiesen: %v", fehler)
		}
	})
}

func TestEinAbgelaufenesTokenGiltNicht(t *testing.T) {
	t.Parallel()

	gestern := time.Date(2026, 9, 4, 12, 0, 0, 0, time.UTC)
	heute := time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC)

	faelle := []struct {
		name   string
		ablauf string
		gilt   bool
	}{
		{"morgen", heute.Add(24 * time.Hour).Format(time.RFC3339), true},
		{"gestern", gestern.Format(time.RFC3339), false},
		{"jetzt gerade", heute.Format(time.RFC3339), false},
		{"gar kein Ablauf", "", true},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			verwaltung := service.NeueTokenverwaltung(neuesTokendoppel()).
				MitUhr(func() time.Time { return heute })

			neu := einToken()
			neu.Ablauf = fall.ablauf

			_, klartext, fehler := verwaltung.Anlegen(context.Background(), neu)
			if fehler != nil {
				t.Fatalf("Anlegen: %v", fehler)
			}

			_, fehler = verwaltung.Pruefe(
				context.Background(), klartext, domain.ScopeLesen)
			if fall.gilt != (fehler == nil) {
				t.Fatalf("gilt = %v, Fehler = %v", fall.gilt, fehler)
			}

			if fehler != nil && fehler.Error() != abgewiesen {
				t.Errorf("Meldung = %q, erwartet %q", fehler, abgewiesen)
			}
		})
	}
}

func TestEinStolpernderBestandWirdWeitergereicht(t *testing.T) {
	t.Parallel()

	// **Weil ein Bestandsfehler kein Validierungsfehler ist.** Wer ihn in
	// „dieses Token gilt nicht" übersetzte, sagte jemandem, sein Token sei
	// falsch, während die Platte voll ist.
	stolpern := errors.New("die Platte ist voll")

	wege := map[string]func(*service.Tokenverwaltung) error{
		"Anlegen": func(v *service.Tokenverwaltung) error {
			_, _, fehler := v.Anlegen(context.Background(), einToken())

			return fehler
		},
		"Alle": func(v *service.Tokenverwaltung) error {
			_, fehler := v.Alle(context.Background(), 7)

			return fehler
		},
		"Aendern": func(v *service.Tokenverwaltung) error {
			_, fehler := v.Aendern(context.Background(), einToken())

			return fehler
		},
		"Loeschen": func(v *service.Tokenverwaltung) error {
			return v.Loeschen(context.Background(), 1)
		},
	}

	for name, weg := range wege {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ablage := neuesTokendoppel()
			ablage.kaputt = stolpern

			fehler := weg(service.NeueTokenverwaltung(ablage))
			if !errors.Is(fehler, stolpern) {
				t.Fatalf("Fehler = %v, erwartet %v", fehler, stolpern)
			}
		})
	}
}

func TestEinStolpernderBestandLaesstNiemandenHerein(t *testing.T) {
	t.Parallel()

	// **`Pruefe` ist die Ausnahme von der Zeile darüber**, und das ist Absicht:
	// hier steht ein Fremder davor. Ihm den Bestandsfehler zu zeigen, hieße
	// ihm zu sagen, dass es seine Kennung gibt.
	ablage := neuesTokendoppel()
	verwaltung := service.NeueTokenverwaltung(ablage)

	_, klartext, fehler := verwaltung.Anlegen(context.Background(), einToken())
	if fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	ablage.kaputt = errors.New("die Platte ist voll")

	_, fehler = verwaltung.Pruefe(
		context.Background(), klartext, domain.ScopeLesen)
	if fehler == nil {
		t.Fatal("der stolpernde Bestand hat jemanden hereingelassen")
	}

	if fehler.Error() != abgewiesen {
		t.Errorf("Meldung = %q, erwartet %q", fehler, abgewiesen)
	}
}
