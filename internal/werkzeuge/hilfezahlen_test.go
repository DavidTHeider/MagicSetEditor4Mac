// Der Wächter über die Zahlen in den Oberflächentexten.
//
// # Wofür
//
// **Eine Zahl über die Anwendung, die in einem Satz steht, veraltet still.**
// Der Hilfetext sagte am 2026-09-03: *„Das offene Set entscheidet, was die
// **acht** Setlaschen zeigen."* Es sind **dreizehn**, und es waren seit
// mehreren Phasen mehr als acht.
//
// `flaechenzahlen_test.go` daneben hält jede Flächenzahl im Repositorium —
// aber nur in `.md` und `.go`. Der Sprachkatalog ist `.ts` und stand unter
// keinem Wächter; genau dort liest ein Mensch die Zahl.
//
// # Was er nicht hält
//
// **Jede Zahl.** Er kennt die Zahlwörter für die beiden Seitenlisten aus
// `navigation.ts` und prüft die Sätze, in denen „Setlaschen" oder
// „Anwendungsseiten" vorkommt. Eine Zahl über etwas anderes — Kartenfelder,
// Dienste, Formate — käme hier durch; für die gibt es den Wächter nebenan,
// sobald sie in `.md` steht.
//
// **Und beide Sprachen.** Deutsch ist die Ausgangssprache (§9); der englische
// Katalog wird mitgeprüft, weil eine Übersetzung dieselbe Zahl trägt.
package werkzeuge_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// zahlwoerter ordnet die ausgeschriebenen Zahlen ihrem Wert zu.
//
// **Eine eigene Liste**, aus demselben Grund wie in `standderfreigabe_test.go`:
// die dort steht im selben Paket, und zwei kurze Listen sind billiger als eine
// Kante zwischen zwei Wächtern, die verschiedene Fragen stellen.
var zahlwoerter = map[string]int{
	"eine": 1, "zwei": 2, "drei": 3, "vier": 4, "fünf": 5, "sechs": 6,
	"sieben": 7, "acht": 8, "neun": 9, "zehn": 10, "elf": 11, "zwölf": 12,
	"dreizehn": 13, "vierzehn": 14, "fünfzehn": 15, "sechzehn": 16,
	"siebzehn": 17, "achtzehn": 18, "neunzehn": 19, "zwanzig": 20,
	"einundzwanzig": 21, "zweiundzwanzig": 22, "dreiundzwanzig": 23,
	"one": 1, "two": 2, "three": 3, "four": 4, "five": 5, "six": 6,
	"seven": 7, "eight": 8, "nine": 9, "ten": 10, "eleven": 11, "twelve": 12,
	"thirteen": 13, "fourteen": 14, "fifteen": 15, "sixteen": 16,
	"seventeen": 17, "eighteen": 18, "nineteen": 19, "twenty": 20,
}

// zahlVorDemWort findet „acht Setlaschen" oder „13 Setlaschen".
func zahlVorDemWort(wort string) *regexp.Regexp {
	return regexp.MustCompile(`(?i)([\p{L}]+|\d+) ` + wort)
}

func TestDieHilfeNenntDieRichtigeSeitenzahl(t *testing.T) {
	navigation := gelesen(t, filepath.Join("..", "..", "frontend", "src", "navigation.ts"))

	gezaehlt := map[string]int{
		"Setlaschen":       len(schluesselAus(navigation, "export const setschluessel = [")),
		"set tabs":         len(schluesselAus(navigation, "export const setschluessel = [")),
		"Anwendungsseiten": len(schluesselAus(navigation, "export const anwendungsschluessel = [")),
		"app pages":        len(schluesselAus(navigation, "export const anwendungsschluessel = [")),
	}

	for wort, soll := range gezaehlt {
		if soll < 5 {
			t.Fatalf("%q zählt %d — der Leser greift ins Leere", wort, soll)
		}
	}

	for _, datei := range katalogdateien(t) {
		inhalt := gelesen(t, datei)

		for wort, soll := range gezaehlt {
			for _, treffer := range zahlVorDemWort(wort).FindAllStringSubmatch(inhalt, -1) {
				genannt, ist := zahlwoerter[strings.ToLower(treffer[1])]
				if !ist {
					// „die Setlaschen", „alle Setlaschen" — kein Zahlwort,
					// keine Behauptung.
					continue
				}

				if genannt != soll {
					t.Errorf("%s sagt %q — es sind %d (navigation.ts)",
						datei, treffer[0], soll)
				}
			}
		}
	}
}

// katalogdateien sammelt die `.ts`-Dateien beider Sprachkataloge.
func katalogdateien(t *testing.T) []string {
	t.Helper()

	var dateien []string

	for _, sprache := range []string{"de", "en"} {
		ordner := filepath.Join("..", "..", "frontend", "src", "sprache", sprache)

		eintraege, err := os.ReadDir(ordner)
		if err != nil {
			t.Fatalf("%s lesen: %v", ordner, err)
		}

		for _, eintrag := range eintraege {
			if strings.HasSuffix(eintrag.Name(), ".ts") {
				dateien = append(dateien, filepath.Join(ordner, eintrag.Name()))
			}
		}
	}

	if len(dateien) < 10 {
		t.Fatalf("nur %d Katalogdateien gefunden — der Leser greift ins Leere",
			len(dateien))
	}

	return dateien
}
