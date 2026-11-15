// Der Wächter über `docs/KONTRASTE.md`.
//
// Unter jsdom meldet axe `color-contrast` als *unentschieden* — ohne Layout
// gibt es keine gerechnete Farbe. Gemessen wird deshalb im Browser, von Hand,
// und das Ergebnis steht in `docs/KONTRASTE.md`.
//
// Eine Messung von Hand veraltet still: jemand schreibt eine neue Farbe in eine
// Komponente, und die Datei sagt nichts dazu. Also zählt hier jemand nach: jede
// Vordergrundfarbe, die in `frontend/src` benutzt wird, hat dort eine Zeile.
//
// # Was er nicht prüft
//
// **Ob die Zahl stimmt.** Er kann nicht nachmessen — dafür bräuchte er einen
// Browser, und den hat die Pipeline nicht (`docs/entscheidungen/0007`).
//
// **Ob der genannte Grund der richtige ist.** Ein `text-slate-600` auf dunklem
// Grund wäre durch die Zeile „auf bg-white" gedeckt und trotzdem falsch. Der
// Wächter sieht die Klasse, nicht ihre Umgebung.
//
// Beides steht auch in der Datei selbst. Ein Wächter, der die halbe Zusage
// hält, macht die andere unsichtbar — also wird sie ausgesprochen.
package werkzeuge

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// vordergrundfarben findet `text-`, `border-`, `divide-` und `outline-` samt
// ihrem Farbton — mit und ohne `dark:`.
//
// **Nicht `bg-`**: eine Fläche ist ein Grund, kein Vordergrund, und ihre Zeile
// entsteht durch das Paar, in dem sie vorkommt. `bg-sky-800` steht trotzdem in
// der Datei, weil es dort eine Füllung ist — der Wächter verlangt es nur nicht.
//
// # Und deshalb hält er die Farbe, nicht das Paar
//
// **Das ist die Lücke, und sie gehört benannt.** Wer `bg-white` zu
// `bg-slate-500` ändert und `text-slate-600` stehen lässt, bekommt einen
// Kontrast, den niemand gemessen hat — die vorhandene Zeile zu
// `text-slate-600` genügt diesem Wächter, denn sie nennt eine Vordergrundfarbe
// und nicht ihren Grund. Angestrichen vom zweiten Zweitblick über die Freigabe
// 0.3.0 (#1170,
// Befund 11).
//
// **Was das Paar hält, ist die Sichtprüfung** (§3) und `docs/KONTRASTE.md`, wo
// neben jeder Zahl steht, **gegen welchen Grund** gemessen wurde. Ein Wächter,
// der Paare bildete, müsste wissen, welches `text-` unter welchem `bg-`
// landet — das entscheidet der Baum zur Laufzeit und nicht der Quelltext.
//
// **Dasselbe gilt für `dark:`**: der Wächter führt beide Erscheinungsbilder auf
// dieselbe Farbe zurück, und ob die dunkle Fassung ihren eigenen Grund trifft,
// sieht er nicht.
//
// # Jede Familie, keine Liste
//
// Hier stand `(?:slate|sky|red|green|amber)`, und die Anwendung benutzt
// `emerald` — vier Vordergrundfarben im Meldungsband, die dieser Wächter
// **nicht sah**. Sie stehen in `KONTRASTE.md`, aber aus Sorgfalt und nicht,
// weil jemand es verlangt hätte; die nächste hätte niemand bemerkt.
//
// Eine handgeführte Aufzählung ist genau der Fehler, den `navigation.ts` und
// die axe-Prüfung schon einmal gekostet haben: **eine zu kurze Liste ist grün,
// und niemand sieht es** (§3, §10).
//
// **Was das kostet**: `border-spacing-96` und Ähnliches träfe die Regel und
// verlangte eine Zeile, die es nicht braucht. Ein Wächter, der zu viel
// verlangt, wird bemerkt; einer, der zu wenig verlangt, nicht.
// **`ring` gehört dazu, und `white`/`black` auch.** Beides hat gefehlt, und
// beides ist beim Zuschnittrahmen aufgefallen
// (#576): er
// zeichnet `outline-white` über `ring-slate-900`, und **keine** der drei
// Angaben traf die alte Regel — `ring` fehlte in der Familienliste, und
// `-white` hat keine Tonzahl. Ein Rahmen, dessen beide Linien der Wächter
// nicht sieht, ist genau der Fall, für den es ihn gibt.
//
// **Und seit dem 2026-09-03 auch die Namen ohne Tonzahl** — Werkstück J von
// Phase 4.10. Die Gestaltungssprache heißt `text-leise`, `border-kante`,
// `outline-ring`; keiner dieser Namen trägt eine Zahl, und der Wächter sah
// **keinen** von ihnen. Gezählt am selben Tag: 287-mal `text-leise`, 181-mal
// `border-kante`, 160-mal `outline-ring`, 22-mal `text-text` — 668 Klassen,
// über die er grün war, weil sein Muster sie nicht traf.
//
// Die Namen kommen aus `stil.css` und nicht aus einer Aufzählung hier: eine
// zweite Liste liefe auseinander, und die kürzere wäre wieder die grüne.
var vordergrundfarben = regexp.MustCompile(
	`(?:dark:)?((?:text|border|divide|outline|ring)-(?:[a-z]+-\d{2,3}|white|black))`)

