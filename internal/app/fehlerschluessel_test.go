// Der Vertrag zwischen den Fehlerschlüsseln in Go und den Sprachkatalogen
// (AGENTS.md §9).
//
// Er hält **drei** Zusagen, und jede schließt eine andere Lücke:
//
//  1. Was Go setzt, steht in beiden Katalogen. Fehlt es, erscheint in einer
//     englischen Sitzung ein deutscher Satz — und niemand merkt es, weil eine
//     unübersetzte Meldung wie eine Meldung aussieht.
//  2. Was in einem Katalog steht, setzt Go. Sonst wächst dort Text, den nie
//     jemand sieht, und die nächste Übersetzung pflegt ihn mit.
//  3. Jede `Schluessel`-Konstante in `internal/domain` ist angemeldet. Die
//     Anmeldeliste ist von Hand geführt — Go gibt die Konstanten eines Pakets
//     zur Laufzeit nicht her —, und eine von Hand geführte Liste bleibt zurück.
//
// # Was er nicht prüft
//
// Ob die Übersetzung stimmt. Das kann er nicht, und er soll es nicht
// vortäuschen: `frontend/src/sprache/sprache.test.tsx` weist eine englische
// Zeile ab, die den deutschen Satz wiederholt — mehr ist maschinell nicht zu
// haben.
package app_test

import (
	"io/fs"
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

const wurzel = "../.."

// katalogdateien sind die beiden Fehlerkataloge des Frontends.
var katalogdateien = map[string]string{
	"de": "frontend/src/sprache/de/fehler.ts",
	"en": "frontend/src/sprache/en/fehler.ts",
}

// katalogeintrag findet `schluessel:` am Zeilenanfang einer Katalogdatei.
var katalogeintrag = regexp.MustCompile(`(?m)^ {2}([A-Za-z][A-Za-z0-9]*):(?: '|$)`)

// schluesselkonstante findet eine Konstante vom Typ Schluessel im Quelltext.
var schluesselkonstante = regexp.MustCompile(`(?m)^\s*(\w+)\s+Schluessel\s*=\s*"([^"]+)"`)

func TestJederGoSchluesselStehtInBeidenKatalogen(t *testing.T) {
	for sprache, datei := range katalogdateien {
		imKatalog := katalogschluessel(t, datei)

		for _, schluessel := range domain.Fehlerschluessel() {
			t.Run(sprache+"/"+string(schluessel), func(t *testing.T) {
				if !slices.Contains(imKatalog, string(schluessel)) {
					t.Errorf("Go setzt %q, %s kennt ihn nicht", schluessel, datei)
				}
			})
		}
	}
}

func TestJederKatalogschluesselWirdVonGoGesetzt(t *testing.T) {
	angemeldet := make([]string, 0, len(domain.Fehlerschluessel()))
	for _, schluessel := range domain.Fehlerschluessel() {
		angemeldet = append(angemeldet, string(schluessel))
	}

	for sprache, datei := range katalogdateien {
		for _, schluessel := range katalogschluessel(t, datei) {
			t.Run(sprache+"/"+schluessel, func(t *testing.T) {
				if !slices.Contains(angemeldet, schluessel) {
					t.Errorf("%s führt %q, Go setzt ihn nicht — toter Text", datei, schluessel)
				}
			})
		}
	}
}

// TestJedeSchluesselkonstanteIstAngemeldet liest den Quelltext von
// internal/domain nach.
//
// Die Anmeldeliste `alleSchluessel` ist von Hand geführt, und eine von Hand
// geführte Liste bleibt zurück: wer eine Konstante ergänzt und die Zeile
// darunter vergisst, hat einen Schlüssel, den kein Vertrag prüft.
func TestJedeSchluesselkonstanteIstAngemeldet(t *testing.T) {
	angemeldet := domain.Fehlerschluessel()

	// **Die Schlüssel stehen seit dem 2026-09-06 in einer eigenen Datei** —
	// `fehler.go` daneben führt die Fehler**arten**. Der Schnitt folgt der
	// Frage; die Zeilengrenze aus §2 war der Anlass (Werkstück 14 von Phase
	// 4.12).
	//
	// **Und seit F05 in mehreren** (#2344): die Lagen des Starts stehen in
	// `fehlerschluesselstart.go`. Gelesen wird jede Datei des Namens — eine,
	// die hier fehlte, verlöre ihre Schlüssel still aus diesem Vertrag.
	dateien, _ := fs.Glob(os.DirFS(wurzel), "internal/domain/fehlerschluessel*.go") //nolint:errcheck // ein festes Muster

	treffer := make([][]string, 0, len(dateien))
	for _, datei := range dateien {
		treffer = append(treffer, schluesselkonstante.FindAllStringSubmatch(string(projektdatei(t, datei)), -1)...)
	}

	if len(dateien) < 2 || len(treffer) == 0 {
		t.Fatalf("in %v keine Schluessel-Konstante gefunden — hat sich die Schreibweise geändert?", dateien)
	}

	for _, fund := range treffer {
		name, wert := fund[1], domain.Schluessel(fund[2])

		t.Run(name, func(t *testing.T) {
			if !slices.Contains(angemeldet, wert) {
				t.Errorf("die Konstante %s = %q steht nicht in alleSchluessel", name, wert)
			}
		})
	}
}

// katalogschluessel liest die Schlüssel einer Katalogdatei.
// **Nur die flachen Einträge.** `fehler.skript` ist ein **Unterkatalog** — die
// Sätze des Zerteilers
// (#1178) —
// und kein Fehlerschlüssel. Er wird von
// `internal/werkzeuge/skriptmeldungen_test.go` gehalten, und zwar gegen
// `msescript.Meldungsschluesselliste`; ihn hier mitzuzählen hieße, ihn gegen
// die falsche Liste zu halten.
//
// **Erkannt daran, was hinter dem Doppelpunkt steht**: ein flacher Eintrag
// trägt eine Zeichenkette — auf derselben Zeile oder, wenn `prettier`
// umgebrochen hat, gar nichts —, ein Unterkatalog eine geschweifte Klammer.
func katalogschluessel(t *testing.T, datei string) []string {
	t.Helper()

	inhalt := projektdatei(t, datei)

	treffer := katalogeintrag.FindAllStringSubmatch(string(inhalt), -1)
	if len(treffer) == 0 {
		t.Fatalf("%s enthält keinen Eintrag — hat sich der Aufbau geändert?", datei)
	}

	schluessel := make([]string, 0, len(treffer))
	for _, fund := range treffer {
		schluessel = append(schluessel, strings.TrimSpace(fund[1]))
	}

	return schluessel
}

func projektdatei(t *testing.T, pfad string) []byte {
	t.Helper()

	inhalt, err := fs.ReadFile(os.DirFS(wurzel), pfad)
	if err != nil {
		t.Fatalf("%s nicht lesbar: %v", pfad, err)
	}

	return inhalt
}
