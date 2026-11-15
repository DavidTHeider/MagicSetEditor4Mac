// **Die schreibenden Formatwerkzeuge** — Werkstück 7 von Phase 4.6
// (#1293).
//
// **Ein Format ist Daten** (Vermerk 0061): wer ein eigenes Format spielt,
// legt es an, ohne dass jemand das Programm ändert.

package mcp

import (
	"context"
	"encoding/json"
)

// schreibenFormatBauart legt ein Format mit den Zahlen einer Bauart an.
func schreibenFormatBauart(naht Decks) Werkzeug {
	return Werkzeug{
		Name: "schreiben.format.ausbauart", Art: Schreibend,
		Erklaerung: "ÄNDERT den Bestand: legt ein Format mit den Zahlen einer " +
			"Bauart an — »konstruiert«, »commander« oder »limitiert«, siehe " +
			"»lesen.formatbauarten«. Die Zahlen lassen sich danach mit " +
			"»schreiben.format.setzen« ändern.",
		Schema: objekt(map[string]any{
			"bauart":       text("»konstruiert«, »commander« oder »limitiert«."),
			"name":         text("Der Name des Formats. Je Spiel nur einmal."),
			"spiel":        text("Der Name des Spiels, etwa »magic«."),
			"scryfallname": text("Der Schlüssel in Scryfalls legalities, etwa »standard«. Leer, wenn Scryfall dieses Format nicht kennt — dann bleibt jede Karte ungeprüft."),
		}, "bauart", "name", "spiel"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var wahl struct {
				Bauart       string `json:"bauart"`
				Name         string `json:"name"`
				Spiel        string `json:"spiel"`
				Scryfallname string `json:"scryfallname"`
			}

			if err := lies(roh, &wahl); err != nil {
				return "", err
			}

			format, fehler := naht.FormatAusBauart(
				ctx, wahl.Bauart, wahl.Name, wahl.Spiel, wahl.Scryfallname)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(format)
		},
	}
}

// schreibenFormatSetzen legt ein Format an oder schreibt es fort.
func schreibenFormatSetzen(naht Decks) Werkzeug {
	return Werkzeug{
		Name: "schreiben.format.setzen", Art: Schreibend,
		Erklaerung: "ÄNDERT den Bestand: legt ein Format an oder schreibt es " +
			"fort. Ohne id wird angelegt; mit id und fassung wird geändert. " +
			"ACHTUNG bei den vier Zahlen: 0 heißt »keine Grenze« — außer bei " +
			"sideboardgroesse, dort heißt 0 »kein Sideboard« (so ist " +
			"Commander gebaut).",
		Schema: objekt(map[string]any{
			"id":                zahl("Die Kennung — weglassen, um anzulegen."),
			"fassung":           zahl("Die Fassung aus »lesen.formate« — nötig beim Ändern."),
			"name":              text("Der Name des Formats."),
			"spiel":             text("Der Name des Spiels."),
			"mindestgroesse":    zahl("Untergrenze des Hauptdecks; 0 heißt: keine."),
			"hoechstgroesse":    zahl("Obergrenze des Hauptdecks; 0 heißt: keine."),
			"sideboardgroesse":  zahl("Obergrenze des Sideboards; 0 heißt: KEIN Sideboard."),
			"hoechstzahlGleich": zahl("Wie oft dieselbe Karte vorkommen darf; 0 heißt: beliebig oft. Standardländer sind ausgenommen."),
			"scryfallname":      text("Der Schlüssel in Scryfalls legalities."),
		}, "name", "spiel"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var wahl map[string]any

			if err := lies(roh, &wahl); err != nil {
				return "", err
			}

			format, fehler := naht.FormatSetzen(ctx, wahl)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(format)
		},
	}
}

// schreibenFormatLoeschen nimmt ein Format heraus.
func schreibenFormatLoeschen(naht Decks) Werkzeug {
	return Werkzeug{
		Name: "schreiben.format.loeschen", Art: Schreibend,
		Erklaerung: "ÄNDERT den Bestand: nimmt ein Format heraus. Die Decks, " +
			"die darauf standen, BLEIBEN — sie sind danach ungeprüft, nicht " +
			"gelöscht.",
		Schema: objekt(map[string]any{
			"formatID": zahl("Die Kennung des Formats."),
		}, "formatID"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var wahl struct {
				FormatID int64 `json:"formatID"`
			}

			if err := lies(roh, &wahl); err != nil {
				return "", err
			}

			if fehler := naht.FormatLoeschen(ctx, wahl.FormatID); fehler != nil {
				return "", fehler
			}

			return alsJSON(map[string]any{
				"formatID": wahl.FormatID, "geloescht": true,
				"hinweis": "Decks, die auf diesem Format standen, sind jetzt ungeprüft.",
			})
		},
	}
}
