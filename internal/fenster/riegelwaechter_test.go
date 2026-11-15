package fenster

// **Jeder Rückruf, den die Hülle bei Wails anmeldet, hat seinen Riegel** —
// F02 (#2341).
//
// # Warum ein Wächter
//
// Der Riegel hilft nur dort, wo er steht. Ein neuer Haken ohne ihn fiele in
// Wails' Panikweg zurück — und der beendet, hängt oder schützt gar nicht
// (siehe `riegel.go`). Der Wächter liest jede Anmeldung im Paket und
// verlangt `riegeln(` oder `gesichert(` in ihrem Aufruf.
//
// # Was er nicht prüft
//
// **Ob im Rückruf etwas hinter dem Riegel steht**, das selbst eine Goroutine
// startet: die fängt er nicht (für den Schluss hält `innerhalb` in
// `internal/app` das). Und Rückrufe, die Wails außerhalb dieser Namen
// anmeldet.

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// anmeldung findet die Stellen, an denen ein Rückruf bei Wails angemeldet wird.
var anmeldung = regexp.MustCompile(`\.(OnApplicationEvent|OnWindowEvent|RegisterHook|OnShutdown|OnClick)\(`)

func TestJederRueckrufHatSeinenRiegel(t *testing.T) {
	dateien, fehler := filepath.Glob("*.go")
	if fehler != nil {
		t.Fatal(fehler)
	}

	gesehen := 0

	for _, datei := range dateien {
		if strings.HasSuffix(datei, "_test.go") {
			continue
		}

		roh, lesefehler := os.ReadFile(datei) //nolint:gosec // ein Pfad aus Glob im Paket
		if lesefehler != nil {
			t.Fatal(lesefehler)
		}

		text := string(roh)

		for _, stelle := range anmeldung.FindAllStringIndex(text, -1) {
			gesehen++
			aufruf := bisZurKlammer(text[stelle[1]:])

			if !strings.Contains(aufruf, "riegeln(") && !strings.Contains(aufruf, "gesichert(") {
				zeile := strings.Count(text[:stelle[0]], "\n") + 1
				t.Errorf("%s:%d meldet einen Rückruf ohne Riegel an: %s", datei, zeile, strings.TrimSpace(text[stelle[0]:stelle[1]]))
			}
		}
	}

	if gesehen < 8 {
		t.Fatalf("nur %d Anmeldungen gefunden — der Wächter liest daneben", gesehen)
	}
}

// bisZurKlammer gibt den Text bis zur Klammer, die die geöffnete schließt.
func bisZurKlammer(text string) string {
	tiefe := 1

	for stelle, zeichen := range text {
		switch zeichen {
		case '(':
			tiefe++
		case ')':
			tiefe--
		}

		if tiefe == 0 {
			return text[:stelle]
		}
	}

	return text
}
