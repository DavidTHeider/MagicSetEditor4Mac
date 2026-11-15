package app

// **Der Weg zum Relais nimmt die Proxyeinstellung** — NET-02.AK-01
// (#1706).
// Dass tailscale sie dann fragt, misst `internal/fernleitung`; hier steht,
// dass die Fassade ihm dieselbe Wahl gibt wie jedem anderen Klienten.

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestDieFernleitungBekommtDieProxyeinstellung(t *testing.T) {
	t.Parallel()

	anwendung := NeueApp("")
	anwendung.MitEinstellungen(netzeinstellungsprobe{werte: map[string]string{
		domain.NetzProxyHTTPS: "http://proxy.firma:3128",
	}})

	adresse, fehler := anwendung.fernproxywahl()(&http.Request{
		URL: &url.URL{Scheme: "https", Host: "derp1.tailscale.com"},
	})
	if fehler != nil || adresse == nil || adresse.Host != "proxy.firma:3128" {
		t.Errorf("die Wahl gab %v, %v", adresse, fehler)
	}
}
