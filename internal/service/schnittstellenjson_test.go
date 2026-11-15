package service_test

// **Claude Desktop: ein Schlüssel in einer JSON-Datei** — OPS-09b
// (#1961).

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

const desktopprogramm = "/Programme/MagicSetEditor4Mac.app/Contents/MacOS/mse-mcp"

func einDesktopziel(t *testing.T) (ziel service.Schnittstellenziel, datei string) {
	t.Helper()

	ordner := t.TempDir()
	datei = filepath.Join(ordner, "Claude", "claude_desktop_config.json")

	return service.Schnittstellenziel{
		Name: "claudedesktop", Art: service.ArtJSON, Datei: datei, Vorhanden: true,
		Schluessel: "magicseteditor4mac", Programm: desktopprogramm,
		Gedaechtnis: filepath.Join(ordner, "gedaechtnis", "claudedesktop.json"),
	}, datei
}

// ausgang verlangt, dass ein Schritt gelang, und gibt seinen Ausgang.
func ausgang(t *testing.T) func(service.Schnittstellenausgang, error) service.Schnittstellenausgang {
	t.Helper()

	return func(stand service.Schnittstellenausgang, fehler error) service.Schnittstellenausgang {
		t.Helper()

		if fehler != nil {
			t.Fatalf("%+v: %v", stand, fehler)
		}

		return stand
	}
}

// **Einrichten und Widerrufen gehen byte-gleich zurück** — auch die
// Einrückung des Menschen, denn der Stand davor ist gemerkt (AK-02).
func TestClaudeDesktopGehtByteGleichZurueck(t *testing.T) {
	t.Parallel()

	ziel, datei := einDesktopziel(t)
	if fehler := os.MkdirAll(filepath.Dir(datei), 0o700); fehler != nil {
		t.Fatal(fehler)
	}

	vorher := "{\n    \"mcpServers\": {\n        \"anderer\": {\"command\": \"x\"}\n    },\n    \"globalShortcut\": \"\"\n}\n"
	schreibeDatei(t, datei, vorher)

	if stand := service.SchnittstelleLesen(ziel); stand != service.SchnittstelleOffen {
		t.Fatalf("vorher: %q", stand)
	}

	if aus := ausgang(t)(service.SchnittstelleEinrichten(t.Context(), ziel)); aus.Stand != service.SchnittstelleEingerichtet {
		t.Fatalf("Einrichten: %+v", aus)
	}

	eingerichtet := liesDatei(t, datei)
	if !strings.Contains(eingerichtet, `"anderer"`) || !strings.Contains(eingerichtet, desktopprogramm) {
		t.Fatalf("eingerichtet:\n%s", eingerichtet)
	}

	// **Zweimal einrichten ändert nichts.**
	ausgang(t)(service.SchnittstelleEinrichten(t.Context(), ziel))

	if liesDatei(t, datei) != eingerichtet {
		t.Error("zweimal geschrieben")
	}

	aus := ausgang(t)(service.SchnittstelleWiderrufen(t.Context(), ziel))
	if aus.Stand != service.SchnittstelleOffen || !aus.Bytegleich || liesDatei(t, datei) != vorher {
		t.Errorf("widerrufen: %+v\n%s", aus, liesDatei(t, datei))
	}

	if _, fehler := os.Stat(ziel.Gedaechtnis); !errors.Is(fehler, os.ErrNotExist) {
		t.Errorf("das Gedächtnis blieb: %v", fehler)
	}
}

// **Ohne Datei entsteht sie, und widerrufen nimmt sie wieder.**
func TestClaudeDesktopOhneDatei(t *testing.T) {
	t.Parallel()

	ziel, datei := einDesktopziel(t)

	ausgang(t)(service.SchnittstelleEinrichten(t.Context(), ziel))

	if aus := ausgang(t)(service.SchnittstelleWiderrufen(t.Context(), ziel)); !aus.Bytegleich {
		t.Errorf("%+v", aus)
	}

	if _, fehler := os.Stat(datei); !errors.Is(fehler, os.ErrNotExist) {
		t.Errorf("die Datei steht noch: %v", fehler)
	}
}

