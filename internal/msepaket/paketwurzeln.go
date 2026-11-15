package msepaket

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Paketwurzeln findet die Verzeichnisse, in denen Pakete unmittelbar liegen.
//
// # Wozu
//
// [Bildquelle] und [Paketleser] fügen Wurzel und Paketnamen zusammen — sie
// erwarten die Pakete **eine Ebene** unter der Wurzel. Bezogen wird aber ein
// ganzes Repositorium, und `M15-Magic-Pack` legt seine 190 Stile unter
// `M15-Magic-Pack/data/` ab. Ohne diese Suche fand die laufende Anwendung kein
// einziges Bild und kein einziges Modul
// (#206).
//
// **Sie läuft einmal beim Start**, nicht bei jedem Zugriff. Ein Baum mit
// Tausenden Dateien je Nachfrage abzugehen wäre genau der Fehler, den #187 an
// der Bildquelle behoben hat. Wer je Ruf fragt — die Flächen ohne Fenster —,
// nimmt [Wurzelsuche], die das Ergebnis merkt und je Ruf prüft (PERF-02).
//
// # Was sie nicht tut
//
// Sie öffnet kein Paket und prüft keines. Ein Verzeichnis auf `.mse-style` ohne
// lesbaren Inhalt macht seine Wurzel trotzdem zur Wurzel — was darin steht,
// beantwortet [Oeffne], und ein Fehler gehört dorthin und nicht hierher.
//
// **Archive zählen nicht mit.** Ein `.mse-set` darf eine ZIP-Datei sein, und
// [Oeffne] liest sie; als *Wurzel* taugt sie nicht, denn `Bildquelle` und
// `Paketleser` suchen darunter weitere Dateien. Gesucht werden deshalb
// Verzeichnisse.
func Paketwurzeln(wurzeln ...string) []string {
	return paketwurzelnMit(nil, wurzeln)
}

// paketwurzelnMit ist [Paketwurzeln] mit einem Beobachter des Laufs — nil
// heißt: niemand schaut zu.
func paketwurzelnMit(lauf *wurzellauf, wurzeln []string) []string {
	gefunden := make([]string, 0, len(wurzeln))
	gesehen := make(map[string]bool)

	for _, wurzel := range echteWurzeln(wurzeln) {
		lauf.vermerke(wurzel)

		for _, ort := range wurzelnUnter(lauf, wurzel) {
			if !gesehen[ort] {
				gesehen[ort] = true

				gefunden = append(gefunden, ort)
			}
		}
	}

	return gefunden
}

// GrenzePakettiefe ist die Zahl der Ebenen, die unter einer Wurzel abgegangen
// werden.
//
// **Vier, und die Zahl hat einen Grund**: ein bezogenes Repositorium legt seine
// Pakete typischerweise eine bis zwei Ebenen tief ab (`M15-Magic-Pack/data/`).
// Vier lässt Luft für eine Sammlung von Sammlungen und hält den Lauf trotzdem
// kurz.
//
// **Was tiefer liegt, wird nicht gefunden** — und zwar stillschweigend. Wer ein
// Paket sechs Ebenen tief ablegt, sieht ein leeres Ergebnis und keine Meldung.
// Das gehört gesagt, weil es sonst wie „das Paket ist kaputt" aussieht.
const GrenzePakettiefe = 4

// wurzelnUnter geht einen Baum ab und meldet, wo Pakete liegen.
func wurzelnUnter(lauf *wurzellauf, wurzel string) []string {
	var gefunden []string

	// **Der Rückgabewert wird verworfen, und das ist die Aussage**: der Rückruf
	// gibt ausschließlich `nil` oder `fs.SkipDir`, und `WalkDir` gibt genau
	// zurück, was der Rückruf gab. Ein `if` darüber wäre ein Zweig, den kein
	// Test beträte — und ein ungeprüfter Fehlerzweig ist eine Behauptung
	// darüber, was im Ernstfall geschieht (§3).
	//
	//nolint:errcheck // der Grund steht darüber
	_ = filepath.WalkDir(wurzel, func(pfad string, eintrag fs.DirEntry, err error) error {
		// **Ein Lesefehler überspringt den Ast, statt den Lauf zu beenden.**
		// Ein Verzeichnis ohne Leserecht liegt in jedem Heimatverzeichnis; wer
		// daran abbricht, verliert die Pakete daneben mit.
		if err != nil {
			lauf.scheitert()

			return nil //nolint:nilerr // der Grund steht darüber
		}

		if !eintrag.IsDir() {
			return nil
		}

		if istPaketname(eintrag.Name()) && pfad != wurzel {
			return fs.SkipDir
		}

		if tiefe(wurzel, pfad) > GrenzePakettiefe {
			return fs.SkipDir
		}

		// **Die Zeit vor dem Lesen**: was sich danach ändert, sieht der
		// nächste Ruf an einer anderen Zeit.
		lauf.vermerke(pfad)

		if traegtPakete(lauf, pfad) {
			gefunden = append(gefunden, pfad)
		}

		return nil
	})

	return gefunden
}

// tiefe zählt die Ebenen zwischen Wurzel und Pfad.
func tiefe(wurzel, pfad string) int {
	rest, err := filepath.Rel(wurzel, pfad)
	if err != nil || rest == "." {
		return 0
	}

	return strings.Count(rest, string(filepath.Separator)) + 1
}

// traegtPakete sagt, ob unmittelbar in diesem Verzeichnis ein Paket liegt.
func traegtPakete(lauf *wurzellauf, pfad string) bool {
	eintraege, err := os.ReadDir(pfad)
	if err != nil {
		lauf.scheitert()

		return false
	}

	for _, eintrag := range eintraege {
		if eintrag.IsDir() && istPaketname(eintrag.Name()) {
			return true
		}
	}

	return false
}

// istPaketname sagt, ob ein Verzeichnisname ein MSE-Paket benennt.
//
// **Über das Präfix und nicht über eine Aufzählung der sieben Endungen.**
// `docs/FORMATE.md` führt heute sieben; eine achte käme sonst als „kein Paket"
// durch, und niemand sähe es.
func istPaketname(name string) bool {
	punkt := strings.LastIndex(name, ".")

	return punkt >= 0 && strings.HasPrefix(name[punkt:], ".mse-")
}
