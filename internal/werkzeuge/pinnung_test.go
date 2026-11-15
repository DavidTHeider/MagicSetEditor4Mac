// Wächter über die Werkzeugpinnung aus AGENTS.md §1.
//
// Die Versionstabelle in AGENTS.md ist die Fassung, die Menschen lesen; die
// go.mod-Dateien sind die, nach der gebaut wird. Laufen beide auseinander, ist
// die Tabelle eine Behauptung — und zwar eine, die niemand bemerkt, weil ein
// falscher Tabelleneintrag nichts kaputtmacht. Also zählt hier jemand nach.
//
// Geprüft wird in beide Richtungen: jede prüfbare Tabellenzeile braucht eine
// Pinnung mit derselben Version am genannten Ort, und jede tatsächliche Pinnung
// braucht eine Tabellenzeile. Eine Richtung allein ließe die andere Hälfte der
// Zusage unsichtbar (AGENTS.md §10).
//
// # Was dieser Wächter nicht prüft
//
// Diese Datei prüft nur, was über eine `tool`-Direktive oder die `go`-Direktive
// gepinnt ist. Die Frontend-Zeilen — Node, npm, TypeScript, ESLint — stehen in
// `frontendpinnung_test.go`; sie liegen an vier anderen Orten und werden anders
// gelesen.
//
// Übrig bleibt eine einzige Auslassung: die Zeile „GitHub Actions" nennt keine
// Version, sondern einen Commit-Hash je Action. Der Grund steht in `ungeprueft`;
// `TestJedeTabellenzeileIstGeprueftOderBegruendetAusgelassen` hält fest, dass
// keine Zeile stillschweigend durchrutscht — eine neue Zeile in AGENTS.md macht
// diesen Wächter rot, bis jemand entscheidet, wohin sie gehört.
package werkzeuge

import (
	"strings"
	"testing"
)

// pinnungsort verbindet eine Tabellenzeile mit der tatsächlichen Pinnung.
type pinnungsort struct {
	// modul ist der Paketpfad, wie er in der tool-Direktive steht.
	modul string
	// datei ist die go.mod-Datei, die die Tabelle als Ort der Pinnung nennt.
	datei string
}

// gepruefteWerkzeuge ordnet jeder prüfbaren Tabellenzeile den Modulpfad zu,
// unter dem das Werkzeug gepinnt ist. Ohne diese Zuordnung ließe sich die Zeile
// „gofumpt" nicht mit der Direktive „mvdan.cc/gofumpt" zusammenbringen.
var gepruefteWerkzeuge = map[string]pinnungsort{
	// **Die Hülle** — seit ENG-02 eine Abhängigkeit und kein Werkzeug mehr
	// ([0132](../../docs/entscheidungen/0132-wails-3-als-huelle.md)): gebaut
	// wird über `scripts/bauen.sh`, und ein Kommandozeilenwerkzeug gibt es
	// nicht. Dass `@wailsio/runtime` dieselbe Zahl trägt, hält
	// `frontendpinnung_test.go`.
	"Wails":           {modul: "github.com/wailsapp/wails/v3", datei: "go.mod"},
	"golangci-lint":   {modul: "github.com/golangci/golangci-lint/v2/cmd/golangci-lint", datei: "tools/go.mod"},
	"gosec":           {modul: "github.com/securego/gosec/v2/cmd/gosec", datei: "tools/go.mod"},
	"govulncheck":     {modul: "golang.org/x/vuln/cmd/govulncheck", datei: "tools/go.mod"},
	"gofumpt":         {modul: "mvdan.cc/gofumpt", datei: "tools/go.mod"},
	"cyclonedx-gomod": {modul: "github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod", datei: "tools/go.mod"},

	// **Keine Werkzeugpinnung, sondern eine Abhängigkeit.** `x/image` steht im
	// `require`-Block des Wurzelmoduls, nicht hinter einer `tool`-Direktive —
	// geprüft wird trotzdem dasselbe: dass die Zahl in der Tabelle die Zahl in
	// `go.mod` ist. Eine Auslassung wäre hier die bequemere und die falsche
	// Antwort: eine Abhängigkeit, deren Fassung niemand nachhält, verschiebt
	// Goldbilder ([0008](../../docs/entscheidungen/0008-schriftrasterung.md)).
	"golang.org/x/image": {modul: "golang.org/x/image", datei: "go.mod"},

	// **Und aus demselben Grund `x/text`** (Werkstück
	// #1500):
	// es ordnet die Sammlung nach Sprache. Wer die Fassung wechselt, wechselt
	// womöglich die Reihenfolge, in der „Ätherfluss" und „Zauber" stehen — und
	// niemand sähe es, wenn die Zahl in der Tabelle nicht nachgehalten würde.
	"golang.org/x/text": {modul: "golang.org/x/text", datei: "go.mod"},

	// **Der Proxy und das Protokoll** — Werkstück A18 von Phase 5.3
	// (#1533).
	"golang.org/x/net": {modul: "golang.org/x/net", datei: "go.mod"},

	// **Die Entdeckung im selben Netz** — Werkstück N2 von Phase 5.4
	// (#1455).
	"hashicorp/mdns": {modul: "github.com/hashicorp/mdns", datei: "go.mod"},
	"tailcat":        {modul: "github.com/tailscale/tailcat", datei: "go.mod"},

	// **Sicherungen mit Kennwort** — OPS-06
	// (#1723).
	"filippo.io/age": {modul: "filippo.io/age", datei: "go.mod"},
}

