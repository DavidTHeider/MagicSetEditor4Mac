package main

// **Jeder Befehl steht in der Hilfe, und jede Hilfezeile ist ein Befehl** —
// LOC-03.AK-04
// (#1689):
// „CLI-Beschriftungen werden kürzer und handlungsorientiert, mit tatsächlich
// geprüften Beispielen."
//
// # Wozu dieser Wächter da ist
//
// **Weil ein Befehl, den niemand findet, keiner ist.** Beim ersten Lauf
// fehlten **zwei** in der Hilfe: `partien` und `scryfall-bild` — beide seit
// Monaten da, beide über die Kommandozeile erreichbar, und beide nur zu
// finden, wenn man den Quelltext liest.
//
// **Und die Gegenrichtung ist genauso wichtig**: eine Hilfe, die einen
// Befehl nennt, den es nicht gibt, schickt jemanden auf einen Weg, der mit
// „unbekannter Befehl" endet.
//
// # Was er nicht hält
//
// **Ob die Beschreibung trifft.** Er vergleicht Namen, keine Sätze — dass
// hinter `partien` das Richtige steht, sieht nur ein Mensch. Und er prüft
// keine **Beispiele**: die Hilfe nennt Platzhalter wie `<set>`, und ein Lauf,
// der sie ausführte, prüfte den Platzhalter und nicht den Befehl.

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// hilfezeile trifft den Befehlsnamen am Anfang einer Hilfezeile.
//
// **Zwei Leerzeichen und dann ein Kleinbuchstabe** — so ist die Hilfe
// gesetzt, und die Beschreibung steht eingerückt daneben.
var hilfezeile = regexp.MustCompile(`(?m)^ {2}([a-z][a-z0-9-]*)`)

// verteiler trifft die Helfer, an die `lauf.go` weiterreicht:
// `if code, getroffen := sammlungsbefehle(…)`.
//
// **Gesucht und nicht aufgezählt.** Der erste Anlauf zählte vier Dateien auf
// und übersah `kostenjournal.go` — drei Befehle standen in der Hilfe und
// galten als erfunden. Eine Aufzählung veraltet beim nächsten Helfer, und
// sie veraltet **still**.
var verteiler = regexp.MustCompile(`getroffen := ([a-zA-Z]+)\(`)

// verteilungsfunktionen sind die oberste Verteilung samt ihren Helfern.
//
// **`fuehreAus` und `fuehreWeiterAus` sind eine Fallunterscheidung in zwei
// Dateien** (Zeilengrenze, §2); die Helfer beantworten je eine Gruppe mit
// einem zweiten Rückgabewert „war einer von meinen".
//
// **Und nur die oberste Ebene**: die Unterbefehle von `deck` und `format`
// stehen in ihren eigenen Dateien und haben keine eigene Zeile in der Hilfe.
func verteilungsfunktionen(t *testing.T) []string {
	t.Helper()

	treffer := verteiler.FindAllStringSubmatch(gelesene(t, "lauf.go"), -1)

	namen := make([]string, 0, len(treffer)+2)
	namen = append(namen, "fuehreAus", "fuehreWeiterAus")

	for _, einer := range treffer {
		namen = append(namen, einer[1])
	}

	if len(namen) < 5 {
		t.Fatalf("nur %d Verteiler gefunden — das Muster greift daneben",
			len(namen))
	}

	return namen
}

// befehlsfall trifft `case "name":` und `case "eins", "zwei":`.
var befehlsfall = regexp.MustCompile(`(?m)^\tcase ("[a-z0-9-]+"(?:, "[a-z0-9-]+")*):`)

