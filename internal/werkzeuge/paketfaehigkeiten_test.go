package werkzeuge_test

// **Was die Anwendung mit jeder Paketart tun kann** — Werkstück C von
// Phase 4.9 (#1324).
//
// # Warum es diesen Wächter gibt
//
// `docs/FORMATE.md` führt seit dem 2026-09-02 eine Tabelle: je Paketart, was
// die Anwendung mit ihr tun kann, und **über welche Bindung**. Eine solche
// Tabelle veraltet still — jemand benennt eine Bindung um, und die Zeile
// stimmt nicht mehr.
//
// # Was er hält
//
//   - Jede Bindung, die die Tabelle nennt, gibt es in `internal/app`.
//   - Jede Art aus `msepaket` hat dort eine Zeile.
//
// # Was er nicht hält
//
// **Ob die Bindung für diese Art dasselbe leistet.** `Paketdateien` listet
// Dateien; was eine `.mse-locale` darin bedeutet, weiß nur, wer sie öffnet.
// Das steht in der Tabelle daneben und ist die Grenze dieser Messung (§10).

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

func TestJedePaketartHatIhreZeile(t *testing.T) {
	t.Parallel()

	// **Die Zeilen und nicht der Abschnitt.**
	//
	// Der erste Anlauf suchte die Endung irgendwo im Abschnitt — und fand
	// `.mse-locale` im Fließtext darunter („was eine `.mse-locale` darin
	// bedeutet"). Die Gegenprobe blieb grün, und ein Wächter, den keine
	// Verletzung rot macht, ist keiner (§10). **Vierte Wiederholung
	// desselben Fehlers an einem Tag**: ein Instrument, das weiter greift
	// als seine Frage.
	zeilen := tabellenzeilen(
		abschnittAus(t, "## Was die Anwendung mit jeder Paketart tun kann"))

	for _, art := range []msepaket.Art{
		msepaket.ArtSet, msepaket.ArtSpiel, msepaket.ArtStil,
		msepaket.ArtSymbolschrift, msepaket.ArtEinschub,
		msepaket.ArtSprache, msepaket.ArtExportvorlage,
	} {
		if !strings.Contains(zeilen, "| `."+string(art)+"` |") {
			t.Errorf("docs/FORMATE.md führt keine Tabellenzeile für %q", art)
		}
	}
}

func TestJedeGenannteBindungGibtEs(t *testing.T) {
	t.Parallel()

	tabelle := abschnittAus(t, "## Was die Anwendung mit jeder Paketart tun kann")
	bindungen := bindungsnamen(t)

	// **Nur die Namen in Rückstrichen, die wie eine Bindung aussehen.** Die
	// Tabelle nennt daneben Fließtext und Dateinamen; ein Wächter, der die
	// mitzählte, wäre rot für nichts.
	muster := regexp.MustCompile("`([A-Z][A-Za-z]+)`")

	gesehen := 0

	for _, treffer := range muster.FindAllStringSubmatch(tabelle, -1) {
		name := treffer[1]

		// **Die Arten selbst stehen auch in Rückstrichen** — sie beginnen mit
		// einem Punkt und fallen durch das Muster ohnehin nicht auf.
		if name == "MSE2BESTAND" {
			continue
		}

		gesehen++

		if !bindungen[name] {
			t.Errorf("die Tabelle nennt die Bindung %q — die gibt es in "+
				"internal/app nicht", name)
		}
	}

	// **Die Gegenprobe zur Messung**: findet sie keine Namen, ist sie grün
	// und sagt nichts.
	if gesehen < 5 {
		t.Fatalf("in der Tabelle stehen nur %d Bindungsnamen — entweder ist "+
			"sie umgebaut oder dieser Wächter sucht falsch", gesehen)
	}
}

// abschnittAus gibt den Abschnitt einer Überschrift bis zur nächsten `## `.
func abschnittAus(t *testing.T, ueberschrift string) string {
	t.Helper()

	roh, err := os.ReadFile(filepath.Join("..", "..", "docs", "FORMATE.md")) //nolint:gosec // fester Pfad
	if err != nil {
		t.Fatalf("FORMATE.md lesen: %v", err)
	}

	inhalt := string(roh)

	bei := strings.Index(inhalt, ueberschrift)
	if bei == -1 {
		t.Fatalf("in FORMATE.md steht der Abschnitt %q nicht", ueberschrift)
	}

	rest := inhalt[bei+len(ueberschrift):]

	if ende := strings.Index(rest, "\n## "); ende != -1 {
		return rest[:ende]
	}

	return rest
}

// bindungsnamen sammelt die Methoden von `App`.
func bindungsnamen(t *testing.T) map[string]bool {
	t.Helper()

	wurzel := filepath.Join("..", "app")

	eintraege, err := os.ReadDir(wurzel)
	if err != nil {
		t.Fatalf("internal/app lesen: %v", err)
	}

	muster := regexp.MustCompile(`func \(a \*App\) ([A-Z][A-Za-z]*)\(`)
	gefunden := map[string]bool{}

	for _, eintrag := range eintraege {
		if !strings.HasSuffix(eintrag.Name(), ".go") ||
			strings.HasSuffix(eintrag.Name(), "_test.go") {
			continue
		}

		roh, lesefehler := os.ReadFile(filepath.Join(wurzel, eintrag.Name())) //nolint:gosec // fester Pfad
		if lesefehler != nil {
			continue
		}

		for _, treffer := range muster.FindAllStringSubmatch(string(roh), -1) {
			gefunden[treffer[1]] = true
		}
	}

	return gefunden
}

// tabellenzeilen gibt nur die Zeilen, die eine Paketart beschreiben.
//
// **Eine Zeile beginnt mit `| ` und nennt die Endung als erste Zelle.** Was im
// Fließtext daneben steht, gehört nicht zur Tabelle — siehe den Kommentar bei
// der ersten Prüfung.
func tabellenzeilen(abschnitt string) string {
	var gebaut strings.Builder

	for _, zeile := range strings.Split(abschnitt, "\n") {
		if strings.HasPrefix(strings.TrimSpace(zeile), "| `.mse-") {
			gebaut.WriteString(zeile)
			gebaut.WriteString("\n")
		}
	}

	return gebaut.String()
}
