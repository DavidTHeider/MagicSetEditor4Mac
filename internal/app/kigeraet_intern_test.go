package app

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/applefm"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Das Gerätemodell als Anbieter, über Attrappen** — AI-05 (#2042). Kein
// Lauf startet `/usr/bin/fm` (§3); der „Server" ist ein Unix-Socket in diesem
// Prozess, der in der gemessenen Gestalt antwortet (Vermerk 0166).

// probeServer antwortet am Socket mit `antwort` auf jede Frage.
type probeServer struct{ horcher net.Listener }

func (p probeServer) Beende() error { return p.horcher.Close() }

func probestarter(antwort string, gestartet *int) applefm.Starter {
	return func(ctx context.Context, argumente ...string) (applefm.Prozess, error) {
		*gestartet++

		horcher, fehler := (&net.ListenConfig{}).Listen(ctx, "unix", argumente[len(argumente)-1])
		if fehler != nil {
			return nil, fehler
		}

		go func() {
			_ = http.Serve(horcher, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { //nolint:gosec,errcheck // Probe
				if r.URL.Path == applefm.Gesundheitsweg {
					_, _ = w.Write([]byte(`{"models":[{"available":true,"name":"system"}]}`)) //nolint:errcheck // Probe

					return
				}

				_, _ = w.Write([]byte(antwort)) //nolint:errcheck // Probe
			}))
		}()

		return probeServer{horcher: horcher}, nil
	}
}

func geraeteapp(t *testing.T, einrichtung applefm.Einrichtung, antwort string, gestartet *int) *App {
	t.Helper()

	ordner, fehler := os.MkdirTemp("", "fm")
	if fehler != nil {
		t.Fatalf("%v", fehler)
	}

	t.Cleanup(func() { _ = os.RemoveAll(ordner) }) //nolint:errcheck // nur zum Aufräumen

	a := &App{}
	a.MitGeraetemodell(func(context.Context) applefm.Einrichtung { return einrichtung },
		probestarter(antwort, gestartet), ordner)
	t.Cleanup(a.geraetBeenden)

	return a
}

// **Bereit: es nennt seine Fassung und sein Modell — ohne zu starten.**
func TestDasGeraetemodellNenntSichOhneZuStarten(t *testing.T) {
	t.Parallel()

	gestartet := 0
	quelle := geraeteapp(t, applefm.EinrichtungBereit, "", &gestartet).geraetebauplan()()

	if fassung, fehler := quelle.Fassung(t.Context()); fehler != nil || fassung == "" {
		t.Errorf("%q %v", fassung, fehler)
	}

	modelle, _, fehler := quelle.Modelle(t.Context())
	if fehler != nil || len(modelle) != 1 || modelle[0].Ort != domain.KIOrtLokal ||
		!modelle[0].Kann(domain.KIStrukturierteGenerierung) {
		t.Errorf("%+v %v", modelle, fehler)
	}

	if gestartet != 0 {
		t.Errorf("gestartet: %d", gestartet)
	}
}

// **Nicht aktiviert: die Fassung sagt, was fehlt** (AI-05.AK-03).
func TestDieFassungSagtWasFehlt(t *testing.T) {
	t.Parallel()

	for einrichtung, satz := range map[applefm.Einrichtung]string{
		applefm.EinrichtungLizenz:          "sudo fm license",
		applefm.EinrichtungNichtVerfuegbar: "Apple Intelligence",
		applefm.EinrichtungFehlt:           "macOS 27",
	} {
		gestartet := 0

		_, fehler := geraeteapp(t, einrichtung, "", &gestartet).geraetebauplan()().Fassung(t.Context())
		if !errors.Is(fehler, domain.ErrValidation) || !strings.Contains(fehler.Error(), satz) {
			t.Errorf("%s: %v", einrichtung, fehler)
		}
	}
}

// **Ein Vorschlag startet den Dienst einmal und liest die Antwort.**
func TestEinVorschlagStartetEinmal(t *testing.T) {
	t.Parallel()

	gestartet := 0
	a := geraeteapp(t, applefm.EinrichtungBereit,
		`{"model":"system","choices":[{"finish_reason":"stop","message":{"content":"{\"name\":\"Funke\"}","refusal":null}}]}`,
		&gestartet)

	for range 2 {
		antwort, fehler := a.geraetebauplan()().Vorschlagen(t.Context(), "Anweisung", applefm.Modell, "Auftrag",
			map[string]any{"type": "object"})
		if fehler != nil || antwort.Inhalt != `{"name":"Funke"}` || antwort.Modell != applefm.Modell {
			t.Errorf("%+v %v", antwort, fehler)
		}
	}

	if gestartet != 1 {
		t.Errorf("gestartet: %d", gestartet)
	}
}

// **Eine Ablehnung ist ein eigener Ausgang** — die Quelle markiert sie, und
// der Lauf darüber macht daraus einen Vorschlag ohne Felder (AI-05.AK-06).
func TestEineAblehnungNenntDenRueckfall(t *testing.T) {
	t.Parallel()

	gestartet := 0
	a := geraeteapp(t, applefm.EinrichtungBereit,
		`{"model":"system","choices":[{"finish_reason":"stop","message":{"content":"Diese Anfrage kann nicht bearbeitet werden.","refusal":null}}]}`,
		&gestartet)

	_, fehler := a.geraetebauplan()().Vorschlagen(t.Context(), "", applefm.Modell, "Kampf", nil)
	if !errors.Is(fehler, errKIAbgelehnt) || !strings.Contains(fehler.Error(), "nicht bearbeitet") {
		t.Errorf("%v", fehler)
	}
}

// **Das Gerätemodell rechnet hier** — ohne diese Zeile stufte `ortVon` jeden
// Anbieter außer Ollama als Cloud ein, und ein Vorschlag verlangte eine
// Freigabe, die es für ein Modell auf diesem Mac nicht braucht.
func TestDasGeraetemodellRechnetLokal(t *testing.T) {
	t.Parallel()

	if ort := ortVon(geraeteanbieter, ""); ort != domain.KIOrtLokal {
		t.Errorf("%q", ort)
	}
}
