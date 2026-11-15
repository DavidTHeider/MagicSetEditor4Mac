package service_test

// **CLI und MCP per Knopf einrichten und widerrufen** — OPS-09, Schnitt 1
// (#1907).
//
// Gemessen an echten Dateien in einem eigenen Verzeichnis: einrichten ist
// idempotent, widerrufen hinterlässt die Datei byte-gleich (AK-02), und was
// fremd ist, wird erkannt und nicht angefasst (AK-01).

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func einCodexziel(t *testing.T) (ziel service.Schnittstellenziel, datei string) {
	t.Helper()

	ordner := filepath.Join(t.TempDir(), ".codex")
	if fehler := os.MkdirAll(ordner, 0o700); fehler != nil {
		t.Fatal(fehler)
	}

	datei = filepath.Join(ordner, "config.toml")

	return service.Schnittstellenziel{
		Name: "codex", Datei: datei, Vorhanden: true,
		Block:    "[mcp_servers.magicseteditor4mac]\ncommand = \"/Programme/MagicSetEditor4Mac.app/Contents/MacOS/mse-mcp\"\n",
		Anspruch: "[mcp_servers.magicseteditor4mac]",
	}, datei
}

func schreibeDatei(t *testing.T, pfad, inhalt string) {
	t.Helper()

	if fehler := os.WriteFile(pfad, []byte(inhalt), 0o600); fehler != nil {
		t.Fatal(fehler)
	}
}

func liesDatei(t *testing.T, pfad string) string {
	t.Helper()

	inhalt, fehler := os.ReadFile(pfad) //nolint:gosec // eine Datei des Laufs
	if fehler != nil {
		t.Fatal(fehler)
	}

	return string(inhalt)
}

func TestEinrichtenUndWiderrufenSindUmkehrbar(t *testing.T) {
	t.Parallel()

	ziel, datei := einCodexziel(t)
	vorher := "model = \"gpt-6.1-sol\"\n\n[mcp_servers.anderer]\ncommand = \"x\"\n"
	schreibeDatei(t, datei, vorher)

	if stand := service.SchnittstelleLesen(ziel); stand != service.SchnittstelleOffen {
		t.Fatalf("vorher: %q", stand)
	}

	if aus, fehler := service.SchnittstelleEinrichten(t.Context(), ziel); fehler != nil || aus.Stand != service.SchnittstelleEingerichtet {
		t.Fatalf("Einrichten: %+v, %v", aus, fehler)
	}

	eingerichtet := liesDatei(t, datei)

	// **Zweimal einrichten ändert nichts mehr.**
	if _, fehler := service.SchnittstelleEinrichten(t.Context(), ziel); fehler != nil || liesDatei(t, datei) != eingerichtet {
		t.Errorf("zweimal: %v", fehler)
	}

	if aus, fehler := service.SchnittstelleWiderrufen(t.Context(), ziel); fehler != nil || aus.Stand != service.SchnittstelleOffen {
		t.Fatalf("Widerrufen: %+v, %v", aus, fehler)
	}

	if nachher := liesDatei(t, datei); nachher != vorher {
		t.Errorf("nicht byte-gleich:\n%q\n%q", nachher, vorher)
	}
}

// **Ohne Datei entsteht sie — und widerrufen nimmt sie wieder**, damit
// nichts zurückbleibt, was vorher nicht da war.
func TestOhneDateiEntstehtSieUndGehtWieder(t *testing.T) {
	t.Parallel()

	ziel, datei := einCodexziel(t)

	if _, fehler := service.SchnittstelleEinrichten(t.Context(), ziel); fehler != nil {
		t.Fatalf("Einrichten: %v", fehler)
	}

	if angaben, fehler := os.Stat(datei); fehler != nil || angaben.Mode().Perm() != 0o600 {
		t.Fatalf("die Datei: %v", fehler)
	}

	if _, fehler := service.SchnittstelleWiderrufen(t.Context(), ziel); fehler != nil {
		t.Fatalf("Widerrufen: %v", fehler)
	}

	if _, fehler := os.Stat(datei); !errors.Is(fehler, os.ErrNotExist) {
		t.Errorf("die Datei steht noch: %v", fehler)
	}
}

// **Ein fremder Eintrag mit demselben Namen wird erkannt und nicht
// überschrieben** (AK-01).
func TestEinFremderEintragBleibt(t *testing.T) {
	t.Parallel()

	ziel, datei := einCodexziel(t)
	vorher := "[mcp_servers.magicseteditor4mac]\ncommand = \"/woanders/mse-mcp\"\n"
	schreibeDatei(t, datei, vorher)

	if stand := service.SchnittstelleLesen(ziel); stand != service.SchnittstelleFremd {
		t.Fatalf("Stand %q", stand)
	}

	if _, fehler := service.SchnittstelleEinrichten(t.Context(), ziel); !errors.Is(fehler, domain.ErrConflict) {
		t.Errorf("Einrichten: %v", fehler)
	}

	if _, fehler := service.SchnittstelleWiderrufen(t.Context(), ziel); !errors.Is(fehler, domain.ErrConflict) {
		t.Errorf("Widerrufen: %v", fehler)
	}

	if liesDatei(t, datei) != vorher {
		t.Error("ein fremder Eintrag wurde angefasst")
	}
}

