// **Keine Bindung ohne Aufrufer** —
// #1291,
// Werkstück E.
//
// # Warum es diesen Wächter gibt
//
// `AGENTS.md` §2 sagt es an einer anderen Stelle und mit Nachdruck: *„Ein
// Paket ohne Aufrufer ist keine Anwendungsfunktion; das ist die erste
// Dimension aus §10, und sie steht dort zuerst, weil sie schon einmal gefehlt
// hat."* Damals war es `internal/svg` — vollständig, geprüft, und **niemand
// konnte eine SVG erzeugen**.
//
// Für die Bindungen galt dieselbe Gefahr und kein Wächter. `flaechenlauf` hält,
// dass eine Fähigkeit an mehreren Flächen gefahren wird; dass das **Fenster**
// seine eigene Bindung überhaupt ruft, hielt niemand.
//
// **Gemessen am 2026-08-30**, bevor dieser Lauf entstand: 142 Bindungen, und
// jede wird gerufen. Der Wächter hält also einen Zustand, der schon stimmt —
// das ist der günstigste Zeitpunkt, ihn aufzuschreiben.
//
// # Zwei Umgehungen, die der Zweitblick gefunden hat
//
// Der erste Anlauf las die Namen mit einem **Muster** über `func (b
// *Bindungen)` — eine Methode mit einem anderen Empfängernamen fiel heraus,
// und genau dieser Fehler war in `bindungszaehler_test.go` schon einmal
// bezahlt worden. Er liest die Namen jetzt aus demselben **Syntaxbaum**.
//
// Und er suchte den Namen im **rohen** Quelltext: ein Kommentar oder eine
// Zeichenkette galt als Aufrufer. Beides wird jetzt vorher entfernt (Befund 8
// des Zweitblicks über Phase 4.4).
//
// # Was er hält
//
// Jeder Name aus `func (b *Bindungen) …` kommt in `frontend/src` vor —
// **außerhalb** der Dateien, die ihn nur **deklarieren**.
//
// # Was er **nicht** hält
//
//   - **Nicht, dass der Aufruf erreichbar ist.** Ein Knopf hinter einer
//     Bedingung, die nie zutrifft, zählt hier mit. Das sagt kein Wächter; das
//     misst der Lauf über die laufende Brücke.
//   - **Nicht die unexportierten Helfer.** `waehlenUndOeffnen` ist ein Weg des
//     Fensters in eine Fähigkeit und keine Bindung — Wails erzeugt für sie
//     keinen Aufruf (#1287).
package werkzeuge

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// nurDeklaration sind die Dateien, die eine Bindung **nennen**, ohne sie zu
// rufen: die Fassade zählt sie auf, die `api…`-Dateien geben ihr einen Typ.
//
// **Ohne sie wäre der Lauf wertlos**: `fassade.ts` führt jeden Namen, und der
// Wächter wäre immer grün.
var nurDeklaration = regexp.MustCompile(`^(fassade|api[a-z]*)\.tsx?$|\.test\.`)

// TestJedeBindungWirdGerufen hält §2s Satz über Aufrufer.
func TestJedeBindungWirdGerufen(t *testing.T) {
	t.Parallel()

	// **Derselbe Syntaxbaumleser wie im Zähler** — nicht ein zweites Muster
	// daneben, das dieselbe Frage anders beantwortet. Dafür steht dieser Lauf
	// in `package werkzeuge` und nicht in `werkzeuge_test`: über die
	// Paketgrenze reicht kein Helfer hinüber.
	namen := bindungenAusVerzeichnis(t, true)
	if len(namen) < 100 {
		t.Fatalf("nur %d Bindungen gefunden — das Muster trifft nicht", len(namen))
	}

	rufend := rufenderQuelltext(t)

	for _, name := range namen {
		if !wirdGerufen(rufend, name) {
			t.Errorf("die Bindung %q wird nirgends gerufen — sie ist keine "+
				"Anwendungsfunktion, sondern eine Zusage ohne Einlösung (§2)", name)
		}
	}
}

