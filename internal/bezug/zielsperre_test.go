package bezug_test

// **Ein Bezug wählt nur ins Öffentliche** — SEC-P4
// (#2000).
//
// # Warum die Läufe dieses Pakets die Schleife öffnen
//
// **Jede Gegenstelle hier ist ein `httptest`-Server auf `127.0.0.1`**, und
// der Bezug wählt seit #2000 nur noch öffentlich gerichtete Adressen.
// [TestMain] setzt deshalb den benannten Schalter [netz.SchleifeUmgebung] —
// er öffnet die Schleife und **nur** sie. Der Lauf unten nimmt ihn wieder
// weg und zeigt, was ohne ihn gilt.

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/bezug"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

func TestMain(m *testing.M) {
	if fehler := os.Setenv(netz.SchleifeUmgebung, "an"); fehler != nil {
		panic(fehler)
	}

	os.Exit(m.Run())
}

// **Ohne den Schalter fragt der Bezug die Schleife nicht** — weder für den
// Stand noch für das Archiv, und die Absage steht am Feld der Adresse.
func TestEinBezugVonDerSchleifeWirdAbgewiesen(t *testing.T) {
	t.Setenv(netz.SchleifeUmgebung, "")

	var gefragt atomic.Int64

	innen := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			gefragt.Add(1)
			schreiben.WriteHeader(http.StatusOK)
		}))
	t.Cleanup(innen.Close)

	for name, stand := range map[string]string{
		"Archiv":       "0123456789abcdef0123456789abcdef01234567",
		"Standprüfung": "v1.0",
	} {
		t.Run(name, func(t *testing.T) {
			_, fehler := bezug.Hole(t.Context(), bezug.Repositorium{
				Adresse: innen.URL, Marken: innen.URL + "/tags/", Stand: stand,
			}, t.TempDir(), "probe")

			var befund *domain.FieldError
			if !errors.As(fehler, &befund) || befund.Field != "adresse" {
				t.Fatalf("%v statt eines Befunds an der Adresse", fehler)
			}

			if !errors.Is(fehler, netz.ErrZielGesperrt) {
				t.Errorf("der Grund ist verloren: %v", fehler)
			}
		})
	}

	if n := gefragt.Load(); n != 0 {
		t.Errorf("die Schleife wurde %d-mal gefragt", n)
	}
}
