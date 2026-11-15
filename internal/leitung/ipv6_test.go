package leitung_test

// **Dieselbe Leitung über IPv6** — NET-02.AK-01
// (#1706):
// „IPv4/IPv6 … ergänzen die Matrix".
//
// Die Peeradresse nimmt ein IPv6-Literal an (`domain.PruefeAdresse` nennt
// `[2001:db8::1]:41234` als Beispiel), und bis hierhin ging kein Lauf über
// IPv6 — die Zusage stand in einer Fehlermeldung und nirgends sonst.
//
// **Was er nicht zeigt**: ein IPv6-Netz zwischen zwei Rechnern. Er geht über
// `[::1]`, die Rückschleife.

import (
	"net"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/leitung"
)

func TestUeberIPv6KommtEineNachrichtAn(t *testing.T) {
	t.Parallel()

	// **Ohne IPv6 auf der Rückschleife gibt es nichts zu messen** — dann
	// sagt der Lauf das, statt still grün zu sein.
	probe, fehler := (&net.ListenConfig{}).Listen(t.Context(), "tcp6", "[::1]:0")
	if fehler != nil {
		t.Skipf("diese Maschine hat kein IPv6 auf der Rückschleife: %v", fehler)
	}

	_ = probe.Close() //nolint:errcheck // nur eine Probe

	hiesige, dortige := zweiIdentitaeten(t)

	horcher, fehler := leitung.Horche(t.Context(), "[::1]:0", dortige)
	if fehler != nil {
		t.Fatalf("Horche: %v", fehler)
	}

	t.Cleanup(func() { _ = horcher.Schliesse() }) //nolint:errcheck // nur ein Lauf

	if !strings.HasPrefix(horcher.Adresse(), "[::1]:") {
		t.Fatalf("gehorcht wird auf %q", horcher.Adresse())
	}

	angenommen := make(chan *leitung.Leitung, 1)

	go func() {
		dort, nimmfehler := horcher.Nimm(t.Context())
		if nimmfehler != nil {
			close(angenommen)

			return
		}

		angenommen <- dort
	}()

	hier, fehler := leitung.Waehle(t.Context(), horcher.Adresse(), hiesige, dortige.Kennung())
	if fehler != nil {
		t.Fatalf("Waehle: %v", fehler)
	}

	t.Cleanup(func() { _ = hier.Schliesse() }) //nolint:errcheck // nur ein Lauf

	dort, offen := <-angenommen
	if !offen {
		t.Fatal("Nimm kam ohne Leitung zurück")
	}

	t.Cleanup(func() { _ = dort.Schliesse() }) //nolint:errcheck // nur ein Lauf

	if sendefehler := hier.Sende(t.Context(), []byte("über sechs")); sendefehler != nil {
		t.Fatalf("Sende: %v", sendefehler)
	}

	gelesen, fehler := dort.Empfange(t.Context())
	if fehler != nil || string(gelesen) != "über sechs" {
		t.Fatalf("angekommen ist %q, %v", gelesen, fehler)
	}
}
