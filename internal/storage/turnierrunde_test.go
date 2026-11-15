package storage_test

// **Zwei Rufe legen nicht zweimal dieselbe Runde** — Befund 8 des Zweitblicks
// über Phase 4.7
// (#1294).
//
// # Der Befund
//
// Der Dienst liest den Stand, rechnet den Plan und schreibt ihn. Die
// Transaktion ermittelte nur `MAX(nummer)+1` und prüfte den gelesenen Stand
// **nicht noch einmal**. Zwei Rufe konnten denselben abgeschlossenen Stand
// lesen; nach dem Festschreiben des ersten legte der zweite seinen bereits
// gerechneten, gleichen Plan als **nächste** Runde ab.
//
// **Das lokale `useVorgang` hilft dagegen nicht**: es sperrt einen zweiten
// Ruf derselben Fenstersitzung, nicht einen zweiten Prozess, ein zweites
// Fenster oder Kommandozeile, MCP und REST.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Wer eine Runde legt, sagt, die wievielte er meint.**
func TestEineZweiteRundeAusDemselbenStandWirdAbgewiesen(t *testing.T) {
	ablage := turnierablage(t)

	turnier := turnierMitTeilnehmern(t, ablage, "Ada", "Bert")

	paarungen := []domain.Paarung{
		{A: turnier.Teilnehmer[0].ID, B: turnier.Teilnehmer[1].ID},
	}

	// **Beide rechnen mit derselben Ausgangslage**: die erste Runde.
	if _, err := ablage.RundeLegen(probekontext(t), turnier.ID, 1, paarungen); err != nil {
		t.Fatalf("die erste Runde: %v", err)
	}

	_, err := ablage.RundeLegen(probekontext(t), turnier.ID, 1, paarungen)
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("Fehler ist %v, erwartet ErrConflict", err)
	}

	gelesen, lesefehler := ablage.Turnier(probekontext(t), turnier.ID)
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if len(gelesen.Runden) != 1 {
		t.Errorf("%d Runden im Bestand — erwartet eine", len(gelesen.Runden))
	}
}

// **Und die richtige Nummer geht durch.**
func TestDieNaechsteRundeGehtDurch(t *testing.T) {
	ablage := turnierablage(t)

	turnier := turnierMitTeilnehmern(t, ablage, "Ada", "Bert")

	paarungen := []domain.Paarung{
		{A: turnier.Teilnehmer[0].ID, B: turnier.Teilnehmer[1].ID},
	}

	for nummer := 1; nummer <= 3; nummer++ {
		gelegt, err := ablage.RundeLegen(probekontext(t), turnier.ID, nummer, paarungen)
		if err != nil {
			t.Fatalf("Runde %d: %v", nummer, err)
		}

		if gelegt.Nummer != nummer {
			t.Errorf("die Runde trägt die Nummer %d, erwartet %d",
				gelegt.Nummer, nummer)
		}
	}
}
