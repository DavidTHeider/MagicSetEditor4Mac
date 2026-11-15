package mcp

// **`lesen.set.probleme`** — PERF02
// (#2352):
// das eine nennt den Stand, das andere prüft erst.

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

type problemprobe struct {
	geprueft []int64
	fehler   error
}

func (p *problemprobe) Problemstand(context.Context, int64) (Problemstand, error) {
	return Problemstand{Urteil: "ungeprueft", Karten: []Kartenproblem{}}, p.fehler
}

func (p *problemprobe) ProblemePruefen(_ context.Context, setID int64) (Problemstand, error) {
	p.geprueft = append(p.geprueft, setID)

	return Problemstand{Urteil: "aktuell", Sicher: true, Geprueft: 3, Gesamt: 3, Karten: []Kartenproblem{{KarteID: "7", Befunde: []string{"x"}}}, Zahl: 1}, p.fehler
}

func TestDasProblemwerkzeugLiestOderPrueft(t *testing.T) {
	t.Parallel()

	probe := &problemprobe{}
	werkzeuge := problemwerkzeuge(probe)

	if len(werkzeuge) != 2 || werkzeuge[0].Name != "lesen.set.probleme" || werkzeuge[1].Name != "lesen.set.probleme.pruefen" || werkzeuge[0].Art != Lesend || werkzeuge[1].Art != Lesend {
		t.Fatalf("%+v", werkzeuge)
	}

	stand, err := werkzeuge[0].Ruf(t.Context(), json.RawMessage(`{"setID": 4}`))
	if gelesen := standAus(t, stand, err); gelesen.Urteil != "ungeprueft" || gelesen.Karten == nil || len(probe.geprueft) != 0 {
		t.Errorf("ohne pruefen: %+v, geprüft %v", gelesen, probe.geprueft)
	}

	geprueft, err := werkzeuge[1].Ruf(t.Context(), json.RawMessage(`{"setID": 4}`))
	if gelesen := standAus(t, geprueft, err); gelesen.Urteil != "aktuell" || len(gelesen.Karten) != 1 || gelesen.Karten[0].KarteID != "7" || len(probe.geprueft) != 1 || probe.geprueft[0] != 4 {
		t.Errorf("mit pruefen: %+v, geprüft %v", gelesen, probe.geprueft)
	}
}

// standAus liest die Antwort des Werkzeugs.
func standAus(t *testing.T, antwort string, err error) Problemstand {
	t.Helper()

	if err != nil {
		t.Fatal(err)
	}

	var stand Problemstand
	if fehler := json.Unmarshal([]byte(antwort), &stand); fehler != nil {
		t.Fatalf("%v: %s", fehler, antwort)
	}

	return stand
}

func TestDasProblemwerkzeugNenntFehler(t *testing.T) {
	t.Parallel()

	for _, werkzeug := range problemwerkzeuge(&problemprobe{fehler: domain.ErrNotFound}) {
		if _, err := werkzeug.Ruf(t.Context(), json.RawMessage(`{"setID": 9}`)); !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("%s: %v", werkzeug.Name, err)
		}

		for _, argumente := range []string{`{}`, `kein JSON`} {
			if _, err := werkzeug.Ruf(t.Context(), json.RawMessage(argumente)); err == nil {
				t.Errorf("%s: %s ging durch", werkzeug.Name, argumente)
			}
		}
	}
}
