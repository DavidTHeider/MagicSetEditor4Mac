// **Das Zufallspack über die Rückschleife** —
// #889,
// Befund 7 des Zweitblicks über Phase W.

package rest

import (
	"context"
	"net/http"
)

// Pack ist, was der Dienst dafür braucht.
type Pack interface {
	Packarten(ctx context.Context, setID int64) (Packangebot, error)
	PackZiehen(ctx context.Context, setID int64, art, same string) (Packziehung, error)
}

// Packangebot ist die Antwort auf »was lässt sich hier ziehen?«.
type Packangebot struct {
	Arten []Packart `json:"arten"`

	// Befunde nennt die Packarten, deren Filter sich nicht ausrechnen ließ.
	//
	// **Sie stehen in der Antwort und nicht im Log** (§8): eine Packart, die
	// still leer bleibt, sieht aus wie ein Set ohne passende Karten.
	Befunde []string `json:"befunde"`
}

// Packart ist eine Art samt der Zahl erreichbarer Karten.
type Packart struct {
	// Name steht **unübersetzt** da — so heißt er im Paket (§9).
	Name     string `json:"name"`
	Waehlbar bool   `json:"waehlbar"`
	Karten   int    `json:"karten"`
}

// Packziehung sind die gezogenen Karten in ihrer Ziehreihenfolge.
type Packziehung struct {
	Karten  []Packkarte `json:"karten"`
	Befunde []string    `json:"befunde"`
}

// Packkarte ist eine gezogene Karte.
type Packkarte struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// packarten gibt die Packarten eines Sets.
func (d *Dienst) packarten(schreiben http.ResponseWriter, anfrage *http.Request) {
	setID, gut := kennung(schreiben, anfrage, "setID")
	if !gut {
		return
	}

	angebot, fehler := d.naehte.Pack.Packarten(anfrage.Context(), setID)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, angebot)
}

// packziehen zieht ein Pack.
//
// **Der Same steht im Weg, nicht im Rumpf**, und er ist Pflicht: dieselbe Zahl
// gibt dieselbe Ziehung, und ohne sie ließe sich kein Ergebnis wiederholen,
// über das jemand im Fenster spricht.
//
// **Als Text**: eine Zahl über 2⁵³ verlöre in JSON ihre letzten Stellen.
func (d *Dienst) packziehen(schreiben http.ResponseWriter, anfrage *http.Request) {
	setID, gut := kennung(schreiben, anfrage, "setID")
	if !gut {
		return
	}

	ziehung, fehler := d.naehte.Pack.PackZiehen(anfrage.Context(), setID,
		anfrage.PathValue("art"), anfrage.URL.Query().Get("same"))
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, ziehung)
}
