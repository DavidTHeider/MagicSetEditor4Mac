// Der Fähigkeitskatalog bleibt vollständig
// (#886).
//
// # Was dieser Wächter hält
//
// **Die Vollständigkeit in beide Richtungen.** Jede Methode auf `*Bindungen`
// steht in `docs/FAEHIGKEITEN.md` bei genau einer Fähigkeit, und jeder dort
// genannte Name ist eine solche Methode. Dasselbe für die MCP-Werkzeuge.
//
// **Eine neue Bindung erzwingt damit eine Entscheidung**: gehört sie zu einer
// vorhandenen Fähigkeit, oder ist sie eine neue? Ohne ihn verschwände sie
// still — und die Frage, was welcher Fläche fehlt, wäre wieder
// unbeantwortbar. Genau das war der Anlass, am 2026-08-23: 67 Bindungen, 9
// Werkzeuge, keine Kommandozeile, und kein Name für das, was gekonnt wird.
//
// # Was er nicht hält, und das steht auch in der Datei
//
//   - **Ob die Zuordnung fachlich stimmt.** Dass `Zuschnittgrenze` und
//     `Zuschnittziehen` **eine** Fähigkeit sind und nicht zwei, ist eine
//     Aussage über das, was sie tun. Ein Namensmuster ist es nicht.
//   - **Ob die Spalte „Dialog" stimmt.** Sie ist gemessen und nicht abgeleitet;
//     wer eine Bindung umbaut, muss sie neu setzen.
//   - **Ob eine Fähigkeit an einer Fläche fehlen darf.** `Symbolpalette` auf
//     der Kommandozeile wäre eine Antwort auf keine Frage.
package werkzeuge

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// katalogpfad ist der Fähigkeitskatalog.
const katalogpfad = "../../docs/FAEHIGKEITEN.md"

// spaltenzahl ist die Zahl der Striche einer Fähigkeitszeile.
//
// **Sechs Spalten heißen sieben Striche** — der erste und der letzte begrenzen.
// Die Zahl steht hier, weil sie an drei Stellen gebraucht wird und beim
// Hinzufügen einer fünften Fläche an allen dreien stimmen muss.
const spaltenzahl = 7

// inBackenzeichen trifft jeden `Bezeichner` in der Datei.
var inBackenzeichen = regexp.MustCompile("`([A-Za-z0-9_.]+)`")

// mcpwerkzeug trifft einen Werkzeugnamen aus `internal/mcp`.
var mcpwerkzeug = regexp.MustCompile(`"((?:lesen|schreiben)\.[a-z][a-z0-9.]*)"`)

// katalognamen sind die Bezeichner aus den **Tabellenzeilen** des Katalogs.
//
// **Nur aus den Tabellen, nicht aus dem Fließtext.** Der erste Anlauf las die
// ganze Datei, und der Wächter forderte eine Bindung namens `App` — der
// Erklärtext nennt sie in Backenzeichen. Ein Wächter, der über Prosa
// stolpert, zwingt dazu, die Prosa nach ihm zu richten (§10).
func katalognamen(t *testing.T) []string {
	t.Helper()

	roh, err := os.ReadFile(katalogpfad) //nolint:gosec // ein fester Pfad im Repositorium
	if err != nil {
		t.Fatalf("%s lesen: %v", katalogpfad, err)
	}

	var namen []string

	for _, zeile := range strings.Split(string(roh), "\n") {
		// Eine Fähigkeitszeile hat sechs Spalten und beginnt mit einem Strich.
		//
		// **Sechs seit dem 2026-08-23**, mit REST als vierter Fläche
		// (#997).
		if !strings.HasPrefix(zeile, "| ") || strings.Count(zeile, "|") != spaltenzahl {
			continue
		}

		// **Nur die Spalten Fenster und MCP.** Die Spalten CLI und REST nennen
		// einen **Aufruf** (`mse sets`, `GET /sets`), keinen Bezeichner — wer
		// sie mitläse, forderte eine Bindung namens `mse`.
		spalten := strings.Split(zeile, "|")
		for _, spalte := range spalten[2:4] {
			for _, treffer := range inBackenzeichen.FindAllStringSubmatch(spalte, -1) {
				namen = append(namen, treffer[1])
			}
		}
	}

	if len(namen) == 0 {
		t.Fatal("im Katalog steht kein einziger Bezeichner — der Leser greift daneben")
	}

	return namen
}

