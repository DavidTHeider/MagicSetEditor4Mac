package domain_test

// **Aus jedem Zustand führt ein Ausweg, den jede Partei allein gehen kann**
// — TRADE-03.AK-04 (#1717).
//
// Nach der Einigung gab es keinen: aus `angenommen`, `vorbereitet` und
// `uebergabeOffen` führte weder ein Abbruch noch eine Klärung. Ein
// Gegenüber, das verschwand, hielt den Vorgang und die festgehaltenen
// Exemplare für immer fest. `TestJederZustandStehtInDerMatrix` sah es nicht:
// ihm genügt irgendein Übergang hinaus, auch einer, der beide braucht.
//
// **Ausgenommen ist der Entwurf**: ihn sieht nur, wer ihn anlegte.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestAusJedemZustandFuehrtEinAuswegDenJedeParteiGehenKann(t *testing.T) {
	t.Parallel()

	for _, zustand := range domain.Handelszustaende() {
		if domain.HandelEnde(zustand) || zustand == domain.HandelEntwurf {
			continue
		}

		ausweg := false

		for _, uebergang := range domain.Handelsmatrix() {
			if uebergang.Von == zustand && uebergang.Wer == domain.HandelnderJede && !uebergang.Alle &&
				(domain.HandelEnde(uebergang.Nach) || uebergang.Nach == domain.HandelKlaerung) {
				ausweg = true
			}
		}

		if !ausweg {
			t.Errorf("aus %q kommt keine Partei allein heraus", zustand)
		}
	}
}

// **Ein Abbruch nach der Einigung gibt die Zusagen frei**; nach begonnener
// Übergabe gibt es keinen Abbruch, sondern die Klärung — die Karten können
// schon unterwegs sein.
func TestNachDerEinigungGehtEsHinausOhneZusagen(t *testing.T) {
	t.Parallel()

	for _, zustand := range []domain.Handelszustand{domain.HandelAngenommen, domain.HandelVorbereitet} {
		for _, wer := range []string{"banna", "bbert"} {
			danach, fehler := einHandelsstand(zustand).Schreite(domain.SchrittAbbrechen, wer)
			if fehler != nil || danach.Zustand != domain.HandelAbgebrochen {
				t.Errorf("%s bricht %q ab: %q, %v", wer, zustand, danach.Zustand, fehler)
			}
		}

		if wirkung := domain.ZusagenwirkungFuer(zustand, domain.HandelAbgebrochen); wirkung != domain.ZusagenFreigeben {
			t.Errorf("ein Abbruch aus %q wirkt %q", zustand, wirkung)
		}
	}

	offen := einHandelsstand(domain.HandelUebergabeOffen)

	if _, fehler := offen.Schreite(domain.SchrittAbbrechen, "bbert"); fehler == nil {
		t.Error("eine begonnene Übergabe ließ sich abbrechen")
	}

	if danach, fehler := offen.Schreite(domain.SchrittKlaeren, "bbert"); fehler != nil || danach.Zustand != domain.HandelKlaerung {
		t.Errorf("klären aus der Übergabe: %q, %v", danach.Zustand, fehler)
	}
}

// **Welcher Beschluss offen ist** — die Stimmen gehören zu ihm, und ohne ihn
// gibt es keine (TRADE-03.AK-02 und NAT-01): „Angenommen" neben „Zugestimmt:
// niemand" las sich wie ein Widerspruch.
func TestDerOffeneBeschluss(t *testing.T) {
	t.Parallel()

	for zustand, erwartet := range map[domain.Handelszustand]domain.Handelsschritt{
		domain.HandelAngeboten:     domain.SchrittZustimmen,
		domain.HandelVerhandlung:   domain.SchrittZustimmen,
		domain.HandelAngenommen:    "",
		domain.HandelVorbereitet:   domain.SchrittUebergeben,
		domain.HandelAbgleichOffen: domain.SchrittBestaetigen,
	} {
		if beschluss := einHandelsstand(zustand).OffenerBeschluss(); beschluss != erwartet {
			t.Errorf("%s: %q, erwartet %q", zustand, beschluss, erwartet)
		}
	}
}
