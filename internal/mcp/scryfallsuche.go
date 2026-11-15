package mcp

// **Die Kartensuche über MCP** — Werkstück 1 von Phase 3.6
// (#1167).
//
// **Für ein Sprachmodell ist sie die wichtigere Hälfte des Scryfall-Zugangs.**
// Alle übrigen Wege verlangen einen Kartennamen, wie Scryfall ihn führt — wer
// ihn nicht kennt, kommt nicht weiter. Eine Suche gibt Kandidaten, und die
// Kennung darin ist der Bezug, an dem alles Weitere hängt.

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Fremdkartensuche ist, was der MCP-Server von der Suche braucht.
type Fremdkartensuche interface {
	Kartensuche(ctx context.Context, frage string) (Suchergebnis, error)
}

// Suchergebnis ist, was die Suche zurückgibt.
type Suchergebnis struct {
	Treffer      []Suchtrefferang `json:"treffer"`
	Weitere      bool             `json:"weitere"`
	AlterStunden int              `json:"alterStunden"`
	Befunde      []string         `json:"befunde"`
}

// Suchtrefferang ist eine gefundene Karte.
type Suchtrefferang struct {
	Kennung    string `json:"kennung"`
	Name       string `json:"name"`
	Setcode    string `json:"setcode"`
	Setname    string `json:"setname"`
	Nummer     string `json:"nummer"`
	Seltenheit string `json:"seltenheit"`
	Manakosten string `json:"manakosten"`
	Typzeile   string `json:"typzeile"`
	Bild       string `json:"bild"`
}

// suchwerkzeug gibt das eine Werkzeug — oder keines ohne Naht.
func suchwerkzeug(suche Fremdkartensuche) []Werkzeug {
	if suche == nil {
		return nil
	}

	return []Werkzeug{{
		Name: "lesen.scryfall.suche", Art: Lesend, Idempotent: true,
		Erklaerung: "Sucht Karten bei Scryfall. ÄNDERT NICHTS. " +
			"Die Frage ist Scryfalls eigene Suchsyntax — ein Kartenname " +
			"genügt, aber »t:creature c:r cmc<=3« geht auch. " +
			"Jeder Treffer nennt die KENNUNG des Drucks; sie ist es, die " +
			"»schreiben.scryfall.karte«, »lesen.preise« und die Sammlung " +
			"entgegennehmen. »weitere« sagt, ob Scryfall mehr hat — dann " +
			"die Frage enger fassen, es wird nicht nachgeladen.",
		Schema: objekt(map[string]any{
			"frage": text("Die Suchfrage in Scryfalls Syntax."),
		}, "frage"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var argumente struct {
				Frage string `json:"frage"`
			}

			if err := lies(roh, &argumente); err != nil {
				return "", err
			}

			if argumente.Frage == "" {
				return "", fmt.Errorf(
					"%w: es wurde keine Frage genannt", domain.ErrValidation)
			}

			ergebnis, fehler := suche.Kartensuche(ctx, argumente.Frage)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(ergebnis)
		},
	}}
}