// imKatalogGezaehlt gibt, wie oft jeder Bezeichner in den Tabellen steht.
func imKatalogGezaehlt(t *testing.T) map[string]int {
	t.Helper()

	haeufigkeit := map[string]int{}
	for _, name := range katalognamen(t) {
		haeufigkeit[name]++
	}

	return haeufigkeit
}

// bindungenAusApp gibt die Methoden auf `*Bindungen` — **über den Syntaxbaum**.
//
// # Warum nicht mehr über ein Muster
//
// Der Regex verlangte den Empfängernamen `b`: `func (x *Bindungen) Neu(` war
// für ihn keine Bindung, und eine neue Methode verschwände lautlos aus dem
// Katalog. Ein Wächter, der an einer Schreibgewohnheit hängt, hält seine
// Zusage so lange, wie sich alle daran halten (Zweitblick über Phase T).
//
// **Hier auch die nicht exportierten**, und das ist der Unterschied zum
// Zähler im Einleitungssatz: `waehlenUndOeffnen` bindet Wails nicht — aber sie
// ist ein Weg des Fensters in dieselbe Fähigkeit, und wer sie wegließe, hätte
// einen Weg ohne Namen. Beide lesen denselben Baum
// (`bindungszaehler_test.go`), und der Unterschied ist ein Wahrheitswert.
//
// **Was der Baum nicht sieht**: eine Bindung, die zur Laufzeit entsteht. Wails
// erzeugt seine Aufrufe aus genau diesen Methoden, also gibt es sie nicht —
// aber gesagt ist es hiermit.
func bindungenAusApp(t *testing.T) []string {
	t.Helper()

	// **Die drei Zugänge des Sitzungstors sind keine Wege** (Nachlauf über
	// ID-02, #1701): `app`, `tor` und `frei` führen eine Bindung zur
	// Anwendung, nicht in eine Fähigkeit. Sie stehen hier mit Namen, damit
	// ein vierter nicht still dazukommt. `ruf` ist seit BK-01 (#2109) der
	// Kontext hinter dem Tor — derselbe Zugang, kein Weg.
	return slices.DeleteFunc(bindungenAusVerzeichnis(t, false), func(name string) bool {
		return name == "app" || name == "tor" || name == "frei" || name == "ruf"
	})
}

// ausQuelltext sammelt alle Treffer eines Musters unter einem Verzeichnis.
func ausQuelltext(t *testing.T, verzeichnis string, muster *regexp.Regexp) []string {
	t.Helper()

	var namen []string

	fehler := filepath.WalkDir(filepath.Join(wurzel, verzeichnis),
		func(pfad string, eintrag os.DirEntry, err error) error {
			if err != nil || eintrag.IsDir() ||
				!strings.HasSuffix(pfad, ".go") || strings.HasSuffix(pfad, "_test.go") {
				return err
			}

			inhalt, lesefehler := os.ReadFile(pfad) //nolint:gosec // Quelldatei im Repositorium
			if lesefehler != nil {
				return lesefehler
			}

			for _, treffer := range muster.FindAllStringSubmatch(string(inhalt), -1) {
				if !slices.Contains(namen, treffer[1]) {
					namen = append(namen, treffer[1])
				}
			}

			return nil
		})
	if fehler != nil {
		t.Fatalf("%s durchgehen: %v", verzeichnis, fehler)
	}

	if len(namen) == 0 {
		t.Fatalf("in %s fand sich nichts — das Muster greift daneben", verzeichnis)
	}

	return namen
}

