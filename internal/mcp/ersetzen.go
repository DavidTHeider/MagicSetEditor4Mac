package mcp

// **Suchen und Ersetzen über MCP** — Werkstück 2 von Phase 3.7
// (#1168).
//
// # Zwei Werkzeuge, und das erste schreibt nicht
//
// Ein Sprachmodell, das über dreihundert Karten ersetzt, soll **vorher** sagen
// können, was daraus wird — und der Mensch soll es lesen. `lesen.ersetzen`
// gibt die Liste, `schreiben.ersetzen` tut es. §16 Regel 6 sagt dasselbe für
// fremde Daten: nichts wird still übernommen.

import (
	"context"
	"encoding/json"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Fremdersetzung ist, was der MCP-Server vom Ersetzen braucht.
type Fremdersetzung interface {
	ErsetzungVorschau(ctx context.Context, setID int64, frage domain.Ersetzungsfrage) (
		domain.Ersetzungsvorschau, error)
	ErsetzungAusfuehren(ctx context.Context, setID int64, frage domain.Ersetzungsfrage) (
		domain.Ersetzungsvorschau, error)
}

// ersetzungswerkzeuge gibt die beiden — oder keines ohne Naht.
func ersetzungswerkzeuge(ersetzung Fremdersetzung) (lesend, schreibend []Werkzeug) {
	if ersetzung == nil {
		return nil, nil
	}

	return []Werkzeug{{
		Name: "lesen.ersetzen", Art: Lesend, Idempotent: true,
		Erklaerung: "Sagt, welche Feldwerte eines Sets sich ändern würden. " +
			"ÄNDERT NICHTS. Durchsucht nur Textfelder — ein Wahlfeld wie " +
			"»rarity« bleibt außen vor. »stellen« ist auf " +
			"200 Einträge begrenzt; »treffer« zählt alles.",
		Schema: ersetzungsschema(),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			return ersetzungsruf(ctx, roh, ersetzung.ErsetzungVorschau)
		},
	}}, []Werkzeug{{
		Name: "schreiben.ersetzen", Art: Schreibend,
		Erklaerung: "Ersetzt in den Textfeldern eines Sets. Zuvor " +
			"»lesen.ersetzen« rufen und die Änderung zeigen lassen — " +
			"zurückgenommen wird sie als EIN Schritt.",
		Schema: ersetzungsschema(),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			return ersetzungsruf(ctx, roh, ersetzung.ErsetzungAusfuehren)
		},
	}}
}

// ersetzungsschema ist das Schema beider Werkzeuge — **eines**, weil sie
// dieselbe Frage stellen.
func ersetzungsschema() map[string]any {
	return objekt(map[string]any{
		"setID":  zahl("Die Kennung des Sets aus »lesen.sets«."),
		"text":   text("Der gesuchte Text. Ohne Auszeichnungen — »<i>« zählt nicht mit."),
		"ersatz": text("Was an seine Stelle tritt; leer streicht ihn."),
		"feld": text("Nur dieses Kartenfeld; leer heißt: alle Textfelder. " +
			"Feldnamen kommen aus »lesen.felder«."),
		"grossKleinBeachten": jaNein("Groß- und Kleinschreibung unterscheiden."),
		"ganzesWort":         jaNein("Nur ganze Wörter treffen."),
	}, "setID", "text")
}

// ersetzungsruf ist der gemeinsame Rumpf beider Werkzeuge.
func ersetzungsruf(
	ctx context.Context, roh json.RawMessage,
	arbeit func(context.Context, int64, domain.Ersetzungsfrage) (
		domain.Ersetzungsvorschau, error),
) (string, error) {
	var argumente struct {
		SetID              int64  `json:"setID"`
		Text               string `json:"text"`
		Ersatz             string `json:"ersatz"`
		Feld               string `json:"feld"`
		GrossKleinBeachten bool   `json:"grossKleinBeachten"`
		GanzesWort         bool   `json:"ganzesWort"`
	}

	if err := lies(roh, &argumente); err != nil {
		return "", err
	}

	if err := pruefeKennung("setID", argumente.SetID); err != nil {
		return "", err
	}

	frage := domain.Ersetzungsfrage{
		Text:               argumente.Text,
		Ersatz:             argumente.Ersatz,
		Feld:               argumente.Feld,
		GrossKleinBeachten: argumente.GrossKleinBeachten,
		GanzesWort:         argumente.GanzesWort,
	}

	if befunde := frage.Validate(); len(befunde) > 0 {
		return "", befunde[0]
	}

	schau, err := arbeit(ctx, argumente.SetID, frage)
	if err != nil {
		return "", err
	}

	return alsJSON(map[string]any{
		"stellen": stellenliste(schau.Stellen),
		"karten":  schau.Karten,
		"treffer": schau.Treffer,
		"weitere": schau.Weitere,
	})
}

// stellenliste macht aus den Stellen eine Liste — **immer eine**, nie `null`.
func stellenliste(stellen []domain.Ersetzungsstelle) []map[string]any {
	liste := make([]map[string]any, 0, len(stellen))

	for _, stelle := range stellen {
		liste = append(liste, map[string]any{
			"karteID": stelle.KarteID,
			"nummer":  stelle.Nummer + 1,
			"feld":    stelle.Feld,
			"vorher":  stelle.Vorher,
			"nachher": stelle.Nachher,
			"treffer": stelle.Treffer,
		})
	}

	return liste
}