// sprachfarben ist dasselbe Muster für die Namen aus `stil.css`.
//
// **Es wird beim ersten Gebrauch gebaut**, weil es die Datei braucht.
func sprachfarben(t *testing.T) *regexp.Regexp {
	t.Helper()

	namen := farbnamen(t)

	return regexp.MustCompile(
		`(?:dark:)?((?:text|border|divide|outline|ring)-(?:` +
			strings.Join(namen, "|") + `))\b`)
}

// ohneQuelltextkommentar räumt Zeilen- und Blockkommentare aus TSX.
//
// **Weil ein Kommentar über eine Farbe sprechen darf.** Derselbe Fall wie in
// `feldpruefung_test.go`: ein Hinweis, warum eine Farbe **nicht** genommen
// wurde, machte diesen Wächter rot — und ein Wächter, den man nicht erklären
// darf, ohne ihn rot zu machen, erzieht zum Schweigen.
//
// **Grob und mit Absicht**: eine Zeichenkette, die `//` enthält, verliert
// hier ihren Rest. Das kostet nichts — gesucht werden Tailwind-Klassen, und
// die stehen in keiner Adresse.
func ohneQuelltextkommentar(quelle string) string {
	ohneBloecke := regexp.MustCompile(`(?s)/\*.*?\*/`).ReplaceAllString(quelle, " ")

	zeilen := strings.Split(ohneBloecke, "\n")
	for i, zeile := range zeilen {
		if stelle := strings.Index(zeile, "//"); stelle >= 0 {
			zeilen[i] = zeile[:stelle]
		}
	}

	return strings.Join(zeilen, "\n")
}

// vordergrundfarbenAus liest die Farben aus einem Stück Quelltext.
//
// **Als eigene Funktion, damit die Reichweite prüfbar ist.** Solange die Regel
// nur über den echten Baum lief, ließ sich nicht zeigen, was sie **nicht**
// findet — und genau das war der Mangel (§10).
func vordergrundfarbenAus(inhalt string, weitere ...*regexp.Regexp) []string {
	gefunden := []string{}

	for _, muster := range append([]*regexp.Regexp{vordergrundfarben}, weitere...) {
		for _, treffer := range muster.FindAllStringSubmatch(inhalt, -1) {
			gefunden = append(gefunden, treffer[1])
		}
	}

	return gefunden
}

