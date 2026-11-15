package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"
)

// **Die Quelle der Befehle** — Werkstück U12 von Phase 5.1 (#1423).
func TestDieQuelleFuelltDieBefehleBeimStartOderSagtWarumNicht(t *testing.T) {
	t.Parallel()

	anwendung := NeueApp("").MitBefehlsquelle(func(_ *sql.DB, pfad string) ([]Befehl, error) {
		return []Befehl{{Name: "lesen.probe", Erklaerung: pfad, Ruf: func(context.Context, json.RawMessage) (string, error) { return "", nil }}}, nil
	})
	anwendung.befehleAus(nil)

	if len(anwendung.befehle) != 1 || anwendung.befehle[0].Name != "lesen.probe" {
		t.Errorf("Befehle: %+v", anwendung.befehle)
	}

	// **Eine Quelle, die scheitert, hält den Start nicht auf** — die Zentrale
	// bleibt leer und sagt es beim ersten Blick.
	kaputt := NeueApp("").MitBefehlsquelle(func(*sql.DB, string) ([]Befehl, error) { return nil, errors.New("kein Bestand") })
	kaputt.befehleAus(nil)

	if len(kaputt.befehle) != 0 {
		t.Errorf("kaputte Quelle: %+v", kaputt.befehle)
	}

	// Ohne Quelle bleibt es beim Stand — auch ohne Fehler.
	ohne := NeueApp("")
	ohne.befehleAus(nil)

	if ohne.befehle != nil {
		t.Errorf("ohne Quelle: %+v", ohne.befehle)
	}
}