// goZeile ist die Tabellenzeile, die nicht über eine tool-Direktive gepinnt ist,
// sondern über die `go`-Direktive des Wurzelmoduls.
const goZeile = "Go"

// ungeprueft nennt die Tabellenzeilen, die dieser Wächter nicht prüft, samt
// Grund. Ein leerer Grund ist keiner — der Text steht hier, damit eine
// Auslassung eine Entscheidung bleibt und nicht zur Gewohnheit wird.
var ungeprueft = map[string]string{
	"GitHub Actions": "je Commit-Hash statt fester Version; ein Hash lässt sich hier gegen nichts abgleichen",
}

// werkzeugdateien nennt die go.mod-Dateien mit tool-Direktiven, in fester
// Reihenfolge — eine Kartenwanderung machte die Meldungen unvergleichbar.
var werkzeugdateien = []string{"tools/go.mod"}

// ---------------------------------------------------------------- Die Prüfungen

func TestVersionstabelleStimmtMitDerPinnungUeberein(t *testing.T) {
	tabelle := versionstabelle(t)
	module := werkzeugmodule(t)
	wurzelmodul := gomodLesen(t, "go.mod")

	for _, zeile := range tabelle {
		if grund, ausgelassen := ungeprueft[zeile.werkzeug]; ausgelassen {
			t.Run(zeile.werkzeug+" wird hier nicht geprüft", func(t *testing.T) {
				t.Skip(grund)
			})

			continue
		}

		if zeile.werkzeug == goZeile {
			t.Run("Go ist in go.mod gepinnt", func(t *testing.T) {
				pruefeVersion(t, "go.mod", zeile, wurzelmodul.goVersion)
			})

			continue
		}

		if _, imFrontend := frontendorte[zeile.werkzeug]; imFrontend {
			continue // geprüft in frontendpinnung_test.go
		}

		ort, bekannt := gepruefteWerkzeuge[zeile.werkzeug]
		if !bekannt {
			t.Errorf("Tabellenzeile %q hat keine Entsprechung: weder eine Pinnung in "+
				"gepruefteWerkzeuge noch eine begründete Auslassung in ungeprueft", zeile.werkzeug)

			continue
		}

		t.Run(zeile.werkzeug+" ist in "+ort.datei+" gepinnt", func(t *testing.T) {
			gepinnt := module[ort.datei].version(ort.modul)
			if ort.datei == "go.mod" {
				gepinnt = wurzelmodul.version(ort.modul)
			}

			pruefeVersion(t, ort.datei, zeile, gepinnt)
		})
	}
}

func TestJedesGepinnteWerkzeugStehtInDerTabelle(t *testing.T) {
	tabelle := versionstabelle(t)
	module := werkzeugmodule(t)

	genannt := make(map[string]string, len(tabelle))
	for _, zeile := range tabelle {
		genannt[zeile.werkzeug] = zeile.version
	}

	for _, datei := range werkzeugdateien {
		for _, pfad := range module[datei].werkzeuge {
			name, bekannt := tabellenname(pfad)
			t.Run("die Pinnung "+pfad+" steht in der Tabelle", func(t *testing.T) {
				if !bekannt {
					t.Fatalf("%s pinnt %q, aber gepruefteWerkzeuge kennt den Pfad nicht — "+
						"neues Werkzeug ohne Zeile in AGENTS.md §1?", datei, pfad)
				}
				if _, steht := genannt[name]; !steht {
					t.Fatalf("%s pinnt %q, aber die Tabelle in AGENTS.md §1 nennt %q nicht",
						datei, pfad, name)
				}
			})
		}
	}
}

