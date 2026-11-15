package service_test

// **Die Fehlerwege der Decklisteneinfuhr** — Werkstück `T8` von Phase 5.6
// (#1473).
//
// **Eine eigene Datei**, damit `deckeinfuhr_test.go` unter der Zeilengrenze
// aus §2 bleibt: dort steht, was zugesagt ist, hier, was schiefgehen kann.
//
// # Warum es dafür ein eigenes Doppel gibt
//
// **Weil [deckdoppel] nur einen Schalter hat.** Sein `fehler` lässt **jeden**
// Weg scheitern, und damit endet `Uebernehmen` schon am Lesen — der Zweig, in
// dem das Legen scheitert, bliebe unbetreten und sähe aus wie eine geprüfte
// Zusage (§10).

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/sammelliste"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// haltedoppel ist ein Deckhalter mit einem Schalter je Weg.
type haltedoppel struct {
	deck domain.Deck

	// lesefehlerAb lässt `Deck` ab dem n-ten Ruf scheitern — 0 heißt nie.
	lesefehlerAb int
	lesungen     int

	legefehler error
	gelegt     int
}

func (h *haltedoppel) Deck(_ context.Context, _ int64) (domain.Deck, error) {
	h.lesungen++

	if h.lesefehlerAb != 0 && h.lesungen >= h.lesefehlerAb {
		return domain.Deck{}, errKaputt
	}

	return h.deck, nil
}

func (h *haltedoppel) PlatzSetzen(
	_ context.Context, _ int64, platz domain.Deckplatz,
) (domain.Deckplatz, error) {
	h.gelegt++

	if h.legefehler != nil {
		return domain.Deckplatz{}, h.legefehler
	}

	h.deck.Plaetze = append(h.deck.Plaetze, platz)

	return platz, nil
}

// eineZeile ist eine bestätigte Zeile.
func eineZeile() []service.Deckzeile {
	return []service.Deckzeile{{
		Printing: gueltigerDruck, Menge: 4,
		Bereich: domain.BereichHauptdeck,
	}}
}

// **Ein Deck, das nicht zu lesen ist, gibt keine Vorschau.**
func TestOhneLesbaresDeckGibtEsKeineVorschau(t *testing.T) {
	t.Parallel()

	halter := &haltedoppel{lesefehlerAb: 1}

	einfuhr := service.NeueDeckeinfuhr(&deutungsdoppel{}, halter)

	if _, fehler := einfuhr.Vorschau(t.Context(), 1, "4 Bolt"); fehler == nil {
		t.Error("die Vorschau kam ohne Deck zurück")
	}
}

// **Und eine Deutung, die klagt, kommt durch.**
//
// **Nicht verschluckt**: ohne Netz sagt sie es, und eine leere Vorschau
// sähe aus wie „nichts gefunden" (§8).
func TestEineKlagendeDeutungKommtDurch(t *testing.T) {
	t.Parallel()

	einfuhr := service.NeueDeckeinfuhr(
		&deutungsdoppel{fehler: errKaputt}, &haltedoppel{})

	if _, fehler := einfuhr.Vorschau(t.Context(), 1, "4 Bolt"); fehler == nil {
		t.Error("die Klage wurde verschluckt")
	}
}

// **Eine Zeile ohne Menge ist eine Karte** — und eine ohne Treffer ist keine.
func TestEineZeileOhneMengeUndOhneTreffer(t *testing.T) {
	t.Parallel()

	deutung := service.Listendeutung{Zeilen: []service.Zeilendeutung{
		{
			Zeile:   sammelliste.Zeile{Name: "Lightning Bolt", Abschnitt: "deck"},
			Treffer: []service.Suchtreffer{{Kennung: gueltigerDruck}},
			Drucke:  1,
		},
		{
			Zeile:  sammelliste.Zeile{Menge: 2, Name: "Gibtsnicht", Abschnitt: "deck"},
			Drucke: 0,
		},
	}}

	einfuhr := service.NeueDeckeinfuhr(
		&deutungsdoppel{deutung: deutung}, &haltedoppel{})

	vorschau, fehler := einfuhr.Vorschau(t.Context(), 1, "Bolt\nGibtsnicht")
	if fehler != nil {
		t.Fatalf("die Vorschau: %v", fehler)
	}

	if vorschau.Zeilen[0].Menge != 1 {
		t.Errorf("die Zeile ohne Menge hat %d", vorschau.Zeilen[0].Menge)
	}

	if vorschau.Zeilen[1].Uebernehmbar {
		t.Error("die Zeile ohne Treffer ist übernehmbar")
	}

	if vorschau.Zeilen[1].Befund == "" {
		t.Error("die Zeile ohne Treffer sagt nicht, warum")
	}
}

