package app

// **Die drei Ziele aus Schnitt 2** — Claude Desktop, Claude Code und der
// Dienst als LaunchAgent, OPS-09b
// (#1961,
// Vermerk [0159](../../docs/entscheidungen/0159-drei-ziele-ohne-block.md)).

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Der Name, unter dem dieses Programm sich bei den Klienten einträgt, und
// das Etikett seines LaunchAgents.
const (
	mcpName      = "magicseteditor4mac"
	dienstlabel  = Kennung + ".dienst"
	nichtGeladen = 3 // `launchctl bootout` für einen Dienst, der nicht geladen ist
)

// zweiteZiele baut die drei Ziele über Zuhause, Werkzeugort und `PATH`.
func zweiteZiele(heim, werkzeug string, pfad []string, ausfuehren befehlsausfuehrung) []service.Schnittstellenziel {
	stuetze := filepath.Join(heim, "Library", "Application Support")
	eigene := filepath.Join(stuetze, "MagicSetEditor4Mac", "schnittstellen")
	mcp := filepath.Join(werkzeug, "mse-mcp")
	claude := programmIn(pfad, "claude")
	plist := filepath.Join(heim, "Library", "LaunchAgents", dienstlabel+".plist")
	domaene := "gui/" + strconv.Itoa(os.Getuid())
	ansage := filepath.Join(eigene, "dienst.txt")

	return []service.Schnittstellenziel{
		{
			Name: "claudedesktop", Art: service.ArtJSON,
			Datei:      filepath.Join(stuetze, "Claude", "claude_desktop_config.json"),
			Vorhanden:  ordnerDa(filepath.Join(stuetze, "Claude")),
			Schluessel: mcpName, Programm: mcp,
			Gedaechtnis: filepath.Join(eigene, "claudedesktop.json"),
		},
		{
			Name: "claudecode", Art: service.ArtBefehl,
			Datei:      filepath.Join(heim, ".claude.json"),
			Vorhanden:  claude != "" || ordnerDa(filepath.Join(heim, ".claude")),
			Schluessel: mcpName, Programm: mcp,
			Einrichtbefehle: [][]string{{claudeOder(claude), "mcp", "add", "-s", "user", mcpName, mcp}},
			Widerrufbefehle: [][]string{{claudeOder(claude), "mcp", "remove", "-s", "user", mcpName}},
			Ausfuehren:      ausfuehren,
		},
		{
			Name: "dienst", Art: service.ArtAgent, Datei: plist, Vorhanden: true,
			Schluessel: dienstlabel, Begleitordner: eigene, Ansage: ansage,
			Block:           dienstplist(filepath.Join(werkzeug, "mse-rest"), ansage, filepath.Join(eigene, "dienst.log")),
			Einrichtbefehle: [][]string{{"launchctl", "bootstrap", domaene, plist}},
			Widerrufbefehle: [][]string{{"launchctl", "bootout", domaene + "/" + dienstlabel}},
			Altlasten:       []service.Altagent{altdienst(heim, domaene)},
			Ausfuehren:      duldetNichtGeladen(ausfuehren),
		},
	}
}

// altdienst ist der Dienst unter dem Etikett von vor 0.9.0 — **eine
// Beschreibung für Widerruf und Betriebsband** (Zweitblick über Phase 4,
// Befund 7). Die plist von damals nennt `mse-rest` wie die heutige.
func altdienst(heim, domaene string) service.Altagent {
	return service.Altagent{
		Datei:      filepath.Join(heim, "Library", "LaunchAgents", altdienstlabel+".plist"),
		Schluessel: altdienstlabel, Programm: "mse-rest",
		Widerrufbefehle: [][]string{{"launchctl", "bootout", domaene + "/" + altdienstlabel}},
	}
}

// befehlsausfuehrung führt einen Befehl aus.
type befehlsausfuehrung func(ctx context.Context, befehl []string) error

