// Wächter über die Frontend-Hälfte der Versionstabelle aus AGENTS.md §1.
//
// `pinnung_test.go` führte Node, npm, TypeScript und ESLint als „ungeprüft, weil
// das Frontend noch nicht existiert". Mit Issue #4 existiert es, und eine
// Auslassung, deren Grund weggefallen ist, ist nur noch eine Lücke. Also stehen
// die vier Zeilen jetzt hier.
//
// Die Fassungen stehen an vier verschiedenen Orten, und die Tabelle nennt sie:
//
//	Node        .github/workflows/ci.yml    NODE_VERSION
//	npm         frontend/package.json       packageManager
//	TypeScript  frontend/package-lock.json  node_modules/typescript
//	ESLint      frontend/package-lock.json  node_modules/eslint
//
// TypeScript und ESLint werden bewusst aus der **Sperrdatei** gelesen, nicht aus
// `package.json`: dort dürfen Bereiche stehen, und reproduzierbar ist der Stand
// über `npm ci`. Eine Prüfung gegen den Bereich prüfte die Absicht, nicht das,
// was installiert wird.
//
// # Was dieser Wächter nicht prüft
//
// Die Zeile „GitHub Actions" bleibt in `ungeprueft`: sie nennt keine Version,
// sondern „je Commit-Hash", und ein Hash lässt sich gegen nichts abgleichen, was
// im Repository steht. Wer prüfen will, ob ein Hash zur danebenstehenden Fassung
// gehört, braucht das Netz — das gehört nicht in einen Einheitstest.
package werkzeuge

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// frontendorte ordnet den vier Tabellenzeilen ihre Fundstelle zu.
var frontendorte = map[string]struct {
	datei   string
	ablesen func(t *testing.T) string
}{
	"Node":       {datei: ".github/workflows/", ablesen: nodeVersion},
	"npm":        {datei: "frontend/package.json", ablesen: npmVersion},
	"TypeScript": {datei: "frontend/package-lock.json", ablesen: sperrdateiVersion("typescript")},
	"ESLint":     {datei: "frontend/package-lock.json", ablesen: sperrdateiVersion("eslint")},

	// **Der gezeichnete Spieltisch** — Werkstück G3 von Phase 5.5
	// (#1460).
	//
	// **Aus der Sperrdatei und nicht aus `package.json`**, aus demselben Grund
	// wie TypeScript und ESLint: dort steht, was gewünscht war, hier steht,
	// was installiert wird. Bei einer Grafikbibliothek ist der Unterschied
	// nicht akademisch — eine andere Fassung zeichnet ein anderes Bild.
	"PixiJS": {datei: "frontend/package-lock.json", ablesen: sperrdateiVersion("pixi.js")},

	// **Die Bauteilbibliothek** — UX-07
	// (#1679,
	// [0125](../../docs/entscheidungen/0125-mantine-als-bauteilbibliothek.md)).
	//
	// **Gelesen wird `@mantine/core`**, und `@mantine/hooks` trägt dieselbe
	// Zahl — sein Peerbereich verlangt sie **ohne Spielraum**. Zwei Zeilen in
	// der Tabelle wären zwei Angaben über eine Zahl; der Lauf daneben hält
	// stattdessen, dass beide gleich stehen.
	"Mantine": {datei: "frontend/package-lock.json", ablesen: sperrdateiVersion("@mantine/core")},

	// **Der Testläufer der Oberfläche** — ENG-04
	// (#1729).
	//
	// **Aus der Sperrdatei**, wie die vier darüber. Und aus demselben Grund
	// wie bei Mantine steht nur **eine** Zeile in der Tabelle:
	// `@vitest/coverage-v8` trägt dieselbe Zahl, und der Lauf unten hält sie
	// zusammen.
	"Vitest": {datei: "frontend/package-lock.json", ablesen: sperrdateiVersion("vitest")},
}

func TestFrontendVersionstabelleStimmtMitDerPinnungUeberein(t *testing.T) {
	for _, zeile := range versionstabelle(t) {
		ort, gehoertHierher := frontendorte[zeile.werkzeug]
		if !gehoertHierher {
			continue
		}

		t.Run(zeile.werkzeug+" ist in "+ort.datei+" gepinnt", func(t *testing.T) {
			gepinnt := ort.ablesen(t)

			if gepinnt != zeile.version {
				t.Errorf("%s: AGENTS.md §1 nennt %s, gepinnt ist %s",
					zeile.werkzeug, zeile.version, gepinnt)
			}
			if !strings.Contains(zeile.ort, ort.datei) {
				t.Errorf("%s: AGENTS.md §1 nennt als Ort der Pinnung %q, tatsächlich steht sie in %s",
					zeile.werkzeug, zeile.ort, ort.datei)
			}
		})
	}
}

