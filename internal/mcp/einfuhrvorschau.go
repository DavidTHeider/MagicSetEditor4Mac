// **Eine Einfuhr vorab ansehen, über MCP** — DOM-03.AK-01
// (#1698).
//
// Einordnung nach §16: **erweiternd** — zwei lesende Werkzeuge kommen hinzu,
// und die zwei schreibenden nehmen ein freiwilliges Feld
// »zusammenfuehrung«. Wer es nicht schickt, bekommt, was er bisher bekam.

package mcp

import (
	"context"
	"encoding/json"
)

// Vorschauposten ist eine Zeile der Vorschau.
type Vorschauposten struct {
	Zeile       int    `json:"zeile"`
	Bezeichnung string `json:"bezeichnung"`
	Menge       int    `json:"menge"`
	Zustand     string `json:"zustand"`
	Sprache     string `json:"sprache"`
	Wunsch      bool   `json:"wunsch"`
	Art         string `json:"art"`
	Vorher      int    `json:"vorher"`
	Nachher     int    `json:"nachher"`
}

// Sammelvorschau sagt, was eine Einfuhr täte.
type Sammelvorschau struct {
	Posten  []Vorschauposten `json:"posten"`
	Befunde []string         `json:"befunde"`
}

// zusammenfuehrungsfeld beschreibt das freiwillige Feld der zwei
// schreibenden Werkzeuge.
func zusammenfuehrungsfeld() map[string]any {
	return text("Was mit einer Zeile geschieht, die es schon gibt: " +
		"»addieren« (die Vorgabe) erhöht sie, »auslassen« lässt sie, wie sie " +
		"ist. Eine Zeile, die dieselbe Einfuhr eben angelegt hat, wächst in " +
		"beiden Fällen. Was geschähe, sagt vorher die Vorschau.")
}

// lesenListenvorschau sagt, was »schreiben.liste.einlesen« täte.
func lesenListenvorschau(einfuhr Listeneinfuhr) Werkzeug {
	return Werkzeug{
		Name: "lesen.liste.vorschau", Art: Lesend,
		Erklaerung: "Schreibt NICHTS: sagt je Zeile, was »schreiben.liste." +
			"einlesen« mit denselben Zeilen täte — »neu«, »zusammen« (trifft " +
			"eine Zeile, die die Sammlung schon führt, mit vorher und nachher) " +
			"oder »wiederholt« (trifft eine frühere Zeile derselben Einfuhr).",
		Schema: objekt(map[string]any{
			"zeilen": listenzeilenschema(),
		}, "zeilen"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var argumente struct {
				Zeilen []Listeneinfuhrzeile `json:"zeilen"`
			}

			if err := lies(roh, &argumente); err != nil {
				return "", err
			}

			vorschau, err := einfuhr.Vorschauen(ctx, argumente.Zeilen)
			if err != nil {
				return "", err
			}

			return alsJSON(vorschau)
		},
	}
}

// lesenSammlungsvorschau sagt, was »schreiben.sammlung.einlesen« täte.
func lesenSammlungsvorschau(ausfuhr Sammlungsausfuhr) Werkzeug {
	return Werkzeug{
		Name: "lesen.sammlung.vorschau", Art: Lesend,
		Erklaerung: "Schreibt NICHTS, auch keinen Druck ins Register: sagt je " +
			"Zeile eines CSV, was »schreiben.sammlung.einlesen« täte — »neu«, " +
			"»zusammen« (mit vorher und nachher) oder »wiederholt«. Die eigene " +
			"Ausfuhr ein zweites Mal eingelesen träfe jede Zeile.",
		Schema: objekt(map[string]any{
			"text": text("Das CSV, mit einer Kopfzeile — wie beim Einlesen."),
		}, "text"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var argumente struct {
				Text string `json:"text"`
			}

			if err := lies(roh, &argumente); err != nil {
				return "", err
			}

			vorschau, err := ausfuhr.CSVVorschauen(ctx, argumente.Text)
			if err != nil {
				return "", err
			}

			return alsJSON(vorschau)
		},
	}
}
