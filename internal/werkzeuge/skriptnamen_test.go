//go:build fremdpakete

// **Jeder Skriptname von MSE2 ist hier erreichbar** — Werkstück A von
// Phase 4.9 (#1324).
//
// # Warum hinter einem Build-Tag
//
// `/Users/david/mse2-quelle` liegt **neben** dem Repositorium und nicht darin
// (§11). Ein Lauf, der ihn braucht, kann in der Pipeline nicht stehen — und
// einer, der ihn braucht und ohne ihn grün ist, wäre schlimmer als keiner.
//
//	MSE2_QUELLE=~/mse2-quelle go test -tags fremdpakete ./internal/werkzeuge/ -run Skriptnamen -v
//
// # Was er hält
//
// **Erreichbar heißt: aus einer Vorlage rufbar** — und das sind zwei Orte.
// Die Bibliothek trägt 113 Namen; sechs weitere gibt `internal/service` in die
// Umgebung, weil sie Dateien schreiben oder Symbole zeichnen und ein Blatt
// beides nicht darf (§2). Zusammen decken sie MSE2 ab.
//
// # Was er nicht hält
//
// **Ob die Funktion dasselbe tut.** Er zählt Namen an zwei Registrierstellen —
// dieselbe Grenze, die das Issue selbst nennt. Zwei Funktionen desselben
// Namens können Verschiedenes tun.
//
// **Und er sieht nur, was über `ctx.setVariable` angemeldet wird.** Was MSE2
// anders bereitstellt, zählt er nicht.
//
// # Die zweite Richtung, seit dem 2026-09-07
//
// **Er hielt bis dahin nur „MSE2 → hier".** Ein Name, den **dieses** Programm
// anbietet und MSE2 nicht, ging durch — und genau das ist am 2026-09-07
// passiert: `rule` und `warning_if_neq` sind in `basic.cpp` als
// `SCRIPT_FUNCTION` **definiert**, aber `init_script_basic_functions`
// registriert keinen der beiden ([basic.cpp:729](file:///Users/david/mse2-quelle/src/script/functions/basic.cpp)).
// `warning_if_neq` ist ein interner Helfer, auf den der Zerteiler bei
// `assert(x == y)` zeigt; `rule` heißt dort „delayed closure" und wird
// **nirgends** benutzt. Fassung f9d9356 vom 2021-07-06.
//
// Sie wurden hier trotzdem als öffentliche Namen eingetragen — mit der
// Begründung, die Portierung sei „unvollständig". Sie war es nicht: die
// beiden fehlten nicht, es gibt sie dort gar nicht.
//
// **Ein Wächter, der die halbe Zusage hält, macht die andere unsichtbar**
// (`AGENTS.md` §10). Er hält jetzt beide.
package werkzeuge_test

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// ausService sind die Namen, die `internal/service` in die Umgebung gibt.
//
// **Eine Aufzählung mit Grund, keine Suche.** Wer einen Namen hier einträgt,
// sagt, warum er nicht in der Bibliothek steht — und der Lauf unten prüft,
// dass er wirklich gesetzt wird.
var ausService = map[string]string{
	"copy_file": "kopiert eine Datei auf die Platte; ein Blatt fasst kein " +
		"Dateisystem an (§2)",
	"symbol_variation": "braucht die Symbolfassungen des Spiels und der " +
		"Vorlage, und die kennt nur, wer das Paket geöffnet hat",
	"symbols_to_html": "braucht die Symbolschrift des Spiels, um aus Zeichen " +
		"Bilder zu machen",
	"write_image_file": "schreibt eine Bilddatei auf die Platte; dasselbe " +
		"wie bei copy_file",
	"write_set_file": "schreibt ein ganzes Set auf die Platte und braucht " +
		"dafür den Paketschreiber",
	"write_text_file": "schreibt eine Textdatei auf die Platte; dasselbe " +
		"wie bei copy_file",
}

