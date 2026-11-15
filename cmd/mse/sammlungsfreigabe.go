package main

// **Die freigegebenen Sichten an der Kommandozeile** — Werkstück `T3` von
// Phase 5.6 (#1468).
//
//	mse kontakt-freigaben <kennung>             was einem Gegenüber gilt
//	mse kontakt-freigeben <kennung> <zweck> <umfang> [behalten|druck…]
//	                                            eine Sicht erteilen
//	mse kontakt-widerrufen <kennung> <zweck> <umfang>
//	                                            sie zurücknehmen
//	mse kontakt-sicht <kennung> <zweck>         was es wirklich sähe
//
// **`kontakt-` und nicht `freigabe-`**: `freigaben` gibt es seit Phase 4.12
// und meint die Freigabe eines **Profils** an einen zweiten Nutzer dieses
// Rechners.
//
// **Die Vorschau ist der Punkt.** „Deine Preise gehen nicht hinaus" ist ein
// Satz; `mse freigabesicht` ist der Blick darauf — und er geht durch denselben
// Weg wie die Antwort an das Gegenüber.

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// sichtfreigabenbefehl nennt, was einem Gegenüber gilt.
func sichtfreigabenbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	return sichtfreigabelauf(ctx, bestandsort, an, rest, "kontakt-freigaben", 1,
		func(ctx context.Context, anwendung *app.App, teile []string) app.SammlungsfreigabenAntwort {
			return anwendung.Sammlungsfreigaben(ctx, teile[0])
		})
}

// sichtfreigebenbefehl erteilt eine Sicht.
//
// **Die vierte Stelle ist die Behaltemenge oder die Auswahl** — je nach
// Umfang, und das ist keine Doppeldeutigkeit: bei `dubletten` ist es eine
// Zahl, bei `auswahl` sind es Druckkennungen, und bei den drei anderen
// weist die Domäne beides ab.
func sichtfreigebenbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) < 3 {
		an.klage("kontakt-freigeben nimmt mindestens drei Argumente.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		return schreibeSichtfreigaben(an, app.PeerflaecheUeber(db).SammlungFreigeben(
			ctx, sichtfreigabeAus(rest)))
	})
}

// sichtfreigabewiderrufbefehl nimmt eine Sicht zurück.
func sichtfreigabewiderrufbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	return sichtfreigabelauf(ctx, bestandsort, an, rest, "kontakt-widerrufen", 3,
		func(ctx context.Context, anwendung *app.App, teile []string) app.SammlungsfreigabenAntwort {
			return anwendung.SammlungsfreigabeWiderrufen(
				ctx, teile[0], teile[1], teile[2])
		})
}

// sichtvorschaubefehl zeigt, was ein Gegenüber wirklich sähe.
func sichtvorschaubefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 2 {
		an.klage("kontakt-sicht nimmt zwei Argumente.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		antwort := app.PeerflaecheUeber(db).Freigabesicht(ctx, rest[0], rest[1])
		if antwort.Fehler != nil {
			an.klageUeber(antwort.Fehler.AlsFehler())

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, antwort)
		}

		// **Die letzte Spalte sagt, was die Zeile ist** — `angebot` oder
		// `wunsch` (Befund 7 des Zweitblicks über Phase 8): ohne sie sah ein
		// Wunsch aus wie ein Angebot desselben Drucks. Angehängt, damit jede
		// Spalte davor bleibt, wo sie war — erweiternd (§16).
		for _, zeile := range antwort.Zeilen {
			art := "angebot"
			if zeile.Wunsch {
				art = "wunsch"
			}

			an.sag("%s\t%d\t%s\t%s\t%s\t%s\t%s\n", zeile.Referenz, zeile.Menge,
				zeile.Zustand, zeile.Sprache, zeile.Finish, zeile.Bezeichnung, art)
		}

		return an.abschluss()
	})
}

