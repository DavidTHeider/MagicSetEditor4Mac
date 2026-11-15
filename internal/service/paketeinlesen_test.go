package service_test

// **Einen `.mse-installer` einlesen** — FMT-02.AK-04
// (#1691).
//
// So wird eine Vorlage bei MSE2 weitergegeben; bis hierher las dieses
// Programm keinen. Gegen dieselben Zusagen wie die Übernahme aus einer
// MSE2-Installation: was schon daliegt, wird **übergangen und nicht
// überschrieben**, ein unlesbares Paket kommt **nicht** ins Verzeichnis, und
// was eingelesen ist, trägt den Installer als Herkunft.

import (
	"archive/zip"
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestEinInstallerWirdEingelesen(t *testing.T) {
	ort, verwalter, _ := kopierplatz(t)
	quelle := installerdatei(t, map[string]string{
		"installer":                               "package:\n\tname: probe-mitgebracht.mse-include\n",
		"probe-mitgebracht.mse-include/include":   "mse version: 2.0.0\nshort name: Mitgebracht\n",
		"probe-mitgebracht.mse-include/bild.png":  "png",
		"probe-mitgebracht.mse-include/a/b/c.txt": "tief",
	})

	ergebnis, err := verwalter.Einlesen(t.Context(), quelle)
	if err != nil {
		t.Fatalf("einlesen: %v", err)
	}

	if len(ergebnis.Uebernommen) != 1 || len(ergebnis.Befunde) != 0 {
		t.Fatalf("Ergebnis %+v", ergebnis)
	}

	paket := ergebnis.Uebernommen[0]
	if paket.Herkunft != quelle || paket.HerkunftsartIn(ort) != domain.PaketUebernommen {
		t.Errorf("Herkunft %q, Art %q", paket.Herkunft, paket.HerkunftsartIn(ort))
	}

	if inhalt := datei(t, filepath.Join(ort, "probe-mitgebracht.mse-include", "a", "b", "c.txt")); inhalt != "tief" {
		t.Errorf("ausgepackt %q", inhalt)
	}

	// **Der Bauort bleibt nicht liegen.**
	eintraege, _ := os.ReadDir(ort) //nolint:errcheck // leer heißt hier: nichts liegen geblieben
	for _, eintrag := range eintraege {
		if strings.HasPrefix(eintrag.Name(), "mse4-einlesen-") {
			t.Errorf("liegen geblieben: %s", eintrag.Name())
		}
	}
}

func TestEinVorhandenesPaketWirdUebergangen(t *testing.T) {
	ort, verwalter, _ := kopierplatz(t)
	quelle := installerdatei(t, map[string]string{
		"installer":                             "package:\n\tname: probe-mitgebracht.mse-include\n",
		"probe-mitgebracht.mse-include/include": "mse version: 2.0.0\n",
	})

	if _, err := verwalter.Einlesen(t.Context(), quelle); err != nil {
		t.Fatal(err)
	}

	bearbeitet := filepath.Join(ort, "probe-mitgebracht.mse-include", "include")
	if err := os.WriteFile(bearbeitet, []byte("mse version: 2.0.0\n# meins\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	ergebnis, err := verwalter.Einlesen(t.Context(), quelle)
	if err != nil {
		t.Fatal(err)
	}

	if !slices.Equal(ergebnis.Uebergangen, []string{"probe-mitgebracht.mse-include"}) || len(ergebnis.Uebernommen) != 0 {
		t.Errorf("Ergebnis %+v", ergebnis)
	}

	if !strings.Contains(datei(t, bearbeitet), "# meins") {
		t.Error("die Bearbeitung wurde überschrieben")
	}
}

// **Was der Installer benennt, kommt mit** — und ein Paket ohne Hauptdatei
// bleibt draußen, statt als halbes im Verzeichnis zu stehen.
func TestBefundeDesInstallersKommenMit(t *testing.T) {
	ort, verwalter, _ := kopierplatz(t)
	quelle := installerdatei(t, map[string]string{
		"installer": "package:\n\tname: probe-gut.mse-include\npackage:\n\tname: probe-kaputt.mse-game\n" +
			"package:\n\tname: probe-leer.mse-style\n",
		"probe-gut.mse-include/include":          "mse version: 2.0.0\n",
		"probe-gut.mse-include/../../weg.txt":    "fremd",
		"probe-kaputt.mse-game/nur-ein-bild.png": "png",
	})

	ergebnis, err := verwalter.Einlesen(t.Context(), quelle)
	if err != nil {
		t.Fatal(err)
	}

	if len(ergebnis.Uebernommen) != 1 || ergebnis.Uebernommen[0].Name != "probe-gut" {
		t.Errorf("übernommen %+v", ergebnis.Uebernommen)
	}

	// Der Ausbruch, das Paket ohne Hauptdatei, das Paket ohne Dateien.
	if len(ergebnis.Befunde) != 3 {
		t.Errorf("Befunde %+v", ergebnis.Befunde)
	}

	for _, name := range []string{"probe-kaputt.mse-game", "probe-leer.mse-style"} {
		if _, err := os.Stat(filepath.Join(ort, name)); !errors.Is(err, os.ErrNotExist) {
			t.Errorf("%s liegt im Verzeichnis", name)
		}
	}
}

func TestWasKeinInstallerIstWirdAbgewiesen(t *testing.T) {
	_, verwalter, _ := kopierplatz(t)
	ablage := t.TempDir()

	keinZip := filepath.Join(ablage, "text.mse-installer")
	if err := os.WriteFile(keinZip, []byte("kein Archiv"), 0o600); err != nil {
		t.Fatal(err)
	}

	ohneHaupt := installerdatei(t, map[string]string{"probe.mse-game/game": "mse version: 2.0.0\n"})

	for name, quelle := range map[string]string{
		"eine fehlende Datei":        filepath.Join(ablage, "fehlt.mse-installer"),
		"ein Ordner":                 ablage,
		"kein Archiv":                keinZip,
		"ein Archiv ohne Hauptdatei": ohneHaupt,
	} {
		t.Run(name, func(t *testing.T) {
			_, err := verwalter.Einlesen(t.Context(), quelle)

			var feldfehler *domain.FieldError
			if !errors.As(err, &feldfehler) || feldfehler.Field != "quelle" {
				t.Errorf("Fehler %v, erwartet einen an der Quelle", err)
			}
		})
	}
}

func TestOhneFestenOrtWirdNichtsEingelesen(t *testing.T) {
	verwalter := service.NeuerPaketverwalter(&ablage{}, "")

	if _, err := verwalter.Einlesen(t.Context(), "egal.mse-installer"); err == nil {
		t.Error("ohne Paketverzeichnis wurde eingelesen")
	}
}

// **Ein Katalog, der nicht mitkommt, ist ein Fehler — und das Paket liegt
// trotzdem da** (§7, Regel 3); das Ergebnis sagt, was geschehen ist.
func TestEinScheiternderAbgleichWirdGemeldet(t *testing.T) {
	ort, verwalter, katalog := kopierplatz(t)
	katalog.fehler = errors.New("der Bestand ist gesperrt")

	quelle := installerdatei(t, map[string]string{
		"installer":                             "package:\n\tname: probe-mitgebracht.mse-include\n",
		"probe-mitgebracht.mse-include/include": "mse version: 2.0.0\n",
	})

	ergebnis, err := verwalter.Einlesen(t.Context(), quelle)
	if err == nil || ergebnis == nil {
		t.Fatalf("Ergebnis %+v, Fehler %v", ergebnis, err)
	}

	if _, err := os.Stat(filepath.Join(ort, "probe-mitgebracht.mse-include", "include")); err != nil {
		t.Errorf("das Paket liegt nicht da: %v", err)
	}
}

// installerdatei legt einen Installer aus diesen Einträgen ab.
func installerdatei(t *testing.T, eintraege map[string]string) string {
	t.Helper()

	var puffer bytes.Buffer
	archiv := zip.NewWriter(&puffer)

	namen := make([]string, 0, len(eintraege))
	for name := range eintraege {
		namen = append(namen, name)
	}

	slices.Sort(namen)

	for _, name := range namen {
		schreiber, err := archiv.Create(name)
		if err != nil {
			t.Fatal(err)
		}

		if _, err := schreiber.Write([]byte(eintraege[name])); err != nil {
			t.Fatal(err)
		}
	}

	if err := archiv.Close(); err != nil {
		t.Fatal(err)
	}

	pfad := filepath.Join(t.TempDir(), "probe.mse-installer")
	if err := os.WriteFile(pfad, puffer.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}

	return pfad
}
