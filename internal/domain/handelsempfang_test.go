package domain_test

// **Was eine empfangene Revision behaupten darf** — TRADE-03.AK-01 und
// NAT-01 (#1717).
//
// Eine Revision kam so an, wie der Absender sie schrieb — mit **seinem**
// Zustand und **seinen** Stimmen. Wer `angenommen` mit zwei Stimmen schickte,
// stand hier angenommen, ohne dass jemand zugestimmt hätte.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEineErsteRevisionKommtVomInitiatorUndIstAngeboten(t *testing.T) {
	t.Parallel()

	gut := einHandelsstand(domain.HandelAngeboten)
	gut.Zustimmungen = []string{"banna"}

	if fehler := domain.PruefeErsteRevision(gut, "banna"); fehler != nil {
		t.Fatalf("eine gültige erste Revision: %v", fehler)
	}

	ohneStimmen := einHandelsstand(domain.HandelAngeboten)
	if fehler := domain.PruefeErsteRevision(ohneStimmen, "banna"); fehler != nil {
		t.Errorf("ohne Stimmen: %v", fehler)
	}

	for _, fall := range []struct {
		name     string
		aendere  func(*domain.Handelsangebot)
		absender string
	}{
		{"nicht vom Initiator", func(*domain.Handelsangebot) {}, "bbert"},
		{"schon verbucht", func(a *domain.Handelsangebot) { a.Zustand = domain.HandelLokalVerbucht }, "banna"},
		{"als Entwurf", func(a *domain.Handelsangebot) { a.Zustand = domain.HandelEntwurf }, "banna"},
		{"mit einer fremden Stimme", func(a *domain.Handelsangebot) { a.Zustimmungen = []string{"banna", "bbert"} }, "banna"},
		{"ohne Position", func(a *domain.Handelsangebot) { a.Positionen = nil }, "banna"},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			angebot := einHandelsstand(domain.HandelAngeboten)
			fall.aendere(&angebot)

			if fehler := domain.PruefeErsteRevision(angebot, fall.absender); !errors.Is(fehler, domain.ErrValidation) {
				t.Errorf("Fehler ist %v", fehler)
			}
		})
	}
}

func TestEinEmpfangenerVorschlagIstVerhandlungUndTraegtNurDieEigeneStimme(t *testing.T) {
	t.Parallel()

	hier := einHandelsstand(domain.HandelAngeboten)

	gut := hier.Fortgeschrieben(hier.Positionen)
	gut.Zustimmungen = []string{"bbert"}

	if fehler := hier.NimmtVon(gut, "bbert"); fehler != nil {
		t.Fatalf("ein gültiger Vorschlag: %v", fehler)
	}

	angenommen := hier.Fortgeschrieben(hier.Positionen)
	angenommen.Zustand = domain.HandelAngenommen

	fremdeStimme := hier.Fortgeschrieben(hier.Positionen)
	fremdeStimme.Zustimmungen = []string{"banna"}

	for name, vorschlag := range map[string]domain.Handelsangebot{
		"mit behauptetem Zustand": angenommen,
		"mit fremder Stimme":      fremdeStimme,
	} {
		if fehler := hier.NimmtVon(vorschlag, "bbert"); !errors.Is(fehler, domain.ErrValidation) {
			t.Errorf("%s: Fehler ist %v", name, fehler)
		}
	}

	// **Was `Nimmt` abweist, weist auch `NimmtVon` ab.**
	falscheReihe := hier.Fortgeschrieben(hier.Positionen)
	falscheReihe.Revision = 7

	if fehler := hier.NimmtVon(falscheReihe, "bbert"); !errors.Is(fehler, domain.ErrConflict) {
		t.Errorf("eine falsche Revision: %v", fehler)
	}
}

// **Eine verspätete Stimme gehört zu ihrem Beschluss** — Befund aus der
// Kartierung zu TRADE-03: eine Zustimmung zum Angebot, die erst ankommt, wenn
// hier schon vorbereitet ist, zählte als Stimme zur Übergabe.
func TestEineVerspaeteteStimmeZaehltNichtFuerDenNaechstenBeschluss(t *testing.T) {
	t.Parallel()

	hier := einHandelsstand(domain.HandelVorbereitet)

	drueben := einHandelsstand(domain.HandelAngeboten)
	drueben.Zustimmungen = []string{"bbert"}

	danach, fehler := hier.NimmtStand(drueben, "bbert")
	if fehler != nil {
		t.Fatalf("NimmtStand: %v", fehler)
	}

	if danach.Zustand != domain.HandelVorbereitet || len(danach.Zustimmungen) != 0 {
		t.Errorf("danach %q mit %v", danach.Zustand, danach.Zustimmungen)
	}
}

// **Was hinausgeht** — eine ungeteilte Bewertung bleibt hier, und der
// Abdruck ändert sich dabei nicht; eine geteilte geht mit.
func TestZumVersandBleibtEineUngeteilteBewertungHier(t *testing.T) {
	t.Parallel()

	hier := einHandelsstand(domain.HandelAngeboten)
	hier.Positionen[0].Bewertung = 1995

	hinaus := hier.ZumVersand()
	if hinaus.Positionen[0].Bewertung != 0 || hinaus.Abdruck() != hier.Abdruck() || hier.Positionen[0].Bewertung != 1995 {
		t.Errorf("ungeteilt: hinaus %d, hier %d", hinaus.Positionen[0].Bewertung, hier.Positionen[0].Bewertung)
	}

	hier.BewertungGeteilt = true
	if hinaus = hier.ZumVersand(); hinaus.Positionen[0].Bewertung != 1995 {
		t.Errorf("geteilt: hinaus %d", hinaus.Positionen[0].Bewertung)
	}

	for art, bewegt := range map[domain.Handelsart]bool{
		domain.HandelTausch: true, domain.HandelTeilen: false, domain.HandelDeckleihe: false, domain.HandelMarkt: false,
	} {
		if art.BewegtBestand() != bewegt {
			t.Errorf("%s bewegt Bestand: %v", art, !bewegt)
		}
	}
}
