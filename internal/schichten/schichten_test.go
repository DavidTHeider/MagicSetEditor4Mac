// Die Anwendung der Schichtenregel auf den wirklichen Baum.
//
// Die Regel selbst und ihre Gegenproben stehen in `regeln_test.go`.
//
// # Was dieser Wächter nicht sieht
//
//   - **Keinen Umweg über eine Schnittstelle.** `service` definiert die
//     Repository-Schnittstellen, `storage` erfüllt sie ohne Import — das ist
//     Absicht (AGENTS.md §2) und für einen Importgraphen unsichtbar. Wer die
//     fachliche Zuordnung umdreht, ohne einen Import zu setzen, kommt hier durch.
//   - **Keine fachliche Entscheidung in `internal/app`.** Dass dort kein `if`
//     über einen Domänenzustand steht, ist eine Aussage über Inhalt, nicht über
//     Importe. Das bleibt Sache des Inline-Reviews.
//   - **Fachliche Entscheidungen.** Ob in `internal/app` ein `if` über einen
//     Domänenzustand steht, ist eine Aussage über Inhalt und in keinem
//     Importgraphen zu sehen.
//
// `main.go` wird **mit**geprüft, unter der Kennung `main`: es darf nur
// `internal/app` kennen. Bis zum Zweitblick am Ende von Phase 1 begann die
// Erfassung bei `internal/`, und ein direkter Griff von `main.go` nach
// `storage` oder `service` wäre unbemerkt geblieben — obwohl AGENTS.md §2 den
// Composition Root ausdrücklich `internal/app` zuordnet.
//
// Gelesen werden **alle** `.go`-Dateien, auch die hinter Build-Tags und die
// Tests. Ein `//go:build windows` hebt die Schichtenregel nicht auf, und ein
// Test, der von `domain` nach `service` greift, ist derselbe Verstoß — er
// verrät nur zusätzlich, dass die Domäne mehr weiß, als sie sollte.
package schichten

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// internalVerzeichnis ist der Ort der Schichten, von diesem Paket aus gesehen.
const internalVerzeichnis = ".."

// wurzelpaket ist die Kennung, unter der main.go geführt wird.
const wurzelpaket = "main"

func TestKeinPaketVerletztDieSchichtenregel(t *testing.T) {
	importe := importeLesen(t)

	for _, befund := range pruefe(importe) {
		t.Errorf("%s", befund)
	}
}

// TestJedesPaketIstGefuehrt weist ein Paket ab, das weder eingeordnet noch
// ausgenommen ist.
//
// Ohne ihn entstünde eine Schicht, die niemand eingeordnet hat — und die Regel
// wäre lautlos umgangen, weil `pruefe` über sie nichts zu sagen hätte.
func TestJedesPaketIstGefuehrt(t *testing.T) {
	for paket := range importeLesen(t) {
		t.Run(paket, func(t *testing.T) {
			_, eingeordnet := erlaubt[paket]
			_, ausgenommen := keineSchicht[paket]

			if !eingeordnet && !ausgenommen {
				t.Errorf("das Paket internal/%s steht weder in erlaubt noch in keineSchicht "+
					"— wohin gehört es? (AGENTS.md §2)", paket)
			}
		})
	}
}

// TestJedeGefuehrteSchichtExistiertOderIstAngekuendigt sagt, welche Schichten
// es noch nicht gibt.
//
// Er schlägt nicht fehl: `erlaubt` beschreibt den Zielzustand aus AGENTS.md §2,
// und in Phase 1 fehlt das meiste davon planmäßig. Er meldet den Stand, damit
// eine Zeile, die zu einer nie gebauten Schicht gehört, nicht unbemerkt für
// immer stehen bleibt.
func TestJedeGefuehrteSchichtExistiertOderIstAngekuendigt(t *testing.T) {
	vorhanden := importeLesen(t)

	for paket := range erlaubt {
		if _, gibtEs := vorhanden[paket]; !gibtEs {
			t.Logf("internal/%s ist in der Regel geführt und existiert noch nicht", paket)
		}
	}
}

