// **Die Zustimmungserklärung zählt auf, was wirklich im Diagnosetext steht**
// (#1213,
// Befund 6).
//
// # Warum es einen Wächter braucht und nicht nur eine Korrektur
//
// Die Diagnoseseite sagt **vor** dem Anzeigen und Weitergeben, was der Text
// enthält — er geht danach an Fremde. Die Aufzählung ist damit die Grundlage
// einer **Zustimmung**, und eine Zustimmung zu einem Inhalt, den man nicht
// kennt, ist keine.
//
// Der Nachlauf über Phase 3.5 hat gefunden, dass zwei neu aufgenommene
// Inhalte — Schemastand und Paketliste — dort fehlten. **Der Fall wurde
// behoben und die Klasse blieb offen**: als am 2026-08-27 die Zeile über
// verwaiste Zeilen dazukam
// (#1220),
// fehlte sie in beiden Aufzählungen wieder, und niemand sah es (§10).
//
// # Wie er misst
//
// **Am erzeugten Text und nicht am Quelltext.** Ein Muster über `diagnose.go`
// fand auch `texte.DiagnoseNichtLesbar + ": "` — das ist ein Wert **in** einer
// Zeile und keine Aufschrift, und der Wächter hätte verlangt, „nicht zu lesen"
// in die Zustimmung zu schreiben. Was eine Aufschrift ist, entscheidet der
// Text: eine Zeichenkette am **Zeilenanfang**, gefolgt von `: `.
//
// # Was er nicht hält
//
//   - **Dass der Satz stimmt.** Er prüft, ob das Wort vorkommt, nicht ob es im
//     richtigen Zusammenhang steht. „Keine Pakete" enthielte „Pakete".
//   - **Dass die Aufzählung nicht zu viel verspricht.** Ein Inhalt, den der
//     Text gar nicht mehr schreibt, bliebe dort stehen; zu viel versprochen
//     ist hier die harmlose Seite.
//   - **Was ohne Aufschrift geschrieben wird** — die Kopfzeile mit Fassung und
//     Rechnerart und das Protokollende. Sie haben keine, und dass sie genannt
//     sind, liest ein Mensch.
package app_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// aufschrift findet eine Zeile der Gestalt `Aufschrift: Wert`.
//
// **Am Zeilenanfang und ohne führende Leerzeichen**: die Paketzeilen sind
// eingerückt, und sie sind Werte und keine Abschnitte.
var aufschrift = regexp.MustCompile(`(?m)^([^\s:][^:\n]*): `)

// zustimmungssatz findet die Aufzählung `inhalt:` im Sprachkatalog.
//
// **Der Ausdruck nennt seinen Block.** Zweimal ist er daran vorbeigelaufen,
// und beide Male sah es aus wie ein Befund am Inhalt:
//
//   - Er suchte an **sechs Zeichen Einrückung**. Die waren richtig, solange
//     der Block zwei Ebenen tief in `seiten.ts` lag; nach dem Umzug nach
//     `seitentechnik.ts` (O9, #1433) wurden es vier.
//   - Er nahm dann das **erste** `inhalt:` überhaupt — und das war ab O10
//     (#1434) der Satz des **Diagnosepakets**, nicht der der Diagnose.
//
// Beide Male galt derselbe Satz: eine Stelle, die über ihre Umgebung
// gefunden wird, ist keine Zusage über ihren Inhalt.
var zustimmungssatz = regexp.MustCompile(
	`(?s)\n\s*diagnose:\s*\{.*?\n\s*inhalt:\n?\s*'(.*?)',\n`)

func TestDieZustimmungNenntJedenDiagnoseabschnitt(t *testing.T) {
	for _, sprache := range []string{"de", "en"} {
		t.Run(sprache, func(t *testing.T) {
			// **Ein eigenes Zuhause mit echtem Protokoll** (§3): ohne die
			// Datei schriebe die Diagnose „Kein Protokoll: …" an den
			// Zeilenanfang, und der Wächter verlangte einen Satz über die
			// Abwesenheit eines Inhalts.
			zuhause := t.TempDir()
			t.Setenv("HOME", zuhause)

			protokollAnlegen(t)

			anwendung := app.NeueApp("")
			anwendung.SpracheSetzen(sprache)

			antwort := anwendung.Diagnose(t.Context())
			if antwort.Fehler != nil {
				t.Fatalf("%+v", antwort.Fehler)
			}

			satz := strings.ToLower(erklaerung(t, sprache))

			gefunden := aufschrift.FindAllStringSubmatch(antwort.Text, -1)
			if len(gefunden) < 3 {
				t.Fatalf("%d beschriftete Zeilen — der Wächter greift ins Leere:\n%s",
					len(gefunden), antwort.Text)
			}

			for _, zeile := range gefunden {
				if !strings.Contains(satz, strings.ToLower(zeile[1])) {
					t.Errorf("die Diagnose schreibt einen Abschnitt %q, "+
						"und die Zustimmungserklärung nennt ihn nicht", zeile[1])
				}
			}
		})
	}
}

// protokollAnlegen legt eine leere Protokolldatei im gesetzten Zuhause an.
func protokollAnlegen(t *testing.T) {
	t.Helper()

	pfad, err := app.Protokollpfad()
	if err != nil {
		t.Fatalf("Protokollpfad: %v", err)
	}

	if fehler := os.MkdirAll(filepath.Dir(pfad), 0o750); fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	if fehler := os.WriteFile(pfad, []byte("zeile\n"), 0o600); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}
}

// erklaerung liest `technik.diagnose.inhalt` aus einem Sprachkatalog.
func erklaerung(t *testing.T, sprache string) string {
	t.Helper()

	// **`seitentechnik.ts` und nicht `seiten.ts`** — die Texte der
	// Technikseite stehen seit Werkstück O9 (#1433) in einer eigenen Datei,
	// weil `seiten.ts` sonst über die Zeilengrenze aus §2 ginge. Ein Katalog
	// darf sich verteilen, solange er als einer ankommt (§9).
	pfad := filepath.Join("..", "..", "frontend", "src", "sprache", sprache,
		"seitentechnik.ts")

	inhalt, err := os.ReadFile(pfad) // #nosec G304 -- fester Pfad im Repository
	if err != nil {
		t.Fatalf("%s lesen: %v", pfad, err)
	}

	gefunden := zustimmungssatz.FindStringSubmatch(string(inhalt))
	if gefunden == nil {
		t.Fatalf("in %s steht keine Aufzählung `inhalt:` — der Wächter greift ins Leere", pfad)
	}

	return gefunden[1]
}