// TestKeineFrontendzeileGiltNochAlsUngeprueft hält fest, dass die vier Zeilen
// nicht doppelt geführt werden.
//
// Ohne ihn ließe sich eine Zeile hier prüfen und in `ungeprueft` zugleich
// entschuldigen — und dann bliebe sie beim nächsten Umbau als „ungeprüft"
// stehen, während dieser Wächter stillschweigend weiterliefe.
func TestKeineFrontendzeileGiltNochAlsUngeprueft(t *testing.T) {
	for werkzeug := range frontendorte {
		t.Run(werkzeug+" steht nicht mehr in ungeprueft", func(t *testing.T) {
			if grund, entschuldigt := ungeprueft[werkzeug]; entschuldigt {
				t.Fatalf("%q wird hier geprüft, gilt in pinnung_test.go aber als "+
					"ausgelassen mit dem Grund %q", werkzeug, grund)
			}
		})
	}
}

// nodeVersion liest NODE_VERSION aus dem Integrationsworkflow.
func nodeVersion(t *testing.T) string {
	t.Helper()

	roh := projektdatei(t, ".github/workflows/ci.yml")

	for _, zeile := range strings.Split(string(roh), "\n") {
		gestutzt := strings.TrimSpace(zeile)
		if !strings.HasPrefix(gestutzt, "NODE_VERSION:") {
			continue
		}

		wert := strings.TrimSpace(strings.TrimPrefix(gestutzt, "NODE_VERSION:"))

		return strings.Trim(wert, `"'`)
	}

	t.Fatal(".github/workflows/ci.yml nennt kein NODE_VERSION")

	return ""
}

// npmVersion liest das Feld packageManager aus frontend/package.json.
//
// Das Feld ist die Deklaration, nicht die Durchsetzung: ohne Corepack liest
// `npm ci` es gar nicht, und eine falsche Angabe liefe kommentarlos durch.
// Wirksam wird sie erst dadurch, dass die Stufe `format` `npm --version`
// dagegen hält — dieser Test prüft nur, dass Tabelle und Feld übereinstimmen.
func npmVersion(t *testing.T) string {
	t.Helper()

	var beschreibung struct {
		PackageManager string `json:"packageManager"`
	}
	if err := json.Unmarshal(projektdatei(t, "frontend/package.json"), &beschreibung); err != nil {
		t.Fatalf("frontend/package.json nicht lesbar: %v", err)
	}

	name, version, getrennt := strings.Cut(beschreibung.PackageManager, "@")
	if !getrennt || name != "npm" {
		t.Fatalf("packageManager ist %q, erwartet war npm@<version>", beschreibung.PackageManager)
	}

	return version
}

// sperrdateiVersion liefert eine Ableserfunktion für ein Paket aus package-lock.json.
func sperrdateiVersion(paket string) func(t *testing.T) string {
	return func(t *testing.T) string {
		t.Helper()

		var sperrdatei struct {
			Packages map[string]struct {
				Version string `json:"version"`
			} `json:"packages"`
		}
		if err := json.Unmarshal(projektdatei(t, "frontend/package-lock.json"), &sperrdatei); err != nil {
			t.Fatalf("frontend/package-lock.json nicht lesbar: %v", err)
		}

		eintrag, vorhanden := sperrdatei.Packages["node_modules/"+paket]
		if !vorhanden {
			t.Fatalf("frontend/package-lock.json führt kein node_modules/%s", paket)
		}

		return eintrag.Version
	}
}

