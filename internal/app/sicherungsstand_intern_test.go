package app

// **Der Stand der Sicherung für die Flächen** — OPS01
// (#2347).

import (
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestDerSicherungsstandAlsDTO(t *testing.T) {
	t.Parallel()

	jetzt := time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC)
	vermerk := domain.Sicherungsvermerk{
		Gesichert: jetzt.AddDate(0, 0, -16), Anlass: domain.SicherungVonHand, Flaeche: domain.FlaecheCLI,
		Datei: "hand.msesicherung", Groesse: 9, Pruefsumme: strings.Repeat("a", 64), Verschluesselt: true,
	}
	plan := vermerk
	plan.Anlass, plan.Gesichert = domain.SicherungNachPlan, jetzt.AddDate(0, 0, -8)

	faelle := []struct {
		name    string
		stand   domain.Sicherungsstand
		plan    domain.Sicherungsvermerk
		bekannt bool
		ordner  string
		dto     SicherungsstandDTO
	}{
		{
			"nie gesichert, kein Plan", domain.SicherungsstandBei(domain.Sicherungsvermerk{}, false, jetzt),
			domain.Sicherungsvermerk{},
			false, "",
			SicherungsstandDTO{Stufe: "unbekannt"},
		},
		{
			"alt, ein Plan, der fällig ist", domain.SicherungsstandBei(vermerk, true, jetzt), plan, true, "/Volumes/Stick",
			SicherungsstandDTO{
				Stufe: "alt", Gesichert: "2026-09-29T12:00:00Z", Tage: 16, Anlass: "hand", Flaeche: "cli",
				Datei: "hand.msesicherung", Verschluesselt: true,
				Plan: SicherungsplanDTO{Ordner: "/Volumes/Stick", Letzter: "2026-10-07T12:00:00Z", Faellig: true},
			},
		},
		{
			"ein Plan ohne Vermerk ist fällig", domain.SicherungsstandBei(domain.Sicherungsvermerk{}, false, jetzt),
			domain.Sicherungsvermerk{},
			false, "/Volumes/Stick",
			SicherungsstandDTO{Stufe: "unbekannt", Plan: SicherungsplanDTO{Ordner: "/Volumes/Stick", Faellig: true}},
		},
	}

	for _, fall := range faelle {
		if dto := sicherungsstandAlsDTO(fall.stand, fall.plan, fall.bekannt, fall.ordner, jetzt); dto != fall.dto {
			t.Errorf("%s:\n%+v\nerwartet\n%+v", fall.name, dto, fall.dto)
		}
	}
}