// sichtanfragebefehl fährt `mse sicht-anfragen <kennung>` — P11 (#2379):
// fragt ein Gegenüber über die Leitung nach dem, was es diesem Profil zum
// Handeln freigibt. **Dieselbe Ausgabe wie `kontakt-sicht`**, denn es ist
// dieselbe Gestalt von der anderen Seite.
func sichtanfragebefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 1 {
		an.klage("sicht-anfragen nimmt die Kennung des Gegenübers.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		antwort := app.PeerflaecheUeber(db).HandelSichtAnfragen(ctx, rest[0])
		if antwort.Fehler != nil {
			an.klageUeber(antwort.Fehler.AlsFehler())

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, antwort)
		}

		// **Die letzte Spalte sagt, was die Zeile ist** — `angebot` oder
		// `wunsch` (Befund 7 des Zweitblicks über Phase 8): ohne sie sah ein
		// Wunsch aus wie ein Angebot desselben Drucks. Angehängt, damit jede
		// Spalte davor bleibt, wo sie war — erweiternd (§16).
		for _, zeile := range antwort.Zeilen {
			art := "angebot"
			if zeile.Wunsch {
				art = "wunsch"
			}

			an.sag("%s\t%d\t%s\t%s\t%s\t%s\t%s\n", zeile.Referenz, zeile.Menge,
				zeile.Zustand, zeile.Sprache, zeile.Finish, zeile.Bezeichnung, art)
		}

		return an.abschluss()
	})
}

// sichtfreigabeAus baut die Freigabe aus den Stellungsargumenten.
//
// **Eine Zahl an der vierten Stelle ist die Behaltemenge, alles andere sind
// Drucke.** Der Umfang entscheidet, was gilt; die Domäne weist ab, was nicht
// zusammenpasst — hier wird nur gelesen.
func sichtfreigabeAus(rest []string) app.SammlungsfreigabeDTO {
	freigabe := app.SammlungsfreigabeDTO{
		Empfaenger: rest[0], Zweck: rest[1], Umfang: rest[2],
		Auswahl: []string{},
	}

	for _, weiteres := range rest[3:] {
		if zahl, fehler := strconv.Atoi(weiteres); fehler == nil {
			freigabe.Behalten = zahl

			continue
		}

		freigabe.Auswahl = append(freigabe.Auswahl, weiteres)
	}

	return freigabe
}

// sichtfreigabelauf fährt einen Weg und schreibt die Liste.
func sichtfreigabelauf(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
	name string, argumente int,
	fahre func(context.Context, *app.App, []string) app.SammlungsfreigabenAntwort,
) int {
	if len(rest) != argumente {
		an.klage("%s nimmt %d Argumente.\n%s", name, argumente, hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		return schreibeSichtfreigaben(an, fahre(ctx, app.PeerflaecheUeber(db), rest))
	})
}

// schreibeSichtfreigaben gibt die Liste aus — als Text oder als JSON.
func schreibeSichtfreigaben(an *schreiber, antwort app.SammlungsfreigabenAntwort) int {
	if antwort.Fehler != nil {
		an.klageUeber(antwort.Fehler.AlsFehler())

		return CodeFachlich
	}

	nachladenMelden(an, antwort.Nachladefehler)

	if an.alsJSON {
		return alsJSONSchreiben(an, antwort)
	}

	for _, freigabe := range antwort.Freigaben {
		an.sag("%s\t%s\t%s\n", freigabe.Zweck, freigabe.Umfang,
			sichtbefristung(freigabe))

		for _, printing := range freigabe.Auswahl {
			an.sag("\tdruck\t%s\n", printing)
		}
	}

	return an.abschluss()
}

// sichtbefristung sagt, bis wann eine Freigabe gilt.
//
// **„unbefristet" und nicht eine leere Spalte**: eine Freigabe ohne Ablauf ist
// die gewöhnliche, und eine Lücke an dieser Stelle läse sich wie ein Fehler.
func sichtbefristung(freigabe app.SammlungsfreigabeDTO) string {
	if freigabe.Ablauf == "" {
		return "unbefristet"
	}

	return "bis " + freigabe.Ablauf
}
