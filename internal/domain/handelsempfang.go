package domain

// **Was eine empfangene Revision behaupten darf** — TRADE-03.AK-01 und
// NAT-01 (#1717).
//
// # Der Fund, für den es diese Regeln gibt
//
// **Eine Revision kam so an, wie der Absender sie schrieb** — mit seinem
// Zustand und seinen Stimmen. [Handelsangebot.Nimmt] prüft Kopf und Kette,
// aber Zustand und Stimmen gehen nicht in den Abdruck ein und entgingen ihm:
// wer eine erste Revision als `lokalVerbucht` schickte oder einen Vorschlag
// mit beiden Stimmen, stand hier dort, ohne dass jemand zugestimmt hätte.
// [Handelsangebot.NimmtStand] hielt dieselbe Regel für Stimmen längst (Befund
// 2 des Zweitblicks über Phase 5.6) — die zwei anderen Wege nicht.
//
// # Was eine Revision behaupten darf
//
//   - **Eine erste** kommt vom Initiator, steht auf `angeboten` — sie geht
//     erst nach dem Anbieten hinaus — und trägt höchstens seine Stimme.
//   - **Ein Vorschlag** ist, was [Handelsangebot.Fortgeschrieben] baut:
//     `verhandlung` und höchstens die Stimme des Absenders.

import (
	"fmt"
	"slices"
)

// PruefeErsteRevision weist eine erste Revision ab, die mehr behauptet, als
// ihr Absender sagen darf.
func PruefeErsteRevision(angebot Handelsangebot, absender string) error {
	if fehler := angebot.Pruefe(); fehler != nil {
		return fehler
	}

	switch {
	case absender != angebot.Initiator:
		return fmt.Errorf("%w: eine erste Revision schickt der Initiator %q, nicht %q",
			ErrValidation, angebot.Initiator, absender)

	case angebot.Zustand != HandelAngeboten:
		return fmt.Errorf("%w: eine erste Revision steht auf %q, nicht auf %q",
			ErrValidation, HandelAngeboten, angebot.Zustand)
	}

	return nurEigeneStimme(angebot, absender)
}

// NimmtVon ist [Handelsangebot.Nimmt] für einen Vorschlag, der über die
// Leitung kam — und er behauptet keinen Zustand und keine fremde Stimme.
func (a Handelsangebot) NimmtVon(vorschlag Handelsangebot, absender string) error {
	if fehler := a.NimmtVorschlagVon(vorschlag, absender); fehler != nil {
		return fehler
	}

	if vorschlag.Zustand != HandelVerhandlung {
		return fmt.Errorf("%w: ein Vorschlag steht auf %q, nicht auf %q",
			ErrValidation, HandelVerhandlung, vorschlag.Zustand)
	}

	return nurEigeneStimme(vorschlag, absender)
}

// nurEigeneStimme weist Stimmen ab, die nicht die des Absenders sind.
func nurEigeneStimme(angebot Handelsangebot, absender string) error {
	if fremde := slices.IndexFunc(angebot.Zustimmungen, func(stimme string) bool {
		return stimme != absender
	}); fremde >= 0 {
		return fmt.Errorf("%w: %q stimmt für %q — für andere stimmt niemand",
			ErrValidation, absender, angebot.Zustimmungen[fremde])
	}

	return nil
}
