package mcp

// **Wer antwortet hier** — Werkstück O9 von Phase 5.2
// (#1433).
//
// **Auch wenn `initialize` die Fassung schon nennt.** Der Handschlag geschieht
// **einmal** je Sitzung; ein Modell, das später fragt, hat ihn nicht mehr —
// und ein Werkzeug ist die Form, in der es fragen kann.

import (
	"context"
	"encoding/json"
)

// Meta ist die Antwort von `lesen.meta`.
type Meta struct {
	Fassung   string `json:"fassung"`
	Plattform string `json:"plattform"`
	Go        string `json:"go"`
	Flaeche   string `json:"flaeche"`

	// Stand, Gebaut und Veraendert sagen, aus welchem Commit dieses Programm
	// gebaut ist — zur Diagnose und nicht als zweite Fassung (ENG-06.AK-04,
	// #1731). Bindungsvertrag ist der Abdruck dessen, was zwischen Fenster
	// und Go über den Draht geht.
	Stand           string `json:"stand"`
	Gebaut          string `json:"gebaut"`
	Veraendert      bool   `json:"veraendert"`
	Bindungsvertrag string `json:"bindungsvertrag"`
}

// Metaauskunft ist die Naht zur Anwendung.
type Metaauskunft interface {
	Meta(ctx context.Context) (Meta, error)
}

func metawerkzeug(naht Metaauskunft) []Werkzeug {
	if naht == nil {
		return nil
	}

	return []Werkzeug{{
		Name: "lesen.meta", Art: Lesend, Idempotent: true,
		Erklaerung: "Nennt Fassung, Plattform, Go-Fassung, den Commit, aus dem " +
			"dieses Programm gebaut ist, den Abdruck des Bindungsvertrags und " +
			"die Fläche, über die dieser Ruf kam. Für einen Fehlerbericht und um zwei Prozesse " +
			"an einem Bestand auseinanderzuhalten. Was diese Fläche KANN, " +
			"steht nicht hier, sondern in »tools/list«. ÄNDERT NICHTS.",
		Schema: objekt(map[string]any{}),
		Ruf: func(ctx context.Context, _ json.RawMessage) (string, error) {
			meta, fehler := naht.Meta(ctx)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(meta)
		},
	}}
}
