package rest_test

import (
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

// imWeg trifft einen eingetragenen Endpunkt: `weg.HandleFunc("GET /sets", …)`.
//
// **Auch über einen Zeilenumbruch hinweg.** Der erste Anlauf verlangte das
// Muster in **einer** Zeile — und `gofmt` bricht einen langen Aufruf um. Der
// Nachlauf zum Zweitblick über Phase U hat es nachgestellt: ein
//
//	weg.HandleFunc(
//		"DELETE /sets/{setID}",
//		d.loeschen,
//	)
//
// ging durch **beide** Wächter, und der neue Weg hatte keine der vier Klassen.
var imWeg = regexp.MustCompile(`weg\.HandleFunc\(\s*"([A-Z]+) ([^"]+)"`)

// **Jeder Endpunkt im Quelltext steht in `alleEndpunkte`**
// (#997).
//
// # Wozu dieser Wächter da ist
//
// [0046](../../docs/entscheidungen/0046-doch-ein-dienst-auf-der-rueckschleife.md)
// nennt als Bedingung 2: wird die Regel „vier Testklassen je Endpunkt" einmal
// übergangen, fällt die Entscheidung für diesen Dienst wieder. **Übergangen
// wird sie nicht durch Absicht, sondern durch Vergessen** — jemand trägt einen
// siebten Weg ein und schreibt keine Zeile in die Liste.
//
// Danach hätte dieser Weg **keine** der vier Klassen, und alles wäre grün.
//
// # Was er nicht prüft
//
// **Dass die Klassen etwas taugen.** Er zählt, dass jeder Weg in der Liste
// steht; ob der Aufruf in `gueltig` wirklich durchgeht, hält
// `TestJederEndpunktKenntSeineVierKlassen` selbst — und dass er es hält, steht
// in seiner Gegenprobe.
//
// **Einen Weg, der nicht über `weg.HandleFunc` eingetragen wird.** Wer einen
// eigenen `http.Handler` daneben hängte, käme hier durch. Dagegen steht die
// Zeilengrenze aus §2 und die Übersichtlichkeit von `wegeEintragen` — kein
// Wächter, und das ist hiermit gesagt.
//
// **Und `peerwege.go`** seit dem 2026-09-14 (N1, #1454): die Tafel ging mit
// den zehn Peerwegen über die Zeilengrenze aus §2, und ein Wächter, der nur
// eine von zwei Dateien liest, hält die Hälfte.
//
// **Gelesen wird `wegetafel.go`**, seit die Tafel dort steht (2026-08-26) —
// **und `wegeleute.go`**, seit die Wege zu Nutzern und Profilen dort stehen
// (2026-09-08, §2).
// Als sie noch in `endpunkte.go` lag und der Wächter dort suchte, war der
// Umzug ein roter Lauf — der Zweig „das Muster greift daneben" hat gehalten,
// was er zusagt.
func TestJederEndpunktStehtInDerListe(t *testing.T) {
	t.Parallel()

	quelle := allewegequellen(t)

	treffer := imWeg.FindAllStringSubmatch(quelle, -1)
	if len(treffer) == 0 {
		t.Fatal("in endpunkte.go fand sich kein Endpunkt — das Muster greift daneben")
	}

	for _, eintrag := range treffer {
		methode, muster := eintrag[1], eintrag[2]

		if !inDerListe(methode, muster) {
			t.Errorf("%s %s steht in keiner Zeile von alleEndpunkte — "+
				"ohne sie hat dieser Weg keine der vier Klassen", methode, muster)
		}
	}
}

// **Und umgekehrt: keine Zeile ohne Weg.**
//
// Ohne diese Richtung bliebe eine Zeile für einen entfernten Endpunkt stehen
// und prüfte vier Klassen an einem Weg, den es nicht mehr gibt — grün, und
// ohne Aussage.
func TestJedeZeileHatIhrenWeg(t *testing.T) {
	t.Parallel()

	quelle := allewegequellen(t)

	for _, ziel := range alleEndpunkte {
		gefunden := false

		for _, eintrag := range imWeg.FindAllStringSubmatch(quelle, -1) {
			if eintrag[1] == ziel.methode && passt(eintrag[2], ziel.weg) {
				gefunden = true

				break
			}
		}

		if !gefunden {
			t.Errorf("%q (%s %s) trifft keinen eingetragenen Weg",
				ziel.name, ziel.methode, ziel.weg)
		}
	}
}

// inDerListe sagt, ob ein Muster von einer Zeile getroffen wird.
func inDerListe(methode, muster string) bool {
	return slices.ContainsFunc(alleEndpunkte, func(ziel endpunkt) bool {
		return ziel.methode == methode && passt(muster, ziel.weg)
	})
}

// passt vergleicht ein Wegmuster mit einem konkreten Weg.
//
// **Platzhalterweise**: `/sets/{setID}/karten` trifft `/sets/1/karten`. Ein
// Vergleich auf Gleichheit ginge nicht, und einer, der nur die Länge zählte,
// hielte `/sets/1/felder` für denselben Weg.
func passt(muster, weg string) bool {
	// **Die Abfrage gehört nicht zum Weg.** `?name=…` steht in der Zeile,
	// weil ein gültiger Aufruf sie braucht — `net/http` sieht sie nicht.
	weg, _, _ = strings.Cut(weg, "?")

	teileM := strings.Split(strings.Trim(muster, "/"), "/")
	teileW := strings.Split(strings.Trim(weg, "/"), "/")

	if len(teileM) != len(teileW) {
		return false
	}

	for stelle, teil := range teileM {
		if strings.HasPrefix(teil, "{") {
			continue
		}

		if teil != teileW[stelle] {
			return false
		}
	}

	return true
}

// allewegequellen liest jede Datei, in der Wege eingetragen werden.
//
// **Eine Stelle und nicht zwei.** Die Aufzählung stand zweimal wörtlich da,
// und beim fünften Umzug (N4/N5: `wegedateien.go`) fiel genau das auf — eine
// von beiden hätte man vergessen können, und dann prüfte der eine Lauf mehr
// als der andere. Dass es hier bisher gutging, lag daran, dass beide im
// selben Commit geändert wurden.
//
// **Fail-closed bleibt es**: findet das Muster nichts, endet der Lauf — das
// steht bei seinen Aufrufern.
//
// **Und jede Datei, nicht eine Aufzählung** — DOM-04 (#1699). Hier stand eine
// Liste von sechs Namen; eine siebte Wegedatei (`turnierwege.go`) entging
// dem Lauf, und ihre Wege galten als ungeschrieben. Gelesen wird deshalb
// jede Quelldatei des Pakets, die einen Weg einträgt.
func allewegequellen(t *testing.T) string {
	t.Helper()

	dateien, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("die Quelldateien finden: %v", err)
	}

	var quelle string

	for _, datei := range dateien {
		if strings.HasSuffix(datei, "_test.go") {
			continue
		}

		if inhalt := lies(t, datei); imWeg.MatchString(inhalt) {
			quelle += inhalt
		}
	}

	return quelle
}