// **Jeder Befehl der obersten Ebene steht in der Hilfe.**
func TestJederBefehlStehtInDerHilfe(t *testing.T) {
	t.Parallel()

	fehlend := []string{}

	for name := range befehlsnamen(t) {
		if !strings.Contains(hilfetext(t), "\n  "+name+" ") &&
			!strings.Contains(hilfetext(t), "\n  "+name+"\n") &&
			!strings.Contains(hilfetext(t), "`  "+name+" ") {
			fehlend = append(fehlend, name)
		}
	}

	if len(fehlend) > 0 {
		t.Errorf("diese Befehle gibt es und die Hilfe nennt sie nicht:\n%s",
			strings.Join(fehlend, "\n"))
	}
}

// **Und jede Hilfezeile nennt einen Befehl, den es gibt.**
func TestJedeHilfezeileNenntEinenBefehl(t *testing.T) {
	t.Parallel()

	befehle := befehlsnamen(t)

	erfunden := []string{}

	for _, treffer := range hilfezeile.FindAllStringSubmatch(hilfetext(t), -1) {
		name := treffer[1]

		// **»mse« steht als Programmname in den Beispielen** und ist kein
		// Befehl.
		if name == "mse" || befehle[name] {
			continue
		}

		erfunden = append(erfunden, name)
	}

	if len(erfunden) > 0 {
		t.Errorf("die Hilfe nennt diese Befehle, und es gibt sie nicht:\n%s",
			strings.Join(erfunden, "\n"))
	}
}

// befehlsnamen sammelt die Fälle der obersten Verteilung.
//
// **Aus beiden Hälften** (`lauf.go` und `laufweiter.go`): sie sind eine
// Fallunterscheidung in zwei Dateien, und ein Wächter, der nur eine liest,
// hält die Hälfte.
//
// **Und nur die oberste Ebene**: die Unterbefehle von `deck` und `format`
// stehen eingerückt und werden hier nicht gezählt — ihre Zeile in der Hilfe
// gehört zum Hauptbefehl.
func befehlsnamen(t *testing.T) map[string]bool {
	t.Helper()

	namen := map[string]bool{}

	for _, funktion := range verteilungsfunktionen(t) {
		for _, treffer := range befehlsfall.FindAllStringSubmatch(
			rumpfVon(t, funktion), -1) {
			for _, name := range strings.Split(treffer[1], ", ") {
				namen[strings.Trim(name, `"`)] = true
			}
		}
	}

	if len(namen) < 100 {
		t.Fatalf("nur %d Befehle gefunden — das Muster greift daneben",
			len(namen))
	}

	return namen
}

// rumpfVon gibt den Quelltext einer Funktion dieses Pakets.
//
// **Von ihrer Zeile bis zur nächsten auf Spalte eins** — Go setzt die
// schließende Klammer dorthin, und die nächste Deklaration fängt dort an.
func rumpfVon(t *testing.T, funktion string) string {
	t.Helper()

	eintraege, fehler := os.ReadDir(".")
	if fehler != nil {
		t.Fatalf("das Verzeichnis lesen: %v", fehler)
	}

	for _, eintrag := range eintraege {
		name := eintrag.Name()

		if !strings.HasSuffix(name, ".go") ||
			strings.HasSuffix(name, "_test.go") {
			continue
		}

		quelle := gelesene(t, name)

		anfang := strings.Index(quelle, "\nfunc "+funktion+"(")
		if anfang < 0 {
			continue
		}

		rest := quelle[anfang+1:]

		if ende := strings.Index(rest, "\n}\n"); ende >= 0 {
			return rest[:ende]
		}

		return rest
	}

	t.Fatalf("die Funktion %q gibt es nicht", funktion)

	return ""
}

// hilfetext gibt beide Hälften der Hilfe.
func hilfetext(t *testing.T) string {
	t.Helper()

	return gelesene(t, "hilfe.go") + gelesene(t, "hilfebetrieb.go")
}

// gelesene gibt den Inhalt einer Datei dieses Verzeichnisses.
func gelesene(t *testing.T, name string) string {
	t.Helper()

	roh, fehler := os.ReadFile(name) //nolint:gosec // eine Datei des Pakets
	if fehler != nil {
		t.Fatalf("%s lesen: %v", name, fehler)
	}

	return string(roh)
}