func TestJedeTabellenzeileIstGeprueftOderBegruendetAusgelassen(t *testing.T) {
	for _, zeile := range versionstabelle(t) {
		t.Run("die Zeile "+zeile.werkzeug+" ist zugeordnet", func(t *testing.T) {
			_, geprueft := gepruefteWerkzeuge[zeile.werkzeug]
			grund, ausgelassen := ungeprueft[zeile.werkzeug]

			_, imFrontend := frontendorte[zeile.werkzeug]

			switch {
			case zeile.werkzeug == goZeile:
			case imFrontend && (geprueft || ausgelassen):
				t.Fatalf("%q wird in frontendpinnung_test.go geprüft und ist hier "+
					"zugleich geführt", zeile.werkzeug)
			case imFrontend:
			case geprueft && ausgelassen:
				t.Fatalf("%q gilt zugleich als geprüft und als ausgelassen", zeile.werkzeug)
			case geprueft:
			case ausgelassen && strings.TrimSpace(grund) == "":
				t.Fatalf("%q ist ausgelassen, aber ohne Grund", zeile.werkzeug)
			case ausgelassen:
			default:
				t.Fatalf("%q ist weder geprüft noch begründet ausgelassen", zeile.werkzeug)
			}
		})
	}
}

// pruefeVersion vergleicht Version und Ort einer Tabellenzeile mit dem, was
// tatsächlich in der go.mod-Datei steht.
func pruefeVersion(t *testing.T, datei string, zeile tabellenzeile, gepinnt string) {
	t.Helper()

	if gepinnt == "" {
		t.Fatalf("%s pinnt %q nicht — AGENTS.md §1 nennt %s",
			datei, zeile.werkzeug, zeile.version)
	}
	if gepinnt != zeile.version {
		t.Errorf("%s: AGENTS.md §1 nennt %s, gepinnt ist %s",
			zeile.werkzeug, zeile.version, gepinnt)
	}
	if !strings.Contains(zeile.ort, datei) {
		t.Errorf("%s: AGENTS.md §1 nennt als Ort der Pinnung %q, tatsächlich steht sie in %s",
			zeile.werkzeug, zeile.ort, datei)
	}
}

// tabellenname sucht zu einem Werkzeugpfad den Namen, unter dem er in der
// Tabelle geführt wird.
func tabellenname(pfad string) (string, bool) {
	for name, ort := range gepruefteWerkzeuge {
		if ort.modul == pfad {
			return name, true
		}
	}

	return "", false
}

// ------------------------------------------------------------- AGENTS.md lesen

// tabellenzeile ist eine Zeile der Versionstabelle aus AGENTS.md §1.
type tabellenzeile struct {
	werkzeug string
	version  string
	ort      string
}

// versionstabelle liest die Tabelle „Werkzeug | Version | Ort der Pinnung" aus
// dem Abschnitt §1 von AGENTS.md. Bewusst nur aus §1: die Datei enthält weitere
// Tabellen, und eine davon versehentlich mitzulesen ergäbe stille Falschbefunde.
func versionstabelle(t *testing.T) []tabellenzeile {
	t.Helper()

	roh := projektdatei(t, "AGENTS.md")

	var (
		zeilen      []tabellenzeile
		imAbschnitt bool
		inTabelle   bool
	)

	for _, roheZeile := range strings.Split(string(roh), "\n") {
		switch {
		case strings.HasPrefix(roheZeile, "## 1 "):
			imAbschnitt = true

			continue
		case strings.HasPrefix(roheZeile, "## "):
			imAbschnitt = false

			continue
		}

		if !imAbschnitt {
			continue
		}

		if !strings.HasPrefix(strings.TrimSpace(roheZeile), "|") {
			inTabelle = false

			continue
		}

		spalten := tabellenspalten(roheZeile)
		if len(spalten) != 3 {
			continue
		}

		switch {
		case spalten[0] == "Werkzeug":
			inTabelle = true
		case !inTabelle, strings.HasPrefix(spalten[0], "---"):
		default:
			zeilen = append(zeilen, tabellenzeile{
				werkzeug: ohneAuszeichnung(spalten[0]),
				version:  ohneAuszeichnung(spalten[1]),
				ort:      spalten[2],
			})
		}
	}

	if len(zeilen) == 0 {
		t.Fatal("in AGENTS.md §1 keine Versionstabelle gefunden")
	}

	return zeilen
}

// tabellenspalten zerlegt eine Markdown-Tabellenzeile in ihre Zellen.
func tabellenspalten(zeile string) []string {
	teile := strings.Split(strings.Trim(strings.TrimSpace(zeile), "|"), "|")

	spalten := make([]string, 0, len(teile))
	for _, teil := range teile {
		spalten = append(spalten, strings.TrimSpace(teil))
	}

	return spalten
}

// ohneAuszeichnung entfernt die Markdown-Auszeichnung aus einer Zelle.
func ohneAuszeichnung(zelle string) string {
	return strings.TrimSpace(strings.ReplaceAll(zelle, "`", ""))
}
