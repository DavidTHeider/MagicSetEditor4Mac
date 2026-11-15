package app

// **Die Schnittstellen an der Naht** — OPS-09, Schnitt 1
// (#1907).
//
// **Jeder Lauf hat sein eigenes Zuhause** (§3): die Orte sind eingesetzt,
// und das echte `~/.zprofile` bekommt kein Lauf zu sehen.

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func eineEinrichtende(t *testing.T, pfad string) (anwendung *App, heim string) {
	t.Helper()

	heim = t.TempDir()
	if fehler := os.MkdirAll(filepath.Join(heim, ".codex"), 0o700); fehler != nil {
		t.Fatal(fehler)
	}

	anwendung = NeueApp("").MitSchnittstellenorten(
		func() (string, error) { return heim, nil },
		func() (string, error) { return "/Programme/MagicSetEditor4Mac.app/Contents/MacOS", nil },
		func(context.Context) string { return pfad },
	)

	return anwendung, heim
}

func standVon(antwort SchnittstellenAntwort, name string) SchnittstelleDTO {
	for _, eine := range antwort.Schnittstellen {
		if eine.Name == name {
			return eine
		}
	}

	return SchnittstelleDTO{}
}

func TestDieSchnittstellenLassenSichEinrichtenUndWiderrufen(t *testing.T) {
	t.Parallel()

	anwendung, heim := eineEinrichtende(t, "/usr/bin:/bin")
	ctx := context.Background()

	vorher := anwendung.Schnittstellen(ctx)
	if vorher.Fehler != nil || standVon(vorher, "codex").Stand != "offen" ||
		standVon(vorher, "befehlszeile").Stand != "offen" {
		t.Fatalf("vorher: %+v", vorher)
	}

	// **Die Vorschau nennt jede Zeile** — samt den Marken (AK-01).
	if zeilen := strings.Join(standVon(vorher, "codex").Zeilen, "\n"); !strings.Contains(zeilen,
		`command = "/Programme/MagicSetEditor4Mac.app/Contents/MacOS/mse-mcp"`) ||
		!strings.Contains(zeilen, "# >>> MagicSetEditor4Mac") {
		t.Errorf("Zeilen: %s", zeilen)
	}

	for _, name := range []string{"codex", "befehlszeile"} {
		if antwort := anwendung.SchnittstelleEinrichten(ctx, name); antwort.Fehler != nil ||
			standVon(antwort, name).Stand != "eingerichtet" {
			t.Fatalf("%s einrichten: %+v", name, antwort)
		}
	}

	zprofil, fehler := os.ReadFile(filepath.Join(heim, ".zprofile")) //nolint:gosec // eine Datei des Laufs
	if fehler != nil || !strings.Contains(string(zprofil), `export PATH="$HOME/.local/bin:$PATH"`) {
		t.Errorf(".zprofile: %q, %v", zprofil, fehler)
	}

	for _, name := range []string{"codex", "befehlszeile"} {
		if antwort := anwendung.SchnittstelleWiderrufen(ctx, name); antwort.Fehler != nil ||
			standVon(antwort, name).Stand != "offen" {
			t.Fatalf("%s widerrufen: %+v", name, antwort)
		}
	}

	for _, datei := range []string{".zprofile", filepath.Join(".codex", "config.toml")} {
		if _, fehler := os.Stat(filepath.Join(heim, datei)); !errors.Is(fehler, os.ErrNotExist) {
			t.Errorf("%s steht noch: %v", datei, fehler)
		}
	}
}

// **Was schon im PATH steht, braucht keinen Eintrag**, und ohne Codex gibt
// es keinen für Codex.
func TestUnnoetigUndNichtGefunden(t *testing.T) {
	t.Parallel()

	anwendung, heim := eineEinrichtende(t, "")
	anwendung.schnittstellen.pfad = func(context.Context) string {
		return filepath.Join(heim, ".local", "bin") + string(os.PathListSeparator) + "/usr/bin"
	}

	if fehler := os.RemoveAll(filepath.Join(heim, ".codex")); fehler != nil {
		t.Fatal(fehler)
	}

	antwort := anwendung.Schnittstellen(context.Background())
	if standVon(antwort, "befehlszeile").Stand != "unnoetig" || standVon(antwort, "codex").Stand != "nichtGefunden" {
		t.Errorf("Stand: %+v", antwort.Schnittstellen)
	}

	if einrichten := anwendung.SchnittstelleEinrichten(context.Background(), "codex"); einrichten.Fehler == nil {
		t.Error("Codex ohne Codex eingerichtet")
	}
}

