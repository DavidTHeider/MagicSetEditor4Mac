package app_test

import (
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Die Abschrift hält gegen die Tabelle** — Werkstück U12 von Phase 5.1 (#1423).
//
// Dieselbe Regel wie beim Abschreiben: je Zeile der Fähigkeitentabelle jedes
// MCP-Werkzeug mit der Fähigkeit, den `mse`-Befehlen und den Wegen der Zeile.
func TestDieBefehlsverweiseStehenSoInDenFaehigkeiten(t *testing.T) {
	t.Parallel()

	roh, err := os.ReadFile("../../docs/FAEHIGKEITEN.md")
	if err != nil {
		t.Fatal(err)
	}

	werkzeugmuster := regexp.MustCompile("`([a-z]+\\.[a-z0-9_.]+)`")
	climuster := regexp.MustCompile("`(mse [^`]+)`")
	restmuster := regexp.MustCompile("`((?:GET|POST|PUT|DELETE|PATCH) [^`]+)`")

	erwartet := map[string]app.Befehlsverweis{}

	for _, zeile := range strings.Split(string(roh), "\n") {
		if !strings.HasPrefix(zeile, "| ") || strings.HasPrefix(zeile, "| Fähigkeit |") || strings.HasPrefix(zeile, "| ---") {
			continue
		}

		zellen := strings.Split(strings.Trim(strings.TrimSpace(zeile), "|"), " | ")
		if len(zellen) != 6 {
			continue
		}

		clis := make([]string, 0, 2)
		rests := make([]string, 0, 2)
		for _, treffer := range climuster.FindAllStringSubmatch(zellen[3], -1) {
			clis = append(clis, treffer[1])
		}

		for _, treffer := range restmuster.FindAllStringSubmatch(zellen[4], -1) {
			rests = append(rests, treffer[1])
		}

		for _, treffer := range werkzeugmuster.FindAllStringSubmatch(zellen[2], -1) {
			erwartet[treffer[1]] = app.Befehlsverweis{
				Faehigkeit: strings.ReplaceAll(strings.TrimSpace(zellen[0]), "**", ""),
				CLI:        strings.Join(clis, " · "),
				REST:       strings.Join(rests, " · "),
			}
		}
	}

	abschrift := app.Befehlsverweise()

	for name, verweis := range erwartet {
		if abschrift[name] != verweis {
			t.Errorf("%s: die Abschrift sagt %+v, die Tabelle %+v", name, abschrift[name], verweis)
		}
	}

	for name := range abschrift {
		if _, da := erwartet[name]; !da {
			t.Errorf("%s steht in der Abschrift, aber nicht in der Tabelle", name)
		}
	}

	if len(erwartet) < 150 {
		t.Errorf("nur %d Werkzeuge in der Tabelle gefunden — stimmt die Regel noch?", len(erwartet))
	}
}
