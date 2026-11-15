package service_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Eine unveränderte Zeile eines Felds mit Leerzeichen im Namen bleibt, wie
// sie war** — gefunden an
// #2298.
//
// Die Datei schreibt `rule_text:`, der Bestand führt `rule text`. Suchte der
// Schreiber den Bestandswert mit dem Schlüssel der Datei, fand er keinen,
// warf die Zeile weg und setzte sie aus dem Bestand neu zusammen — mit
// anderen Abständen und ohne den Kommentar darunter. Bei `extra_data:` kam
// so die flache Fassung eines älteren Bestands in die Datei.
func TestEineZeileMitLeerzeichenImFeldnamenBehaeltIhreBytes(t *testing.T) {
	const quelle = "mse version: 2.0.0\n" +
		"game: probe\n" +
		"card:\n" +
		"\tname: Nebelhüterin\n" +
		"\trule_text:   Fliegend\n" +
		"\t# ein Kommentar\n" +
		"\tcasting_cost: 2RR\n"

	bestand := quellenDouble{
		leserDouble: leserDouble{
			koepfe: []domain.SetHead{{ID: 7, Name: "Sternengarten", Game: "probe"}},
			karten: []*domain.CardRow{{ID: 1, Number: 0, Values: []domain.Value{
				{Field: "name", Text: "Nebelhüterin"},
				{Field: "rule text", Text: "Fliegend"},
				{Field: "casting cost", Text: "2RR"},
			}}},
		},
		quelle: &domain.Quelle{Datei: "set", Inhalt: []byte(quelle)},
	}

	ort := filepath.Join(t.TempDir(), "wieder.mse-set")

	if _, err := service.NeuerSetschreiber(bestand, &ortsschreiber{}).
		Schreibe(t.Context(), 7, ort); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	gelesen, lesefehler := os.ReadFile(filepath.Join(ort, "set")) // #nosec G304 -- aus t.TempDir()
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if string(gelesen) != quelle {
		t.Errorf("Abweichung:\n%q\nerwartet\n%q", gelesen, quelle)
	}
}
