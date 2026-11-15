package mcp

// **Die Gesundheit über MCP** — Werkstück O9 von Phase 5.2
// (#1433).
//
// **Ein lesendes Werkzeug, und es ist idempotent**: zweimal gefragt heißt
// zweimal dieselbe Auskunft über denselben Augenblick — es ändert nichts.

import (
	"context"
	"encoding/json"
)

// Fremdzustand ist ein fremder Dienst in der Antwort.
type Fremdzustand struct {
	Dienst     string `json:"dienst"`
	Erreichbar bool   `json:"erreichbar"`
	Grund      string `json:"grund"`
}

// Gesundheit ist die Antwort von `lesen.gesundheit`.
type Gesundheit struct {
	Lebt   bool           `json:"lebt"`
	Bereit bool           `json:"bereit"`
	Grund  string         `json:"grund"`
	Fremde []Fremdzustand `json:"fremde"`
}

// Gesundheitsauskunft ist die Naht zur Anwendung.
type Gesundheitsauskunft interface {
	Gesundheit(ctx context.Context) (Gesundheit, error)
}

func gesundheitswerkzeug(naht Gesundheitsauskunft) []Werkzeug {
	if naht == nil {
		return nil
	}

	return []Werkzeug{{
		Name: "lesen.gesundheit", Art: Lesend, Idempotent: true,
		Erklaerung: "Sagt, ob dieser Prozess lebt und ob er arbeiten kann. " +
			"»bereit« hängt AUSSCHLIESSLICH am eigenen Bestand: »grund« ist " +
			"»bestandZu« (keiner geöffnet), »wanderungLaeuft« (Migrationen " +
			"laufen) oder »bestandUnlesbar« (er antwortet nicht) — leer, wenn " +
			"alles steht. »fremde« nennt die fremden Dienste und ihre " +
			"Erreichbarkeit; ein Ausfall dort macht diesen Prozess NICHT " +
			"unbereit. ÄNDERT NICHTS.",
		Schema: objekt(map[string]any{}),
		Ruf: func(ctx context.Context, _ json.RawMessage) (string, error) {
			gesund, fehler := naht.Gesundheit(ctx)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(gesund)
		},
	}}
}
