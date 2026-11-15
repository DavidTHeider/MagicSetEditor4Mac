package schichten

// **Kein Unterprozess nimmt ein Argument aus der Ferne** — ausgelagert aus
// `netz_test.go`, weil die Datei mit dem Eintrag für `internal/fernleitung`
// über die Zeilengrenze aus §2 ging.
//
// Der Schnitt folgt der Frage: dort, **wer ins Netz darf** — hier, **was
// dieses Programm startet**. Zwei Wächter, zwei Fragen.

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// **Kein Unterprozess nimmt sein Programm aus der Ferne** —
// #1170,
// Befund 7 des ersten Zweitblicks über die Freigabe.
//
// # Der Befund
//
// [TestNurBezugKenntDasNetz] begründete seine Unterprozesslücke mit „`os/exec`
// kommt im Projekt nicht vor". Seit dem 2026-08-25 kommt es vor — zweimal, in
// den ausgelieferten Dateien `protokoll_darwin.go` und `startmeldung_darwin.go`.
//
// **Die Begründung war falsch, die Lücke geblieben.** Eine Ausrede prüft
// niemand nach, weil sie wie eine Begründung aussieht (§10).
//
// # Was dieser Wächter hält
//
// Jeder `exec.CommandContext` nennt sein Programm als **Zeichenkettenliteral**.
// Ein Aufruf, dessen erstes Argument aus einer Veränderlichen kommt, macht die
// Pipeline rot — denn dann entscheidet der Zustand des Programms, was startet.
//
// # Ein zweiter Fall seit N6
//
// **Ein Programm neben diesem** — Werkstück N6 von Phase 5.4
// (#1572).
// Das Fenster startet `mse-rest` und `mse-mcp` aus **seinem eigenen
// Verzeichnis**, und das ist zur Bauzeit keine Zeichenkette: der Ort kommt aus
// `os.Executable()`.
//
// **Die Zusage bleibt dieselbe, der Ausdruck wird genauer**: erlaubt ist
// zusätzlich `filepath.Join(…, "<literal>")` — der **Name** steht im
// Quelltext, das Verzeichnis kommt vom Betriebssystem. Wer ihn setzen könnte,
// müsste dieses Programm verschieben; wer das kann, hat es ohnehin ersetzt.
//
// **Was die Lockerung nicht aufgibt**: ein Name aus einer Veränderlichen
// bleibt rot. `filepath.Join(ort, name)` fällt weiterhin durch, und genau das
// hat den ersten Anlauf von N6 angehalten.
//
// # Was er **nicht** hält
//
// **Ein fest verdrahtetes `curl`.** `exec.CommandContext(ctx, "/usr/bin/curl", …)`
// käme durch. Das ist kein Versehen: eine Liste erlaubter Programme wäre eine
// zweite Aufzählung neben der Netzgrenze, und die Frage gehört strukturell
// gestellt (#227). Was hier ausgeschlossen ist, ist der **unbeaufsichtigte**
// Fall — ein Programmname, den jemand von außen setzen kann.
//
// **Und nicht, dass das Verzeichnis stimmt.** Es kommt aus `os.Executable()`;
// dass dort das richtige Programm liegt, ist eine Eigenschaft des Bündels und
// keine dieses Wächters.
func TestKeinUnterprozessNimmtEinArgumentAusDerFerne(t *testing.T) {
	gefunden := 0

	fehler := filepath.WalkDir(internalVerzeichnis,
		func(pfad string, eintrag os.DirEntry, err error) error {
			if err != nil || eintrag.IsDir() || !strings.HasSuffix(pfad, ".go") ||
				strings.HasSuffix(pfad, "_test.go") {
				return err
			}

			roh, lesefehler := os.ReadFile(pfad) //nolint:gosec // Quelldatei im Repositorium
			if lesefehler != nil {
				return lesefehler
			}

			for _, treffer := range unterprozess.FindAllStringSubmatch(string(roh), -1) {
				gefunden++

				if !programmStehtImQuelltext(treffer[1]) {
					t.Errorf("%s startet ein Programm, dessen Name nicht im Quelltext "+
						"steht (%s) — wer ihn setzen kann, entscheidet, was läuft",
						pfad, strings.TrimSpace(treffer[1]))
				}
			}

			return nil
		})
	if fehler != nil {
		t.Fatalf("%s durchgehen: %v", internalVerzeichnis, fehler)
	}

	if gefunden == 0 {
		t.Fatal("kein exec.CommandContext gefunden — der Leser greift daneben, " +
			"oder die Begründung in TestNurBezugKenntDasNetz gehört umgeschrieben")
	}
}

// unterprozess trifft das erste Argument nach dem Kontext.
//
// **Bis zum Klammerschluss und nicht bis zum ersten Komma** — seit N6 steht
// dort ein `filepath.Join(ort, "mse-rest")`, und ein Ausdruck, der beim
// Komma darin abbräche, läse `filepath.Join(ort` und hielte es für eine
// Veränderliche.
var unterprozess = regexp.MustCompile(
	`(?s)exec\.CommandContext\(\s*\w+,\s*((?:[^,()]|\([^()]*\))+)`)

// programmStehtImQuelltext sagt, ob der Programmname ein Literal ist.
//
// **Zwei Gestalten und keine dritte**: eine Zeichenkette, oder ein
// `filepath.Join`, dessen **letztes** Stück eine ist. Der Grund steht im Kopf.
func programmStehtImQuelltext(ausdruck string) bool {
	ausdruck = strings.TrimSpace(ausdruck)

	if strings.HasPrefix(ausdruck, `"`) {
		return true
	}

	innen, istJoin := strings.CutPrefix(ausdruck, "filepath.Join(")
	if !istJoin {
		return false
	}

	innen = strings.TrimSuffix(strings.TrimSpace(innen), ")")

	stuecke := strings.Split(innen, ",")
	letztes := strings.TrimSpace(stuecke[len(stuecke)-1])

	return strings.HasPrefix(letztes, `"`)
}