// **Ohne bestätigte Zeile wird nichts gelegt.**
//
// **Eine Beanstandung und kein leerer Erfolg**: „ich habe nichts ausgewählt"
// und „es ist nichts hineingekommen" sind zwei Auskünfte.
func TestOhneBestaetigteZeileWirdNichtsGelegt(t *testing.T) {
	t.Parallel()

	halter := &haltedoppel{}
	einfuhr := service.NeueDeckeinfuhr(&deutungsdoppel{}, halter)

	_, fehler := einfuhr.Uebernehmen(t.Context(), 1, domain.Deckstand(domain.Deck{}), nil)
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("der Fehler ist keine Beanstandung: %v", fehler)
	}

	if halter.lesungen != 0 {
		t.Error("das Deck wurde trotzdem gelesen")
	}
}

// **Jeder Weg der Übernahme reicht seinen Fehler durch — und legt dabei
// nichts halb.**
func TestJederWegDerUebernahmeReichtSeinenFehlerDurch(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		halter *haltedoppel
		zeilen []service.Deckzeile
		gelegt int
	}{
		"das Deck ist nicht zu lesen": {
			halter: &haltedoppel{lesefehlerAb: 1}, zeilen: eineZeile(),
		},
		"eine Zeile ist kein Platz": {
			halter: &haltedoppel{},
			zeilen: []service.Deckzeile{
				{Printing: gueltigerDruck, Menge: 4, Bereich: domain.BereichHauptdeck},
				{Printing: "kein Druck", Menge: 1, Bereich: domain.BereichHauptdeck},
			},
		},
		"das Legen scheitert": {
			halter: &haltedoppel{legefehler: errKaputt}, zeilen: eineZeile(),
			gelegt: 1,
		},
		"der neue Stand ist nicht zu lesen": {
			halter: &haltedoppel{lesefehlerAb: 2}, zeilen: eineZeile(),
			gelegt: 1,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			einfuhr := service.NeueDeckeinfuhr(&deutungsdoppel{}, fall.halter)

			_, fehler := einfuhr.Uebernehmen(t.Context(), 1,
				domain.Deckstand(domain.Deck{}), fall.zeilen)
			if fehler == nil {
				t.Fatal("der Fehler wurde verschluckt")
			}

			// **Die zweite Hälfte der Zusage**: eine abgewiesene Zeile
			// erreicht den Speicher gar nicht. Ohne sie bliebe die halbe
			// Liste im Deck.
			if fall.halter.gelegt != fall.gelegt {
				t.Errorf("es wurden %d Plätze gelegt, erwartet %d",
					fall.halter.gelegt, fall.gelegt)
			}
		})
	}
}

// **Und der Bericht nennt den neuen Stand.**
//
// **Damit eine zweite Liste ohne zweite Vorschau hineingeht** — ein Ruf
// dafür wäre einer zu viel, und wer ihn vergäße, bekäme einen Konflikt, den
// er nicht versteht.
func TestDerBerichtNenntDenNeuenStand(t *testing.T) {
	t.Parallel()

	halter := &haltedoppel{}
	vorher := domain.Deckstand(halter.deck)

	bericht, fehler := service.NeueDeckeinfuhr(&deutungsdoppel{}, halter).
		Uebernehmen(t.Context(), 1, vorher, eineZeile())
	if fehler != nil {
		t.Fatalf("die Übernahme: %v", fehler)
	}

	if bericht.Stand == vorher || bericht.Stand == "" {
		t.Errorf("der Bericht nennt den Stand %q, vorher war er %q",
			bericht.Stand, vorher)
	}
}