// # Was dieser Wächter nicht unterscheidet
//
// **Ob eine Zeile die helle oder die dunkle Ansicht beschreibt.** Er entfernt
// `dark:` und ordnet nach der Farbe; eine einzige Zeile genügt damit auch für
// eine Farbe, die nur im dunklen Bild vorkommt. Angestrichen hat es der
// Zweitblick über Phase O.
//
// **Ableiten lässt es sich nicht**, und das ist gemessen: `dark:text-slate-950`
// steht auf `bg-sky-300` — einem **hellen** Grund, benutzt im dunklen Bild
// (`Knopf.tsx:79`). Eine Regel „dunkler Grund heißt dunkle Ansicht" hielte
// genau diesen Fall für einen Fehler.
//
// Damit bräuchte jede Zeile eine ausdrückliche Angabe der Ansicht. Das ist ein
// Formatwechsel über 26 Zeilen und steht als
// #832 — hier
// steht solange, was fehlt, statt dass es niemand weiß.
func TestJedeVordergrundfarbeStehtInKontraste(t *testing.T) {
	gemessen, err := os.ReadFile(filepath.Join("..", "..", "docs", "KONTRASTE.md"))
	if err != nil {
		t.Fatalf("KONTRASTE.md nicht lesbar: %v", err)
	}

	benutzt := map[string][]string{}
	sprache := sprachfarben(t)

	wurzel := filepath.Join("..", "..", "frontend", "src")

	if err := filepath.WalkDir(wurzel, func(pfad string, eintrag os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// **`.ts` zählt mit, seit eine Farbtabelle dort steht** (#1164).
		//
		// `bausteine/farben.ts` ordnet jeder Art des Lexers eine Klasse zu —
		// sechs Vordergrundfarben, von denen dieser Wächter keine sah, weil er
		// nur `.tsx` las. Eine Klasse ist eine Klasse, gleich in welcher Datei
		// sie steht; die Endung war nie die Aussage.
		if eintrag.IsDir() || strings.HasSuffix(pfad, ".test.tsx") ||
			strings.HasSuffix(pfad, ".test.ts") ||
			(!strings.HasSuffix(pfad, ".tsx") && !strings.HasSuffix(pfad, ".ts")) {
			return nil
		}

		inhalt, err := os.ReadFile(pfad) //nolint:gosec // Quelltext im Repository
		if err != nil {
			return err
		}

		// **Ohne Kommentare** — seit dem 2026-09-05.
		//
		// # Was hier vorher schiefging
		//
		// Der Wächter las **rohen** Text, und ein Kommentar, der eine
		// **verworfene** Farbe nennt („`border-kante` und nicht
		// `border-grund`"), machte ihn rot. Genau derselbe Fall stand schon
		// in `feldpruefung_test.go`, und der Satz dort gilt hier
		// unverändert: *ein Wächter, den man nicht erklären darf, ohne ihn
		// rot zu machen, erzieht zum Schweigen.*
		//
		// **Die Zusage schrumpft dabei nicht**: eine Farbe, die wirklich
		// benutzt wird, steht im Quelltext und nicht in einem Kommentar.
		//
		// **`ohneQuelltextkommentar` und nicht `ohneKommentar`.** Es gibt in
		// diesem Verzeichnis **zwei** Funktionen dieses Namens, in zwei
		// Paketen: die eine schneidet einen **Zeilen**kommentar aus einer
		// `go.mod`-Zeile ab, die andere räumt eine ganze TSX-Datei. Der
		// erste Anlauf hier griff die falsche — und sie schnitt bei der
		// ersten `//`-Stelle **die ganze Datei** ab. Der Wächter fand danach
		// **null** Farben und sagte es; ohne den Satz „keine einzige Farbe
		// gefunden" wäre er grün gewesen und hätte nichts mehr geprüft
		// (§10, die Reichweite einer Messung).
		for _, farbe := range vordergrundfarbenAus(ohneQuelltextkommentar(string(inhalt)), sprache) {
			benutzt[farbe] = append(benutzt[farbe], filepath.Base(pfad))
		}

		return nil
	}); err != nil {
		t.Fatalf("frontend/src nicht lesbar: %v", err)
	}

	if len(benutzt) == 0 {
		t.Fatal("keine einzige Farbe gefunden — der Wächter greift ins Leere")
	}

	werte := gemesseneWerte(string(gemessen))

	for farbe, dateien := range benutzt {
		wert, gefunden := werte[farbe]
		if !gefunden {
			t.Errorf("%s wird in %s benutzt, hat aber keine **Zeile mit Wert** "+
				"in docs/KONTRASTE.md", farbe, strings.Join(dateien, ", "))

			continue
		}

		if schwelle := schwelleFuer(farbe); wert < schwelle {
			t.Errorf("%s steht mit %.2f in docs/KONTRASTE.md — unter %.1f, benutzt in %s",
				farbe, wert, schwelle, strings.Join(dateien, ", "))
		}
	}
}

// Die zwei Schwellen aus `docs/KONTRASTE.md`.
//
// **Fließtext braucht mehr als ein Rand** — 4,5 nach WCAG 1.4.3 gegen 3 nach
// 1.4.11. Der erste Anlauf legte 3 an **alle** an; angestrichen hat es der
// Zweitblick über Phase O.
//
// **Beide kosten keine Ausnahme.** Ausgezählt am 2026-08-22: von den 24 Farben
// mit einer Zeile liegt keine unter ihrer Schwelle — sie waren eingehalten und
// nur nicht zugesichert.
//
// **Was die 4,5 überzeichnet**: WCAG erlaubt großem Text ab 24 px ebenfalls 3.
// Eine Farbe, die **nur** groß gesetzt wird, verlangt dieser Wächter strenger,
// als sie müsste. Heute gibt es keine solche; käme eine, gehört die Ausnahme
// hierher und nicht eine gesenkte Schwelle für alle.
const (
	mindestverhaeltnis     = 3.0
	mindestverhaeltnisText = 4.5
)

