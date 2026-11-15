package storage

// **Zwischen der Frage nach dem Stand und dem Schreiben wandert niemand** —
// Zweitblick über Phase 2 von 0.9.0, Befund 2 (#2323): die Fassungswache
// fragte vor einem Ruf, und eine Wanderung konnte festschreiben, bevor der
// Ruf schrieb. Die Freigabe hält dafür die geteilte Sperre über die Wanderung,
// und die Wanderung braucht sie allein.
//
// **Ohne `t.Parallel()`**, wo ein Lauf [Wanderungsfrist] herabsetzt.

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// bestandInEinerDatei legt einen heutigen Bestand in einer Datei an.
func bestandInEinerDatei(t *testing.T) *sql.DB {
	t.Helper()

	db, err := Oeffnen(probekontext(t), filepath.Join(t.TempDir(), "bestand.db"))
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = db.Close() }) //nolint:errcheck // nur zum Aufräumen

	if err := Migriere(probekontext(t), db, nil); err != nil {
		t.Fatal(err)
	}

	return db
}

func TestWaehrendDerFreigabeWandertNiemand(t *testing.T) {
	vorher := Wanderungsfrist
	Wanderungsfrist = 150 * time.Millisecond

	t.Cleanup(func() { Wanderungsfrist = vorher })

	db := bestandInEinerDatei(t)
	weitere, _ := kuenftigeWanderung(t, "kuenftig", "CREATE TABLE kuenftig (wert INTEGER) STRICT;")

	frei, err := HalteSchemafreigabe(probekontext(t), db)
	if err != nil {
		t.Fatalf("die Freigabe über dem eigenen Stand: %v", err)
	}

	if fehler := MigriereMit(probekontext(t), db, weitere, nil); !errors.Is(fehler, domain.ErrBelegt) {
		t.Errorf("unter der Freigabe wanderte es: %v", fehler)
	}

	frei()

	if fehler := MigriereMit(probekontext(t), db, weitere, nil); fehler != nil {
		t.Fatalf("nach der Freigabe: %v", fehler)
	}

	// **Und danach sagt die Freigabe nein** — mit der neueren Fassung.
	var neuer *domain.NeuereFassung

	if _, fehler := HalteSchemafreigabe(probekontext(t), db); !errors.As(fehler, &neuer) {
		t.Errorf("über dem neueren Bestand: %v", fehler)
	}
}

// **Eine Freigabe wartet, solange gewandert wird.**
func TestDieFreigabeWartetAufDieWanderung(t *testing.T) {
	vorher := Wanderungsfrist
	Wanderungsfrist = 150 * time.Millisecond

	t.Cleanup(func() { Wanderungsfrist = vorher })

	db := bestandInEinerDatei(t)

	wandert, err := wanderungHalten(probekontext(t), db)
	if err != nil {
		t.Fatal(err)
	}

	if _, fehler := HalteSchemafreigabe(probekontext(t), db); !errors.Is(fehler, domain.ErrBelegt) {
		t.Errorf("während der Wanderung: %v, erwartet belegt", fehler)
	}

	wandert()

	frei, fehler := HalteSchemafreigabe(probekontext(t), db)
	if fehler != nil {
		t.Fatalf("danach: %v", fehler)
	}

	frei()
}

// **Wer den Stand nicht erfährt, schreibt nicht** — Befund 3: hier galt ein
// Lesefehler als Zustimmung.
func TestOhneStandKeineFreigabe(t *testing.T) {
	t.Parallel()

	db, err := Oeffnen(probekontext(t), filepath.Join(t.TempDir(), "bestand.db"))
	if err != nil {
		t.Fatal(err)
	}

	_ = db.Close() //nolint:errcheck // geschlossen gefragt

	if frei, fehler := HalteSchemafreigabe(probekontext(t), db); fehler == nil || frei != nil {
		t.Errorf("über einer geschlossenen Verbindung: %v", fehler)
	}
}

// **Ohne Nachbardatei keine Freigabe** — wie bei der Wanderung (Befund 4).
func TestOhneNachbardateiKeineFreigabe(t *testing.T) {
	t.Parallel()

	db := bestandInEinerDatei(t)
	ort, _ := dateiVon(probekontext(t), db) //nolint:errcheck // eben migriert

	if err := os.Remove(wanderungspfad(ort)); err != nil {
		t.Fatal(err)
	}

	if err := os.Mkdir(wanderungspfad(ort), 0o700); err != nil {
		t.Fatal(err)
	}

	if _, fehler := HalteSchemafreigabe(probekontext(t), db); !errors.Is(fehler, domain.ErrGesperrt) {
		t.Errorf("ohne Nachbardatei: %v, erwartet gesperrt", fehler)
	}
}

// **Im Arbeitsspeicher gibt es nichts zu sperren**, und der Stand wird
// trotzdem gefragt.
func TestImArbeitsspeicherFragtDieFreigabeNur(t *testing.T) {
	t.Parallel()

	db := leererBestand(t)
	if fehler := Migriere(probekontext(t), db, nil); fehler != nil {
		t.Fatal(fehler)
	}

	frei, fehler := HalteSchemafreigabe(probekontext(t), db)
	if fehler != nil {
		t.Fatalf("im Arbeitsspeicher: %v", fehler)
	}

	frei()
}

// **Ein Vorblick über einer geschlossenen Verbindung ist ein Fehler** — seit
// die Sperre Pflicht ist (Befund 4), erreicht ihn kein Weg über `Migriere`
// mehr; die Wanderung im Griff fragt ihn trotzdem.
func TestDerVorblickUeberEinerGeschlossenenVerbindung(t *testing.T) {
	t.Parallel()

	db := leererBestand(t)
	_ = db.Close() //nolint:errcheck // geschlossen gefragt

	if fehler := wandernImGriff(probekontext(t), db, migrationen, PruefeBestand, nil, grundlinie); fehler == nil {
		t.Error("über einer geschlossenen Verbindung kein Fehler")
	}
}
