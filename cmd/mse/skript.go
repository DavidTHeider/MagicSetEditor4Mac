// **Ein Ausdruck im Kontext einer Karte** —
// #1070.
//
// §15 nennt MSEs eigene Kommandozeile als **Orakel** dieses Projekts:
//
//	magicseteditor --cli <SET>.mse-set -q --raw
//
// Dieser Befehl ist die Antwort darauf — dieselbe Frage, an dieses Programm
// gestellt, und damit vergleichbar.
package main

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// skriptbefehl wertet einen Ausdruck aus.
//
// **Eine Karte ist Pflicht.** MSEs Konsole arbeitet an einem Set und kennt
// „die aufgeschlagene Karte" aus dem Fenster; ohne Fenster gibt es keine, und
// eine erfundene wäre eine Umgebung, die niemand nachvollziehen kann.
func skriptbefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) < 2 {
		an.klage("skript braucht die Kartenkennung und den Ausdruck.\n%s", hilfe)

		return CodeBedienung
	}

	karteID, gut := kennung(an, "die Kartenkennung", rest[0])
	if !gut {
		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		naht, err := app.SkriptnahtUeber(db)
		if err != nil {
			an.klageUeber(err)

			return CodeFachlich
		}

		ergebnis, auswertfehler := naht.Auswerten(ctx, karteID, rest[1])
		if auswertfehler != nil {
			an.klageUeber(auswertfehler)

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, ergebnis)
		}

		// **Der Quelltext, nicht der Text** — so zeigt MSE es
		// (`src/cli/cli_main.cpp`, Zeile 234, Fassung f9d9356 vom
		// 2021-07-06).
		an.sag("%s\n", ergebnis.Quelltext)

		// **Die Befunde auf den Fehlerstrom**, nicht in die Ausgabe: wer
		// das Ergebnis weiterverarbeitet, will den Vorbehalt nicht darin
		// haben.
		for _, befund := range ergebnis.Befunde {
			an.klage("%s", befund)
		}

		return an.abschluss()
	})
}

// namensbefehl zählt auf, was ein Ausdruck an einer Karte nennen darf
// (#1164).
//
// **Ein Name je Zeile.** Wer sie weiterverarbeitet, hat sie damit schon
// zerlegt; wer sie liest, sucht in einer Spalte schneller als in einer Reihe.
// Mit `--json` kommt die Liste als Feld.
func namensbefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) < 1 {
		an.klage("skriptnamen braucht die Kartenkennung.\n%s", hilfe)

		return CodeBedienung
	}

	karteID, gut := kennung(an, "die Kartenkennung", rest[0])
	if !gut {
		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		naht, err := app.SkriptnahtUeber(db)
		if err != nil {
			an.klageUeber(err)

			return CodeFachlich
		}

		namen, namensfehler := naht.Namen(ctx, karteID)
		if namensfehler != nil {
			an.klageUeber(namensfehler)

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, map[string]any{"namen": namen})
		}

		for _, name := range namen {
			an.sag("%s\n", name)
		}

		return an.abschluss()
	})
}
