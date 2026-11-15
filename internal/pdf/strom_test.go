package pdf_test

import (
	"bytes"
	"compress/zlib"
	"io"
	"regexp"
	"testing"
)

// ausgepackterStrom holt den ersten Bildstrom aus einem PDF und packt ihn aus.
//
// **Von Hand und ohne Bibliothek.** Ein PDF-Leser als Testabhängigkeit prüfte
// dieselbe Sache zweimal: dass wir und er dasselbe für ein PDF halten. Hier
// wird gesucht, was der Schreiber hingeschrieben haben muss — und der fremde
// Leser steht daneben (`quartzMisst`).
func ausgepackterStrom(t *testing.T, dokument []byte) []byte {
	t.Helper()

	stelle := regexp.MustCompile(`(?s)/Subtype /Image.*?stream\r?\n`).FindIndex(dokument)
	if stelle == nil {
		t.Fatalf("kein Bildobjekt im Dokument:\n%s", ersteZeilen(dokument))
	}

	ende := bytes.Index(dokument[stelle[1]:], []byte("\nendstream"))
	if ende < 0 {
		t.Fatal("der Bildstrom endet nicht")
	}

	leser, err := zlib.NewReader(bytes.NewReader(dokument[stelle[1] : stelle[1]+ende]))
	if err != nil {
		t.Fatalf("der Bildstrom ist nicht gepackt: %v", err)
	}

	defer func() { _ = leser.Close() }() //nolint:errcheck // gelesen ist gelesen

	roh, err := io.ReadAll(leser)
	if err != nil {
		t.Fatalf("auspacken: %v", err)
	}

	return roh
}

// ersteZeilen kürzt ein Dokument für eine Fehlermeldung.
func ersteZeilen(dokument []byte) string {
	if len(dokument) > 400 {
		return string(dokument[:400]) + " …"
	}

	return string(dokument)
}
