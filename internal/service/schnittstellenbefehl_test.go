package service_test

// **Claude Code und der LaunchAgent: was der Klient selbst schreibt** —
// OPS-09b (#1961).
//
// Gerufen wird hier weder `claude` noch `launchctl`: ein nachgestellter
// Klient schreibt, was der echte schriebe.

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// nachgestellterClaude schreibt `mcpServers` wie `claude mcp add/remove`.
type nachgestellterClaude struct {
	datei   string
	gerufen [][]string
	fehler  error
}

func (n *nachgestellterClaude) ausfuehren(_ context.Context, befehl []string) error {
	n.gerufen = append(n.gerufen, befehl)
	if n.fehler != nil {
		return n.fehler
	}

	switch befehl[2] {
	case "add":
		return os.WriteFile(n.datei, []byte(`{"numStartups": 3, "mcpServers": {"`+befehl[5]+
			`": {"type": "stdio", "command": "`+befehl[6]+`", "args": [], "env": {}}}}`), 0o600)
	case "remove":
		return os.WriteFile(n.datei, []byte(`{"numStartups": 3, "mcpServers": {}}`), 0o600)
	}

	return nil
}

func einCodeziel(t *testing.T) (service.Schnittstellenziel, *nachgestellterClaude) {
	t.Helper()

	klient := &nachgestellterClaude{datei: filepath.Join(t.TempDir(), ".claude.json")}

	return service.Schnittstellenziel{
		Name: "claudecode", Art: service.ArtBefehl, Datei: klient.datei, Vorhanden: true,
		Schluessel: "magicseteditor4mac", Programm: desktopprogramm,
		Einrichtbefehle: [][]string{{"claude", "mcp", "add", "-s", "user", "magicseteditor4mac", desktopprogramm}},
		Widerrufbefehle: [][]string{{"claude", "mcp", "remove", "-s", "user", "magicseteditor4mac"}},
		Ausfuehren:      klient.ausfuehren,
	}, klient
}

// **Einrichten und Widerrufen über den Befehl des Klienten** — und der
// Stand aus seiner Datei, ohne ihn zu starten.
func TestClaudeCodeUeberSeinenBefehl(t *testing.T) {
	t.Parallel()

	ziel, klient := einCodeziel(t)

	if stand := service.SchnittstelleLesen(ziel); stand != service.SchnittstelleOffen {
		t.Fatalf("vorher: %q", stand)
	}

	if aus := ausgang(t)(service.SchnittstelleEinrichten(t.Context(), ziel)); aus.Stand != service.SchnittstelleEingerichtet {
		t.Fatalf("Einrichten: %+v", aus)
	}

	if aus := ausgang(t)(service.SchnittstelleWiderrufen(t.Context(), ziel)); aus.Stand != service.SchnittstelleOffen || !aus.Bytegleich {
		t.Fatalf("Widerrufen: %+v", aus)
	}

	if len(klient.gerufen) != 2 || klient.gerufen[0][2] != "add" || klient.gerufen[1][2] != "remove" {
		t.Errorf("gerufen: %v", klient.gerufen)
	}
}

// **Ein veralteter Eintrag geht zuerst** — `claude mcp add` weist einen
// Namen ab, den es schon gibt.
func TestClaudeCodeRepariertEinenVeraltetenEintrag(t *testing.T) {
	t.Parallel()

	ziel, klient := einCodeziel(t)
	schreibeDatei(t, klient.datei, `{"mcpServers": {"magicseteditor4mac": {"command": "/Alt/MagicSetEditor4Mac.app/Contents/MacOS/mse-mcp"}}}`)

	if aus := ausgang(t)(service.SchnittstelleEinrichten(t.Context(), ziel)); aus.Stand != service.SchnittstelleEingerichtet {
		t.Fatalf("%+v", aus)
	}

	if len(klient.gerufen) != 2 || klient.gerufen[0][2] != "remove" || klient.gerufen[1][2] != "add" {
		t.Errorf("gerufen: %v", klient.gerufen)
	}
}

// **Ein Fehler des Klienten kommt an**, und der Stand bleibt, wie er war.
func TestClaudeCodeSagtSeinenFehler(t *testing.T) {
	t.Parallel()

	ziel, klient := einCodeziel(t)
	klient.fehler = errors.New("claude ist nicht angemeldet")

	aus, fehler := service.SchnittstelleEinrichten(t.Context(), ziel)
	if !errors.Is(fehler, klient.fehler) || aus.Stand != service.SchnittstelleOffen {
		t.Errorf("%+v, %v", aus, fehler)
	}
}

