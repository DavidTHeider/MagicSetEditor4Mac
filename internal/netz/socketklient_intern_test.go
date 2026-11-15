package netz

import (
	"errors"
	"net/http"
	"testing"
)

// **Der Socketklient wählt nur seinen einen Pfad** — die Zielprüfung für
// einen Weg, der keine Netzadresse hat (#2000, AI-05 #2042).
func TestDerSocketklientWaehltNurSeinenPfad(t *testing.T) {
	t.Parallel()

	kontrolle := socketkontrolle("/kurz/fm.sock")

	if fehler := kontrolle("unix", "/kurz/fm.sock", nil); fehler != nil {
		t.Errorf("der eigene Pfad: %v", fehler)
	}

	for _, ziel := range [][2]string{{"unix", "/anderer.sock"}, {"tcp", "127.0.0.1:1976"}} {
		if fehler := kontrolle(ziel[0], ziel[1], nil); !errors.Is(fehler, ErrZielGesperrt) {
			t.Errorf("%v: %v", ziel, fehler)
		}
	}
}

// **Fehlt der Socket, ist die Anfrage ein Fehler** — und kein Hängen.
func TestEinFehlenderSocketIstEinFehler(t *testing.T) {
	t.Parallel()

	klient := NeuerKlientUeberSocket(UeblicheFristen(), "/gibt/es/nicht.sock")

	anfrage, fehler := http.NewRequestWithContext(t.Context(), http.MethodGet, "http://fm/health", http.NoBody)
	if fehler != nil {
		t.Fatalf("%v", fehler)
	}

	if antwort, fehler := klient.unten.Do(anfrage); fehler == nil {
		_ = antwort.Body.Close() //nolint:errcheck // nur zum Aufräumen
		t.Error("eine Anfrage ohne Socket ging durch")
	}
}
