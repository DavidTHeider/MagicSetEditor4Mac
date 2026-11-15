package fernleitung

// **Die Messung am echten Klienten** — nicht am Haken, sondern daran, dass
// tailscales DERP-Klient in der gepinnten Fassung ihn fragt und dann durch
// den Proxy wählt. Ein Test, der nur prüft, dass der Haken gesetzt ist,
// sagte nichts darüber, ob ihn jemand liest (§10).
//
// **Nur über die Schleifenadresse**: das Relais trägt einen Namen unter
// `.invalid`, und der Proxy ist ein Lauscher hier, der mitschreibt und
// abweist.

import (
	"bufio"
	"context"
	"net"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"tailscale.com/derp/derphttp"
	"tailscale.com/net/netmon"
	"tailscale.com/tailcfg"
	"tailscale.com/types/key"
)

// einProxylauscher nimmt eine Verbindung an, liest ihren Kopf und weist ab.
func einProxylauscher(t *testing.T) (adresse string, gesehen <-chan *http.Request) {
	t.Helper()

	lauscher, fehler := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0")
	if fehler != nil {
		t.Fatal(fehler)
	}

	t.Cleanup(func() { _ = lauscher.Close() }) //nolint:errcheck // nur ein Lauf

	kanal := make(chan *http.Request, 1)

	go func() {
		verbindung, fehler := lauscher.Accept()
		if fehler != nil {
			return
		}

		defer func() { _ = verbindung.Close() }() //nolint:errcheck // nur ein Lauf

		anfrage, fehler := http.ReadRequest(bufio.NewReader(verbindung))
		if fehler != nil {
			return
		}

		kanal <- anfrage

		_, _ = verbindung.Write([]byte("HTTP/1.1 502 Bad Gateway\r\nContent-Length: 0\r\n\r\n")) //nolint:errcheck // die Absage ist eine Höflichkeit
	}()

	return lauscher.Addr().String(), kanal
}

func TestDerDerpklientWaehltDurchDenProxy(t *testing.T) {
	adresse, gesehen := einProxylauscher(t)

	MitProxywahl(func(*http.Request) (*url.URL, error) {
		return url.Parse("http://anna:geheim@" + adresse)
	})

	t.Cleanup(func() { MitProxywahl(func(*http.Request) (*url.URL, error) { return nil, nil }) })

	klient := derphttp.NewRegionClient(key.NewNode(), t.Logf, netmon.NewStatic(),
		func() *tailcfg.DERPRegion {
			return &tailcfg.DERPRegion{RegionID: 900, RegionCode: "probe", Nodes: []*tailcfg.DERPNode{
				{Name: "900a", RegionID: 900, HostName: "relais.beispiel.invalid"},
			}}
		})

	t.Cleanup(func() { _ = klient.Close() }) //nolint:errcheck // nur ein Lauf

	ctx, abbrechen := context.WithTimeout(t.Context(), 10*time.Second)
	defer abbrechen()

	// **Scheitern muss es** — der Lauscher weist ab. Gemessen wird, was bei
	// ihm ankam.
	if fehler := klient.Connect(ctx); fehler == nil {
		t.Fatal("die Verbindung kam zustande")
	}

	select {
	case anfrage := <-gesehen:
		if anfrage.Method != http.MethodConnect || anfrage.Host != "relais.beispiel.invalid:443" {
			t.Errorf("beim Proxy kam %s %s an", anfrage.Method, anfrage.Host)
		}

		if kopf := anfrage.Header.Get("Proxy-Authorization"); !strings.HasPrefix(kopf, "Basic ") {
			t.Errorf("die Anmeldung fehlt: %q", kopf)
		}
	default:
		t.Fatal("beim Proxy kam nichts an")
	}
}
