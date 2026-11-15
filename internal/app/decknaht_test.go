package app_test

// **Die Decknaht** — Werkstück 7 von Phase 4.6
// (#1293).
//
// Geprüft wird, was diese Schicht **tut**: umrechnen, verpacken, Fehler
// übersetzen. Was fachlich gilt, prüft `internal/service`.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// deckanwendung baut eine App mit allen vier Decknähten.
func deckanwendung(t *testing.T) (
	*app.App, *deckprobe, *formatprobe, *pruefprobe, *abgleichprobe,
) {
	t.Helper()

	decks := &deckprobe{}
	formate := &formatprobe{}
	pruefung := &pruefprobe{}
	abgleich := &abgleichprobe{}

	anwendung := app.NeueApp("")
	anwendung.MitDecknaht(decks)
	anwendung.MitFormatnaht(formate)
	anwendung.MitDeckpruefung(pruefung)
	anwendung.MitDeckabgleich(abgleich)

	return anwendung, decks, formate, pruefung, abgleich
}

// **Ein Deck kommt mit gezählten Bereichen und Kennungen als Text.**
func TestEinDeckKommtGezaehltUndAlsText(t *testing.T) {
	anwendung, decks, _, _, _ := deckanwendung(t)

	decks.deck = domain.Deck{
		ID: 7, Fassung: 2, Name: "Aggro", Spiel: "magic", FormatID: 5,
		Plaetze: []domain.Deckplatz{
			{ID: 1, Printing: "a", Menge: 4, Bereich: domain.BereichHauptdeck},
			{ID: 2, KarteID: 9, Menge: 1, Bereich: domain.BereichSideboard},
		},
	}

	antwort := anwendung.Deck(t.Context(), "7")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if antwort.Deck.ID != "7" || antwort.Deck.Fassung != "2" ||
		antwort.Deck.FormatID != "5" {
		t.Errorf("%+v", antwort.Deck)
	}

	if antwort.Deck.Hauptdeck != 4 || antwort.Deck.Sideboard != 1 {
		t.Errorf("Hauptdeck %d, Sideboard %d",
			antwort.Deck.Hauptdeck, antwort.Deck.Sideboard)
	}

	// **Genau eines von beiden ist gefüllt** — und die Null wird zur leeren
	// Zeichenkette, nicht zu „0".
	if antwort.Deck.Plaetze[0].KarteID != "" ||
		antwort.Deck.Plaetze[1].Printing != "" ||
		antwort.Deck.Plaetze[1].KarteID != "9" {
		t.Errorf("%+v", antwort.Deck.Plaetze)
	}
}

// **Ein Deck ohne Format trägt eine leere Kennung, keine Null.**
func TestEinDeckOhneFormatTraegtEineLeereKennung(t *testing.T) {
	anwendung, decks, _, _, _ := deckanwendung(t)

	decks.deck = domain.Deck{ID: 7, Name: "Aggro", Spiel: "magic"}

	if kennung := anwendung.Deck(t.Context(), "7").Deck.FormatID; kennung != "" {
		t.Errorf("FormatID %q, erwartet leer", kennung)
	}
}

// **Die Übersicht kommt als Liste — auch im Fehlerfall.**
//
// Ein `null` dort ließe jede Seite mit einer leeren Tabelle stehen; die
// Zusage dieses Pakets ist die leere Liste.
func TestDieDeckuebersichtIstImmerEineListe(t *testing.T) {
	anwendung, decks, _, _, _ := deckanwendung(t)

	decks.fehler = errors.New("der Bestand schweigt")

	antwort := anwendung.Decks(t.Context(), "magic")
	if antwort.Fehler == nil {
		t.Fatal("der Fehler ging verloren")
	}

	if antwort.Decks == nil {
		t.Error("die Liste ist null")
	}
}

