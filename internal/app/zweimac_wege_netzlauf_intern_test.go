//go:build netzlauf

package app

// **Die Wege des Messlaufs „Zwei Macs auf einem"** — der Vermittler mit
// Laufzeit, die Adresse von `en0` und die Zusammenfassung einer Messreihe.
// Herausgelöst aus `zweimac_netzlauf_intern_test.go`, als er an die
// Zeilengrenze aus §2 stieß.

import (
	"fmt"
	"net"
	"slices"
	"sort"
	"testing"
	"time"
)

// verzoegerer ist eine Leitung mit Laufzeit: TCP auf der Schleife, jedes
// Stück je Richtung um `laufzeit` später.
func verzoegerer(t *testing.T, ziel string, laufzeit time.Duration) string {
	t.Helper()

	horcher, fehler := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0")
	if fehler != nil {
		t.Fatalf("der Vermittler horcht nicht: %v", fehler)
	}

	t.Cleanup(func() { _ = horcher.Close() }) //nolint:errcheck // nur ein Lauf

	go func() {
		for {
			hinein, annahmefehler := horcher.Accept()
			if annahmefehler != nil {
				return
			}

			hinaus, waehlfehler := (&net.Dialer{}).DialContext(t.Context(), "tcp", ziel)
			if waehlfehler != nil {
				_ = hinein.Close() //nolint:errcheck // nur ein Lauf

				continue
			}

			go laufzeitkopie(hinaus, hinein, laufzeit)
			go laufzeitkopie(hinein, hinaus, laufzeit)
		}
	}()

	return horcher.Addr().String()
}

type stueck struct {
	daten   []byte
	gelesen time.Time
}

// laufzeitkopie liest, stempelt und schreibt nach der Laufzeit — ohne das
// Lesen aufzuhalten.
func laufzeitkopie(ziel, quelle net.Conn, laufzeit time.Duration) {
	unterwegs := make(chan stueck, 1024)

	go func() {
		defer close(unterwegs)

		puffer := make([]byte, 64<<10)

		for {
			gelesen, fehler := quelle.Read(puffer)
			if gelesen > 0 {
				unterwegs <- stueck{daten: slices.Clone(puffer[:gelesen]), gelesen: time.Now()}
			}

			if fehler != nil {
				return
			}
		}
	}()

	for eines := range unterwegs {
		time.Sleep(time.Until(eines.gelesen.Add(laufzeit)))

		if _, fehler := ziel.Write(eines.daten); fehler != nil {
			break
		}
	}

	_ = ziel.Close()   //nolint:errcheck // nur ein Lauf
	_ = quelle.Close() //nolint:errcheck // nur ein Lauf
}

// en0 gibt die IPv4-Adresse, die ein anderer Mac im Netz wählen würde.
func en0(t *testing.T) string {
	t.Helper()

	schnitt, fehler := net.InterfaceByName("en0")
	if fehler != nil {
		t.Skipf("kein en0: %v", fehler)
	}

	adressen, fehler := schnitt.Addrs()
	if fehler != nil {
		t.Skipf("en0 nennt keine Adresse: %v", fehler)
	}

	for _, adresse := range adressen {
		if netz, passt := adresse.(*net.IPNet); passt && netz.IP.To4() != nil {
			return netz.IP.String()
		}
	}

	t.Skip("en0 hat keine IPv4-Adresse")

	return ""
}

// messreihe hält Dauern und fasst sie zusammen.
type messreihe []time.Duration

func (m messreihe) String() string {
	if len(m) == 0 {
		return "—"
	}

	sortiert := slices.Clone(m)
	sort.Slice(sortiert, func(i, j int) bool { return sortiert[i] < sortiert[j] })

	stelle := func(anteil float64) time.Duration {
		return sortiert[min(len(sortiert)-1, int(anteil*float64(len(sortiert))))]
	}

	return fmt.Sprintf("Median %s · p95 %s · max %s (n=%d)",
		stelle(0.5).Round(100*time.Microsecond), stelle(0.95).Round(100*time.Microsecond),
		sortiert[len(sortiert)-1].Round(100*time.Microsecond), len(sortiert))
}

// warteAufNebenbeiText wartet, bis ein Satz beim Gegenüber ankommt.
func warteAufNebenbeiText(t *testing.T, nebenbei chan NebenbeiDTO, text string) {
	t.Helper()

	frist := time.After(10 * time.Second)

	for {
		select {
		case eines := <-nebenbei:
			if eines.Text == text {
				return
			}
		case <-frist:
			t.Fatalf("%q kam nicht an", text)
		}
	}
}
