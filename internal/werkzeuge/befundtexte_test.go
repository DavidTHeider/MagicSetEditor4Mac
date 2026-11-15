package werkzeuge_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// **Ein Befundtext nennt keine Auszeichnung, ohne sie zu erklären**
// (#675).
//
// # Warum es diesen Wächter gibt
//
// Die Anwendung soll auch für Einsteiger zu bedienen sein. Ein Befund wie
//
//	`<sym>` steht im Text, aber das Feld nennt keine Symbolschrift
//
// setzt voraus, dass jemand weiß, was `<sym>` ist — und wer das weiß, braucht
// den Befund nicht. §9 zieht die Grenze: was aus einem **Paket** kommt, bleibt
// stehen; was **dieses Programm** sagt, ist für Menschen.
//
// # Was er prüft, und was nicht
//
// Er sucht in den Meldungstexten von `service` und `render` nach einer
// spitzen Klammer vor einem Buchstaben oder Schrägstrich — der Form einer
// MSE-Auszeichnung.
//
// **Er prüft nicht, ob eine Meldung verständlich ist.** Das kann er nicht;
// dafür steht dieser Satz hier, und dafür liest jemand sie.
//
// **Und er prüft nicht die Skriptfehler.** `msescript` meldet, was an einem
// Ausdruck falsch ist — dort ist die Sprache der Gegenstand, nicht das
// Beiwerk.
func TestKeinBefundtextNenntEineAuszeichnung(t *testing.T) {
	t.Parallel()

	auszeichnung := regexp.MustCompile(`"[^"]*<[a-zA-Z/][^"]*"`)

	var gefunden []string

	for _, paket := range []string{"internal/service", "internal/render"} {
		eintraege, err := os.ReadDir(filepath.Join("..", "..", paket))
		if err != nil {
			t.Fatalf("%s lesen: %v", paket, err)
		}

		for _, eintrag := range eintraege {
			name := eintrag.Name()
			if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
				continue
			}

			pfad := filepath.Join("..", "..", paket, name)

			roh, err := os.ReadFile(pfad) // #nosec G304 -- eigener Baum
			if err != nil {
				t.Fatalf("%s lesen: %v", pfad, err)
			}

			for nummer, zeile := range strings.Split(string(roh), "\n") {
				if !strings.Contains(zeile, "Meldung") || strings.HasPrefix(strings.TrimSpace(zeile), "//") {
					continue
				}

				if auszeichnung.MatchString(zeile) {
					gefunden = append(gefunden,
						filepath.Join(paket, name)+":"+itoa(nummer+1)+" — "+strings.TrimSpace(zeile))
				}
			}
		}
	}

	if len(gefunden) > 0 {
		t.Errorf("%d Befundtexte nennen eine Auszeichnung:\n      %s",
			len(gefunden), strings.Join(gefunden, "\n      "))
	}
}

// itoa spart den Import von `strconv` für eine Zeilennummer.
func itoa(zahl int) string {
	if zahl == 0 {
		return "0"
	}

	ziffern := ""
	for zahl > 0 {
		ziffern = string(rune('0'+zahl%10)) + ziffern
		zahl /= 10
	}

	return ziffern
}
