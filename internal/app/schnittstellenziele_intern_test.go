package app

// **Claude Desktop, Claude Code und der Dienst an der Anwendung** — OPS-09b
// (#1961).
// Gerufen wird weder `claude` noch `launchctl`.

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

const probewerkzeug = "/Programme/MagicSetEditor4Mac.app/Contents/MacOS"

// beendetMit ist ein Befehl, der mit einem Rückgabewert endete.
type beendetMit int

func (b beendetMit) Error() string { return "beendet" }
func (b beendetMit) ExitCode() int { return int(b) }

// eineZweiteEinrichtende baut eine Anwendung mit Zuhause, einem `claude` im
// PATH und einer Ausführung, die mitschreibt und `claude` nachstellt.
func eineZweiteEinrichtende(t *testing.T) (anwendung *App, heim string, gerufen *[][]string) {
	t.Helper()

	heim = t.TempDir()
	if fehler := os.MkdirAll(filepath.Join(heim, "Library", "Application Support", "Claude"), 0o700); fehler != nil {
		t.Fatal(fehler)
	}

	pfad := t.TempDir()
	if fehler := os.WriteFile(filepath.Join(pfad, "claude"), []byte("#!/bin/sh\n"), 0o700); fehler != nil { //nolint:gosec // ein Programm des Laufs
		t.Fatal(fehler)
	}

	gerufen = &[][]string{}
	claudejson := filepath.Join(heim, ".claude.json")

	anwendung = NeueApp("").MitSchnittstellenorten(
		func() (string, error) { return heim, nil },
		func() (string, error) { return probewerkzeug, nil },
		func(context.Context) string { return pfad },
	).MitSchnittstellenbefehlen(func(_ context.Context, befehl []string) error {
		*gerufen = append(*gerufen, befehl)

		switch {
		case strings.HasSuffix(befehl[0], "claude") && befehl[2] == "add":
			return os.WriteFile(claudejson, []byte(`{"mcpServers": {"`+mcpName+`": {"command": "`+befehl[6]+`"}}}`), 0o600)
		case strings.HasSuffix(befehl[0], "claude"):
			return os.WriteFile(claudejson, []byte(`{"mcpServers": {}}`), 0o600)
		case befehl[1] == "bootout":
			return beendetMit(nichtGeladen)
		}

		return nil
	})

	return anwendung, heim, gerufen
}

// **Alle drei Ziele lassen sich einrichten und widerrufen** — und die
// Antwort nennt sie mit ihrer Art.
func TestDieZweitenZieleKommenUndGehen(t *testing.T) {
	t.Parallel()

	anwendung, heim, gerufen := eineZweiteEinrichtende(t)
	ctx := context.Background()

	for _, name := range []string{"claudedesktop", "claudecode", "dienst"} {
		if stand := standVon(anwendung.Schnittstellen(ctx), name); stand.Stand != "offen" || stand.Art == "" {
			t.Fatalf("%s vorher: %+v", name, stand)
		}

		if antwort := anwendung.SchnittstelleEinrichten(ctx, name); antwort.Fehler != nil ||
			standVon(antwort, name).Stand != "eingerichtet" {
			t.Fatalf("%s einrichten: %+v", name, antwort)
		}

		if antwort := anwendung.SchnittstelleWiderrufen(ctx, name); antwort.Fehler != nil ||
			standVon(antwort, name).Stand != "offen" || antwort.NurEintrag != "" {
			t.Fatalf("%s widerrufen: %+v", name, antwort)
		}
	}

	claude := (*gerufen)[0]
	if !strings.HasSuffix(claude[0], "/claude") || strings.Join(claude[1:], " ") !=
		"mcp add -s user "+mcpName+" "+filepath.Join(probewerkzeug, "mse-mcp") {
		t.Errorf("claude: %v", claude)
	}

	if laden := (*gerufen)[2]; laden[0] != "launchctl" || laden[1] != "bootstrap" ||
		laden[3] != filepath.Join(heim, "Library", "LaunchAgents", dienstlabel+".plist") {
		t.Errorf("launchctl: %v", laden)
	}
}

// **Hat der Mensch Claude Desktops Datei seither geändert, sagt die Antwort,
// dass nur der eigene Eintrag ging.**
func TestNurDerEigeneEintragGingWirdGesagt(t *testing.T) {
	t.Parallel()

	anwendung, heim, _ := eineZweiteEinrichtende(t)
	ctx := context.Background()
	datei := filepath.Join(heim, "Library", "Application Support", "Claude", "claude_desktop_config.json")

	anwendung.SchnittstelleEinrichten(ctx, "claudedesktop")

	inhalt, fehler := os.ReadFile(datei) //nolint:gosec // eine Datei des Laufs
	if fehler != nil {
		t.Fatal(fehler)
	}

	if fehler := os.WriteFile(datei, append([]byte(" "), inhalt...), 0o600); fehler != nil { //nolint:gosec // eine Datei des Laufs
		t.Fatal(fehler)
	}

	if antwort := anwendung.SchnittstelleWiderrufen(ctx, "claudedesktop"); antwort.NurEintrag != "claudedesktop" {
		t.Errorf("%+v", antwort)
	}
}

