package app

// **Was die ausführliche Stufe je Ruf kostet** — PERF03
// (#2353),
// Abnahme 3: höchstens 5 ms im p95 zusätzlich, gemessen über eine echte
// Weiche mit Datei und Drehung — die Zeile je Ruf (`profilnaht.go`), das
// Nachsehen vor jeder Zeile (`protokolldrehung.go`) und das Schwärzen.
//
// **Ohne `t.Parallel()`**: der Lauf setzt den Schreiber des Prozesses. **Was
// er nicht misst**: das Unified Log — es nimmt ab „info", die Zeile je Ruf
// erreicht es nicht (`protokollweiche.go`).

import (
	"io"
	"log/slog"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// rufzeiten misst `wieOft` Rufe des Fensters und gibt ihr p95.
func rufzeiten(t *testing.T, anwendung *App, wieOft int) time.Duration {
	t.Helper()

	dauern := make([]time.Duration, 0, wieOft)

	for range wieOft {
		anfang := time.Now()
		anwendung.Metriken(t.Context())
		dauern = append(dauern, time.Since(anfang))
	}

	slices.Sort(dauern)

	return dauern[len(dauern)*95/100]
}

func TestDieAusfuehrlicheStufeKostetHoechstensFuenfMillisekunden(t *testing.T) {
	vorher := slog.Default()
	t.Cleanup(func() { slog.SetDefault(vorher) })

	anwendung := NeueApp("")

	weiche := neueWeicheMit(t.TempDir(), io.Discard, io.Discard, os.UserHomeDir)
	t.Cleanup(func() { _ = weiche.Schliessen() }) //nolint:errcheck // Aufräumen

	regel := domain.ProtokollregelAus(map[string]string{domain.ProtokollStufe: domain.ProtokollstufeAusfuehrlich})
	regel.Datei = true

	slog.SetDefault(slog.New(&geschwaerzt{innen: slog.NewTextHandler(io.Discard, nil)}))
	ueblich := rufzeiten(t, anwendung, 1000)

	weiche.Einrichten(regel)
	slog.SetDefault(weiche.Logger())

	ausfuehrlich := rufzeiten(t, anwendung, 1000)
	zusatz := ausfuehrlich - ueblich

	t.Logf("kennzahl protokoll.ausfuehrlich.p95.zusatz_us=%d", zusatz.Microseconds())
	t.Logf("p95 üblich %s, ausführlich %s", ueblich, ausfuehrlich)

	if zusatz > 5*time.Millisecond {
		t.Errorf("die ausführliche Stufe kostet im p95 %s zusätzlich, erwartet höchstens 5 ms", zusatz)
	}

	if weiche.Stand().Datei == "" {
		t.Error("ohne Datei gemessen")
	}
}
