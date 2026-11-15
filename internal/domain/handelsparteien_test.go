package domain_test

// **Zwei Parteien sind zwei** — Befund des Inline-Reviews zu
// #1637.
//
// # Was hier gemessen wird
//
// **Dass derselbe Name nicht zweimal unter den Parteien steht.** Der Vertrag
// verlangte bisher **zwei** Parteien und den Initiator darunter — beides
// hält auch `[anna, anna]`.
//
// # Warum das mehr ist als eine Formsache
//
// **Weil ein solcher Handel nie fertig wird.** `zustimmen`, `uebergeben` und
// `bestaetigen` wechseln den Zustand erst, wenn **jede** Partei gestimmt hat
// ([Handelsangebot.mitStimme]); wer zweimal dasteht, stimmt einmal, und
// `len(stimmen) == len(Parteien)` wird nie wahr. Der Vorgang stünde für
// immer in `angeboten` — und niemand sähe, warum.
//
// **Bis zum 2026-09-19 kam der Fall über keine Fläche herein.** Seit `T1`
// eine hat (#1637), kommt er über `HandelEroeffnen`: wer sich selbst als
// Gegenüber einträgt, bekommt jetzt eine Absage statt eines toten Vorgangs.

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Derselbe Name zweimal ist keine zweite Partei.**
func TestDieselbePartieZweimalIstKeinHandel(t *testing.T) {
	t.Parallel()

	angebot := domain.Handelsangebot{
		Vertragsfassung: domain.Handelsvertragsfassung,
		Vorgang:         "h1", Revision: 1,
		Art: domain.HandelTausch, Initiator: "anna",
		Parteien: []string{"anna", "anna"},
		Zustand:  domain.HandelEntwurf,
		Positionen: []domain.Handelsposition{{
			Von: "anna", Printing: "abc-123", Menge: 1,
			Guete: "NM", Sprache: "de", Finish: "nonfoil",
		}},
	}

	fehler := angebot.Pruefe()
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("es kam durch: %v", fehler)
	}

	// **Die Absage nennt den Namen, der doppelt steht** — sonst sucht
	// jemand in einer Liste von Kennungen.
	if !strings.Contains(fehler.Error(), "anna") {
		t.Errorf("die Absage nennt den Namen nicht: %v", fehler)
	}
}

// **Und zwei verschiedene gehen durch** — die Gegenprobe.
//
// **Ohne sie sagte das Nein oben nichts**: eine Prüfung, die alles abweist,
// wäre auch grün.
func TestZweiVerschiedeneParteienGehenDurch(t *testing.T) {
	t.Parallel()

	angebot := domain.Handelsangebot{
		Vertragsfassung: domain.Handelsvertragsfassung,
		Vorgang:         "h1", Revision: 1,
		Art: domain.HandelTausch, Initiator: "anna",
		Parteien: []string{"anna", "bert"},
		Zustand:  domain.HandelEntwurf,
		Positionen: []domain.Handelsposition{{
			Von: "anna", Printing: "abc-123", Menge: 1,
			Guete: "NM", Sprache: "de", Finish: "nonfoil",
		}},
	}

	if fehler := angebot.Pruefe(); fehler != nil {
		t.Fatalf("zwei verschiedene kamen nicht durch: %v", fehler)
	}
}
