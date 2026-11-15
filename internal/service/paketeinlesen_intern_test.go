package service

// **Die Grenzen und Fehlerwege des Einlesens, die nur von innen erreichbar
// sind** — FMT-02.AK-04
// (#1691):
// die Grenzen sind Paketvariablen wie beim Bezug, der Bauort ein Feld.

import (
	"archive/zip"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

const eingelesen = "probe-mitgebracht.mse-include"

// einlesbar legt einen Installer mit einem Include ab.
func einlesbar(t *testing.T, weitere ...string) string {
	t.Helper()

	pfad := filepath.Join(t.TempDir(), "probe.mse-installer")

	datei, err := os.Create(pfad) // #nosec G304 -- im Testverzeichnis
	if err != nil {
		t.Fatal(err)
	}

	archiv := zip.NewWriter(datei)
	eintraege := append([]string{"installer", eingelesen + "/include"}, weitere...)
	inhalte := map[string]string{
		"installer":             "package:\n\tname: " + eingelesen + "\n",
		eingelesen + "/include": "mse version: 2.0.0\n",
	}

	for _, name := range eintraege {
		schreiber, err := archiv.Create(name)
		if err != nil {
			t.Fatal(err)
		}

		if _, err := schreiber.Write([]byte(inhalte[name] + "x")); err != nil {
			t.Fatal(err)
		}
	}

	if err := errors.Join(archiv.Close(), datei.Close()); err != nil {
		t.Fatal(err)
	}

	return pfad
}

// einleseplatz ist ein Verwalter mit eigenem Bauort.
func einleseplatz(t *testing.T) (ort string, verwalter *Paketverwalter) {
	t.Helper()

	ort = t.TempDir()
	verwalter = NeuerPaketverwalter(&kopierablage{}, ort)
	verwalter.bau = t.TempDir()

	return ort, verwalter
}

func TestEinInstallerMitZuVielenEintraegenWirdAbgewiesen(t *testing.T) {
	vorher := installereintraege
	installereintraege = 2

	t.Cleanup(func() { installereintraege = vorher })

	_, verwalter := einleseplatz(t)

	_, err := verwalter.Einlesen(t.Context(), einlesbar(t, eingelesen+"/bild.png"))

	var feldfehler *domain.FieldError
	if !errors.As(err, &feldfehler) || !strings.Contains(feldfehler.Message, "mehr als 2 Einträge") {
		t.Errorf("Fehler %v", err)
	}
}

// **Der Deckel ist ein Befund mit eigenem Satz** — „größer als erlaubt"
// sagt, was zu tun ist; „ließ sich nicht auspacken" nicht.
func TestDerDeckelHaeltDasEinlesen(t *testing.T) {
	vorher := installerdeckel
	installerdeckel = 4

	t.Cleanup(func() { installerdeckel = vorher })

	ort, verwalter := einleseplatz(t)

	ergebnis, err := verwalter.Einlesen(t.Context(), einlesbar(t))
	if err != nil {
		t.Fatal(err)
	}

	if len(ergebnis.Befunde) != 1 || !strings.Contains(ergebnis.Befunde[0].Meldung, "größer als erlaubt") {
		t.Errorf("Befunde %+v", ergebnis.Befunde)
	}

	if _, err := os.Stat(filepath.Join(ort, eingelesen)); !errors.Is(err, os.ErrNotExist) {
		t.Error("ein halb ausgepacktes Paket liegt im Verzeichnis")
	}
}

func TestEinUnbrauchbarerBauortHaeltDasEinlesen(t *testing.T) {
	_, verwalter := einleseplatz(t)

	keinOrdner := filepath.Join(t.TempDir(), "datei")
	if err := os.WriteFile(keinOrdner, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	verwalter.bau = keinOrdner

	ergebnis, err := verwalter.Einlesen(t.Context(), einlesbar(t))
	if err != nil {
		t.Fatal(err)
	}

	if len(ergebnis.Befunde) != 1 || ergebnis.Befunde[0].Meldung != "für das Paket fand sich kein Bauort" {
		t.Errorf("Befunde %+v", ergebnis.Befunde)
	}
}

func TestEinVerschlossenesVerzeichnisHaeltDasEinlesen(t *testing.T) {
	ort, verwalter := einleseplatz(t)

	if err := os.Chmod(ort, 0o500); err != nil { //nolint:gosec // das Verschließen ist der Fall
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = os.Chmod(ort, 0o700) }) //nolint:errcheck,gosec // nur für das Aufräumen

	ergebnis, err := verwalter.Einlesen(t.Context(), einlesbar(t))
	if err != nil {
		t.Fatal(err)
	}

	if len(ergebnis.Befunde) != 1 || ergebnis.Befunde[0].Meldung != "das Paket ließ sich nicht ins Verzeichnis legen" {
		t.Errorf("Befunde %+v", ergebnis.Befunde)
	}
}
