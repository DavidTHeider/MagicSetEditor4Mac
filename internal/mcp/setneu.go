// **Ein neues, leeres Set** —
// #891.
//
// # Wofür ein Sprachmodell es braucht
//
// Um **bei null anzufangen**. Bis dahin kam ein Set nur aus einer Datei — ein
// Modell, das gebeten wird „bau mir ein Set mit fünf Karten", hatte keinen
// ersten Schritt.
//
// **Und es sagt keine Idempotenz zu.** Zwei Aufrufe mit demselben Namen legen
// **zwei** Sets an, und das ist Absicht: ein Set hat keinen Schlüssel außer
// seiner Kennung, und zwei Sets desselben Namens sind erlaubt. Wer prüfen
// will, ob es schon eines gibt, ruft vorher `lesen.sets`.

package mcp

import (
	"context"
	"encoding/json"
)

// Setanleger ist, was dieses Werkzeug braucht.
type Setanleger interface {
	SetAnlegen(ctx context.Context, name, spiel, stil string) (int64, error)
}

// Setanlage ist die Antwort — die Kennung des neuen Sets.
//
// **Als Zahl und nicht als Text**, wie bei den Kartenwerkzeugen daneben: die
// übrigen Werkzeuge nehmen Kennungen als Zahl entgegen, und eine Antwort, die
// sie als Text gäbe, verlangte vom Modell eine Umwandlung, die es nirgends
// sonst braucht.
type Setanlage struct {
	ID int64 `json:"ID"`
}

// setneuwerkzeug gibt das schreibende Werkzeug.
func setneuwerkzeug(anleger Setanleger) []Werkzeug {
	return []Werkzeug{
		{
			Name: "schreiben.set.neu", Art: Schreibend,
			Erklaerung: "ÄNDERT den Bestand: legt ein LEERES Set an und gibt seine " +
				"Kennung. name ist der Setname, spiel das MSE-Spiel (»magic«, " +
				"»probe«), stil der Stilbogen — VERLANGT, und einer des Spiels: " +
				"er beginnt mit »<spiel>-«. Ohne ihn wird abgewiesen. " +
				"Ein zweiter Aufruf mit demselben Namen legt ein ZWEITES Set an; " +
				"wer das nicht will, ruft vorher »lesen.sets«. " +
				"Karten kommen danach über »schreiben.karte.anlegen«.",
			Schema: objekt(map[string]any{
				"name":  text("Der Name des Sets."),
				"spiel": text("Das MSE-Spiel, etwa »magic«."),
				"stil": text("Der Stilbogen, etwa »magic-m15«. VERLANGT, und " +
					"einer des Spiels: er beginnt mit »<spiel>-«."),
			}, "name", "spiel", "stil"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				var argumente struct {
					Name  string `json:"name"`
					Spiel string `json:"spiel"`
					Stil  string `json:"stil"`
				}

				if err := lies(roh, &argumente); err != nil {
					return "", err
				}

				kennung, fehler := anleger.SetAnlegen(
					ctx, argumente.Name, argumente.Spiel, argumente.Stil)
				if fehler != nil {
					return "", fehler
				}

				return alsJSON(Setanlage{ID: kennung})
			},
		},
	}
}
