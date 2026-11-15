package service_test

// **Was der Zweitblick über Phase 6.12 fand** — Z1, Z5 und Z6
// (#1905).

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

// **Z1 — Reparieren nach fremder Änderung macht das alte Gedächtnis nicht
// wieder gültig.** Sonst legte Widerrufen den Stand von vor dem ersten
// Einrichten zurück, und was der Mensch seither schrieb, wäre fort.
func TestReparierenNachFremderAenderungBehaeltSie(t *testing.T) {
	t.Parallel()

	ziel, datei := einDesktopziel(t)
	if fehler := os.MkdirAll(filepath.Dir(datei), 0o700); fehler != nil {
		t.Fatal(fehler)
	}

	schreibeDatei(t, datei, "{\n  \"globalShortcut\": \"\"\n}\n")
	ausgang(t)(service.SchnittstelleEinrichten(t.Context(), ziel))

	schreibeDatei(t, datei, strings.Replace(liesDatei(t, datei), "{", "{\n  \"neu\": true,", 1))

	umgezogen := ziel
	umgezogen.Programm = "/Neu/MagicSetEditor4Mac.app/Contents/MacOS/mse-mcp"

	if stand := service.SchnittstelleLesen(umgezogen); stand != service.SchnittstelleVeraltet {
		t.Fatalf("nach dem Umzug: %q", stand)
	}

	ausgang(t)(service.SchnittstelleEinrichten(t.Context(), umgezogen))

	aus := ausgang(t)(service.SchnittstelleWiderrufen(t.Context(), umgezogen))
	nachher := liesDatei(t, datei)

	if aus.Bytegleich || !strings.Contains(nachher, `"neu": true`) || strings.Contains(nachher, "magicseteditor4mac") {
		t.Errorf("%+v\n%s", aus, nachher)
	}
}

// **Z5 — derselbe Programmname ist kein Nachweis.** Ein Eintrag mit
// eigenen Argumenten oder eigener Umgebung gehört dem, der ihn so schrieb;
// `claude mcp add` schreibt selbst `type`, leere `args` und leeres `env`,
// und die bleiben die eigenen.
func TestDerselbeProgrammnameIstKeinNachweis(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		name, eintrag string
		stand         service.Schnittstellenstand
	}{
		{"anderes Programm, eigene Argumente", `{"command": "/anderes/mse-mcp", "args": ["--profil", "x"]}`, service.SchnittstelleFremd},
		{"eigenes Programm mit Umgebung", `{"command": "` + desktopprogramm + `", "args": [], "env": {"A": "1"}}`, service.SchnittstelleFremd},
		{"eigenes Programm mit Argumenten", `{"command": "` + desktopprogramm + `", "args": ["--x"]}`, service.SchnittstelleFremd},
		{"wie claude mcp add es schreibt", `{"type": "stdio", "command": "` + desktopprogramm + `", "args": [], "env": {}}`, service.SchnittstelleEingerichtet},
		{"umgezogen, wie claude mcp add es schreibt", `{"type": "stdio", "command": "/Alt/MagicSetEditor4Mac.app/Contents/MacOS/mse-mcp", "args": [], "env": {}}`, service.SchnittstelleVeraltet},
		{"umgezogen, ohne args", `{"command": "/Alt/MagicSetEditor4Mac.app/Contents/MacOS/mse-mcp"}`, service.SchnittstelleVeraltet},
		// **N3 des Nachlaufs: die Gestalt allein ist kein Nachweis** — ein
		// `mse-mcp` außerhalb eines Bündels hat jemand anders hingelegt.
		{"fremder Pfad, eigene Gestalt", `{"command": "/anderes/mse-mcp", "args": []}`, service.SchnittstelleFremd},
		{"ein anderer Typ", `{"type": "sse", "command": "` + desktopprogramm + `"}`, service.SchnittstelleFremd},
		{"ohne Befehl", `{"args": []}`, service.SchnittstelleFremd},
		{"Argumente als Text", `{"command": "` + desktopprogramm + `", "args": "--x"}`, service.SchnittstelleFremd},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			ziel, datei := einDesktopziel(t)
			if fehler := os.MkdirAll(filepath.Dir(datei), 0o700); fehler != nil {
				t.Fatal(fehler)
			}

			schreibeDatei(t, datei, `{"mcpServers": {"magicseteditor4mac": `+fall.eintrag+`}}`)

			if stand := service.SchnittstelleLesen(ziel); stand != fall.stand {
				t.Errorf("Stand %q, erwartet %q", stand, fall.stand)
			}
		})
	}
}

