//go:build fremdpakete

// Jede Ordnungszeichenkette aus **bezogenen** Paketen, einmal durch.
//
// # Warum das hinter einem Build-Tag steht
//
// `testdaten/` enthält keine fremden Vorlagen (§11,
// [0005](../../docs/entscheidungen/0005-keine-fremden-vorlagen.md)). Dieser Lauf
// schließt die Lücke, ohne sie ins Repository zu holen. Er läuft nicht in der
// Pipeline, zählt nicht zur Abdeckung und wird von Hand angestoßen:
//
//	MSE_FREMDPAKETE="$HOME/Library/Application Support/MagicSetEditor4Mac/pakete/M15-Magic-Pack/data" \
//	  go test -tags fremdpakete ./internal/msescript/ -run Fremdpakete -v
//
// # Was er prüft, und was nicht
//
// Er sammelt jede Zeichenkette, die in einem `sort_text@(order: …)` steht, und
// hält sie gegen den Zerleger. Erwartet wird **nicht**, dass alle durchgehen:
// die runde Klammer ist nicht erschlossen (#98) und wird gemeldet. Erwartet
// wird, dass jede Ordnung entweder gelesen oder **benannt abgelehnt** wird —
// eine dritte Antwort gibt es nicht, und ein Absturz oder ein stilles Ergebnis
// wäre genau die, vor der das Issue warnt.
//
// **Und er zählt.** Wer die Zahlen liest, sieht, wie viel von der Sprache
// heute trägt — und wie viel eine Freigabe verschweigen würde.
package msescript_test

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// ordnungsmuster findet `sort_text(order: "…")` samt der Zeichenkette.
//
// **Über einen Ausdruck und nicht über den Zerteiler**: gesucht wird in
// Skriptdateien, die zu einem Bruchteil aus Skript bestehen — und ein
// Zerteilerfehler an einer anderen Stelle machte diesen Lauf blind für das, was
// er prüfen soll.
var ordnungsmuster = regexp.MustCompile(`sort_text@?\(\s*order:\s*"((?:[^"\\]|\\.)*)"`)

func TestOrdnungenAusFremdpaketen(t *testing.T) {
	wurzel := os.Getenv("MSE_FREMDPAKETE")
	if wurzel == "" {
		t.Skip("MSE_FREMDPAKETE ist nicht gesetzt — kein Verzeichnis mit bezogenen Paketen")
	}

	gefunden := map[string]int{}

	//nolint:gosec // ein Lauf über MSE_FREMDPAKETE; die Bäume dort sind bezogen und liegen still (§3)
	fehler := filepath.WalkDir(wurzel, func(pfad string, eintrag os.DirEntry, err error) error {
		if err != nil || eintrag.IsDir() {
			return nil //nolint:nilerr // ein unlesbarer Zweig ist kein Grund abzubrechen
		}

		inhalt, lesefehler := os.ReadFile(pfad) //nolint:gosec // derselbe Lauf, derselbe Baum
		if lesefehler != nil {
			return nil //nolint:nilerr // eine unlesbare Datei zählt nicht mit, hält aber den Lauf nicht an
		}

		for _, treffer := range ordnungsmuster.FindAllStringSubmatch(string(inhalt), -1) {
			gefunden[entkleide(treffer[1])]++
		}

		return nil
	})
	if fehler != nil {
		t.Fatalf("%s nicht durchlaufbar: %v", wurzel, fehler)
	}

	if len(gefunden) == 0 {
		t.Fatalf("keine einzige Ordnung in %s — der Ausdruck greift nicht mehr", wurzel)
	}

	ordnungen := make([]string, 0, len(gefunden))
	for ordnung := range gefunden {
		ordnungen = append(ordnungen, ordnung)
	}

	sort.Strings(ordnungen)

	gelesen, gemeldet := 0, 0

	for _, ordnung := range ordnungen {
		_, err := msescript.AuswertenQuelle(t.Context(),
			`sort_text(order: `+alsQuelltext(ordnung)+`, "2WWU/R10T")`, msescript.Bibliothek())

		switch {
		case err == nil:
			gelesen++

		case strings.Contains(err.Error(), "nicht erschlossen"):
			gemeldet++

			t.Logf("gemeldet: %s", ordnung)

		default:
			t.Errorf("%q scheitert mit einer anderen Meldung als der benannten: %v",
				ordnung, err)
		}
	}

	t.Logf("%d verschiedene Ordnungen: %d gelesen, %d benannt abgelehnt",
		len(ordnungen), gelesen, gemeldet)

	if gelesen == 0 {
		t.Error("keine einzige Ordnung war lesbar")
	}
}

// entkleide macht aus dem Quelltext einer MSE-Script-Zeichenkette ihren Wert.
//
// Nur die Fälle, die vorkommen: `\\` wird zu `\`, `\"` zu `"`. Mehr zu deuten
// hieße eine Schreibweise zu behaupten, die kein Paket führt.
func entkleide(roh string) string {
	return strings.NewReplacer(`\\`, `\`, `\"`, `"`).Replace(roh)
}

// alsQuelltext schreibt eine Ordnung wieder als MSE-Script-Zeichenkette.
func alsQuelltext(wert string) string {
	return `"` + strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(wert) + `"`
}
