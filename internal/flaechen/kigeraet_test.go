// **Apples Gerätemodell über jede Fläche ohne Fenster** — AI-05.AK-08
// (#2042).
//
// # Was hier gemessen wird
//
// **Der Sitzungsschalter an MCP und REST** — derselbe Stand an beiden — und
// **der Vorrang ohne bereiten Anbieter** an allen drei: ein Testlauf sieht
// kein Gerätemodell (der Riegel in `internal/applefm`), und Ollama ist
// ausgeschaltet. Dann sagt jede Fläche denselben fachlichen Satz, und **keine
// fragt eine Cloud** — das ist die Zusage, die hier zählt.
//
// **Was er nicht deckt**: das echte Modell. Das tut der Echtlauf hinter
// `fmlauf` (`internal/applefm/fmlauf_test.go`), und die Hüllenprobe am
// Fenster.
package flaechen_test

import (
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Abschalten für die Sitzung gibt an beiden Flächen denselben Stand.**
func TestDasGeraetemodellLaesstSichFuerDieSitzungAbschalten(t *testing.T) {
	umgebung := append(eigenerBestand(t), "HOME="+t.TempDir())

	an := starteDienst(t, umgebung)

	aus, istFehler := mcpRuf(t, umgebung, "schreiben.ki.geraetemodell.sitzung",
		map[string]any{"an": false})
	if istFehler || !strings.Contains(aus, `"an": false`) {
		t.Errorf("MCP: %s", aus)
	}

	status, rumpf := an.ruf(t, http.MethodPost, "/ki/geraetemodell/sitzung", `{"an":false}`)
	if status != http.StatusOK || !strings.Contains(rumpf, `"an":false`) {
		t.Errorf("REST: %d %s", status, rumpf)
	}
}

// **Der Vorrang ohne bereiten lokalen Anbieter scheitert überall gleich** —
// und an der Kommandozeile auch mit `--ohne-geraetemodell`, dem Schalter je
// Aufruf.
func TestDerVorrangOhneLokalenAnbieterUeberJedeFlaeche(t *testing.T) {
	umgebung := append(eigenerBestand(t), "HOME="+t.TempDir())

	// **Eine echte Karte**: ohne sie scheiterte der Ruf an MCP und REST schon
	// am Bezug, an der Kommandozeile erst am Anbieter — gemessen wäre dann
	// der Unterschied der Reihenfolge und nicht der Vorrang.
	_, karte := eineAbnahmekarte(t, umgebung)
	kennung := strconv.FormatInt(karte, 10)

	an := starteDienst(t, umgebung)

	abweisungGleich(t, "ein Vorschlag mit Vorrang ohne bereiten Anbieter", abweisung{
		klasse: klasseValidierung,
		kern:   "kein lokaler Anbieter ist bereit",
		fenster: func() *app.Fehler {
			return fensterAus(t, umgebung).KIVorschlagen(t.Context(), kennung, app.KIVorschlagsbezug{
				KarteID: karte, Anbieter: "vorrang", Auftrag: "Ein Name", Nur: []string{"name"},
			}).Fehler
		},
		cli: func() (string, int) {
			aus, fehler, code := kommandozeile(t, umgebung,
				"ki-vorschlag", kennung, "vorrang", "", "Ein Name", "--nur=name",
				"--ohne-geraetemodell")

			return aus + fehler, code
		},
		mcp: func() (string, bool) {
			return mcpRuf(t, umgebung, "schreiben.ki.vorschlag", map[string]any{
				"karteId": karte, "anbieter": "vorrang", "alias": "", "auftrag": "Ein Name",
				"nur": []string{"name"},
			})
		},
		rest: func() (int, string) {
			return an.ruf(t, http.MethodPost, "/karten/"+kennung+"/ki-vorschlag",
				`{"anbieter":"vorrang","alias":"","auftrag":"Ein Name","nur":["name"]}`)
		},
	})
}