// **Die Vorschau zeigt je Art, was geschieht** (AK-01).
func TestDieVorschauZeigtJeArtWasGeschieht(t *testing.T) {
	t.Parallel()

	anwendung, _, _ := eineZweiteEinrichtende(t)
	antwort := anwendung.Schnittstellen(context.Background())

	for name, erwartet := range map[string]string{
		"claudedesktop": `"` + mcpName + `": {"command": "` + filepath.Join(probewerkzeug, "mse-mcp") + `"`,
		"claudecode":    "$ ",
		"dienst":        "<string>" + filepath.Join(probewerkzeug, "mse-rest") + "</string>",
		"codex":         domain.BlockAnfang,
	} {
		if zeilen := strings.Join(standVon(antwort, name).Zeilen, "\n"); !strings.Contains(zeilen, erwartet) {
			t.Errorf("%s: %q fehlt in\n%s", name, erwartet, zeilen)
		}
	}
}

// **Ohne Claude ist Claude Code nicht gefunden** — und die Vorschau nennt
// trotzdem den Befehl.
func TestOhneClaudeIstClaudeCodeNichtGefunden(t *testing.T) {
	t.Parallel()

	heim := t.TempDir()
	anwendung := NeueApp("").MitSchnittstellenorten(
		func() (string, error) { return heim, nil },
		func() (string, error) { return probewerkzeug, nil },
		func(context.Context) string { return "/gibtsnicht" },
	)

	code := standVon(anwendung.Schnittstellen(context.Background()), "claudecode")
	if code.Stand != "nichtGefunden" || !strings.Contains(strings.Join(code.Zeilen, " "), "$ claude mcp add") {
		t.Errorf("%+v", code)
	}

	if desktop := standVon(anwendung.Schnittstellen(context.Background()), "claudedesktop"); desktop.Stand != "nichtGefunden" {
		t.Errorf("%+v", desktop)
	}
}

// **Die echte Ausführung**: ein geglückter Befehl ist still, ein
// gescheiterter sagt, welches Programm — ohne Pfad im Satz (§8). **Gestartet
// wird nur, was die Anwendung kennt**: `claude` aus dem Ordner, in dem die
// Anmeldeshell ihn fand, und `launchctl` aus `/bin`.
func TestDieEchteAusfuehrung(t *testing.T) {
	t.Parallel()

	ordner := t.TempDir()
	claude := filepath.Join(ordner, "claude")

	probe := func(skript string) {
		t.Helper()

		if fehler := os.WriteFile(claude, []byte("#!/bin/sh\n"+skript+"\n"), 0o700); fehler != nil { //nolint:gosec // ein Programm des Laufs
			t.Fatal(fehler)
		}
	}

	ctx := context.Background()

	probe("exit 0")

	if fehler := befehlAusfuehren(ctx, []string{claude, "mcp"}); fehler != nil {
		t.Errorf("geglückt: %v", fehler)
	}

	probe("echo kaputt >&2; exit 1")

	if fehler := befehlAusfuehren(ctx, []string{claude, "mcp"}); !errors.Is(fehler, domain.ErrConflict) ||
		strings.Contains(fehler.Error(), "/") {
		t.Errorf("gescheitert: %v", fehler)
	}

	// **`launchctl version` liest nur** — und belegt, dass `/bin` gemeint ist.
	if fehler := befehlAusfuehren(ctx, []string{"launchctl", "version"}); fehler != nil {
		t.Errorf("launchctl: %v", fehler)
	}

	// **Und was den Dienst dieses Rechners verändert, ruft ein Testlauf nie**
	// — Zweitblick über Phase 2, Befund 6: das Ziel ist `gui/<UID>/<Etikett>`,
	// und ein umgelenktes Zuhause lenkt das nicht um.
	for _, unterbefehl := range []string{"kickstart", "bootstrap", "bootout", "kill"} {
		if fehler := befehlAusfuehren(ctx, []string{"launchctl", unterbefehl, "gui/0/" + dienstlabel}); !errors.Is(fehler, domain.ErrValidation) {
			t.Errorf("launchctl %s im Testlauf: %v", unterbefehl, fehler)
		}
	}

	for name, befehl := range map[string][]string{
		"ein fremdes Programm": {"/bin/sh", "-c", "exit 0"},
		"claude ohne Ort":      {"claude", "mcp"},
	} {
		if fehler := befehlAusfuehren(ctx, befehl); !errors.Is(fehler, domain.ErrValidation) {
			t.Errorf("%s: %v", name, fehler)
		}
	}
}

