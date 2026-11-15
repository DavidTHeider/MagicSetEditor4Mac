package service_test

// **Ein Diagnosepaket auf die Platte schreiben** — Werkstück O10 von Phase 5.2
// (#1434).
//
// **Hier und nicht nur über `internal/app`**: die Naht gehört diesem Paket.
// Eine Naht, die nur über ihren Aufrufer geprüft wird, ist ungeprüft, sobald
// der Aufrufer wegfällt — derselbe Satz wie bei der Vorgangskennung (O9).

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestEinDiagnosepaketTraegtSeineTeileUndSeinManifest(t *testing.T) {
	t.Parallel()

	ziel := filepath.Join(t.TempDir(), "probe.msediagnose")

	fehler := service.SchreibeDiagnosepaket(ziel, "diagnose.json", "0.4.0",
		[]service.Paketteil{{Name: "fassung.txt", Inhalt: []byte("0.4.0\n")}},
		time.Now())
	if fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	archiv, oeffnungsfehler := zip.OpenReader(ziel)
	if oeffnungsfehler != nil {
		t.Fatalf("öffnen: %v", oeffnungsfehler)
	}

	defer func() {
		if schliessfehler := archiv.Close(); schliessfehler != nil {
			t.Errorf("schließen: %v", schliessfehler)
		}
	}()

	inhalte := map[string]string{}

	for _, eintrag := range archiv.File {
		offen, err := eintrag.Open()
		if err != nil {
			t.Fatalf("%s öffnen: %v", eintrag.Name, err)
		}

		roh, lesefehler := io.ReadAll(offen)
		if lesefehler != nil {
			t.Fatalf("%s lesen: %v", eintrag.Name, lesefehler)
		}

		if schliessfehler := offen.Close(); schliessfehler != nil {
			t.Errorf("%s schließen: %v", eintrag.Name, schliessfehler)
		}

		inhalte[eintrag.Name] = string(roh)
	}

	switch {
	case inhalte["fassung.txt"] != "0.4.0\n":
		t.Errorf("der Teil kam nicht an: %q", inhalte["fassung.txt"])
	case inhalte["diagnose.json"] == "":
		t.Errorf("kein Manifest: %v", inhalte)
	}
}

// **Ein bestehendes Paket wird nicht weggenommen** — und der Fehler sagt es.
func TestEinDiagnosepaketNimmtKeinBestehendesWeg(t *testing.T) {
	t.Parallel()

	ziel := filepath.Join(t.TempDir(), "probe.msediagnose")
	if fehler := os.WriteFile(ziel, []byte("gehört jemandem"), 0o600); fehler != nil {
		t.Fatalf("vorbelegen: %v", fehler)
	}

	fehler := service.SchreibeDiagnosepaket(ziel, "diagnose.json", "0.4.0",
		[]service.Paketteil{{Name: "fassung.txt", Inhalt: []byte("0.4.0\n")}},
		time.Now())
	if fehler == nil {
		t.Fatal("das bestehende Paket wurde weggenommen")
	}

	gelesen, lesefehler := os.ReadFile(ziel) // #nosec G304 -- Pfad aus t.TempDir()
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if string(gelesen) != "gehört jemandem" {
		t.Errorf("die bestehende Datei ist verändert: %q", gelesen)
	}
}
