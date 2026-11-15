// **Jede Fähigkeit an mehreren Flächen hat einen Flächenlauf**
// (#1169).
//
// # Der Befund, aus dem dieser Wächter entstanden ist
//
// `docs/FAEHIGKEITEN.md` sagt seit dem 2026-08-23: „Dafür gibt es je Fähigkeit
// einen Lauf, der **denselben** Arbeitsablauf über jede Fläche fährt und
// dieselben fachlichen Fehler erwartet."
//
// **Es waren 51 Läufe für 59 mehrflächige Fähigkeiten** — und die Verteilung
// war nicht die, die man vermutet: 32 Fähigkeiten hatten einen Lauf, 27 gar
// keinen. Der Satz stand als Zusage da, und niemand hielt ihn nach.
//
// Das ist genau die Gestalt aus §10: der Wächter daneben zählt, was **da**
// ist — Bindungen, Werkzeuge, Endpunkte —, und die Zusage **darüber** hielt
// keiner.
//
// # Was dieser Wächter hält
//
// Für jede Zeile, die an **mindestens zwei** Flächen geführt wird: mindestens
// eine ihrer Kennungen — MCP-Werkzeug, Kommandozeilenaufruf oder
// REST-Endpunkt — kommt unter `internal/flaechen` vor.
//
// **Zwei und nicht vier.** Eine Fähigkeit, die es nur am Fenster gibt, hat
// nichts zu vergleichen; ab der zweiten Fläche ist „überall dasselbe" eine
// Aussage, und eine Aussage gehört gefahren.
//
// # Was er **nicht** hält, und das ist die Hälfte, die zählt
//
//   - **Ob der Lauf wirklich vergleicht.** Er sieht, dass eine Fläche
//     angefasst wird — nicht, dass zwei dasselbe tun. Ein Lauf, der `mse
//     preise` ruft und das Ergebnis wegwirft, macht ihn grün. Was ihn hält,
//     ist der Name des Laufs: `…UeberJedeFlaeche` ist eine Behauptung, die
//     ein Mensch liest.
//   - **Ob die Fehlerklassen dabei sind.** §16 verlangt „dieselben fachlichen
//     Fehler"; ob ein Lauf einen Fehlerfall führt, weiß nur, wer ihn liest.
//   - **Das Fenster.** Die Läufe fahren Kommandozeile, MCP und Dienst; die
//     Wails-Naht steht daneben und wird von `internal/app` geprüft. Genau
//     dort liegt die Naht, vor der §3 warnt — ein `omitempty` auf einem Slice
//     lässt eine leere Liste verschwinden, und keine der beiden Ebenen sieht
//     es. Das ist Werkstück 3 von #1169 und noch offen.
package werkzeuge

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// flaechenverzeichnis ist der Ort der Flächenläufe.
const flaechenverzeichnis = "../flaechen"

// kennungImKatalog trifft jede `Kennung` in einer Flächenspalte.
//
// **Weiter als [inBackenzeichen]**: dort geht es um Bezeichner, hier auch um
// Aufrufe mit Leerzeichen und Schrägstrichen — `mse scryfall-suche`,
// `GET /sets/{id}/felder`.
var kennungImKatalog = regexp.MustCompile("`([^`]+)`")

// TestJedeMehrflaechigeFaehigkeitHatEinenFlaechenlauf ist die Zusage aus §16.
func TestJedeMehrflaechigeFaehigkeitHatEinenFlaechenlauf(t *testing.T) {
	laeufe := flaechenlaeufe(t)

	ohne := 0

	for _, zeile := range mehrflaechigeZeilen(t) {
		if zeile.getroffen(laeufe) {
			continue
		}

		ohne++

		t.Errorf("die Fähigkeit %q hat keinen Flächenlauf — keine der Kennungen %v kommt "+
			"unter internal/flaechen vor (AGENTS.md §16)", zeile.name, zeile.kennungen)
	}

	if ohne > 0 {
		t.Logf("%d von %d mehrflächigen Fähigkeiten ohne Lauf",
			ohne, len(mehrflaechigeZeilen(t)))
	}
}

