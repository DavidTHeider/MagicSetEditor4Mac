package werkzeuge

// **Die Wurzeln einer Naht kommen aus einer Regel** — FMT-02, S10.NAT-03
// (#1691).
//
// `msepaket.Paketwurzeln` nennt nur Orte, **an denen Pakete liegen**. Ein
// frisches Paketverzeichnis trägt keines; wer seine Wurzeln damit beim Start
// festlegt, sieht jedes später angelegte Paket erst nach einem Neustart. Die
// Regel dagegen — der verwaltete Ort gilt immer — steht in
// `app/paketorte.go` (`paketwurzelnFuer`) und im Composition Root
// (`app/verdrahtung.go`, mit seinen Nähten für Vorlagen- und Saatort). Sechs
// Nähte riefen `Paketwurzeln` daneben ohne sie; gemessen hat es die
// Hüllenprobe `fmt02f`.
//
// **Seit PERF-02 (#2169) auch über die Wurzelsuche**: `paketorte.go` ruft
// `wurzelsuche.Paketwurzeln` — dieselbe Regel, gemerkt. Eine zweite
// Wurzelsuche anderswo wäre eine zweite Stelle, und die zählt er mit.
//
// # Was er nicht prüft
//
// **Ob ein Aufrufer die Regel an anderer Stelle nachbaut** — etwa die Wurzeln
// von Hand zusammensetzt. Er sieht den Aufruf, nicht die Absicht.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPaketwurzelnKommenAusEinerRegel(t *testing.T) {
	erlaubt := map[string]bool{"paketorte.go": true, "verdrahtung.go": true}

	dateien, err := filepath.Glob(filepath.Join("..", "app", "*.go"))
	if err != nil {
		t.Fatal(err)
	}

	gesehen := 0

	for _, datei := range dateien {
		if strings.HasSuffix(datei, "_test.go") {
			continue
		}

		roh, err := os.ReadFile(datei) // #nosec G304 -- Quelltext dieses Repositoriums
		if err != nil {
			t.Fatal(err)
		}

		quelle := string(roh)
		if !strings.Contains(quelle, "msepaket.Paketwurzeln(") &&
			!strings.Contains(quelle, "wurzelsuche.Paketwurzeln(") &&
			!strings.Contains(quelle, "msepaket.NeueWurzelsuche(") {
			continue
		}

		gesehen++

		if !erlaubt[filepath.Base(datei)] {
			t.Errorf("%s ruft msepaket.Paketwurzeln selbst — der verwaltete Ort fiele heraus, "+
				"solange er leer ist; über paketwurzelnFuer gehen", datei)
		}
	}

	// **Die Gegenrichtung**: ohne die beiden erlaubten Stellen hätte der
	// Wächter nichts gesehen und wäre trotzdem grün.
	if gesehen != len(erlaubt) {
		t.Errorf("%d Dateien rufen Paketwurzeln, erwartet %d — ist die Regel umgezogen?", gesehen, len(erlaubt))
	}
}