// **Hat der Mensch die Datei seither geändert, geht nur der eigene Eintrag**
// — und das wird gesagt, statt ihm seine Änderung zu nehmen.
func TestClaudeDesktopNachFremderAenderung(t *testing.T) {
	t.Parallel()

	ziel, datei := einDesktopziel(t)

	ausgang(t)(service.SchnittstelleEinrichten(t.Context(), ziel))

	geaendert := strings.Replace(liesDatei(t, datei), "{", "{\n  \"neu\": true,", 1)
	schreibeDatei(t, datei, geaendert)

	aus := ausgang(t)(service.SchnittstelleWiderrufen(t.Context(), ziel))
	nachher := liesDatei(t, datei)

	if aus.Bytegleich || aus.Stand != service.SchnittstelleOffen ||
		!strings.Contains(nachher, `"neu": true`) || strings.Contains(nachher, "magicseteditor4mac") {
		t.Errorf("%+v\n%s", aus, nachher)
	}
}

// **Die Stände**: fremd, veraltet, nicht gefunden — und was einrichten und
// widerrufen dann tun.
func TestClaudeDesktopStaende(t *testing.T) {
	t.Parallel()

	eintrag := func(befehl string) string {
		return `{"mcpServers": {"magicseteditor4mac": {"command": "` + befehl + `"}}}`
	}

	for _, fall := range []struct {
		name, inhalt string
		vorhanden    bool
		stand        service.Schnittstellenstand
		einrichten   error
	}{
		{"fremd", eintrag("/bin/anderes"), true, service.SchnittstelleFremd, domain.ErrConflict},
		{"veraltet", eintrag("/Alt/MagicSetEditor4Mac.app/Contents/MacOS/mse-mcp"), true, service.SchnittstelleVeraltet, nil},
		{"fremder Pfad, eigene Gestalt", eintrag("/anderes/mse-mcp"), true, service.SchnittstelleFremd, domain.ErrConflict},
		{"kaputt", `{"mcpServers": `, true, service.SchnittstelleFremd, domain.ErrConflict},
		{"kein Objekt", `{"mcpServers": {"magicseteditor4mac": 3}}`, true, service.SchnittstelleFremd, domain.ErrConflict},
		{"nicht gefunden", ``, false, service.SchnittstelleNichtGefunden, domain.ErrValidation},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			ziel, datei := einDesktopziel(t)
			ziel.Vorhanden = fall.vorhanden

			if fall.inhalt != "" {
				if fehler := os.MkdirAll(filepath.Dir(datei), 0o700); fehler != nil {
					t.Fatal(fehler)
				}

				schreibeDatei(t, datei, fall.inhalt)
			}

			if stand := service.SchnittstelleLesen(ziel); stand != fall.stand {
				t.Fatalf("Stand %q", stand)
			}

			aus, fehler := service.SchnittstelleEinrichten(t.Context(), ziel)
			if !errors.Is(fehler, fall.einrichten) || (fall.einrichten == nil && aus.Stand != service.SchnittstelleEingerichtet) {
				t.Errorf("Einrichten: %+v, %v", aus, fehler)
			}

			if fall.stand == service.SchnittstelleFremd {
				if _, fehler := service.SchnittstelleWiderrufen(t.Context(), ziel); !errors.Is(fehler, domain.ErrConflict) {
					t.Errorf("Widerrufen: %v", fehler)
				}

				if liesDatei(t, datei) != fall.inhalt {
					t.Error("ein fremder Eintrag wurde angefasst")
				}
			}
		})
	}
}

// **Ein Gedächtnis, das sich nicht lesen lässt, ist keines** — widerrufen
// nimmt dann nur den eigenen Eintrag.
func TestClaudeDesktopMitKaputtemGedaechtnis(t *testing.T) {
	t.Parallel()

	ziel, datei := einDesktopziel(t)

	ausgang(t)(service.SchnittstelleEinrichten(t.Context(), ziel))
	schreibeDatei(t, ziel.Gedaechtnis, "kein json")

	if aus := ausgang(t)(service.SchnittstelleWiderrufen(t.Context(), ziel)); aus.Bytegleich ||
		strings.Contains(liesDatei(t, datei), "magicseteditor4mac") {
		t.Errorf("%+v", aus)
	}
}
