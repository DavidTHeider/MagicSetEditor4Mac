package main

// **Die eigenen Partien an der Kommandozeile** —
// #1600.
//
// # Warum es den Weg auch hier gibt
//
// **Weil die Kennung einer Partie sonst nirgends steht** (§16). Wer über die
// Kommandozeile spielt, bekommt sie beim Anlegen einmal ausgegeben — und
// danach müsste er sie sich aufschreiben. Dieselbe Frage stellt sich am
// Fenster, und dort beantwortet sie eine Liste.
//
// **Eine eigene Datei**, weil `partie.go` mit ihr an die Zeilengrenze käme
// (§2) — und weil sie eine andere Frage beantwortet: dort steht, was man mit
// **einer** Partie tut, hier, **welche** es gibt.

import (
	"context"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// partienbefehl nennt die Partien dieses Profils.
func partienbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 0 {
		an.klage("partien nimmt keine Angaben entgegen — welche es gibt, "+
			"hängt am Profil.\n%s", hilfe)

		return CodeBedienung
	}

	return mitPartieflaeche(ctx, bestandsort, an,
		func(ctx context.Context, flaeche *app.Partieflaeche) int {
			return partienAusgeben(an, flaeche.Partien(ctx))
		})
}

// partienAusgeben schreibt die Liste — als Text oder als JSON.
//
// **Und die leere Liste sagt es ausdrücklich.** Eine Ausgabe ohne Zeile
// sähe aus wie ein Programm, das nichts getan hat (§6).
func partienAusgeben(an *schreiber, antwort app.PartienAntwort) int {
	if antwort.Fehler != nil {
		an.klageUeber(antwort.Fehler.AlsFehler())

		return CodeFachlich
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, antwort)
	}

	if len(antwort.Partien) == 0 {
		an.sag("Keine Partie. Fang eine mit »mse partie-anlegen« an.\n")

		return an.abschluss()
	}

	for _, zeile := range antwort.Partien {
		namen := make([]string, 0, len(zeile.Sitze))
		for _, sitz := range zeile.Sitze {
			namen = append(namen, sitz.Kennung+" ("+sitz.Name+")")
		}

		an.sag("%s — %s, %d Objekte, Sitze: %s%s\n",
			zeile.ID, zeile.Stand, zeile.Objekte, strings.Join(namen, ", "),
			eigenerSitzAngabe(zeile))
	}

	return an.abschluss()
}

// eigenerSitzAngabe hängt an die Zeile, welcher Sitz der eigene ist und ob die
// Partie auf ihn wartet — UXR-F16
// (#2031).
//
// **Nichts, wo das Profil nicht sitzt**: die Zeile bleibt dann, wie sie war,
// und wer sie ausschneidet, findet an der alten Stelle dasselbe (§16,
// erweiternd). **Dieselben Felder wie bei `--json`** — `eigenerSitz` und
// `dran` —, nur in Worten: die Auskunft ist an jeder Fläche eine, und gerechnet
// hat sie `service`.
func eigenerSitzAngabe(zeile app.PartiezeileDTO) string {
	if zeile.EigenerSitz == "" {
		return ""
	}

	angabe := ", dein Sitz: " + zeile.EigenerSitz

	if zeile.Dran {
		angabe += ", dran"
	}

	return angabe
}