// befehlAusfuehren ist die echte Ausführung — **mit der Ausgabe im Grund**
// und nicht im Satz: dort stünden Pfade, und die erreichen die Oberfläche
// nicht (§8).
//
// **Das Programm ist ein Literal** (`internal/schichten/unterprozess_test.go`):
// `launchctl` aus `/bin`, `claude` aus dem Ordner, in dem die Anmeldeshell
// ihn fand. Ein anderer Name startet nichts, und ein `claude` ohne Ort auch
// nicht — sonst liefe ein `claude` aus dem Arbeitsverzeichnis.
func befehlAusfuehren(ctx context.Context, befehl []string) error {
	var programm *exec.Cmd

	switch name := filepath.Base(befehl[0]); {
	case befehl[0] == "launchctl" && (testlauf || os.Getenv(Dienstsperre) == "an") && !nurLesendesLaunchctl(befehl):
		// **Ein Testlauf verändert nie den Dienst dieses Rechners** —
		// Zweitblick über Phase 2, Befund 6: das Ziel ist `gui/<UID>/<Etikett>`,
		// und ein umgelenktes Zuhause lenkt das nicht um. **Und ein gebautes
		// Kind eines Laufs auch nicht** (Nachlauf, Befund 6): es heißt nicht
		// `*.test` und bekommt dafür [Dienstsperre].
		return domain.MitGrund(domain.ErrValidation, nil, "ein Testlauf ruft launchctl %s nicht", befehl[1])
	case befehl[0] == "launchctl":
		programm = exec.CommandContext(ctx, "/bin/launchctl", befehl[1:]...) // #nosec G204 -- die Argumente setzt diese Anwendung zusammen
	case befehl[0] == "defaults" && (testlauf || os.Getenv(Dienstsperre) == "an"):
		// **Und die Vorgaben dieses Rechners auch nicht** (R02, #2334): der
		// Umzug der Kennung liest und schreibt sie über cfprefsd, und der
		// folgt nicht jedem umgelenkten Zuhause.
		return domain.MitGrund(domain.ErrValidation, nil, "ein Testlauf ruft defaults %s nicht", befehl[1])
	case befehl[0] == "defaults":
		programm = exec.CommandContext(ctx, "/usr/bin/defaults", befehl[1:]...) // #nosec G204 -- die Argumente setzt diese Anwendung zusammen
	case befehl[0] == "plutil":
		// **Ohne Sperre**: `plutil` schreibt nur die Datei, die es genannt
		// bekommt — beim Umzug der Kennung eine eigene im Zwischenordner (R02).
		programm = exec.CommandContext(ctx, "/usr/bin/plutil", befehl[1:]...) // #nosec G204 -- die Argumente setzt diese Anwendung zusammen
	case name == "claude" && filepath.IsAbs(befehl[0]):
		ordner := filepath.Dir(befehl[0])
		programm = exec.CommandContext(ctx, filepath.Join(ordner, "claude"), befehl[1:]...) // #nosec G204 -- der Name steht hier, der Ordner kommt aus dem PATH der Anmeldeshell
	default:
		return domain.MitGrund(domain.ErrValidation, nil,
			"%s startet diese Anwendung nicht — steht es im PATH der Anmeldeshell?", name)
	}

	if ausgabe, fehler := programm.CombinedOutput(); fehler != nil {
		return domain.MitGrund(domain.ErrConflict, fmt.Errorf("%w: %s", fehler, ausgabe),
			"%s hat das nicht angenommen", filepath.Base(befehl[0]))
	}

	return nil
}

// Dienstsperre ist die Umgebungsvariable, mit der ein Lauf seinen gebauten
// Kindprozessen jedes verändernde `launchctl` sperrt — dieselbe Form wie die
// Sperre für Apples Gerätemodell (`applefm.Sperre`).
const Dienstsperre = "MSE_DIENST_GESPERRT"

// nurLesendesLaunchctl sagt, ob ein Ruf an `launchctl` nur fragt.
func nurLesendesLaunchctl(befehl []string) bool {
	return len(befehl) > 1 && (befehl[1] == "version" || befehl[1] == "print" || befehl[1] == "list")
}

