package mcp

// **Der Dienstkatalog als Werkzeug** — Werkstück A1 von Phase 5.3
// (#1440).
//
// Ein Sprachmodell, das »such mir die Karte« hört, muss vorher wissen, ob
// die Suche überhaupt gehen kann — und wenn nicht, warum: abgeschaltet,
// ohne Zugang, gesperrt oder für die Frage nicht gewählt. Dieselbe Auskunft
// wie am Fenster, an der Kommandozeile und im Dienst (§16).

import (
	"context"
	"encoding/json"
)

// Fremdkatalog ist die Naht zu `internal/app`.
type Fremdkatalog interface {
	Dienstkatalog(ctx context.Context) ([]Dienstfaehigkeit, error)
}

// Dienstvertrag ist der Vertrag, wie das Werkzeug ihn nennt.
type Dienstvertrag struct {
	Basisadresse string   `json:"basisadresse"`
	Endpunkte    []string `json:"endpunkte"`
	Zugang       string   `json:"zugang"`
	Tarif        string   `json:"tarif"`
	Grenze       string   `json:"grenze"`
	Frist        string   `json:"frist"`
	Bedingungen  string   `json:"bedingungen"`
	Abgerufen    string   `json:"abgerufen"`
	Beleg        string   `json:"beleg"`
}

// Dienstnachweis ist der letzte Lauf gegen den echten Dienst.
type Dienstnachweis struct {
	Lauf                 string   `json:"lauf"`
	Datum                string   `json:"datum"`
	Fassung              string   `json:"fassung"`
	Endpunkte            []string `json:"endpunkte"`
	GiltFuerDieseFassung bool     `json:"giltFuerDieseFassung"`
}

// Dienstfaehigkeit ist eine Fähigkeit eines Dienstes samt Status.
type Dienstfaehigkeit struct {
	Dienst   string          `json:"dienst"`
	Frage    string          `json:"frage"`
	Status   string          `json:"status"`
	Grund    string          `json:"grund"`
	Angabe   string          `json:"angabe"`
	HatWahl  bool            `json:"hatWahl"`
	Gewaehlt bool            `json:"gewaehlt"`
	Wirkung  string          `json:"wirkung"`
	Vertrag  Dienstvertrag   `json:"vertrag"`
	Nachweis *Dienstnachweis `json:"nachweis"`
}

// katalogwerkzeug gibt das Werkzeug — oder nichts, wenn die Naht fehlt.
func katalogwerkzeug(katalog Fremdkatalog) []Werkzeug {
	if katalog == nil {
		return nil
	}

	return []Werkzeug{{
		Name: "lesen.dienstkatalog", Art: Lesend, Idempotent: true,
		Erklaerung: "Nennt jede Fähigkeit der fremden Kartendienste samt Status. " +
			"ÄNDERT NICHTS und schickt keine Anfrage. »status« ist bereit, aus, " +
			"zugangFehlt, gestoert, gesperrt oder nichtGewaehlt; »grund« sagt, " +
			"was zuerst dagegen steht. »vertrag« nennt Adresse, Endpunkte, Zugang, " +
			"Tarif, Grenze, Frist und Bedingungen mit Abrufdatum und Beleg; " +
			"»nachweis« den letzten Lauf gegen den echten Dienst — oder null, " +
			"wenn es keinen gibt. Ein Nachweis aus einer anderen Fassung ist keine " +
			"Prüfung dieser: »giltFuerDieseFassung« sagt es.",
		Schema: objekt(map[string]any{}),
		Ruf: func(ctx context.Context, _ json.RawMessage) (string, error) {
			eintraege, fehler := katalog.Dienstkatalog(ctx)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(map[string]any{"eintraege": liste(eintraege)})
		},
	}}
}
