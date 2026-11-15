package service_test

// **Eine Deckliste wird ein Deck** — Werkstück `T8` von Phase 5.6
// (#1473).
//
// # Was hier zugesagt wird
//
//   - Die Vorschau **schreibt nichts** und bindet sich an den Stand, aus dem
//     sie entstand.
//   - Hauptdeck und Sideboard bleiben getrennt, Mengen bleiben Mengen.
//   - Eine Zeile, die mehrere Drucke trifft, wird **nicht** übernommen —
//     „keine geratenen Drucke" steht so im Issue.
//   - Die Übernahme scheitert, wenn sich das Deck dazwischen geändert hat.
//   - **Und die Sammlung bleibt unberührt** — auch nach zwei Läufen. Ein
//     Deck ist kein Besitz.

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/sammelliste"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// dieProbeliste nennt zwei Bereiche und eine Merkliste.
const dieProbeliste = `Deck
4 Lightning Bolt (M11) 149
2 Lightning Bolt (M11) 149

Sideboard
3 Counterspell (M11) 50

Maybeboard
1 Brainstorm (M11) 51`

// deutungsdoppel gibt eine vorbereitete Listendeutung.
type deutungsdoppel struct {
	deutung service.Listendeutung
	fehler  error

	// text ist, was gedeutet werden sollte — für die Zusage, dass die
	// Einfuhr den Text durchreicht und nicht selbst zerteilt.
	text string
}

func (d *deutungsdoppel) Deuten(_ context.Context, text string) (
	service.Listendeutung, error,
) {
	d.text = text

	return d.deutung, d.fehler
}

// deutungMitAbschnitten baut die Deutung, die der Zerteiler für
// [dieProbeliste] gäbe — eindeutig, wo ein Setcode dabeisteht.
func deutungMitAbschnitten() service.Listendeutung {
	treffer := func(kennung, name string) []service.Suchtreffer {
		return []service.Suchtreffer{{
			Kennung: kennung, Name: name, Setcode: "m11", Nummer: "149",
		}}
	}

	return service.Listendeutung{Zeilen: []service.Zeilendeutung{
		{
			Zeile:   zeile(4, "Lightning Bolt", "deck"),
			Treffer: treffer(gueltigerDruck, "Lightning Bolt"), Drucke: 1,
		},
		{
			Zeile:   zeile(2, "Lightning Bolt", "deck"),
			Treffer: treffer(gueltigerDruck, "Lightning Bolt"), Drucke: 1,
		},
		{
			Zeile:   zeile(3, "Counterspell", "sideboard"),
			Treffer: treffer(zweiterDruck, "Counterspell"), Drucke: 1,
		},
		{
			Zeile:   zeile(1, "Brainstorm", "maybeboard"),
			Treffer: treffer(dritterDruck, "Brainstorm"), Drucke: 1,
		},
	}}
}

// zweiterDruck und dritterDruck sind weitere gültige Kennungen.
const (
	zweiterDruck = "b1b2c3d4-0000-0000-0000-000000000000"
	dritterDruck = "c1b2c3d4-0000-0000-0000-000000000000"
)

// zeile baut eine gelesene Listenzeile.
func zeile(menge int, name, abschnitt string) sammelliste.Zeile {
	return sammelliste.Zeile{
		Menge: menge, Name: name, Setcode: "m11", Nummer: "149",
		Abschnitt: abschnitt, Roh: name,
	}
}

// deckMitEinfuhr baut Einfuhr und Deckspeicher samt einem leeren Deck.
func deckMitEinfuhr(t *testing.T, deutung service.Listendeutung) (
	*service.Deckeinfuhr, *deckdoppel, domain.Deck,
) {
	t.Helper()

	speicher := &deckdoppel{}

	deck, fehler := speicher.Anlegen(t.Context(),
		domain.Deck{Name: "Probe", Spiel: "magic"})
	if fehler != nil {
		t.Fatalf("das Deck anlegen: %v", fehler)
	}

	return service.NeueDeckeinfuhr(
		&deutungsdoppel{deutung: deutung}, speicher), speicher, deck
}

