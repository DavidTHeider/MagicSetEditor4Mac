package werkzeuge_test

// **Wer in den Besitz schreibt** — Werkstück `T2` von Phase 5.6
// (#1467).
//
// # Warum es diesen Wächter gibt
//
// **Weil eine Zusage nur so viel wert ist wie der engste Weg an ihr vorbei.**
// Seit `T2` fragen [storage.Sammlungsspeicher.Aendern] und `…Loeschen` die
// offenen Zusagen, bevor sie schreiben. Das Issue verlangt, die **ganze
// Schreibklasse** zu schließen: Einzelbearbeitung, Massenpflege,
// Import/Merge, Löschen, Profilübernahme und Wiederherstellung.
//
// **Gemessen am 2026-09-18**: es gibt genau **drei** Anweisungen, die Zeilen
// der Sammlung schreiben, und alle drei stehen in
// `internal/storage/sammlung.go`. Massenpflege und Listeneinfuhr haben keinen
// eigenen SQL-Weg — sie gehen durch dieselbe Tür. Damit ist die Klasse
// geschlossen, **solange keine vierte dazukommt**, und genau das hält dieser
// Wächter.
//
// # Was er nicht hält
//
// **Dass die Prüfung an der richtigen Stelle steht.** Er sieht, wo geschrieben
// wird, nicht ob davor jemand gefragt hat; dafür gibt es
// `internal/storage/zusagenschutz_test.go`, und der fährt den Fall.
//
// **Und er sieht die Wiederherstellung nicht.** Sie tauscht die **Datei** und
// geht an jedem SQL vorbei; was danach mit offenen Zusagen ist, ist eine
// fachliche Frage und keine der Schreibwege.

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// besitzschreiber findet jede Anweisung, die Zeilen der Sammlung schreibt.
//
// **Auf die Tabelle genau.** `sammlung_herkunft` ist eine andere Tabelle und
// trägt keinen Besitz; ein Muster ohne die Wortgrenze zählte sie mit und wäre
// ein Wächter, der etwas anderes prüft, als sein Kopf sagt.
var besitzschreiber = regexp.MustCompile(
	`(?i)(INSERT INTO|UPDATE|DELETE FROM)\s+sammlung\b(?:\s|$|\x60)`)

// besitzschreibstellen sind die Dateien, die es dürfen.
var besitzschreibstellen = map[string]string{
	"internal/storage/sammlung.go": "der Weg der Bearbeitung — " +
		"Anlegen, Aendern und Loeschen, und die beiden letzten fragen " +
		"vorher die offenen Zusagen (T2, #1467)",

	// **Der zweite Weg, und er ist keine Aufweichung des ersten.** Ein
	// Handelsabschluss fragt die Zusagen nicht, ob er stören würde — er löst
	// sie ein: jeder Abgang muss von einer Zusage dieses Vorgangs mit
	// **bestätigter Übergabe** gedeckt sein, und deren Fassung muss gegen den
	// heutigen Stand der Besitzzeile rechnen (§7, Regel 2). Ohne beides wird
	// nicht gebucht. Zugänge berühren keine Zusage: sie nehmen nichts weg.
	"internal/storage/handelsbestand.go": "der Weg des Handels — Abgang nur " +
		"gegen eine bestätigte Übergabe passender Fassung, Zugang ohne " +
		"Preis (T6, #1471)",
}

func TestNurEinWegSchreibtInDenBesitz(t *testing.T) {
	t.Parallel()

	gefunden := map[string]int{}

	wurzel := filepath.Join("..", "..")

	if err := filepath.Walk(wurzel, func(pfad string, info os.FileInfo, err error) error {
		if err != nil {
			return nil //nolint:nilerr // eine unlesbare Stelle ist kein Schreibweg
		}

		// **Fremde Bäume gar nicht erst betreten.** Der Modulzwischenspeicher
		// unter `.go/` allein kostete beim ersten Lauf 36 Sekunden — ein
		// Wächter, der eine halbe Minute dauert, wird abgeschaltet.
		if info.IsDir() {
			if istFremd(info.Name()) {
				return filepath.SkipDir
			}

			return nil
		}

		if !strings.HasSuffix(pfad, ".go") || strings.HasSuffix(pfad, "_test.go") {
			return nil
		}

		roh, lesefehler := os.ReadFile(pfad) //nolint:gosec // der Baum dieses Repositoriums
		if lesefehler != nil {
			return nil //nolint:nilerr // dito
		}

		if treffer := besitzschreiber.FindAllString(string(roh), -1); treffer != nil {
			kurz, _ := filepath.Rel(wurzel, pfad) //nolint:errcheck // unter der Wurzel gefunden
			gefunden[filepath.ToSlash(kurz)] = len(treffer)
		}

		return nil
	}); err != nil {
		t.Fatalf("den Baum durchgehen: %v", err)
	}

	for datei, zahl := range gefunden {
		if _, erlaubt := besitzschreibstellen[datei]; !erlaubt {
			t.Errorf("%s schreibt %dmal in die Sammlung — trage die Stelle "+
				"in `besitzschreibstellen` ein und sag, wie sie die offenen "+
				"Zusagen prüft (T2, #1467)", datei, zahl)
		}
	}

	// **Die Gegenrichtung**: eine eingetragene Stelle, die nicht mehr
	// schreibt, ist eine Erlaubnis ohne Fall — und die nächste Änderung
	// glaubt, sie sei geprüft worden.
	for datei := range besitzschreibstellen {
		if gefunden[datei] == 0 {
			t.Errorf("%s steht in `besitzschreibstellen` und schreibt nicht "+
				"mehr — nimm die Zeile heraus", datei)
		}
	}
}

// istFremd sagt, ob ein Verzeichnis nicht zu diesem Quelltext gehört.
func istFremd(name string) bool {
	switch name {
	case ".go", ".git", "node_modules", "frontend", "tools", "saat", "testdaten":
		return true
	}

	return false
}
