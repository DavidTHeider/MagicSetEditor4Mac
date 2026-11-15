package domain_test

// **Die Reihenfolge der Revisionen** — Werkstück `T1` von Phase 5.6
// (#1466).
//
// # Was hier gemessen wird
//
// **Dass zwei Vorschläge gegen denselben Stand einen Konflikt ergeben und
// nicht zwei Revisionen mit der Nummer N.** Wer beide annähme, hätte zwei
// Wahrheiten, und beim ersten Unterschied wüsste niemand, welche gilt.
//
// **Der Initiator sequenziert**, und zwar gegen den erwarteten Basisstand:
// ein Vorschlag nennt den Abdruck der Revision, auf der er aufsetzt. Stimmt
// der nicht mehr, hat inzwischen jemand anderes geschrieben.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// einVorschlag gibt die nächste Revision zu einem Angebot.
func einVorschlag(auf domain.Handelsangebot, menge int) domain.Handelsangebot {
	anders := auf.Positionen[0]
	anders.Menge = menge

	return auf.Fortgeschrieben([]domain.Handelsposition{anders})
}

// **Ein Vorschlag auf den heutigen Stand wird angenommen.**
func TestEinVorschlagAufDenHeutigenStandGehtDurch(t *testing.T) {
	t.Parallel()

	heute := einAngebot()
	heute.Zustand = domain.HandelAngeboten

	if fehler := heute.Nimmt(einVorschlag(heute, 3)); fehler != nil {
		t.Errorf("der Vorschlag wird abgewiesen: %v", fehler)
	}
}

// **Was einen Vorschlag scheitern lässt — und woran man es unterscheidet.**
//
// **Zwei Fehlerarten und nicht eine**: ein Stand von vorhin ist ein
// [domain.ErrConflict] — noch einmal lesen und noch einmal vorschlagen hilft.
// Ein fremder Vorgang oder eine unbekannte Vertragsfassung ist ein
// [domain.ErrValidation] — daran ändert kein Wiederholen etwas.
func TestWasEinenVorschlagScheiternLaesst(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		verbiegen func(*domain.Handelsangebot)
		art       error
	}{
		"eine alte Annahme": {
			verbiegen: func(v *domain.Handelsangebot) { v.Eltern = "von gestern" },
			art:       domain.ErrConflict,
		},
		"eine übersprungene Revision": {
			verbiegen: func(v *domain.Handelsangebot) { v.Revision = 7 },
			art:       domain.ErrConflict,
		},
		"ein fremder Vorgang": {
			verbiegen: func(v *domain.Handelsangebot) { v.Vorgang = "v99" },
			art:       domain.ErrValidation,
		},
		"eine andere Art": {
			verbiegen: func(v *domain.Handelsangebot) { v.Art = domain.HandelMarkt },
			art:       domain.ErrValidation,
		},
		"eine andere Partei": {
			verbiegen: func(v *domain.Handelsangebot) {
				v.Parteien = []string{"anna", "cara"}
			},
			art: domain.ErrValidation,
		},
		"ein anderer Initiator": {
			verbiegen: func(v *domain.Handelsangebot) { v.Initiator = "bert" },
			art:       domain.ErrValidation,
		},
		"eine unbekannte Vertragsfassung": {
			verbiegen: func(v *domain.Handelsangebot) { v.Vertragsfassung = 99 },
			art:       domain.ErrValidation,
		},
		"ein Vorschlag, der selbst nicht prüft": {
			verbiegen: func(v *domain.Handelsangebot) { v.Positionen[0].Menge = 0 },
			art:       domain.ErrValidation,
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			heute := einAngebot()
			heute.Zustand = domain.HandelAngeboten

			vorschlag := einVorschlag(heute, 3)
			fall.verbiegen(&vorschlag)

			if fehler := heute.Nimmt(vorschlag); !errors.Is(fehler, fall.art) {
				t.Errorf("%q ergibt %v", name, fehler)
			}
		})
	}
}

