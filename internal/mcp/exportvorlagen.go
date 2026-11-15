// **Die Exportvorlagen über MCP** —
// #1291.
//
// # Warum ein Modell sie braucht
//
// Weil es die Fähigkeit ist, für die MSE2 seine Vorlagen mitbringt: das
// `M15-Magic-Pack` führt zwölf `.mse-export-template` — Cockatrice, MWS,
// LackeyCCG, Planesculptors. Ein Modell, das ein Set „für Cockatrice
// ausgeben" soll, hatte bis dahin keinen Weg dorthin.

package mcp

import (
	"context"
	"encoding/json"
)

// Vorlagenausgabe ist, was die drei Werkzeuge dieser Datei brauchen.
type Vorlagenausgabe interface {
	// Vorlagen nennt die Exportvorlagen, die zu einem Set passen.
	Vorlagen(ctx context.Context, setID int64) ([]Vorlagenzeile, []string, error)

	// Ausfuehren schreibt das Ergebnis an einen Ort, an dem nichts liegt.
	Ausfuehren(ctx context.Context, setID int64, paket, ziel string,
		wahl map[string]string) (befunde []string, fehler error)

	// OptionSetzen behält den Wert einer Vorlagenoption.
	OptionSetzen(ctx context.Context, vorlage, feld, wert string) error
}

// Vorlagenzeile ist eine Exportvorlage.
type Vorlagenzeile struct {
	Paket    string           `json:"paket"`
	Name     string           `json:"name"`
	Dateityp string           `json:"dateityp"`
	Optionen []Vorlagenoption `json:"optionen"`
}

// Vorlagenoption ist ein einstellbares Feld einer Exportvorlage.
type Vorlagenoption struct {
	Name string `json:"name"`
	Art  string `json:"art"`
	Wert string `json:"wert"`
}

// exportvorlagen gibt die drei Werkzeuge rund um Exportvorlagen.
func exportvorlagen(ausgabe Vorlagenausgabe) []Werkzeug {
	if ausgabe == nil {
		return nil
	}

	return []Werkzeug{
		{
			Name: "lesen.vorlagen", Art: Lesend, Idempotent: true,
			Erklaerung: "Nennt die Exportvorlagen, die zum Spiel eines Sets gehören — " +
				"samt ihren einstellbaren Feldern und deren Werten. »dateityp« sagt, " +
				"welche Endung das Ergebnis tragen soll.",
			Schema: objekt(map[string]any{
				"setID": zahl("Die Kennung des Sets aus »lesen.sets«."),
			}, "setID"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				wahl, err := kennung(roh)
				if err != nil {
					return "", err
				}

				vorlagen, befunde, lesefehler := ausgabe.Vorlagen(ctx, wahl)
				if lesefehler != nil {
					return "", lesefehler
				}

				return alsJSON(map[string]any{
					"vorlagen": liste(vorlagen), "befunde": liste(befunde),
				})
			},
		},
		{
			Name: "schreiben.vorlage.ausfuehren", Art: Schreibend,
			Erklaerung: "ÄNDERT die Platte: führt eine Exportvorlage über das ganze Set " +
				"aus und schreibt das Ergebnis an einen Ort, an dem noch nichts liegt. " +
				"Ein vorhandenes Ziel wird abgewiesen. »wahl« setzt einstellbare Felder " +
				"für DIESEN Lauf, ohne sie zu behalten.",
			Schema: objekt(map[string]any{
				"setID": zahl("Die Kennung des Sets aus »lesen.sets«."),
				"paket": text("Der Paketname der Vorlage aus »lesen.vorlagen«."),
				"ziel":  text("Der Pfad, an den geschrieben wird — dort darf nichts liegen."),
				// **`wahl` ist eine freie Abbildung** — welche Felder es
				// gibt, sagt `lesen.vorlagen` je Vorlage; ein festes Schema
				// hier wäre eine zweite, veraltende Liste.
				"wahl": map[string]any{
					"type":        "object",
					"description": "Optionswerte für DIESEN Lauf: {feld: wert}.",
				},
			}, "setID", "paket", "ziel"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				var argumente struct {
					SetID int64             `json:"setID"`
					Paket string            `json:"paket"`
					Ziel  string            `json:"ziel"`
					Wahl  map[string]string `json:"wahl"`
				}

				if err := lies(roh, &argumente); err != nil {
					return "", err
				}

				if err := pruefeKennung("setID", argumente.SetID); err != nil {
					return "", err
				}

				befunde, err := ausgabe.Ausfuehren(
					ctx, argumente.SetID, argumente.Paket, argumente.Ziel, argumente.Wahl)
				if err != nil {
					return "", err
				}

				return alsJSON(map[string]any{
					"pfad": argumente.Ziel, "befunde": liste(befunde),
				})
			},
		},
		{
			Name: "schreiben.vorlagenoption", Art: Schreibend,
			Erklaerung: "ÄNDERT den Bestand: behält den Wert eines einstellbaren Feldes " +
				"einer Exportvorlage. Er gilt ab dem nächsten Lauf — für einen " +
				"einzelnen genügt »wahl« bei »schreiben.vorlage.ausfuehren«.",
			Schema: objekt(map[string]any{
				"vorlage": text("Der Paketname der Vorlage aus »lesen.vorlagen«."),
				"feld":    text("Der Name des Feldes aus »lesen.vorlagen«."),
				"wert":    text("Der Wert, der behalten wird."),
			}, "vorlage", "feld", "wert"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				var argumente struct {
					Vorlage string `json:"vorlage"`
					Feld    string `json:"feld"`
					Wert    string `json:"wert"`
				}

				if err := lies(roh, &argumente); err != nil {
					return "", err
				}

				if err := ausgabe.OptionSetzen(
					ctx, argumente.Vorlage, argumente.Feld, argumente.Wert); err != nil {
					return "", err
				}

				return alsJSON(map[string]any{
					"vorlage": argumente.Vorlage, "feld": argumente.Feld,
					"wert": argumente.Wert,
				})
			},
		},
	}
}
