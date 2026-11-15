package domain_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestOffenlegungPruefe(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		gabe domain.Offenlegung
		gilt bool
	}{
		"gibt an einen frei": {
			gabe: domain.Offenlegung{Von: "anna", An: []string{"ben"}},
			gilt: true,
		},
		"gibt an alle frei": {
			gabe: domain.Offenlegung{
				Von: "anna", An: []string{domain.OffenlegungAnAlle},
			},
			gilt: true,
		},
		"ohne Geber": {
			gabe: domain.Offenlegung{An: []string{"ben"}},
		},
		"ohne Empfänger": {
			gabe: domain.Offenlegung{Von: "anna"},
		},
		"leerer Empfänger": {
			gabe: domain.Offenlegung{Von: "anna", An: []string{""}},
		},
		"an sich selbst": {
			gabe: domain.Offenlegung{Von: "anna", An: []string{"anna"}},
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			fehler := fall.gabe.Pruefe()

			if fall.gilt && fehler != nil {
				t.Fatalf("erwartet gültig, bekam %v", fehler)
			}

			if !fall.gilt && !errors.Is(fehler, domain.ErrValidation) {
				t.Fatalf("erwartet ErrValidation, bekam %v", fehler)
			}
		})
	}
}

func TestMitleser(t *testing.T) {
	t.Parallel()

	freigaben := []domain.Offenlegung{
		{Von: "anna", An: []string{"ben"}},
		{Von: "cem", An: []string{domain.OffenlegungAnAlle}},
	}

	faelle := map[string]struct {
		fuer     string
		erwartet []string
	}{
		"sich selbst sieht jeder":   {fuer: "dora", erwartet: []string{"dora", "cem"}},
		"der benannte Empfänger":    {fuer: "ben", erwartet: []string{"ben", "anna", "cem"}},
		"ein Protokoll ohne Angabe": {fuer: "", erwartet: []string{"cem"}},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			darf := domain.Mitleser(freigaben, fall.fuer)

			if len(darf) != len(fall.erwartet) {
				t.Fatalf("erwartet %v, bekam %v", fall.erwartet, darf)
			}

			for _, wer := range fall.erwartet {
				if !darf[wer] {
					t.Fatalf("%q fehlt in %v", wer, darf)
				}
			}
		})
	}
}

func TestRueckschauMitFreigabe(t *testing.T) {
	t.Parallel()

	strom := []domain.Ereignis{{
		PartieID: "p", Folge: 1, Art: "ziehen",
		Oeffentlich: map[string]any{"sitz": "anna"},
		Privat: []domain.Ergaenzung{
			{Empfaenger: "anna", Inhalt: map[string]any{"karte": "Wald"}},
			{Empfaenger: "ben", Inhalt: map[string]any{"karte": "Insel"}},
		},
		Kette: "abc",
	}}

	t.Run("ohne Freigabe bleibt es beim eigenen", func(t *testing.T) {
		t.Parallel()

		gesehen := domain.RueckschauFuer(strom, "ben", nil)

		if len(gesehen[0].Privat) != 1 ||
			gesehen[0].Privat[0].Empfaenger != "ben" {
			t.Fatalf("erwartet nur die eigene Ergänzung, bekam %v",
				gesehen[0].Privat)
		}
	})

	t.Run("mit Freigabe kommt die fremde dazu", func(t *testing.T) {
		t.Parallel()

		gesehen := domain.RueckschauFuer(strom, "ben",
			[]domain.Offenlegung{{Von: "anna", An: []string{"ben"}}})

		if len(gesehen[0].Privat) != 2 {
			t.Fatalf("erwartet beide Ergänzungen, bekam %v", gesehen[0].Privat)
		}
	})

	t.Run("die Reihenfolge des Stroms bleibt", func(t *testing.T) {
		t.Parallel()

		gesehen := domain.RueckschauFuer(strom, "ben",
			[]domain.Offenlegung{{Von: "anna", An: []string{"ben"}}})

		if gesehen[0].Privat[0].Empfaenger != "anna" {
			t.Fatalf("erwartet die Reihenfolge des Ereignisses, bekam %v",
				gesehen[0].Privat)
		}
	})

	t.Run("eine fremde Freigabe hilft niemand anderem", func(t *testing.T) {
		t.Parallel()

		gesehen := domain.RueckschauFuer(strom, "cem",
			[]domain.Offenlegung{{Von: "anna", An: []string{"ben"}}})

		if len(gesehen[0].Privat) != 0 {
			t.Fatalf("erwartet nichts, bekam %v", gesehen[0].Privat)
		}
	})

	t.Run("der Strom selbst bleibt unberührt", func(t *testing.T) {
		t.Parallel()

		domain.RueckschauFuer(strom, "ben", nil)

		if len(strom[0].Privat) != 2 {
			t.Fatalf("der Strom wurde gekürzt: %v", strom[0].Privat)
		}
	})
}