// **Veraltet ist ein eigener Block mit anderem Inhalt** — die Anwendung liegt
// woanders; einrichten repariert ihn an seiner Stelle (AK-03).
func TestEinVeralteterEintragWirdRepariert(t *testing.T) {
	t.Parallel()

	ziel, datei := einCodexziel(t)
	alt := ziel
	alt.Block = "[mcp_servers.magicseteditor4mac]\ncommand = \"/alt/mse-mcp\"\n"

	if _, fehler := service.SchnittstelleEinrichten(t.Context(), alt); fehler != nil {
		t.Fatal(fehler)
	}

	if stand := service.SchnittstelleLesen(ziel); stand != service.SchnittstelleVeraltet {
		t.Fatalf("Stand %q", stand)
	}

	if aus, fehler := service.SchnittstelleEinrichten(t.Context(), ziel); fehler != nil || aus.Stand != service.SchnittstelleEingerichtet {
		t.Errorf("repariert: %+v, %v", aus, fehler)
	}

	if _, fehler := service.SchnittstelleWiderrufen(t.Context(), ziel); fehler != nil {
		t.Fatal(fehler)
	}

	if _, fehler := os.Stat(datei); !errors.Is(fehler, os.ErrNotExist) {
		t.Errorf("die Datei steht noch: %v", fehler)
	}
}

// **Ein Klient, der nicht da ist, wird nicht eingerichtet** — und was nicht
// nötig ist, auch nicht.
func TestNichtDaUndNichtNoetig(t *testing.T) {
	t.Parallel()

	ziel, datei := einCodexziel(t)

	fehlt := ziel
	fehlt.Vorhanden = false

	if stand := service.SchnittstelleLesen(fehlt); stand != service.SchnittstelleNichtGefunden {
		t.Errorf("Stand %q", stand)
	}

	if _, fehler := service.SchnittstelleEinrichten(t.Context(), fehlt); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("Einrichten: %v", fehler)
	}

	unnoetig := ziel
	unnoetig.Unnoetig = true

	if stand := service.SchnittstelleLesen(unnoetig); stand != service.SchnittstelleUnnoetig {
		t.Errorf("Stand %q", stand)
	}

	if _, fehler := service.SchnittstelleEinrichten(t.Context(), unnoetig); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("Einrichten: %v", fehler)
	}

	if _, fehler := os.Stat(datei); !errors.Is(fehler, os.ErrNotExist) {
		t.Errorf("es entstand eine Datei: %v", fehler)
	}
}

// **Eine Datei, die sich nicht lesen lässt, wird nicht überschrieben.**
func TestEineUnlesbareDateiBleibt(t *testing.T) {
	t.Parallel()

	ziel, datei := einCodexziel(t)
	if fehler := os.Mkdir(datei, 0o700); fehler != nil {
		t.Fatal(fehler)
	}

	if stand := service.SchnittstelleLesen(ziel); stand != service.SchnittstelleFremd {
		t.Errorf("Stand %q", stand)
	}

	if _, fehler := service.SchnittstelleEinrichten(t.Context(), ziel); fehler == nil {
		t.Error("über ein Verzeichnis eingerichtet")
	}

	if _, fehler := service.SchnittstelleWiderrufen(t.Context(), ziel); fehler == nil {
		t.Error("aus einem Verzeichnis widerrufen")
	}
}

// **Ein verknüpftes Profil bleibt verknüpft** — wer `~/.zprofile` aus einem
// Dotfiles-Verzeichnis holt, findet danach den Block dort und die
// Verknüpfung, wo sie war.
func TestEinSymlinkBleibtEinSymlink(t *testing.T) {
	t.Parallel()

	ordner := t.TempDir()
	echt := filepath.Join(ordner, "dotfiles-zprofile")
	verknuepft := filepath.Join(ordner, ".zprofile")

	schreibeDatei(t, echt, "export EDITOR=vi\n")

	if fehler := os.Symlink(echt, verknuepft); fehler != nil {
		t.Fatal(fehler)
	}

	ziel := service.Schnittstellenziel{
		Name: "befehlszeile", Datei: verknuepft, Vorhanden: true,
		Block: "export PATH=\"$HOME/.local/bin:$PATH\"\n",
	}

	if _, fehler := service.SchnittstelleEinrichten(t.Context(), ziel); fehler != nil {
		t.Fatal(fehler)
	}

	if angaben, fehler := os.Lstat(verknuepft); fehler != nil || angaben.Mode()&os.ModeSymlink == 0 {
		t.Errorf("die Verknüpfung ist fort: %v", fehler)
	}

	if !bytes.Contains([]byte(liesDatei(t, echt)), []byte(".local/bin")) {
		t.Error("der Block steht nicht in der verknüpften Datei")
	}

	if _, fehler := service.SchnittstelleWiderrufen(t.Context(), ziel); fehler != nil || liesDatei(t, echt) != "export EDITOR=vi\n" {
		t.Errorf("Widerrufen: %v", fehler)
	}
}

// **Eine Datei über der Grenze wird nicht gelesen und nicht geschrieben** —
// ein Shellprofil von einem Mebibyte ist keines, das dieses Programm
// versteht, und es ganz zu lesen hieße, einer fremden Datei den Speicher zu
// überlassen.
func TestEineZuGrosseDateiBleibt(t *testing.T) {
	t.Parallel()

	ziel, datei := einCodexziel(t)
	gross := bytes.Repeat([]byte("# x\n"), service.Schnittstellendateigrenze/4+1)
	schreibeDatei(t, datei, string(gross))

	if stand := service.SchnittstelleLesen(ziel); stand != service.SchnittstelleFremd {
		t.Errorf("Stand %q", stand)
	}

	if _, fehler := service.SchnittstelleEinrichten(t.Context(), ziel); fehler == nil {
		t.Error("in eine zu große Datei eingerichtet")
	}

	if liesDatei(t, datei) != string(gross) {
		t.Error("die Datei wurde angefasst")
	}
}
