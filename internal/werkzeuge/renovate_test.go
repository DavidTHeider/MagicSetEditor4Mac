// Wächter über `renovate.json5` — ENG-04.AK-02
// (#1729).
//
// **Die Pinnregeln aus AGENTS.md §1 bleiben wirksam**, verlangt das
// Kriterium — und Renovate kennt sie nur, wenn sie in seiner Konfiguration
// stehen. Eine Regel, die dort fehlt, fällt nicht auf: Renovate schlägt dann
// schlicht vor, was §1 zurückhält, und der Vorschlag sieht aus wie jeder
// andere. So stand es auf dem Dashboard, bevor es die Datei gab (0129).
//
// **Geprüft wird gegen die Quelle der Regel, nicht gegen eine Abschrift**: die
// Grenze für TypeScript ist der Peerbereich von `typescript-eslint` in der
// Sperrdatei, die für ESLint der von `eslint-plugin-jsx-a11y`, die für
// `@types/node` die Node-Fassung des Workflows. Hebt eines davon seine
// Grenze, wird dieser Wächter rot, bis `renovate.json5` folgt — und
// umgekehrt.
//
// # Was dieser Wächter nicht prüft
//
//   - **Ob Renovate die Datei so liest.** Gelesen wird hier ein Ausschnitt von
//     JSON5, nicht die Sprache; ob Renovate sie annimmt, zeigt sein Dashboard
//     (#1784) nach dem nächsten Lauf.
//   - **Ob eine Gruppe greift.** Dass die Regel dasteht, hält er; welches PR
//     daraus wird, entscheidet Renovate.
package werkzeuge