// wirdGerufen sagt, ob ein Name im Quelltext **gerufen oder weitergegeben**
// wird — nicht bloß erwähnt.
//
// # Der Befund, aus dem diese Funktion entstanden ist
//
// Hier stand `strings.Contains(rufend, name)`. Kommentare und Zeichenketten
// waren zwar schon herausgenommen, aber **eine Erwähnung ist noch kein
// Aufruf**: `void api().DiagnoseSichern` ohne Klammern kam durch, und damit
// jede Bindung, deren Name irgendwo als Wort steht
// (#1346).
//
// # Warum drei Zeichen und kein Zerteiler
//
// Für TypeScript liegt in diesem Werkzeugkasten keiner (§1) — die ehrliche
// Zwischenstufe ist, die Gestalt zu verlangen, in der ein Aufruf aussieht.
// **Gemessen am 2026-08-31** über alle 153 Bindungen: 151 stehen als
// `name(`, zwei als **Wert** hinter `api()` und vor einem Komma —
// `punktaendern(api().PunktEinfuegen, teil, punkt)` in `usePunktwege.ts`.
// Deshalb zählen `(`, `,` und `)`.
//
// # Was sie nicht hält
//
// Dass der Aufruf **erreicht** wird. Ein `api().Sets()` in einem Zweig, den
// niemand betritt, kommt hier durch — dafür ist die Abdeckung da.
func wirdGerufen(quelltext, name string) bool {
	for stelle := 0; ; {
		fund := strings.Index(quelltext[stelle:], name)
		if fund < 0 {
			return false
		}

		stelle += fund + len(name)

		// **Ein längerer Name ist ein anderer Name.** `Sets` steckt in
		// `SetsFuerSpiel`; ohne diese Grenze wäre die zweite Bindung der
		// Beleg für die erste.
		if stelle < len(quelltext) && istWortzeichen(quelltext[stelle]) {
			continue
		}

		nach := strings.TrimLeft(quelltext[stelle:], " \t\n\r")
		if nach != "" && strings.ContainsRune("(,)", rune(nach[0])) {
			return true
		}
	}
}

// istWortzeichen sagt, ob ein Byte einen Bezeichner fortsetzen kann.
func istWortzeichen(zeichen byte) bool {
	return zeichen == '_' || zeichen == '$' ||
		(zeichen >= '0' && zeichen <= '9') ||
		(zeichen >= 'a' && zeichen <= 'z') ||
		(zeichen >= 'A' && zeichen <= 'Z')
}

// **Die Gegenprobe steht als Lauf da, nicht als Erinnerung an einen Versuch.**
//
// §10: *ein Wächter beweist nichts, bevor eine gezielte Verletzung ihn rot
// gemacht hat.* Der Nachlauf über Phase 4.4 hat `void api().DiagnoseSichern`
// ohne Klammern gefahren und einen grünen Lauf bekommen; dieser Lauf hält
// genau das fest.
func TestEineErwaehnungIstNochKeinAufruf(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		quelltext string
		gerufen   bool
	}{
		"gerufen":                     {"await api().Sets()", true},
		"als Wert übergeben":          {"lade(api().Sets, teil)", true},
		"als letztes Argument":        {"lade(teil, api().Sets)", true},
		"nur erwähnt":                 {"void api().Sets", false},
		"in einer Zuweisung ohne Ruf": {"const x = api().Sets\n", false},
		"längerer Name":               {"await api().SetsFuerSpiel()", false},
		"gar nicht da":                {"await api().Karten()", false},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if wirdGerufen(fall.quelltext, "Sets") != fall.gerufen {
				t.Errorf("%q: gerufen=%v, erwartet %v",
					fall.quelltext, !fall.gerufen, fall.gerufen)
			}
		})
	}
}

