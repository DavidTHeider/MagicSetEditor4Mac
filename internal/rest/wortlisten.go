// **Die Auswahllisten eines Kartenfeldes über die Rückschleife** —
// #1074.

package rest

import (
	"context"
	"net/http"
)

// Wortlisten ist, was der Dienst für die Auswahl braucht.
type Wortlisten interface {
	Wortlisten(ctx context.Context, karteID int64, feld string) ([]Wortbereich, error)

	Wortwahl(
		ctx context.Context, setID, karteID int64, feld string, bereich int, weg []int,
	) error
}

// Wortbereich ist eine Stelle eines Feldes mit ihrem Menü.
type Wortbereich struct {
	Liste     string        `json:"liste"`
	Text      string        `json:"text"`
	Kaestchen bool          `json:"kaestchen"`
	Gewaehlt  int           `json:"gewaehlt"`
	Eintraege []Worteintrag `json:"eintraege"`
}

// Worteintrag ist eine Zeile im Menü.
type Worteintrag struct {
	Name          string        `json:"name"`
	IstPraefix    bool          `json:"istPraefix"`
	ZeileDarunter bool          `json:"zeileDarunter"`
	Aktiv         bool          `json:"aktiv"`
	Kinder        []Worteintrag `json:"kinder"`
}

// wortlisten gibt die Auswahl eines Feldes.
func (d *Dienst) wortlisten(schreiben http.ResponseWriter, anfrage *http.Request) {
	karteID, gut := kennung(schreiben, anfrage, "karteID")
	if !gut {
		return
	}

	bereiche, fehler := d.naehte.Wortwahl.Wortlisten(
		anfrage.Context(), karteID, anfrage.PathValue("feld"))
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	// **Eine leere Liste und kein `null`**: ein Aufrufer, der über die Antwort
	// läuft, prüfte sonst auf zwei Dinge statt auf eines.
	if bereiche == nil {
		bereiche = []Wortbereich{}
	}

	ja(schreiben, http.StatusOK, map[string]any{"bereiche": bereiche})
}

// wortwaehlen wählt einen Eintrag und schreibt das Feld.
//
// **`POST` und nicht `PUT`**: gesetzt wird kein Wert, sondern eine Wahl
// angewandt — und ein Präfix wird dabei **umgeschaltet**. Zweimal dasselbe zu
// schicken gibt nicht dasselbe Ergebnis, und genau das sagt `POST`.
func (d *Dienst) wortwaehlen(schreiben http.ResponseWriter, anfrage *http.Request) {
	setID, gutesSet := kennung(schreiben, anfrage, "setID")
	if !gutesSet {
		return
	}

	karteID, guteKarte := kennung(schreiben, anfrage, "karteID")
	if !guteKarte {
		return
	}

	var rumpf struct {
		Bereich int   `json:"bereich"`
		Weg     []int `json:"weg"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	feld := anfrage.PathValue("feld")

	if fehler := d.naehte.Wortwahl.Wortwahl(anfrage.Context(),
		setID, karteID, feld, rumpf.Bereich, rumpf.Weg); fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"feld": feld})
}
