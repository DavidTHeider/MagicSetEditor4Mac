package domain_test

// **Wo ein Handel für mich steht** — TRADE-01.AK-01/02
// (#1715).

import (
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// lage baut einen Vorgang zwischen `aich` (Initiator) und `bdu`.
func lage(zustand domain.Handelszustand, stimmen ...string) domain.Handelsangebot {
	return domain.Handelsangebot{
		Vertragsfassung: 1, Vorgang: "v-1", Revision: 1,
		Art: domain.HandelTausch, Initiator: "aich",
		Parteien: []string{"aich", "bdu"}, Zustand: zustand, Zustimmungen: stimmen,
	}
}

func TestDieLageEinesHandels(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		name     string
		angebot  domain.Handelsangebot
		ich      string
		erwartet domain.Handelslage
		schritt  domain.Handelsschritt
	}{
		{"mein Entwurf wartet auf mich", lage(domain.HandelEntwurf), "aich", domain.LageDran, domain.SchrittAnbieten},
		{"ein Angebot an mich wartet auf meine Zustimmung", lage(domain.HandelAngeboten), "bdu", domain.LageDran, domain.SchrittZustimmen},
		{"wer zugestimmt hat, wartet", lage(domain.HandelAngeboten, "bdu"), "bdu", domain.LageWartet, ""},
		{"vorbereiten darf nur der Initiator", lage(domain.HandelAngenommen), "bdu", domain.LageWartet, ""},
		{"der Initiator bereitet vor", lage(domain.HandelAngenommen), "aich", domain.LageDran, domain.SchrittVorbereiten},
		{"verbucht, aber unbestätigt", lage(domain.HandelAbgleichOffen, "aich"), "aich", domain.LageWartet, ""},
		{"die Gegenbestätigung steht aus", lage(domain.HandelAbgleichOffen, "aich"), "bdu", domain.LageDran, domain.SchrittBestaetigen},
		{"eine Klärung braucht einen Menschen", lage(domain.HandelKlaerung), "bdu", domain.LageDran, ""},
		{"aber nicht einen Fremden", lage(domain.HandelKlaerung), "cer", domain.LageWartet, ""},
		{"beidseitig verbucht ist abgeschlossen", lage(domain.HandelBeidseitigVerbucht), "aich", domain.LageAbgeschlossen, ""},
		{"abgelehnt ist abgebrochen", lage(domain.HandelAbgelehnt), "aich", domain.LageAbgebrochen, ""},
		{"abgelaufen ist abgebrochen", lage(domain.HandelAbgelaufen), "bdu", domain.LageAbgebrochen, ""},
		{"wer nicht dabei ist, ist nie dran", lage(domain.HandelAngeboten), "cer", domain.LageWartet, ""},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			if lage := fall.angebot.Lage(fall.ich); lage != fall.erwartet {
				t.Errorf("Lage ist %q, erwartet %q", lage, fall.erwartet)
			}

			if schritt := fall.angebot.Naechster(fall.ich); schritt != fall.schritt {
				t.Errorf("nächster Schritt ist %q, erwartet %q", schritt, fall.schritt)
			}
		})
	}
}

// **Was ich darf, ist mehr als der nächste Schritt** — die Ausgänge
// gehören dazu, in der Reihenfolge der Matrix.
func TestDieMoeglichenSchritte(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		name     string
		angebot  domain.Handelsangebot
		ich      string
		erwartet []domain.Handelsschritt
	}{
		{"die Gegenseite eines Angebots", lage(domain.HandelAngeboten), "bdu", []domain.Handelsschritt{
			domain.SchrittGegenangebot, domain.SchrittZustimmen, domain.SchrittAblehnen, domain.SchrittAblaufen,
		}},
		{"der Initiator eines Angebots", lage(domain.HandelAngeboten), "aich", []domain.Handelsschritt{
			domain.SchrittZustimmen, domain.SchrittAbbrechen, domain.SchrittAblaufen,
		}},
		{"nach der eigenen Stimme bleiben die Ausgänge", lage(domain.HandelAngeboten, "aich"), "aich", []domain.Handelsschritt{
			domain.SchrittAbbrechen, domain.SchrittAblaufen,
		}},
		{"ein Ende hat keine", lage(domain.HandelBeidseitigVerbucht), "aich", nil},
		{"wer nicht dabei ist, hat keine", lage(domain.HandelAngeboten), "cer", nil},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			if schritte := fall.angebot.Moegliche(fall.ich); !slices.Equal(schritte, fall.erwartet) {
				t.Errorf("möglich sind %v, erwartet %v", schritte, fall.erwartet)
			}
		})
	}
}

// **Einseitig verbucht ist nicht fertig** (AK-04) — die Liste zeigt es.
func TestUnbestaetigtIstNurDerWegZwischenDenBestaenden(t *testing.T) {
	t.Parallel()

	for _, zustand := range domain.Handelszustaende() {
		erwartet := zustand == domain.HandelLokalVerbucht || zustand == domain.HandelAbgleichOffen
		if lage(zustand).Unbestaetigt() != erwartet {
			t.Errorf("%s: unbestätigt ist %v", zustand, !erwartet)
		}
	}
}

func TestVierLagen(t *testing.T) {
	t.Parallel()

	if lagen := domain.Handelslagen(); !slices.Equal(lagen, []domain.Handelslage{
		domain.LageDran, domain.LageWartet, domain.LageAbgeschlossen, domain.LageAbgebrochen,
	}) {
		t.Errorf("die Lagen sind %v", lagen)
	}

	if !domain.LageWartet.Gueltig() || domain.Handelslage("irgendwo").Gueltig() {
		t.Error("Gueltig unterscheidet nicht")
	}
}