// Die zugesagte macOS-Untergrenze und die des Bündels müssen dieselbe sein.
//
// AGENTS.md §12 sagt „ab macOS 13"; `build/darwin/Info.plist` sagt, was das
// Bündel wirklich verlangt. Eine Zusage in zwei Dateien läuft auseinander, und
// die im Bündel ist die, die gilt — beim Zweitblick am Ende von Phase 1 stand
// dort noch 10.13.0, also eine Spanne, die nie jemand geprüft hat.
func TestDieMacOSUntergrenzeStimmtMitDenLeitlinienUeberein(t *testing.T) {
	ausBuendel := normiert(untergrenzeAusPlist(t))

	// **Drei** Zusagestellen, nicht zwei.
	//
	// Der Nachlauf des Zweitblicks hat die dritte gefunden: `README.md` sagt
	// Anwendenden „ab macOS 13", und sie kann unabhängig weglaufen. Dieselbe
	// Zusage an drei Orten ist dreimal die Gelegenheit, sie zu brechen.
	//
	// **Und eine vierte, die keine Zusage ist, sondern die Einlösung**
	// (ENG-02): `scripts/macos.sh` sagt cgo, wofür es baut. Stünde dort eine
	// höhere Zahl, trüge das Bündel „ab macOS 13" und das Programm darin
	// liefe dort nicht.
	quellen := map[string]*regexp.Regexp{
		"AGENTS.md":        untergrenzeInAgents,
		"README.md":        untergrenzeInReadme,
		"scripts/macos.sh": untergrenzeFuerCgo,
	}

	for datei, ausdruck := range quellen {
		t.Run(datei, func(t *testing.T) {
			zusage := normiert(untergrenzeAus(t, datei, ausdruck))

			if zusage != ausBuendel {
				t.Errorf("%s sagt macOS %s, build/darwin/Info.plist verlangt %s",
					datei, zusage, ausBuendel)
			}
		})
	}
}

// **Und der Prüfstand nennt kein macOS unter der Grenze** — Befund 3 des
// Zweitblicks über Phase I: nach dem Heben auf 15.2 führte er weiter
// „macOS 13 bis 26" als vorhanden, in allen acht Sprachen. Er ist die
// fünfte Zusagestelle, und die einzige, die Anwendende im Fenster lesen.
//
// **Was er nicht prüft**: ob die genannten Fassungen belegt sind — das
// sagt der Beleg in der Zeile daneben.
func TestDerPruefstandNenntKeinMacOSUnterDerGrenze(t *testing.T) {
	grenze := normiert(untergrenzeAusPlist(t))

	for _, treffer := range macOSFassung.FindAllStringSubmatch(string(projektdatei(t, "internal/app/pruefstand.go")), -1) {
		if fassungKleiner(normiert(treffer[1]), grenze) {
			t.Errorf("der Prüfstand nennt %q, das Bündel verlangt macOS %s", treffer[0], grenze)
		}
	}
}

var macOSFassung = regexp.MustCompile(`macOS (\d+(?:\.\d+)*)`)

// fassungKleiner vergleicht zwei normierte Fassungen Zahl für Zahl.
func fassungKleiner(a, b string) bool {
	x, y := strings.Split(a, "."), strings.Split(b, ".")
	for i := range x {
		m, _ := strconv.Atoi(x[i]) //nolint:errcheck // normiert: nur Ziffern
		n, _ := strconv.Atoi(y[i]) //nolint:errcheck // normiert: nur Ziffern
		if m != n {
			return m < n
		}
	}

	return false
}

// normiert bringt „13" und „13.0" auf dieselbe Form.
//
// Die vollständige Angabe wird verglichen, nicht nur die führende Zahl: ein
// späteres „macOS 13.1" gegen ein Bündel mit 13.0 wäre sonst grün geblieben.
func normiert(version string) string {
	teile := strings.Split(version, ".")
	for len(teile) < 3 {
		teile = append(teile, "0")
	}

	return strings.Join(teile[:3], ".")
}

var (
	// `\d+(?:\.\d+)*` und nicht `[\d.]+`: der zweite Ausdruck liest den
	// Satzpunkt mit, und aus „ab macOS 13." wird „13.".
	untergrenzeInAgents = regexp.MustCompile(`Ziel ist \*\*darwin/arm64\*\*, ab macOS (\d+(?:\.\d+)*)`)
	untergrenzeInReadme = regexp.MustCompile(`macOS \(Apple Silicon\), ab macOS (\d+(?:\.\d+)*)`)
	untergrenzeFuerCgo  = regexp.MustCompile(`(?m)^export MACOSX_DEPLOYMENT_TARGET=(\d+(?:\.\d+)*)$`)
)

func untergrenzeAus(t *testing.T, datei string, ausdruck *regexp.Regexp) string {
	t.Helper()

	treffer := ausdruck.FindSubmatch(projektdatei(t, datei))
	if treffer == nil {
		t.Fatalf("%s nennt keine macOS-Untergrenze — hat sich der Wortlaut geändert?", datei)
	}

	return string(treffer[1])
}

var plistUntergrenze = regexp.MustCompile(
	`(?s)<key>LSMinimumSystemVersion</key>\s*<string>([^<]+)</string>`)

