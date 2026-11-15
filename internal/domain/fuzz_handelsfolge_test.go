package domain_test

// **Ein Handel unter einer fremden Folge von Schritten, Vorschlägen und
// Ständen** — QA-06b
// (#1967),
// Fläche 78 aus `docs/FUZZZIELE.md`.
//
// # Drei Wege, auf denen sich ein Handel bewegt
//
//   - **ein Schritt** einer Partei — [domain.Handelsangebot.Schreite];
//   - **ein Vorschlag**, die nächste Revision über die Leitung —
//     [domain.Handelsangebot.NimmtVon];
//   - **ein Stand** der Gegenseite mit ihren Stimmen —
//     [domain.Handelsangebot.NimmtStand].
//
// Der Fuzzer wählt je zwei Byte Weg, Absender und Ziel, über einem Vorgang
// mit zwei oder drei Parteien. Was ein Fremder über die Leitung schickt,
// kommt so in jeder Reihenfolge an.
//
// # Was ein Fund ist — festgelegt **vor** dem ersten Lauf
//
//   - Panic
//   - abgelehnt ohne Klasse — [domain.ErrValidation] oder [domain.ErrConflict]
//   - **abgelehnt, und der Stand ist ein anderer**; angenommen, und der Stand
//     **davor** ist ein anderer (eine geteilte Scheibe)
//   - angenommen, und das Angebot besteht [domain.Handelsangebot.Pruefe] nicht
//   - **ein Zustandswechsel, den die Matrix nicht kennt** — für einen Schritt
//     genau die Kante, für einen Vorschlag die Kante `gegenangebot` aus dem
//     heutigen Zustand; das Fenster bietet „vorschlagen" nur dann an
//     (TRADE-02,
//     #1716:
//     „nach der Einigung gibt es nichts mehr vorzuschlagen")
//
// **Hier stand vor dem ersten Lauf auch die Rolle** — die Kante, „die der
// Absender gehen darf" ([domain.Handelsangebot.Moegliche]). Die Saat fand
// beides: Vorschläge nach der Einigung **und** den Initiator, der auf sein
// eigenes Angebot vorschlägt. Behoben ist das erste. Das zweite prüfte nur
// [domain.Handelsangebot.NimmtVon], weil nur dort der Absender bekannt ist;
// der eigene Vorschlag nennt keinen, und eine Prüfung allein auf der
// empfangenden Seite ließe beide Seiten auseinanderlaufen. Die Rolle hält
// weiterhin nur das Fenster — offen im Bericht zu QA-06b.
//   - **ein Stand der Gegenseite, der einen Abschlussschritt nachspielt** —
//     `verbuchen`, `abgleichen` und `bestaetigen` verlangen einen Nachweis
//     (`handelsstandnahme.go`)
//   - Stimmen, die keine Partei sind, doppelt stehen oder alle Parteien
//     umfassen — dann hätte der Beschluss schon fallen müssen
//
//	go test ./internal/domain -run Fuzz -fuzz FuzzHandelsfolge -fuzztime 2m