// **Eine große Datei des Klienten wird gelesen** — sie gehört ihm, und hier
// wird sie nur gelesen.
func TestClaudeCodeLiestEineGrosseDatei(t *testing.T) {
	t.Parallel()

	ziel, klient := einCodeziel(t)
	gross := `{"projects": "` + strings.Repeat("x", 2<<20) + `", "mcpServers": {}}`
	schreibeDatei(t, klient.datei, gross)

	if stand := service.SchnittstelleLesen(ziel); stand != service.SchnittstelleOffen {
		t.Errorf("Stand %q", stand)
	}
}

func einAgentenziel(t *testing.T) (ziel service.Schnittstellenziel, gerufen *[][]string) {
	t.Helper()

	ordner := t.TempDir()
	gerufen = &[][]string{}

	return service.Schnittstellenziel{
		Name: "dienst", Art: service.ArtAgent, Vorhanden: true,
		Datei:           filepath.Join(ordner, "LaunchAgents", "com.probe.dienst.plist"),
		Schluessel:      "com.probe.dienst",
		Block:           "<plist><string>com.probe.dienst</string><string>/x/mse-rest</string></plist>\n",
		Begleitordner:   filepath.Join(ordner, "eigen"),
		Ansage:          filepath.Join(ordner, "eigen", "dienst.txt"),
		Einrichtbefehle: [][]string{{"launchctl", "bootstrap", "gui/501", "plist"}},
		Widerrufbefehle: [][]string{{"launchctl", "bootout", "gui/501/com.probe.dienst"}},
		Ausfuehren: func(_ context.Context, befehl []string) error {
			*gerufen = append(*gerufen, befehl)

			return nil
		},
	}, gerufen
}

// **Der LaunchAgent: plist schreiben, laden, entladen, löschen.**
func TestDerLaunchAgentKommtUndGeht(t *testing.T) {
	t.Parallel()

	ziel, gerufen := einAgentenziel(t)

	if aus := ausgang(t)(service.SchnittstelleEinrichten(t.Context(), ziel)); aus.Stand != service.SchnittstelleEingerichtet {
		t.Fatalf("Einrichten: %+v", aus)
	}

	if angaben, fehler := os.Stat(ziel.Begleitordner); fehler != nil || angaben.Mode().Perm() != 0o700 {
		t.Errorf("der Begleitordner: %v", fehler)
	}

	if liesDatei(t, ziel.Datei) != ziel.Block {
		t.Error("die plist ist nicht die eigene")
	}

	// **Die Ansage, wie der laufende Dienst sie schreibt** (SEC02, #2355).
	schreibeDatei(t, ziel.Ansage, "Marke abc\n")

	if aus := ausgang(t)(service.SchnittstelleWiderrufen(t.Context(), ziel)); aus.Stand != service.SchnittstelleOffen {
		t.Fatalf("Widerrufen: %+v", aus)
	}

	if _, fehler := os.Stat(ziel.Datei); !errors.Is(fehler, os.ErrNotExist) {
		t.Errorf("die plist steht noch: %v", fehler)
	}

	// **Und die Marke geht mit dem Dienst**: sie gilt nicht mehr, und eine
	// Datei, die eine tote Marke nennt, sähe aus wie ein Zugang.
	if _, fehler := os.Stat(ziel.Ansage); !errors.Is(fehler, os.ErrNotExist) {
		t.Errorf("die Ansage steht noch: %v", fehler)
	}

	if len(*gerufen) != 2 || (*gerufen)[0][1] != "bootstrap" || (*gerufen)[1][1] != "bootout" {
		t.Errorf("gerufen: %v", *gerufen)
	}
}

