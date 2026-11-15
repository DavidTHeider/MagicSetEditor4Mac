// Die Fassung steht an **fünf** Orten — dieser Wächter hält den fünften.
//
// # Warum es einen fünften gibt
//
// `scripts/releaseversion.sh` hält vier zusammen: `wails.json`,
// `frontend/package.json`, die oberste datierte Überschrift in `CHANGELOG.md`
// und den Tag. Sie sind alle **Dateien**, und die Freigabe liest sie mit einem
// Shell-Werkzeug.
//
// **Der MCP-Server ist keine Datei.** Er meldet seine Fassung in der Antwort
// auf `initialize`, und sie muss im Programm stehen — ein eigenes Binärprogramm
// hat kein `Info.plist`, aus dem es sie läse. Bis zum 2026-08-23 stand dort
// fest `"0.0.0"`
// (#896):
// eine Angabe, die wie eine Fassung aussieht und keine ist.
//
// # Warum hier und nicht in `releaseversion.sh`
//
// Weil der Ort Go-Quelltext ist und die Prüfung dann eine Go-Datei mit `sed`
// zerlegte. Sie steht hier bei den übrigen Wächtern über Dateien der
// Projektwurzel — und sie läuft in **jedem** Lauf, nicht nur im Freigabelauf.
// Das ist der Unterschied, der zählt: eine Fassung, die erst beim Freigeben
// auffällt, ist ein halbes Jahr lang falsch gewesen.
//
// # Was er nicht prüft
//
//   - **Ob die Fassung die richtige ist.** Er hält zwei Orte gegeneinander.
//     Ob 0.1.0 der Stand ist, den jemand freigeben will, sagt der Mensch.
//   - **Die anderen drei Orte.** Die hält `versionen_pruefen`, und er hält
//     sie **nur im Freigabelauf**. Wer beide zusammen will, ruft
//     `./scripts/release.sh`.
//   - **Dass der MCP-Server die Konstante wirklich meldet.** Das prüft
//     `internal/mcp/vorstellung_test.go` dort, wo die Antwort entsteht.
package werkzeuge_test

import (
	"regexp"
	"testing"
)

// fassungImQuelltext findet `const Fassung = "…"`.
var fassungImQuelltext = regexp.MustCompile(`(?m)^const Fassung = "([^"]*)"`)

// fassungInWails findet die Fassung in `wails.json`.
var fassungInWails = regexp.MustCompile(`"productVersion"\s*:\s*"([^"]*)"`)

func TestDieFassungImQuelltextStimmtMitWailsUeberein(t *testing.T) {
	imCode := einzelfund(t, fassungImQuelltext, lies(t, "internal/app/fassung.go"),
		"internal/app/fassung.go nennt keine Konstante Fassung")
	inWails := einzelfund(t, fassungInWails, lies(t, "wails.json"),
		"wails.json nennt keine productVersion")

	if imCode != inWails {
		t.Errorf("internal/app/fassung.go nennt %q, wails.json nennt %q", imCode, inWails)
	}

	// **Und sie ist keine Lücke mit Ziffern.** `0.0.0` sähe wie eine Fassung
	// aus und wäre keine — genau die Angabe, die der MCP-Server bis zum
	// 2026-08-23 meldete.
	if imCode == "0.0.0" || imCode == "" {
		t.Errorf("die Fassung ist %q — das ist keine Fassung, sondern eine Lücke", imCode)
	}
}

// einzelfund gibt die erste Fanggruppe oder bricht ab.
func einzelfund(t *testing.T, muster *regexp.Regexp, text, klage string) string {
	t.Helper()

	treffer := muster.FindStringSubmatch(text)
	if treffer == nil {
		t.Fatal(klage)
	}

	return treffer[1]
}
