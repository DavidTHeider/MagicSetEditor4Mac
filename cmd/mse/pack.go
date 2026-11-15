// **Das Zufallspack auf der Kommandozeile** —
// #889,
// Befund 7 des Zweitblicks über Phase W.
//
// Zwei Befehle: `packarten` zeigt, was ein Set anbietet, `pack` zieht. Die
// Trennung ist dieselbe wie bei den Schlüsselwörtern — zuerst sehen, dann tun.
package main

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// packartenbefehl zeigt die Packarten eines Sets.
func packartenbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) < 1 {
		an.klage("packarten braucht die Setkennung.\n%s", hilfe)

		return CodeBedienung
	}

	setID, gut := kennung(an, "die Setkennung", rest[0])
	if !gut {
		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		anwendung, err := app.PackUeber(db)
		if err != nil {
			an.klageUeber(err)

			return CodeFachlich
		}

		antwort, lesefehler := anwendung.Arten(ctx, setID)
		if lesefehler != nil {
			an.klageUeber(lesefehler)

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, antwort)
		}

		zeigePackarten(an, antwort)

		return an.abschluss()
	})
}

// packbefehl zieht ein Pack.
//
// **Der Same steht im Aufruf**, nicht in einer Uhr. Dieselbe Zahl gibt
// dieselbe Ziehung — ohne das könnte niemand ein Ergebnis noch einmal
// herstellen, und die Fläche gäbe eine andere Antwort als das Fenster (§16).
func packbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) < 3 {
		an.klage("pack braucht Set, Packart und Samen.\n%s", hilfe)

		return CodeBedienung
	}

	setID, gut := kennung(an, "die Setkennung", rest[0])
	if !gut {
		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		anwendung, err := app.PackUeber(db)
		if err != nil {
			an.klageUeber(err)

			return CodeFachlich
		}

		antwort, ziehfehler := anwendung.Ziehen(ctx, setID, rest[1], rest[2])
		if ziehfehler != nil {
			an.klageUeber(ziehfehler)

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, antwort)
		}

		zeigePackziehung(an, antwort)

		return an.abschluss()
	})
}

// zeigePackarten schreibt die Arten in lesbarer Form.
//
// **Die Zahl steht dabei, weil eine leere Packart nichts zieht** und niemand
// sähe, warum: ein Set ohne seltene Karten gibt kein Booster.
func zeigePackarten(an *schreiber, antwort app.PackAntwort) {
	for _, art := range antwort.Arten {
		waehlbar := "—"
		if art.Waehlbar {
			waehlbar = "wählbar"
		}

		an.sag("%s\t%d\t%s\n", art.Name, art.Karten, waehlbar)
	}

	zeigePackbefunde(an, antwort.Befunde)
}

// zeigePackziehung schreibt die gezogenen Karten in ihrer Ziehreihenfolge.
//
// **Die Reihenfolge ist die Auskunft** — so legt MSE das Pack hin, und wer
// sie sortierte, gäbe eine andere Antwort als das Fenster (§16).
func zeigePackziehung(an *schreiber, antwort app.PackziehungAntwort) {
	for _, karte := range antwort.Karten {
		an.sag("%s\t%s\n", karte.ID, packkartenname(karte))
	}

	zeigePackbefunde(an, antwort.Befunde)
}

// packkartenname gibt den Namen einer gezogenen Karte.
//
// **Leer ist kein Fehler**: eine Karte ohne `name` gibt es, und ein Platzhalter
// verschleierte, dass sie so im Bestand steht.
func packkartenname(karte app.KarteDTO) string {
	for _, wert := range karte.Werte {
		if wert.Feld == "name" {
			return wert.Wert
		}
	}

	return ""
}

// zeigePackbefunde nennt die Packarten, deren Filter nicht aufging.
//
// **Sie stehen da und nicht im Verborgenen** (§8): eine Packart, die still
// leer bleibt, sieht aus wie ein Set ohne passende Karten.
func zeigePackbefunde(an *schreiber, befunde []string) {
	for _, befund := range befunde {
		an.klage("%s\n", befund)
	}
}
