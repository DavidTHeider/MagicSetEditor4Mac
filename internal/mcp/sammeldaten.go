package mcp

import (
	"context"
	"encoding/json"
)

// Fremdsammeldaten ist, was die Sammeldatenwerkzeuge brauchen.
type Fremdsammeldaten interface {
	Stand(ctx context.Context) (Sammelstand, error)
	Holen(ctx context.Context) (Sammelstand, error)
}

// Sammelstand sagt, was auf der Platte liegt.
//
// **Die Marken sind die Zusage** — siehe [Satz].
type Sammelstand struct {
	Art       string `json:"Art"`
	Pfad      string `json:"Pfad"`
	Vorhanden bool   `json:"Vorhanden"`
	Bytes     int64  `json:"Bytes"`
	Angeboten int64  `json:"Angeboten"`
	Stand     string `json:"Stand"`
}

// sammeldatenwerkzeuge gibt das lesende und das holende Werkzeug.
//
// # Warum zwei und nicht eines mit einer Fahne
//
// **Weil das eine liest und das andere 25 MB lädt.** Ein Werkzeug mit einem
// `holen: true` wäre für ein Sprachmodell ein Aufruf wie jeder andere — und
// die Liste in `tools/list` sagte nicht mehr, welcher davon etwas tut. Die
// Trennung ist dieselbe wie zwischen `lesen.` und `schreiben.` (0010).
func sammeldatenwerkzeuge(daten Fremdsammeldaten) (lesend, holend []Werkzeug) {
	lesend = []Werkzeug{{
		Name: "lesen.scryfall.sammeldaten", Art: Lesend, Idempotent: true,
		Erklaerung: "Sagt, ob die Sammeldatei von Scryfall auf dieser Maschine liegt " +
			"und wie groß sie ist. ÄNDERT NICHTS und lädt nichts. Die Datei ist " +
			"der Rückfall, aus dem Kartenangaben kommen, wenn Scryfall " +
			"nicht erreichbar ist.",
		Ruf: func(ctx context.Context, _ json.RawMessage) (string, error) {
			stand, fehler := daten.Stand(ctx)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(stand)
		},
	}}

	holend = []Werkzeug{{
		Name: "schreiben.scryfall.sammeldaten", Art: Schreibend,
		Erklaerung: "ÄNDERT diese Maschine: lädt die Sammeldatei von Scryfall " +
			"herunter — rund 25 MB — und legt sie ab, eine vorhandene wird " +
			"ersetzt. Ändert **keine** Karte und kein Set. Danach arbeiten die " +
			"Scryfall-Wege auch ohne Netz, mit einem Stand von bis zu einem Tag.",
		Ruf: func(ctx context.Context, _ json.RawMessage) (string, error) {
			stand, fehler := daten.Holen(ctx)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(stand)
		},
	}}

	return lesend, holend
}
