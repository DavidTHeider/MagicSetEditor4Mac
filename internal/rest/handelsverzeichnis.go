package rest

// **Die Liste der Handelsvorgänge und ihre Entwürfe** — TRADE-01
// (#1715).
//
// `GET /handel` ist die Liste mit dem Filter in der Abfrage; ein Entwurf
// ist eine Ressource des Vorgangs (`PUT` schreibt ihn neu, `DELETE`
// löscht ihn), die Kopie entsteht mit `POST`, und das Archiv ist ein
// Zustand, den `PUT` setzt und `DELETE` nimmt.

import (
	"net/http"
	"strconv"
)

// Handelsfilter ist der Filter aus der Abfrage.
type Handelsfilter struct {
	Partner string   `json:"partner"`
	Lagen   []string `json:"lagen"`
	Suche   string   `json:"suche"`
	Marke   string   `json:"marke"`
	Seit    string   `json:"seit"`
	Bis     string   `json:"bis"`
	Ordnung string   `json:"ordnung"`
	Archiv  bool     `json:"archiv"`
}

// Handelspartner ist ein Gegenüber einer Zeile.
type Handelspartner struct {
	Kennung string `json:"kennung"`
	Name    string `json:"name"`
}

// Handelsuebersichtszeile ist ein Vorgang in der Liste.
type Handelsuebersichtszeile struct {
	Vorgang      Handelsvorgang   `json:"vorgang"`
	Lage         string           `json:"lage"`
	Naechster    string           `json:"naechster"`
	Gegenueber   []Handelspartner `json:"gegenueber"`
	Karten       []string         `json:"karten"`
	Marken       []string         `json:"marken"`
	Angelegt     string           `json:"angelegt"`
	Geaendert    string           `json:"geaendert"`
	Archiviert   bool             `json:"archiviert"`
	Unbestaetigt bool             `json:"unbestaetigt"`
}

// handelsuebersichtZeigen gibt die Liste — `?lage=` darf mehrfach stehen.
func (d *Dienst) handelsuebersichtZeigen(schreiben http.ResponseWriter, anfrage *http.Request) {
	abfrage := anfrage.URL.Query()
	filter := Handelsfilter{
		Partner: abfrage.Get("partner"), Lagen: abfrage["lage"], Suche: abfrage.Get("suche"),
		Marke: abfrage.Get("marke"), Seit: abfrage.Get("seit"), Bis: abfrage.Get("bis"),
		Ordnung: abfrage.Get("ordnung"),
	}

	if roh := abfrage.Get("archiv"); roh != "" {
		archiv, fehler := strconv.ParseBool(roh)
		if fehler != nil {
			nein(schreiben, http.StatusBadRequest, "validierung", "archiv ist true oder false")

			return
		}

		filter.Archiv = archiv
	}

	zeilen, fehler := d.naehte.Leitung.HandelUebersicht(anfrage.Context(), filter)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"zeilen": ohneNil(zeilen)})
}

// handelEntwurfAendern schreibt einen Entwurf neu — der Vorgang steht im
// Pfad und nicht im Rumpf.
func (d *Dienst) handelEntwurfAendern(schreiben http.ResponseWriter, anfrage *http.Request) {
	var wunsch Handelswunsch

	if !gelesen(schreiben, anfrage, &wunsch) {
		return
	}

	wunsch.Vorgang = anfrage.PathValue("vorgang")

	vorgang, fehler := d.naehte.Leitung.HandelEntwurfAendern(anfrage.Context(), wunsch)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, vorgang)
}

// handelEntwurfLoeschen löscht einen Entwurf.
func (d *Dienst) handelEntwurfLoeschen(schreiben http.ResponseWriter, anfrage *http.Request) {
	if fehler := d.naehte.Leitung.HandelEntwurfLoeschen(anfrage.Context(),
		anfrage.PathValue("vorgang")); fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	schreiben.WriteHeader(http.StatusNoContent)
}

// handelKopieren legt aus einem Vorgang einen neuen Entwurf an.
func (d *Dienst) handelKopieren(schreiben http.ResponseWriter, anfrage *http.Request) {
	kopie, fehler := d.naehte.Leitung.HandelKopieren(anfrage.Context(), anfrage.PathValue("vorgang"))
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusCreated, kopie)
}

// handelArchivieren legt einen beendeten Vorgang ins Archiv (`PUT`) oder
// holt ihn heraus (`DELETE`).
func (d *Dienst) handelArchivieren(an bool) http.HandlerFunc {
	return func(schreiben http.ResponseWriter, anfrage *http.Request) {
		if fehler := d.naehte.Leitung.HandelArchivieren(anfrage.Context(),
			anfrage.PathValue("vorgang"), an); fehler != nil {
			fehlerhaft(schreiben, fehler)

			return
		}

		schreiben.WriteHeader(http.StatusNoContent)
	}
}
