// **Go-Quelltext ohne seine Kommentare** — der gemeinsame Leser für jeden
// Wächter, der einen Zugriff sucht und keine Erwähnung
// (#1346).
//
// # Der Befund, aus dem er entstanden ist
//
// §10 sagt: *ein Wächter, der die halbe Zusage hält, macht die andere
// unsichtbar.* Drei Wächter suchten mit `strings.Contains` in **rohem**
// Quelltext nach einer Zeichenfolge und behaupteten damit, sie werde
// **benutzt**. Der Nachlauf über Phase 4.4 hat für jeden eine gezielte
// Verletzung gefahren, und jede blieb grün:
//
//   - `MitDrosselwerk(nil)` statt `MitDrosselwerk(drosselwerk)`, während die
//     erwartete Form im Kommentar darüber stand
//   - der `os.Getenv("MSE_SCHRIFTWURZEL")`-Zugriff entfernt, der Text im
//     Kommentar geblieben
//   - die produktive Behandlung von `Mischart` entfernt, während
//     `internal/svg/svg.go` das Feld im Kommentar nennt
//
// # Warum ein Zerteiler und kein zweites Muster
//
// Weil ein Muster über Kommentare dieselbe Frage nur anders falsch
// beantwortet. `go/parser` weiß, was ein Kommentar ist; ohne
// `parser.ParseComments` fällt er heraus, und `printer.Fprint` schreibt
// zurück, was übrig bleibt. **Zeichenketten bleiben stehen** — sie sind
// Quelltext, und `Getenv("MSE_SCHRIFTWURZEL")` lebt davon.
//
// # Was er nicht hält
//
// Er sagt, dass eine Zeichenfolge **im Code** steht, nicht dass sie an der
// richtigen Stelle steht. Ein `MitDrosselwerk(drosselwerk)` in einer
// Funktion, die niemand ruft, kommt hier durch. Wer das braucht, sucht im
// Baum — so wie `bindungenAusQuelltext` es tut.
package werkzeuge

import (
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// GoOhneKommentare gibt den Quelltext einer Go-Datei ohne ihre Kommentare.
func GoOhneKommentare(t *testing.T, pfad string) string {
	t.Helper()

	menge := token.NewFileSet()

	baum, fehler := parser.ParseFile(menge, pfad, nil, parser.SkipObjectResolution)
	if fehler != nil {
		t.Fatalf("%s zerteilen: %v", pfad, fehler)
	}

	var gebaut strings.Builder
	if err := printer.Fprint(&gebaut, menge, baum); err != nil {
		t.Fatalf("%s schreiben: %v", pfad, err)
	}

	return gebaut.String()
}

// GoVerzeichnisOhneKommentare tut dasselbe für ein ganzes Verzeichnis.
//
// **Ohne die Tests**, denn dort steht jeder Name ohnehin — ein Wächter, der
// sie mitläse, wäre grün, sobald jemand einen Testfall schreibt.
func GoVerzeichnisOhneKommentare(t *testing.T, verzeichnis string) string {
	t.Helper()

	eintraege, err := os.ReadDir(verzeichnis)
	if err != nil {
		t.Fatalf("%s lesen: %v", verzeichnis, err)
	}

	var alles strings.Builder

	for _, eintrag := range eintraege {
		name := eintrag.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}

		alles.WriteString(GoOhneKommentare(t, filepath.Join(verzeichnis, name)))
		alles.WriteString("\n")
	}

	if alles.Len() == 0 {
		t.Fatalf("in %s fand sich keine Go-Datei — der Leser greift daneben", verzeichnis)
	}

	return alles.String()
}

// **Die Gegenprobe steht als Lauf da, nicht als Erinnerung an einen Versuch.**
//
// §10 verlangt sie: *ein Wächter beweist nichts, bevor eine gezielte
// Verletzung ihn rot gemacht hat.* Eine Gegenprobe von Hand ist einmal
// gelaufen und danach eine Behauptung — diese läuft bei jedem `make ci` mit.
//
// **Sie prüft beide Richtungen**, denn ein Leser, der alles wegwirft, wäre
// ebenso grün wie einer, der nichts wegwirft.
func TestGoOhneKommentareNimmtNurDieKommentare(t *testing.T) {
	t.Parallel()

	const probe = `package probe

import "os"

// MitDrosselwerk(drosselwerk) stand hier einmal.
func Wurzel() string {
	/* auch Getenv("MSE_SCHRIFTWURZEL") in einem Blockkommentar */
	return os.Getenv("MSE_SCHRIFTWURZEL") // und am Zeilenende
}
`

	ort := filepath.Join(t.TempDir(), "probe.go")
	if err := os.WriteFile(ort, []byte(probe), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	nackt := GoOhneKommentare(t, ort)

	// **Was im Kommentar stand, ist fort.**
	if strings.Contains(nackt, "MitDrosselwerk") {
		t.Errorf("ein Kommentar blieb stehen: %q", nackt)
	}

	// **Was Quelltext ist, steht noch da** — auch die Zeichenkette darin.
	if !strings.Contains(nackt, `Getenv("MSE_SCHRIFTWURZEL")`) {
		t.Errorf("der Zugriff fehlt: %q", nackt)
	}

	// **Und zwar genau einmal.** Zwei Kommentare nannten ihn ebenfalls; bliebe
	// einer, wäre der Lauf oben grün und dieser rot.
	if zahl := strings.Count(nackt, "MSE_SCHRIFTWURZEL"); zahl != 1 {
		t.Errorf("%d Nennungen, erwartet war eine: %q", zahl, nackt)
	}
}
