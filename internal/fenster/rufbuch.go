package fenster

// **Das Rufbuch: welche Bindungen das Fenster ruft, wie oft, wie lange und
// mit welchem Ausgang** — B01
// (#2220).
//
// # Wo es mitschreibt
//
// **Als äußerste Hülle des Asset-Servers.** Wails 3 schickt jeden Ruf einer
// gebundenen Methode als `POST /wails/runtime` (`transport_http.go`,
// v3.0.0-beta.27), und die Hülle aus `AssetOptions.Middleware` steht in
// Wails' Kette **vor** dem Transport (`application.go`, `middlewares`). Sie
// sieht also den Ruf, wie er ankommt, und die Antwort, wie sie hinausgeht —
// ohne einen Eingriff in die Seite oder in die Bindungen.
//
// # Wann es mitschreibt
//
// **Nur im Abnahmebau** (`abnahme_darwin.go`); sonst gibt es die Hülle gar
// nicht (`abnahme_ohne.go`). Die Hüllenprobe holt den Abzug über die
// Leitung (`rufbuch`) und legt ihn je Drehbuch ab.
//
// # Was es nicht sieht
//
// **Zerteilte Rufe** (Wails' `x-wails-chunk-*`): ihr Name steht im
// zusammengesetzten Rumpf, nicht im einzelnen Teil. Sie werden als
// `zerteilt` gezählt, damit sie nicht still fehlen.

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// bindungsvorsatz steht vor dem Namen jeder gebundenen Methode.
const bindungsvorsatz = "github.com/DavidTHeider/MagicSetEditor4Mac/internal/app.Bindungen."

// rufgrenze ist, wie viel einer Antwort vollständig gelesen wird — darüber
// liest der Späher den Strom mit (`rufausgang.go`) —, und endgrenze, wie
// viel er vom Wert des Fehlerfelds höchstens behält.
const (
	rufgrenze = 1 << 16
	endgrenze = 1 << 12
)

// Rufzahl ist, was das Rufbuch über eine Bindung weiß.
type Rufzahl struct {
	Anzahl    int            `json:"anzahl"`
	Ausgaenge map[string]int `json:"ausgaenge"`
	DauerMs   float64        `json:"dauerMs"`
}

// rufbuch ist die Mitschrift.
type rufbuch struct {
	schloss    sync.Mutex
	eintraege  map[string]*Rufzahl
	stoerungen map[string]Stoerung
}

func neuesRufbuch() *rufbuch { return &rufbuch{eintraege: map[string]*Rufzahl{}} }

// Huelle schreibt jeden Bindungsruf mit, der durch `innen` geht.
func (b *rufbuch) Huelle(innen http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost || req.URL.Path != "/wails/runtime" {
			innen.ServeHTTP(rw, req)

			return
		}

		rumpf, _ := io.ReadAll(req.Body) //nolint:errcheck // ein halber Rumpf geht wie er ist weiter
		req.Body = io.NopCloser(bytes.NewReader(rumpf))

		name := methodeAus(rumpf, req.Header.Get("x-wails-chunk-id") != "")
		if name == "" {
			innen.ServeHTTP(rw, req)

			return
		}

		if b.stoeren(rw, req, name) {
			return
		}

		mitschrift := &mitschreiber{ResponseWriter: rw, status: http.StatusOK}
		beginn := time.Now()
		innen.ServeHTTP(mitschrift, req)
		b.eintragen(name, mitschrift.ausgang(), time.Since(beginn))
	})
}

func (b *rufbuch) eintragen(name, ausgang string, dauer time.Duration) {
	b.schloss.Lock()
	defer b.schloss.Unlock()

	eintrag := b.eintraege[name]
	if eintrag == nil {
		eintrag = &Rufzahl{Ausgaenge: map[string]int{}}
		b.eintraege[name] = eintrag
	}

	eintrag.Anzahl++
	eintrag.Ausgaenge[ausgang]++
	eintrag.DauerMs += float64(dauer.Microseconds()) / 1000
}

// Abzug gibt eine Kopie der Mitschrift — und leert sie, wenn verlangt.
func (b *rufbuch) Abzug(leeren bool) map[string]Rufzahl {
	b.schloss.Lock()
	defer b.schloss.Unlock()

	kopie := make(map[string]Rufzahl, len(b.eintraege))
	for name, eintrag := range b.eintraege {
		ausgaenge := make(map[string]int, len(eintrag.Ausgaenge))
		for art, zahl := range eintrag.Ausgaenge {
			ausgaenge[art] = zahl
		}

		kopie[name] = Rufzahl{Anzahl: eintrag.Anzahl, Ausgaenge: ausgaenge, DauerMs: eintrag.DauerMs}
	}

	if leeren {
		b.eintraege = map[string]*Rufzahl{}
	}

	return kopie
}

// methodeAus liest den Namen der Bindung aus dem Rumpf eines Laufzeitrufs —
// leer, wenn es keiner ist.
func methodeAus(rumpf []byte, zerteilt bool) string {
	if zerteilt {
		return "zerteilt"
	}

	var ruf struct {
		Args struct {
			MethodName string `json:"methodName"`
		} `json:"args"`
	}

	if json.Unmarshal(rumpf, &ruf) != nil {
		return ""
	}

	name, gebunden := strings.CutPrefix(ruf.Args.MethodName, bindungsvorsatz)
	if !gebunden {
		return ""
	}

	return name
}
