// Kein Bedienelement der Saat ohne Wirkung
// (#770).
//
// # Warum die Saat und nicht das Pack
//
// **Eine fremde Vorlage darf tote Optionen führen** — sie gehört uns nicht, und
// `M15-Magic-Pack` ist voll davon. Die Saat gehört uns, und sie wird
// **ausgeliefert** (`saat/LIESMICH.md`): wer die Anwendung auf einem frischen
// Rechner öffnet, sieht sie und sonst nichts.
//
// Am 2026-08-21 führte sie **drei** `styling field:` gegen **0**
// `styling.`-Nutzungen. Wer sie umstellte, sah keine Änderung.
//
// # Was dieser Wächter hält, und was nicht
//
// **Er hält die Klasse**: jedes deklarierte Bedienelement wird irgendwo in der
// Saat gelesen — eine `styling field:` als `styling.<name>`, eine
// `option field:` einer Exportvorlage als `options.<name>`. Eine neue tote
// Option fällt damit auf, statt mitgeliefert zu werden.
//
// **Der Name darf einen Umlaut tragen**, und das ist kein Randfall: die
// Exportvorlage der Saat führt `überschrift`. Ein Muster über `[A-Za-z0-9_]`
// findet `options.überschrift` **gar nicht** und meldete die Option als tot —
// ein Wächter, der beim ersten echten Namen falschen Alarm schlägt. Deshalb
// `\p{L}`.
//
// **Er hält nicht, dass die Wirkung sichtbar ist.** Ein `styling.x` in einem
// Ausdruck, dessen Ergebnis nichts ändert, käme durch. Dafür gibt es den
// zweiten Wächter — `app.TestEineStiloptionDerSaatAendertDasBild` zeichnet die
// Karte zweimal und vergleicht. Der prüft **eine** Option am Bild; dieser hier
// prüft **alle** am Text. Zusammen ergeben sie die Zusage; einzeln hält jeder
// die halbe.
package werkzeuge

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// ohneWirkung sind die Optionen, die heute nichts tun — **mit Grund und
// Fundstelle**.
//
// **Eine Ausnahmeliste und kein Weglassen**: wer die Prüfung für die beiden
// abschaltete, verlöre sie auch für die nächste. So bleibt die Klasse gehalten
// und die Schuld sichtbar.
//
// **Sie ist seit dem 2026-08-22 leer**, und das ist die Geschichte, für die es
// sie gab: `hintergrund` und `rahmen zeigen` standen darin, weil die Saat
// nichts hatte, woran sie hängen konnten. Sie hat es jetzt — zwei
// `extra card field:` vom Typ `color`, wie MSE seinen Kartenrand zeichnet.
//
// **Der Wächter hat es selbst verlangt**: kaum wirkten die beiden, wurde er rot
// mit „steht in `ohneWirkung`, ist aber keine tote Option mehr". Eine
// Ausnahmeliste, die ihre eigenen Einträge einfordert, veraltet nicht.
var ohneWirkung = map[string]string{}

// felderarten sind die Blöcke, die ein Bedienelement erklären, samt der
// Redewendung, mit der eine Vorlage es liest.
//
// **Zwei, nicht eine.** Eine Exportvorlage führt `option field:` und liest sie
// als `options.<name>` — dieselbe Klasse, andere Wörter. Heute sind beide
// Optionen von `probe-liste` benutzt; der Wächter hält, dass es so bleibt.
var felderarten = []struct {
	block   *regexp.Regexp
	zugriff *regexp.Regexp
	wort    string
}{
	{
		block:   regexp.MustCompile(`(?m)^\t*styling field:\s*$`),
		zugriff: regexp.MustCompile(`styling\.([\p{L}0-9_]+)`),
		wort:    "Stiloption",
	},
	{
		block:   regexp.MustCompile(`(?m)^\t*option field:\s*$`),
		zugriff: regexp.MustCompile(`options\.([\p{L}0-9_]+)`),
		wort:    "Ausfuhroption",
	},
}

// feldname findet die `name:`-Zeile eines Blocks.
var feldname = regexp.MustCompile(`(?m)^\t+name:\s*(.+?)\s*$`)

