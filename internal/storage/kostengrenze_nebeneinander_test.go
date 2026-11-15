package storage_test

// **Zwei gleichzeitige Aufträge können die Grenze nicht überschreiten** —
// Werkstück A5 von Phase 5.3
// (#1444).
//
// # Warum es diesen Lauf gibt
//
// **Die Zusage stand seit O12 da und war nie gemessen.** `passtInsBudget`
// prüft in **derselben** Transaktion, in der gebucht wird — das ist die
// richtige Form, und sie trägt nur, solange zwei Transaktionen einander nicht
// überholen können.
//
// # Was `_txlock=immediate` wirklich kauft
//
// **Nicht die Grenze — die Meldung.** Das war die Erwartung beim Schreiben
// dieses Laufs, und die Gegenprobe hat sie widerlegt: dieselbe Datei ohne
// `_txlock=immediate`, am 2026-09-11 gemessen, ließ **einen von dreien**
// durch statt zweier. Die Grenze hielt auch dort — die beiden anderen fielen
// mit `SQLITE_BUSY_SNAPSHOT` um, also mit einem Datenbankfehler, wo einer von
// ihnen hätte gebucht werden dürfen.
//
// **Deshalb prüft dieser Lauf beides**: dass nicht zu viele durchkommen, und
// dass der eine, der abgewiesen wird, es **wegen der Grenze** wird. Die
// zweite Zusicherung ist die, die ohne die Schranke umfällt — und sie ist
// nicht vom Zufall abhängig, weil mit der Schranke jeder Lauf wartet, statt
// sich zu stoßen.

import (
	"database/sql"
	"errors"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// Die Zahlen, gegen die die Läufe drücken.
//
// **Drei Aufträge zu je 400 gegen 1.000**: zwei passen, der dritte nicht.
// Eine Grenze, die ein Vielfaches der Auftragshöhe wäre, ließe jede
// Reihenfolge durch und prüfte nichts.
const (
	dieGrenze   int64 = 1000
	jeAuftrag   int64 = 400
	dieLaeufe         = 3
	dieWaehrung       = "EUR"
)

func TestZweiGleichzeitigeAuftraegeSprengenDieGrenzeNicht(t *testing.T) {
	t.Parallel()

	ort := filepath.Join(t.TempDir(), "bestand.db")

	erste := verbindung(t, ort)

	if err := storage.Migriere(probekontext(t), erste, nil); err != nil {
		t.Fatalf("migrieren: %v", err)
	}

	if fehler := storage.NeuesKostenjournal(erste).BudgetSetzen(probekontext(t),
		1, "anbieter", domain.ZeitraumTag,
		domain.Betrag{Wert: dieGrenze, Waehrung: dieWaehrung}, time.Now(),
	); fehler != nil {
		t.Fatalf("das Budget setzen: %v", fehler)
	}

	// **Drei eigene Verbindungen** — eine geteilte liefe reihum, und der Lauf
	// prüfte eine Reihenfolge statt einer Gleichzeitigkeit.
	geglueckt, fehler := gleichzeitigReservieren(t,
		[]*sql.DB{erste, verbindung(t, ort), verbindung(t, ort)})

	// **Zwei von dreien, und keiner mehr.** Ohne die Zahl wäre der Lauf auch
	// dann grün, wenn gar keiner durchgekommen wäre.
	if geglueckt != 2 {
		t.Errorf("%d Aufträge glückten, erwartet 2 — zusammen %d von %d",
			geglueckt, int64(geglueckt)*jeAuftrag, dieGrenze)
	}

	if len(fehler) != 1 {
		t.Fatalf("%d Aufträge scheiterten, erwartet einer: %v", len(fehler), fehler)
	}

	// **Und der eine scheitert an der Grenze, nicht an der Datenbank.** Das
	// ist die Zusicherung, die ohne `_txlock=immediate` umfällt.
	if !errors.Is(fehler[0], domain.ErrConflict) {
		t.Errorf("der abgewiesene Auftrag scheiterte an %v — erwartet war die Grenze",
			fehler[0])
	}
}

// gleichzeitigReservieren lässt drei Läufe zugleich buchen.
//
// **Alle drei starten am selben Signal.** Ohne es liefe der erste durch,
// bevor der zweite beginnt.
func gleichzeitigReservieren(t *testing.T, verbindungen []*sql.DB) (int, []error) {
	t.Helper()

	var (
		los     = make(chan struct{})
		warten  sync.WaitGroup
		schloss sync.Mutex
		anzahl  int
		fehler  []error
	)

	for lauf := range dieLaeufe {
		warten.Add(1)

		go func() {
			defer warten.Done()

			<-los

			_, gescheitert := storage.NeuesKostenjournal(verbindungen[lauf]).Reservieren(
				probekontext(t), domain.Kostenposten{
					Kennung:       "auftrag-" + string(rune('a'+lauf)),
					ProfilID:      1,
					Konto:         "anbieter",
					Aufgabe:       "karte.text",
					Hoechstbetrag: domain.Betrag{Wert: jeAuftrag, Waehrung: dieWaehrung},
				}, domain.ZeitraumTag, time.Now())

			schloss.Lock()
			defer schloss.Unlock()

			if gescheitert != nil {
				fehler = append(fehler, gescheitert)

				return
			}

			anzahl++
		}()
	}

	close(los)
	warten.Wait()

	return anzahl, fehler
}