// **Bei genau einer Fähigkeit, nicht bei mindestens einer.**
//
// Hier stand `slices.Contains`, und das prüfte die Zusage nur zur Hälfte: eine
// Bindung, die in **zwei** Zeilen steht, blieb grün — gemessen am 2026-08-23
// mit `Sets` in einer zweiten Zeile. Damit war jede Zahl darunter falsch, ohne
// dass jemand es sah, und die Aussage „welche Fläche kann was" beruhte auf
// einer Zuordnung, die keine war (Zweitblick über Phase T).
func TestJedeBindungStehtBeiGenauEinerFaehigkeit(t *testing.T) {
	haeufigkeit := imKatalogGezaehlt(t)

	for _, name := range bindungenAusApp(t) {
		pruefeGenauEinmal(t, haeufigkeit, name, "die Bindung")
	}
}

func TestJedesMCPWerkzeugStehtBeiGenauEinerFaehigkeit(t *testing.T) {
	haeufigkeit := imKatalogGezaehlt(t)

	for _, name := range ausQuelltext(t, "internal/mcp", mcpwerkzeug) {
		pruefeGenauEinmal(t, haeufigkeit, name, "das Werkzeug")
	}
}

// pruefeGenauEinmal hält beide Hälften der Zusage an **einer** Stelle.
func pruefeGenauEinmal(t *testing.T, haeufigkeit map[string]int, name, was string) {
	t.Helper()

	switch anzahl := haeufigkeit[name]; {
	case anzahl == 0:
		t.Errorf("%s %q steht in keiner Fähigkeit — zu welcher gehört sie? (%s)",
			was, name, katalogpfad)
	case anzahl > 1:
		t.Errorf("%s %q steht bei %d Fähigkeiten — es darf genau eine sein (%s)",
			was, name, anzahl, katalogpfad)
	}
}

// **Und umgekehrt**: was im Katalog steht, gibt es auch.
//
// Ohne diese Richtung bliebe eine umbenannte Bindung als Karteileiche stehen,
// und der Katalog verspräche eine Fähigkeit, die keinen Weg mehr hat.
func TestJederNameImKatalogGibtEsWirklich(t *testing.T) {
	bindungen := bindungenAusApp(t)
	werkzeuge := ausQuelltext(t, "internal/mcp", mcpwerkzeug)

	for _, name := range katalognamen(t) {
		istWerkzeug := strings.HasPrefix(name, "lesen.") || strings.HasPrefix(name, "schreiben.")

		switch {
		case istWerkzeug:
			if !slices.Contains(werkzeuge, name) {
				t.Errorf("der Katalog nennt das Werkzeug %q — das gibt es nicht", name)
			}
		case slices.Contains(bindungen, name):
		default:
			t.Errorf("der Katalog nennt %q wie eine Bindung — die gibt es nicht", name)
		}
	}
}

// besetzt sagt, ob eine Spalte einen Eintrag trägt.
//
// **Ein Gedankenstrich ist kein Eintrag.** So schreibt der Katalog „gibt es
// hier nicht", und eine Zählung, die ihn mitnähme, zählte die Lücke als
// Fähigkeit.
func besetzt(spalte string) bool {
	gekuerzt := strings.TrimSpace(spalte)

	return gekuerzt != "" && gekuerzt != "—"
}

// zeileMitZahl trifft eine Zeile der Übersicht: `| … | **37** … |`.
var zeileMitZahl = regexp.MustCompile(`^\| (.+?) \| \*\*(\d+)\*\*`)

