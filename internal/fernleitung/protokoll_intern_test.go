package fernleitung

// **Kein Kennwort ins Protokoll** — Befund Z1 des Zweitblicks über Phase 6.7
// (NET-Teil).
//
// tailcat schreibt ohne eigenen Logger mit `log.Printf` (v0.6.0 und v0.7.0,
// `tailcat.go`), und die gepinnte Tailscale-Fassung setzt bei einer
// Proxyabsage die ganze Proxyadresse in ihren Fehler — samt Benutzer und
// Kennwort. Über `slog.SetDefault` landet `log` im Protokoll der Anwendung.

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"tailscale.com/derp/derphttp"
	"tailscale.com/net/netmon"
	"tailscale.com/tailcfg"
	"tailscale.com/types/key"
)

// sammler hält, was protokolliert wurde.
type sammler struct {
	schloss sync.Mutex
	zeilen  []string
}

func (s *sammler) schreibe(zeile string) {
	s.schloss.Lock()
	defer s.schloss.Unlock()

	s.zeilen = append(s.zeilen, zeile)
}

func (s *sammler) alles() string {
	s.schloss.Lock()
	defer s.schloss.Unlock()

	return strings.Join(s.zeilen, "\n")
}

func TestEinKennwortInEinerAdresseKommtNichtInsProtokoll(t *testing.T) {
	t.Parallel()

	gesammelt := &sammler{}
	protokollNach(gesammelt.schreibe)("derp: %v",
		fmt.Errorf("proxy %s: 407", "http://anna:geheim@proxy.beispiel:3128"))

	zeile := gesammelt.alles()
	if strings.Contains(zeile, "geheim") || strings.Contains(zeile, "anna") ||
		!strings.Contains(zeile, "proxy.beispiel:3128") {
		t.Errorf("protokolliert: %q", zeile)
	}
}

// **An einer echten Proxyabsage** — der Fehler trägt das Kennwort, das
// Protokoll nicht, und die Klasse bleibt.
func TestEineProxyabsageTraegtDasKennwortNichtHinaus(t *testing.T) {
	adresse, _ := einProxylauscher(t)

	MitProxywahl(func(*http.Request) (*url.URL, error) {
		return url.Parse("http://anna:geheim@" + adresse)
	})

	t.Cleanup(func() { MitProxywahl(func(*http.Request) (*url.URL, error) { return nil, nil }) })

	gesammelt := &sammler{}
	klient := derphttp.NewRegionClient(key.NewNode(), protokollNach(gesammelt.schreibe), netmon.NewStatic(),
		func() *tailcfg.DERPRegion {
			return &tailcfg.DERPRegion{RegionID: 900, RegionCode: "probe", Nodes: []*tailcfg.DERPNode{
				{Name: "900a", RegionID: 900, HostName: "relais.beispiel.invalid"},
			}}
		})

	t.Cleanup(func() { _ = klient.Close() }) //nolint:errcheck // nur ein Lauf

	ctx, abbrechen := context.WithTimeout(t.Context(), 10*time.Second)
	defer abbrechen()

	fehler := klient.Connect(ctx)
	if fehler == nil {
		t.Fatal("die Verbindung kam zustande")
	}

	// **Erst die Quelle**: trüge die Absage das Kennwort nicht, prüfte dieser
	// Lauf nichts (`derphttp_client.go` setzt die Proxyadresse mit `%s`).
	if !strings.Contains(fehler.Error(), "geheim") {
		t.Fatalf("die Absage trägt kein Kennwort mehr — der Lauf prüfte nichts: %v", fehler)
	}

	if strings.Contains(gesammelt.alles(), "geheim") {
		t.Errorf("im Protokoll steht das Kennwort: %q", gesammelt.alles())
	}

	bereinigt := ohneZugangsdatenFehler(fehler)
	if strings.Contains(bereinigt.Error(), "geheim") || !errors.Is(bereinigt, fehler) {
		t.Errorf("bereinigt: %v", bereinigt)
	}
}

func TestTailcatBekommtDenBereinigtenLogger(t *testing.T) {
	t.Parallel()

	server, istServer := serverBauen(nil).(*tailcatserver)
	klient, istKlient := klientBauen("probe").(*tailcatklient)

	if !istServer || !istKlient || server.unten.Logf == nil || klient.unten.Logf == nil {
		t.Error("tailcat schreibt mit log.Printf — ohne Bereinigung")
	}

	if ohneZugangsdatenFehler(nil) != nil {
		t.Error("aus keinem Fehler wurde einer")
	}
}

// **Der Logger von tailcat schreibt, wohin `log` schreibt** — bereinigt.
//
// **Nicht parallel**: er biegt die Ausgabe von `log` um, und die ist eine
// für den ganzen Prozess.
func TestDerTailcatloggerSchreibtBereinigtInsLog(t *testing.T) {
	var puffer strings.Builder

	alt := log.Writer()
	log.SetOutput(&puffer)
	t.Cleanup(func() { log.SetOutput(alt) })

	tailcatprotokoll("derp: %s", "http://anna:geheim@proxy:3128")

	if strings.Contains(puffer.String(), "geheim") || !strings.Contains(puffer.String(), "proxy:3128") {
		t.Errorf("geschrieben: %q", puffer.String())
	}
}

// **Keine Bestandteile der Einladung ins Protokoll** — SEC-21 (#2162). Die
// Einladung ist der öffentliche Knotenschlüssel, der Disco-Schlüssel, den
// tailcat selbst den „nicht zu erratenden Teil" der Adresse nennt, und der
// Pre-Shared Key (`ConnInfo`, `tailcat.go`, Zeilen 155–169, v0.6.0 wie
// v0.7.0). tailcat schreibt die ersten beiden mit jeder `NetworkMap` und jedem
// Disco-Schlüsselwechsel hinaus — gemessen am echten Fernweg
// (`docs/belege/sec21-fernweg.md`).
func TestKeineSchluesselDerEinladungInsProtokoll(t *testing.T) {
	t.Parallel()

	knoten := "nodekey:7d860a407d8075966c2a2f2d59cf97af0c16aafca485fee9abc5206c6786ac18"
	disco := "discokey:83df8291d90e0180ed93738c0249de3815baeb8e95a514a11b825a7b9b4b0b2f"

	gesammelt := &sammler{}
	protokollNach(gesammelt.schreibe)("NetworkMap: {\"Key\":%q,\"DiscoKey\":%q}", knoten, disco)
	protokollNach(gesammelt.schreibe)("disco pub key: %s", disco)

	zeilen := gesammelt.alles()
	for _, geheim := range []string{"7d860a407d8075966c2a", "83df8291d90e0180ed93"} {
		if strings.Contains(zeilen, geheim) {
			t.Errorf("das Protokoll trägt einen Schlüssel der Einladung: %q", zeilen)
		}
	}

	// **Die Art bleibt lesbar** — wer das Protokoll liest, sieht, dass dort
	// ein Schlüssel stand.
	if !strings.Contains(zeilen, "nodekey:…") || !strings.Contains(zeilen, "discokey:…") {
		t.Errorf("die Art des Schlüssels fehlt: %q", zeilen)
	}
}
