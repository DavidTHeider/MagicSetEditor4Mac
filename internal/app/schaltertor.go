package app

// **Das Tor der Schalter an der Bindungsgrenze** — OPS-10
// (#1908).
//
// # Warum mit einer Panik, und warum mit einer Marke
//
// **Dieselbe Stelle wie das Tor der Sitzung** (`sitzungstor.go`): Wails
// kennt keinen Haken vor einem Aufruf, und jede gebundene Methode kommt nur
// über [Bindungen.app] an die Anwendung. **Eine Panik gibt Wails aber nur als
// Satz weiter** — `<Methode>: panic: <Error()>` (beta.23,
// `(*BoundMethod).Call`, nachgelesen am 2026-10-01). Damit das Fenster
// dieselbe Abweisung bekommt wie die drei Flächen ohne Fenster, trägt der
// Satz den Fehler als JSON hinter einer Marke, und `frontend/src/anbindung.ts`
// packt ihn wieder aus.
//
// # Woher das Tor den Weg kennt
//
// **Aus dem Aufrufrahmen**: die äußerste exportierte Methode von [Bindungen]
// ist die, die Wails gerufen hat — auch dann, wenn eine Hilfsmethode oder
// ein Abschluss darunter [Bindungen.app] ruft. Unexportierte Helfer und
// Abschlüsse zählen ohnehin nicht; **die äußerste statt der innersten** ist
// heute folgenlos (gemessen am 2026-10-01: nur `SetSichern` ruft eine zweite
// Bindung, `SetSpeichernUnter`, und beide sind Kern) und steht für die erste
// Bindung, die eine andere ruft.

import (
	"context"
	"encoding/json"
	"go/token"
	"reflect"
	"runtime"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Fehlermarke steht vor dem Fehler, den ein Tor als Panik wirft.
const Fehlermarke = "mse4mac-fehler:"

// torfehler ist eine Panik, die einen übersetzten Fehler trägt.
type torfehler struct{ fehler *Fehler }

// Error gibt die Marke und den Fehler als JSON.
func (t torfehler) Error() string {
	roh, _ := json.Marshal(t.fehler) //nolint:errcheck // ein Fehler aus Zeichenketten lässt sich immer schreiben

	return Fehlermarke + string(roh)
}

// bindungsvorsatz ist der Name, unter dem Go die Methoden von [Bindungen]
// im Aufrufrahmen führt.
var bindungsvorsatz = reflect.TypeFor[Bindungen]().PkgPath() + ".(*Bindungen)."

// tore sind beide Tore einer gebundenen Methode: erst die Sitzung, dann die
// Schalter — wer gesperrt ist, erfährt nicht, was abgeschaltet ist.
func (a *App) tore(ctx context.Context) {
	a.lebenstor(ctx)
	a.sitzungstor()
	a.schaltertor(ctx)
}

// lebenstor hält an, solange kein Bestand offen sein kann.
//
// **Nach dem Beginn des Schlusses nichts mehr** (BK-07a, #2115) — als
// Abbruch, vor jeder anderen Frage: der Bestand geht gleich zu. **Und
// während eines Tauschs auch nicht** (A02, #2216) — aber als „belegt", denn
// das vergeht.
//
// **Auch für die Methoden ohne Sitzung** ([Bindungen.frei]): sie fragen
// keine Sitzung, lesen aber denselben Bestand. Bis A02 gingen sie am Tor
// ganz vorbei und trafen während eines Tauschs auf `sql: database is
// closed`.
func (a *App) lebenstor(ctx context.Context) {
	switch a.lauf.tor.Load() {
	case torEndet:
		panic(torfehler{fehler: Uebersetze(ctx, errAnwendungEndet)})
	case torTausch:
		panic(torfehler{fehler: Uebersetze(ctx, errBestandWirdGetauscht)})
	}
}

// schaltertor hält eine gebundene Methode an, deren Weg abgeschaltet ist.
func (a *App) schaltertor(ctx context.Context) {
	methode := gebundeneMethode()
	if methode == "" {
		return
	}

	ctx = a.rufkontext(ctx)

	// **Die Wache beendet nebenbei, was ein anderer Prozess abgeschaltet
	// hat** (OPS-10b, #1978): `mse`, MCP und REST setzen dieselben Schalter,
	// und das Fenster erfährt es erst beim nächsten Weg, der fragt.
	if fehler := (Schalterwache{einstellungen: a.einstellungen, gelesen: a.laeufeBeenden}).Pruefe(
		a.imProfil(ctx), domain.FlaecheFenster, methode); fehler != nil {
		panic(torfehler{fehler: Uebersetze(ctx, fehler)})
	}
}

// gebundeneMethode gibt die äußerste exportierte Methode von [Bindungen] im
// Aufrufrahmen — leer, wenn keine darin steht.
func gebundeneMethode() string {
	rahmen := make([]uintptr, 64)
	folge := runtime.CallersFrames(rahmen[:runtime.Callers(3, rahmen)])
	methode := ""

	for {
		einer, weiter := folge.Next()

		if name, gefunden := strings.CutPrefix(einer.Function, bindungsvorsatz); gefunden &&
			token.IsExported(name) && !strings.Contains(name, ".") {
			methode = name
		}

		if !weiter {
			return methode
		}
	}
}

// Schalter gibt die Schalter und ihren Stand — **Kern**, damit ein
// abgeschalteter Bereich wieder einzuschalten ist.
func (b *Bindungen) Schalter(ctx context.Context) SchalterAntwort {
	return b.app(ctx).Schalter(b.ruf(ctx))
}
