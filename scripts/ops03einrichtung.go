//go:build ignore

// **Die Einrichtung, wie das Fenster sie schreibt — für einen Gast** — OPS03
// (#2349).
//
//	go run scripts/ops03einrichtung.go <zuhause> <Werkzeugort im Gast>
//
// Richtet unter `<zuhause>` ein, was *Alles einrichten* am Fenster anlegt —
// **mit dem Code der Anwendung selbst** (`App.SchnittstellenAlleEinrichten`),
// nicht mit einer Abschrift: die Verweise in `~/.local/bin`, den Block im
// Shellprofil, den Eintrag für Codex und für Claude Desktop samt Gedächtnis,
// die plist des Dienstes. **Ausgeführt wird nichts**: die Befehle, die das
// Einrichten an `launchctl` gäbe, stehen in `<zuhause>/befehle.txt`, und der
// Gast führt sie aus (`scripts/huellenprobe/gast/ops03.pl`).
//
// **Fremde Nachbarn** liegen vorher in jeder Datei; der Gast vergleicht sie
// nach dem Widerruf Byte für Byte mit `<zuhause>/nachbarn.json`.
//
// **Claude Code fehlt mit Absicht**: im Gast gibt es kein `claude`, und ohne
// es bliebe der Eintrag stehen — das steht in der Administration.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "Aufruf: go run scripts/ops03einrichtung.go <zuhause> <Werkzeugort im Gast>")
		os.Exit(2)
	}

	heim, werkzeug := os.Args[1], os.Args[2]

	nachbarn := map[string]string{
		".zprofile":                "export EDITOR=vi\n",
		".codex/config.toml":       "[mcp_servers.anderer]\ncommand = \"/opt/anderer\"\n",
		".local/bin/mein-werkzeug": "#!/bin/sh\n",
		"Library/Application Support/Claude/claude_desktop_config.json": "{\n  \"mcpServers\": {\n    \"anderer\": {\"command\": \"/opt/anderer\"}\n  }\n}\n",
		"Library/LaunchAgents/de.anderer.plist":                         "<plist/>\n",
	}

	for name, inhalt := range nachbarn {
		pfad := filepath.Join(heim, name)
		if err := os.MkdirAll(filepath.Dir(pfad), 0o700); err != nil {
			panic(err)
		}

		if err := os.WriteFile(pfad, []byte(inhalt), 0o600); err != nil {
			panic(err)
		}
	}

	var befehle []string

	anwendung := app.NeueApp("").MitSchnittstellenorten(
		func() (string, error) { return heim, nil },
		func() (string, error) { return werkzeug, nil },
		func(context.Context) string { return "/usr/bin:/bin" },
	).MitSchnittstellenbefehlen(func(_ context.Context, befehl []string) error {
		befehle = append(befehle, strings.Join(befehl, "\t"))

		return nil
	})

	antwort := anwendung.SchnittstellenAlleEinrichten(context.Background())
	if antwort.Fehler != nil {
		panic(antwort.Fehler.Meldung)
	}

	for _, ausgang := range antwort.Ausgaenge {
		fmt.Printf("%s\t%s\n", ausgang.Name, ausgang.Ausgang)
	}

	roh, err := json.MarshalIndent(nachbarn, "", "  ")
	if err != nil {
		panic(err)
	}

	if err := os.WriteFile(filepath.Join(heim, "nachbarn.json"), roh, 0o600); err != nil {
		panic(err)
	}

	if err := os.WriteFile(filepath.Join(heim, "befehle.txt"), []byte(strings.Join(befehle, "\n")+"\n"), 0o600); err != nil {
		panic(err)
	}
}
