package app_test

// **Die Absprache über Prozessgrenzen** — Werkstück A15 von Phase 5.3
// (#1514).
//
// # Warum das hier gemessen wird und nicht nur im Dienst
//
// **Weil der Dienst nur eine Schnittstelle kennt.** Ob `flock` wirklich zwei
// Griffe unterscheidet, sagt kein Doppel — das sagt das Betriebssystem. Das
// Issue verlangt es wörtlich: „prozessübergreifende Konkurrenz ausdrücklich
// **messen** und begrenzen."
//
// **Zwei Griffe sind zwei Prozesse, soweit es `flock` betrifft.** Die Sperre
// gilt der **offenen Datei** und nicht dem Prozess; zwei `Oeffnen` auf
// denselben Pfad verhalten sich zueinander wie zwei Programme. Das steht so
// im Kopf von `internal/dateisperre`, und dieser Lauf lebt davon.

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/dateisperre"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Der zweite Griff bekommt die Sperre nicht.**
func TestDerZweiteGriffBekommtDieIndexsperreNicht(t *testing.T) {
	t.Parallel()

	ort := filepath.Join(t.TempDir(), "bestand.db")
	sperre := app.Indexsperre(func() string { return ort })

	freigeben, genommen, fehler := sperre.Nehmen()
	if fehler != nil || !genommen {
		t.Fatalf("der erste Griff: %v, %v", fehler, genommen)
	}

	_, zweiter, zweitfehler := sperre.Nehmen()
	if zweitfehler != nil || zweiter {
		t.Fatalf("der zweite Griff: %v, %v", zweitfehler, zweiter)
	}

	// **Und nach der Freigabe geht es wieder.** Ohne diese Zeile stünde
	// „genommen" ohne Gegenprobe da: eine Sperre, die nie fällt, weist den
	// zweiten auch dann ab, wenn niemand baut.
	freigeben()

	dritter := nachDerFreigabe(t, sperre)
	dritter()
}

// nachDerFreigabe greift, bis die Sperre wieder frei ist.
//
// **Nicht sofort, und das ist gemessen**: `flock` gilt für die offene
// Dateibeschreibung, und ein Test daneben, der gerade einen Kindprozess
// forkt, hält sie bis zu dessen `exec` mit — `O_CLOEXEC` greift erst dort. In
// `make ci` kam der dritte Griff am 2026-09-27 so einmal mit „belegt" zurück.
// **Die Zusage bleibt**: nach der Freigabe ist die Sperre frei — nur nicht in
// derselben Mikrosekunde.
func nachDerFreigabe(t *testing.T, sperre service.Prozesssperre) func() {
	t.Helper()

	bis := time.Now().Add(2 * time.Second)

	for {
		freigeben, genommen, fehler := sperre.Nehmen()
		if fehler != nil {
			t.Fatalf("der dritte Griff: %v", fehler)
		}

		if genommen {
			return freigeben
		}

		if time.Now().After(bis) {
			t.Fatal("der dritte Griff: die Sperre fiel nicht")
		}

		time.Sleep(10 * time.Millisecond)
	}
}

// **Ohne Ort gibt es keine Absprache — und der Aufbau läuft trotzdem.**
//
// **Ein Bestand im Arbeitsspeicher hat keinen Ort**, an dem eine Sperrdatei
// liegen könnte. Ihn deshalb nicht bauen zu lassen wäre eine Sperre, die
// etwas anderes verhindert als zwei Schreiber.
func TestOhneOrtGibtEsKeineIndexsperre(t *testing.T) {
	t.Parallel()

	sperre := app.Indexsperre(func() string { return "" })

	for range 2 {
		freigeben, genommen, fehler := sperre.Nehmen()
		if fehler != nil || !genommen {
			t.Fatalf("ohne Ort: %v, %v", fehler, genommen)
		}

		freigeben()
	}
}

// **Ein Ort, an dem keine Datei entstehen kann, ist ein Fehler.**
//
// **Und keine Abweisung**: „ein anderer baut" ist eine Auskunft, „das
// Verzeichnis gibt es nicht" ist etwas anderes.
func TestEinUnmoeglicherOrtIstEinFehler(t *testing.T) {
	t.Parallel()

	sperre := app.Indexsperre(func() string {
		return filepath.Join(t.TempDir(), "gibtesnicht", "bestand.db")
	})

	if _, genommen, fehler := sperre.Nehmen(); fehler == nil || genommen {
		t.Fatalf("der unmögliche Ort: %v, %v", fehler, genommen)
	}
}

// **Ein Erbe baut nicht** — #2401: ein Griff, der die Sperre nur kurz hält,
// wie ein Kind zwischen `fork` und `exec`, heißt nicht „ein anderer baut".
func TestEinKurzerHalterHaeltDenIndexbauNichtAuf(t *testing.T) {
	t.Parallel()

	ort := filepath.Join(t.TempDir(), "bestand.db")

	erbe, err := dateisperre.Oeffnen(ort + app.Indexbauendung)
	if err != nil {
		t.Fatal(err)
	}

	if err := erbe.Allein(); err != nil {
		t.Fatal(err)
	}

	uhr := time.AfterFunc(20*time.Millisecond, func() { _ = erbe.Schliessen() }) //nolint:errcheck // losgelassen ist losgelassen
	t.Cleanup(func() { uhr.Stop() })

	freigeben, genommen, fehler := app.Indexsperre(func() string { return ort }).Nehmen()
	if fehler != nil || !genommen {
		t.Fatalf("nach einem Halter von 20 ms: %v, %v", fehler, genommen)
	}

	freigeben()
}
