// **Decks an der Kommandozeile** — Werkstück 7 von Phase 4.6
// (#1293).
//
// Einordnung nach §16: **erweiternd** — neue Befehle, keine geänderten.

package main

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// deckbefehl bedient `mse decks` und `mse deck …`.
func deckbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	return mitDecks(ctx, bestandsort, an, func(ctx context.Context, flaeche *app.Deckflaeche) int {
		if len(rest) == 0 {
			return decklisteZeigen(ctx, flaeche, an, "")
		}

		switch rest[0] {
		case "neu":
			return deckAnlegen(ctx, flaeche, an, rest[1:])

		case "umbenennen":
			return deckUmbenennen(ctx, flaeche, an, rest[1:])

		case "loeschen":
			return deckLoeschen(ctx, flaeche, an, rest[1:])

		// **Die Kopie** — DOM-03.AK-02 (#1698).
		case "kopieren":
			return deckKopieren(ctx, flaeche, an, rest[1:])

		// **Aus einem Set ein Deck** — FMT-04 (#1693).
		case "aus-set":
			return deckAusSet(ctx, flaeche, an, rest[1:])

		case "wuenschen":
			return deckWuenschen(ctx, flaeche, an, rest[1:])

		case "legen":
			return deckplatzLegen(ctx, flaeche, an, rest[1:])

		case "nehmen":
			return deckplatzNehmen(ctx, flaeche, an, rest[1:])

		case "verschieben":
			return deckplatzVerschieben(ctx, flaeche, an, rest[1:])

		case "format":
			return deckFormatWaehlen(ctx, flaeche, an, rest[1:])

		case "pruefen":
			return deckPruefen(ctx, flaeche, an, rest[1:])

		case "bestand":
			return deckAbgleichen(ctx, flaeche, an, rest[1:])

		// **Die Zuordnung** — Werkstück U9 von Phase 5.1 (#1420).
		case "zuordnungen":
			return deckZuordnungen(ctx, flaeche, an, rest[1:])

		case "zuordnen":
			return deckZuordnen(ctx, flaeche, an, rest[1:])

		case "zuordnung-loesen":
			return deckZuordnungLoesen(ctx, flaeche, an, rest[1:])

		case "fehlkarten":
			return deckFehlkarten(ctx, flaeche, an, rest[1:])

		// **Das Deck als Liste** — NET-04 (#1708).
		case "liste":
			return deckListe(ctx, flaeche, an, rest[1:])
		}

		// **Ohne Unterbefehl ist das erste Wort eine Kennung** — `mse deck 7`
		// zeigt das Deck. Dasselbe Muster wie bei `stile` und `vorlagen`.
		return deckZeigen(ctx, flaeche, an, rest[0])
	})
}

// decksbefehl bedient `mse decks [spiel]`.
func decksbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	spiel := ""
	if len(rest) > 0 {
		spiel = rest[0]
	}

	return mitDecks(ctx, bestandsort, an, func(ctx context.Context, flaeche *app.Deckflaeche) int {
		return decklisteZeigen(ctx, flaeche, an, spiel)
	})
}

// decklisteZeigen nennt die Decks.
func decklisteZeigen(
	ctx context.Context, flaeche *app.Deckflaeche, an *schreiber, spiel string,
) int {
	antwort := flaeche.Decks(ctx, spiel)
	if antwort.Fehler != nil {
		return klageUeberNaht(an, antwort.Fehler)
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, antwort)
	}

	for _, deck := range antwort.Decks {
		an.sag("%s\t%s\t%s\t%d+%d\n",
			deck.ID, deck.Spiel, deck.Name, deck.Hauptdeck, deck.Sideboard)
	}

	return an.abschluss()
}

// deckZeigen gibt ein Deck samt seinen Plätzen.
func deckZeigen(
	ctx context.Context, flaeche *app.Deckflaeche, an *schreiber, kennung string,
) int {
	antwort := flaeche.Deck(ctx, kennung)
	if antwort.Fehler != nil {
		return klageUeberNaht(an, antwort.Fehler)
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, antwort)
	}

	an.sag("%s (%s), Fassung %s — %d im Hauptdeck, %d im Sideboard\n",
		antwort.Deck.Name, antwort.Deck.Spiel, antwort.Deck.Fassung,
		antwort.Deck.Hauptdeck, antwort.Deck.Sideboard)

	for _, platz := range antwort.Deck.Plaetze {
		bezug := platz.Printing
		if bezug == "" {
			bezug = "karte:" + platz.KarteID
		}

		an.sag("%s\t%d×\t%s\t%s\n", platz.ID, platz.Menge, platz.Bereich, bezug)
	}

	return an.abschluss()
}

// deckAnlegen legt ein leeres Deck an.
func deckAnlegen(
	ctx context.Context, flaeche *app.Deckflaeche, an *schreiber, rest []string,
) int {
	if len(rest) < 2 {
		an.klage("deck neu braucht einen Namen und ein Spiel.\n%s", hilfe)

		return CodeBedienung
	}

	notiz := ""
	if len(rest) > 2 {
		notiz = rest[2]
	}

	antwort := flaeche.DeckAnlegen(ctx, rest[0], rest[1], notiz)
	if antwort.Fehler != nil {
		return klageUeberNaht(an, antwort.Fehler)
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, antwort)
	}

	an.sag("%s\n", antwort.Deck.ID)

	return an.abschluss()
}

// mitDecks öffnet den Bestand und baut die Deckwege.
//
// **Ein fehlender Scryfall-Zugang ist hier kein Abbruch**: er trifft nur die
// Prüfung, und die meldet es selbst. Wer deswegen `mse decks` verweigerte,
// machte aus einem fehlenden Zuhause einen Totalausfall.
func mitDecks(
	ctx context.Context, bestandsort string, an *schreiber,
	arbeit func(context.Context, *app.Deckflaeche) int,
) int {
	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		// **Der Fehler betrifft nur die Prüfung** und wird bewusst nicht
		// gemeldet: `deck pruefen` sagt es selbst, und `mse decks` hätte
		// nichts davon.
		flaeche, _ := app.DeckflaecheUeber(db) //nolint:errcheck // siehe oben

		return arbeit(ctx, flaeche)
	})
}

// klageUeberNaht meldet den Befund einer Naht und gibt den Rückgabewert.
//
// **Ein fachliches Nein und kein Bedienfehler**: die Angaben waren lesbar,
// der Bestand sagt nein.
func klageUeberNaht(an *schreiber, fehler *app.Fehler) int {
	an.klageUeberFehler(fehler)

	return CodeFachlich
}
