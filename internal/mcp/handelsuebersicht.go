package mcp

// **Die Liste der Handelsvorgänge über MCP** — TRADE-01
// (#1715).
//
// **Lesend, aus demselben Grund wie `lesen.handel`**: welche Handel es gibt
// und wo sie stehen, ist Auskunft über den eigenen Bestand und verlässt den
// Rechner nicht. Entwürfe bearbeiten, kopieren, löschen und archivieren gibt
// es hier **nicht** — schon das Eröffnen eines Entwurfs fehlt an dieser
// Fläche ([0106](../../docs/entscheidungen/0106-drei-vertrauensgrenzen-und-nur-eine-wird-gebaut.md)).

import (
	"context"
	"encoding/json"
)

// Handelsfilter ist, was ein Modell von der Liste sehen will — dieselben
// Angaben wie an den anderen Flächen.
type Handelsfilter struct {
	Partner string   `json:"partner"`
	Lagen   []string `json:"lagen"`
	Suche   string   `json:"suche"`
	Marke   string   `json:"marke"`
	Seit    string   `json:"seit"`
	Bis     string   `json:"bis"`
	Ordnung string   `json:"ordnung"`
	Archiv  bool     `json:"archiv"`
}

// Handelspartner ist ein Gegenüber einer Zeile.
type Handelspartner struct {
	Kennung string `json:"kennung"`
	Name    string `json:"name,omitempty"`
}

// Handelsuebersichtszeile ist ein Vorgang in der Liste.
type Handelsuebersichtszeile struct {
	Vorgang      Handelsstand     `json:"vorgang"`
	Lage         string           `json:"lage"`
	Naechster    string           `json:"naechster,omitempty"`
	Gegenueber   []Handelspartner `json:"gegenueber"`
	Karten       []string         `json:"karten"`
	Marken       []string         `json:"marken"`
	Angelegt     string           `json:"angelegt"`
	Geaendert    string           `json:"geaendert"`
	Archiviert   bool             `json:"archiviert"`
	Unbestaetigt bool             `json:"unbestaetigt"`
}

// handelsuebersichtwerkzeug liest die Liste.
func handelsuebersichtwerkzeug(auskunft Handelsauskunft) Werkzeug {
	return Werkzeug{
		Name: "lesen.handelsuebersicht", Art: Lesend, Idempotent: true,
		Erklaerung: "Nennt die Handelsvorgänge dieses Profils — je Vorgang die letzte " +
			"Revision mit Lage (dran, wartet, abgeschlossen, abgebrochen), nächstem " +
			"Schritt, Gegenübern, Karten und Marken. Filtert nach Gegenüber, Lagen, " +
			"Suche, Marke und Zeitraum; ohne `archiv` bleibt das Archiv draußen. " +
			"LIEST NUR: einen Handel bewegt über diese Fläche niemand.",
		Schema: objekt(map[string]any{
			"partner": text("Die Peerkennung eines Gegenübers."),
			"lagen": map[string]any{
				"type": "array", "items": map[string]any{"type": "string"},
				"description": "dran, wartet, abgeschlossen oder abgebrochen.",
			},
			"suche":   text("Teil einer Kennung, eines Namens, einer Karte oder Marke."),
			"marke":   text("Eine Marke."),
			"seit":    text("Ein Tag (2026-09-30) oder ein Zeitpunkt nach RFC 3339."),
			"bis":     text("Ein Tag oder ein Zeitpunkt — ein Tag zählt bis zu seinem Ende."),
			"ordnung": text("geaendert (Vorgabe), angelegt, lage oder partner."),
			"archiv":  map[string]any{"type": "boolean", "description": "Auch archivierte Vorgänge."},
		}),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var filter Handelsfilter
			if err := lies(roh, &filter); err != nil {
				return "", err
			}

			zeilen, err := auskunft.Handelsuebersicht(ctx, filter)
			if err != nil {
				return "", err
			}

			return alsJSON(map[string]any{"zeilen": zeilen})
		},
	}
}
