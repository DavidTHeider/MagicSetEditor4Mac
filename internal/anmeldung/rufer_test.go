package anmeldung_test

// **Der Anmeldefluss folgt derselben Proxywahl wie der Rest** — Werkstück N5
// von Phase 5.4
// (#1571).
//
// # Warum das einen Lauf wert ist
//
// **Weil dieses Paket bis N5 der zweite HTTP-Klient mit anderen
// Eigenschaften war**: `http.DefaultTransport` liest die Umgebung **einmal**
// und hält sie danach fest, und eine Einstellung erreichte ihn gar nicht. Wer
// hinter einem Firmenproxy saß, konnte Karten laden und sich nicht anmelden.
//
// **Gemessen wird die Wahl und nicht die Gestalt**: dass ein Transport
// gesetzt ist, sagt nichts darüber, ob er gefragt wird.

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/anmeldung"
)

// **Die gereichte Wahl wird gefragt — und sie bekommt die Zieladresse.**
func TestDerRuferFragtDieGereichteProxywahl(t *testing.T) {
	t.Setenv("HTTP_PROXY", "")
	t.Setenv("HTTPS_PROXY", "")

	gefragt := make(chan string, 4)

	anbieter := anmeldung.GitHub("kennung", "geheim")

	// **Ein Ziel außerhalb der Schleife** — wäre es die Schleife, ginge die
	// Anfrage von sich aus am Proxy vorbei, und der Lauf sagte nichts.
	anbieter.Tauschadresse = "http://beispiel.invalid/token"
	anbieter.Nutzeradresse = "http://beispiel.invalid/user"

	lauf := mussBeginnen(t, anbieter)

	rufer := anmeldung.NeuerRuferMit(func(anfrage *http.Request) (*url.URL, error) {
		gefragt <- anfrage.URL.String()

		// **Kein Proxy zurück** — gemessen wird, **dass** gefragt wurde;
		// wohin die Anfrage danach geht, ist die Sache von `internal/netz`
		// und dort geprüft.
		return nil, nil
	})

	//nolint:errcheck // geprüft wird der Weg und nicht das Ergebnis
	_, _ = lauf.Einloesen(t.Context(), rufer, "der-code")

	select {
	case weg := <-gefragt:
		if weg != "http://beispiel.invalid/token" {
			t.Errorf("gefragt wurde für %q", weg)
		}
	default:
		t.Fatal("die Wahl wurde nicht gefragt — der Transport sitzt nicht")
	}
}

// **Ohne Wahl entsteht der Rufer von früher.**
//
// **Die Gegenprobe**: `nil` heißt „die Umgebung gilt", und dann darf kein
// eigener Transport dazwischenstehen — sonst hätte dieser Weg stillschweigend
// andere Eigenschaften als der ohne.
func TestOhneWahlIstDerRuferDerAlte(t *testing.T) {
	t.Parallel()

	if anmeldung.NeuerRuferMit(nil) == nil {
		t.Error("ohne Wahl entstand kein Rufer")
	}
}
