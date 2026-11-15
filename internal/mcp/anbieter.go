package mcp

// **Anbieterstand und Fremdverlauf als Werkzeuge** — Werkstück A5 von
// Phase 5.3 (#1444).
//
// **Eine eigene Datei**, weil `preise.go` mit ihnen über 400 Zeilen ginge (§2).

import (
	"context"
	"encoding/json"
)

// Anbietersicht ist das Kontingent an einem Anbieterkonto.
type Anbietersicht struct {
	Dienst     string `json:"dienst"`
	Gesamt     int64  `json:"gesamt"`
	Rest       int64  `json:"rest"`
	Verbraucht int64  `json:"verbraucht"`

	// Ueberzugssatz ist Text — Geld je Kredit (0093).
	Ueberzugssatz string `json:"ueberzugssatz"`

	// HoechstensAltMinuten sagt, wie alt die Zahl beim Anbieter sein kann.
	HoechstensAltMinuten int `json:"hoechstensAltMinuten"`
}

// Verlaufsbezug ist, wonach ein Verlaufsbezug fragt.
type Verlaufsbezug struct {
	Quelle    string   `json:"quelle"`
	Printing  string   `json:"printing"`
	Finish    string   `json:"finish"`
	Zustaende []string `json:"zustaende"`
}

// Verlaufssicht ist, was er ergeben hat.
type Verlaufssicht struct {
	Quelle     string   `json:"quelle"`
	Fassung    string   `json:"fassung"`
	Zugeordnet bool     `json:"zugeordnet"`
	Staende    int      `json:"staende"`
	Befunde    []string `json:"befunde"`
}

// anbieterstandwerkzeug gibt das Werkzeug für den Kontostand.
func anbieterstandwerkzeug(werkstatt Preiswerkstatt) Werkzeug {
	return Werkzeug{
		Name: "lesen.anbieterstand", Art: Lesend,
		Erklaerung: "Nennt das Kontingent an einem Anbieterkonto — wie viele " +
			"Kredite der Tarif hat, wie viele noch übrig sind und was ein " +
			"zusätzlicher kostet. FRAGT IM NETZ und ändert nichts. " +
			"»hoechstensAltMinuten« sagt, wie alt die Zahl beim Anbieter sein " +
			"kann: sie ist KEINE Sperre, und die Grenze, die hier gilt, ist " +
			"das lokale Budget. " + vorbehalt,
		Schema: objekt(map[string]any{
			"dienst": text("Der Dienst — heute nur »scrydex«."),
		}, "dienst"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var argumente struct {
				Dienst string `json:"dienst"`
			}

			if err := lies(roh, &argumente); err != nil {
				return "", err
			}

			sicht, fehler := werkstatt.Anbieterstand(ctx, argumente.Dienst)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(sicht)
		},
	}
}

// fremdverlaufwerkzeug gibt das Werkzeug für den Preisverlauf einer Quelle.
func fremdverlaufwerkzeug(werkstatt Preiswerkstatt) Werkzeug {
	return Werkzeug{
		Name: "schreiben.fremdverlauf", Art: Schreibend,
		Erklaerung: "Holt den Preisverlauf eines Drucks bei einer Quelle und " +
			"legt ihn ab. FRAGT IM NETZ und ÄNDERT den Bestand: die " +
			"Preisstände kommen dazu und, beim ersten Mal, " +
			"die Zuordnung des Drucks zur Karte des Anbieters. Der Zustand steht so da, wie die Quelle ihn nennt — " +
			"er wird NICHT in die Skala dieses Programms übersetzt. " + vorbehalt,
		Schema: objekt(map[string]any{
			"quelle":   text("Die Quelle — heute nur »scrydex«."),
			"printing": text("Die Scryfall-Kennung des Drucks."),
			"finish":   text("Wahlfrei: nonfoil, foil oder etched."),
			"zustaende": textliste(
				"Wahlfrei: nur diese Erhaltungen durchlassen.",
				"Eine Erhaltung in der Ausdrucksweise der Quelle."),
		}, "quelle", "printing"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var wunsch Verlaufsbezug

			if err := lies(roh, &wunsch); err != nil {
				return "", err
			}

			sicht, fehler := werkstatt.Fremdpreisverlauf(ctx, wunsch)
			if fehler != nil {
				return "", fehler
			}

			// **Leere Listen bleiben Listen** — ein `null` sähe für ein
			// Sprachmodell aus wie „konnte nicht beantwortet werden".
			if sicht.Befunde == nil {
				sicht.Befunde = []string{}
			}

			return alsJSON(sicht)
		},
	}
}
