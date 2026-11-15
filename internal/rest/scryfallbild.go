package rest

import (
	"context"
	"net/http"
)

// Fremdbilder ist, was der Bildendpunkt braucht.
type Fremdbilder interface {
	Einsetzen(ctx context.Context, setID, karteID int64, feld, name, art string) (
		Bildeinsatz, error)
}

// Bildeinsatz ist, was ein Einsetzen zurückmeldet.
type Bildeinsatz struct {
	Feld               string
	Kuenstler          string
	Nennung            string
	Adresse            string
	IllustratorGesetzt bool
	AlterStunden       int

	// Befunde ist, was die Kartenauskunft tolerant hingenommen hat —
	// der Grund steht bei [service.Bilduebernahme].
	Befunde []string
}

// scryfallBild holt das Kartenbild und setzt es ein.
//
// # Warum die Nennung in der Antwort steht
//
// Weil dieser Dienst eine **Fläche nach außen** ist (0033): was ihn ruft, baut
// eine eigene Oberfläche, und Scryfall verlangt Künstler und Copyright neben
// dem Bild. Ein Aufrufer, der die Nennung erst suchen müsste, ließe sie weg.
func (d *Dienst) scryfallBild(schreiben http.ResponseWriter, anfrage *http.Request) {
	setID, gutesSet := kennung(schreiben, anfrage, "setID")
	karteID, guteKarte := kennung(schreiben, anfrage, "karteID")

	if !gutesSet || !guteKarte {
		return
	}

	var rumpf struct {
		Feld string `json:"feld"`
		Name string `json:"name"`
		Art  string `json:"art"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	// **Ohne Angabe die ganze Karte** — dieselbe Vorgabe wie an der
	// Kommandozeile. Sie trägt Künstler und Copyright im Bild selbst.
	if rumpf.Art == "" {
		rumpf.Art = "normal"
	}

	eingesetzt, fehler := d.naehte.Bilder.Einsetzen(anfrage.Context(),
		setID, karteID, rumpf.Feld, rumpf.Name, rumpf.Art)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{
		"feld":               eingesetzt.Feld,
		"kuenstler":          eingesetzt.Kuenstler,
		"nennung":            eingesetzt.Nennung,
		"adresse":            eingesetzt.Adresse,
		"illustratorGesetzt": eingesetzt.IllustratorGesetzt,
		"alterStunden":       eingesetzt.AlterStunden,

		// **Leer und nicht nil**: `null` und `[]` sind für einen Aufrufer
		// nicht dasselbe, und die Antwort soll nicht danach aussehen, als
		// hätte das Feld gefehlt.
		"befunde": ohneNil(eingesetzt.Befunde),
	})
}
