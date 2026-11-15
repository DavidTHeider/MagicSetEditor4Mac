package mcp

import (
	"context"
	"encoding/json"
)

// Fremdbilder ist, was das Bildwerkzeug braucht.
type Fremdbilder interface {
	// Einsetzen holt das Bild und legt es an der Karte ab.
	Einsetzen(ctx context.Context, setID, karteID int64, feld, name, art string) (
		Bildeinsatz, error)
}

// Bildeinsatz ist, was ein Einsetzen zurückmeldet.
//
// **Die Marken sind die Zusage** — dieselbe Überlegung wie bei den Zeilen in
// `gestalt.go`: ohne sie nähme `encoding/json` die Go-Feldnamen, und ein
// umbenanntes Feld verschöbe stillschweigend eine Fläche, die
// [0033](../../docs/entscheidungen/0033-was-ueber-fassungen-hinweg-zugesagt-wird.md)
// über Fassungen hinweg zusagt. Angestrichen von den Flächenläufen
// (#1028):
// die Namen standen groß, aber nur, weil Go sie so schreibt.
//
// **Die Nennung reist mit**, und das ist keine Höflichkeit: Scryfall verlangt
// Künstler und Copyright in derselben Oberfläche wie das Bild. Ein Gegenüber,
// das dieses Werkzeug ruft, baut eine Oberfläche — und bekommt hier, was
// hineingehört.
type Bildeinsatz struct {
	Feld               string `json:"Feld"`
	Kuenstler          string `json:"Kuenstler"`
	Nennung            string `json:"Nennung"`
	Adresse            string `json:"Adresse"`
	IllustratorGesetzt bool   `json:"IllustratorGesetzt"`
	AlterStunden       int    `json:"AlterStunden"`

	// Befunde ist, was die Kartenauskunft tolerant hingenommen hat —
	// eine Drift beim Anbieter, die niemanden aufhält und die niemand
	// verschweigen soll (§16). Der Grund steht bei [service.Bilduebernahme].
	Befunde []string `json:"Befunde"`
}

// scryfallbildwerkzeug gibt das Werkzeug, das ein Bild einsetzt.
//
// # Warum es eines ist und nicht zwei
//
// Der Feldabgleich zeigt erst und schreibt dann, weil ein Regeltext von
// Scryfall einen getippten überschreiben könnte. **Ein Bild hat diesen Fall
// nicht**: es gibt keinen Vorschlag zu prüfen — entweder man will das Bild zu
// diesem Kartennamen oder nicht.
//
// # Und warum 0044 es trotzdem trägt
//
// [0044](../../docs/entscheidungen/0044-was-ein-werkzeug-schreiben-darf.md)
// misst die **Umkehrbarkeit**. Das Werkzeug meldet den Feldnamen und den
// vorigen Künstler zurück; zurückgenommen wird über `schreiben.feld` — genau
// wie bei der Übernahme darüber.
func scryfallbildwerkzeug(bilder Fremdbilder) []Werkzeug {
	return []Werkzeug{
		{
			Name: "schreiben.scryfall.bild", Art: Schreibend,
			Erklaerung: "ÄNDERT den Bestand: holt das Kartenbild von Scryfall und setzt es " +
				"in ein Bildfeld. Schreibt zusätzlich den Künstler nach »illustrator« — " +
				"Scryfall verlangt, dass Künstler und Copyright neben dem Bild stehen. " +
				"art ist »klein«, »normal« oder »ausschnitt«; »ausschnitt« ist die " +
				"Illustration ohne Rahmen und wird ohne bekannten Künstler abgewiesen.",
			Schema: objekt(map[string]any{
				"setID":   zahl("Die Kennung des Sets aus »lesen.sets«."),
				"karteID": zahl("Die Kennung der Karte aus »lesen.karten«."),
				"feld":    text("Der Name des Bildfelds, etwa »image«."),
				"name":    text("Der Kartenname, wie Scryfall ihn führt."),
				"art":     text("»klein«, »normal« oder »ausschnitt«."),
			}, "setID", "karteID", "feld", "name", "art"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				var argumente struct {
					SetID   int64  `json:"setID"`
					KarteID int64  `json:"karteID"`
					Feld    string `json:"feld"`
					Name    string `json:"name"`
					Art     string `json:"art"`
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

				eingesetzt, fehler := bilder.Einsetzen(ctx, argumente.SetID,
					argumente.KarteID, argumente.Feld, argumente.Name, argumente.Art)
				if fehler != nil {
					return "", fehler
				}

				return alsJSON(eingesetzt)
			},
		},
	}
}
