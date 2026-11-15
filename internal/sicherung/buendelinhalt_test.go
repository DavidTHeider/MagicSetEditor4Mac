package sicherung_test

// **Ein Teil aus dem Arbeitsspeicher und ein eigener Manifestname** —
// Werkstück O10 von Phase 5.2
// (#1434).

import (
	"archive/zip"
	"io"
	"path/filepath"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/sicherung"
)

func TestEinTeilOhneDateiKommtAusDemInhalt(t *testing.T) {
	t.Parallel()

	ziel := filepath.Join(t.TempDir(), "probe.msediagnose")

	fehler := sicherung.PackenAls(ziel, "diagnose.json", []sicherung.Teil{
		{Name: "uebersicht.json", Inhalt: []byte(`{"teile":[]}`)},
		{Name: "leer.txt"},
	}, sicherung.Angaben{Fassung: "0.4.0", Erstellt: time.Now()})
	if fehler != nil {
		t.Fatalf("packen: %v", fehler)
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

	namen := map[string]string{}

	for _, eintrag := range archiv.File {
		namen[eintrag.Name] = lies(t, eintrag)
	}

	switch {
	// **Der eigene Manifestname**, und der der Sicherung steht nicht daneben:
	// zwei Formate unter einem Namen wären die Verwechslung, gegen die es die
	// Unterscheidung gibt.
	case namen["diagnose.json"] == "":
		t.Errorf("kein Manifest »diagnose.json«: %v", namen)
	case namen[sicherung.Manifestname] != "":
		t.Errorf("das Sicherungsmanifest steht im Diagnosepaket")
	case namen["uebersicht.json"] != `{"teile":[]}`:
		t.Errorf("der Inhalt kam nicht an: %q", namen["uebersicht.json"])

	// **Ein leerer Teil ist ein Teil ohne Inhalt und kein Fehler.**
	case namen["leer.txt"] != "":
		t.Errorf("der leere Teil trug %q", namen["leer.txt"])
	}

	if _, da := namen["leer.txt"]; !da {
		t.Error("der leere Teil fehlt ganz")
	}
}

// lies gibt den Inhalt eines Archiveintrags.
func lies(t *testing.T, eintrag *zip.File) string {
	t.Helper()

	offen, fehler := eintrag.Open()
	if fehler != nil {
		t.Fatalf("%s öffnen: %v", eintrag.Name, fehler)
	}

	defer func() { _ = offen.Close() }() //nolint:errcheck // nur gelesen

	roh, lesefehler := io.ReadAll(offen)
	if lesefehler != nil {
		t.Fatalf("%s lesen: %v", eintrag.Name, lesefehler)
	}

	return string(roh)
}
