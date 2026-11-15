package domain_test

// **Eine Grenze ist Klartext** — gefunden am 2026-09-24 in WebKit, auf der
// Seite *Eine Partie anfangen*: „was auf einer Karte \*\*ohne\*\* Spielskript
// steht" stand dort mit den Sternen.
//
// # Warum ein Lauf und nicht nur die Korrektur
//
// **Weil die Grenzen an vier Flächen erscheinen, und an keiner wird
// Auszeichnung gelesen**: das Fenster setzt sie als Text, die
// Kommandozeile druckt sie, MCP und der Dienst reichen sie als JSON weiter.
// Die Sterne stammten aus der Gewohnheit dieses Projekts, im **Kommentar**
// zu betonen — in einer Zeichenkette sind sie ein Fehler, den erst ein
// Mensch am Bildschirm sieht (§3).
//
// **Was der Lauf nicht hält**: andere Texte des Kerns — Meldungen, Befunde.
// Er deckt die Matrix beider Profile ab, wo die drei Funde lagen.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEineGrenzeTraegtKeineAuszeichnung(t *testing.T) {
	t.Parallel()

	for _, profil := range []domain.Regelprofil{
		domain.Grundprofil("magic"),
		domain.Commanderprofil("magic"),
	} {
		for _, eintrag := range profil.Matrix.Eintraege {
			for _, grenze := range eintrag.Grenzen {
				if strings.Contains(grenze, "**") || strings.Contains(grenze, "`") {
					t.Errorf("%s/%s: Auszeichnung in der Grenze %q",
						profil.Kennung, eintrag.Bereich, grenze)
				}
			}
		}
	}
}