// TestJedeFlaechenkennungGehoertZuEinerFaehigkeit geht die Gegenrichtung.
//
// **Ohne sie wüchsen Läufe ohne Zeile.** Ein Werkzeug, das kein Katalogeintrag
// nennt, wäre an einer Fläche vorhanden und in der Übersicht unsichtbar —
// dieselbe Lücke wie beim Wächter nebenan, nur andersherum.
//
// **Geprüft wird über die MCP-Namen**, weil die eindeutig sind: `mse karten`
// steht auch in einem Kommentar, `GET /sets` in jeder zweiten Zeile.
func TestJedeFlaechenkennungGehoertZuEinerFaehigkeit(t *testing.T) {
	katalog := strings.Join(katalognamen(t), "\n")

	for _, werkzeug := range werkzeugeInLaeufen(t) {
		if !strings.Contains(katalog, werkzeug) {
			t.Errorf("das Werkzeug %q wird in einem Flächenlauf gerufen, "+
				"steht aber in keiner Katalogzeile", werkzeug)
		}
	}
}

// katalogzeile ist eine Fähigkeit mit ihren Kennungen ohne Fenster.
type katalogzeile struct {
	name      string
	kennungen []string
}

// getroffen sagt, ob eine der Kennungen in den Läufen vorkommt.
//
// **Ein Kommandozeilenbefehl wird als Zeichenkettenliteral gesucht.** Der
// Katalog schreibt ihn als `mse bestand-sichern` — so ruft man ihn auf —, ein
// Lauf schreibt `kommandozeile(t, umgebung, "bestand-sichern", …)`. Wer nur
// die Katalogform sucht, findet sie **im Kommentar** und nicht im Aufruf.
//
// Genau so ist es gewesen: `Eine Sicherung zurückspielen` galt als geprüft,
// weil ein Kommentar den Aufruf **erklärte** — sichtbar wurde es erst, als
// dieser Wächter am 2026-08-28 aufhörte, Kommentare mitzulesen (#1170,
// Befund 12). Der Lauf dazu gab es die ganze Zeit; belegt hat ihn dieser
// Wächter nicht.
func (k katalogzeile) getroffen(laeufe string) bool {
	for _, kennung := range k.kennungen {
		if strings.Contains(laeufe, kennung) {
			return true
		}

		if befehl, istBefehl := strings.CutPrefix(kennung, "mse "); istBefehl &&
			strings.Contains(laeufe, `"`+befehl+`"`) {
			return true
		}
	}

	return false
}

// mehrflaechigeZeilen liest die Zeilen mit mindestens zwei Flächen.
func mehrflaechigeZeilen(t *testing.T) []katalogzeile {
	t.Helper()

	roh, err := os.ReadFile(katalogpfad) //nolint:gosec // ein fester Pfad im Repositorium
	if err != nil {
		t.Fatalf("%s lesen: %v", katalogpfad, err)
	}

	var zeilen []katalogzeile

	for _, zeile := range strings.Split(string(roh), "\n") {
		if !strings.HasPrefix(zeile, "| ") || strings.Count(zeile, "|") != spaltenzahl {
			continue
		}

		spalten := strings.Split(zeile, "|")

		name := strings.TrimSpace(spalten[1])
		if name == "Fähigkeit" || name == "Spalte" || strings.Trim(name, "- :") == "" {
			continue
		}

		gefuehrt, kennungen := flaechenAusZeile(spalten)
		if gefuehrt < 2 {
			continue
		}

		zeilen = append(zeilen, katalogzeile{name: kuerze(name), kennungen: kennungen})
	}

	if len(zeilen) == 0 {
		t.Fatal("keine mehrflächige Zeile gefunden — der Leser greift daneben")
	}

	return zeilen
}

// flaechenAusZeile zählt die geführten Flächen und sammelt die Kennungen.
//
// **Das Fenster zählt mit, seine Bindungen aber nicht**: eine Bindung ist kein
// Weg, den ein Flächenlauf gehen kann — er baut `cmd/mse` und `cmd/mse-mcp`
// und fährt den Dienst, nicht das Fenster.
func flaechenAusZeile(spalten []string) (gefuehrt int, kennungen []string) {
	for stelle, spalte := range spalten[2:6] {
		inhalt := strings.TrimSpace(spalte)
		if inhalt == "" || inhalt == "—" {
			continue
		}

		gefuehrt++

		// stelle 0 ist das Fenster.
		if stelle == 0 {
			continue
		}

		for _, treffer := range kennungImKatalog.FindAllStringSubmatch(inhalt, -1) {
			if kennung := strings.TrimSpace(treffer[1]); kennung != "" {
				kennungen = append(kennungen, kennung)
			}
		}
	}

	return gefuehrt, kennungen
}