// **Z6 — ein Entladen, das scheitert, hält das Einrichten an.** Sonst
// scheiterte das Laden am noch geladenen Dienst, die plist ginge, und ein
// Dienst liefe ohne Datei, die ihn erklärt.
func TestEinGescheitertesEntladenHaeltDasReparierenAn(t *testing.T) {
	t.Parallel()

	ziel, gerufen := einAgentenziel(t)
	if fehler := os.MkdirAll(filepath.Dir(ziel.Datei), 0o700); fehler != nil {
		t.Fatal(fehler)
	}

	alt := "<plist><string>com.probe.dienst</string><string>/alt/mse-rest</string></plist>\n"
	schreibeDatei(t, ziel.Datei, alt)

	klemmt := errors.New("Boot-out failed: 5: Input/output error")
	ziel.Ausfuehren = func(_ context.Context, befehl []string) error {
		*gerufen = append(*gerufen, befehl)
		if befehl[1] == "bootout" {
			return klemmt
		}

		return nil
	}

	if _, fehler := service.SchnittstelleEinrichten(t.Context(), ziel); !errors.Is(fehler, klemmt) {
		t.Fatalf("das Entladen scheiterte, und es kam %v", fehler)
	}

	if liesDatei(t, ziel.Datei) != alt || slices.ContainsFunc(*gerufen, func(b []string) bool { return b[1] == "bootstrap" }) {
		t.Errorf("trotzdem geladen oder geschrieben: %v", *gerufen)
	}
}

// **Z6 — und ein Laden, das scheitert, nimmt die plist nur, wenn der Dienst
// sich entladen ließ.** Lässt er sich nicht entladen, bleibt sie — und der
// Fehler sagt beides.
func TestEinGescheitertesLadenNimmtDiePlistNurNachDemEntladen(t *testing.T) {
	t.Parallel()

	laedtNicht := errors.New("Bootstrap failed: 5: Input/output error")
	entlaedtNicht := errors.New("Boot-out failed: 5: Input/output error")

	for name, entladen := range map[string]error{"entladen geht": nil, "entladen klemmt": entlaedtNicht} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ziel, gerufen := einAgentenziel(t)
			ziel.Ausfuehren = func(_ context.Context, befehl []string) error {
				*gerufen = append(*gerufen, befehl)
				if befehl[1] == "bootstrap" {
					return laedtNicht
				}

				return entladen
			}

			_, fehler := service.SchnittstelleEinrichten(t.Context(), ziel)
			if !errors.Is(fehler, laedtNicht) {
				t.Fatalf("das Laden scheiterte, und es kam %v", fehler)
			}

			_, offen := os.Stat(ziel.Datei)

			switch {
			case entladen == nil && !errors.Is(offen, os.ErrNotExist):
				t.Error("die plist blieb, obwohl der Dienst sich entladen ließ")
			case entladen != nil && (offen != nil || !errors.Is(fehler, entlaedtNicht)):
				t.Errorf("die plist ging, obwohl der Dienst noch geladen sein kann: %v, %v", offen, fehler)
			}
		})
	}
}

// **Ein Fremder bleibt fremd** — Z5 an einer Datei, die er schrieb.
func TestEinFremderMitGleichemNamenWirdNichtRepariert(t *testing.T) {
	t.Parallel()

	ziel, datei := einDesktopziel(t)
	if fehler := os.MkdirAll(filepath.Dir(datei), 0o700); fehler != nil {
		t.Fatal(fehler)
	}

	fremd := `{"mcpServers": {"magicseteditor4mac": {"command": "/anderes/mse-mcp", "args": ["--x"]}}}`
	schreibeDatei(t, datei, fremd)

	if _, fehler := service.SchnittstelleEinrichten(t.Context(), ziel); !errors.Is(fehler, domain.ErrConflict) {
		t.Errorf("ein fremder Eintrag wurde repariert: %v", fehler)
	}

	if liesDatei(t, datei) != fremd {
		t.Error("der fremde Eintrag wurde angefasst")
	}
}

// **Z1, die Gegenrichtung: ist die Datei unverändert, gilt das Gedächtnis
// weiter** — ein Umzug der Anwendung repariert den Eintrag, und widerrufen
// geht danach trotzdem byte-gleich zum Anfang.
func TestReparierenOhneFremdeAenderungBleibtByteGleich(t *testing.T) {
	t.Parallel()

	ziel, datei := einDesktopziel(t)
	if fehler := os.MkdirAll(filepath.Dir(datei), 0o700); fehler != nil {
		t.Fatal(fehler)
	}

	vorher := "{\n    \"globalShortcut\": \"\"\n}\n"
	schreibeDatei(t, datei, vorher)
	ausgang(t)(service.SchnittstelleEinrichten(t.Context(), ziel))

	umgezogen := ziel
	umgezogen.Programm = "/Neu/MagicSetEditor4Mac.app/Contents/MacOS/mse-mcp"
	ausgang(t)(service.SchnittstelleEinrichten(t.Context(), umgezogen))

	if aus := ausgang(t)(service.SchnittstelleWiderrufen(t.Context(), umgezogen)); !aus.Bytegleich || liesDatei(t, datei) != vorher {
		t.Errorf("%+v\n%s", aus, liesDatei(t, datei))
	}
}

