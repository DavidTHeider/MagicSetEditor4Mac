package app

// **Alles widerrufen nimmt genau das Eigene — auch wenn das Programm schon
// fehlt** — OPS03
// (#2349),
// Befund A-10 des Audits 0.9.0: wer das Programm in den Papierkorb legte,
// behielt einen LaunchAgent, Verknüpfungen in `~/.local/bin` und Einträge
// bei Codex und Claude, die auf ein fehlendes `mse-mcp` zeigten.
//
// # Die Lage
//
// Eingerichtet wird aus einem Bündel A, widerrufen aus einem Bündel B — A
// gibt es dann nicht mehr. Neben jedem eigenen Eintrag steht ein fremder,
// und der bleibt bytegleich. **Der Befehlsweg ist eingesetzt**: `launchctl`
// und `claude` dieses Rechners fasst ein Lauf nicht an (§3).

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// buendelIn legt ein Bündel mit den drei Programmen an und gibt seinen
// Ordner `Contents/MacOS`.
func buendelIn(t *testing.T, ort, name string) string {
	t.Helper()

	macos := filepath.Join(ort, name, "Contents", "MacOS")
	for _, programm := range []string{"mse", "mse-mcp", "mse-rest"} {
		legeAn(t, filepath.Join(macos, programm), "#!/bin/sh\n")
	}

	return macos
}

// nachbarn sind die fremden Einträge neben jedem eigenen — Pfad und Inhalt.
func nachbarn(heim string) map[string]string {
	stuetze := filepath.Join(heim, "Library", "Application Support")

	return map[string]string{
		filepath.Join(heim, ".zprofile"):                                   "export EDITOR=vi\n",
		filepath.Join(heim, ".codex", "config.toml"):                       "[mcp_servers.anderer]\ncommand = \"/opt/anderer\"\n",
		filepath.Join(stuetze, "Claude", "claude_desktop_config.json"):     `{"mcpServers":{"anderer":{"command":"/opt/anderer","args":[]}}}`,
		filepath.Join(heim, ".local", "bin", "mein-werkzeug"):              "#!/bin/sh\n",
		filepath.Join(heim, "Library", "LaunchAgents", "de.anderer.plist"): "<plist/>\n",
	}
}

// fremdBleibt prüft, dass die fremden Einträge noch dastehen — `genau`
// verlangt jede Datei bytegleich wie vor dem Einrichten, sonst nur die, in
// die nichts Eigenes kam.
func fremdBleibt(t *testing.T, heim, wann string, genau bool) {
	t.Helper()

	for pfad, erwartet := range nachbarn(heim) {
		gelesen := inhalt(t, pfad)

		switch base := filepath.Base(pfad); {
		case genau || base == "mein-werkzeug" || base == "de.anderer.plist":
			if gelesen != erwartet {
				t.Errorf("%s: %s ist nicht mehr bytegleich: %q", wann, pfad, gelesen)
			}
		case base == "claude_desktop_config.json":
			// **Claude Desktop schreibt seine JSON neu** — gesucht wird der
			// Eintrag, nicht der Wortlaut; bytegleich wird sie beim Widerrufen.
			if !strings.Contains(gelesen, `"anderer"`) || !strings.Contains(gelesen, "/opt/anderer") {
				t.Errorf("%s: der fremde Eintrag in %s fehlt: %q", wann, pfad, gelesen)
			}
		default:
			if !strings.Contains(gelesen, strings.TrimSuffix(erwartet, "\n")) {
				t.Errorf("%s: der fremde Eintrag in %s fehlt: %q", wann, pfad, gelesen)
			}
		}
	}
}

func ausgaengeVon(antwort SchnittstellenAlleAntwort) map[string]string {
	ausgaenge := map[string]string{}
	for _, eines := range antwort.Ausgaenge {
		ausgaenge[eines.Name] = eines.Ausgang
	}

	return ausgaenge
}

