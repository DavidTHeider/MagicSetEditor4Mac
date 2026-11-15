package service_test

// **Was an einer Schnittstelle schiefgeht** — OPS-09 (#1907). Herausgelöst
// aus `schnittstellen_test.go`; dort steht, was gelingt.
//
// **Jeder Fehlerweg lässt die Datei, wie sie war** — oder meldet, dass er es
// nicht konnte.

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// sperren entzieht einem Ort das Schreibrecht — und gibt es nach dem Lauf
// zurück, sonst räumte niemand das Verzeichnis weg.
func sperren(t *testing.T, ort string, recht os.FileMode) {
	t.Helper()

	if fehler := os.Chmod(ort, recht); fehler != nil {
		t.Fatal(fehler)
	}

	t.Cleanup(func() { _ = os.Chmod(ort, 0o700) }) //nolint:errcheck,gosec // nur aufgeräumt
}

// **Ein halber eigener Block ist fremd** — wer ihn ergänzte, riete, wo er
// endet.
func TestEinHalberEigenerBlockIstFremd(t *testing.T) {
	t.Parallel()

	ziel, datei := einCodexziel(t)
	schreibeDatei(t, datei, domain.BlockAnfang+"\n[mcp_servers.magicseteditor4mac]\n")

	if stand := service.SchnittstelleLesen(ziel); stand != service.SchnittstelleFremd {
		t.Errorf("Stand %q", stand)
	}
}

// **Widerrufen ohne Eintrag tut nichts** — und meldet nichts.
func TestWiderrufenOhneEintragTutNichts(t *testing.T) {
	t.Parallel()

	ziel, datei := einCodexziel(t)
	schreibeDatei(t, datei, "model = \"x\"\n")

	if aus, fehler := service.SchnittstelleWiderrufen(t.Context(), ziel); fehler != nil || aus.Stand != service.SchnittstelleOffen {
		t.Errorf("Widerrufen: %+v, %v", aus, fehler)
	}

	if liesDatei(t, datei) != "model = \"x\"\n" {
		t.Error("die Datei wurde angefasst")
	}
}

// **Liegt eine Datei, wo das Verzeichnis hingehört, wird nichts eingerichtet.**
func TestEinZielUnterEinerDateiMeldetSich(t *testing.T) {
	t.Parallel()

	ziel, datei := einCodexziel(t)
	schreibeDatei(t, datei, "")
	ziel.Datei = filepath.Join(datei, "unter", "config.toml")

	if _, fehler := service.SchnittstelleEinrichten(t.Context(), ziel); fehler == nil {
		t.Error("unter einer Datei eingerichtet")
	}
}

// **Ein Verzeichnis ohne Schreibrecht** lässt weder einrichten noch
// widerrufen — beides meldet sich, und die Datei bleibt.
func TestEinGesperrtesVerzeichnisMeldetSich(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		name, vorher string
	}{
		{"die Datei würde leer", ""},
		{"es bleibt etwas stehen", "model = \"x\"\n"},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			ziel, datei := einCodexziel(t)
			if fall.vorher != "" {
				schreibeDatei(t, datei, fall.vorher)
			}

			if _, fehler := service.SchnittstelleEinrichten(t.Context(), ziel); fehler != nil {
				t.Fatal(fehler)
			}

			eingerichtet := liesDatei(t, datei)
			sperren(t, filepath.Dir(datei), 0o500)

			if _, fehler := service.SchnittstelleWiderrufen(t.Context(), ziel); fehler == nil {
				t.Error("aus einem gesperrten Verzeichnis widerrufen")
			}

			if liesDatei(t, datei) != eingerichtet {
				t.Error("die Datei wurde trotzdem angefasst")
			}
		})
	}

	ziel, datei := einCodexziel(t)
	sperren(t, filepath.Dir(datei), 0o500)

	if _, fehler := service.SchnittstelleEinrichten(t.Context(), ziel); fehler == nil {
		t.Error("in ein gesperrtes Verzeichnis eingerichtet")
	}

	// **Und ein Verzeichnis, das erst entstehen müsste**, entsteht dort nicht.
	ziel.Datei = filepath.Join(filepath.Dir(datei), "neu", "config.toml")

	if _, fehler := service.SchnittstelleEinrichten(t.Context(), ziel); fehler == nil {
		t.Error("ein Verzeichnis in einem gesperrten angelegt")
	}
}

// **Eine Datei, die sich nicht öffnen lässt, ist fremd.**
func TestEineVerschlosseneDateiIstFremd(t *testing.T) {
	t.Parallel()

	ziel, datei := einCodexziel(t)
	schreibeDatei(t, datei, "model = \"x\"\n")
	sperren(t, datei, 0o000)

	if stand := service.SchnittstelleLesen(ziel); stand != service.SchnittstelleFremd {
		t.Errorf("Stand %q", stand)
	}

	if _, fehler := service.SchnittstelleWiderrufen(t.Context(), ziel); fehler == nil || errors.Is(fehler, domain.ErrConflict) {
		t.Errorf("Widerrufen: %v", fehler)
	}
}