// **Z1 — ein Gedächtnis, das sich nicht vergessen lässt, hält das
// Reparieren an**, bevor die Datei geändert wird.
func TestEinUnvergesslichesGedaechtnisHaeltDasReparierenAn(t *testing.T) {
	t.Parallel()

	ziel, datei := einDesktopziel(t)
	if fehler := os.MkdirAll(filepath.Dir(datei), 0o700); fehler != nil {
		t.Fatal(fehler)
	}

	schreibeDatei(t, datei, "{}\n")
	ausgang(t)(service.SchnittstelleEinrichten(t.Context(), ziel))

	geaendert := strings.Replace(liesDatei(t, datei), "{", "{\n  \"neu\": true,", 1)
	schreibeDatei(t, datei, geaendert)
	sperren(t, filepath.Dir(ziel.Gedaechtnis), 0o500)

	umgezogen := ziel
	umgezogen.Programm = "/Neu/MagicSetEditor4Mac.app/Contents/MacOS/mse-mcp"

	if _, fehler := service.SchnittstelleEinrichten(t.Context(), umgezogen); fehler == nil {
		t.Fatal("repariert, obwohl das alte Gedächtnis blieb")
	}

	if liesDatei(t, datei) != geaendert {
		t.Error("die Datei wurde trotzdem geändert")
	}
}

// **N5 des Nachlaufs — eine gescheiterte Reparatur lässt das Gedächtnis,
// wie es war.** Hier wurde seine Prüfsumme fortgeschrieben, bevor die Datei
// es war; scheiterte das Schreiben, passte das Gedächtnis zu keiner Datei
// mehr, und ein Widerruf ließ die Datei stehen, die es vorher gar nicht gab.
func TestEineGescheiterteReparaturLaesstDasGedaechtnis(t *testing.T) {
	t.Parallel()

	ziel, datei := einDesktopziel(t)
	if fehler := os.MkdirAll(filepath.Dir(datei), 0o700); fehler != nil {
		t.Fatal(fehler)
	}

	ausgang(t)(service.SchnittstelleEinrichten(t.Context(), ziel))

	umgezogen := ziel
	umgezogen.Programm = "/Neu/MagicSetEditor4Mac.app/Contents/MacOS/mse-mcp"

	sperren(t, filepath.Dir(datei), 0o500)

	if _, fehler := service.SchnittstelleEinrichten(t.Context(), umgezogen); fehler == nil {
		t.Fatal("repariert, obwohl der Ordner gesperrt war")
	}

	if fehler := os.Chmod(filepath.Dir(datei), 0o700); fehler != nil { //nolint:gosec // ein Ordner dieses Tests, er braucht das Ausführungsrecht
		t.Fatal(fehler)
	}

	if aus := ausgang(t)(service.SchnittstelleWiderrufen(t.Context(), ziel)); !aus.Bytegleich {
		t.Errorf("nicht byte-gleich: %+v", aus)
	}

	if _, fehler := os.Stat(datei); !errors.Is(fehler, os.ErrNotExist) {
		t.Errorf("die Datei, die es vorher nicht gab, steht noch: %v", fehler)
	}
}

// **N5 — scheitert nach dem Reparieren das Fortschreiben des Gedächtnisses,
// geht nichts verloren**: es passt dann zu keiner Datei, und Widerrufen nimmt
// nur den eigenen Eintrag.
func TestEinGescheitertesFortschreibenNimmtNurDenEintrag(t *testing.T) {
	t.Parallel()

	ziel, datei := einDesktopziel(t)
	if fehler := os.MkdirAll(filepath.Dir(datei), 0o700); fehler != nil {
		t.Fatal(fehler)
	}

	schreibeDatei(t, datei, "{\"anderes\": 1}\n")
	ausgang(t)(service.SchnittstelleEinrichten(t.Context(), ziel))

	umgezogen := ziel
	umgezogen.Programm = "/Neu/MagicSetEditor4Mac.app/Contents/MacOS/mse-mcp"

	sperren(t, filepath.Dir(ziel.Gedaechtnis), 0o500)

	if _, fehler := service.SchnittstelleEinrichten(t.Context(), umgezogen); fehler == nil {
		t.Fatal("das Gedächtnis ließ sich nicht fortschreiben, und es kam kein Fehler")
	}

	if fehler := os.Chmod(filepath.Dir(ziel.Gedaechtnis), 0o700); fehler != nil { //nolint:gosec // ein Ordner dieses Tests, er braucht das Ausführungsrecht
		t.Fatal(fehler)
	}

	aus := ausgang(t)(service.SchnittstelleWiderrufen(t.Context(), umgezogen))
	if nachher := liesDatei(t, datei); aus.Bytegleich || !strings.Contains(nachher, `"anderes"`) ||
		strings.Contains(nachher, "magicseteditor4mac") {
		t.Errorf("%+v\n%s", aus, nachher)
	}
}
