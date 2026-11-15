package werkzeuge

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// lokaleEinbindung trifft `@import './name.css';` — ein Blatt aus demselben
// Verzeichnis. `@import 'tailwindcss'` und Pakete trifft es nicht.
var lokaleEinbindung = regexp.MustCompile(`@import '\./([^']+)';`)

// StilblattMitEinbindungen gibt `frontend/src/stil.css` samt den lokalen
// Blättern zurück, die es einbindet — so, wie Tailwind es liest.
//
// **Seit AP-2** (#2359)
// stehen Merkmale auch in `erscheinungsmerkmale.css`. Ein Wächter, der nur
// `stil.css` las, sah `text-hinweis` danach nicht mehr als Farbe — und
// `TestJedeVordergrundfarbeStehtInKontraste` verlangte für sie keine Zeile
// mehr, **grün und still schwächer**. Exportiert, weil
// `farbklassen_test.go` im äußeren Testpaket steht.
func StilblattMitEinbindungen(t *testing.T) string {
	t.Helper()

	return mitEinbindungen(t, filepath.Join("..", "..", "frontend", "src"), "stil.css")
}

func mitEinbindungen(t *testing.T, ordner, datei string) string {
	t.Helper()

	inhalt, err := os.ReadFile(filepath.Join(ordner, datei)) //nolint:gosec // Pfad aus dem Baum
	if err != nil {
		t.Fatalf("%s nicht lesbar: %v", datei, err)
	}

	return lokaleEinbindung.ReplaceAllStringFunc(string(inhalt), func(treffer string) string {
		return mitEinbindungen(t, ordner, lokaleEinbindung.FindStringSubmatch(treffer)[1])
	})
}

func TestStilblattLiestSeineEinbindungenMit(t *testing.T) {
	t.Parallel()

	stil := StilblattMitEinbindungen(t)

	for _, merkmal := range []string{"--color-grund:", "--color-hinweis:", "--color-auswahlkante:"} {
		if !strings.Contains(stil, merkmal) {
			t.Errorf("%s fehlt im eingebundenen Stilblatt", merkmal)
		}
	}

	if strings.Contains(stil, "@import './") {
		t.Error("eine lokale Einbindung blieb stehen")
	}
}