import (
	"errors"
	"reflect"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Jeder Wechsel ist einer der Matrix, und wer ablehnt, lässt den Stand
// stehen.**
func FuzzHandelsfolge(f *testing.F) {
	for _, saat := range [][]byte{
		{0, 0, 2, 1, 2, 0, 3, 0},
		{0, 0, 1, 1, 64, 1, 2, 0, 2, 1},
		{0, 0, 2, 0, 2, 1, 3, 0, 4, 0, 4, 1, 64, 0},
		{0, 0, 2, 0, 2, 1, 3, 0, 4, 0, 4, 1, 128, 1, 5, 0, 128, 0},
		{0, 0, 9, 1, 10, 0, 11, 2},
	} {
		f.Add(false, saat)
		f.Add(true, saat)
	}

	f.Fuzz(func(t *testing.T, dreiParteien bool, folge []byte) {
		heute := handelsbeginn(dreiParteien)

		for stelle := 0; stelle+1 < len(folge) && stelle < 2*64; stelle += 2 {
			wer := "fremd"
			if teil := int(folge[stelle+1]); teil < len(heute.Parteien) {
				wer = heute.Parteien[teil]
			}

			heute = handelsschritt(t, heute, folge[stelle], wer)
		}
	})
}

// handelsbeginn ist ein Entwurf, wie `HandelEroeffnen` ihn anlegt.
func handelsbeginn(dreiParteien bool) domain.Handelsangebot {
	parteien := []string{"anna", "bert"}
	if dreiParteien {
		parteien = append(parteien, "clara")
	}

	return domain.Handelsangebot{
		Vertragsfassung: domain.Handelsvertragsfassung,
		Vorgang:         "v1", Revision: 1, Art: domain.HandelTausch,
		Initiator: "anna", Parteien: parteien, Zustand: domain.HandelEntwurf,
		Positionen: []domain.Handelsposition{{Von: "anna", Printing: "abc-123", Menge: 2}},
	}
}

// handelsschritt geht einen Weg: die unteren sechs Bit wählen den Schritt,
// die oberen zwei den Weg — Schritt, Vorschlag oder Stand.
func handelsschritt(t *testing.T, heute domain.Handelsangebot, wahl byte, wer string) domain.Handelsangebot {
	t.Helper()

	vorher := klonen(heute)
	schritte := domain.Handelsschritte()
	schritt := schritte[int(wahl&63)%len(schritte)]

	var (
		danach domain.Handelsangebot
		fehler error
		weg    string
	)

	switch wahl >> 6 {
	case 1:
		weg = "Vorschlag"
		danach = heute.Fortgeschrieben([]domain.Handelsposition{{Von: wer, Printing: "x", Menge: 1 + int(wahl&7)}})

		if wahl&8 != 0 {
			danach.Zustimmungen = []string{wer}
		}

		fehler = heute.NimmtVon(danach, wer)
		if fehler == nil && !slices.ContainsFunc(domain.Handelsmatrix(), func(u domain.Handelsuebergang) bool {
			return u.Schritt == domain.SchrittGegenangebot && u.Von == heute.Zustand
		}) {
			t.Errorf("ein Vorschlag von %q auf %q angenommen — dort geht kein Gegenangebot", wer, heute.Zustand)
		}

	case 2:
		weg = "Stand"
		danach, fehler = heute.NimmtStand(fremderStand(heute, schritt, wer), wer)

		if fehler == nil && danach.Zustand != heute.Zustand && slices.ContainsFunc(domain.Handelsmatrix(),
			func(u domain.Handelsuebergang) bool {
				return u.Nach == danach.Zustand && domain.IstAbschlussschritt(u.Schritt)
			}) {
			t.Errorf("ein Stand von %q spielte einen Abschluss nach: %q → %q", wer, heute.Zustand, danach.Zustand)
		}

	default:
		weg = string(schritt)
		danach, fehler = heute.Schreite(schritt, wer)

		if fehler == nil && danach.Zustand != heute.Zustand && !kanteDa(schritt, heute.Zustand, danach.Zustand) {
			t.Errorf("%s von %q: %q → %q steht in keiner Kante", schritt, wer, heute.Zustand, danach.Zustand)
		}
	}

	if !reflect.DeepEqual(heute, vorher) {
		t.Fatalf("%s von %q hat den Stand davor verändert", weg, wer)
	}

	if fehler != nil {
		if !errors.Is(fehler, domain.ErrValidation) && !errors.Is(fehler, domain.ErrConflict) {
			t.Errorf("%s von %q abgelehnt ohne Klasse: %v", weg, wer, fehler)
		}

		if weg != "Vorschlag" && !reflect.DeepEqual(danach, heute) {
			t.Errorf("%s von %q abgelehnt, und der Stand ist ein anderer", weg, wer)
		}

		return heute
	}

	handelsstandPruefen(t, weg, danach)

	return danach
}

// fremderStand ist, was die Gegenseite schickt: derselbe Stand, ein Schritt
// weiter, und die Stimme des Absenders.
func fremderStand(heute domain.Handelsangebot, schritt domain.Handelsschritt, wer string) domain.Handelsangebot {
	fremd := klonen(heute)
	fremd.Zustimmungen = append(fremd.Zustimmungen, wer)

	for _, kante := range domain.Handelsmatrix() {
		if kante.Schritt == schritt && kante.Von == heute.Zustand {
			fremd.Zustand = kante.Nach
		}
	}

	return fremd
}

// kanteDa sagt, ob die Matrix diesen Wechsel für diesen Schritt kennt.
func kanteDa(schritt domain.Handelsschritt, von, nach domain.Handelszustand) bool {
	return slices.ContainsFunc(domain.Handelsmatrix(), func(u domain.Handelsuebergang) bool {
		return u.Schritt == schritt && u.Von == von && u.Nach == nach
	})
}

// handelsstandPruefen hält ein angenommenes Angebot.
func handelsstandPruefen(t *testing.T, weg string, angebot domain.Handelsangebot) {
	t.Helper()

	if fehler := angebot.Pruefe(); fehler != nil {
		t.Errorf("nach %s ungültig: %v", weg, fehler)
	}

	gesehen := map[string]bool{}
	for _, stimme := range angebot.Zustimmungen {
		if gesehen[stimme] || !slices.Contains(angebot.Parteien, stimme) {
			t.Errorf("nach %s stimmen %v", weg, angebot.Zustimmungen)
		}

		gesehen[stimme] = true
	}

	if len(angebot.Zustimmungen) >= len(angebot.Parteien) {
		t.Errorf("nach %s haben alle zugestimmt, und %q steht noch", weg, angebot.Zustand)
	}
}

// klonen gibt eine Kopie ohne geteilte Scheiben.
func klonen(angebot domain.Handelsangebot) domain.Handelsangebot {
	angebot.Parteien = slices.Clone(angebot.Parteien)
	angebot.Positionen = slices.Clone(angebot.Positionen)
	angebot.Zustimmungen = slices.Clone(angebot.Zustimmungen)

	return angebot
}
