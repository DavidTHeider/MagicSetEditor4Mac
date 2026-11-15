package main

// **Die eigene Peeridentität an der Kommandozeile** — Werkstück N1 von
// Phase 5.4
// (#1454).
//
// `mse peer` sagt, womit sich dieses Profil ausweist; `mse peer-anlegen`
// legt die erste an, `mse peer-rotieren` löst eine bestehende ab.
//
// **Drei Befehle und nicht einer mit Schaltern**: fragen, anlegen und
// ablösen sind drei Handlungen, und die letzte verlangt von jedem bekannten
// Kontakt eine neue Prüfung. Ein `--rotieren` an einem Lesebefehl wäre die
// Gestalt, in der jemand das eine tippt und das andere bekommt.

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// peerbefehl nennt die eigene Identität.
func peerbefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	return peerlauf(ctx, bestandsort, an, rest, "peer",
		func(ctx context.Context, anwendung *app.App) app.PeeridentitaetAntwort {
			return anwendung.Peeridentitaet(ctx)
		})
}

// peeranlegenbefehl legt die erste Identität an.
func peeranlegenbefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	return peerlauf(ctx, bestandsort, an, rest, "peer-anlegen",
		func(ctx context.Context, anwendung *app.App) app.PeeridentitaetAntwort {
			return anwendung.PeeridentitaetSicherstellen(ctx)
		})
}

// peerrotierenbefehl löst die bestehende Identität ab.
func peerrotierenbefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	return peerlauf(ctx, bestandsort, an, rest, "peer-rotieren",
		func(ctx context.Context, anwendung *app.App) app.PeeridentitaetAntwort {
			return anwendung.PeeridentitaetRotieren(ctx)
		})
}

// peerlauf fährt einen der drei Wege und schreibt das Ergebnis.
func peerlauf(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
	name string,
	fahre func(context.Context, *app.App) app.PeeridentitaetAntwort,
) int {
	if len(rest) != 0 {
		an.klage("%s nimmt keine Argumente.\n%s", name, hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		antwort := fahre(ctx, app.PeerflaecheUeber(db))
		if antwort.Fehler != nil {
			an.klageUeber(antwort.Fehler.AlsFehler())

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, antwort)
		}

		an.sag("stand\t%s\n", antwort.Stand)

		if antwort.Kennung != "" {
			an.sag("kennung\t%s\n", antwort.Kennung)
			an.sag("fingerabdruck\t%s\n", antwort.Fingerabdruck)
		}

		if antwort.Seit != "" {
			an.sag("seit\t%s\n", antwort.Seit)
		}

		if antwort.Vorher != "" {
			an.sag("vorher\t%s\t%s\n", antwort.Vorher, antwort.VorherFingerabdruck)
		}

		for _, befund := range antwort.Befunde {
			an.sag("befund\t%s\n", befund)
		}

		return an.abschluss()
	})
}