// importeLesen sammelt je Paket unter internal/ die Importpfade aller
// .go-Dateien.
func importeLesen(t *testing.T) map[string][]Einfuhr {
	t.Helper()

	eintraege, err := os.ReadDir(internalVerzeichnis)
	if err != nil {
		t.Fatalf("internal/ nicht lesbar: %v", err)
	}

	importe := make(map[string][]Einfuhr, len(eintraege)+1)

	for _, eintrag := range eintraege {
		if !eintrag.IsDir() {
			continue
		}

		importe[eintrag.Name()] = paketImporte(t, filepath.Join(internalVerzeichnis, eintrag.Name()))
	}

	// Das Wurzelpaket gehört dazu.
	importe[wurzelpaket] = paketImporte(t, filepath.Join(internalVerzeichnis, ".."))

	// **Und jedes weitere Programm unter `cmd/`.**
	//
	// Seit #143 gibt es ein zweites: der MCP-Server ist ein eigener Prozess
	// ([0010](../../docs/entscheidungen/0010-mcp-als-eigener-prozess.md)). Ohne
	// diese Zeilen stünde er außerhalb jeder Regel — und ein zweiter
	// Composition Root, den niemand prüft, ist genau die Stelle, an der eine
	// Schicht still umgangen wird.
	cmdVerzeichnis := filepath.Join(internalVerzeichnis, "..", "cmd")

	programme, err := os.ReadDir(cmdVerzeichnis)
	if err != nil && !os.IsNotExist(err) {
		t.Fatalf("cmd/ nicht lesbar: %v", err)
	}

	for _, programm := range programme {
		if !programm.IsDir() {
			continue
		}

		importe["cmd/"+programm.Name()] = paketImporte(t, filepath.Join(cmdVerzeichnis, programm.Name()))
	}

	return importe
}

// paketImporte liest die Importe eines Verzeichnisses samt Unterverzeichnissen.
func paketImporte(t *testing.T, verzeichnis string) []Einfuhr {
	t.Helper()

	var gesammelt []Einfuhr

	// Nur die oberste Ebene beim Wurzelpaket: darunter liegen internal/,
	// tools/ und frontend/, und die haben ihre eigenen Regeln.
	nurObersteEbene := filepath.Clean(verzeichnis) == filepath.Clean(filepath.Join(internalVerzeichnis, ".."))

	fehler := filepath.WalkDir(verzeichnis, func(pfad string, eintrag os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// Ein Symlink ist kein Paket: in einem Worktree zeigt `.go` auf den
		// Cache des Hauptcheckouts, und wer ihm folgt, liest ein Verzeichnis
		// als Datei — der vierte Lauf von A1 (#1440) endete so.
		if eintrag.Type()&os.ModeSymlink != 0 {
			return nil
		}
		if eintrag.IsDir() {
			if nurObersteEbene && filepath.Clean(pfad) != filepath.Clean(verzeichnis) {
				return filepath.SkipDir
			}

			return nil
		}
		if !strings.HasSuffix(pfad, ".go") {
			return nil
		}

		// Der Pfad **innerhalb** des Pakets, nicht der bloße Dateiname.
		//
		// Mit `filepath.Base` hätte ein späteres `internal/app/unterpaket/app.go`
		// als das erlaubte `app.go` gegolten — der Nachlauf des Zweitblicks hat
		// darauf hingewiesen. Eine Ausnahme, die auf einen Namen hört statt auf
		// eine Stelle, wandert mit dem Namen.
		innen, fehlerImPfad := filepath.Rel(verzeichnis, pfad)
		if fehlerImPfad != nil {
			return fehlerImPfad
		}

		for _, ziel := range dateiImporte(t, pfad) {
			gesammelt = append(gesammelt, Einfuhr{Datei: filepath.ToSlash(innen), Pfad: ziel})
		}

		return nil
	})
	if fehler != nil {
		t.Fatalf("%s nicht durchlaufbar: %v", verzeichnis, fehler)
	}

	return gesammelt
}

// dateiImporte zerteilt eine Datei bis zum Ende ihres Importblocks.
//
// `parser.ImportsOnly` und nicht der ganze Baum: gebraucht werden nur die
// Importe, und eine Datei mit einem Syntaxfehler im Rumpf soll diesen Wächter
// nicht zum Schweigen bringen — dafür ist `go build` zuständig.
func dateiImporte(t *testing.T, pfad string) []string {
	t.Helper()

	baum, err := parser.ParseFile(token.NewFileSet(), pfad, nil, parser.ImportsOnly)
	if err != nil {
		t.Fatalf("%s nicht zerteilbar: %v", pfad, err)
	}

	pfade := make([]string, 0, len(baum.Imports))
	for _, eintrag := range baum.Imports {
		wert, fehler := strconv.Unquote(eintrag.Path.Value)
		if fehler != nil {
			t.Fatalf("%s: Importpfad %s nicht lesbar: %v", pfad, eintrag.Path.Value, fehler)
		}

		pfade = append(pfade, wert)
	}

	return pfade
}
