// **Zwei Prozesse an einem Hostbudget** — Befund Z7-03 von Phase 3.2
// (#1163).
//
// # Warum echte Prozesse
//
// `flock` gilt für die **offene Datei**, nicht für den Prozess. Zwei Griffe im
// selben Programm ständen einander im Weg — ein Test mit zwei Griffen prüfte,
// dass sich die Anwendung gegen sich selbst aussperrt, und das ist das
// Gegenteil der Zusage.
package netz_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// helferschalter sagt dem zweiten Prozess, welche Absprache er benutzen soll.
const helferschalter = "MSE_ABSPRACHE"

// TestMain lässt den Testlauf auch der zweite Prozess sein.
func TestMain(m *testing.M) {
	if pfad := os.Getenv(helferschalter); pfad != "" {
		schickeUeberDieAbsprache(pfad)

		return
	}

	if ort := os.Getenv(ladeschalter); ort != "" {
		holeUeberDenSpeicher(ort)

		return
	}

	os.Exit(m.Run())
}

// schickeUeberDieAbsprache holt sich zweimal die Erlaubnis und meldet, wann.
func schickeUeberDieAbsprache(pfad string) {
	grenze := netz.NeueProzessgrenze(pfad, 200*time.Millisecond, time.Now, nil)

	for range 2 {
		if err := grenze.Warte(context.Background()); err != nil {
			os.Exit(2)
		}

		if _, err := os.Stdout.WriteString(
			strconv.FormatInt(time.Now().UnixNano(), 10) + "\n"); err != nil {
			os.Exit(3)
		}
	}
}

// **Zwei Prozesse halten zusammen den Abstand ein.**
//
// # Was gemessen wird
//
// Vier Erlaubnisse, zwei je Prozess, bei einem Abstand von 200 ms. Liefen die
// Prozesse unabhängig, lägen je zwei davon dicht beieinander — ein Prozess
// wartet nur auf sich selbst. Mit der Absprache liegt **zwischen je zweien**
// der Abstand.
//
// **Die Schwelle ist großzügig**: gemessen wird gegen die Hälfte des
// Abstands. Was der Lauf zeigen soll, ist der Unterschied zwischen „warten"
// und „gar nicht warten", nicht die Genauigkeit einer Uhr.
func TestZweiProzesseHaltenZusammenDenAbstand(t *testing.T) {
	pfad := filepath.Join(t.TempDir(), "budget.drossel")

	zeiten := make(chan int64, 4)

	var laeuft sync.WaitGroup

	for range 2 {
		laeuft.Add(1)

		go func() {
			defer laeuft.Done()

			for _, wann := range zweiterSchickt(t, pfad) {
				zeiten <- wann
			}
		}()
	}

	laeuft.Wait()
	close(zeiten)

	var gesammelt []int64
	for wann := range zeiten {
		gesammelt = append(gesammelt, wann)
	}

	if len(gesammelt) != 4 {
		t.Fatalf("es kamen %d Zeitpunkte statt vier: %v", len(gesammelt), gesammelt)
	}

	slices.Sort(gesammelt)

	for i := 1; i < len(gesammelt); i++ {
		abstand := time.Duration(gesammelt[i] - gesammelt[i-1])
		if abstand < 100*time.Millisecond {
			t.Errorf("zwischen Nummer %d und %d lagen nur %v — die Absprache griff nicht",
				i-1, i, abstand)
		}
	}
}

// zweiterSchickt startet einen Helfer und gibt seine beiden Zeitpunkte.
func zweiterSchickt(t *testing.T, pfad string) []int64 {
	t.Helper()

	// #nosec G204,G702 -- der Testlauf selbst
	befehl := exec.CommandContext(t.Context(), os.Args[0])
	befehl.Env = append(os.Environ(), helferschalter+"="+pfad)

	roh, err := befehl.Output()
	if err != nil {
		t.Errorf("der Helfer scheiterte: %v", err)

		return nil
	}

	var zeiten []int64

	for _, zeile := range strings.Split(strings.TrimSpace(string(roh)), "\n") {
		wann, fehler := strconv.ParseInt(zeile, 10, 64)
		if fehler != nil {
			t.Errorf("unlesbarer Zeitpunkt %q: %v", zeile, fehler)

			return nil
		}

		zeiten = append(zeiten, wann)
	}

	return zeiten
}
