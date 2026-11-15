package mcp

// **Das Werkzeug zum Nachzeichnen** —
// #1340.

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

// nachzeichnerdoppel antwortet, wie die Naht es täte.
type nachzeichnerdoppel struct {
	fehler error
	setID  int64
	pfad   string
}

func (n *nachzeichnerdoppel) SymbolNachzeichnen(
	_ context.Context, setID int64, bildpfad string,
) (any, error) {
	n.setID, n.pfad = setID, bildpfad

	if n.fehler != nil {
		return nil, n.fehler
	}

	return map[string]any{"name": "symbol.mse-symbol", "teile": 1, "punkte": 12}, nil
}

// **Ohne Naht gibt es das Werkzeug nicht.**
//
// Dieselbe Form wie bei den anderen: ein Werkzeug, dessen Naht fehlt, meldete
// sich beim Aufruf statt beim Auflisten — und ein Sprachmodell wählte es.
func TestOhneNahtGibtEsKeinNachzeichnenwerkzeug(t *testing.T) {
	t.Parallel()

	if werkzeuge := symbolnachzeichnen(nil); werkzeuge != nil {
		t.Errorf("%d Werkzeuge ohne Naht", len(werkzeuge))
	}
}

// **Mit Naht ist es genau eines, und es schreibt.**
func TestDasNachzeichnenwerkzeugIstSchreibend(t *testing.T) {
	t.Parallel()

	werkzeuge := symbolnachzeichnen(&nachzeichnerdoppel{})
	if len(werkzeuge) != 1 {
		t.Fatalf("%d Werkzeuge, erwartet 1", len(werkzeuge))
	}

	if werkzeuge[0].Art != Schreibend {
		t.Errorf("Art %v, erwartet Schreibend", werkzeuge[0].Art)
	}

	if werkzeuge[0].Name != "schreiben.symbolnachzeichnen" {
		t.Errorf("Name %q", werkzeuge[0].Name)
	}
}

// **Es reicht Kennung und Pfad durch und gibt die Antwort als JSON.**
func TestDasNachzeichnenwerkzeugReichtDurch(t *testing.T) {
	t.Parallel()

	doppel := &nachzeichnerdoppel{}
	werkzeug := symbolnachzeichnen(doppel)[0]

	antwort, err := werkzeug.Ruf(t.Context(),
		json.RawMessage(`{"setID":7,"bildpfad":"/tmp/k.png"}`))
	if err != nil {
		t.Fatalf("Ruf: %v", err)
	}

	if doppel.setID != 7 || doppel.pfad != "/tmp/k.png" {
		t.Errorf("durchgereicht wurden %d und %q", doppel.setID, doppel.pfad)
	}

	if !strings.Contains(antwort, "symbol.mse-symbol") {
		t.Errorf("die Antwort nennt den Namen nicht: %s", antwort)
	}
}

// **Ein unlesbares Argument ist ein Fehler.**
func TestDasNachzeichnenwerkzeugWeistUnlesbaresAb(t *testing.T) {
	t.Parallel()

	werkzeug := symbolnachzeichnen(&nachzeichnerdoppel{})[0]

	if _, err := werkzeug.Ruf(t.Context(), json.RawMessage(`{"setID":"sieben"}`)); err == nil {
		t.Error("eine Zeichenkette als Kennung kam durch")
	}
}

// **Und ein Fehler der Naht kommt durch.**
func TestDasNachzeichnenwerkzeugReichtDenFehlerDurch(t *testing.T) {
	t.Parallel()

	kaputt := errors.New("die Bilddatei ließ sich nicht öffnen")
	werkzeug := symbolnachzeichnen(&nachzeichnerdoppel{fehler: kaputt})[0]

	_, err := werkzeug.Ruf(t.Context(), json.RawMessage(`{"setID":1,"bildpfad":"x"}`))
	if !errors.Is(err, kaputt) {
		t.Errorf("Fehler %v, erwartet %v", err, kaputt)
	}
}