func TestKeinBedienelementDerSaatOhneWirkung(t *testing.T) {
	t.Parallel()

	dateien := saatdateien(t)
	erklaert := map[string]bool{}

	for _, art := range felderarten {
		// **Je Paket, nicht über alle** — aus dem Zweitblick über Phase L
		// (2026-08-22). Der erste Anlauf sammelte die Nutzungen über
		// **sämtliche** Saatdateien: eine zweite tote `titelfarbe` in einem
		// anderen Paket wäre durch die Nutzung in `probe-schlicht` gedeckt
		// gewesen. Eine Vorlage liest nur ihre **eigenen** Optionen.
		for pfad, inhalt := range dateien {
			gelesen := map[string]bool{}
			for _, treffer := range art.zugriff.FindAllStringSubmatch(ohneKommentare(inhalt), -1) {
				gelesen[treffer[1]] = true
			}

			for _, name := range blocknamen(inhalt, art.block) {
				// **Leerzeichen werden Unterstriche** — so setzt
				// `stilumgebung.go` die Namen in die Skriptumgebung, und so
				// stehen sie in jeder Vorlage.
				if gelesen[strings.ReplaceAll(name, " ", "_")] {
					continue
				}

				grund, benannt := ohneWirkung[name]
				if !benannt {
					t.Errorf("%s führt die %s %q, und niemand liest sie — "+
						"ein ausgeliefertes Bedienelement ohne Wirkung (#770)",
						filepath.Base(filepath.Dir(pfad)), art.wort, name)

					continue
				}

				erklaert[name] = true
				t.Logf("  %-16s ohne Wirkung: %s", name, grund)
			}
		}
	}

	// **Und eine Ausnahme, die niemand mehr braucht, verschwindet.** Sonst
	// stünde die Schuld noch da, wenn sie längst getilgt ist.
	for name := range ohneWirkung {
		if !erklaert[name] {
			t.Errorf("%q steht in `ohneWirkung`, ist aber keine tote Option "+
				"der Saat mehr — der Eintrag gehört gelöscht", name)
		}
	}
}

// blocknamen gibt die `name:`-Zeile jedes Blocks dieser Art.
func blocknamen(inhalt string, block *regexp.Regexp) []string {
	var aus []string

	for _, kopf := range block.FindAllStringIndex(inhalt, -1) {
		rest := inhalt[kopf[1]:]

		// **Nur bis zum nächsten Block**: sonst nähme der letzte Kopf den
		// Namen eines Feldes, das gar nicht zu ihm gehört.
		if ende := block.FindStringIndex(rest); ende != nil {
			rest = rest[:ende[0]]
		}

		if treffer := feldname.FindStringSubmatch(rest); treffer != nil {
			aus = append(aus, treffer[1])
		}
	}

	return aus
}

// saatdateien gibt den Inhalt jeder `style`- und `game`-Datei der Saat.
func saatdateien(t *testing.T) map[string]string {
	t.Helper()

	aus := map[string]string{}
	wurzel := filepath.Join("..", "..", "saat")

	fehler := filepath.Walk(wurzel, func(pfad string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		switch info.Name() {
		case "style", "game", "export-template":
		default:
			return nil
		}

		roh, err := os.ReadFile(pfad) //nolint:gosec // der Pfad kommt aus dem eigenen Baumlauf
		if err != nil {
			return err
		}

		aus[pfad] = string(roh)

		return nil
	})
	if fehler != nil {
		t.Fatalf("die Saat lesen: %v", fehler)
	}

	if len(aus) == 0 {
		t.Fatal("keine einzige Saatdatei gefunden — der Wächter sieht den Baum nicht")
	}

	return aus
}

// ohneKommentare nimmt die Kommentarzeilen aus einer Paketdatei.
//
// **Sonst deckt eine Erwähnung die Wirkung** — aus dem Zweitblick über Phase L.
// `saat/probe-schlicht.mse-style` erklärt in einem Kommentar, wozu seine
// Zusatzfelder da sind, und nennt dabei `styling.hintergrund`. Der erste Anlauf
// zählte das als Nutzung: ein Wächter, den ein Satz Prosa zufriedenstellt, hält
// nichts.
//
// **Eine Kommentarzeile beginnt mit `#`**, gleich wie weit sie eingerückt ist
// (`docs/FORMATE.md`).
func ohneKommentare(inhalt string) string {
	zeilen := strings.Split(inhalt, "\n")
	behalten := make([]string, 0, len(zeilen))

	for _, zeile := range zeilen {
		if strings.HasPrefix(strings.TrimSpace(zeile), "#") {
			continue
		}

		behalten = append(behalten, zeile)
	}

	return strings.Join(behalten, "\n")
}
