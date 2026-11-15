package mcp

// **Kann die neueste Fassung diesen Bestand öffnen?** — Werkstück O6 von
// Phase 5.2 (#1430).

import (
	"context"
	"encoding/json"
)

// Hindernis ist ein Grund, warum eine Fassung nicht passt.
type Hindernis struct {
	Art    string `json:"art"`
	Stelle string `json:"stelle"`
}

// Vertraeglichkeit ist die Antwort von `lesen.vertraeglichkeit`.
type Vertraeglichkeit struct {
	Fassung     string      `json:"fassung"`
	Urteil      string      `json:"urteil"`
	Bereit      bool        `json:"bereit"`
	Hindernisse []Hindernis `json:"hindernisse"`
}

// Vertraeglichkeitsauskunft ist die Naht zur Anwendung.
type Vertraeglichkeitsauskunft interface {
	Vertraeglichkeit(ctx context.Context) (Vertraeglichkeit, error)
}

func vertraeglichkeitswerkzeug(naht Vertraeglichkeitsauskunft) []Werkzeug {
	if naht == nil {
		return nil
	}

	return []Werkzeug{{
		Name: "lesen.vertraeglichkeit", Art: Lesend, Idempotent: true,
		Erklaerung: "Sagt, ob die neueste Freigabe DIESEN Bestand öffnen kann. " +
			"»urteil« ist »vertraeglich«, »unvertraeglich« oder »ungeklaert« — " +
			"letzteres heißt, dass die Freigabe kein Manifest trägt, und ist " +
			"KEIN Nein. Verglichen wird nicht eine Schemanummer, sondern Name " +
			"und Prüfsumme jeder verbuchten Wanderung; »hindernisse« nennt " +
			"jede Stelle. »bereit« sagt hier nur, dass die Fassung passt — ob " +
			"ein Wechsel jetzt vorbereitet werden darf, entscheidet das " +
			"Fenster, weil es dafür eine Sicherung macht. HOLT ETWAS AUS DEM " +
			"NETZ. ÄNDERT NICHTS.",
		Schema: objekt(map[string]any{}),
		Ruf: func(ctx context.Context, _ json.RawMessage) (string, error) {
			urteil, fehler := naht.Vertraeglichkeit(ctx)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(urteil)
		},
	}}
}
