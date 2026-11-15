package app

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/applefm"
)

// **Die Fehlerwege des Gerätemodells** — AI-05 (#2042). Jeder steht für sich
// und lässt keinen halben Stand zurück.

// **Ohne Bestand sagt der Vorschlag, was fehlt** — beim Vorrang wie bei einer
// ausdrücklichen Wahl.
func TestOhneBestandSagenBeideWegeWasFehlt(t *testing.T) {
	t.Parallel()

	a := &App{}

	if antwort := a.vorschlagGewaehlt(t.Context(), kifrage{}, KIVorschlagsbezug{}); antwort.Fehler == nil {
		t.Errorf("gewählt: %+v", antwort)
	}

	if antwort := a.vorschlagMitVorrang(t.Context(), kifrage{}, KIVorschlagsbezug{}); antwort.Fehler == nil {
		t.Errorf("Vorrang: %+v", antwort)
	}
}

// **Ohne die Rolle bleibt der Schalter, wie er war.**
func TestDerSitzungsschalterBrauchtDieRolle(t *testing.T) {
	t.Parallel()

	a := (&App{}).MitWache(&abweisendeWache{})
	a.MitGeraetemodell(nil, nil, "")

	antwort := a.KIGeraetemodellFuerSitzung(t.Context(), false)
	if antwort.Fehler == nil || !antwort.An || a.geraetSitzungAus() {
		t.Errorf("%+v, aus %v", antwort, a.geraetSitzungAus())
	}
}

// **Abschalten beendet den laufenden Dienst, Anschalten startet keinen.**
func TestDerSitzungsschalterBeendetDenDienst(t *testing.T) {
	t.Parallel()

	gestartet := 0
	a := geraeteapp(t, applefm.EinrichtungBereit,
		`{"model":"system","choices":[{"finish_reason":"stop","message":{"content":"{}","refusal":null}}]}`,
		&gestartet)

	if _, fehler := a.geraetebauplan()().Vorschlagen(t.Context(), "", applefm.Modell, "x", nil); fehler != nil {
		t.Fatalf("%v", fehler)
	}

	if aus := a.KIGeraetemodellFuerSitzung(t.Context(), false); aus.An || aus.Fehler != nil || !a.geraetSitzungAus() {
		t.Errorf("aus: %+v", aus)
	}

	// **Aus heißt auch für einen, der den Anbieter schon gewählt hat**
	// (Z2 des Zweitblicks über Phase 6.10): der Dienst selbst hält den Start
	// auf.
	if _, fehler := a.geraetebauplan()().Vorschlagen(t.Context(), "", applefm.Modell, "x", nil); !errors.Is(fehler, applefm.ErrAbgeschaltet) {
		t.Errorf("nach aus: %v", fehler)
	}

	if an := a.KIGeraetemodellFuerSitzung(t.Context(), true); !an.An || an.Fehler != nil || a.geraetSitzungAus() {
		t.Errorf("an: %+v", an)
	}

	if gestartet != 1 {
		t.Errorf("gestartet: %d", gestartet)
	}

	// **Und an heißt: der nächste Vorschlag startet wieder.**
	if _, fehler := a.geraetebauplan()().Vorschlagen(t.Context(), "", applefm.Modell, "x", nil); fehler != nil || gestartet != 2 {
		t.Errorf("nach an: %v, gestartet %d", fehler, gestartet)
	}
}

// **Im Testlauf startet der echte Weg nicht** — mit leerem Starter und Ordner
// nimmt die Naht `applefm.Prozessstarter` und `$TMPDIR`, und der Riegel hält.
func TestDieVorgabenDerNahtHaltenAmRiegel(t *testing.T) {
	t.Parallel()

	a := (&App{}).MitGeraetemodell(nil, nil, "")

	_, fehler := a.geraetebauplan()().Vorschlagen(t.Context(), "", applefm.Modell, "x", nil)
	if !errors.Is(fehler, applefm.ErrImLauf) {
		t.Errorf("%v", fehler)
	}

	// **Und ohne laufenden Dienst tut Beenden nichts** — auch das des Prozesses.
	GeraetBeenden()
}

// **Eine Antwort, die sich nicht lesen lässt, ist ein Fehler.**
func TestEinFehlerDesDienstesIstEinFehler(t *testing.T) {
	t.Parallel()

	gestartet := 0
	a := geraeteapp(t, applefm.EinrichtungBereit, `kein JSON`, &gestartet)

	if _, fehler := a.geraetebauplan()().Vorschlagen(t.Context(), "", applefm.Modell, "x", nil); fehler == nil {
		t.Error("kein Fehler")
	}
}

// stoerrischerServer antwortet wie `fm serve` und lässt sich nicht sauber
// beenden.
type stoerrischerServer struct{ horcher net.Listener }

func (s stoerrischerServer) Beende() error {
	_ = s.horcher.Close() //nolint:errcheck // der Fehler darunter ist der gemeinte

	return errors.New("hängt")
}

// **Ein Beenden, das scheitert, wird gemeldet** — und hält nichts an.
func TestEinScheiterndesBeendenWirdGemeldet(t *testing.T) {
	t.Parallel()

	ordner, fehler := os.MkdirTemp("", "fm")
	if fehler != nil {
		t.Fatalf("%v", fehler)
	}

	t.Cleanup(func() { _ = os.RemoveAll(ordner) }) //nolint:errcheck // nur zum Aufräumen

	a := (&App{}).MitGeraetemodell(
		func(context.Context) applefm.Einrichtung { return applefm.EinrichtungBereit },
		func(ctx context.Context, argumente ...string) (applefm.Prozess, error) {
			horcher, err := (&net.ListenConfig{}).Listen(ctx, "unix", argumente[len(argumente)-1])
			if err != nil {
				return nil, err
			}

			go func() {
				_ = http.Serve(horcher, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { //nolint:gosec,errcheck // Probe
					_, _ = w.Write([]byte(`{"models":[{"available":true,"name":"system"}]}`)) //nolint:errcheck // Probe
				}))
			}()

			return stoerrischerServer{horcher: horcher}, nil
		}, ordner)

	if _, err := a.geraetedienst().Klient(t.Context()); err != nil {
		t.Fatalf("%v", err)
	}

	a.geraetBeenden()

	if einrichtung := a.geraeteeinrichtung(t.Context()); einrichtung != applefm.EinrichtungBereit {
		t.Errorf("%q", einrichtung)
	}

	// **Auch das Abschalten für die Sitzung meldet es** und schaltet trotzdem ab.
	if _, err := a.geraetedienst().Klient(t.Context()); err != nil {
		t.Fatalf("%v", err)
	}

	if aus := a.KIGeraetemodellFuerSitzung(t.Context(), false); aus.An || aus.Fehler != nil {
		t.Errorf("aus: %+v", aus)
	}
}

// **Für die Sitzung abgeschaltet ist nicht »von der Richtlinie
// abgeschaltet«** — sonst nennte der Katalog einen falschen Grund und
// schickte an den falschen Schalter (§8).
func TestDieSitzungIstKeineRichtlinie(t *testing.T) {
	t.Parallel()

	befunde := vonDerRichtlinieAbgeschaltet([]kianbieter{
		{name: geraeteanbieter, gewaehlt: true, eingeschaltet: false, sitzungAus: true},
	})
	if len(befunde) != 0 {
		t.Errorf("%+v", befunde)
	}
}
