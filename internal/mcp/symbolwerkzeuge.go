// **Die Werkzeuge des Symboleditors** — abgetrennt von `symbol.go`, weil dort
// die **Schemata** stehen und hier die **Rufe**
// (#1391).
//
// Der Schnitt folgt der Frage: dort steht, was ein Aufrufer schicken darf,
// hier, was damit geschieht.

package mcp

import (
	"context"
	"encoding/json"
)

// symbolwerkzeuge gibt die Werkzeuge des Editors — acht lesend, eines
// schreibend, dazu das Lesen des abgelegten Symbols.
func symbolwerkzeuge(werkstatt Symbolwerkstatt) []Werkzeug {
	alle := []Werkzeug{
		{
			Name: "lesen.symbol", Art: Lesend, Idempotent: true,
			Erklaerung: "Gibt das abgelegte Symbol eines Sets als Teile. ÄNDERT NICHTS. " +
				"Das Ergebnis geht in »teile« der Rechnungen; zurückgeschrieben wird " +
				"mit »schreiben.symbol«.",
			Schema: objekt(map[string]any{
				"setID": zahl("Die Kennung des Sets aus »lesen.sets«."),
				"name":  text("Der Dateiname im Set — leer für das Setsymbol."),
			}, "setID"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				var wahl struct {
					SetID int64  `json:"setID"`
					Name  string `json:"name"`
				}

				if err := lies(roh, &wahl); err != nil {
					return "", err
				}

				if err := pruefeKennung("setID", wahl.SetID); err != nil {
					return "", err
				}

				teile, fehler := werkstatt.SymbolLesen(ctx, wahl.SetID, wahl.Name)
				if fehler != nil {
					return "", fehler
				}

				return alsJSON(map[string]any{"Teile": teile})
			},
		},
		{
			Name: "schreiben.symbol", Art: Schreibend,
			Erklaerung: "ÄNDERT den Bestand: legt eine gezeichnete Symbolzeichnung neben " +
				"das Set und gibt ihren Dateinamen. Ein vorhandenes Symbol desselben " +
				"Namens wird ERSETZT. Zurückzunehmen ist das, indem die vorigen Teile " +
				"— aus »lesen.symbol«, VOR dem Schreiben geholt — wieder abgelegt werden.",
			Schema: objekt(map[string]any{
				"setID": zahl("Die Kennung des Sets aus »lesen.sets«."),
				"teile": rohliste("Die Teile des Symbols — aus einer Rechnung."),
				"unter": text("Der Dateiname im Set. Leer: das Setsymbol."),
			}, "setID", "teile"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				var wahl struct {
					SetID int64           `json:"setID"`
					Teile json.RawMessage `json:"teile"`
					Unter string          `json:"unter"`
				}

				if err := lies(roh, &wahl); err != nil {
					return "", err
				}

				if err := pruefeKennung("setID", wahl.SetID); err != nil {
					return "", err
				}

				name, fehler := werkstatt.SymbolAblegen(ctx, wahl.SetID, wahl.Teile, wahl.Unter)
				if fehler != nil {
					return "", fehler
				}

				return alsJSON(map[string]any{"Name": name})
			},
		},
	}

	alle = append(alle, Werkzeug{
		Name: "lesen.symbolpalette", Art: Lesend, Idempotent: true,
		Erklaerung: "Zählt auf, was sich in ein Kartenfeld einfügen lässt — Code, " +
			"Beschriftung und Art. ÄNDERT NICHTS. Für Manakosten und ähnliche " +
			"Felder: hier stehen die Codes, die die Vorlage kennt. Die Bildchen, " +
			"die das Fenster dazu zeigt, sind NICHT dabei — sie sind die " +
			"Darstellung eines Codes, nicht der Code.",
		Schema: objekt(map[string]any{
			"karteID": zahl("Die Kennung der Karte aus »lesen.karten«."),
			"feld":    text("Der Feldname, etwa »casting cost«."),
		}, "karteID", "feld"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var wahl struct {
				KarteID int64  `json:"karteID"`
				Feld    string `json:"feld"`
			}

			if err := lies(roh, &wahl); err != nil {
				return "", err
			}

			if err := pruefeKennung("karteID", wahl.KarteID); err != nil {
				return "", err
			}

			eintraege, fehler := werkstatt.Palette(ctx, wahl.KarteID, wahl.Feld)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(map[string]any{"Eintraege": eintraege})
		},
	})

	for _, rechnung := range symbolrechnungen() {
		alle = append(alle, alsWerkzeug(werkstatt, rechnung))
	}

	return alle
}

// alsWerkzeug macht aus einer Rechnung ein Werkzeug.
//
// **Die Schleife ist Absicht.** Alle neun rechnen dasselbe Muster: Argumente
// hinein, ein Ergebnis heraus, kein Bestand dazwischen. Sie einzeln
// auszuschreiben wäre neunmal derselbe Rumpf — und beim zehnten schriebe
// jemand ihn anders.
func alsWerkzeug(werkstatt Symbolwerkstatt, rechnung symbolrechnung) Werkzeug {
	return Werkzeug{
		Name: rechnung.name, Art: Lesend, Idempotent: true,
		Erklaerung: rechnung.erklaerung,
		Schema:     objekt(rechnung.felder, rechnung.pflicht...),
		Ruf: func(_ context.Context, roh json.RawMessage) (string, error) {
			ergebnis, fehler := werkstatt.SymbolRechnen(rechnung.name, roh)
			if fehler != nil {
				return "", fehler
			}

			// **Unter einem Namen und nicht roh.** Jede Antwort dieses
			// Servers ist ein Objekt; eine nackte Liste wäre die einzige,
			// die aus der Reihe fällt — und ein Modell, das `.Ergebnis`
			// liest, fände sie nicht.
			return alsJSON(map[string]any{"Ergebnis": ergebnis})
		},
	}
}
