// **Die REST-Spalte des Fähigkeitskatalogs, gegen den Quelltext gehalten**
// (#997).
//
// **Eine eigene Datei, weil sie eine eigene Frage beantwortet** (§2) — und
// weil `faehigkeiten_test.go` sonst über die Zeilengrenze liefe.
package werkzeuge

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

// restweg trifft einen Eintrag der REST-Spalte: `GET /sets`.
var restweg = regexp.MustCompile("`([A-Z]+) (/[^`]*)`")

// imDienst trifft einen eingetragenen Endpunkt in `internal/rest/endpunkte.go`.
//
// **Auch über einen Zeilenumbruch hinweg** — `gofmt` bricht einen langen
// Aufruf um, und ein umbrochener ging durch (Nachlauf zum Zweitblick über
// Phase U).
var imDienst = regexp.MustCompile(`weg\.HandleFunc\(\s*"([A-Z]+) ([^"]+)"`)

// **Jeder Endpunkt des Dienstes steht im Katalog — und umgekehrt**
// (#997).
//
// # Warum das dieselbe Frage ist wie bei Bindungen und Werkzeugen
//
// Ohne diese Prüfung wäre die REST-Spalte **Prosa**: jemand trägt einen
// siebten Endpunkt ein, schreibt keine Zeile, und der Katalog behauptet
// weiterhin sechs. Genau das war die Lage vor Phase T — nur mit MCP.
//
// # Was er nicht hält
//
// **Ob die Zuordnung fachlich stimmt.** Dass `POST /sets` dieselbe Fähigkeit
// ist wie `SetOeffnen` und `schreiben.set.einfahren`, ist eine Aussage über
// das, was sie tun.
//
// **Die Platzhalter.** Der Katalog schreibt `/sets/{id}/karten`, der Quelltext
// `/sets/{setID}/karten`. Verglichen wird deshalb **stellenweise**, und ein
// Platzhalter trifft jeden anderen. Wer zwei verschiedene Wege mit gleich
// vielen Stellen und gleichen festen Teilen hätte, käme hier durch — den gibt
// es nicht, und das ist hiermit gesagt.
func TestJederRestwegStehtImKatalog(t *testing.T) {
	imQuelltext := restwegeAusDienst(t)
	imKatalog := restwegeAusKatalog(t)

	for _, weg := range imQuelltext {
		if !trifftEinen(weg, imKatalog) {
			t.Errorf("der Endpunkt %q steht in keiner Fähigkeit — zu welcher gehört er? (%s)",
				weg, katalogpfad)
		}
	}

	for _, weg := range imKatalog {
		if !trifftEinen(weg, imQuelltext) {
			t.Errorf("der Katalog nennt %q — den Endpunkt gibt es nicht", weg)
		}
	}
}

// restwegeAusDienst liest die Endpunkte aus dem Quelltext.
func restwegeAusDienst(t *testing.T) []string {
	t.Helper()

	// **Jede Quelldatei des Dienstes, die Wege einträgt** — und keine
	// Aufzählung. Hier standen sechs Namen, und jeder Umzug der Tafel (Leute,
	// Peers, ganze Dateien) verlangte eine Zeile mehr; die siebte Datei
	// (`turnierwege.go`, DOM-04, #1699) fehlte, und der Katalog nannte elf
	// Wege, die es angeblich nicht gab. Eine vergessene Datei wäre in der
	// Gegenrichtung still gewesen: ihre Wege stünden ungeprüft da.
	dateien, err := filepath.Glob(filepath.Join(wurzel, "internal", "rest", "*.go"))
	if err != nil {
		t.Fatalf("die Quelldateien des Dienstes finden: %v", err)
	}

	var roh []byte

	for _, datei := range dateien {
		if strings.HasSuffix(datei, "_test.go") {
			continue
		}

		teil, err := os.ReadFile(datei) //nolint:gosec // eine Quelldatei im Repositorium
		if err != nil {
			t.Fatalf("%s lesen: %v", datei, err)
		}

		roh = append(roh, teil...)
	}

	treffer := imDienst.FindAllStringSubmatch(string(roh), -1)

	wege := make([]string, 0, len(treffer))
	for _, eintrag := range treffer {
		wege = append(wege, eintrag[1]+" "+eintrag[2])
	}

	if len(wege) == 0 {
		t.Fatal("in endpunkte.go fand sich kein Endpunkt — das Muster greift daneben")
	}

	return wege
}

// restwegeAusKatalog liest die REST-Spalte.
func restwegeAusKatalog(t *testing.T) []string {
	t.Helper()

	var wege []string

	for _, zeile := range faehigkeitszeilen(t) {
		spalten := strings.Split(zeile, "|")
		for _, treffer := range restweg.FindAllStringSubmatch(spalten[5], -1) {
			wege = append(wege, treffer[1]+" "+treffer[2])
		}
	}

	if len(wege) == 0 {
		t.Fatal("in der REST-Spalte fand sich nichts — der Leser greift daneben")
	}

	return wege
}

// trifftEinen sagt, ob ein Weg von einem der anderen getroffen wird.
func trifftEinen(weg string, andere []string) bool {
	return slices.ContainsFunc(andere, func(kandidat string) bool {
		return gleicherWeg(weg, kandidat)
	})
}

// gleicherWeg vergleicht zwei Wegmuster stellenweise.
//
// **Ein Platzhalter trifft jeden anderen Teil.** `/sets/{id}/karten` und
// `/sets/{setID}/karten` sind derselbe Weg; die Namen dürfen sich
// unterscheiden, weil der Katalog für Menschen geschrieben ist und der
// Quelltext für `net/http`.
func gleicherWeg(links, rechts string) bool {
	teileL := strings.Split(links, "/")
	teileR := strings.Split(rechts, "/")

	if len(teileL) != len(teileR) {
		return false
	}

	for stelle, teil := range teileL {
		if strings.HasPrefix(teil, "{") || strings.HasPrefix(teileR[stelle], "{") {
			continue
		}

		if teil != teileR[stelle] {
			return false
		}
	}

	return true
}