func untergrenzeAusPlist(t *testing.T) string {
	t.Helper()

	treffer := plistUntergrenze.FindSubmatch(projektdatei(t, "build/darwin/Info.plist"))
	if treffer == nil {
		t.Fatal("build/darwin/Info.plist nennt kein LSMinimumSystemVersion")
	}

	return string(treffer[1])
}

// **Und `@mantine/hooks` trägt dieselbe Zahl** — UX-07.
//
// Mantines Peerbereich verlangt es ohne Spielraum: `@mantine/core@9.6.3` nennt
// `"@mantine/hooks": "9.6.3"`, nicht einen Bereich. Zwei Zeilen in der Tabelle
// wären zwei Angaben über dieselbe Zahl und liefen auseinander; dieser Lauf
// hält sie zusammen, ohne der Tabelle eine zweite Zeile zu geben.
func TestMantinesHakenTragenDieselbeFassung(t *testing.T) {
	t.Parallel()

	kern := sperrdateiVersion("@mantine/core")(t)
	haken := sperrdateiVersion("@mantine/hooks")(t)

	if kern != haken {
		t.Errorf("@mantine/core steht bei %q, @mantine/hooks bei %q — sein "+
			"Peerbereich verlangt dieselbe Fassung ohne Spielraum", kern, haken)
	}
}

// **Und `@vitest/coverage-v8` trägt dieselbe Zahl** — ENG-04.
//
// # Warum das kein Schönheitsfehler ist
//
// **Der Peerbereich verlangt sie ohne Spielraum, und er ist nicht
// wahlfrei.** Am 2026-09-21 aus der Sperrdatei gelesen:
// `@vitest/coverage-v8` nennt `"vitest": "4.1.11"` und `"@vitest/browser":
// "4.1.11"`, und `peerDependenciesMeta` erklärt **nur den zweiten** für
// wahlfrei. Wer eines von beiden allein hebt, bekommt keine Auflösung —
// nicht eine Warnung, sondern einen Abbruch.
//
// **Was dieser Lauf nicht hält**: dass die Fassung frei von bekannten Lücken
// ist. Das tut `npm audit` in der Stufe `security`, und **wie weit es beißt**,
// steht in [`HAERTUNG.md`](../../docs/HAERTUNG.md) — ab `high` und nicht
// darunter.
func TestVitestsAbdeckungTraegtDieselbeFassung(t *testing.T) {
	t.Parallel()

	laeufer := sperrdateiVersion("vitest")(t)
	abdeckung := sperrdateiVersion("@vitest/coverage-v8")(t)

	if laeufer != abdeckung {
		t.Errorf("vitest steht bei %q, @vitest/coverage-v8 bei %q — sein "+
			"Peerbereich verlangt dieselbe Fassung ohne Spielraum", laeufer, abdeckung)
	}
}

// **Und `@wailsio/runtime` trägt die Fassung der Hülle** — ENG-02 (#1727).
//
// # Warum das kein Schönheitsfehler ist
//
// Die Oberfläche ruft die Anwendung über dieses Paket (`anbindung.ts`), und
// am anderen Ende antwortet der Go-Teil derselben Bibliothek. **Beide sprechen
// ein Protokoll, das keine Fassungszusage hat** — Wails 3 ist eine
// Vorabfassung (0124). Laufen die zwei auseinander, bricht nicht der Bau,
// sondern ein Ruf zur Laufzeit, und den sieht nur das Fenster.
//
// **Was dieser Lauf nicht hält**: dass beide Hälften wirklich miteinander
// reden. Das zeigt der Nahtlauf im Browser (`make dev-browser`) und am Ende
// das Fenster (`docs/WAILS3.md`).
func TestDieLaufzeitDerOberflaecheTraegtDieFassungDerHuelle(t *testing.T) {
	t.Parallel()

	oberflaeche := sperrdateiVersion("@wailsio/runtime")(t)
	huelle := strings.TrimPrefix(gomodLesen(t, "go.mod").version("github.com/wailsapp/wails/v3"), "v")

	if huelle == "" {
		t.Fatal("go.mod nennt github.com/wailsapp/wails/v3 nicht")
	}

	if oberflaeche != huelle {
		t.Errorf("@wailsio/runtime steht bei %q, github.com/wailsapp/wails/v3 bei %q — "+
			"beide Hälften sprechen dasselbe Protokoll ohne Fassungszusage", oberflaeche, huelle)
	}
}
