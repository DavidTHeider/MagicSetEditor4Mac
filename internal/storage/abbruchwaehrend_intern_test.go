package storage

// **Ein Abbruch während der Abfrage kommt als Abbruch an** — Nahtlauf vom
// 2026-09-07.
//
// # Der Befund
//
// Zwei Kartenvorschauen zugleich: die ältere wird abgebrochen (`AGENTS.md` §7,
// die Sonderregel für die Vorschau). Der Nahtlauf erwartete dafür den Code
// `abgebrochen` und bekam **`unbekannt`** — im Fenster steht dann „Es ist ein
// unerwarteter Fehler aufgetreten", wo gar nichts unerwartet ist.
//
// Im Protokoll stand der Grund: `Karte lesen: sqlite3: interrupted`. SQLite
// bricht eine **laufende** Abfrage über `sqlite3_interrupt` ab und gibt seinen
// eigenen Fehler zurück, nicht `context.Canceled`.
//
// # Warum der erste Versuch, es nachzustellen, grün war
//
// **Ein Kontext, der schon abgebrochen ist, erreicht den Treiber gar nicht.**
// `database/sql` prüft ihn vorher und gibt `context.Canceled` selbst zurück —
// eine Probe mit `abbrechen()` **vor** der Abfrage sagt über diesen Weg
// nichts. Sie war grün und maß nichts (§10).
//
// **Dieser Test bricht deshalb erst ab, während die Abfrage läuft.** Damit sie
// lange genug läuft, zählt sie über eine rekursive Reihe — ohne Tabelle und
// ohne Bestand.
//
// **Ein interner Test**, weil er den echten Treiberfehler durch [einordnen]
// schickt: geprüft wird nicht, was SQLite tut, sondern was **dieses Programm**
// daraus macht.

import (
	"context"
	"errors"
	"testing"
	"time"

	sqlite "github.com/ncruces/go-sqlite3"
)

// **Der Treiberfehler wird zu einem Abbruch — und behält seinen Grund.**
func TestEinAbbruchWaehrendDerAbfrageWirdEingeordnet(t *testing.T) {
	roh := abgebrochenerTreiberfehler(t)

	eingeordnet := einordnen(roh)

	if !errors.Is(eingeordnet, context.Canceled) {
		t.Errorf("eingeordnet als %v — kein Abbruch. Im Fenster steht dann "+
			"„ein unerwarteter Fehler\", wo nichts unerwartet ist",
			eingeordnet)
	}

	// **Der Grund reist mit** (§8): wer im Protokoll nachsieht, findet
	// weiterhin `sqlite3: interrupted`.
	if !errors.Is(eingeordnet, roh) {
		t.Errorf("der Treiberfehler ist aus der Kette gefallen: %v", eingeordnet)
	}
}

// abgebrochenerTreiberfehler holt einen echten `interrupted` vom Treiber.
//
// **Kein nachgebauter.** Ein Fehler, den der Test selbst schreibt, prüfte,
// dass `einordnen` auf das reagiert, was der Test annimmt — und nicht auf das,
// was der Treiber gibt (§10: die Erinnerung eines Werkzeugs ist kein
// Prüfergebnis).
func abgebrochenerTreiberfehler(t *testing.T) error {
	t.Helper()

	// **Über [Oeffnen] und nicht über `sql.Open`** — der Verbindungshaken
	// setzt die Pragmas, und ohne ihn prüfte dieser Test eine Verbindung, die
	// es im Betrieb nicht gibt. Migriert wird nicht: die Abfrage unten fasst
	// keine Tabelle an.
	db, oeffnungsfehler := Oeffnen(t.Context(), ImArbeitsspeicher)
	if oeffnungsfehler != nil {
		t.Fatalf("öffnen: %v", oeffnungsfehler)
	}

	t.Cleanup(func() {
		if fehler := db.Close(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	ctx, abbrechen := context.WithCancel(t.Context())

	// **Abgebrochen wird, während gerechnet wird.** Die Frist ist großzügig
	// genug, dass die Abfrage sicher läuft, und kurz genug, dass sie sicher
	// nicht fertig ist.
	go func() {
		time.Sleep(50 * time.Millisecond)
		abbrechen()
	}()

	var zahl int64

	// **Eine Reihe über zehn Millionen Zeilen** — sie braucht Sekunden und
	// fasst nichts an.
	fehler := db.QueryRowContext(ctx, `
		WITH RECURSIVE reihe(n) AS (
			SELECT 1 UNION ALL SELECT n + 1 FROM reihe WHERE n < 10000000
		)
		SELECT count(*) FROM reihe`).Scan(&zahl)

	if fehler == nil {
		t.Fatalf("die Abfrage lief durch und gab %d — sie war zu schnell, "+
			"und dieser Test misst nichts", zahl)
	}

	// **Der Code wird genannt und nicht verworfen.** `ExtendedErrorCode`
	// erfüllt `error`, und ein `_` darauf ist für `errcheck` ein verworfener
	// Fehler — zu Recht: er sagt hier, **welcher** Abbruch es war.
	code, ist := sqlitecode(fehler)
	if !ist {
		t.Fatalf("%v trägt keinen SQLite-Code — dieser Test misst etwas "+
			"anderes als den Abbruch", fehler)
	}

	if code.Code() != sqlite.INTERRUPT {
		t.Fatalf("der Treiber gab %v statt INTERRUPT — der Abbruch kam nicht "+
			"an, und dieser Test misst etwas anderes", code)
	}

	return fehler
}
