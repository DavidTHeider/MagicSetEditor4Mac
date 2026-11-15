// **Die Aktualisierung über die Rückschleife** —
// #1100,
// Werkstück 7.
//
// # Warum es sie hier gibt und bei MCP nicht
//
// §16: eine neue Fähigkeit landet nicht an einer Fläche. Kommandozeile und
// Dienst bedienen **Skripte auf demselben Rechner** — „liegt eine neue Fassung
// vor?" ist für beide dieselbe sinnvolle Frage.
//
// Der MCP-Server bedient ein **Sprachmodell** an Karten und Sets. Ein Werkzeug,
// das 188 MB auf die Platte des Benutzers legt, gehört nicht in diese Hand;
// der Grund steht in `docs/FAEHIGKEITEN.md` und ist keine Erfindung, sondern
// §16, Regel 6: nichts wird still übernommen.
//
// # Einordnung nach §16
//
// **Erweiternd.** Zwei neue Adressen; an keiner bestehenden ändert sich etwas.

package rest

import (
	"context"
	"net/http"
)

// Fassungen ist, was die beiden Endpunkte der Aktualisierung brauchen.
//
// **Mit einem hiesigen Verbund und nicht mit `aktualisierung.Freigabe`**:
// `rest` ist ein Blatt und kennt `domain`. Die Naht wird im Composition Root
// von `cmd/mse-rest` bedient — dieselbe Richtung wie bei allen anderen.
type Fassungen interface {
	// NeuesteFassung sagt, was freigegeben ist.
	NeuesteFassung(ctx context.Context) (Fassungsstand, error)

	// FassungLaden holt das Abbild und gibt seinen Ort.
	FassungLaden(ctx context.Context) (string, error)
}

// Fassungsstand ist die Auskunft über die Fassungen.
type Fassungsstand struct {
	Laufend     string `json:"laufend"`
	Freigegeben string `json:"freigegeben"`
	Seite       string `json:"seite"`
	Neuer       bool   `json:"neuer"`
}

// fassung sagt, welche Fassung freigegeben ist.
func (d *Dienst) fassung(schreiben http.ResponseWriter, anfrage *http.Request) {
	stand, fehler := d.naehte.Fassungen.NeuesteFassung(anfrage.Context())
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, stand)
}

// fassungLaden holt das Abbild ins Downloadverzeichnis.
//
// **201 und nicht 200**: es entsteht etwas, und die Antwort sagt, wo.
func (d *Dienst) fassungLaden(schreiben http.ResponseWriter, anfrage *http.Request) {
	pfad, fehler := d.naehte.Fassungen.FassungLaden(anfrage.Context())
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusCreated, map[string]any{"pfad": pfad})
}
