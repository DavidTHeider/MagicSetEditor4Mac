package rest

// **Der Dienstkatalog im Dienst** — Werkstück A1 von Phase 5.3
// (#1440):
// `GET /dienstkatalog`, ohne Parameter, ohne Nebenwirkung.

import (
	"context"
	"net/http"
)

// Fremdkatalog ist die Naht zu `internal/app`.
type Fremdkatalog interface {
	Dienstkatalog(ctx context.Context) ([]Dienstfaehigkeit, error)
}

// Dienstvertrag ist der Vertrag, wie der Dienst ihn nennt.
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

// dienstkatalog beantwortet `GET /dienstkatalog`.
func (d *Dienst) dienstkatalog(schreiben http.ResponseWriter, anfrage *http.Request) {
	eintraege, fehler := d.naehte.Katalog.Dienstkatalog(anfrage.Context())
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	if eintraege == nil {
		eintraege = []Dienstfaehigkeit{}
	}

	ja(schreiben, http.StatusOK, map[string]any{"eintraege": eintraege})
}