func TestAllesWiderrufenNimmtNurEigenesAuchOhneProgramm(t *testing.T) {
	t.Parallel()

	heim, programme := t.TempDir(), t.TempDir()
	a := buendelIn(t, programme, "MagicSetEditor4Mac.app")

	for _, ordner := range []string{".codex", ".claude"} {
		if err := os.MkdirAll(filepath.Join(heim, ordner), 0o700); err != nil {
			t.Fatal(err)
		}
	}

	for pfad, text := range nachbarn(heim) {
		legeAn(t, pfad, text)
	}

	buch := &befehlsbuch{}
	anwendungAus := func(werkzeug string) *App {
		return NeueApp("").MitSchnittstellenorten(
			func() (string, error) { return heim, nil },
			func() (string, error) { return werkzeug, nil },
			func(context.Context) string { return "/usr/bin:/bin" },
		).MitSchnittstellenbefehlen(buch.ausfuehren)
	}

	eingerichtet := anwendungAus(a).SchnittstellenAlleEinrichten(t.Context())
	if eingerichtet.Fehler != nil {
		t.Fatalf("einrichten: %+v", eingerichtet.Fehler)
	}

	for _, name := range []string{"befehlszeile", "codex", "claudedesktop", "claudecode", "dienst", befehlsziel} {
		if ausgang := ausgaengeVon(eingerichtet)[name]; ausgang != AusgangEingerichtet {
			t.Errorf("einrichten %s: %q", name, ausgang)
		}
	}

	// **Was `claude mcp add` schriebe** — der Befehlsweg ist eingesetzt.
	legeAn(t, filepath.Join(heim, ".claude.json"), `{"mcpServers":{"anderer":{"command":"/opt/anderer"},`+
		`"magicseteditor4mac":{"type":"stdio","command":"`+filepath.Join(a, "mse-mcp")+`","args":[],"env":{}}}}`)

	fremdBleibt(t, heim, "nach dem Einrichten", false)

	// **Das Programm ist fort** — widerrufen wird aus einem anderen Bündel.
	if err := os.RemoveAll(filepath.Dir(filepath.Dir(a))); err != nil {
		t.Fatal(err)
	}

	b := buendelIn(t, t.TempDir(), "MagicSetEditor4Mac 2.app")

	zurueck := anwendungAus(b).SchnittstellenAlleWiderrufen(t.Context())
	if zurueck.Fehler != nil {
		t.Fatalf("widerrufen: %+v", zurueck.Fehler)
	}

	for _, name := range []string{"befehlszeile", "codex", "claudedesktop", "claudecode", "dienst", befehlsziel} {
		if ausgang := ausgaengeVon(zurueck)[name]; ausgang != AusgangWiderrufen {
			t.Errorf("widerrufen %s: %q", name, ausgang)
		}
	}

	for _, eigenes := range []string{
		filepath.Join(heim, ".local", "bin", "mse"), filepath.Join(heim, ".local", "bin", "mse-mcp"),
		filepath.Join(heim, "Library", "LaunchAgents", dienstlabel+".plist"),
	} {
		if _, err := os.Lstat(eigenes); !os.IsNotExist(err) {
			t.Errorf("%s steht noch (%v)", eigenes, err)
		}
	}

	for _, datei := range []string{".zprofile", filepath.Join(".codex", "config.toml")} {
		if gelesen := inhalt(t, filepath.Join(heim, datei)); strings.Contains(gelesen, "magicseteditor4mac") ||
			strings.Contains(gelesen, ".local/bin") {
			t.Errorf("in %s steht noch Eigenes: %q", datei, gelesen)
		}
	}

	fremdBleibt(t, heim, "nach dem Widerrufen", true)

	if befehle := strings.Join(buch.befehle, "\n"); !strings.Contains(befehle, "mcp remove -s user magicseteditor4mac") ||
		!strings.Contains(befehle, "launchctl bootout") {
		t.Errorf("die Befehle waren %q", befehle)
	}

	// **Ein zweites Mal ist nichts mehr zu tun** — und das ist kein Fehler.
	legeAn(t, filepath.Join(heim, ".claude.json"), `{"mcpServers":{"anderer":{"command":"/opt/anderer"}}}`)

	nochmal := anwendungAus(b).SchnittstellenAlleWiderrufen(t.Context())
	if nochmal.Fehler != nil {
		t.Fatalf("ein zweites Mal: %+v", nochmal.Fehler)
	}

	for name, ausgang := range ausgaengeVon(nochmal) {
		if ausgang != AusgangUnveraendert {
			t.Errorf("ein zweites Mal %s: %q", name, ausgang)
		}
	}

	fremdBleibt(t, heim, "nach dem zweiten Widerrufen", true)
}

// **Gemischt heißt veraltet** — sonst bliebe beim Widerrufen eine eigene
// Verknüpfung neben einer fremden liegen.
func TestDerStandDerVerweise(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		staende  []string
		erwartet service.Schnittstellenstand
	}{
		"beide eigen":        {[]string{befehlEigen, befehlEigen}, service.SchnittstelleEingerichtet},
		"beide frei":         {[]string{befehlFrei, befehlFrei}, service.SchnittstelleOffen},
		"beide fremd":        {[]string{befehlFremd, befehlFremd}, service.SchnittstelleFremd},
		"fremd und frei":     {[]string{befehlFremd, befehlFrei}, service.SchnittstelleOffen},
		"eigen und frei":     {[]string{befehlEigen, befehlFrei}, service.SchnittstelleVeraltet},
		"eigen und fremd":    {[]string{befehlEigen, befehlFremd}, service.SchnittstelleVeraltet},
		"veraltet und eigen": {[]string{befehlVeraltet, befehlEigen}, service.SchnittstelleVeraltet},
	} {
		var vorschau Befehlsvorschau
		for _, stand := range fall.staende {
			vorschau.Eintraege = append(vorschau.Eintraege, Befehlseintrag{Stand: stand})
		}

		if stand := befehlszielstand(vorschau); stand != fall.erwartet {
			t.Errorf("%s: %s, erwartet %s", name, stand, fall.erwartet)
		}
	}
}
