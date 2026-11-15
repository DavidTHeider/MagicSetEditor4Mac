package mcp

// **Bereiche und Funktionen abschalten, am Werkzeugserver** — OPS-10
// (#1908).
//
// **Die Frage stellt die Wurzel**, wie bei der Schreibwache: `internal/mcp`
// weiß nicht, was abgeschaltet ist. Es fragt je Ruf mit dem Namen des
// Werkzeugs, und die Antwort kommt als Fehler des Werkzeugs zurück.

import (
	"context"
	"encoding/json"
)

// MitSchaltern setzt die Frage, die jedes Werkzeug vor seinem Lauf stellt.
func (s *Server) MitSchaltern(naht func(ctx context.Context, werkzeug string) error) *Server {
	s.schalter = naht

	return s
}

// abgeschaltet sagt, ob dieses Werkzeug gerade abgeschaltet ist.
func (s *Server) abgeschaltet(ctx context.Context, werkzeug Werkzeug) error {
	if s.schalter == nil {
		return nil
	}

	return s.schalter(ctx, werkzeug.Name)
}

// Schalterauskunft ist, was `lesen.schalter` braucht.
type Schalterauskunft interface {
	Schalter(ctx context.Context) ([]Schalter, error)
}

// Schalter ist ein Schalter mit seinem wirksamen Stand.
type Schalter struct {
	Kennung string `json:"kennung"`
	Art     string `json:"art"`
	Reife   string `json:"reife"`
	Vorgabe string `json:"vorgabe"`
	An      bool   `json:"an"`
}

func schalterwerkzeug(quelle Schalterauskunft) []Werkzeug {
	if quelle == nil {
		return nil
	}

	return []Werkzeug{{
		Name: "lesen.schalter", Art: Lesend, Idempotent: true,
		Erklaerung: "Nennt die Bereiche und Funktionen, die sich abschalten lassen, " +
			"je mit Art (bereich, funktion), Reife (stabil, vorschau, auslaufend), " +
			"Vorgabe und ob er an ist. Ein Werkzeug eines abgeschalteten Bereichs " +
			"weist mit »dieser Bereich ist abgeschaltet« ab. ÄNDERT NICHTS: " +
			"geschaltet wird über schreiben.einstellung mit schalter.<kennung>.",
		Schema: objekt(nil),
		Ruf: func(ctx context.Context, _ json.RawMessage) (string, error) {
			stand, err := quelle.Schalter(ctx)
			if err != nil {
				return "", err
			}

			return alsJSON(stand)
		},
	}}
}
