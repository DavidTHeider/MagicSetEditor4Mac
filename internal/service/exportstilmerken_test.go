package service_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein Stil je Name und Lauf, nicht je Karte** — E04
// (#2241).
//
// Das Profil von `magic-namecheck` über Black (3.788 Karten, 28,6 s gegen
// MSE2s 6,2 s) zeigte das Stilpaket je Karte geöffnet und geschlossen: `mitStilen`
// fragte die Stilquelle für jede Karte, deren `stylesheet` gesetzt ist. Die
// Feldberechnung merkte ihre Antworten schon (Befund 3 eines früheren
// Zweitblicks) — behoben war der Fall, nicht die Klasse.
func TestDieAusfuhrFragtJedenStilEinmal(t *testing.T) {
	t.Parallel()

	gefragt := map[string]int{}
	stile := func(paket string) (msepaket.Stil, bool) {
		gefragt[paket]++
		if paket == "fehlt" {
			return msepaket.Stil{}, false
		}

		return msepaket.Stil{Paket: "probe-" + paket + ".mse-style", VollerName: "Stil " + paket, Breite: 375, Hoehe: 523, DPI: 150}, true
	}

	karten := make([]*domain.CardRow, 0, 40)
	for i := range 40 {
		stil := []string{"eigen", "anders", "fehlt"}[i%3]
		karten = append(karten, &domain.CardRow{Values: []domain.Value{
			{Field: "name", Text: fmt.Sprintf("Karte %d", i)}, {Field: "stylesheet", Text: stil},
		}})
	}

	ausgabe, befunde, err := probelauf(service.Ausfuhrmittel{Stile: stile}).Fuehre(t.Context(),
		msepaket.Exportvorlage{Skript: `to_text(for each card in cards do (if card.stylesheet == nil then "-" else card.stylesheet.full_name) + ";") + set.stylesheet.full_name`},
		karten, nil, nil)
	if err != nil {
		t.Fatalf("ausführen: %v — %v", err, befunde)
	}

	for name, anzahl := range gefragt {
		if anzahl != 1 {
			t.Errorf("die Stilquelle wurde nach %q %d-mal gefragt, erwartet einmal", name, anzahl)
		}
	}

	if !strings.HasPrefix(ausgabe, "Stil eigen;Stil anders;-;") || !strings.HasSuffix(ausgabe, "Stil ") {
		t.Errorf("die Ausgabe hat sich verändert: %q", ausgabe)
	}

	// **Der Befund bleibt je Karte** — ein gemerkter Stil darf ihn nicht
	// verschlucken: jede Karte mit fehlendem Stil nennt sich.
	fehlend := 0
	for _, befund := range befunde {
		if strings.Contains(befund.Meldung, `"fehlt"`) {
			fehlend++
		}
	}

	if fehlend != 13 {
		t.Errorf("%d Befunde für den fehlenden Stil, erwartet 13 (einer je Karte)", fehlend)
	}
}
