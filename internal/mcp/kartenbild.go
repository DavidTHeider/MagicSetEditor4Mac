// **Ein Kartenbild von der Platte einsetzen** —
// #1291.
//
// **Eine eigene Datei neben `scryfallbild.go`**, weil sie eine andere Frage
// beantwortet: dort kommt das Bild aus dem Netz, hier von der Platte dieses
// Rechners — und der Unterschied ist nicht nur die Quelle, sondern auch, wer
// die Rechte daran hat (§11).

package mcp

import (
	"context"
	"encoding/json"
)

// Kartenbilder setzt ein Bild aus einer Datei in ein Kartenfeld.
type Kartenbilder interface {
	// AusDatei liest eine Bilddatei, schneidet sie zu und schreibt das Feld.
	AusDatei(ctx context.Context, setID, karteID int64, feld, pfad string) (Bildmasse, error)
}

// Bildmasse sagt, was gelesen und worauf zugeschnitten wurde.
type Bildmasse struct {
	Feld         string `json:"feld"`
	QuelleBreite int    `json:"quelleBreite"`
	QuelleHoehe  int    `json:"quelleHoehe"`
	ZielBreite   int    `json:"zielBreite"`
	ZielHoehe    int    `json:"zielHoehe"`
}

// kartenbildwerkzeug gibt das Werkzeug, das ein Bild einsetzt.
func kartenbildwerkzeug(bilder Kartenbilder) []Werkzeug {
	if bilder == nil {
		return nil
	}

	return []Werkzeug{{
		Name: "schreiben.karte.bild", Art: Schreibend,
		Erklaerung: "ÄNDERT eine Karte: liest eine Bilddatei von DIESEM Rechner, " +
			"schneidet sie auf die Fläche des Feldes zu und trägt sie ein. Der " +
			"Zuschnitt ist der, den Magic Set Editor selbst vorschlägt. Ein " +
			"vorhandenes Bild wird ersetzt — der Verlauf hält den vorigen Wert.",
		Schema: objekt(map[string]any{
			"setID":   zahl("Die Kennung des Sets aus »lesen.sets«."),
			"karteID": zahl("Die Kennung der Karte aus »lesen.karten«."),
			"feld":    text("Das Bildfeld der Karte, etwa »image« oder »kunst«."),
			"pfad":    text("Der Pfad zur Bilddatei (.png, .jpg) auf diesem Rechner."),
		}, "setID", "karteID", "feld", "pfad"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var argumente struct {
				SetID   int64  `json:"setID"`
				KarteID int64  `json:"karteID"`
				Feld    string `json:"feld"`
				Pfad    string `json:"pfad"`
			}

			if err := lies(roh, &argumente); err != nil {
				return "", err
			}

			if err := pruefeKennung("setID", argumente.SetID); err != nil {
				return "", err
			}

			if err := pruefeKennung("karteID", argumente.KarteID); err != nil {
				return "", err
			}

			masse, err := bilder.AusDatei(
				ctx, argumente.SetID, argumente.KarteID, argumente.Feld, argumente.Pfad)
			if err != nil {
				return "", err
			}

			return alsJSON(masse)
		},
	}}
}
