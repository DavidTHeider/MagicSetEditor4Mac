package main

// **Eine Leitung ohne Fenster** —
// #1634,
// entschieden in
// [0121](../../docs/entscheidungen/0121-eine-leitung-ohne-fenster.md).
//
// # Warum ein Befehl, der bleibt
//
// **Weil dieses Programm zwei davon ausliefert.** `mse-rest` hält einen
// Webdienst, `mse-mcp` spricht über die Standardeingabe; beide laufen, bis
// jemand sie beendet. Der Satz „ein Prozess, der bis ⌃C daran sitzt, ist kein
// Befehl" stand fünfmal in `docs/FAEHIGKEITEN.md` und war zweimal widerlegt.
//
// **Nichts horcht von selbst.** Die Zusage aus `N2` bleibt: ohne bewusste
// Auswahl geschieht nichts — und ein Befehl, den jemand tippt, **ist** eine.
//
// # Und warum die wirkliche Adresse ausgegeben wird
//
// **Weil `:0` einen Port vom Betriebssystem bekommt.** Wer sie nicht nennte,
// ließe den Menschen daneben raten, was er seinem Gegenüber sagen soll.

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// zustellungssicht ist, was die Kommandozeile über eine Zustellung schreibt.
//
// **Eine eigene Gestalt** — die Wails-DTO ist **innen**, diese Fläche
// **außen** (0033, Regel 5). Dieselbe Trennung wie bei `belegsicht`: jene
// trägt ein Fehlerfeld, das hier nie gefüllt sein kann, weil ein Fehler
// vorher über den Rückgabewert hinausgeht.
type zustellungssicht struct {
	Gegenueber string `json:"gegenueber"`
	Getragen   int    `json:"getragen"`
}

// horchenbefehl fährt `mse horchen <adresse>` — bis ⌃C.
func horchenbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 1 {
		an.klage("horchen nimmt die Adresse, unter der gehorcht wird.\n%s",
			hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an,
		func(ctx context.Context, db *sql.DB) int {
			flaeche := app.PeerflaecheUeber(db)

			antwort := flaeche.HorchStarten(ctx, rest[0])
			if antwort.Fehler != nil {
				an.klageUeber(antwort.Fehler.AlsFehler())

				return CodeFachlich
			}

			// **Die Adresse geht auf die Standardausgabe** — sie ist das
			// Ergebnis dieses Befehls; alles Weitere ist Protokoll und
			// gehört auf die Fehlerausgabe (0121).
			an.sag("%s\n", antwort.Stand.Adresse)

			// **Und dann wird gewartet.** Beendet wird über ⌃C oder
			// SIGTERM; `main` hängt beides an diesen Kontext.
			<-ctx.Done()

			// **Ohne den abgebrochenen Kontext** — er ist der Grund, warum
			// wir hier stehen, und ein Beenden über ihn käme gar nicht mehr
			// bis zum Lauscher.
			if beendet := flaeche.HorchBeenden(
				context.WithoutCancel(ctx)); beendet.Fehler != nil {
				an.klageUeber(beendet.Fehler.AlsFehler())

				return CodeFachlich
			}

			return an.abschluss()
		})
}

// handelszustellungbefehl fährt `mse handelszustellung <kennung>`.
//
// **Er bleibt nicht.** Er baut eine Leitung, trägt hinüber, was im Ausgang
// liegt, und endet — die Gegenrichtung zum Befehl darüber.
func handelszustellungbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 1 {
		an.klage("handelszustellung nimmt die Kennung des Gegenübers.\n%s",
			hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an,
		func(ctx context.Context, db *sql.DB) int {
			bericht := app.PeerflaecheUeber(db).Handelszustellung(ctx, rest[0])
			if bericht.Fehler != nil {
				an.klageUeber(bericht.Fehler.AlsFehler())

				return CodeFachlich
			}

			sicht := zustellungssicht{
				Gegenueber: bericht.Gegenueber, Getragen: bericht.Getragen,
			}

			if an.alsJSON {
				return alsJSONSchreiben(an, sicht)
			}

			an.sag("%s\t%d\n", sicht.Gegenueber, sicht.Getragen)

			return an.abschluss()
		})
}
