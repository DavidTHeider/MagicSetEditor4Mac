// **Profilvorgaben über REST** — Werkstück 15 von Phase 4.12
// (#1351),
// Punkte 4 bis 7.
//
// **Zwei Wege, beide erweiternd** (§16).
//
// # Warum `PUT` auf die Art und nicht auf das Profil
//
// **Weil die Art der Ort ist.** `PUT /profile/{id}/vorgaben/{art}` sagt: hier
// stehen die Werte dieser Art, und derselbe Ruf zweimal führt zum selben
// Zustand. Ein `PUT` auf das Profil müsste alle fünf Arten tragen — und wer
// eine ändern wollte, schriebe vier andere mit.

package rest

import (
	"context"
	"net/http"
)

// Vorgabenzeile ist eine Art samt ihrem Stand.
type Vorgabenzeile struct {
	Art        string `json:"art"`
	Mehrwertig bool   `json:"mehrwertig"`

	Werte    []string         `json:"werte"`
	Moeglich []Vorgabeangebot `json:"moeglich"`

	// Verwaist sind gewählte Werte, die es nicht mehr gibt.
	Verwaist []string `json:"verwaist"`
}

// Vorgabeangebot ist ein Wert, den ein Profil wählen kann.
type Vorgabeangebot struct {
	Wert         string `json:"wert"`
	Beschriftung string `json:"beschriftung"`
}

// Vorgabewerte ist der Rumpf eines `PUT` — **nur die Werte**.
//
// **Kennung und Art stehen im Weg** (`json:"-"` gibt es hier gar nicht): der
// Ort ist die Art, und ein Rumpf, der sie wiederholte, ließe sich mit einer
// anderen füllen.
type Vorgabewerte struct {
	Werte []string `json:"werte"`
}

// Vorgabengesetzt ist die Antwort auf ein Setzen: der ganze Stand — und
// der Grund, aus dem er nach geglücktem Schreiben fehlen kann (§7, Regel 3;
// #1970).
//
// **Dieselbe Gestalt wie bisher** (`{"vorgaben": […]}`) und ein Feld dazu,
// das nur dasteht, wenn etwas fehlt — erweiternd im Sinne von 0033.
type Vorgabengesetzt struct {
	Vorgaben       []Vorgabenzeile `json:"vorgaben"`
	Nachladefehler *Befund         `json:"nachladefehler,omitempty"`
}

// Fremdvorgaben führt die Profilvorgaben.
type Fremdvorgaben interface {
	Vorgaben(ctx context.Context, profilID string) ([]Vorgabenzeile, error)
	VorgabeSetzen(ctx context.Context, profilID, art string, werte []string) (
		Vorgabengesetzt, error)
}

// vorgaben gibt die Vorgaben eines Profils samt ihrem Angebot.
func (d *Dienst) vorgaben(schreiben http.ResponseWriter, anfrage *http.Request) {
	zeilen, fehler := d.naehte.Vorgaben.Vorgaben(
		anfrage.Context(), anfrage.PathValue("profilID"))
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"vorgaben": zeilen})
}

// vorgabeSetzen ersetzt die Werte **einer** Art.
func (d *Dienst) vorgabeSetzen(schreiben http.ResponseWriter, anfrage *http.Request) {
	var rumpf Vorgabewerte

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	werte := rumpf.Werte
	if werte == nil {
		// **Ein fehlendes `werte` räumt ab.** Es als Auslassung zu lesen
		// hieße, einen `PUT` ohne Wirkung zuzulassen — und der ist bei einem
		// Ort, der die Werte *ist*, keine sinnvolle Anfrage.
		werte = []string{}
	}

	gesetzt, fehler := d.naehte.Vorgaben.VorgabeSetzen(anfrage.Context(),
		anfrage.PathValue("profilID"), anfrage.PathValue("art"), werte)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, gesetzt)
}