import (
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// renovateregel ist ein Eintrag aus `packageRules`, soweit er hier zählt.
type renovateregel struct {
	pakete          []string
	allowedVersions string
	gruppe          string
	rangeStrategy   string
	text            string
}

// Die Felder einer Regel, wie `renovate.json5` sie schreibt.
var (
	renovatePakete    = regexp.MustCompile(`matchPackageNames:\s*\[([^\]]*)\]`)
	renovateGrenze    = regexp.MustCompile(`allowedVersions:\s*'([^']*)'`)
	renovateGruppe    = regexp.MustCompile(`groupName:\s*'([^']*)'`)
	renovateStrategie = regexp.MustCompile(`rangeStrategy:\s*'([^']*)'`)
	renovateName      = regexp.MustCompile(`'([^']*)'`)
)

// renovateregeln zerlegt `packageRules` in seine Einträge — an den
// geschweiften Klammern der zweiten Ebene, Kommentarzeilen ausgenommen.
func renovateregeln(t *testing.T) []renovateregel {
	t.Helper()

	roh := string(projektdatei(t, "renovate.json5"))

	anfang := strings.Index(roh, "packageRules:")
	if anfang < 0 {
		t.Fatal("renovate.json5 hat keine packageRules")
	}

	var (
		regeln []renovateregel
		tiefe  int
		block  strings.Builder
	)

	for _, zeile := range strings.Split(roh[anfang:], "\n") {
		if strings.HasPrefix(strings.TrimSpace(zeile), "//") {
			continue
		}

		for _, zeichen := range zeile {
			switch zeichen {
			case '{':
				tiefe++
			case '}':
				tiefe--
			}

			if tiefe > 0 || zeichen == '}' {
				block.WriteRune(zeichen)
			}

			if zeichen == '}' && tiefe == 0 {
				regeln = append(regeln, regelLesen(block.String()))
				block.Reset()
			}
		}

		block.WriteRune('\n')

		if tiefe < 0 {
			break
		}
	}

	return regeln
}

// regelLesen liest die Felder eines Eintrags.
func regelLesen(text string) renovateregel {
	regel := renovateregel{text: text}

	if fund := renovatePakete.FindStringSubmatch(text); fund != nil {
		for _, name := range renovateName.FindAllStringSubmatch(fund[1], -1) {
			regel.pakete = append(regel.pakete, name[1])
		}
	}

	for ziel, ausdruck := range map[*string]*regexp.Regexp{
		&regel.allowedVersions: renovateGrenze,
		&regel.gruppe:          renovateGruppe,
		&regel.rangeStrategy:   renovateStrategie,
	} {
		if fund := ausdruck.FindStringSubmatch(text); fund != nil {
			*ziel = fund[1]
		}
	}

	return regel
}

// regelFuer gibt die letzte Regel, die ein Paket nennt und das Feld setzt —
// die letzte, weil bei Renovate eine spätere Regel eine frühere überschreibt.
func regelFuer(regeln []renovateregel, paket string, feld func(renovateregel) string) string {
	var wert string

	for _, regel := range regeln {
		for _, name := range regel.pakete {
			passt := name == paket || (strings.HasSuffix(name, "/**") && strings.HasPrefix(paket, strings.TrimSuffix(name, "**")))
			if passt && feld(regel) != "" {
				wert = feld(regel)
			}
		}
	}

	return wert
}

// obereGrenze gibt aus einem Peerbereich die kleinste ausgeschlossene
// Fassung: `>=4.8.4 <6.1.0` → `<6.1.0`, `^3 || … || ^9` → `<10.0.0`.
func obereGrenze(t *testing.T, bereich string) string {
	t.Helper()

	if fund := regexp.MustCompile(`<\s*(\d+\.\d+\.\d+)`).FindStringSubmatch(bereich); fund != nil {
		return "<" + fund[1]
	}

	hoechste := 0

	for _, fund := range regexp.MustCompile(`\^(\d+)`).FindAllStringSubmatch(bereich, -1) {
		hauptfassung, err := strconv.Atoi(fund[1])
		if err != nil {
			t.Fatalf("%q ist keine Hauptfassung: %v", fund[1], err)
		}

		hoechste = max(hoechste, hauptfassung)
	}

	if hoechste == 0 {
		t.Fatalf("aus %q lässt sich keine obere Grenze lesen", bereich)
	}

	return "<" + strconv.Itoa(hoechste+1) + ".0.0"
}

func TestRenovateHaeltDieGrenzenAusDenPeerbereichen(t *testing.T) {
	regeln := renovateregeln(t)
	sperre := sperrdatei(t)
	grenze := func(regel renovateregel) string { return regel.allowedVersions }

	for paket, quelle := range map[string]struct{ von, peer string }{
		"typescript": {von: "typescript-eslint", peer: "typescript"},
		"eslint":     {von: "eslint-plugin-jsx-a11y", peer: "eslint"},
		"@eslint/js": {von: "eslint-plugin-jsx-a11y", peer: "eslint"},
	} {
		t.Run(paket, func(t *testing.T) {
			bereich := sperre[quelle.von].PeerDependencies[quelle.peer]
			if bereich == "" {
				t.Fatalf("%s nennt keinen Peerbereich für %s — die Grenze aus §1 hat ihren Grund verloren", quelle.von, quelle.peer)
			}

			if erwartet, gesetzt := obereGrenze(t, bereich), regelFuer(regeln, paket, grenze); gesetzt != erwartet {
				t.Errorf("renovate.json5 setzt für %s %q, der Peerbereich %q von %s verlangt %q",
					paket, gesetzt, bereich, quelle.von, erwartet)
			}
		})
	}

	// **Die Typen folgen der Laufzeit** (§1): Node aus dem Workflow.
	node := nodeVersion(t)
	hauptfassung, err := strconv.Atoi(strings.SplitN(node, ".", 2)[0])
	if err != nil {
		t.Fatalf("Node %q hat keine Hauptfassung: %v", node, err)
	}

	if erwartet, gesetzt := "<"+strconv.Itoa(hauptfassung+1)+".0.0", regelFuer(regeln, "@types/node", grenze); gesetzt != erwartet {
		t.Errorf("renovate.json5 setzt für @types/node %q, Node %s verlangt %q", gesetzt, node, erwartet)
	}
}

// TestRenovateHebtGekoppeltePaketeZusammen hält die Gruppen, in denen §1 zwei
// Pakete an **eine** Zahl bindet — getrennt gehoben gäbe es keine Auflösung.
func TestRenovateHebtGekoppeltePaketeZusammen(t *testing.T) {
	regeln := renovateregeln(t)
	gruppe := func(regel renovateregel) string { return regel.gruppe }
	strategie := func(regel renovateregel) string { return regel.rangeStrategy }

	for _, paar := range [][2]string{
		{"@mantine/core", "@mantine/hooks"},
		{"vitest", "@vitest/coverage-v8"},
		{"github.com/wailsapp/wails/v3", "@wailsio/runtime"},
	} {
		erste, zweite := regelFuer(regeln, paar[0], gruppe), regelFuer(regeln, paar[1], gruppe)
		if erste == "" || erste != zweite {
			t.Errorf("%s und %s stehen in den Gruppen %q und %q — sie gehören in eine", paar[0], paar[1], erste, zweite)
		}

		for _, paket := range paar {
			if regelFuer(regeln, paket, strategie) != "pin" {
				t.Errorf("%s ist nicht auf die Stelle genau gepinnt (rangeStrategy 'pin')", paket)
			}
		}
	}

	// **Und PixiJS für sich**: es ändert das Bild, und das sieht nur eine
	// Sichtprüfung — gebündelt mit anderem ginge sie unter.
	pixi := regelFuer(regeln, "pixi.js", gruppe)
	if pixi == "" {
		t.Fatal("pixi.js hat keine eigene Gruppe")
	}

	for _, regel := range regeln {
		if regel.gruppe == pixi && (len(regel.pakete) != 1 || regel.pakete[0] != "pixi.js") {
			t.Errorf("die Gruppe %q fasst mehr als pixi.js: %v", pixi, regel.pakete)
		}
	}
}

// TestRenovateMergtNieVonSelbst hält die drei Sätze aus 0129, an denen der
// ganze Ablauf hängt: nichts ohne Freigabe, nichts von selbst gemergt, und Go
// selbst gar nicht — es hängt über `GOTOOLCHAIN` an drei Stellen (§1).
func TestRenovateMergtNieVonSelbst(t *testing.T) {
	roh := string(projektdatei(t, "renovate.json5"))

	for _, satz := range []string{"automerge: false", "':dependencyDashboardApproval'"} {
		if !strings.Contains(roh, satz) {
			t.Errorf("renovate.json5 enthält %q nicht mehr", satz)
		}
	}

	for _, regel := range renovateregeln(t) {
		if strings.Contains(regel.text, "matchDatasources: ['golang-version']") &&
			strings.Contains(regel.text, "enabled: false") {
			return
		}
	}

	t.Error("renovate.json5 schaltet die Datenquelle golang-version nicht ab")
}