// schwelleFuer gibt die Schwelle, die für diese Farbe gilt.
func schwelleFuer(farbe string) float64 {
	if strings.HasPrefix(farbe, "text-") {
		return mindestverhaeltnisText
	}

	return mindestverhaeltnis
}

// messzeile liest eine Zeile der Form `| `farbe` | `grund` | 4,76 | … |`.
//
// **Eine Zeile mit Wert und nicht eine Erwähnung.** Hier stand
// `strings.Contains(gemessen, "`"+farbe+"`")`, und „irgendwo" schloss den
// Fließtext ein — auch den Satz, der sagte, eine Farbe sei **abgeschafft**.
// `border-slate-400` stand vier Stellen lang im JSX, während die Datei ihre
// Hebung behauptete, und der Lauf war grün
// (#816).
var messzeile = regexp.MustCompile(
	"^\\|\\s*`((?:text|border|divide|outline|ring)-[a-z0-9-]+)`\\s*\\|[^|]*\\|\\s*([0-9]+[,.][0-9]+)\\s*\\|")

// gemesseneWerte gibt je Farbe das **kleinste** gemessene Verhältnis.
//
// Das kleinste, weil eine Farbe auf mehreren Gründen steht — hell und dunkel —
// und die schlechtere Paarung entscheidet, ob sie taugt.
func gemesseneWerte(datei string) map[string]float64 {
	werte := map[string]float64{}

	for _, zeile := range strings.Split(datei, "\n") {
		treffer := messzeile.FindStringSubmatch(zeile)
		if treffer == nil {
			continue
		}

		wert, err := strconv.ParseFloat(strings.Replace(treffer[2], ",", ".", 1), 64)
		if err != nil {
			continue
		}

		if vorher, da := werte[treffer[1]]; !da || wert < vorher {
			werte[treffer[1]] = wert
		}
	}

	return werte
}

// TestDerFarbwaechterFindetJedeFamilie ist die Gegenprobe.
//
// **Ohne sie wäre die Erweiterung eine Behauptung.** Der Wächter lief zuvor
// gegen eine Liste von fünf Familien und übersah `emerald`; dass er jetzt jede
// findet, steht hier und nicht im Kommentar (§10).
func TestDerFarbwaechterFindetJedeFamilie(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name, quelle string
		erwartet     []string
	}{
		{"slate", `className="text-slate-900"`, []string{"text-slate-900"}},

		// **Die Familie, an der es aufgefallen ist.**
		{
			"emerald", `'border-emerald-700 text-emerald-800'`,
			[]string{"border-emerald-700", "text-emerald-800"},
		},

		{
			"eine, die es hier noch nicht gibt", `className="text-teal-700"`,
			[]string{"text-teal-700"},
		},
		{"mit dark:", `className="dark:text-violet-300"`, []string{"text-violet-300"}},
		{
			"alle vier Anfänge", `text-a-100 border-b-200 divide-c-300 outline-d-400`,
			[]string{"text-a-100", "border-b-200", "divide-c-300", "outline-d-400"},
		},

		// **Eine Fläche verlangt keine Zeile** — der Grund steht bei
		// [vordergrundfarben].
		{"bg bleibt draußen", `className="bg-sky-800"`, nil},

		// **Und was keine Farbe ist, bleibt draußen.** Eine Ziffer allein
		// reicht nicht; verlangt sind zwei oder drei.
		{"outline-offset-2", `className="outline-offset-2"`, nil},
		{"border-t-2", `className="border-t-2"`, nil},
		{"text-sm", `className="text-sm"`, nil},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			ist := vordergrundfarbenAus(fall.quelle)
			if len(ist) != len(fall.erwartet) {
				t.Fatalf("gefunden %v, erwartet %v", ist, fall.erwartet)
			}

			for i := range fall.erwartet {
				if ist[i] != fall.erwartet[i] {
					t.Errorf("gefunden %v, erwartet %v", ist, fall.erwartet)
				}
			}
		})
	}
}
