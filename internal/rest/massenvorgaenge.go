// **Massenvorgänge und Preise in Masse, über den Dienst** — Werkstück 6 von
// Phase 4.5 (#1292).
//
// Einordnung nach §16: **erweiternd** — drei Wege kommen hinzu.

package rest

import (
	"context"
	"net/http"
)

// Massenzeile ist ein Exemplar samt seiner gelesenen Fassung.
type Massenzeile struct {
	ID      int64 `json:"id"`
	Fassung int64 `json:"fassung"`
}

// Massenaenderung sagt, was an den genannten Zeilen zu ändern ist.
//
// **Zeiger heißen „unverändert"** — ein leerer Lagerort ist eine Angabe.
type Massenaenderung struct {
	Zeilen   []Massenzeile `json:"zeilen"`
	Menge    *int          `json:"menge"`
	Zustand  *string       `json:"zustand"`
	Lagerort *string       `json:"lagerort"`
	Wunsch   *bool         `json:"wunsch"`
}

// Massenwerkstatt ist, was der Dienst für Vorgänge über eine Auswahl braucht.
type Massenwerkstatt interface {
	MasseAendern(ctx context.Context, aenderung Massenaenderung) (any, error)
	MasseLoeschen(ctx context.Context, zeilen []Massenzeile) (any, error)
	MassePreise(ctx context.Context, hoechstensStunden int) (any, error)
}

// masseAendern ändert viele Exemplare auf einmal.
func (d *Dienst) masseAendern(schreiben http.ResponseWriter, anfrage *http.Request) {
	var rumpf Massenaenderung

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	bericht, fehler := d.naehte.Masse.MasseAendern(anfrage.Context(), rumpf)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, bericht)
}

// masseLoeschen nimmt viele Exemplare heraus.
//
// **`POST` und nicht `DELETE`**, weil die Zeilen samt Fassungen in den Rumpf
// gehören — und ein `DELETE` hat keinen (derselbe Grund wie bei
// `/sammlung/{id}`, nur andersherum gelöst: dort steht die Fassung in der
// Abfrage, hier sind es fünfzig).
func (d *Dienst) masseLoeschen(schreiben http.ResponseWriter, anfrage *http.Request) {
	var rumpf struct {
		Zeilen []Massenzeile `json:"zeilen"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	bericht, fehler := d.naehte.Masse.MasseLoeschen(anfrage.Context(), rumpf.Zeilen)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, bericht)
}

// massePreise holt einen Schub Preise.
func (d *Dienst) massePreise(schreiben http.ResponseWriter, anfrage *http.Request) {
	var rumpf struct {
		HoechstensStunden int `json:"hoechstensStunden"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	schub, fehler := d.naehte.Masse.MassePreise(
		anfrage.Context(), rumpf.HoechstensStunden)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, schub)
}