func TestJederSkriptnameVonMSE2IstErreichbar(t *testing.T) {
	ort := os.Getenv("MSE2_QUELLE")
	if ort == "" {
		t.Skip("MSE2_QUELLE ist nicht gesetzt — ohne den Quelltext von MSE2 " +
			"lässt sich nichts vergleichen")
	}

	dort := namenAusMSE2(t, ort)
	hier := msescript.Bibliothek().AlleNamen()

	var fehlend []string

	for _, name := range dort {
		if slices.Contains(hier, name) {
			continue
		}

		grund, aufgezaehlt := ausService[name]
		if !aufgezaehlt {
			fehlend = append(fehlend, name)

			continue
		}

		if len(grund) < 20 {
			t.Errorf("%q steht in ausService, aber der Grund ist zu kurz — "+
				"eine Ausnahme ohne Begründung ist eine Lücke", name)
		}

		if !wirdGesetzt(t, name) {
			t.Errorf("%q soll aus internal/service kommen, wird dort aber "+
				"nicht gesetzt", name)
		}
	}

	if len(fehlend) > 0 {
		t.Errorf("%d Namen von MSE2 sind hier nicht erreichbar: %s",
			len(fehlend), strings.Join(fehlend, " "))
	}

	// **Die Gegenprobe zur Messung**: findet sie gar nichts, ist sie grün und
	// sagt nichts. 113 sind es heute.
	if len(dort) < 100 {
		t.Fatalf("in MSE2 wurden nur %d Namen gefunden — entweder ist die "+
			"Registrierung umgebaut oder dieser Wächter sucht falsch", len(dort))
	}

	pruefeErfundeneNamen(t, dort, hier)
}

// eigeneNamen sind die Namen dieses Programms ohne Entsprechung in MSE2 —
// jeder mit seinem Grund.
//
// **Ein leerer Grund ist keiner**, wie bei jeder Ausnahmeliste hier.
//
// # Die sechs sind alle dieselbe Anpassung
//
// **In MSE2 sind sie feste Formen des Zerteilers, keine registrierten Namen**
// — nachgezählt am 2026-09-07: `min`, `max`, `rgb`, `rgba`, `true` und `false`
// kommen in `src/script/parser.cpp` (Fassung f9d9356 vom 2021-07-06) als
// `token == _("…")`
// vor und in **keiner** `setVariable`-Zeile.
//
// Hier sind es Funktionen der Standardbibliothek. Das steht seit jeher im
// Kopf von `klammerOffen` (`internal/msescript/zerteiler.go`) und ist dort
// begründet: ihr Aufruf geht denselben Weg wie jeder andere, und die
// Fehlerstelle liegt an der Klammer statt am Wort — „eine Vereinheitlichung,
// keine Verschlechterung".
//
// **Sie sind damit `adapted` und nicht erfunden**: dieselbe Fähigkeit, ein
// anderer Ort. Eine Vorlage, die `min(1,2)` schreibt, läuft hier **und** dort.
var eigeneNamen = map[string]string{
	"min":   "feste Form des MSE2-Zerteilers, hier Bibliotheksfunktion — dieselbe Fähigkeit, anderer Ort",
	"max":   "feste Form des MSE2-Zerteilers, hier Bibliotheksfunktion — dieselbe Fähigkeit, anderer Ort",
	"rgb":   "feste Form des MSE2-Zerteilers (I_TERNARY/I_RGB), hier Bibliotheksfunktion",
	"rgba":  "feste Form des MSE2-Zerteilers (I_QUATERNARY/I_RGBA), hier Bibliotheksfunktion",
	"true":  "Literal des MSE2-Zerteilers (parser.cpp:490), hier ein Name in der Umgebung",
	"false": "Literal des MSE2-Zerteilers (parser.cpp:492), hier ein Name in der Umgebung",
}

