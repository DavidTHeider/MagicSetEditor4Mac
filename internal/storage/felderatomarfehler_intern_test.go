package storage

import (
	"database/sql"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Fehlerwege des Schreibens mehrerer Felder** —
// #714.
//
// Paketintern und über den Stolpertreiber, wie bei den übrigen
// Transaktionswegen. **Und sie gehören geprüft, weil ein halber Verbund eine
// Karte hinterließe, die niemand gemeint hat** — der Regeltext neu, der
// Flavortext alt.
func TestDasSchreibenMehrererFelderMeldetSeineFehlerwege(t *testing.T) {
	faelle := map[string]string{
		"die Transaktion beginnt nicht": "",
		"der Commit scheitert":          "stolper-feld-commit",
	}

	for name, treiber := range faelle {
		t.Run(name, func(t *testing.T) {
			zweiterSchreibzaehler = 0

			db := geschlosseneDatenbank(t)

			if treiber != "" {
				geoeffnet, err := sql.Open(treiber, "")
				if err != nil {
					t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
				}

				t.Cleanup(func() { _ = geoeffnet.Close() })

				db = geoeffnet
			}

			if _, _, err := NeuerKartenschreiber(db).FelderSetzenMitVorwert(
				probekontext(t), 1, []domain.Value{{Field: "name", Text: "Eisball"}},
			); err == nil {
				t.Errorf("%s: kein Fehler", name)
			}
		})
	}
}