// rufenderQuelltext gibt den Teil des Frontends, der Bindungen **ruft**.
func rufenderQuelltext(t *testing.T) string {
	t.Helper()

	var gebaut strings.Builder

	wurzel := filepath.Join("..", "..", "frontend", "src")

	fehler := filepath.WalkDir(wurzel, func(pfad string, eintrag os.DirEntry, err error) error {
		if err != nil || eintrag.IsDir() {
			return err //nolint:wrapcheck // der Aufrufer meldet ihn
		}

		if !zaehltAlsRufer(pfad) {
			return nil
		}

		roh, lesefehler := os.ReadFile(pfad) //nolint:gosec // ein Pfad aus dem Repositorium
		if lesefehler != nil {
			return lesefehler //nolint:wrapcheck // der Aufrufer meldet ihn
		}

		gebaut.WriteString(ohneKommentareUndTexte(string(roh)))

		return nil
	})
	if fehler != nil {
		t.Fatalf("frontend/src abgehen: %v", fehler)
	}

	if gebaut.Len() == 0 {
		t.Fatal("kein rufender Quelltext gefunden — das Muster trifft nicht")
	}

	return gebaut.String()
}

// zaehltAlsRufer sagt, ob eine Datei des Frontends als Rufer einer Bindung
// zählt — **nicht**, wenn sie unter `src/test/` liegt.
func zaehltAlsRufer(pfad string) bool {
	name := filepath.Base(pfad)
	if !strings.HasSuffix(name, ".ts") && !strings.HasSuffix(name, ".tsx") {
		return false
	}

	if strings.Contains(filepath.ToSlash(pfad), "/frontend/src/test/") {
		return false
	}

	return !nurDeklaration.MatchString(name)
}

// kommentarOderText trifft einen Zeilen- oder Blockkommentar und jede
// Zeichenkette — auch die mit Rückwärtsakzenten.
var kommentarOderText = regexp.MustCompile(
	"//[^\n]*|(?s)/\\*.*?\\*/|`[^`]*`|\"(?:[^\"\\\\\n]|\\\\.)*\"|'(?:[^'\\\\\n]|\\\\.)*'")

// ohneKommentareUndTexte nimmt aus einem Quelltext alles, was kein Aufruf sein
// kann.
//
// **Der Grund ist Befund 8 des Zweitblicks über Phase 4.4**: `strings.Contains`
// über den rohen Text zählte einen **Kommentar** als Aufrufer. Wer eine
// Bindung entfernt und den Kommentar darüber stehen lässt, hätte einen grünen
// Lauf und keine Fähigkeit.
//
// **Es ist kein Zerteiler**, und das steht hier, damit niemand ihn dafür hält:
// eine Zeichenkette mit `//` darin verkürzt die Zeile, ein Aufruf in einem
// Vorlagenliteral geht verloren. Für diese Frage — kommt der Name irgendwo
// **außerhalb** von Kommentar und Text vor — trägt es.
func ohneKommentareUndTexte(quelltext string) string {
	return kommentarOderText.ReplaceAllString(quelltext, " ")
}

// **Eine Vorrichtung ruft nicht** — SEC04
// (#2357),
// Befund Z-02. Unter `frontend/src/test/` liegen Helfer der Tests; ein Ruf
// dort belegt keinen Verbraucher im Fenster, so wenig wie einer in einer
// `.test.`-Datei.
func TestEineVorrichtungRuftNicht(t *testing.T) {
	t.Parallel()

	faelle := map[string]bool{
		filepath.Join("frontend", "src", "haken", "partie.ts"):           true,
		filepath.Join("frontend", "src", "seiten", "Spielseite.tsx"):     true,
		filepath.Join("frontend", "src", "test", "partieentwurf.ts"):     false,
		filepath.Join("frontend", "src", "test", "unterordner", "x.tsx"): false,
		filepath.Join("frontend", "src", "haken", "partie.test.tsx"):     false,
		filepath.Join("frontend", "src", "fassade.ts"):                   false,
		filepath.Join("frontend", "src", "apikarten.ts"):                 false,
		filepath.Join("frontend", "src", "stil.css"):                     false,
	}

	for pfad, erwartet := range faelle {
		if zaehltAlsRufer(filepath.Join("..", "..", pfad)) != erwartet {
			t.Errorf("%s: als Rufer %v, erwartet %v", pfad, !erwartet, erwartet)
		}
	}
}