// pruefeErfundeneNamen weist ab, was dieses Programm anbietet und MSE2 nicht.
//
// # Warum das eine Prüfung wert ist
//
// **Weil ein zusätzlicher Name eine stillschweigende Semantikänderung ist**
// (§15, Regel 4). Eine Vorlage, die ihn benutzt, läuft hier und in MSE2 nicht
// — und wer sie schreibt, merkt es erst beim Weitergeben. Die Richtung
// „MSE2 → hier" fängt das nicht: sie prüft nur, dass nichts fehlt.
//
// # Was er nicht prüft
//
// **Ob der Grund stimmt.** Ein Eintrag in [eigeneNamen] mit einem plausiblen
// Satz geht durch; ob der Satz zutrifft, sieht nur ein Mensch.
func pruefeErfundeneNamen(t *testing.T, dort, hier []string) {
	t.Helper()

	var erfunden []string

	for _, name := range hier {
		if slices.Contains(dort, name) {
			continue
		}

		grund, aufgezaehlt := eigeneNamen[name]
		if !aufgezaehlt {
			erfunden = append(erfunden, name)

			continue
		}

		if len(grund) < 20 {
			t.Errorf("%q steht in eigeneNamen, aber der Grund ist zu kurz — "+
				"eine Ausnahme ohne Begründung ist eine Lücke", name)
		}
	}

	if len(erfunden) > 0 {
		sort.Strings(erfunden)
		t.Errorf("%d Namen gibt es hier und in MSE2 nicht: %s — eine Vorlage, "+
			"die sie benutzt, läuft dort nicht (§15, Regel 4). Wer einen "+
			"absichtlich anbietet, trägt ihn mit Grund in eigeneNamen ein",
			len(erfunden), strings.Join(erfunden, " "))
	}
}

// namenAusMSE2 liest die Namen, die MSE2 über `ctx.setVariable` anmeldet.
func namenAusMSE2(t *testing.T, ort string) []string {
	t.Helper()

	wurzel := filepath.Join(ort, "src", "script", "functions")

	eintraege, err := os.ReadDir(wurzel)
	if err != nil {
		t.Fatalf("%s lesen: %v", wurzel, err)
	}

	muster := regexp.MustCompile(`ctx\.setVariable\(_\("([a-z_0-9]+)"\)`)
	gefunden := map[string]bool{}

	for _, eintrag := range eintraege {
		if !strings.HasSuffix(eintrag.Name(), ".cpp") {
			continue
		}

		roh, lesefehler := os.ReadFile(filepath.Join(wurzel, eintrag.Name())) //nolint:gosec // aus MSE2_QUELLE
		if lesefehler != nil {
			t.Fatalf("%s lesen: %v", eintrag.Name(), lesefehler)
		}

		for _, treffer := range muster.FindAllStringSubmatch(string(roh), -1) {
			gefunden[treffer[1]] = true
		}
	}

	namen := make([]string, 0, len(gefunden))
	for name := range gefunden {
		namen = append(namen, name)
	}

	slices.Sort(namen)

	return namen
}

// wirdGesetzt sucht `Setze("<name>"` in `internal/service`.
func wirdGesetzt(t *testing.T, name string) bool {
	t.Helper()

	wurzel := filepath.Join("..", "service")

	eintraege, err := os.ReadDir(wurzel)
	if err != nil {
		t.Fatalf("internal/service lesen: %v", err)
	}

	for _, eintrag := range eintraege {
		if !strings.HasSuffix(eintrag.Name(), ".go") ||
			strings.HasSuffix(eintrag.Name(), "_test.go") {
			continue
		}

		roh, lesefehler := os.ReadFile(filepath.Join(wurzel, eintrag.Name())) //nolint:gosec // fester Pfad
		if lesefehler != nil {
			continue
		}

		if strings.Contains(string(roh), `Setze("`+name+`"`) {
			return true
		}
	}

	return false
}
