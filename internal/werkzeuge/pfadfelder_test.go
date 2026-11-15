package werkzeuge

// **Jeder Pfad eines offenen Servers reist unter einem bekannten Namen** —
// OPS-02.AK-04 (#1719).
//
// # Was er hält
//
// **Beide Richtungen der Zusage aus `internal/rest/pfadvertrag.go`**:
//
//  1. Jedes Feld in `internal/rest` und `internal/mcp`, dessen Name nach
//     einem Pfad klingt (`pfad`, `ziel`, `datei`, `verzeichnis`, `ordner` —
//     auch als Teil), heißt so, wie `domain.IstPfadfeld` es kennt. Ein
//     `zielordner` fiele sonst am Vertrag vorbei.
//  2. Wer einen offenen Server baut — `cmd/mse-rest` und der Werkzeugserver
//     über HTTP —, setzt den Vertrag.
//
// # Was er nicht hält
//
// **Einen Pfad unter einem Namen, der nicht nach einem klingt** — ein
// `quelle` mit einem Verzeichnis sähe er nicht. Das hält seit #2073
// `pfadstrukturen_test.go`: es zählt jedes `quelle`, `ort` und `bild` und
// verlangt eine Entscheidung.

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// pfadklang findet Namen, die nach einem Pfad klingen — als Marke oder als
// Eigenschaft eines Schemas.
var pfadklang = regexp.MustCompile(`(?:json:"|^\s*")([a-z]*(?:pfad|ziel|datei|verzeichnis|ordner)[a-z]*)"`)

// keinPfadfeld sind Namen, die nach einem Pfad klingen und keiner sind —
// mit Grund.
var keinPfadfeld = map[string]string{
	"unterverzeichnis": "ein Verzeichnis **im Archiv** eines Pakets, nicht auf dieser Platte",
	"ziele":            "die Karten, an denen ein KI-Auftrag arbeitet",
	"dateien":          "eine Zahl von Dateien",
	"dateityp":         "die Art einer Ausgabe, etwa »png«",
	"dateiname":        "der Name der Hauptdatei ohne Verzeichnis, etwa »ausfuhr.xml« (Probelauf, #2015)",
	"zielrevision":     "eine Revision des Suchindex",
	"offiziell":        "ein Wort, das »ziel« nur buchstabiert",
}

func TestJedesPfadfeldDesDienstesIstBenannt(t *testing.T) {
	t.Parallel()

	for _, paket := range []string{"internal/rest", "internal/mcp"} {
		dateien, fehler := filepath.Glob(filepath.Join(wurzel, paket, "*.go"))
		if fehler != nil {
			t.Fatal(fehler)
		}

		for _, datei := range dateien {
			if strings.HasSuffix(datei, "_test.go") {
				continue
			}

			roh, lesefehler := os.ReadFile(datei) //nolint:gosec // Quelltext dieses Repositoriums
			if lesefehler != nil {
				t.Fatal(lesefehler)
			}

			for zeile := range strings.SplitSeq(string(roh), "\n") {
				for _, treffer := range pfadklang.FindAllStringSubmatch(zeile, -1) {
					name := treffer[1]
					if _, frei := keinPfadfeld[name]; frei || domain.IstPfadfeld(name) {
						continue
					}

					t.Errorf("%s: »%s« klingt nach einem Pfad, und der Pfadvertrag kennt den Namen nicht",
						strings.TrimPrefix(datei, wurzel+"/"), name)
				}
			}
		}
	}
}

func TestJederOffeneServerSetztDenPfadvertrag(t *testing.T) {
	t.Parallel()

	for _, datei := range []string{"cmd/mse-rest/dienst.go", "cmd/mse-mcp/ueberhttp.go"} {
		if !strings.Contains(lies(t, datei), "MitPfadvertrag(app.Dienstpfadvertrag(") {
			t.Errorf("%s baut einen offenen Server ohne Pfadvertrag", datei)
		}
	}
}