// **Die Vorschau trennt die Bereiche, zählt die Mengen und schreibt nichts.**
func TestDieVorschauTrenntHauptdeckUndSideboard(t *testing.T) {
	t.Parallel()

	einfuhr, speicher, deck := deckMitEinfuhr(t, deutungMitAbschnitten())

	vorschau, fehler := einfuhr.Vorschau(t.Context(), deck.ID, dieProbeliste)
	if fehler != nil {
		t.Fatalf("die Vorschau: %v", fehler)
	}

	if speicher.gelegt != 0 {
		t.Errorf("die Vorschau hat %d Plätze gelegt", speicher.gelegt)
	}

	if vorschau.Stand != domain.Deckstand(deck) {
		t.Errorf("die Vorschau bindet sich an %q statt an %q",
			vorschau.Stand, domain.Deckstand(deck))
	}

	// **Zwei Zeilen über dieselbe Karte werden eine mit sechs** — Mengen
	// bleiben Mengen, und zwei Plätze über denselben Druck wären zwei Zeilen
	// über dasselbe.
	if len(vorschau.Zeilen) != 3 {
		t.Fatalf("%d Zeilen statt 3: %+v", len(vorschau.Zeilen), vorschau.Zeilen)
	}

	erwartet := map[domain.Deckbereich]int{
		domain.BereichHauptdeck: 6, domain.BereichSideboard: 3,
	}

	gezaehlt := map[domain.Deckbereich]int{}

	for _, zeile := range vorschau.Zeilen {
		if zeile.Uebernehmbar {
			gezaehlt[zeile.Bereich] += zeile.Menge
		}
	}

	for bereich, menge := range erwartet {
		if gezaehlt[bereich] != menge {
			t.Errorf("%s hat %d Stück, erwartet %d", bereich,
				gezaehlt[bereich], menge)
		}
	}

	// **Die Merkliste steht da und ist nicht übernehmbar.** Wer sie ausließe,
	// ließe jemanden rätseln, wo seine Zeile geblieben ist (§8).
	var merkliste int

	for _, zeile := range vorschau.Zeilen {
		if !zeile.Uebernehmbar {
			merkliste++

			if zeile.Befund == "" {
				t.Errorf("%q wird nicht übernommen und sagt nicht warum",
					zeile.Name)
			}
		}
	}

	if merkliste != 1 {
		t.Errorf("%d Zeilen bleiben draußen, erwartet 1", merkliste)
	}
}

// **Eine mehrdeutige Zeile wird nicht geraten.**
func TestEineMehrdeutigeZeileWirdNichtUebernommen(t *testing.T) {
	t.Parallel()

	deutung := service.Listendeutung{Zeilen: []service.Zeilendeutung{{
		Zeile: zeile(4, "Counterspell", "deck"),
		Treffer: []service.Suchtreffer{
			{Kennung: gueltigerDruck, Name: "Counterspell", Setcode: "m11"},
			{Kennung: zweiterDruck, Name: "Counterspell", Setcode: "m12"},
		},
		Drucke: 2,
	}}}

	einfuhr, _, deck := deckMitEinfuhr(t, deutung)

	vorschau, fehler := einfuhr.Vorschau(t.Context(), deck.ID, "4 Counterspell")
	if fehler != nil {
		t.Fatalf("die Vorschau: %v", fehler)
	}

	if len(vorschau.Zeilen) != 1 {
		t.Fatalf("%d Zeilen statt 1", len(vorschau.Zeilen))
	}

	if vorschau.Zeilen[0].Uebernehmbar {
		t.Error("die mehrdeutige Zeile ist übernehmbar")
	}

	if vorschau.Zeilen[0].Printing != "" {
		t.Errorf("es wurde ein Druck geraten: %q", vorschau.Zeilen[0].Printing)
	}

	// **Die Auswahl steht dabei** — sonst könnte niemand entscheiden.
	if len(vorschau.Zeilen[0].Treffer) != 2 {
		t.Errorf("%d Treffer stehen zur Wahl", len(vorschau.Zeilen[0].Treffer))
	}
}