// kuerze macht aus einer Tabellenzelle einen lesbaren Namen.
func kuerze(name string) string {
	name = strings.NewReplacer("`", "", "*", "", "[", "", "]", "").Replace(name)

	if schnitt := strings.Index(name, " —"); schnitt > 0 {
		name = name[:schnitt]
	}

	return strings.TrimSpace(name)
}

// flaechenlaeufe gibt den Quelltext unter `internal/flaechen` — **ohne
// Kommentare**.
//
// # Warum ohne
//
// Er fügte bis zum 2026-08-28 den vollständigen Quelltext samt Kommentaren
// zusammen und suchte darin nach einer Kennung. **Eine Kennung in einen
// Kommentar zu schreiben machte den Lauf grün** — ohne dass irgendetwas
// gefahren wäre. Angestrichen vom zweiten Zweitblick über die Freigabe 0.3.0
// (#1170,
// Befund 12).
//
// Der Fall ist nicht hypothetisch: die Läufe **erklären** sich ausführlich,
// und jede dieser Erklärungen nennt Kennungen.
//
// # Was er weiterhin **nicht** hält
//
// Dass ein Lauf, der eine Kennung nennt, sie auch **fährt**. Eine Kennung in
// einer ungenutzten Konstanten genügt ihm, und ob zwei Flächen dasselbe
// antworten, sieht er ohnehin nicht — dafür gibt es je Fähigkeit einen
// eigenen Lauf. Das steht schon im Kopf dieser Datei und bleibt wahr.
func flaechenlaeufe(t *testing.T) string {
	t.Helper()

	var teile []string

	fehler := filepath.WalkDir(flaechenverzeichnis,
		func(pfad string, eintrag os.DirEntry, err error) error {
			if err != nil || eintrag.IsDir() || !strings.HasSuffix(pfad, ".go") {
				return err
			}

			inhalt, lesefehler := os.ReadFile(pfad) //nolint:gosec // Quelldatei im Repositorium
			if lesefehler != nil {
				return lesefehler
			}

			inhalt = []byte(ohneGoKommentare(string(inhalt)))

			teile = append(teile, string(inhalt))

			return nil
		})
	if fehler != nil {
		t.Fatalf("%s durchgehen: %v", flaechenverzeichnis, fehler)
	}

	if len(teile) == 0 {
		t.Fatal("unter internal/flaechen liegt keine Go-Datei — der Leser greift daneben")
	}

	return strings.Join(teile, "\n")
}

// werkzeugeInLaeufen gibt die MCP-Werkzeuge, die ein Lauf ruft.
func werkzeugeInLaeufen(t *testing.T) []string {
	t.Helper()

	var namen []string

	for _, treffer := range mcpwerkzeug.FindAllStringSubmatch(flaechenlaeufe(t), -1) {
		if !slicesEnthaelt(namen, treffer[1]) {
			namen = append(namen, treffer[1])
		}
	}

	if len(namen) == 0 {
		t.Fatal("in den Flächenläufen steht kein Werkzeugname — das Muster greift daneben")
	}

	return namen
}

// slicesEnthaelt sagt, ob ein Name schon dabei ist.
func slicesEnthaelt(namen []string, gesucht string) bool {
	for _, name := range namen {
		if name == gesucht {
			return true
		}
	}

	return false
}

// kommentare findet Zeilen- und Blockkommentare in Go.
//
// **Denselben Befund gab es schon einmal** — der Zweitblick über Phase L fand
// ihn für Paketdateien, und daneben steht seither `ohneKommentare` in
// `saatbedienelemente_test.go` mit dem Satz "sonst deckt eine Erwähnung die
// Wirkung". Dieser Wächter hat ihn wiederholt.
//
// **Zeichenketten bleiben stehen.** Ein `//` in einer Zeichenkette käme hier
// als Kommentar durch — das ist die Ungenauigkeit, die dieser Ausdruck sich
// leistet, und sie geht in die **sichere** Richtung: sie entfernt zu viel,
// nicht zu wenig. Ein Lauf, dessen Kennung in einer Zeichenkette hinter `//`
// steht, machte den Wächter rot statt grün.
var kommentare = regexp.MustCompile(`(?s)/\*.*?\*/|//[^\n]*`)

// ohneGoKommentare gibt den Quelltext ohne seine Erklärungen.
func ohneGoKommentare(quelltext string) string {
	return kommentare.ReplaceAllString(quelltext, "")
}
