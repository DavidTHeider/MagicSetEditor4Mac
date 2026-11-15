package rest

// **Der Verweis auf einen echten Druck über die Rückschleife** — Werkstück 2
// von Phase 3.6
// (#1167).

import (
	"context"
	"net/http"
	"strings"
)

// Verweissicht ist ein Verweis, wie diese Fläche ihn sieht.
type Verweissicht struct {
	KarteID  string `json:"karteID"`
	Printing string `json:"printing"`

	// DruckID ist die Zeile des Drucks in diesem Bestand (A2, #1441) — als
	// Zeichenkette wie jede Kennung, leer für keine; Vorlaeufig sagt, dass
	// der Druck keinen Alias hat. Beides **erweiternd**: alte Aufrufer lesen
	// weiter `printing`.
	DruckID    string `json:"druckID"`
	Vorlaeufig bool   `json:"vorlaeufig"`

	Name    string `json:"name"`
	Setcode string `json:"setcode"`
	Nummer  string `json:"nummer"`
	Adresse string `json:"adresse"`
	Gesetzt string `json:"gesetzt"`
}

// Fremdverweise ist, was der Dienst von den Verweisen braucht.
type Fremdverweise interface {
	VerweisSetzen(ctx context.Context, karteID int64, printing string) (Verweissicht, error)
	Verweis(ctx context.Context, karteID int64) (*Verweissicht, error)
	VerweisLoesen(ctx context.Context, karteID int64) error
	Verweise(ctx context.Context, setID int64) ([]Verweissicht, error)
}

// verweisZeigen nennt den Druck, auf den eine Karte verweist.
//
// **`GET`, und ohne Verweis kommt `200` mit `null`** — nicht `404`. Eine Karte
// ohne Verweis ist der Normalfall; ein `404` sagte „diese Karte gibt es
// nicht", und das wäre etwas anderes.
func (d *Dienst) verweisZeigen(schreiben http.ResponseWriter, anfrage *http.Request) {
	karteID, gut := kennung(schreiben, anfrage, "karteID")
	if !gut {
		return
	}

	verweis, fehler := d.naehte.Verweise.Verweis(anfrage.Context(), karteID)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"verweis": verweis})
}

// verweisSetzen bezieht eine Karte auf einen Druck.
func (d *Dienst) verweisSetzen(schreiben http.ResponseWriter, anfrage *http.Request) {
	karteID, gut := kennung(schreiben, anfrage, "karteID")
	if !gut {
		return
	}

	var rumpf struct {
		Printing string `json:"printing"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	if strings.TrimSpace(rumpf.Printing) == "" {
		nein(schreiben, http.StatusBadRequest, "validierung",
			"der Druck fehlt — er gehört als \"printing\" in den Rumpf")

		return
	}

	verweis, fehler := d.naehte.Verweise.VerweisSetzen(
		anfrage.Context(), karteID, rumpf.Printing)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"verweis": verweis})
}

// verweisLoesen nimmt den Verweis einer Karte zurück.
func (d *Dienst) verweisLoesen(schreiben http.ResponseWriter, anfrage *http.Request) {
	karteID, gut := kennung(schreiben, anfrage, "karteID")
	if !gut {
		return
	}

	if fehler := d.naehte.Verweise.VerweisLoesen(anfrage.Context(), karteID); fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"karteID": karteID, "geloest": true})
}

// setverweiseZeigen nennt die Verweise aller Karten eines Sets.
func (d *Dienst) setverweiseZeigen(schreiben http.ResponseWriter, anfrage *http.Request) {
	setID, gut := kennung(schreiben, anfrage, "setID")
	if !gut {
		return
	}

	verweise, fehler := d.naehte.Verweise.Verweise(anfrage.Context(), setID)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"verweise": ohneNilVerweise(verweise)})
}

// ohneNilVerweise macht aus einer fehlenden Liste eine leere.
//
// **Weil `null` und `[]` an einer Naht nicht dasselbe sind** (§3): ein
// `omitempty` oder ein `nil`-Slice lässt eine leere Liste verschwinden, und
// die Seite bleibt leer, ohne dass eine Testebene es sieht.
func ohneNilVerweise(verweise []Verweissicht) []Verweissicht {
	if verweise == nil {
		return []Verweissicht{}
	}

	return verweise
}
