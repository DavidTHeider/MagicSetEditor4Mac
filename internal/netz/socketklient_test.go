package netz_test

import (
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// **Ein Klient über einen Unix-Socket** — AI-05
// (#2042,
// Vermerk 0166). `fm serve --socket` ist der einzige Weg zu Apples Gerätemodell,
// der keinen Port öffnet.
//
// # Was er zusagt
//
//   - **Jede Anfrage geht an den Socket**, gleich welcher Name in der Adresse
//     steht.
//   - **Kein Vermittlungsrechner**, auch wenn die Umgebung einen nennt — ein
//     Proxy für einen Socket auf diesem Rechner schickte die Anfrage hinaus.
//   - **Keine Weiterleitung**: eine Antwort 3xx kommt als solche zurück.

// socketDienst lauscht an einem kurzen Pfad — `t.TempDir()` wäre unter macOS
// länger als die 104 Bytes, die ein Unix-Socket fasst.
func socketDienst(t *testing.T, behandeln http.HandlerFunc) string {
	t.Helper()

	ordner, fehler := os.MkdirTemp("", "ns")
	if fehler != nil {
		t.Fatalf("Ordner: %v", fehler)
	}

	t.Cleanup(func() { _ = os.RemoveAll(ordner) }) //nolint:errcheck // nur zum Aufräumen

	pfad := filepath.Join(ordner, "s")

	horcher, fehler := (&net.ListenConfig{}).Listen(t.Context(), "unix", pfad)
	if fehler != nil {
		t.Fatalf("horchen: %v", fehler)
	}

	dienst := &http.Server{Handler: behandeln, ReadHeaderTimeout: netz.UeblicheFristen().Kopf}
	go func() { _ = dienst.Serve(horcher) }() //nolint:errcheck // endet mit Close

	t.Cleanup(func() { _ = dienst.Close() }) //nolint:errcheck // nur zum Aufräumen

	return pfad
}

func TestEinSocketklientFragtDenSocket(t *testing.T) {
	t.Setenv("HTTP_PROXY", "http://127.0.0.1:9")
	t.Setenv("HTTPS_PROXY", "http://127.0.0.1:9")

	pfad := socketDienst(t, func(antwort http.ResponseWriter, anfrage *http.Request) {
		if anfrage.URL.Path == "/weiter" {
			http.Redirect(antwort, anfrage, "http://example.com/", http.StatusFound)

			return
		}

		_, _ = antwort.Write([]byte(`{"weg":"` + anfrage.URL.Path + `"}`)) //nolint:errcheck,gosec // Probe
	})

	klient := netz.NeuerKlientUeberSocket(netz.UeblicheFristen(), pfad)

	antwort, fehler := netz.Hole(t.Context(), klient, nil, netz.Kenner{}, "http://irgendwer/gesund", 1<<10)
	if fehler != nil {
		t.Fatalf("holen: %v", fehler)
	}

	if antwort.Status != http.StatusOK || string(antwort.Rumpf) != `{"weg":"/gesund"}` {
		t.Errorf("%d %s", antwort.Status, antwort.Rumpf)
	}

	weiter, fehler := netz.Hole(t.Context(), klient, nil, netz.Kenner{}, "http://irgendwer/weiter", 1<<10)
	if fehler == nil && weiter.Status != http.StatusFound {
		t.Errorf("eine Weiterleitung wurde befolgt: %d", weiter.Status)
	}
}
