package app

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// TestAusfuehrlichesGehtNichtInsSystemprotokoll hält Nachlauf über Phase 6.4,
// N5: auf der ausführlichen Stufe steht jeder Ruf des Fensters, und über das
// Unified Log startete jede Zeile einen eigenen `logger`. **Das Systemprotokoll
// nimmt ab „info"**; die Datei nimmt, was eingestellt ist.
func TestAusfuehrlichesGehtNichtInsSystemprotokoll(t *testing.T) {
	ordner := t.TempDir()

	weiche, system, _ := weicheMitPuffern(t, ordner)
	weiche.Einrichten(domain.ProtokollregelAus(map[string]string{
		domain.ProtokollStufe: domain.ProtokollstufeAusfuehrlich,
	}))

	schreiber := weiche.Logger()
	schreiber.DebugContext(context.Background(), "eine ausführliche Zeile")
	schreiber.InfoContext(context.Background(), "eine gewöhnliche Zeile")

	if strings.Contains(system.String(), "ausführliche") || !strings.Contains(system.String(), "gewöhnliche") {
		t.Errorf("Systemprotokoll: %q", system.String())
	}

	datei, fehler := os.ReadFile(filepath.Join(ordner, Protokollname)) // #nosec G304 -- aus t.TempDir()
	if fehler != nil || !strings.Contains(string(datei), "ausführliche") {
		t.Errorf("Datei: %q, %v", datei, fehler)
	}

	if !schreiber.Enabled(context.Background(), slog.LevelDebug) {
		t.Error("die ausführliche Stufe gilt nicht")
	}
}

// wechselhaft scheitert beim ersten Schreiben und dann nicht mehr.
type wechselhaft struct{ rufe int }

func (w *wechselhaft) Write(zeile []byte) (int, error) {
	w.rufe++
	if w.rufe == 1 {
		return 0, os.ErrDeadlineExceeded
	}

	return len(zeile), nil
}

// TestDerJuengsteAusgangEinesKanalsZaehlt hält Nachlauf über Phase 6.4,
// N13: ein einmaliger Zeitablauf von `logger` nannte das Unified Log bis zum
// nächsten Einrichten „fehlt", obwohl es längst wieder schrieb.
func TestDerJuengsteAusgangEinesKanalsZaehlt(t *testing.T) {
	t.Parallel()

	kanal := &kanalschreiber{name: KanalSystem, ziel: &wechselhaft{}}

	kanal.schreiben([]byte("erste\n"))

	if kanal.gescheitert() == "" {
		t.Fatal("der Fehler wurde nicht bemerkt")
	}

	kanal.schreiben([]byte("zweite\n"))

	if grund := kanal.gescheitert(); grund != "" {
		t.Errorf("der Fehler klebt: %q", grund)
	}
}
