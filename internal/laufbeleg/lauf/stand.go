package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/laufbeleg"
)

// stand ist, was `git` über den Baum sagt.
type stand struct {
	commit      string
	aenderungen string
	sauber      bool
}

// standLesen fragt `git` — und gilt als unsauber, wenn es nicht antwortet.
//
// **Zweimal gelesen, vor und nach dem Lauf**: ein Baum, der sich dazwischen
// ändert, war nicht der, den der Commit nennt.
//
// **Der eigene Beleg zählt nicht mit** (`beleg`): ihn schreibt der vorige
// Lauf, und bis er eingecheckt ist, hieße sonst jeder weitere Beleg »anderer
// Stand« — gefunden am 2026-09-22, als ein blockierter Lauf seinen Beleg
// hinterließ.
func standLesen(ctx context.Context, kasten werkzeugkasten, beleg string) stand {
	commit, commitfehler := kasten.git(ctx, "rev-parse", "HEAD")
	aenderungen, statusfehler := kasten.git(ctx, "status", "--porcelain", "--", ".",
		":(exclude)"+beleg)

	if commitfehler != nil {
		commit = ""
	}

	return stand{
		commit:      commit,
		aenderungen: aenderungen,
		sauber:      commitfehler == nil && statusfehler == nil && aenderungen == "",
	}
}

// inhaltLesen gibt den Hash über das, was der Lauf misst — die Pfade des
// Plans und den Plan selbst, wie `git ls-files -s` sie mit ihren Blobs nennt.
//
// **Der Inhalt und nicht der Commit** sagt, ob ein Beleg noch passt: jeder PR
// wird per Squash gemergt, und danach liegt der gemessene Commit in keiner
// Geschichte mehr. Nach der Frist von `git gc` gäbe es ihn nicht einmal.
func inhaltLesen(ctx context.Context, kasten werkzeugkasten, misst []string, planpfad string) (string, error) {
	argumente := append(append([]string{"ls-files", "-s", "--"}, misst...), planpfad)

	liste, fehler := kasten.git(ctx, argumente...)
	if fehler != nil {
		return "", fehler
	}

	summe := sha256.Sum256([]byte(liste))

	return "sha256:" + hex.EncodeToString(summe[:]), nil
}

// benannt gibt den Stand für eine Zeile.
func (s stand) benannt() string {
	switch {
	case s.commit == "":
		return "ohne Commit"
	case s.sauber:
		return kurz(s.commit) + " (sauber)"
	default:
		return kurz(s.commit) + " (mit Änderungen)"
	}
}

// kurz gibt die ersten sieben Zeichen eines Commits.
func kurz(commit string) string {
	return commit[:min(7, len(commit))]
}

// zeigeEingabe zeigt eine Eingabe.
func zeigeEingabe(zu *anzeige, eingabe laufbeleg.Eingabe) {
	if eingabe.Fehlt != "" {
		zu.sag("  ✗ %s — %s\n", eingabe.Name, eingabe.Fehlt)

		return
	}

	zu.sag("  · %s — %d Dateien, %s\n", eingabe.Name, eingabe.Dateien,
		strings.TrimPrefix(eingabe.Hash, "sha256:")[:12])
}

// zusammenfassen zeigt, was der Beleg sagt — und an jedem Fall, der nicht
// bestanden hat, warum.
func zusammenfassen(zu *anzeige, pfad string, beleg laufbeleg.Beleg) {
	zu.sag("\nBeleg %s: %s — %d von %d Fällen gemessen\n",
		pfad, beleg.Status, beleg.Gemessen, beleg.Erwartet)

	if beleg.Grund != "" {
		zu.sag("  %s\n", beleg.Grund)
	}

	for _, ergebnis := range beleg.Faelle {
		if ergebnis.Ausfuehrung == laufbeleg.Bestanden && ergebnis.Ergebnis == laufbeleg.Bestanden {
			continue
		}

		zu.sag("  %s %s.%s — gemessen %s, Ergebnis %s: %s\n",
			zeichen(ergebnis), zu.kuerzer(ergebnis.Paket), ergebnis.Test,
			ergebnis.Ausfuehrung, ergebnis.Ergebnis, ergebnis.Grund)
	}
}

// zeichen gibt das Zeichen eines Falls.
func zeichen(ergebnis laufbeleg.Fallergebnis) string {
	if ergebnis.Art == laufbeleg.Trend && ergebnis.Ausfuehrung == laufbeleg.Bestanden {
		return "~"
	}

	return "✗"
}

// kuerzer nimmt das Modul vom Paket.
func (a *anzeige) kuerzer(paket string) string {
	return strings.TrimPrefix(strings.TrimPrefix(paket, a.modul), "/")
}
