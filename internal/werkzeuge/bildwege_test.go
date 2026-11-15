// Der Wächter über die Bildvorsätze — Werkstück G2 von Phase 5.5
// (#1459).
//
// # Wofür es ihn gibt
//
// **Weil ein Bildpfad zwei Leser hat.** In Go sagt [vorschauweg.Pfad]
// beziehungsweise [vorschauweg.Spielpfad], unter welchem Vorsatz der Weg
// antwortet. Unter `wails dev` steht aber **Vite** davor, und es beantwortet
// jeden unbekannten Weg mit `index.html`; Wails greift auf seinen Handler nur
// zurück, wenn der Entwicklungsserver 404 oder 405 meldet.
//
// **Gemessen am 2026-09-15**: `/spielbild/` fehlte in `vite.config.ts`, `GET
// /spielbild/<hash>.png` gab `200 text/html`, und am Spieltisch hingen
// achtzehn Bilder mit `naturalWidth` null. Im **gebauten** Programm wäre es
// nie aufgefallen — dort gibt es kein Vite.
//
// # Warum er eine Ableitung prüft und keine Liste
//
// **Weil eine zweite Liste wieder vergessen wird.** Der erste Anlauf schrieb
// die Vorsätze in `vite.config.ts` noch einmal hin, und dieser Wächter zählte
// nach — das fängt den Fehler, aber erst in `make ci`. Seitdem **liest**
// `vite.config.ts` die Konstanten aus `internal/vorschauweg`, und geprüft
// wird, dass ihr Ausdruck jede davon erfasst.
//
// **Der Ausdruck wird aus `vite.config.ts` gelesen und nicht hier
// nachgebaut** — sonst prüfte dieser Lauf seine eigene Kopie.
//
// # Was er nicht prüft
//
// **Ob die Middleware das Richtige tut.** Er sieht, dass ein Vorsatz gefunden
// wird, nicht dass er zu einer 404 führt. Das steht eine Zeile darunter im
// Quelltext und wäre nur mit einem laufenden Vite zu messen — die
// Sichtprüfung hat es getan (§3).
package werkzeuge

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/vorschauweg"
)

// vorsatzausdruck holt den Ausdruck, mit dem `vite.config.ts` sucht.
var vorsatzausdruck = regexp.MustCompile(
	`const vorsatzregel = /([^/]+(?:\\/[^/]*)*)/gm`)

// **Der Ausdruck aus `vite.config.ts` findet jeden Bildvorsatz aus Go.**
func TestDerViteausdruckFindetJedenBildvorsatz(t *testing.T) {
	t.Parallel()

	gefunden := vorsaetzeWieViteSieFindet(t)

	for _, vorsatz := range []string{vorschauweg.Pfad, vorschauweg.Spielpfad} {
		if !gefunden[vorsatz] {
			t.Errorf("%q antwortet in Go, wird aber vom Ausdruck in "+
				"`frontend/vite.config.ts` nicht gefunden — unter `wails dev` "+
				"gibt Vite darauf `index.html`, und das Bild bleibt leer",
				vorsatz)
		}
	}
}

// vorsaetzeWieViteSieFindet wendet den Ausdruck aus `vite.config.ts` auf die
// Quellen von `internal/vorschauweg` an.
func vorsaetzeWieViteSieFindet(t *testing.T) map[string]bool {
	t.Helper()

	konfigurat, fehler := os.ReadFile(filepath.Join(
		"..", "..", "frontend", "vite.config.ts"))
	if fehler != nil {
		t.Fatalf("vite.config.ts nicht lesbar: %v", fehler)
	}

	treffer := vorsatzausdruck.FindStringSubmatch(string(konfigurat))
	if treffer == nil {
		t.Fatal("`const vorsatzregel = /…/gm` steht nicht in vite.config.ts " +
			"— entweder ist die Ableitung weg, oder sie heißt anders und " +
			"dieser Wächter prüft nichts mehr")
	}

	// **Von JavaScript nach Go**: der Ausdruck dort sucht in einer Datei, in
	// der jedes Anführungszeichen schon zu einem Apostroph geworden ist.
	regel, fehler := regexp.Compile(strings.ReplaceAll(
		"(?m)"+treffer[1], `\/`, "/"))
	if fehler != nil {
		t.Fatalf("der Ausdruck aus vite.config.ts ist keiner: %v", fehler)
	}

	gefunden := map[string]bool{}

	eintraege, fehler := os.ReadDir(filepath.Join("..", "vorschauweg"))
	if fehler != nil {
		t.Fatalf("internal/vorschauweg nicht lesbar: %v", fehler)
	}

	for _, eintrag := range eintraege {
		name := eintrag.Name()
		if !strings.HasSuffix(name, ".go") ||
			strings.HasSuffix(name, "_test.go") {
			continue
		}

		inhalt, fehler := os.ReadFile( //nolint:gosec // Quelltext im Repository
			filepath.Join("..", "vorschauweg", name))
		if fehler != nil {
			t.Fatalf("%s nicht lesbar: %v", name, fehler)
		}

		mitApostroph := strings.ReplaceAll(string(inhalt), `"`, "'")

		for _, gesehen := range regel.FindAllStringSubmatch(mitApostroph, -1) {
			gefunden[gesehen[1]] = true
		}
	}

	// **Die Gegenprobe gegen einen Ausdruck, der nichts findet** (§10): eine
	// leere Sammlung wäre grün und wertlos.
	if len(gefunden) == 0 {
		t.Fatal("der Ausdruck aus vite.config.ts findet in " +
			"internal/vorschauweg keinen einzigen Vorsatz")
	}

	return gefunden
}
