// **Die Kennung ist `io.github.davidtheider.magicseteditor4mac`** — R02
// (#2334),
// entschieden in [0191](../../docs/entscheidungen/0191-die-kennung-ist-io-github-davidtheider.md).
//
// # Was der Wächter hält
//
// Bündel, Abnahmebündel, LaunchAgent und Dokumenttyp tragen die neue
// Kennung, und **die alte steht nur dort, wo sie gesucht wird** — in der
// Liste der Altkennungen, aus der der Umzug liest, und in der Erkennung eines
// alten eigenen Bündels. Eine dritte Stelle mit `com.wails.` ist ein Befund:
// sie schriebe unter der alten Kennung weiter.
//
// **Was er nicht hält**: ob der Umzug wirkt. Das zeigt der Lauf im Gast.
package werkzeuge_test

import (
	"os/exec"
	"strings"
	"testing"
)

// neueKennung ist die aus Vermerk 0191.
const neueKennung = "io.github.davidtheider.magicseteditor4mac"

// altkennungsorte sind die Dateien, in denen die alte Kennung stehen darf —
// je mit dem Grund.
var altkennungsorte = map[string]string{
	"internal/app/kennungsumzug.go":          "die Liste, aus der der Umzug liest",
	"internal/service/schnittstellenjson.go": "ein MCP-Eintrag in ein altes eigenes Bündel heißt veraltet und nicht fremd",
	"scripts/huellenorte.mjs":                "die Orte der alten Kennung gehören zu den echten Orten, die keine Probe anfassen darf",
	"scripts/kennunggast.mjs":                "der Lauf im Gast prüft, dass das alte Bündel die alte trägt",
	"scripts/huellenprobe/gast/kennung.pl":   "der Treiber legt den alten Stand an, von dem umgezogen wird",
	"scripts/m02gast.mjs":                    "der Dienst von 0.8.0 läuft unter dessen Etikett",
}

func TestDieNeueKennungStehtInBauUndDienst(t *testing.T) {
	t.Parallel()

	for datei, erwartet := range map[string]string{
		"build/darwin/Info.plist":             "<string>" + neueKennung + "</string>",
		"scripts/bauen.sh":                    `CFBundleIdentifier -string "` + neueKennung + `.abnahme"`,
		"scripts/huellenprobe.mjs":            `'` + neueKennung + `.abnahme'`,
		"internal/app/kennungsumzug.go":       `const Kennung = "` + neueKennung + `"`,
		"internal/app/schnittstellenziele.go": `dienstlabel  = Kennung + ".dienst"`,
	} {
		if !strings.Contains(lies(t, datei), erwartet) {
			t.Errorf("%s nennt %s nicht", datei, erwartet)
		}
	}

	if !strings.Contains(lies(t, "build/darwin/Info.plist"), "<string>"+neueKennung+".umschlag</string>") {
		t.Error("der Typ für .msepackage trägt die neue Kennung nicht")
	}
}

func TestDieAlteKennungStehtNurWoSieGesuchtWird(t *testing.T) {
	t.Parallel()

	// **Was git führt** — ein gebautes Bündel unter `build/bin` trägt die
	// Kennung, mit der es gebaut wurde, und gehört nicht dazu.
	ausgabe, err := exec.CommandContext(t.Context(), "git", "-C", "../..", "ls-files",
		"--", "build", "cmd", "internal", "scripts", "main.go").Output()
	if err != nil {
		t.Fatalf("git kennt diesen Baum nicht: %v", err)
	}

	gelesen := 0

	for name := range strings.SplitSeq(strings.TrimSpace(string(ausgabe)), "\n") {
		// **Tests dürfen sie nennen** — sie bauen die alte Lage nach.
		if strings.Contains(name, "_test.go") || strings.Contains(name, ".test.") || altkennungsorte[name] != "" {
			continue
		}

		gelesen++

		if strings.Contains(strings.ToLower(lies(t, name)), "com.wails.") {
			t.Errorf("%s nennt die alte Kennung com.wails.* — gehört sie in die Liste der Altkennungen?", name)
		}
	}

	if gelesen < 1000 {
		t.Fatalf("nur %d Dateien gelesen — der Ruf trifft den Baum nicht", gelesen)
	}
}
