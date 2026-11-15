// Die Handgriffe, mit denen die Tests eine Sicherung verderben.
//
// **Eine eigene Datei**, weil sie eine eigene Frage beantworten: `archiv_test`
// sagt, was gelten soll — hier steht, wie man einen Verstoß herstellt.
package sicherung_test

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"testing"
)

// umbauen schreibt ein Archiv neu und lässt dabei einen Eintrag ändern.
//
// **Neu geschrieben und nicht bearbeitet.** Ein ZIP an Ort und Stelle zu
// ändern hieße, das Format selbst nachzubauen — und dann prüfte der Test den
// Nachbau.
func umbauen(t *testing.T, pfad string, aendern func(name string, inhalt []byte) ([]byte, bool)) {
	t.Helper()

	roh, err := os.ReadFile(pfad) // #nosec G304 -- aus t.TempDir()
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	leser, err := zip.NewReader(bytes.NewReader(roh), int64(len(roh)))
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	var gebaut bytes.Buffer

	schreiber := zip.NewWriter(&gebaut)

	for _, eintrag := range leser.File {
		offen, oeffnungsfehler := eintrag.Open()
		if oeffnungsfehler != nil {
			t.Fatalf("%s öffnen: %v", eintrag.Name, oeffnungsfehler)
		}

		inhalt, lesefehler := io.ReadAll(offen)

		_ = offen.Close() //nolint:errcheck // gelesen ist gelesen

		if lesefehler != nil {
			t.Fatalf("%s lesen: %v", eintrag.Name, lesefehler)
		}

		neu, behalten := aendern(eintrag.Name, inhalt)
		if !behalten {
			continue
		}

		ziel, anlegefehler := schreiber.Create(eintrag.Name)
		if anlegefehler != nil {
			t.Fatalf("%s anlegen: %v", eintrag.Name, anlegefehler)
		}

		if _, schreibfehler := ziel.Write(neu); schreibfehler != nil {
			t.Fatalf("%s schreiben: %v", eintrag.Name, schreibfehler)
		}
	}

	if fehler := schreiber.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	if fehler := os.WriteFile(pfad, gebaut.Bytes(), 0o600); fehler != nil {
		t.Fatalf("zurückschreiben: %v", fehler)
	}
}

// verfaelschen tauscht den Inhalt eines Teils aus.
func verfaelschen(t *testing.T, pfad, name string, neu []byte) {
	t.Helper()

	umbauen(t, pfad, func(eintrag string, inhalt []byte) ([]byte, bool) {
		if eintrag == name {
			return neu, true
		}

		return inhalt, true
	})
}

// entfernen nimmt einen Teil aus dem Archiv.
func entfernen(t *testing.T, pfad, name string) {
	t.Helper()

	umbauen(t, pfad, func(eintrag string, inhalt []byte) ([]byte, bool) {
		return inhalt, eintrag != name
	})
}

// nurEintrag schreibt ein Archiv mit genau einem Eintrag — ohne Manifest.
func nurEintrag(t *testing.T, pfad, name string, inhalt []byte) {
	t.Helper()

	var gebaut bytes.Buffer

	schreiber := zip.NewWriter(&gebaut)

	ziel, err := schreiber.Create(name)
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if _, fehler := ziel.Write(inhalt); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if fehler := schreiber.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	if fehler := os.WriteFile(pfad, gebaut.Bytes(), 0o600); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}
}

// benennen tauscht den Namen eines Eintrags im Archiv.
func benennen(t *testing.T, pfad, alt, neu string) {
	t.Helper()

	roh, err := os.ReadFile(pfad) // #nosec G304 -- aus t.TempDir()
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	leser, err := zip.NewReader(bytes.NewReader(roh), int64(len(roh)))
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	var gebaut bytes.Buffer

	schreiber := zip.NewWriter(&gebaut)

	for _, eintrag := range leser.File {
		offen, oeffnungsfehler := eintrag.Open()
		if oeffnungsfehler != nil {
			t.Fatalf("%s öffnen: %v", eintrag.Name, oeffnungsfehler)
		}

		inhalt, lesefehler := io.ReadAll(offen)

		_ = offen.Close() //nolint:errcheck // gelesen ist gelesen

		if lesefehler != nil {
			t.Fatalf("%s lesen: %v", eintrag.Name, lesefehler)
		}

		name := eintrag.Name
		if name == alt {
			name = neu
		}

		// **Über den Kopf und nicht über `Create`.** `zip.Writer.Create`
		// säubert den Namen; ein Eintrag `../hinaus.txt` käme als
		// `hinaus.txt` heraus — und dann prüfte der Test nichts.
		ziel, anlegefehler := schreiber.CreateHeader(&zip.FileHeader{Name: name})
		if anlegefehler != nil {
			t.Fatalf("%s anlegen: %v", name, anlegefehler)
		}

		if _, schreibfehler := ziel.Write(inhalt); schreibfehler != nil {
			t.Fatalf("%s schreiben: %v", name, schreibfehler)
		}
	}

	if fehler := schreiber.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	if fehler := os.WriteFile(pfad, gebaut.Bytes(), 0o600); fehler != nil {
		t.Fatalf("zurückschreiben: %v", fehler)
	}
}