// faehigkeitszeilen sind die Tabellenzeilen, die eine Fähigkeit beschreiben.
func faehigkeitszeilen(t *testing.T) []string {
	t.Helper()

	roh, err := os.ReadFile(katalogpfad) //nolint:gosec // ein fester Pfad im Repositorium
	if err != nil {
		t.Fatalf("%s lesen: %v", katalogpfad, err)
	}

	var zeilen []string

	for _, zeile := range strings.Split(string(roh), "\n") {
		// **Die Kopfzeile an ihrer ersten Zelle**, nicht an „Fähigkeit |"
		// irgendwo in der Zeile: so fiel bis zum 2026-10-01 „Ein Bild
		// zuschneiden" aus jeder Zählung, weil ihre Bemerkung mit „die
		// Fähigkeit |" endet — gefunden von der Schaltertafel (OPS-10, #1908).
		if !strings.HasPrefix(zeile, "| ") || strings.Count(zeile, "|") != spaltenzahl ||
			strings.TrimSpace(strings.Split(zeile, "|")[1]) == "Fähigkeit" ||
			strings.HasPrefix(zeile, "| ---") {
			continue
		}

		zeilen = append(zeilen, zeile)
	}

	if len(zeilen) == 0 {
		t.Fatal("keine einzige Fähigkeitszeile gefunden — der Leser greift daneben")
	}

	return zeilen
}

// **Die Zahlen in der Übersicht stimmen mit den Tabellen darunter.**
//
// Eine Zahl in Fließtext, die niemand nachzählt, veraltet still. Dieser
// Wächter zählt sie nach — und macht damit sichtbar, wenn eine neue Fähigkeit
// nur an **einer** Fläche landet: die Lückenzahl stimmt dann nicht mehr.
//
// **Was er nicht sagt**: ob eine Fähigkeit an einer Fläche fehlen darf.
func TestDieZahlenImKatalogSindGezaehlt(t *testing.T) {
	zeilen := faehigkeitszeilen(t)

	// **`insgesamt` ist die Zeilenzahl, `am Fenster` nicht.**
	//
	// Hier stand für beide `len(zeilen)`, und das war eine Behauptung im
	// Gewand einer Zählung: eine Fähigkeit mit `—` in der Fensterspalte hätte
	// weiter als Fensterfähigkeit gezählt. Gemessen am 2026-08-23 — die Zeile
	// eingesetzt, beide Zahlen erhöht, Lauf grün (Zweitblick über Phase T).
	gezaehlt := map[string]int{"Fähigkeiten insgesamt": len(zeilen)}

	for _, zeile := range zeilen {
		spalten := strings.Split(zeile, "|")

		if besetzt(spalten[2]) {
			gezaehlt["davon am **Fenster**"]++
		}

		if strings.Contains(spalten[3], "`lesen.") || strings.Contains(spalten[3], "`schreiben.") {
			gezaehlt["davon über **MCP**"]++
		}

		if besetzt(spalten[4]) {
			gezaehlt["davon über **CLI**"]++
		}

		if besetzt(spalten[5]) {
			gezaehlt["davon über **REST**"]++
		}

		if strings.Contains(spalten[6], "✋") {
			gezaehlt["davon mit **Dateidialog**"]++
		}
	}

	roh, err := os.ReadFile(katalogpfad) //nolint:gosec // ein fester Pfad im Repositorium
	if err != nil {
		t.Fatalf("%s lesen: %v", katalogpfad, err)
	}

	gefunden := map[string]bool{}

	for _, zeile := range strings.Split(string(roh), "\n") {
		treffer := zeileMitZahl.FindStringSubmatch(zeile)
		if treffer == nil {
			continue
		}

		erwartet, kennen := gezaehlt[treffer[1]]
		if !kennen {
			continue
		}

		gefunden[treffer[1]] = true

		if genannt := treffer[2]; genannt != strconv.Itoa(erwartet) {
			t.Errorf("die Übersicht nennt für %q die Zahl %s — gezählt sind %d",
				treffer[1], genannt, erwartet)
		}
	}

	// **Sonst wäre eine entfernte Zeile grün.** Eine Übersicht, die eine Zahl
	// nicht mehr nennt, hat sie nicht widerlegt, sondern verschwiegen.
	for zeile := range gezaehlt {
		if !gefunden[zeile] {
			t.Errorf("die Übersicht nennt %q nicht mehr", zeile)
		}
	}
}
