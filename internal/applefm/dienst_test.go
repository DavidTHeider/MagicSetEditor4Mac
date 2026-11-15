package applefm_test

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/applefm"
)

// **Aktiviert heißt drei Dinge zugleich** (Vermerk 0166): das Werkzeug ist da,
// die Lizenz ist angenommen, das Modell ist verfügbar — gelesen an den
// Exit-Codes, nicht am Text, denn die Ausgabe ist teils lokalisiert.
func TestDieErkennungLiestExitCodes(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name       string
		da         bool
		lizenz     int
		verfuegbar int
		erwartet   applefm.Einrichtung
	}{
		{"alles da", true, 0, 0, applefm.EinrichtungBereit},
		{"kein Werkzeug", false, 0, 0, applefm.EinrichtungFehlt},
		{"Lizenz nicht angenommen", true, 69, 0, applefm.EinrichtungLizenz},
		{"Modell nicht verfügbar", true, 0, 1, applefm.EinrichtungNichtVerfuegbar},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			var gerufen []string

			einrichtung := applefm.Erkenne(t.Context(), func(string) bool { return fall.da },
				func(_ context.Context, argumente ...string) (int, error) {
					gerufen = append(gerufen, strings.Join(argumente, " "))
					if argumente[0] == "license" {
						return fall.lizenz, nil
					}

					return fall.verfuegbar, nil
				})
			if einrichtung != fall.erwartet {
				t.Errorf("%q, erwartet %q", einrichtung, fall.erwartet)
			}

			// **Nie `sudo`, nie die Annahme der Lizenz** — gefragt wird nur der Stand.
			for _, ruf := range gerufen {
				if strings.Contains(ruf, "sudo") || (strings.Contains(ruf, "license") && !strings.Contains(ruf, "--status")) {
					t.Errorf("gerufen: %q", ruf)
				}
			}
		})
	}
}

// **Ein Ruf, der gar nicht läuft, ist nicht bereit** — und kein Absturz.
func TestEinScheiternderRufIstNichtBereit(t *testing.T) {
	t.Parallel()

	einrichtung := applefm.Erkenne(t.Context(), func(string) bool { return true },
		func(context.Context, ...string) (int, error) { return -1, errors.New("kein Prozess") })
	if einrichtung != applefm.EinrichtungLizenz {
		t.Errorf("%q", einrichtung)
	}
}

// **Ein Testlauf sieht kein Gerätemodell — auf jedem Rechner gleich.** Der
// Riegel lässt `fm` nicht laufen; ohne diese Regel hieße das auf einem Mac
// mit macOS 27 „Lizenz fehlt" und auf jedem anderen „fehlt", und ein Lauf
// über den Katalog sähe je nach Rechner anders aus.
func TestEinTestlaufSiehtKeinGeraetemodell(t *testing.T) {
	t.Parallel()

	if einrichtung := applefm.Erkenne(t.Context(), func(string) bool { return true }, applefm.Exitcode); einrichtung != applefm.EinrichtungFehlt {
		t.Errorf("%q", einrichtung)
	}
}

// **Der Socketpfad bleibt unter 104 Bytes** — sonst legte `fm serve` keinen
// Socket an und lauschte still auf allen Schnittstellen (Vermerk 0166).
func TestEinZuLangerSocketpfadWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	if fehler := applefm.PruefeSocketpfad(strings.Repeat("a", 104)); fehler == nil {
		t.Error("ein Pfad von 104 Bytes ging durch")
	}

	if fehler := applefm.PruefeSocketpfad("/kurz/fm.sock"); fehler != nil {
		t.Errorf("%v", fehler)
	}
}

// probeprozess ist ein `fm serve`, der an seinem Socket antwortet.
type probeprozess struct{ horcher net.Listener }

func (p *probeprozess) Beende() error { return p.horcher.Close() }

// **Der Dienst startet bei Bedarf, wartet auf `/health` und beendet sich.**
func TestDerDienstStartetBeiBedarfUndBeendet(t *testing.T) {
	t.Parallel()

	ordner, fehler := os.MkdirTemp("", "fm")
	if fehler != nil {
		t.Fatalf("%v", fehler)
	}

	t.Cleanup(func() { _ = os.RemoveAll(ordner) }) //nolint:errcheck // nur zum Aufräumen

	var gestartet []string

	var prozess *probeprozess

	dienst := applefm.NeuerDienst(ordner, func(ctx context.Context, argumente ...string) (applefm.Prozess, error) {
		gestartet = append(gestartet, strings.Join(argumente, " "))
		horcher, err := (&net.ListenConfig{}).Listen(ctx, "unix", argumente[len(argumente)-1])
		if err != nil {
			return nil, err
		}

		go func() {
			_ = http.Serve(horcher, http.HandlerFunc(func(antwort http.ResponseWriter, _ *http.Request) { //nolint:gosec,errcheck // Probe
				_, _ = antwort.Write([]byte(`{"models":[{"available":true,"name":"system"}]}`)) //nolint:errcheck // Probe
			}))
		}()

		prozess = &probeprozess{horcher: horcher}

		return prozess, nil
	})

	klient, fehler := dienst.Klient(t.Context())
	if fehler != nil {
		t.Fatalf("Klient: %v", fehler)
	}

	if fehler := klient.Gesund(t.Context()); fehler != nil {
		t.Errorf("gesund: %v", fehler)
	}

	// **Ein zweiter Ruf startet keinen zweiten Prozess.**
	if _, fehler := dienst.Klient(t.Context()); fehler != nil || len(gestartet) != 1 {
		t.Errorf("gestartet: %v, %v", gestartet, fehler)
	}

	if !strings.HasPrefix(gestartet[0], "serve --socket ") {
		t.Errorf("gestartet: %q", gestartet[0])
	}

	if info, fehler := os.Stat(filepath.Dir(strings.TrimPrefix(gestartet[0], "serve --socket "))); fehler != nil || info.Mode().Perm() != 0o700 {
		t.Errorf("Ordner des Sockets: %v %v", info, fehler)
	}

	if fehler := dienst.Beende(); fehler != nil {
		t.Errorf("beenden: %v", fehler)
	}
}

// **Testläufe starten `/usr/bin/fm` nie** — der Riegel steht im echten Starter.
func TestDerEchteStarterIstImTestlaufGesperrt(t *testing.T) {
	t.Parallel()

	if _, fehler := applefm.Prozessstarter(t.Context(), "serve"); !errors.Is(fehler, applefm.ErrImLauf) {
		t.Errorf("%v", fehler)
	}
}
