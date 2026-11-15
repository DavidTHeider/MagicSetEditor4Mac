package app

// **Jede Bindung der Tafel weist ab, wenn ihre Schalter aus sind** —
// OPS-10.AK-02 (#1908).
//
// **Über die Reflexion und nicht über eine Auswahl**: eine Liste von Hand
// prüfte, was jemand für wichtig hielt. Gerufen wird jede Bindung mit
// Nullwerten — das Tor steht vor jeder Arbeit, und eine Bindung, die
// trotzdem etwas täte, hätte ihr Tor verfehlt und fiele hier auf.

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestJedeBindungDerTafelWeistAbgeschaltetAb(t *testing.T) {
	t.Parallel()

	alleAus := map[string]string{}
	for _, schalter := range domain.Schalterverzeichnis() {
		alleAus[domain.SchalterEinstellung(schalter.Kennung)] = domain.SchalterAus
	}

	anwendung := NeueApp("")
	anwendung.MitEinstellungen(&merkendeEinstellungen{werte: alleAus})

	gebunden := reflect.ValueOf(NeueBindungen(anwendung))
	geprueft := 0

	for weg := range domain.EingebauteWegetafel()[domain.FlaecheFenster] {
		methode := gebunden.MethodByName(weg)
		if !methode.IsValid() {
			// **Eine Hilfsmethode im Katalog** — `dienstStarten` etwa — ist
			// keine Bindung; die Tafel führt sie, das Fenster ruft sie nicht.
			continue
		}

		geprueft++

		t.Run(weg, func(t *testing.T) {
			pruefeAbgewiesen(t, methode)
		})
	}

	if geprueft < 200 {
		t.Errorf("nur %d Bindungen geprüft — die Tafel greift daneben", geprueft)
	}
}

// pruefeAbgewiesen ruft eine Bindung mit Nullwerten und verlangt die
// Abweisung des Tors.
func pruefeAbgewiesen(t *testing.T, methode reflect.Value) {
	t.Helper()

	defer func() {
		grund := recover()

		panik, istTor := grund.(torfehler)
		if !istTor {
			t.Fatalf("kein Tor: %v", grund)
		}

		var fehler Fehler
		if entpackt := json.Unmarshal([]byte(strings.TrimPrefix(panik.Error(), Fehlermarke)), &fehler); entpackt != nil ||
			fehler.Code != CodeAbgeschaltet || fehler.Werte["schalter"] == "" {
			t.Errorf("%+v, %v", fehler, entpackt)
		}
	}()

	// **Der Rufkontext ist kein Nullwert** (BK-01, #2109) — Wails füllt ihn
	// immer.
	argumente := make([]reflect.Value, methode.Type().NumIn())
	for stelle := range argumente {
		argumente[stelle] = reflect.Zero(methode.Type().In(stelle))
		if methode.Type().In(stelle) == reflect.TypeFor[context.Context]() {
			argumente[stelle] = reflect.ValueOf(t.Context())
		}
	}

	methode.Call(argumente)
}
