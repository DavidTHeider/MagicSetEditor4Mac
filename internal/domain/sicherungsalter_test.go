package domain_test

import (
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestDasAlterEinerSicherung(t *testing.T) {
	t.Parallel()

	gesichert := einVermerk().Gesichert

	faelle := []struct {
		name    string
		bekannt bool
		jetzt   time.Time
		stufe   domain.Sicherungsstufe
		tage    int
	}{
		{"ohne Vermerk heißt unbekannt", false, gesichert, domain.SicherungsstufeUnbekannt, 0},
		{"im selben Augenblick ist frisch", true, gesichert, domain.SicherungsstufeFrisch, 0},
		{
			"ein Tag weniger eine Sekunde zählt null Tage", true,
			gesichert.Add(24*time.Hour - time.Second), domain.SicherungsstufeFrisch, 0,
		},
		{
			"genau vierzehn Tage sind noch frisch", true,
			gesichert.Add(14 * 24 * time.Hour), domain.SicherungsstufeFrisch, 14,
		},
		{
			"eine Sekunde über vierzehn Tage ist alt", true,
			gesichert.Add(14*24*time.Hour + time.Second), domain.SicherungsstufeAlt, 14,
		},
		{
			"sechzehn Tage sind alt", true,
			gesichert.Add(16*24*time.Hour + time.Hour), domain.SicherungsstufeAlt, 16,
		},
		{
			"eine Minute nach der Uhr ist noch frisch", true,
			gesichert.Add(-time.Minute), domain.SicherungsstufeFrisch, 0,
		},
		{
			"mehr als eine Minute nach der Uhr ist unstimmig", true,
			gesichert.Add(-time.Minute - time.Second), domain.SicherungsstufeUnstimmig, 0,
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			stand := domain.SicherungsstandBei(einVermerk(), fall.bekannt, fall.jetzt)

			if stand.Stufe != fall.stufe || stand.Tage != fall.tage {
				t.Errorf("Stufe %q mit %d Tagen, erwartet %q mit %d", stand.Stufe, stand.Tage, fall.stufe, fall.tage)
			}

			if fall.bekannt != (stand.Letzter == einVermerk()) {
				t.Errorf("der letzte Vermerk ist %+v — bekannt: %v", stand.Letzter, fall.bekannt)
			}
		})
	}
}

func TestDerPlanIstFaellig(t *testing.T) {
	t.Parallel()

	gesichert := einVermerk().Gesichert

	faelle := []struct {
		name    string
		bekannt bool
		jetzt   time.Time
		faellig bool
	}{
		{"ohne Planvermerk ist er fällig", false, gesichert, true},
		{"nach einem Tag nicht", true, gesichert.Add(24 * time.Hour), false},
		{"eine Sekunde vor der Woche nicht", true, gesichert.Add(7*24*time.Hour - time.Second), false},
		{"nach einer Woche ist er fällig", true, gesichert.Add(7 * 24 * time.Hour), true},
		{"ein Vermerk nach der Uhr hält ihn nicht auf", true, gesichert.Add(-time.Hour), true},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			if faellig := domain.SicherungsplanFaellig(einVermerk(), fall.bekannt, fall.jetzt); faellig != fall.faellig {
				t.Errorf("fällig: %v, erwartet %v", faellig, fall.faellig)
			}
		})
	}
}

func TestDerPlanBehaeltSiebenEigene(t *testing.T) {
	t.Parallel()

	tage := func(namen ...int) []domain.Sicherungsvermerk {
		vermerke := make([]domain.Sicherungsvermerk, 0, len(namen))
		for _, tag := range namen {
			vermerk := einVermerk()
			vermerk.Anlass = domain.SicherungNachPlan
			vermerk.Gesichert = vermerk.Gesichert.AddDate(0, 0, tag)
			vermerk.Datei = vermerk.Gesichert.Format("plan-2006-01-02")
			vermerke = append(vermerke, vermerk)
		}

		return vermerke
	}

	faelle := []struct {
		name   string
		eigene []domain.Sicherungsvermerk
		weg    []string
	}{
		{"keine eigenen", nil, nil},
		{"sieben bleiben alle", tage(0, 7, 14, 21, 28, 35, 42), nil},
		{"die älteste von acht geht", tage(7, 14, 21, 28, 35, 42, 49, 0), []string{"plan-2026-10-14"}},
		{
			"von zehn gehen die drei ältesten, gleich in welcher Reihenfolge sie kommen",
			tage(63, 0, 56, 7, 49, 14, 42, 21, 35, 28),
			[]string{"plan-2026-10-28", "plan-2026-10-21", "plan-2026-10-14"},
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			weg := domain.SicherungsplanAusduennen(fall.eigene)

			namen := make([]string, 0, len(weg))
			for _, vermerk := range weg {
				namen = append(namen, vermerk.Datei)
			}

			if len(namen) != len(fall.weg) {
				t.Fatalf("weg: %v, erwartet %v", namen, fall.weg)
			}

			for i := range namen {
				if namen[i] != fall.weg[i] {
					t.Errorf("weg: %v, erwartet %v", namen, fall.weg)
				}
			}
		})
	}
}
