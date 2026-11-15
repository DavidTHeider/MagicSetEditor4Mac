// **Eine Deckliste an der Kommandozeile einlesen** — Werkstück `T8` von
// Phase 5.6 (#1473).
//
// §16: dieselbe Fähigkeit, überall dieselbe Antwort.
//
// # Zwei Befehle, weil es zwei Schritte sind
//
// `deck-liste` liest von der Standardeingabe und **schreibt nichts** — es
// sagt, was ins Deck käme, und nennt den **Stand**, an den es gebunden ist.
// `deck-liste-uebernehmen` legt, was eindeutig ist, und verlangt denselben
// Stand zurück.
//
// # Warum der Stand über die Kommandozeile geht
//
// **Weil es sonst keine Bindung gäbe.** Zwei Rufe an der Kommandozeile sind
// zwei Prozesse; dazwischen kann jemand am Fenster eine Karte hineinlegen.
// Wer den Stand mitgibt, bekommt einen Konflikt statt einer Deckliste über
// einer anderen (§7, Regel 2).
//
// **Und wer ihn ausdrücklich nicht will**, schreibt `-` und übernimmt damit
// den Stand, der gerade da ist — ein Befehl in einem Skript hat keine
// Vorschau gesehen. Das ist eine bewusste Aufgabe der Bindung und deshalb
// sichtbar.
package main

import (
	"context"
	"database/sql"
	"fmt"
	"io"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// derFreieStand hebt die Bindung an eine Vorschau auf.
const derFreieStand = "-"

// deckListebefehl deutet eine Deckliste für ein Deck.
func deckListebefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 1 {
		an.klage("deck-liste nimmt die Kennung des Decks.\n%s", hilfe)

		return CodeBedienung
	}

	deckID, gelesen := kennung(an, "das Deck", rest[0])
	if !gelesen {
		return CodeBedienung
	}

	text, lesefehler := io.ReadAll(an.eingabe)
	if lesefehler != nil {
		an.klageUeber(fmt.Errorf("die Liste lesen: %w", lesefehler))

		return CodeBedienung
	}

	return mitDeckeinfuhr(ctx, bestandsort, an,
		func(ctx context.Context, naht *app.Deckeinfuhrnaht) int {
			vorschau, err := naht.Vorschau(ctx, deckID, string(text))
			if err != nil {
				an.klageUeber(err)

				return CodeFachlich
			}

			return deckvorschauSchreiben(an, vorschau)
		})
}

// deckvorschauSchreiben gibt eine Deckvorschau aus.
func deckvorschauSchreiben(an *schreiber, vorschau *app.Deckvorschau) int {
	if an.alsJSON {
		return alsJSONSchreiben(an, vorschau)
	}

	// **Der Stand zuerst** — wer übernehmen will, braucht ihn, und er soll
	// nicht hinter dreihundert Zeilen stehen.
	an.sag("stand\t%s\n", vorschau.Stand)
	an.sag("hauptdeck\t%d\nsideboard\t%d\n",
		vorschau.Hauptdeck, vorschau.Sideboard)

	for _, befund := range vorschau.Befunde {
		an.sag("!\t%s\n", befund)
	}

	for _, zeile := range vorschau.Zeilen {
		if !zeile.Uebernehmbar {
			an.sag("?\t%s\t%s\n", zeile.Name, zeile.Befund)

			continue
		}

		an.sag("=\t%s\t%d\t%s\t%s\n", zeile.Bereich, zeile.Menge,
			zeile.Printing, zeile.Name)
	}

	return an.abschluss()
}

// deckListeUebernehmenbefehl legt die eindeutigen Zeilen ins Deck.
func deckListeUebernehmenbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 2 {
		an.klage("deck-liste-uebernehmen nimmt die Kennung des Decks und "+
			"den Stand aus der Vorschau (»-« für den gerade geltenden).\n%s",
			hilfe)

		return CodeBedienung
	}

	deckID, gelesen := kennung(an, "das Deck", rest[0])
	if !gelesen {
		return CodeBedienung
	}

	text, lesefehler := io.ReadAll(an.eingabe)
	if lesefehler != nil {
		an.klageUeber(fmt.Errorf("die Liste lesen: %w", lesefehler))

		return CodeBedienung
	}

	return mitDeckeinfuhr(ctx, bestandsort, an,
		func(ctx context.Context, naht *app.Deckeinfuhrnaht) int {
			return deckListeLegen(ctx, an, naht, deckID, rest[1], string(text))
		})
}

// deckListeLegen deutet, wählt das Eindeutige und legt es.
func deckListeLegen(
	ctx context.Context, an *schreiber, naht *app.Deckeinfuhrnaht,
	deckID int64, stand, text string,
) int {
	vorschau, err := naht.Vorschau(ctx, deckID, text)
	if err != nil {
		an.klageUeber(err)

		return CodeFachlich
	}

	zeilen, offen := uebernehmbareZeilen(vorschau)

	if len(zeilen) == 0 {
		an.sag("keine übernehmbare Zeile — %d offen\n", offen)

		return an.abschluss()
	}

	if stand == derFreieStand {
		stand = vorschau.Stand
	}

	bericht, err := naht.Uebernehmen(ctx, deckID, stand, zeilen)
	if err != nil {
		an.klageUeber(err)

		return CodeFachlich
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, map[string]any{
			"gelegt": bericht.Gelegt, "stueck": bericht.Stueck,
			"stand": bericht.Stand, "offen": offen,
		})
	}

	an.sag("%d Plätze, %d Karten, %d offen\nstand\t%s\n",
		bericht.Gelegt, bericht.Stueck, offen, bericht.Stand)

	return an.abschluss()
}

// uebernehmbareZeilen nimmt aus der Vorschau, was ohne Rückfrage geht.
//
// **Und sagt, wie viele es nicht sind.** Eine Einfuhr, die 40 von 300 Zeilen
// legt und „fertig" meldet, ist eine stille Kappung (§10).
func uebernehmbareZeilen(vorschau *app.Deckvorschau) (
	zeilen []app.Deckeinfuhrzeile, offen int,
) {
	zeilen = []app.Deckeinfuhrzeile{}

	for _, zeile := range vorschau.Zeilen {
		if !zeile.Uebernehmbar {
			offen++

			continue
		}

		zeilen = append(zeilen, app.Deckeinfuhrzeile{
			Printing: zeile.Printing, Menge: zeile.Menge,
			Bereich: zeile.Bereich,
		})
	}

	return zeilen, offen
}

// mitDeckeinfuhr baut die Naht über dem Bestand und führt die Arbeit aus.
func mitDeckeinfuhr(
	ctx context.Context, bestandsort string, an *schreiber,
	arbeit func(context.Context, *app.Deckeinfuhrnaht) int,
) int {
	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		naht, err := app.DeckeinfuhrUeber(db)
		if err != nil {
			an.klageUeber(err)

			return CodeFachlich
		}

		return arbeit(ctx, naht)
	})
}
