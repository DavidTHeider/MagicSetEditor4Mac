// **Eine Kartenliste an der Kommandozeile einlesen** — Werkstück 2 von
// Phase 4.5 (#1292).
//
// §16: dieselbe Fähigkeit, überall dieselbe Antwort.
//
// # Zwei Befehle, weil es zwei Schritte sind
//
// `liste-deuten` liest von der Standardeingabe und **schreibt nichts** — es
// sagt, welche Zeile welchen Druck meint und welche offen ist.
// `liste-einlesen` schreibt, was eindeutig ist.
//
// **Der Grund steht in §16, Regel 6**: nichts wird still übernommen. Wer
// „Counterspell" einliest, meint einen von 72 Drucken, und welchen, weiß nur
// er selbst.
package main

import (
	"context"
	"database/sql"
	"fmt"
	"io"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// listeDeutenbefehl liest eine Liste von der Standardeingabe und deutet sie.
func listeDeutenbefehl(ctx context.Context, bestandsort string, an *schreiber) int {
	text, fehler := io.ReadAll(an.eingabe)
	if fehler != nil {
		an.klageUeber(fmt.Errorf("die Liste lesen: %w", fehler))

		return CodeBedienung
	}

	return mitEinfuhr(ctx, bestandsort, an, func(ctx context.Context, naht *app.Listeneinfuhrnaht) int {
		deutung, err := naht.Deuten(ctx, string(text))
		if err != nil {
			an.klageUeber(err)

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, deutung)
		}

		for _, befund := range deutung.Befunde {
			an.sag("!\t%s\n", befund)
		}

		for _, zeile := range deutung.Zeilen {
			an.sag("%s\t%s\n", zeilenlage(zeile), zeile.Roh)

			// **Die Auswahl steht darunter, eingerückt.** Wer sie nicht
			// braucht, überliest sie; wer wählen muss, hat die Kennungen da.
			if !zeile.Eindeutig {
				for _, treffer := range zeile.Treffer {
					an.sag("\t\t%s\t%s\t%s\t#%s\n", treffer.Printing,
						treffer.Name, treffer.Setcode, treffer.Nummer)
				}
			}
		}

		return an.abschluss()
	})
}

// zeilenlage gibt die Marke am Anfang einer Zeile.
//
// **Drei Zeichen und keine Sätze.** Eine Liste von dreihundert Zeilen liest
// sich mit einer Spalte; mit einem Satz je Zeile liest sie sich nicht.
func zeilenlage(zeile app.Listendeutungszeile) string {
	switch {
	case zeile.Befund != "":
		return "!"

	case zeile.Eindeutig:
		return "="

	default:
		return "?"
	}
}

// listeEinlesenbefehl liest eine Liste und übernimmt, was eindeutig ist.
//
// # Warum nur das Eindeutige
//
// **Weil der Rest eine Entscheidung ist und keine Auslassung** (§16, Regel 6).
// Was offen bleibt, steht danach als Zahl da — und `liste-deuten` sagt, welche
// Zeilen es sind und welche Drucke zur Wahl stehen.
//
// **Wer eine offene Zeile eintragen will**, wählt ihren Druck und nimmt
// `exemplar <druck> <menge>`. Das ist derselbe Weg, den es vorher schon gab —
// nur muss die Kennung jetzt nicht mehr von Hand gefunden werden.
func listeEinlesenbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	rest, zusammenfuehrung := zusammenfuehrungAus(rest)

	text, fehler := io.ReadAll(an.eingabe)
	if fehler != nil {
		an.klageUeber(fmt.Errorf("die Liste lesen: %w", fehler))

		return CodeBedienung
	}

	zustand := domain.ZustandNearMint
	if len(rest) > 0 {
		zustand = domain.Zustand(rest[0])
	}

	return mitEinfuhr(ctx, bestandsort, an, func(ctx context.Context, naht *app.Listeneinfuhrnaht) int {
		deutung, err := naht.Deuten(ctx, string(text))
		if err != nil {
			an.klageUeber(err)

			return CodeFachlich
		}

		zeilen, offen := eindeutigeZeilen(deutung, zustand)

		if len(zeilen) == 0 {
			an.sag("keine eindeutige Zeile — %d offen\n", offen)

			return an.abschluss()
		}

		bericht, err := naht.Einlesen(ctx, zeilen, zusammenfuehrung)
		if err != nil {
			an.klageUeber(err)

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, map[string]any{
				"angelegt": bericht.Angelegt, "erhoeht": bericht.Erhoeht,
				"ausgelassen": bericht.Ausgelassen,
				"stueck":      bericht.Stueck, "offen": offen,
				"befunde": bericht.Befunde,
			})
		}

		an.sag("%d angelegt, %d erhöht, %d Stück, %d offen%s\n",
			len(bericht.Angelegt), bericht.Erhoeht, bericht.Stueck, offen, ausgelassenText(bericht))

		for _, befund := range bericht.Befunde {
			an.sag("!\t%s\n", befund)
		}

		return an.abschluss()
	})
}

// eindeutigeZeilen nimmt aus einer Deutung, was ohne Rückfrage geht.
//
// **Und sagt, wie viele es nicht sind.** Eine Einfuhr, die 40 von 300 Zeilen
// schreibt und „fertig" meldet, ist eine stille Kappung (§10).
func eindeutigeZeilen(deutung *app.Listendeutung, zustand domain.Zustand) (
	zeilen []app.Listeneinfuhrzeile, offen int,
) {
	zeilen = []app.Listeneinfuhrzeile{}

	for _, zeile := range deutung.Zeilen {
		if !zeile.Eindeutig {
			offen++

			continue
		}

		treffer := zeile.Treffer[0]

		zeilen = append(zeilen, app.Listeneinfuhrzeile{
			Printing: treffer.Printing,
			Menge:    zeile.Menge,
			Zustand:  string(zustand),
			Sprache:  domain.Vorgabesprache,
			Finish:   zeile.Finish,
			Name:     treffer.Name,
			Setcode:  treffer.Setcode,
			Nummer:   treffer.Nummer,
		})
	}

	return zeilen, offen
}

// mitEinfuhr baut die Naht über dem Bestand und führt die Arbeit aus.
func mitEinfuhr(
	ctx context.Context, bestandsort string, an *schreiber,
	arbeit func(context.Context, *app.Listeneinfuhrnaht) int,
) int {
	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		naht, err := app.ListeneinfuhrUeber(db)
		if err != nil {
			an.klageUeber(err)

			return CodeFachlich
		}

		return arbeit(ctx, naht)
	})
}
