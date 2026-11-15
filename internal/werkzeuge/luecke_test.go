// **Eine Lücke ohne Grund macht die Pipeline rot**
// (#1107).
//
// # Was dieser Wächter hält
//
// §16 sagt: *„Eine neue Fähigkeit landet nicht an einer Fläche. Was das
// Fenster kann, können Kommandozeile und MCP — **oder es steht begründet,
// warum nicht.**"* Der zweite Halbsatz war bis zum 2026-08-25 unbewacht: der
// Katalog hatte für den Grund eine eigene Spalte, und **22 von 78** Zeilen
// ließen sie leer.
//
// **Eine leere Zelle sieht aus wie ein Grund und ist keiner.** Sie sagt nicht
// „hier gehört nichts hin", sondern „hier hat niemand nachgedacht" — dieselbe
// Klasse wie §10, *„Auch eine Ablehnung braucht Deckung"*.
//
// # Was er nicht hält
//
// **Ob der Grund stimmt.** Er sieht, dass in der Spalte etwas steht, nicht was
// es taugt. „Weil nicht" käme durch. Das liest ein Mensch — der Wächter
// erzwingt nur, dass es überhaupt jemand aufgeschrieben hat.
//
// **Und nicht, ob die Fähigkeit an der genannten Fläche dasselbe tut.** Dafür
// gibt es je Fähigkeit einen Lauf über alle Flächen (§16).
package werkzeuge

import (
	"strings"
	"testing"
)

// TestJedeLueckeNenntIhrenGrund hält die Begründungspflicht.
func TestJedeLueckeNenntIhrenGrund(t *testing.T) {
	t.Parallel()

	ohneGrund := 0

	for _, zeile := range faehigkeitszeilen(t) {
		spalten := strings.Split(zeile, "|")

		luecken := fehlendeFlaechen(spalten)
		if len(luecken) == 0 {
			continue
		}

		if strings.TrimSpace(spalten[6]) != "" {
			continue
		}

		ohneGrund++

		t.Errorf("%q fehlt an %s und nennt keinen Grund (%s)",
			strings.TrimSpace(spalten[1]), strings.Join(luecken, " und "), katalogpfad)
	}

	if ohneGrund > 0 {
		t.Logf("%d Zeile(n) ohne Grund — bauen oder begründen, beides ist eine Antwort",
			ohneGrund)
	}
}

// fehlendeFlaechen nennt die Flächen, an denen eine Zeile nichts stehen hat.
//
// **Das Fenster steht nicht darunter.** Der Katalog ist aus den Bindungen
// gewachsen; eine Zeile ohne Fensterspalte gibt es nicht, und wenn eine
// entstünde, wäre das ein anderer Befund als eine fehlende Fläche daneben.
func fehlendeFlaechen(spalten []string) []string {
	var luecken []string

	if !besetzt(spalten[3]) {
		luecken = append(luecken, "MCP")
	}

	if !besetzt(spalten[4]) {
		luecken = append(luecken, "CLI")
	}

	if !besetzt(spalten[5]) {
		luecken = append(luecken, "REST")
	}

	return luecken
}
