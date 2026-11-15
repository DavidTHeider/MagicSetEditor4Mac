package anthropic_test

import (
	"math"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/anthropic"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// **Eine Wartefrist über die Dauer hinaus läuft nicht um** (J05, #2269):
// 9.223.372.037 Sekunden mal `time.Second` sind negativ, und der Klient
// fragte sofort wieder, statt zu warten.
func TestEineUebergrosseWartefristWirdDieLaengste(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			schreiben.Header().Set("retry-after", "9223372037")
			schreiben.WriteHeader(http.StatusTooManyRequests)
			_, _ = schreiben.Write([]byte( //nolint:errcheck // die Gegenstelle
				`{"error":{"type":"rate_limit_error","message":"x"}}`))
		}))
	t.Cleanup(server.Close)

	klient := anthropic.NeuerKlient(server.URL, netz.KlientUeber(server.Client()),
		anthropic.Zugang{Schluessel: "geheim"})

	_, _, _ = klient.Modelle(t.Context()) //nolint:errcheck // gefragt ist die Frist

	if warten := klient.Grenzen().Warten; warten != math.MaxInt64 {
		t.Errorf("gewartet werden soll %v", warten)
	}
}