// duldetNichtGeladen lässt `launchctl bootout` für einen Dienst durch, der
// nicht geladen ist — sonst bliebe nach einem Neustart die plist stehen, weil
// sich nichts entladen ließ.
func duldetNichtGeladen(ausfuehren befehlsausfuehrung) befehlsausfuehrung {
	return func(ctx context.Context, befehl []string) error {
		fehler := ausfuehren(ctx, befehl)

		var beendet interface{ ExitCode() int }
		if len(befehl) > 1 && befehl[1] == "bootout" && errors.As(fehler, &beendet) &&
			beendet.ExitCode() == nichtGeladen {
			return nil
		}

		return fehler
	}
}

// programmIn sucht ein Programm im `PATH` — leer, wenn es dort keines gibt.
func programmIn(pfad []string, name string) string {
	for _, eintrag := range pfad {
		ort := filepath.Join(eintrag, name)
		if angaben, fehler := os.Stat(ort); fehler == nil && !angaben.IsDir() {
			return ort
		}
	}

	return ""
}

// claudeOder gibt den gefundenen `claude` — oder den bloßen Namen, damit
// eine Vorschau etwas nennt.
func claudeOder(gefunden string) string {
	if gefunden == "" {
		return "claude"
	}

	return gefunden
}

// dienstplist ist die plist des LaunchAgents.
//
// **Ohne Port und ohne Zugang**: `mse-rest` sucht einen freien Port und sagt
// ihn samt Marke an. In der plist steht damit nichts, was ein anderer Mensch
// am Rechner nicht lesen dürfte. **Neu gestartet nur nach einem Fehler**:
// endet der Dienst, weil sein Schalter aus ist (OPS-10), bleibt er aus.
//
// **Die Ansage schreibt das Programm selbst** (SEC02, #2355, Vermerk 0192):
// `-ansage` nennt die Datei, und `mse-rest` ersetzt sie je Start mit 0600.
// Bis dahin zeigten `StandardOutPath` und `StandardErrorPath` auf dieselbe
// Datei, und launchd hängte an — jede Marke seit der Einrichtung stand darin.
// **Die Standardausgabe geht deshalb ins Leere**, und die Fehlerausgabe in
// eine eigene Datei, die das Programm beim Start begrenzt.
func dienstplist(programm, ansage, protokoll string) string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>` + dienstlabel + `</string>
	<key>ProgramArguments</key>
	<array>
		<string>` + xmlText(programm) + `</string>
		<string>-ansage</string>
		<string>` + xmlText(ansage) + `</string>
	</array>
	<key>RunAtLoad</key>
	<true/>
	<key>KeepAlive</key>
	<dict>
		<key>SuccessfulExit</key>
		<false/>
	</dict>
	<key>StandardOutPath</key>
	<string>/dev/null</string>
	<key>StandardErrorPath</key>
	<string>` + xmlText(protokoll) + `</string>
</dict>
</plist>
`
}

// xmlText maskiert einen Pfad für die plist — ein `&` im Namen eines Ordners
// machte sie sonst unlesbar.
func xmlText(text string) string {
	var maskiert strings.Builder
	_ = xml.EscapeText(&maskiert, []byte(text)) //nolint:errcheck // ein strings.Builder schreibt ohne Fehler

	return maskiert.String()
}

// vorschauzeilen gibt, was einrichten schreibt oder ausführt — je Art (AK-01:
// jede Datei und jede Zeile vorher).
func vorschauzeilen(ziel service.Schnittstellenziel) []string {
	switch ziel.Art {
	case service.ArtJSON:
		programm, _ := json.Marshal(ziel.Programm) //nolint:errcheck // eine Zeichenkette lässt sich immer schreiben

		return []string{
			`"mcpServers": {`,
			`  "` + ziel.Schluessel + `": {"command": ` + string(programm) + `, "args": []}`,
			"}",
		}
	case service.ArtVerknuepfung:
		return ziel.Verknuepfung.Zeilen
	case service.ArtBefehl, service.ArtAgent:
		var zeilen []string
		if ziel.Art == service.ArtAgent {
			zeilen = strings.Split(strings.TrimRight(ziel.Block, "\n"), "\n")
		}

		for _, befehl := range ziel.Einrichtbefehle {
			zeilen = append(zeilen, "$ "+strings.Join(befehl, " "))
		}

		return zeilen
	}

	return strings.Split(domain.BlockAnfang+"\n"+ziel.Block+domain.BlockEnde, "\n")
}
