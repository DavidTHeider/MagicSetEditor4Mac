// Wer geändert hat, steht im Verlauf.
//
// **Weil zwei Prozesse schreiben** ([0010](../../docs/entscheidungen/0010-mcp-als-eigener-prozess.md)):
// das Fenster und der MCP-Server, auf derselben Datei. Eine Rücknahme, die nicht
// sagt, wessen Änderung sie umkehrt, ist im Zwei-Schreiber-Betrieb eine
// Vermutung (#52).
package storage_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestDerVerlaufHaeltFestWerGeaendertHat(t *testing.T) {
	faelle := []struct {
		name     string
		bauen    func(*storage.Kartenschreiber) *storage.Kartenschreiber
		erwartet domain.Urheber
	}{
		{
			// **Das Fenster ist die Voreinstellung**, weil es der Weg ist, den
			// jeder Tastendruck nimmt. Wer ihn eigens ansagen müsste, vergäße es
			// einmal — und dann stünde eine Tipparbeit unter fremdem Namen.
			name:     "ohne Ansage schreibt das Fenster",
			bauen:    func(k *storage.Kartenschreiber) *storage.Kartenschreiber { return k },
			erwartet: domain.UrheberFenster,
		},
		{
			name: "der Server sagt sich an",
			bauen: func(k *storage.Kartenschreiber) *storage.Kartenschreiber {
				return k.Als(domain.UrheberServer)
			},
			erwartet: domain.UrheberServer,
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			db, setID, karteID := bestandMitKarte(t)

			schreiber := fall.bauen(storage.NeuerKartenschreiber(db))
			if _, _, err := schreiber.FeldSetzenMitVorwert(
				probekontext(t), karteID, "name", "Eisball"); err != nil {
				t.Fatalf("setzen: %v", err)
			}

			verlauf, err := storage.NeuerVerlaufsleser(db).Verlauf(probekontext(t), setID)
			if err != nil {
				t.Fatalf("Verlauf: %v", err)
			}

			if len(verlauf) != 1 {
				t.Fatalf("%d Einträge, erwartet einer", len(verlauf))
			}

			if verlauf[0].Urheber != fall.erwartet {
				t.Errorf("Urheber %q, erwartet %q", verlauf[0].Urheber, fall.erwartet)
			}
		})
	}
}

// **`Als` gibt einen zweiten Schreiber, keinen umgestellten.**
//
// Ein Aufruf, der den Empfänger änderte, stellte jeden anderen Halter mit um —
// und im Composition Root hält `app` denselben Schreiber zweimal, einmal als
// `schreiber` und einmal als `massenschreiber`.
func TestAlsLaesstDenUrsprungInRuhe(t *testing.T) {
	db, setID, karteID := bestandMitKarte(t)

	fenster := storage.NeuerKartenschreiber(db)
	_ = fenster.Als(domain.UrheberServer)

	if _, _, err := fenster.FeldSetzenMitVorwert(
		probekontext(t), karteID, "name", "Eisball"); err != nil {
		t.Fatalf("setzen: %v", err)
	}

	verlauf, err := storage.NeuerVerlaufsleser(db).Verlauf(probekontext(t), setID)
	if err != nil {
		t.Fatalf("Verlauf: %v", err)
	}

	if verlauf[0].Urheber != domain.UrheberFenster {
		t.Errorf("Urheber %q — `Als` hat den Ursprung umgestellt", verlauf[0].Urheber)
	}
}