// **Zwei Vorschläge gegen denselben Stand: der zweite bekommt einen Konflikt.**
//
// **Das ist die Zusage des Issues**, und sie ist der Grund für die ganze
// Kette: „konkurrierende Vorschläge erzeugen Konflikt statt zwei verschiedene
// Revisionen N".
func TestDerZweiteVorschlagGegenDenselbenStandBekommtEinenKonflikt(t *testing.T) {
	t.Parallel()

	heute := einAngebot()
	heute.Zustand = domain.HandelAngeboten

	vonAnna := einVorschlag(heute, 3)
	vonBert := einVorschlag(heute, 4)

	if fehler := heute.Nimmt(vonAnna); fehler != nil {
		t.Fatalf("der erste Vorschlag: %v", fehler)
	}

	// **Der Initiator hat Annas Vorschlag eingetragen** — Berts rechnet
	// damit gegen einen Stand, den es nicht mehr gibt.
	if fehler := vonAnna.Nimmt(vonBert); !errors.Is(fehler, domain.ErrConflict) {
		t.Errorf("der zweite Vorschlag ergibt %v", fehler)
	}
}

// **Aus einem Ende nimmt niemand mehr etwas an.**
func TestAufEinEndeFolgtKeinVorschlag(t *testing.T) {
	t.Parallel()

	heute := einAngebot()
	heute.Zustand = domain.HandelAbgelehnt

	if fehler := heute.Nimmt(einVorschlag(heute, 3)); !errors.Is(
		fehler, domain.ErrConflict) {
		t.Errorf("es kam %v", fehler)
	}
}

// **Ein Vorschlag geht nur, wo ein Gegenangebot geht** — QA-06b
// (#1967),
// gefunden von `FuzzHandelsfolge`.
//
// Das Fenster bietet „vorschlagen" nur an, wo die Matrix `gegenangebot` für
// den Fragenden kennt (TRADE-02, #1716: nach der Einigung gibt es nichts mehr
// vorzuschlagen). Die Leitung nahm einen Vorschlag trotzdem überall an, wo
// der Vorgang nicht zu Ende war — auch über Kommandozeile, REST und die
// Leitung zu einem anderen Rechner. Aus `uebergabeOffen` führte so ein
// Vorschlag nach `verhandlung` und von dort ein `ablehnen` heraus, obwohl
// 0149 genau dort nur die Klärung lässt.
func TestEinVorschlagGehtNurWoEinGegenangebotGeht(t *testing.T) {
	t.Parallel()

	for _, zustand := range []domain.Handelszustand{
		domain.HandelEntwurf, domain.HandelAngenommen, domain.HandelVorbereitet,
		domain.HandelUebergabeOffen, domain.HandelLokalVerbucht,
		domain.HandelAbgleichOffen, domain.HandelKlaerung,
	} {
		heute := einAngebot()
		heute.Zustand = zustand

		if fehler := heute.Nimmt(einVorschlag(heute, 3)); !errors.Is(fehler, domain.ErrConflict) {
			t.Errorf("ein Vorschlag auf %q ergibt %v", zustand, fehler)
		}
	}

	// **Aus `angeboten` und `verhandlung` geht es weiter** — dort kennt die
	// Matrix das Gegenangebot.
	for _, zustand := range []domain.Handelszustand{domain.HandelAngeboten, domain.HandelVerhandlung} {
		heute := einAngebot()
		heute.Zustand = zustand

		if fehler := heute.NimmtVon(einVorschlag(heute, 3), "bert"); fehler != nil {
			t.Errorf("ein Vorschlag auf %q ergibt %v", zustand, fehler)
		}
	}
}

// **Die Vertragsfassung bindet mit.**
//
// **Sonst hätten zwei Fassungen desselben Protokolls dieselbe Unterschrift**,
// und ein Empfänger, der die neuere anders liest, unterschriebe etwas
// anderes, als er gelesen hat.
func TestDieVertragsfassungBindetMit(t *testing.T) {
	t.Parallel()

	eine := einAngebot()

	andere := eine
	andere.Vertragsfassung = eine.Vertragsfassung + 1

	if andere.Abdruck() == eine.Abdruck() {
		t.Error("die Vertragsfassung ändert den Abdruck nicht")
	}
}
