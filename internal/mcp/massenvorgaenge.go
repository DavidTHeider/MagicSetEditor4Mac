// **Massenvorgänge und Preise in Masse, über MCP** — Werkstück 6 von
// Phase 4.5 (#1292).

package mcp

import (
	"context"
	"encoding/json"
)

// Massenzeile ist ein Exemplar samt seiner gelesenen Fassung.
type Massenzeile struct {
	ID      int64 `json:"id"`
	Fassung int64 `json:"fassung"`
}

// Massenaenderung sagt, was an den genannten Zeilen zu ändern ist.
type Massenaenderung struct {
	Zeilen   []Massenzeile `json:"zeilen"`
	Menge    *int          `json:"menge"`
	Zustand  *string       `json:"zustand"`
	Lagerort *string       `json:"lagerort"`
	Wunsch   *bool         `json:"wunsch"`
}

// Massenbericht sagt, was geschehen ist.
//
// **`vorher` ist erweiternd** im Sinne von 0033 (DOM-03.AK-03, #1698): je
// geänderter Zeile der Stand davor samt neuer Fassung — damit ein Aufrufer
// die Änderung mit demselben Werkzeug zurücknehmen kann.
type Massenbericht struct {
	Geaendert int           `json:"geaendert"`
	Befunde   []string      `json:"befunde"`
	Vorher    []Massenstand `json:"vorher"`
}

// Massenstand ist eine Zeile vor der Änderung, mit der Fassung danach.
type Massenstand struct {
	ID       int64  `json:"id"`
	Fassung  int64  `json:"fassung"`
	Menge    int    `json:"menge"`
	Zustand  string `json:"zustand"`
	Lagerort string `json:"lagerort"`
	Wunsch   bool   `json:"wunsch"`
}

// Preisschub sagt, was ein Schub geholt hat und was bleibt.
type Preisschub struct {
	Geholt int `json:"geholt"`
	Offen  int `json:"offen"`

	// Fehlgeschlagen ist die Zahl der Versuche, die nichts eingetragen haben
	// — der Grund steht bei [app.PreisschubDTO].
	Fehlgeschlagen int `json:"fehlgeschlagen"`

	Fertig bool `json:"fertig"`

	// Steht sagt, dass dieser Schub nichts bewegt hat, obwohl Arbeit bleibt.
	Steht bool `json:"steht"`

	Befunde []string `json:"befunde"`
}

// Massenvorgaenge ist, was MCP für Vorgänge über eine Auswahl braucht.
type Massenvorgaenge interface {
	Aendern(ctx context.Context, aenderung Massenaenderung) (Massenbericht, error)
	Loeschen(ctx context.Context, zeilen []Massenzeile) (Massenbericht, error)
	PreiseInMasse(ctx context.Context, hoechstensStunden int) (Preisschub, error)
}

// massenvorgaenge gibt die drei Werkzeuge — alle schreibend.
//
// **Auch das Preisholen**: es trägt Preisstände in den Bestand ein.
func massenvorgaenge(vorgaenge Massenvorgaenge) []Werkzeug {
	if vorgaenge == nil {
		return nil
	}

	zeilenschema := objektliste("Die Zeilen samt ihrer gelesenen Fassung.",
		objekt(map[string]any{
			"id":      zahl("Die Kennung des Exemplars aus »lesen.sammlung«."),
			"fassung": zahl("Die Fassung, die dabei gelesen wurde."),
		}, "id", "fassung"))

	return []Werkzeug{{
		Name: "schreiben.sammlung.masse", Art: Schreibend,
		Erklaerung: "ÄNDERT viele Exemplare auf einmal. Jede Zeile nennt ihre " +
			"KENNUNG UND IHRE FASSUNG — stimmt die Fassung nicht mehr, hat " +
			"jemand anderes geschrieben, und diese Zeile wird übergangen statt " +
			"überschrieben. Was nicht genannt ist, bleibt: ein fehlendes Feld " +
			"heißt »unverändert«, ein leerer Lagerort heißt »Angabe entfernen«.",
		Schema: objekt(map[string]any{
			"zeilen":   zeilenschema,
			"menge":    zahl("Die neue Stückzahl — weglassen heißt unverändert."),
			"zustand":  text("Der neue Zustand — weglassen heißt unverändert."),
			"lagerort": text("Der neue Lagerort — leer entfernt die Angabe."),
			"wunsch":   jaNein("Ob die Zeilen zur Wunschliste zählen."),
		}, "zeilen"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var argumente Massenaenderung

			if err := lies(roh, &argumente); err != nil {
				return "", err
			}

			bericht, err := vorgaenge.Aendern(ctx, argumente)
			if err != nil {
				return "", mitBericht(err, nil, bericht.Geaendert)
			}

			return alsJSON(bericht)
		},
	}, {
		Name: "schreiben.sammlung.masseloeschen", Art: Schreibend,
		Erklaerung: "LÖSCHT viele Exemplare auf einmal — ENDGÜLTIG. Jede Zeile " +
			"nennt ihre Kennung und ihre Fassung; eine Zeile, die inzwischen " +
			"jemand anderes geändert hat, wird übergangen und im Bericht " +
			"genannt.",
		Schema: objekt(map[string]any{"zeilen": zeilenschema}, "zeilen"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var argumente struct {
				Zeilen []Massenzeile `json:"zeilen"`
			}

			if err := lies(roh, &argumente); err != nil {
				return "", err
			}

			bericht, err := vorgaenge.Loeschen(ctx, argumente.Zeilen)
			if err != nil {
				return "", mitBericht(err, nil, bericht.Geaendert)
			}

			return alsJSON(bericht)
		},
	}, {
		Name: "schreiben.sammlung.preisschub", Art: Schreibend,
		Erklaerung: "ÄNDERT den Bestand: holt Preise für die Drucke der " +
			"Sammlung und trägt sie ein — EINEN SCHUB je Aufruf. »offen« sagt, " +
			"wie viele noch fehlen; rufe wieder, bis »fertig« wahr ist. Der " +
			"Grund für die Schübe ist die Drossel bei Scryfall: fünfhundert " +
			"Drucke sind rund fünf Minuten. Ein Druck, dessen Preis jünger ist " +
			"als »hoechstensStunden«, wird übergangen.",
		Schema: objekt(map[string]any{
			"hoechstensStunden": zahl("Wie alt ein Preis sein darf, um zu " +
				"bleiben. 0 holt alles neu."),
		}, "hoechstensStunden"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var argumente struct {
				HoechstensStunden int `json:"hoechstensStunden"`
			}

			if err := lies(roh, &argumente); err != nil {
				return "", err
			}

			schub, err := vorgaenge.PreiseInMasse(ctx, argumente.HoechstensStunden)
			if err != nil {
				return "", err
			}

			return alsJSON(schub)
		},
	}}
}
