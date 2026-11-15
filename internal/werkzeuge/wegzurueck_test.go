// Der Wächter über den Weg zurück —
// Werkstück 5 von Phase 4.13
// (#1352).
//
// # Die Zusage
//
// **Wo ein Programm den Bestand öffnet oder migriert und dabei scheitert,
// steht daneben, wie es weitergeht.** Das Fenster sagt es auf der
// Wiederherstellungsseite ([app.Startbefund] trägt den Rückweg); die drei
// Flächen ohne Fenster sagen es als Zeile, über `app.MitWegZurueck`.
//
// **Warum es einen Wächter braucht**: die Stelle steht fünfmal da — zweimal
// in der Kommandozeile, je zweimal in den beiden Diensten. Eine sechste
// entsteht mit der nächsten Fläche, und sie fiele still aus: der Fehler
// bleibt richtig, er sagt nur nichts mehr über den Weg zurück.
//
// # Was er nicht prüft
//
// **Ob der Satz stimmt.** Dass er die Datei nennt, die diese Fassung wirklich
// öffnen kann, prüft der Flächenlauf gegen das gebaute Programm
// (`internal/flaechen/wegzurueck_test.go`); dass die Auswahl richtig ist,
// prüfen die Läufe in `internal/app`.
//
// **Und er sieht ein Textfenster, keinen Kontrollfluss.** Ein `MitWegZurueck`
// zwölf Zeilen unter dem Öffnen fände er nicht — dafür fände eine AST-Regel
// die Fälle nicht, in denen der Fehler in einer eigenen Zeile geprüft wird.
package werkzeuge

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// bestandsgriffe sind die Aufrufe, nach denen ein Weg zurück stehen muss.
var bestandsgriffe = []string{"storage.Oeffnen(", "storage.Migriere("}

// sichtweite ist, wie viele Zeilen nach dem Griff der Wächter ansieht.
//
// **Zehn**, weil die Behandlung unmittelbar darunter steht: der Aufruf, die
// Prüfung, ein Kommentar von drei Zeilen und die Rückgabe.
const sichtweite = 10

func TestJedeFlaecheNenntDenWegZurueck(t *testing.T) {
	t.Parallel()

	var ohne []string

	for _, datei := range gowurzeln(t, "../../cmd") {
		roh, err := os.ReadFile(datei) //nolint:gosec // Pfad aus dem Baum selbst
		if err != nil {
			t.Fatalf("%s: %v", datei, err)
		}

		zeilen := strings.Split(string(roh), "\n")

		for nummer, zeile := range zeilen {
			if !einBestandsgriff(zeile) {
				continue
			}

			bis := min(nummer+sichtweite, len(zeilen))
			if !strings.Contains(strings.Join(zeilen[nummer:bis], "\n"),
				"MitWegZurueck") {
				ohne = append(ohne,
					filepath.Base(datei)+":"+strings.TrimSpace(zeile))
			}
		}
	}

	if len(ohne) > 0 {
		t.Errorf("diese Stellen öffnen den Bestand, ohne den Weg zurück zu "+
			"nennen — ein Fehler dort lässt jemanden stehen:\n%s",
			strings.Join(ohne, "\n"))
	}
}

// einBestandsgriff sagt, ob eine Zeile den Bestand öffnet oder migriert.
func einBestandsgriff(zeile string) bool {
	for _, griff := range bestandsgriffe {
		if strings.Contains(zeile, griff) {
			return true
		}
	}

	return false
}

// gowurzeln sammelt die Go-Dateien unter einem Verzeichnis — ohne Tests.
func gowurzeln(t *testing.T, ort string) []string {
	t.Helper()

	var gefunden []string

	fehler := filepath.Walk(ort, func(pfad string, angabe os.FileInfo, unterwegs error) error {
		if unterwegs != nil {
			return unterwegs
		}

		if !angabe.IsDir() && strings.HasSuffix(pfad, ".go") &&
			!strings.HasSuffix(pfad, "_test.go") {
			gefunden = append(gefunden, pfad)
		}

		return nil
	})
	if fehler != nil {
		t.Fatalf("%s durchgehen: %v", ort, fehler)
	}

	return gefunden
}