// **Entladen duldet „nicht geladen“ und sonst nichts.**
func TestEntladenDuldetNurNichtGeladen(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		befehl []string
		fehler error
		bleibt bool
	}{
		"nicht geladen":  {[]string{"launchctl", "bootout", "x"}, beendetMit(nichtGeladen), false},
		"anderer Fehler": {[]string{"launchctl", "bootout", "x"}, beendetMit(5), true},
		"beim Laden":     {[]string{"launchctl", "bootstrap", "x"}, beendetMit(nichtGeladen), true},
		"gelungen":       {[]string{"launchctl", "bootout", "x"}, nil, false},
	} {
		geduldet := duldetNichtGeladen(func(context.Context, []string) error { return fall.fehler })
		if fehler := geduldet(context.Background(), fall.befehl); (fehler != nil) != fall.bleibt {
			t.Errorf("%s: %v", name, fehler)
		}
	}
}

// **Ein Pfad mit `&` macht die plist nicht kaputt.**
func TestDiePlistMaskiertPfade(t *testing.T) {
	t.Parallel()

	if plist := dienstplist("/A & B/mse-rest", "/x<y>/ansage", "/x<y>/log"); !strings.Contains(plist, "/A &amp; B/mse-rest") ||
		!strings.Contains(plist, "/x&lt;y&gt;/ansage") || !strings.Contains(plist, "/x&lt;y&gt;/log") {
		t.Errorf("%s", plist)
	}
}

// **launchd hängt die Ansage nicht mehr an** (SEC02, #2355): das Programm
// bekommt die Ansagedatei als Fahne und ersetzt sie je Start selbst; die
// Standardausgabe geht ins Leere, die Fehlerausgabe in eine eigene Datei.
// Zeigte `StandardOutPath` weiter auf die Ansage, schriebe launchd jede
// Marke dazu — und zehn Neustarts ließen zehn darin.
func TestDiePlistReichtDieAnsageAlsFahne(t *testing.T) {
	t.Parallel()

	plist := dienstplist("/p/mse-rest", "/e/dienst.txt", "/e/dienst.log")

	for _, erwartet := range []string{
		"<string>/p/mse-rest</string>\n\t\t<string>-ansage</string>\n\t\t<string>/e/dienst.txt</string>\n\t</array>",
		"<key>StandardOutPath</key>\n\t<string>/dev/null</string>",
		"<key>StandardErrorPath</key>\n\t<string>/e/dienst.log</string>",
	} {
		if !strings.Contains(plist, erwartet) {
			t.Errorf("es fehlt %q in\n%s", erwartet, plist)
		}
	}

	if strings.Count(plist, "/e/dienst.txt") != 1 {
		t.Errorf("die Ansage steht nicht genau einmal — als Fahne — in\n%s", plist)
	}
}

// **Ein gebautes Kind eines Laufs ruft launchctl nie verändernd** — Nachlauf
// über Phase 2 von 0.9.0, Befund 6: `mse`, `mse-mcp` und `mse-rest` heißen in
// den Flächenläufen nicht `*.test`, und ein umgelenktes Zuhause lenkt
// `gui/<UID>/<Etikett>` nicht um. **Nicht parallel**: der Lauf setzt
// [testlauf] und die Umgebung.
func TestDieDienstsperreGiltOhneTestnamen(t *testing.T) {
	vorher := testlauf
	testlauf = false

	t.Cleanup(func() { testlauf = vorher })

	t.Setenv(Dienstsperre, "an")

	if fehler := befehlAusfuehren(t.Context(), []string{"launchctl", "kickstart", "-k", "gui/0/" + dienstlabel}); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("mit der Sperre: %v", fehler)
	}

	// **Die Gegenprobe**: ohne Sperre geht der Ruf hinaus — an ein Ziel, das
	// es nicht gibt, und `launchctl` weist ihn ab.
	t.Setenv(Dienstsperre, "")

	if fehler := befehlAusfuehren(t.Context(), []string{"launchctl", "kickstart", "-k", "gui/0/" + dienstlabel}); !errors.Is(fehler, domain.ErrConflict) {
		t.Errorf("ohne die Sperre: %v", fehler)
	}
}

// **Und die Vorgaben dieses Rechners auch nicht** (R02, #2334) — dieselbe
// Sperre für `defaults`. Die Gegenprobe nennt eine Domäne als **Pfad im
// Testordner**: `defaults` liest dort und nirgends sonst.
func TestDieDienstsperreGiltAuchFuerDieVorgaben(t *testing.T) {
	vorher := testlauf
	testlauf = false

	t.Cleanup(func() { testlauf = vorher })

	ordner := t.TempDir()
	befehl := []string{"defaults", "export", filepath.Join(ordner, "gibtesnicht"), filepath.Join(ordner, "aus.plist")}

	t.Setenv(Dienstsperre, "an")

	if fehler := befehlAusfuehren(t.Context(), befehl); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("mit der Sperre: %v", fehler)
	}

	t.Setenv(Dienstsperre, "")

	// **Gerufen ist gerufen** — ob `defaults` eine leere Domäne ausgibt oder
	// abweist, ist seine Sache; die Sperre darf es nicht sein.
	if fehler := befehlAusfuehren(t.Context(), befehl); errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("ohne die Sperre: %v", fehler)
	}
}
