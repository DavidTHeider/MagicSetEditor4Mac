// **Die Schlüsselwörter eines Sets über die Rückschleife** —
// #1062.

package rest

import (
	"context"
	"net/http"
)

// Schluesselwoerter ist, was der Dienst dafür braucht.
type Schluesselwoerter interface {
	Schluesselwoerter(ctx context.Context, setID int64) (Schluesselwortliste, error)
	SchluesselwortAnlegen(ctx context.Context, setID int64) (int64, error)
	SchluesselwortSetzen(ctx context.Context, setID, wortID int64, feld, wert string) error
	SchluesselwortLoeschen(ctx context.Context, setID, wortID int64) error
}

// Schluesselwortliste ist die Antwort auf »welche gibt es hier?«.
type Schluesselwortliste struct {
	Woerter []Schluesselwort `json:"woerter"`
	Modi    []string         `json:"modi"`
}

// Schluesselwort ist eines des Sets oder seines Spiels.
type Schluesselwort struct {
	// ID ist **0 bei einem des Spiels**: das gehört keinem Set.
	ID int64 `json:"id"`

	Wort       string `json:"wort"`
	Muster     string `json:"muster"`
	Erinnerung string `json:"erinnerung"`
	Regeln     string `json:"regeln"`
	Modus      string `json:"modus"`

	Fest    bool `json:"fest"`
	Gueltig bool `json:"gueltig"`
}

// schluesselwoerter gibt die des Sets und die des Spiels.
func (d *Dienst) schluesselwoerter(schreiben http.ResponseWriter, anfrage *http.Request) {
	setID, gut := kennung(schreiben, anfrage, "setID")
	if !gut {
		return
	}

	liste, fehler := d.naehte.Woerter.Schluesselwoerter(anfrage.Context(), setID)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	// **Leere Listen und kein `null`**: ein Aufrufer, der darüber läuft,
	// prüfte sonst auf zwei Dinge statt auf eines.
	if liste.Woerter == nil {
		liste.Woerter = []Schluesselwort{}
	}

	if liste.Modi == nil {
		liste.Modi = []string{}
	}

	ja(schreiben, http.StatusOK, liste)
}

// schluesselwortAnlegen hängt ein leeres ans Ende.
func (d *Dienst) schluesselwortAnlegen(schreiben http.ResponseWriter, anfrage *http.Request) {
	setID, gut := kennung(schreiben, anfrage, "setID")
	if !gut {
		return
	}

	neu, fehler := d.naehte.Woerter.SchluesselwortAnlegen(anfrage.Context(), setID)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusCreated, map[string]any{"wortID": neu})
}

// schluesselwortSetzen schreibt ein Feld.
//
// **`PUT` und nicht `POST`**: gesetzt wird **ein** Feld auf **einen** Wert.
// Zweimal dasselbe zu schicken gibt dasselbe Ergebnis, und genau das sagt
// `PUT` — dieselbe Wahl wie beim Kartenfeld.
func (d *Dienst) schluesselwortSetzen(schreiben http.ResponseWriter, anfrage *http.Request) {
	setID, wortID, gut := beideSchluesselwortkennungen(schreiben, anfrage)
	if !gut {
		return
	}

	var rumpf struct {
		Wert string `json:"wert"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	feld := anfrage.PathValue("feld")

	if fehler := d.naehte.Woerter.SchluesselwortSetzen(
		anfrage.Context(), setID, wortID, feld, rumpf.Wert); fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"feld": feld})
}

// schluesselwortLoeschen nimmt eines weg.
func (d *Dienst) schluesselwortLoeschen(schreiben http.ResponseWriter, anfrage *http.Request) {
	setID, wortID, gut := beideSchluesselwortkennungen(schreiben, anfrage)
	if !gut {
		return
	}

	if fehler := d.naehte.Woerter.SchluesselwortLoeschen(
		anfrage.Context(), setID, wortID); fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"wortID": wortID})
}

// beideSchluesselwortkennungen liest Set und Schlüsselwort aus dem Weg.
func beideSchluesselwortkennungen(
	schreiben http.ResponseWriter, anfrage *http.Request,
) (setID, wortID int64, gut bool) {
	setID, gut = kennung(schreiben, anfrage, "setID")
	if !gut {
		return 0, 0, false
	}

	wortID, gut = kennung(schreiben, anfrage, "wortID")
	if !gut {
		return 0, 0, false
	}

	return setID, wortID, true
}
