// **Ein Bild als Symbol nachzeichnen, über MCP** — Werkstück F von Phase 4.11
// (#1340).
//
// Einordnung nach §16: **erweiternd** — ein Werkzeug kommt hinzu.

package mcp

import (
	"context"
	"encoding/json"
)

// Symbolnachzeichnung ist, was MCP dafür braucht.
type Symbolnachzeichnung interface {
	// SymbolNachzeichnen liest ein Bild und legt die Zeichnung ins Set.
	//
	// **Ein Zug und nicht zwei.** Am Fenster stehen die Teile erst in der
	// Zeichenfläche, weil dort jemand sie noch anfassen will; kopflos wäre ein
	// Zwischenergebnis, das niemand hält, nur ein zweiter Aufruf mehr.
	SymbolNachzeichnen(ctx context.Context, setID int64, bildpfad string) (any, error)
}

// symbolnachzeichnen gibt das eine Werkzeug.
func symbolnachzeichnen(quelle Symbolnachzeichnung) []Werkzeug {
	if quelle == nil {
		return nil
	}

	return []Werkzeug{{
		Name: "schreiben.symbolnachzeichnen", Art: Schreibend,
		Erklaerung: "Zeichnet eine Bilddatei als Symbol dieses Sets nach und " +
			"legt die Zeichnung ab. Erwartet ein PNG, JPEG oder GIF mit einer " +
			"klaren Silhouette — hell gegen dunkel. Heraus kommen Formen mit " +
			"Bézierpunkten, die sich im Fenster weiterbearbeiten lassen. " +
			"ÜBERSCHREIBT das Symbol, das das Set bisher hatte.",
		Schema: objekt(map[string]any{
			"setID":    zahl("Das Set, dessen Symbol entsteht."),
			"bildpfad": text("Die Bilddatei auf der Platte."),
		}, "setID", "bildpfad"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var argumente struct {
				SetID    int64  `json:"setID"`
				Bildpfad string `json:"bildpfad"`
			}

			if err := json.Unmarshal(roh, &argumente); err != nil {
				return "", err
			}

			antwort, fehler := quelle.SymbolNachzeichnen(ctx, argumente.SetID, argumente.Bildpfad)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(antwort)
		},
	}}
}