// **Eine plist, die sich nicht laden ließ, geht wieder** — kein halber Stand.
func TestEinNichtGeladenerAgentHinterlaesstNichts(t *testing.T) {
	t.Parallel()

	ziel, _ := einAgentenziel(t)
	kaputt := errors.New("Bootstrap failed: 5")

	// **Das Entladen danach gelingt** — so, wie die App „nicht geladen“
	// meldet (Exit 3, gemessen am 2026-10-01). Ein Entladen, das klemmt,
	// steht in `TestEinGescheitertesLadenNimmtDiePlistNurNachDemEntladen`.
	ziel.Ausfuehren = func(_ context.Context, befehl []string) error {
		if befehl[1] == "bootstrap" {
			return kaputt
		}

		return nil
	}

	if _, fehler := service.SchnittstelleEinrichten(t.Context(), ziel); !errors.Is(fehler, kaputt) {
		t.Fatalf("%v", fehler)
	}

	if _, fehler := os.Stat(ziel.Datei); !errors.Is(fehler, os.ErrNotExist) {
		t.Errorf("die plist blieb: %v", fehler)
	}

	// **Und ein Widerruf, der nicht entladen kann, lässt die plist stehen**
	// — sonst liefe ein Dienst ohne Datei, die ihn erklärt.
	ziel.Ausfuehren = func(context.Context, []string) error { return kaputt }
	schreibeDatei(t, ziel.Datei, ziel.Block)
	schreibeDatei(t, ziel.Ansage, "Marke abc\n")

	if _, fehler := service.SchnittstelleWiderrufen(t.Context(), ziel); !errors.Is(fehler, kaputt) {
		t.Fatalf("%v", fehler)
	}

	if liesDatei(t, ziel.Datei) != ziel.Block {
		t.Error("die plist ging trotzdem")
	}

	// **Die Ansage auch**: der Dienst läuft weiter, und seine Marke gilt.
	if liesDatei(t, ziel.Ansage) != "Marke abc\n" {
		t.Error("die Ansage ging, obwohl der Dienst noch läuft")
	}
}

// **Eine Ansage, die sich nicht entfernen lässt, wird genannt** — ein
// Widerruf, der still eine Marke liegen ließe, hielte nicht, was er sagt.
func TestEineAnsageDieBleibtWirdGenannt(t *testing.T) {
	t.Parallel()

	ziel, _ := einAgentenziel(t)

	if _, fehler := service.SchnittstelleEinrichten(t.Context(), ziel); fehler != nil {
		t.Fatalf("Einrichten: %v", fehler)
	}

	// **Ein Ordner mit Inhalt an ihrer Stelle**: `os.Remove` nimmt ihn nicht.
	if fehler := os.MkdirAll(filepath.Join(ziel.Ansage, "darin"), 0o700); fehler != nil {
		t.Fatal(fehler)
	}

	if _, fehler := service.SchnittstelleWiderrufen(t.Context(), ziel); fehler == nil ||
		!strings.Contains(fehler.Error(), "Ansage") {
		t.Errorf("der Widerruf schwieg: %v", fehler)
	}
}

// **Die Stände des Agenten**: veraltet ist eine plist mit dem eigenen
// Etikett und anderem Inhalt, fremd eine ohne.
func TestDieStaendeDesAgenten(t *testing.T) {
	t.Parallel()

	for inhalt, erwartet := range map[string]service.Schnittstellenstand{
		"<plist><string>com.probe.dienst</string><string>/alt/mse-rest</string></plist>\n": service.SchnittstelleVeraltet,
		"<plist><string>jemand.anderes</string></plist>\n":                                 service.SchnittstelleFremd,
	} {
		ziel, gerufen := einAgentenziel(t)
		if fehler := os.MkdirAll(filepath.Dir(ziel.Datei), 0o700); fehler != nil {
			t.Fatal(fehler)
		}

		schreibeDatei(t, ziel.Datei, inhalt)

		if stand := service.SchnittstelleLesen(ziel); stand != erwartet {
			t.Errorf("%q: %q", inhalt, stand)
		}

		_, fehler := service.SchnittstelleEinrichten(t.Context(), ziel)

		switch erwartet {
		case service.SchnittstelleVeraltet:
			// **Erst entladen, dann neu laden** — mit der eigenen plist.
			if fehler != nil || liesDatei(t, ziel.Datei) != ziel.Block ||
				!slices.EqualFunc(*gerufen, [][]string{ziel.Widerrufbefehle[0], ziel.Einrichtbefehle[0]}, slices.Equal) {
				t.Errorf("veraltet: %v %v", fehler, *gerufen)
			}
		default:
			if !errors.Is(fehler, domain.ErrConflict) || len(*gerufen) != 0 {
				t.Errorf("fremd: %v %v", fehler, *gerufen)
			}
		}
	}
}

// **Ein Begleitordner, der nicht entstehen kann, hält vor allem anderen.**
func TestOhneBegleitordnerKeinAgent(t *testing.T) {
	t.Parallel()

	ziel, gerufen := einAgentenziel(t)
	sperre := filepath.Join(t.TempDir(), "datei")
	schreibeDatei(t, sperre, "x")
	ziel.Begleitordner = filepath.Join(sperre, "darunter")

	if _, fehler := service.SchnittstelleEinrichten(t.Context(), ziel); fehler == nil || len(*gerufen) != 0 {
		t.Errorf("%v %v", fehler, *gerufen)
	}
}