// **Ein Umbenennen liest zuerst — sonst verlöre es die Formatwahl.**
//
// `Aendern` schreibt den ganzen Datensatz. Wer nur Name und Notiz mitgäbe,
// setzte `format_id` still auf null, und niemand sähe es, bis jemand prüft.
func TestEinUmbenennenBehaeltDasFormat(t *testing.T) {
	anwendung, decks, _, _, _ := deckanwendung(t)

	decks.deck = domain.Deck{ID: 7, Fassung: 1, Name: "Alt", Spiel: "magic", FormatID: 5}

	antwort := anwendung.DeckUmbenennen(t.Context(), "7", "1", "Neu", "Notiz")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if antwort.Deck.Name != "Neu" || antwort.Deck.FormatID != "5" {
		t.Errorf("%+v", antwort.Deck)
	}
}

// **Eine leere Formatkennung nimmt das Format herunter — und ist kein Befund.**
func TestEineLeereFormatkennungNimmtHerunter(t *testing.T) {
	anwendung, decks, _, _, _ := deckanwendung(t)

	decks.deck = domain.Deck{ID: 7, Name: "Aggro", Spiel: "magic", FormatID: 5}

	antwort := anwendung.DeckFormatWaehlen(t.Context(), "7", "")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if decks.gewaehlt != 0 {
		t.Errorf("gewählt wurde %d, erwartet 0", decks.gewaehlt)
	}
}

// **Eine leere Kartenkennung ist kein Befund** — ein Druck bringt keine mit.
func TestEineLeereKartenkennungIstKeinBefund(t *testing.T) {
	anwendung, decks, _, _, _ := deckanwendung(t)

	antwort := anwendung.DeckplatzLegen(t.Context(), "7", "", "a1b2", 4, "hauptdeck")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if decks.platz.KarteID != 0 || decks.platz.Printing != "a1b2" {
		t.Errorf("%+v", decks.platz)
	}

	if antwort.Platz.ID != "3" || antwort.Platz.KarteID != "" {
		t.Errorf("%+v", antwort.Platz)
	}
}

// **Jede unbrauchbare Kennung nennt ihr Feld** (§8).
func TestJedeUnbrauchbareDeckkennungNenntIhrFeld(t *testing.T) {
	anwendung, _, _, _, _ := deckanwendung(t)

	for _, fall := range []struct {
		name string
		ruf  func() *app.Fehler
		feld string
	}{
		{"Deck lesen", func() *app.Fehler {
			return anwendung.Deck(t.Context(), "x").Fehler
		}, "deckID"},
		{"Deck löschen", func() *app.Fehler {
			return anwendung.DeckLoeschen(t.Context(), "x").Fehler
		}, "deckID"},
		{"umbenennen — Deck", func() *app.Fehler {
			return anwendung.DeckUmbenennen(t.Context(), "x", "1", "N", "").Fehler
		}, "deckID"},
		{"umbenennen — Fassung", func() *app.Fehler {
			return anwendung.DeckUmbenennen(t.Context(), "7", "x", "N", "").Fehler
		}, "fassung"},
		{"Platz legen — Deck", func() *app.Fehler {
			return anwendung.DeckplatzLegen(t.Context(), "x", "", "a", 1, "hauptdeck").Fehler
		}, "deckID"},
		{"Platz legen — Karte", func() *app.Fehler {
			return anwendung.DeckplatzLegen(t.Context(), "7", "x", "", 1, "hauptdeck").Fehler
		}, "karteID"},
		{"Platz nehmen — Deck", func() *app.Fehler {
			return anwendung.DeckplatzNehmen(t.Context(), "x", "1").Fehler
		}, "deckID"},
		{"Platz nehmen — Platz", func() *app.Fehler {
			return anwendung.DeckplatzNehmen(t.Context(), "7", "x").Fehler
		}, "platzID"},
		{"Platz verschieben — Deck", func() *app.Fehler {
			return anwendung.DeckplatzVerschieben(t.Context(), "x", "1", "sideboard").Fehler
		}, "deckID"},
		{"Platz verschieben — Platz", func() *app.Fehler {
			return anwendung.DeckplatzVerschieben(t.Context(), "7", "x", "sideboard").Fehler
		}, "platzID"},
		{"Format wählen — Deck", func() *app.Fehler {
			return anwendung.DeckFormatWaehlen(t.Context(), "x", "5").Fehler
		}, "deckID"},
		{"Format wählen — Format", func() *app.Fehler {
			return anwendung.DeckFormatWaehlen(t.Context(), "7", "x").Fehler
		}, "formatID"},
		{"prüfen", func() *app.Fehler {
			return anwendung.DeckPruefen(t.Context(), "x").Fehler
		}, "deckID"},
		{"abgleichen", func() *app.Fehler {
			return anwendung.DeckAbgleichen(t.Context(), "x").Fehler
		}, "deckID"},
		{"Format löschen", func() *app.Fehler {
			return anwendung.FormatLoeschen(t.Context(), "x").Fehler
		}, "formatID"},
	} {
		t.Run(fall.name, func(t *testing.T) {
			fehler := fall.ruf()
			if fehler == nil {
				t.Fatal("die unbrauchbare Kennung kam durch")
			}

			if fehler.Feld != fall.feld {
				t.Errorf("Feld %q, erwartet %q", fehler.Feld, fall.feld)
			}
		})
	}
}

