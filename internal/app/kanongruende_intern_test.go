//go:build fremdpakete

// Ein Befund nennt eine Karte, an der er auftritt
// (#497).
//
// # Warum das mehr ist als Bequemlichkeit
//
// Der Kanonlauf zählte Gründe und sonst nichts. Aus
//
//	2 × das Bild ist in keinem Paket zu finden
//
// war nicht zu erfahren, **welche** zwei Karten das sind — bei 25.148 Karten
// in zehn Sets. Jede Nachfrage kostete einen eigens geschriebenen Abgriff, und
// genau deshalb stand #497 seit dem 2026-08-18 offen, ohne dass jemand die
// Frage beantworten konnte.
//
// # Ein Beispiel, nicht alle
//
// Eine Klasse mit 159 Vorkommen führt **einen** Namen. Alle zu führen machte
// aus der Ausgabe eine Liste, in der die Klassen untergehen — und die Frage
// lautet „woran sehe ich es", nicht „wo überall".
package app

import (
	"sort"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// kanongrund ist eine Befundklasse samt einer Karte, an der sie auftritt.
type kanongrund struct {
	Anzahl int

	// Satz und Karte sind das **erste** Vorkommen. Weitere überschreiben es
	// nicht: das erste ist wiederholbar, ein späteres hinge an der
	// Reihenfolge der Sets.
	Satz, Karte string
}

// kanongruende zählt Befundklassen und merkt sich je eine Karte.
type kanongruende map[string]*kanongrund

// zaehle nimmt einen Befund auf.
func (g kanongruende) zaehle(klasse, satz, karte string) {
	vorhanden, da := g[klasse]
	if !da {
		g[klasse] = &kanongrund{Anzahl: 1, Satz: satz, Karte: karte}

		return
	}

	vorhanden.Anzahl++
}

// nenneGruende schreibt die Klassen in den Testlauf, häufigste zuerst.
func nenneGruende(t *testing.T, gruende kanongruende) {
	t.Helper()

	namen := make([]string, 0, len(gruende))
	for name := range gruende {
		namen = append(namen, name)
	}

	sort.Slice(namen, func(a, b int) bool {
		return gruende[namen[a]].Anzahl > gruende[namen[b]].Anzahl
	})

	for _, name := range namen {
		grund := gruende[name]
		if grund.Karte == "" {
			t.Logf("  %5d × %s", grund.Anzahl, name)

			continue
		}

		t.Logf("  %5d × %s   (etwa: %s / %s)",
			grund.Anzahl, name, grund.Satz, grund.Karte)
	}
}

// zaehleNachArt summiert die Klassen nach [render.Befundart].
//
// **Die Art kommt aus der Meldung, nicht aus dem Zähler.** `render.ArtAus`
// entscheidet sie an ihrem Wortlaut; hier wird nur summiert.
func zaehleNachArt(gruende kanongruende) map[render.Befundart]int {
	nachArt := map[render.Befundart]int{}

	for grund, eintrag := range gruende {
		nachArt[render.ArtAus(grund)] += eintrag.Anzahl
	}

	return nachArt
}

// satzname streift die Endung `.mse-set` ab.
//
// **Damit die Ausgabe lesbar bleibt**: `White / Nimbus Maze` sagt dasselbe wie
// `White.mse-set / Nimbus Maze` und passt in eine Zeile.
func satzname(datei string) string {
	return strings.TrimSuffix(datei, ".mse-set")
}
