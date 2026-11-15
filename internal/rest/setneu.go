// **Ein neues, leeres Set über die Rückschleife** —
// #891.

package rest

import (
	"context"
	"net/http"
	"strconv"
)

// Setanleger ist, was dieser Endpunkt braucht.
type Setanleger interface {
	SetAnlegen(ctx context.Context, name, spiel, stil string) (int64, error)
}

// setNeu legt ein leeres Set an.
//
// **`POST /sets/neu` und nicht `POST /sets`.** Jener Weg fährt eine **Datei**
// ein; beide auf denselben Weg zu legen und am Rumpf zu unterscheiden hieße,
// zwei Fähigkeiten hinter einer Adresse zu verstecken — und eine Adresse ist
// eine Zusage ([0033](../../docs/entscheidungen/0033-was-ueber-fassungen-hinweg-zugesagt-wird.md)).
func (d *Dienst) setNeu(schreiben http.ResponseWriter, anfrage *http.Request) {
	var rumpf struct {
		Name  string `json:"name"`
		Spiel string `json:"spiel"`

		// Stil ist **verlangt**, und zwar einer des Spiels.
		//
		// **Nicht wahlfrei „wie in MSE2s Dialog"**: `Set::validate` sagt das
		// Gegenteil (`src/data/set.cpp`, Zeilen 143–149, Fassung f9d9356 vom
		// 2021-07-06): ein Set ohne Stil besteht dort die Prüfung nicht, gleich
		// woher es kommt. Die Bedingung steht in `service.Importeur.Neu`, damit
		// sie an **jeder** Fläche gilt.
		Stil string `json:"stil"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	kennung, fehler := d.naehte.Anleger.SetAnlegen(
		anfrage.Context(), rumpf.Name, rumpf.Spiel, rumpf.Stil)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusCreated, map[string]any{
		"id": strconv.FormatInt(kennung, 10),
	})
}
