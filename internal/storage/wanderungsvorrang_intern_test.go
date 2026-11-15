package storage

// **Eine Wanderung verhungert nicht unter überlappenden Rufen** — #2406,
// Befund 4 des Nachlaufs über Phase 2 von 0.9.0.
//
// Jeder Ruf an Dienst und Werkzeugserver hält die Sperre über die Wanderung
// geteilt, und eine Wanderung braucht sie allein. Beide fragen mit `LOCK_NB`
// im Takt; **ein wartender Alleinanspruch hinterlegte keinen Vorrang**, und
// überlappten sich kurze Rufe ohne Lücke, bekam die Wanderung die Sperre nie.
//
// **Ohne `t.Parallel()`**: der Lauf setzt [Wanderungsfrist] herab.

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEineWanderungVerhungertNichtUnterUeberlappendenRufen(t *testing.T) {
	vorher := Wanderungsfrist
	Wanderungsfrist = 2 * time.Second

	t.Cleanup(func() { Wanderungsfrist = vorher })

	db := bestandInEinerDatei(t)
	ctx, aufhoeren := context.WithCancel(probekontext(t))

	var (
		laeufe      sync.WaitGroup
		beantwortet atomic.Int64
		gescheitert atomic.Int64
	)

	// **Drei Rufer, versetzt um ein Drittel ihrer Dauer** — so hält immer
	// einer die geteilte Sperre, und es gibt keine Lücke.
	for versatz := range 3 {
		laeufe.Go(func() {
			time.Sleep(time.Duration(versatz) * 20 * time.Millisecond)

			for ctx.Err() == nil {
				frei, err := HalteSchemafreigabe(ctx, db)
				if err != nil {
					if ctx.Err() == nil {
						gescheitert.Add(1)
					}

					continue
				}

				time.Sleep(60 * time.Millisecond)
				frei()
				beantwortet.Add(1)
			}
		})
	}

	time.Sleep(150 * time.Millisecond)

	beginn := time.Now()
	freigeben, err := wanderungHalten(probekontext(t), db)
	gewartet := time.Since(beginn)

	if err == nil {
		freigeben()
	}

	aufhoeren()
	laeufe.Wait()

	if err != nil {
		t.Fatalf("die Wanderung verhungerte nach %v: %v (%d Rufe beantwortet)", gewartet, err, beantwortet.Load())
	}

	if gescheitert.Load() != 0 || beantwortet.Load() == 0 {
		t.Errorf("%d Rufe gescheitert, %d beantwortet", gescheitert.Load(), beantwortet.Load())
	}

	t.Logf("die Wanderung bekam die Sperre nach %v", gewartet)
}

// **Ohne Schranke wandert niemand und antwortet kein Ruf** — wie ohne die
// Sperre selbst ([TestOhneNachbardateiWandertNiemand]).
func TestOhneSchrankeWandertNiemand(t *testing.T) {
	t.Parallel()

	ort := filepath.Join(t.TempDir(), "bestand.db")

	db, err := Oeffnen(probekontext(t), ort)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = db.Close() }) //nolint:errcheck // nur zum Aufräumen

	if err := Migriere(probekontext(t), db, nil); err != nil {
		t.Fatal(err)
	}

	// **Die Wanderung eben legte die Schranke als Datei an** — an ihrer Stelle
	// steht jetzt ein Ordner.
	if err := os.Remove(ort + Eintrittsendung); err != nil {
		t.Fatal(err)
	}

	if err := os.Mkdir(ort+Eintrittsendung, 0o700); err != nil {
		t.Fatal(err)
	}

	if _, fehler := wanderungHalten(probekontext(t), db); !errors.Is(fehler, domain.ErrGesperrt) {
		t.Errorf("Wanderung ohne Schranke: %v, erwartet gesperrt", fehler)
	}

	if _, fehler := HalteSchemafreigabe(probekontext(t), db); !errors.Is(fehler, domain.ErrGesperrt) {
		t.Errorf("Freigabe ohne Schranke: %v, erwartet gesperrt", fehler)
	}
}
