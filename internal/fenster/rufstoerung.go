package fenster

// **Störungen am Rufbuch** — B03
// (#2222).
//
// Die Zustände „laden" und „Fehler" jeder Seite am gebauten Programm
// herzustellen, ohne eine Seite anzufassen: die Hülle hält einen Ruf um eine
// Verzögerung auf, oder sie beantwortet ihn selbst mit einer fachlichen
// Ablehnung — **in der Gestalt, die jede Antwort hat** (`fehler` mit Code,
// Schlüssel und Satz), sodass die Seite denselben Weg nimmt wie bei einem
// echten Fehler.
//
// **Nur im Abnahmebau** — wie das Rufbuch selbst (`abnahme_darwin.go`).

import (
	"encoding/json"
	"math"
	"net/http"
	"reflect"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// Stoerung sagt, was die Hülle mit den Rufen einer Bindung tut.
type Stoerung struct {
	// Bindung ist der Name der Methode ohne Vorsatz, etwa `Sets`.
	Bindung string `json:"bindung"`

	// VerzoegerungMs hält den Ruf so lange auf, bevor er durchgeht.
	VerzoegerungMs int `json:"verzoegerungMs,omitempty"`

	// Fehler beantwortet den Ruf mit einer fachlichen Ablehnung — er geht
	// nicht durch.
	Fehler bool `json:"fehler,omitempty"`

	// Panik beantwortet den Ruf **so, wie Wails eine Bindungspanik schickt**
	// — 422, `kind: RuntimeError`, die Meldung mit Paketpfad
	// (`transport_http.go`, `httpError`; `bindings.go`, `Call`,
	// v3.0.0-beta.27). Für F03 (#2342): was das Fenster daraus macht, ohne
	// einen echten Fehler in eine Bindung zu legen.
	Panik bool `json:"panik,omitempty"`
}

// bindungspfad ist der Paketpfad, unter dem Wails die Bindungen führt.
var bindungspfad = reflect.TypeFor[app.Bindungen]().PkgPath()

// stoerungsantwort ist die Ablehnung, die eine gestörte Bindung gibt.
const stoerungsantwort = `{"fehler":{"code":"unbekannt","schluessel":"unbekannt","meldung":"Störung der Abnahme"}}`

// Stoeren setzt die Störungen — `nil` hebt alle auf.
func (b *rufbuch) Stoeren(stoerungen []Stoerung) {
	b.schloss.Lock()
	defer b.schloss.Unlock()

	b.stoerungen = map[string]Stoerung{}
	for _, stoerung := range stoerungen {
		b.stoerungen[stoerung.Bindung] = stoerung
	}
}

// stoerungFuer gibt die Störung einer Bindung, wenn es eine gibt.
func (b *rufbuch) stoerungFuer(name string) (Stoerung, bool) {
	b.schloss.Lock()
	defer b.schloss.Unlock()

	stoerung, gestoert := b.stoerungen[name]

	return stoerung, gestoert
}

// stoeren wendet eine Störung an und sagt, ob der Ruf damit beantwortet ist.
// **Ein abgebrochener Ruf wartet die Verzögerung nicht ab** — sonst hielte
// eine überholte Vorschau die nächste auf.
func (b *rufbuch) stoeren(rw http.ResponseWriter, req *http.Request, name string) bool {
	stoerung, gestoert := b.stoerungFuer(name)
	if !gestoert {
		return false
	}

	if stoerung.VerzoegerungMs > 0 {
		uhr := time.NewTimer(stoerdauer(stoerung.VerzoegerungMs))
		defer uhr.Stop()

		select {
		case <-uhr.C:
		case <-req.Context().Done():
			return true
		}
	}

	if stoerung.Panik {
		roh, _ := json.Marshal(map[string]string{ //nolint:errcheck // zwei Zeichenketten lassen sich immer schreiben
			"message": bindungspfad + ".Bindungen." + name + ": panic: Probepanik",
			"kind":    "RuntimeError",
		})

		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = rw.Write(roh) //nolint:errcheck // eine abgebrochene Verbindung hat niemanden mehr, dem es fehlte
		b.eintragen(name, "gestoert", 0)

		return true
	}

	if !stoerung.Fehler {
		return false
	}

	rw.Header().Set("Content-Type", "application/json")
	_, _ = rw.Write([]byte(stoerungsantwort)) //nolint:errcheck // eine abgebrochene Verbindung hat niemanden mehr, dem es fehlte
	b.eintragen(name, "gestoert", 0)

	return true
}

// stoerdauer ist die Verzögerung aus Millisekunden — **sättigend** (Befund 1
// des Nachlaufs über J05, #2269): ein Überlauf hätte die beabsichtigte
// Verzögerung aufgehoben. Dieselbe Rechnung wie `domain.ZeitspanneHoechstens`;
// die Hülle kennt `domain` nicht (`internal/schichten`).
func stoerdauer(ms int) time.Duration {
	switch {
	case ms <= 0:
		return 0
	case int64(ms) > math.MaxInt64/int64(time.Millisecond):
		return math.MaxInt64
	default:
		return time.Duration(ms) * time.Millisecond
	}
}
