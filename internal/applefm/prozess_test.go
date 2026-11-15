package applefm_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/applefm"
)

// **Testläufe rufen `fm` nicht** — weder zum Nachsehen noch zum Starten.
func TestDerRiegelHaeltFmAuf(t *testing.T) {
	t.Parallel()

	if _, fehler := applefm.Exitcode(t.Context(), "license", "--status"); !errors.Is(fehler, applefm.ErrImLauf) {
		t.Errorf("fm lief im Testlauf: %v", fehler)
	}
}

func TestVorhandenPrueftDieDatei(t *testing.T) {
	t.Parallel()

	if !applefm.Vorhanden("/usr/bin/true") || applefm.Vorhanden("/gibt/es/nicht") {
		t.Error("Vorhanden irrt")
	}
}

// kurzerOrdner gibt einen Ordner mit kurzem Pfad — `t.TempDir` ist für einen
// Socket zu lang.
func kurzerOrdner(t *testing.T) string {
	t.Helper()

	ordner, fehler := os.MkdirTemp("", "fm")
	if fehler != nil {
		t.Fatalf("%v", fehler)
	}

	t.Cleanup(func() { _ = os.RemoveAll(ordner) }) //nolint:errcheck // nur zum Aufräumen

	return ordner
}

// stummerStarter gibt einen Prozess, der nie einen Socket anlegt.
func stummerStarter(beendet *bool) applefm.Starter {
	return func(context.Context, ...string) (applefm.Prozess, error) {
		return prozessAus(func() error { *beendet = true; return nil }), nil
	}
}

type prozessAus func() error

func (p prozessAus) Beende() error { return p() }

// **Ein Dienst, der keinen Socket anlegt, wird beendet** — sonst lauschte
// `fm` vielleicht still auf TCP (Vermerk 0166).
func TestEinStummerDienstWirdBeendet(t *testing.T) {
	t.Parallel()

	beendet := false
	dienst := applefm.NeuerDienst(kurzerOrdner(t), stummerStarter(&beendet))

	ctx, abbrechen := context.WithTimeout(t.Context(), 200*time.Millisecond)
	defer abbrechen()

	if _, fehler := dienst.Klient(ctx); fehler == nil || !beendet {
		t.Errorf("%v, beendet %v", fehler, beendet)
	}

	// **Und Beenden ohne laufenden Dienst tut nichts.**
	if fehler := dienst.Beende(); fehler != nil {
		t.Errorf("%v", fehler)
	}
}

// **Der Ordner des Sockets gehört dem Programm** — ein vorhandener mit
// fremden Rechten, eine Datei an seiner Stelle oder ein zu langer Pfad wird
// abgewiesen, und gestartet wird nichts.
func TestDerOrdnerDesSocketsWirdGeprueft(t *testing.T) {
	t.Parallel()

	offen := kurzerOrdner(t)
	if fehler := os.Mkdir(filepath.Join(offen, "mse-fm"), 0o755); fehler != nil { //nolint:gosec // die Probe braucht fremde Rechte
		t.Fatalf("%v", fehler)
	}

	datei := kurzerOrdner(t)
	if fehler := os.WriteFile(filepath.Join(datei, "mse-fm"), nil, 0o600); fehler != nil {
		t.Fatalf("%v", fehler)
	}

	lang := filepath.Join(kurzerOrdner(t), strings.Repeat("l", 90))
	if fehler := os.Mkdir(lang, 0o700); fehler != nil {
		t.Fatalf("%v", fehler)
	}

	for name, ordner := range map[string]string{
		"fremde Rechte": offen, "eine Datei": datei, "zu lang": lang,
		"ohne Elternordner": "/gibt/es/nicht",
	} {
		gestartet := false
		dienst := applefm.NeuerDienst(ordner, func(context.Context, ...string) (applefm.Prozess, error) {
			gestartet = true
			return nil, errors.New("nicht")
		})

		if _, fehler := dienst.Klient(t.Context()); fehler == nil || gestartet {
			t.Errorf("%s: %v, gestartet %v", name, fehler, gestartet)
		}
	}
}

// **Ein Start, der scheitert, ist ein Fehler.**
func TestEinScheiternderStartIstEinFehler(t *testing.T) {
	t.Parallel()

	dienst := applefm.NeuerDienst(kurzerOrdner(t), func(context.Context, ...string) (applefm.Prozess, error) {
		return nil, errors.New("kein Start")
	})

	if _, fehler := dienst.Klient(t.Context()); fehler == nil {
		t.Error("kein Fehler")
	}
}
