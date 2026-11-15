package werkzeuge

// **Die Wegetafel der Schalter ist aus den zwei Katalogen geschrieben** —
// OPS-10.AK-02 (#1908).
//
// # Was er hält
//
// - **Jede Fähigkeit steht in `docs/SCHALTER.md`**, und jede dort in
//   `docs/FAEHIGKEITEN.md` — eine neue Fähigkeit erzwingt die Entscheidung,
//   an welchem Schalter sie hängt, statt still zum Kern zu werden.
// - **Nur Schalter aus dem Verzeichnis**, und Bereiche nur in der Spalte
//   Bereich, Funktionen nur in der Spalte Funktion.
// - **Das Verzeichnis in SCHALTER.md** gleicht `domain.Schalterverzeichnis`.
// - **`internal/domain/schalterwege.json` ist, was beide ergeben.**
//   `SCHALTERWEGE_SCHREIBEN=1` schreibt sie neu.
//
// # Was er nicht hält
//
// **Ob eine Fähigkeit am richtigen Schalter hängt** — das ist eine
// Entscheidung, und sie steht in der Tabelle. Und ob eine Fläche die Tafel
// fragt: das halten die Läufe je Fläche.

import (
	"bytes"
	"encoding/json"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

const (
	schalterkatalog = "../../docs/SCHALTER.md"
	wegetafeldatei  = "../domain/schalterwege.json"
)

// inBackticks fängt jeden Weg einer Zelle — Bindung, Werkzeug, Befehl,
// Endpunkt.
var inBackticks = regexp.MustCompile("`([^`]+)`")

// faehigkeitswege liest je Fähigkeit ihre Wege an den vier Flächen.
func faehigkeitswege(t *testing.T) (namen []string, wege map[string]map[domain.Flaeche][]string) {
	t.Helper()

	wege = map[string]map[domain.Flaeche][]string{}
	flaechen := []domain.Flaeche{domain.FlaecheFenster, domain.FlaecheMCP, domain.FlaecheCLI, domain.FlaecheREST}

	for _, zeile := range faehigkeitszeilen(t) {
		spalten := strings.Split(zeile, "|")
		name := ohneBetonung(spalten[1])
		je := map[domain.Flaeche][]string{}

		for i, flaeche := range flaechen {
			for _, treffer := range inBackticks.FindAllStringSubmatch(spalten[2+i], -1) {
				je[flaeche] = append(je[flaeche], treffer[1])
			}
		}

		namen = append(namen, name)
		wege[name] = je
	}

	return namen, wege
}

// ohneBetonung nimmt die Hervorhebung aus einem Namen.
func ohneBetonung(zelle string) string {
	return strings.ReplaceAll(strings.TrimSpace(zelle), "**", "")
}

// schalterzeilen liest je Fähigkeit ihre Schalter aus SCHALTER.md.
func schalterzeilen(t *testing.T) map[string][]string {
	t.Helper()

	roh, fehler := os.ReadFile(schalterkatalog)
	if fehler != nil {
		t.Fatalf("%s lesen: %v", schalterkatalog, fehler)
	}

	art := map[string]domain.Schalterart{}
	for _, schalter := range domain.Schalterverzeichnis() {
		art[schalter.Kennung] = schalter.Art
	}

	zeilen := map[string][]string{}

	for _, zeile := range strings.Split(string(roh), "\n") {
		spalten := strings.Split(zeile, "|")
		if !strings.HasPrefix(zeile, "| ") || len(spalten) != 5 ||
			strings.TrimSpace(spalten[1]) == "Fähigkeit" || strings.Contains(zeile, "---") {
			continue
		}

		name := ohneBetonung(spalten[1])
		bereich, funktion := strings.TrimSpace(spalten[2]), strings.TrimSpace(spalten[3])

		if _, doppelt := zeilen[name]; doppelt {
			t.Errorf("%q steht zweimal in %s", name, schalterkatalog)
		}

		schalter := []string{}

		if bereich != "kern" {
			if art[bereich] != domain.SchalterBereich {
				t.Errorf("%q: %q ist kein Bereich", name, bereich)
			}

			schalter = append(schalter, bereich)
		}

		if funktion != "—" {
			if art[funktion] != domain.SchalterFunktion {
				t.Errorf("%q: %q ist keine Funktion", name, funktion)
			}

			schalter = append(schalter, funktion)
		}

		zeilen[name] = schalter
	}

	if len(zeilen) == 0 {
		t.Fatalf("keine Zeile in %s — der Leser greift daneben", schalterkatalog)
	}

	return zeilen
}

// TestJedeFaehigkeitHatIhrenSchalter hält beide Kataloge deckungsgleich.
func TestJedeFaehigkeitHatIhrenSchalter(t *testing.T) {
	namen, _ := faehigkeitswege(t)
	schalter := schalterzeilen(t)

	for _, name := range namen {
		if _, steht := schalter[name]; !steht {
			t.Errorf("%q fehlt in %s — an welchem Schalter hängt sie, oder ist sie Kern?", name, schalterkatalog)
		}
	}

	bekannt := map[string]bool{}
	for _, name := range namen {
		bekannt[name] = true
	}

	for name := range schalter {
		if !bekannt[name] {
			t.Errorf("%q steht in %s, aber nicht in FAEHIGKEITEN.md", name, schalterkatalog)
		}
	}
}

// schalterverzeichniszeile trifft eine Zeile des Verzeichnisses.
var schalterverzeichniszeile = regexp.MustCompile("^\\| `([a-z]+)` \\| (Bereich|Funktion) \\| ([a-z]+) \\| (an|aus) \\|")

// TestDasVerzeichnisImKatalogGleichtDerDomaene hält die erste Tabelle.
func TestDasVerzeichnisImKatalogGleichtDerDomaene(t *testing.T) {
	roh, fehler := os.ReadFile(schalterkatalog)
	if fehler != nil {
		t.Fatalf("%s lesen: %v", schalterkatalog, fehler)
	}

	var gelesen []domain.Schalter

	for _, zeile := range strings.Split(string(roh), "\n") {
		if teile := schalterverzeichniszeile.FindStringSubmatch(zeile); teile != nil {
			gelesen = append(gelesen, domain.Schalter{
				Kennung: teile[1],
				Art:     domain.Schalterart(strings.ToLower(teile[2])),
				Reife:   domain.Reife(teile[3]),
				Vorgabe: teile[4] == "an",
			})
		}
	}

	erwartet := domain.Schalterverzeichnis()
	if len(gelesen) != len(erwartet) {
		t.Fatalf("%d Schalter in %s, %d im Verzeichnis", len(gelesen), schalterkatalog, len(erwartet))
	}

	for i := range erwartet {
		if gelesen[i] != erwartet[i] {
			t.Errorf("Zeile %d: %+v, das Verzeichnis sagt %+v", i+1, gelesen[i], erwartet[i])
		}
	}
}

// TestDieWegetafelIstAusDenKatalogenGeschrieben hält die eingebettete Tafel.
func TestDieWegetafelIstAusDenKatalogenGeschrieben(t *testing.T) {
	namen, wege := faehigkeitswege(t)
	schalter := schalterzeilen(t)

	tafel := domain.Wegetafel{}
	kern := map[domain.Flaeche]map[string]bool{}

	for _, name := range namen {
		for flaeche, liste := range wege[name] {
			if tafel[flaeche] == nil {
				tafel[flaeche], kern[flaeche] = map[string][][]string{}, map[string]bool{}
			}

			for _, weg := range liste {
				if flaeche == domain.FlaecheREST {
					weg = domain.RESTWeg(weg)
				}

				if len(schalter[name]) == 0 {
					kern[flaeche][weg] = true

					continue
				}

				tafel[flaeche][weg] = mitSchaltern(tafel[flaeche][weg], schalter[name])
			}
		}
	}

	for flaeche, gekernt := range kern {
		for weg := range gekernt {
			delete(tafel[flaeche], weg)
		}
	}

	geschrieben, fehler := json.MarshalIndent(tafel, "", "\t")
	if fehler != nil {
		t.Fatal(fehler)
	}

	geschrieben = append(geschrieben, '\n')

	if os.Getenv("SCHALTERWEGE_SCHREIBEN") == "1" {
		if schreibfehler := os.WriteFile(wegetafeldatei, geschrieben, 0o600); schreibfehler != nil {
			t.Fatal(schreibfehler)
		}
	}

	vorhanden, fehler := os.ReadFile(wegetafeldatei)
	if fehler != nil {
		t.Fatalf("%s lesen: %v", wegetafeldatei, fehler)
	}

	if !bytes.Equal(vorhanden, geschrieben) {
		t.Errorf("%s ist nicht, was FAEHIGKEITEN.md und SCHALTER.md ergeben — "+
			"SCHALTERWEGE_SCHREIBEN=1 schreibt sie neu", wegetafeldatei)
	}
}

// mitSchaltern hängt eine Schalterliste an, wenn sie noch nicht dasteht.
func mitSchaltern(liste [][]string, schalter []string) [][]string {
	for _, vorhanden := range liste {
		if strings.Join(vorhanden, ",") == strings.Join(schalter, ",") {
			return liste
		}
	}

	return append(liste, schalter)
}