// **Die Übernahme schreibt Plätze und rührt die Sammlung nicht an.**
func TestDieUebernahmeSchreibtPlaetzeUndKeinenBesitz(t *testing.T) {
	t.Parallel()

	einfuhr, speicher, deck := deckMitEinfuhr(t, deutungMitAbschnitten())

	vorschau, fehler := einfuhr.Vorschau(t.Context(), deck.ID, dieProbeliste)
	if fehler != nil {
		t.Fatalf("die Vorschau: %v", fehler)
	}

	bericht, fehler := einfuhr.Uebernehmen(t.Context(), deck.ID,
		vorschau.Stand, uebernehmbare(vorschau))
	if fehler != nil {
		t.Fatalf("die Übernahme: %v", fehler)
	}

	if bericht.Gelegt != 2 || bericht.Stueck != 9 {
		t.Errorf("gelegt %d, Stück %d — erwartet 2 und 9",
			bericht.Gelegt, bericht.Stueck)
	}

	gelegt, gelesen := speicher.Deck(t.Context(), deck.ID)
	if gelesen != nil {
		t.Fatalf("das Deck lesen: %v", gelesen)
	}

	if len(gelegt.Plaetze) != 2 {
		t.Fatalf("%d Plätze im Deck: %+v", len(gelegt.Plaetze), gelegt.Plaetze)
	}

	for _, platz := range gelegt.Plaetze {
		if platz.KarteID != 0 {
			t.Errorf("der Platz nennt eine eigene Karte: %+v", platz)
		}
	}
}

// **Ein zweiter Lauf über denselben Stand scheitert.**
//
// **Und das ist zugleich die Zusage gegen den doppelten Besitz**: wer
// zweimal bestätigt, legt nicht zweimal.
func TestEinZweiterLaufUeberDenselbenStandScheitert(t *testing.T) {
	t.Parallel()

	einfuhr, speicher, deck := deckMitEinfuhr(t, deutungMitAbschnitten())

	vorschau, gesehen := einfuhr.Vorschau(t.Context(), deck.ID, dieProbeliste)
	if gesehen != nil {
		t.Fatalf("die Vorschau: %v", gesehen)
	}

	if _, fehler := einfuhr.Uebernehmen(t.Context(), deck.ID,
		vorschau.Stand, uebernehmbare(vorschau)); fehler != nil {
		t.Fatalf("der erste Lauf: %v", fehler)
	}

	vorher := speicher.gelegt

	_, fehler := einfuhr.Uebernehmen(t.Context(), deck.ID,
		vorschau.Stand, uebernehmbare(vorschau))
	if fehler == nil {
		t.Fatal("der zweite Lauf kam durch")
	}

	if !errors.Is(fehler, domain.ErrConflict) {
		t.Errorf("der Fehler ist kein Konflikt: %v", fehler)
	}

	if speicher.gelegt != vorher {
		t.Errorf("der zweite Lauf hat %d Plätze gelegt",
			speicher.gelegt-vorher)
	}
}

// uebernehmbare macht aus der Vorschau, was bestätigt wird.
func uebernehmbare(vorschau service.Deckvorschau) []service.Deckzeile {
	zeilen := []service.Deckzeile{}

	for _, zeile := range vorschau.Zeilen {
		if zeile.Uebernehmbar {
			zeilen = append(zeilen, service.Deckzeile{
				Printing: zeile.Printing, Menge: zeile.Menge,
				Bereich: zeile.Bereich,
			})
		}
	}

	return zeilen
}