// **Codex wird auch über seinen Befehl gefunden**, ohne Verzeichnis.
func TestCodexWirdUeberSeinenBefehlGefunden(t *testing.T) {
	t.Parallel()

	anwendung, heim := eineEinrichtende(t, "")
	behaelter := t.TempDir()

	if fehler := os.WriteFile(filepath.Join(behaelter, "codex"), []byte("#!/bin/sh\n"), 0o700); fehler != nil { //nolint:gosec // ein Probeprogramm
		t.Fatal(fehler)
	}

	anwendung.schnittstellen.pfad = func(context.Context) string { return behaelter }

	if fehler := os.RemoveAll(filepath.Join(heim, ".codex")); fehler != nil {
		t.Fatal(fehler)
	}

	if stand := standVon(anwendung.Schnittstellen(context.Background()), "codex").Stand; stand != "offen" {
		t.Errorf("Stand %q", stand)
	}
}

// **Was nicht geht, sagt es** — ein unbekannter Name, ein fehlendes
// Zuhause, ein fehlender Werkzeugort.
func TestSchnittstellenfehler(t *testing.T) {
	t.Parallel()

	anwendung, _ := eineEinrichtende(t, "")
	ctx := context.Background()

	if antwort := anwendung.SchnittstelleEinrichten(ctx, "quatsch"); antwort.Fehler == nil {
		t.Error("ein unbekannter Name kam durch")
	}

	kaputt := errors.New("kein Zuhause")
	ohneHeim := NeueApp("").MitSchnittstellenorten(
		func() (string, error) { return "", kaputt },
		func() (string, error) { return "/x", nil },
		func(context.Context) string { return "" },
	)

	if antwort := ohneHeim.Schnittstellen(ctx); antwort.Fehler == nil || antwort.Schnittstellen == nil {
		t.Errorf("ohne Zuhause: %+v", antwort)
	}

	if antwort := ohneHeim.SchnittstelleWiderrufen(ctx, "codex"); antwort.Fehler == nil {
		t.Error("ohne Zuhause widerrufen")
	}

	ohneWerkzeug := NeueApp("").MitSchnittstellenorten(
		func() (string, error) { return t.TempDir(), nil },
		func() (string, error) { return "", kaputt },
		func(context.Context) string { return "" },
	)

	if antwort := ohneWerkzeug.Schnittstellen(ctx); antwort.Fehler == nil {
		t.Error("ohne Werkzeugort")
	}
}

// **Die Orte des Betriebs stehen ab dem Bau** — gelesen, nicht geschrieben:
// derselbe Weg wie die Vorschau der Kommandozeile.
func TestDieOrteDesBetriebsStehenAbDemBau(t *testing.T) {
	t.Parallel()

	antwort := NeueApp("").Schnittstellen(t.Context())
	if antwort.Fehler != nil {
		t.Fatalf("im Betrieb: %+v", antwort.Fehler)
	}

	namen := make([]string, 0, len(antwort.Schnittstellen))
	for _, eine := range antwort.Schnittstellen {
		namen = append(namen, eine.Name)
	}

	if !slices.Equal(namen, []string{"befehlszeile", befehlsziel, "codex", "claudedesktop", "claudecode", "dienst"}) {
		t.Errorf("im Betrieb: %v", namen)
	}
}

// **Die Bindungen reichen durch** — dasselbe Zuhause, derselbe Stand.
func TestDieSchnittstellenbindungenReichenDurch(t *testing.T) {
	t.Parallel()

	anwendung, _ := eineEinrichtende(t, "/usr/bin")
	bindungen := NeueBindungen(anwendung)

	if stand := standVon(bindungen.Schnittstellen(t.Context()), "codex").Stand; stand != "offen" {
		t.Fatalf("vorher %q", stand)
	}

	if stand := standVon(bindungen.SchnittstelleEinrichten(t.Context(), "codex"), "codex").Stand; stand != "eingerichtet" {
		t.Errorf("eingerichtet %q", stand)
	}

	if stand := standVon(bindungen.SchnittstelleWiderrufen(t.Context(), "codex"), "codex").Stand; stand != "offen" {
		t.Errorf("widerrufen %q", stand)
	}
}
