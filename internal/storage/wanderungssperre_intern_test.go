package storage

// **Zwei Starter, eine Sicherung** — der erste Befund aus dem Abnahmelauf von
// M01 (`aufstieg09`), behoben in M02
// (#2346).
//
// Ob gesichert wird, entschied ein Blick in die Buchführung **ohne Sperre**,
// und gesichert wurde vor der Schreibsperre. Zwei Starter — das Fenster und
// ein Dienst nach einer Aktualisierung — sicherten deshalb beide; der zweite
// fand die erste Sicherung, sah den inzwischen gewanderten Bestand und legte
// ihn als zweite „Sicherung vor" ab. Gemessen: `bestand-vor-0.8.0.db` mit 97
// Wanderungen und daneben eine mit 98.

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/dateisperre"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestZweiStarterSichernEinmal(t *testing.T) {
	ort := filepath.Join(t.TempDir(), "bestand.db")

	alt, err := Oeffnen(probekontext(t), ort)
	if err != nil {
		t.Fatal(err)
	}

	if fehler := migriereAusMitVorbereitung(probekontext(t), alt, nurGrundlinie,
		PruefeBestand, nil, grundlinie); fehler != nil {
		t.Fatalf("der Stand vor der Wanderung: %v", fehler)
	}

	_ = alt.Close() //nolint:errcheck // nur angelegt

	// **Der erste sichert lange** — so lange, dass der zweite währenddessen
	// in die Buchführung sieht.
	var gesichert atomic.Int32

	//nolint:unparam // die Gestalt einer Vorbereitung; dieser sichert nur lange
	vorbereiten := func(context.Context) error {
		if gesichert.Add(1) == 1 {
			time.Sleep(200 * time.Millisecond)
		}

		return nil
	}

	var (
		warten sync.WaitGroup
		fehler [2]error
	)

	for i := range fehler {
		warten.Add(1)

		go func() {
			defer warten.Done()

			db, oeffnenfehler := Oeffnen(probekontext(t), ort)
			if oeffnenfehler != nil {
				fehler[i] = oeffnenfehler

				return
			}

			defer func() { _ = db.Close() }() //nolint:errcheck // nur gewandert

			fehler[i] = Migriere(probekontext(t), db, vorbereiten)
		}()

		time.Sleep(50 * time.Millisecond)
	}

	warten.Wait()

	for i, f := range fehler {
		if f != nil {
			t.Errorf("Starter %d: %v", i+1, f)
		}
	}

	if zahl := gesichert.Load(); zahl != 1 {
		t.Errorf("gesichert wurde %d-mal, erwartet einmal — der zweite Starter sah eine "+
			"ausstehende Wanderung, die der erste gerade ging", zahl)
	}
}

// **Wer die Sperre nicht in der Frist bekommt, startet nicht** — er hört
// „belegt" und versucht es gleich noch einmal (F05), statt ohne sie zu
// wandern. Nicht parallel: die Frist ist eine Paketvariable.
func TestEinStarterOhneDieSperreIstBelegt(t *testing.T) { //nolint:paralleltest // siehe darüber
	alteFrist := Wanderungsfrist
	Wanderungsfrist = 50 * time.Millisecond

	t.Cleanup(func() { Wanderungsfrist = alteFrist })

	ort := filepath.Join(t.TempDir(), "bestand.db")

	db, err := Oeffnen(probekontext(t), ort)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = db.Close() }) //nolint:errcheck // nur zum Aufräumen

	halter, err := dateisperre.Oeffnen(ort + Wanderungsendung)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = halter.Schliessen() }) //nolint:errcheck // nur zum Aufräumen

	if fehler := halter.Allein(); fehler != nil {
		t.Fatal(fehler)
	}

	if fehler := Migriere(probekontext(t), db, nil); !errors.Is(fehler, domain.ErrBelegt) {
		t.Errorf("Migriere gab %v, erwartet belegt", fehler)
	}
}

// **Ohne Nachbardatei wandert niemand** — Zweitblick über Phase 2, Befund 4:
// hier stand „wie beim Öffnen", und ein Verzeichnis unter dem Namen der
// Sperre setzte die ganze Koordination außer Kraft. Beim Öffnen schützt die
// Sperre nur das Lesen; eine Wanderung schreibt ohnehin, und wer neben dem
// Bestand nichts anlegen kann, kann ihn auch nicht sichern.
func TestOhneNachbardateiWandertNiemand(t *testing.T) {
	t.Parallel()

	ort := filepath.Join(t.TempDir(), "bestand.db")
	if err := os.Mkdir(ort+Wanderungsendung, 0o700); err != nil {
		t.Fatal(err)
	}

	db, err := Oeffnen(probekontext(t), ort)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = db.Close() }) //nolint:errcheck // nur zum Aufräumen

	if fehler := Migriere(probekontext(t), db, nil); !errors.Is(fehler, domain.ErrGesperrt) {
		t.Errorf("ohne Nachbardatei: %v, erwartet gesperrt", fehler)
	}

	if verbucht, _ := AngewendeteMigrationen(probekontext(t), db); len(verbucht) != 0 { //nolint:errcheck // gezählt
		t.Errorf("ohne Sperre gewandert: %d Wanderungen", len(verbucht))
	}
}

// **Eine Verbindung, die nichts mehr sagt, ist ein Fehler** und nicht „kein
// Ort" (Befund 4): sonst wanderte, wer seinen Ort nicht erfährt, ohne Sperre.
// Ohne Ort ist nur der Bestand im Arbeitsspeicher.
func TestEineGeschlosseneVerbindungNenntKeinenOrt(t *testing.T) {
	t.Parallel()

	db, err := Oeffnen(probekontext(t), filepath.Join(t.TempDir(), "bestand.db"))
	if err != nil {
		t.Fatal(err)
	}

	_ = db.Close() //nolint:errcheck // geschlossen gefragt

	if ort, fehler := dateiVon(probekontext(t), db); fehler == nil {
		t.Errorf("eine geschlossene Verbindung nennt %q ohne Fehler", ort)
	}

	// **Und der Wärter vor der Wanderung fragt über denselben Weg**
	// (Nachlauf über Phase 2, Befund 5): ohne Ort hält er niemanden an.
	if ort, fehler := Bestandsdatei(probekontext(t), db); fehler == nil {
		t.Errorf("Bestandsdatei nennt %q ohne Fehler", ort)
	}

	if _, fehler := wanderungHalten(probekontext(t), db); fehler == nil {
		t.Error("ohne Ort gab es eine Sperre")
	}
}