// **Und jeder Weg meldet den Fehler seiner Naht.**
func TestJederDeckwegMeldetSeinenNahtfehler(t *testing.T) {
	anwendung, decks, formate, pruefung, abgleich := deckanwendung(t)

	kaputt := errors.New("die Naht schweigt")
	decks.fehler, formate.fehler = kaputt, kaputt
	pruefung.fehler, abgleich.fehler = kaputt, kaputt

	for name, ruf := range map[string]func() *app.Fehler{
		"Decks":           func() *app.Fehler { return anwendung.Decks(t.Context(), "").Fehler },
		"Deck":            func() *app.Fehler { return anwendung.Deck(t.Context(), "7").Fehler },
		"DeckAnlegen":     func() *app.Fehler { return anwendung.DeckAnlegen(t.Context(), "A", "magic", "").Fehler },
		"DeckUmbenennen":  func() *app.Fehler { return anwendung.DeckUmbenennen(t.Context(), "7", "1", "A", "").Fehler },
		"DeckLoeschen":    func() *app.Fehler { return anwendung.DeckLoeschen(t.Context(), "7").Fehler },
		"DeckplatzLegen":  func() *app.Fehler { return anwendung.DeckplatzLegen(t.Context(), "7", "", "a", 1, "hauptdeck").Fehler },
		"DeckplatzNehmen": func() *app.Fehler { return anwendung.DeckplatzNehmen(t.Context(), "7", "1").Fehler },
		"DeckplatzVerschieben": func() *app.Fehler {
			return anwendung.DeckplatzVerschieben(t.Context(), "7", "1", "sideboard").Fehler
		},
		"DeckFormatWaehlen": func() *app.Fehler { return anwendung.DeckFormatWaehlen(t.Context(), "7", "5").Fehler },
		"DeckPruefen":       func() *app.Fehler { return anwendung.DeckPruefen(t.Context(), "7").Fehler },
		"DeckAbgleichen":    func() *app.Fehler { return anwendung.DeckAbgleichen(t.Context(), "7").Fehler },
		"Formate":           func() *app.Fehler { return anwendung.Formate(t.Context(), "").Fehler },
		"FormatAusBauart": func() *app.Fehler {
			return anwendung.FormatAusBauart(t.Context(), "konstruiert", "S", "magic", "").Fehler
		},
		"FormatLoeschen": func() *app.Fehler { return anwendung.FormatLoeschen(t.Context(), "5").Fehler },
	} {
		t.Run(name, func(t *testing.T) {
			if ruf() == nil {
				t.Error("der Fehler ging verloren")
			}
		})
	}
}
