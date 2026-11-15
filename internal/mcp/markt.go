package mcp

// **Der Marktplatz als Werkzeug** — Werkstück A4 von Phase 5.3
// (#1443).
//
// **Eine eigene Datei**, weil `preise.go` mit ihr über 400 Zeilen ginge (§2).

import (
	"context"
	"encoding/json"
)

// Marktwunsch ist, wonach ein Marktabruf fragt.
type Marktwunsch struct {
	Quelle    string   `json:"quelle"`
	Printing  string   `json:"printing"`
	Sprache   string   `json:"sprache"`
	Finish    string   `json:"finish"`
	Zustaende []string `json:"zustaende"`
}

// Marktsicht ist, was ein Marktabruf ergeben hat.
type Marktsicht struct {
	Quelle     string          `json:"quelle"`
	Fassung    string          `json:"fassung"`
	Zugeordnet bool            `json:"zugeordnet"`
	Angebote   []Angebotssicht `json:"angebote"`
	Befunde    []string        `json:"befunde"`
}

// Angebotssicht ist ein Posten auf einem fremden Marktplatz.
type Angebotssicht struct {
	Quelle  string `json:"quelle"`
	Kennung string `json:"kennung"`

	// Wert ist der Preis als Text — `"2.17"`. **Nicht als Zahl**: eine
	// Gleitkommazahl verlöre Cent (0093).
	Wert     string `json:"wert"`
	Waehrung string `json:"waehrung"`

	Menge   int `json:"menge"`
	Buendel int `json:"buendel"`

	Finish  string `json:"finish"`
	Zustand string `json:"zustand"`
	Sprache string `json:"sprache"`

	Signiert   bool `json:"signiert"`
	Veraendert bool `json:"veraendert"`
	Gradiert   bool `json:"gradiert"`
	ImUrlaub   bool `json:"imUrlaub"`

	Verkaeufer string `json:"verkaeufer"`
	Land       string `json:"land"`
}

// marktwerkzeug gibt das Werkzeug für die Marktangebote.
//
// **Lesend, und es fragt im Netz** — anders als die übrigen Preiswerkzeuge.
// Ein Marktpreis ist eine Momentaufnahme; es gibt nichts Abgelegtes, das man
// stattdessen zeigen könnte.
func marktwerkzeug(werkstatt Preiswerkstatt) Werkzeug {
	return Werkzeug{
		Name: "lesen.marktangebote", Art: Lesend,
		Erklaerung: "Gibt die Angebote eines Marktplatzes zu einem Druck — " +
			"Preis, Menge, Zustand, Sprache und Verkäufer je Posten. " +
			"FRAGT IM NETZ und ändert am Bestand nur die Zuordnung des Drucks " +
			"zur Kartenfassung des Anbieters. Es wird NICHTS gekauft und " +
			"NICHTS angeboten. Der Zustand steht so da, wie der Anbieter ihn " +
			"nennt — er wird NICHT in die Skala dieses Programms übersetzt, " +
			"weil die Skalen verschieden sind. " + vorbehalt,
		Schema: objekt(map[string]any{
			"quelle":   text("Der Marktplatz — heute nur »cardtrader«."),
			"printing": text("Die Scryfall-Kennung des Drucks."),
			"sprache":  text("Wahlfrei: der Sprachcode, etwa »en« oder »de«."),
			"finish":   text("Wahlfrei: nonfoil, foil oder etched."),
			"zustaende": textliste(
				"Wahlfrei: nur diese Erhaltungen durchlassen.",
				"Eine Erhaltung in der Ausdrucksweise des Anbieters, "+
					"etwa »Near Mint«."),
		}, "quelle", "printing"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var wunsch Marktwunsch

			if err := lies(roh, &wunsch); err != nil {
				return "", err
			}

			sicht, fehler := werkstatt.Marktangebote(ctx, wunsch)
			if fehler != nil {
				return "", fehler
			}

			// **Leere Listen bleiben Listen** — ein `null` im JSON sähe für
			// ein Sprachmodell aus wie „konnte nicht beantwortet werden".
			if sicht.Angebote == nil {
				sicht.Angebote = []Angebotssicht{}
			}

			if sicht.Befunde == nil {
				sicht.Befunde = []string{}
			}

			return alsJSON(sicht)
		},
	}
}
