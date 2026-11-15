//go:build server

package fenster

// **Der Angriff auf den Servermodus, an Wails' echtem Lauscher** — QA-09
// (#1740),
// Befund P1 in `docs/PENTEST.md`.
//
// # Warum hinter dem Bauschalter
//
// **Weil es den Lauscher nur dort gibt.** `scripts/bauen.sh browser` baut mit
// `server`, und `make dev-browser` startet genau diesen Bau mit dem Bestand,
// der gerade gilt. Die Pipeline fährt ihn nicht; gefahren wird er mit
//
//	go test -tags server ./internal/fenster -run Servermodus
//
// **Gemessen wird an der Wirkung**: ob die Bindung gerufen wurde, nicht ob
// eine Zahl zurückkam. Und die eigene Seite fährt denselben Weg und kommt
// durch — sonst hieße ein Nein nur, dass der Weg gar nicht ging (§10).

import (
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// Probedienst ist eine Bindung, die nur zählt, wie oft sie gerufen wurde.
type Probedienst struct{ gerufen atomic.Int64 }

// Merke zählt einen Ruf.
func (p *Probedienst) Merke() string {
	p.gerufen.Add(1)

	return "gerufen"
}

// rufangaben nennen [Probedienst.Merke] — und rufrumpf ruft sie so, wie
// `@wailsio/runtime` es tut.
const (
	rufangaben = `{"call-id":"probe",` +
		`"methodName":"github.com/DavidTHeider/MagicSetEditor4Mac/internal/fenster.Probedienst.Merke","args":[]}`
	rufrumpf = `{"object":0,"method":0,"args":` + rufangaben + `}`
)

func TestImServermodusRuftNurDieEigeneSeiteEineBindung(t *testing.T) {
	dienst := &Probedienst{}
	adresse := starteServermodus(t, dienst)

	angriffe := map[string]func() *http.Request{
		"Rebinding: der Name des Angreifers": func() *http.Request {
			anfrage := ruf(t, http.MethodPost, adresse, "/wails/runtime", rufrumpf)
			anfrage.Host = "angreifer.example:" + portVon(adresse)
			anfrage.Header.Set("Origin", "http://angreifer.example:"+portVon(adresse))

			return anfrage
		},
		"POST einer fremden Seite": func() *http.Request {
			anfrage := ruf(t, http.MethodPost, adresse, "/wails/runtime", rufrumpf)
			anfrage.Header.Set("Origin", "https://angreifer.example")
			anfrage.Header.Set("Content-Type", "text/plain")

			return anfrage
		},
		"Bild auf den Rufweg": func() *http.Request {
			return ruf(t, http.MethodGet, adresse,
				"/wails/runtime?object=0&method=0&args="+url.QueryEscape(rufangaben), "")
		},
	}

	for name, angriff := range angriffe {
		vorher := dienst.gerufen.Load()

		antwort, fehler := http.DefaultClient.Do(angriff())
		if fehler != nil {
			t.Fatalf("%s: %v", name, fehler)
		}

		_ = antwort.Body.Close() //nolint:errcheck // gelesen wird nur der Status

		if nachher := dienst.gerufen.Load(); nachher != vorher {
			t.Errorf("%s: die Bindung wurde gerufen (Status %d)", name, antwort.StatusCode)
		}
	}

	// **Die eigene Seite kommt durch** — mit Herkunft und Kopf wie im Browser.
	eigene := ruf(t, http.MethodPost, adresse, "/wails/runtime", rufrumpf)
	eigene.Header.Set("Origin", "http://"+adresse)
	eigene.Header.Set("Content-Type", "application/json")
	eigene.Header.Set("Sec-Fetch-Site", "same-origin")

	antwort, fehler := http.DefaultClient.Do(eigene)
	if fehler != nil {
		t.Fatalf("eigene Seite: %v", fehler)
	}

	_ = antwort.Body.Close() //nolint:errcheck // gelesen wird nur der Status

	if dienst.gerufen.Load() != 1 {
		t.Errorf("die eigene Seite rief die Bindung %d-mal (Status %d), erwartet einmal",
			dienst.gerufen.Load(), antwort.StatusCode)
	}
}

// starteServermodus startet Wails im Servermodus mit der Weiche der Anwendung
// und gibt `127.0.0.1:<port>`.
func starteServermodus(t *testing.T, dienst *Probedienst) string {
	t.Helper()

	t.Setenv("HOME", t.TempDir())

	frei, fehler := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0")
	if fehler != nil {
		t.Fatalf("einen freien Port suchen: %v", fehler)
	}

	tcp, istTCP := frei.Addr().(*net.TCPAddr)
	if !istTCP {
		t.Fatalf("kein TCP-Lauscher: %T", frei.Addr())
	}

	port := tcp.Port
	_ = frei.Close() //nolint:errcheck // nur der Port war gefragt

	wails := application.New(application.Options{
		Name:     "Rechnerwache",
		Services: []application.Service{application.NewService(dienst)},
		Server:   application.ServerOptions{Host: "127.0.0.1", Port: port},
		Assets: application.AssetOptions{
			Handler:        http.NotFoundHandler(),
			Middleware:     assetweiche(nil),
			DisableLogging: true,
		},
	})

	beendet := make(chan error, 1)

	go func() { beendet <- wails.Run() }()

	t.Cleanup(func() {
		wails.Quit()

		if fehler := <-beendet; fehler != nil {
			t.Errorf("Wails endete mit %v", fehler)
		}
	})

	adresse := "127.0.0.1:" + strconv.Itoa(port)

	for range 100 {
		antwort, fehler := http.Get("http://" + adresse + "/health") //nolint:noctx // nur das Warten auf den Lauscher
		if fehler == nil {
			_ = antwort.Body.Close() //nolint:errcheck // gelesen wird nur, ob er antwortet

			return adresse
		}

		time.Sleep(20 * time.Millisecond)
	}

	t.Fatal("der Servermodus antwortete nicht")

	return ""
}

// ruf baut eine Anfrage an den Servermodus.
func ruf(t *testing.T, methode, adresse, weg, rumpf string) *http.Request {
	t.Helper()

	anfrage, fehler := http.NewRequestWithContext(t.Context(), methode,
		"http://"+adresse+weg, strings.NewReader(rumpf))
	if fehler != nil {
		t.Fatalf("Anfrage bauen: %v", fehler)
	}

	return anfrage
}

// portVon gibt den Port einer Adresse.
func portVon(adresse string) string {
	_, port, _ := net.SplitHostPort(adresse) //nolint:errcheck // die Adresse ist selbst gebaut

	return port
}
