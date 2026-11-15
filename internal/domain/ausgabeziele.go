// **Welches Ziel der Ausgabeassistent je Set zuletzt zeigte** — AP-10
// (#2367),
// Befund B-20 aus `docs/audits/0.9.0/AUDIT-B-OBERFLAECHE.md`.
//
// **new**: MSE2 hat keinen Assistenten, sondern je Ausgabe einen Menübefehl
// (`src/gui/set/window.cpp`, `onFileExport*`, Fassung f9d9356 vom
// 2021-07-06) — ein „letztes Ziel" gibt es dort nicht.
//
// # Warum eine Liste und kein Objekt
//
// **Weil ein Objekt seine Reihenfolge nicht hält.** Eine Kennung wie „4" ist in
// JavaScript ein ganzzahliger Schlüssel, und die stehen immer aufsteigend —
// wer die ältesten Einträge abschneiden will, braucht die Reihenfolge der
// Wahl. Die jüngste Wahl steht hinten.

package domain

import (
	"encoding/json"
	"slices"
	"strconv"
	"strings"
)

// EinstellungAusgabeziele ist der Schlüssel der Liste.
const EinstellungAusgabeziele = "ausgabe.ziele"

// HoechstensGemerkteZiele begrenzt die Liste — das Fenster schneidet die
// ältesten ab, bevor es schreibt.
const HoechstensGemerkteZiele = 200

// Ausgabeziele gibt die fünf Ziele des Assistenten — dieselben, in derselben
// Reihenfolge, wie `ausgabeziele` in `frontend/src/ausgabe.ts`.
func Ausgabeziele() []string {
	return []string{"setdatei", "bilder", "druckbogen", "vorlage", "mtgjson"}
}

// GemerktesZiel ist das Ziel eines Sets.
type GemerktesZiel struct {
	Set  string `json:"set"`
	Ziel string `json:"ziel"`
}

// GemerkteZieleAus liest die Liste — leer heißt keines, und als leere Liste.
func GemerkteZieleAus(wert string) ([]GemerktesZiel, error) {
	if strings.TrimSpace(wert) == "" {
		return []GemerktesZiel{}, nil
	}

	var ziele []GemerktesZiel
	if json.Unmarshal([]byte(wert), &ziele) != nil {
		return nil, NewFieldError("wert", SchluesselValidierung,
			"Die gemerkten Ausgabeziele sind keine Liste in JSON.")
	}

	if len(ziele) > HoechstensGemerkteZiele {
		return nil, NewFieldError("wert", SchluesselValidierung,
			"Es sind %d gemerkte Ausgabeziele, höchstens %d.", len(ziele), HoechstensGemerkteZiele)
	}

	gesehen := map[string]bool{}

	for _, eines := range ziele {
		if kennung, fehler := strconv.ParseInt(eines.Set, 10, 64); fehler != nil || kennung < 1 {
			return nil, NewFieldError("wert", SchluesselValidierung,
				"»%s« ist keine Kennung eines Sets.", eines.Set)
		}

		if !slices.Contains(Ausgabeziele(), eines.Ziel) {
			return nil, NewFieldError("wert", SchluesselValidierung,
				"Das Ausgabeziel »%s« gibt es nicht.", eines.Ziel)
		}

		if gesehen[eines.Set] {
			return nil, NewFieldError("wert", SchluesselValidierung,
				"Das Set %s steht zweimal unter den gemerkten Ausgabezielen.", eines.Set)
		}

		gesehen[eines.Set] = true
	}

	return ziele, nil
}

// PruefeAusgabeziele prüft die Liste, ohne sie zu behalten.
func PruefeAusgabeziele(wert string) error {
	_, fehler := GemerkteZieleAus(wert)

	return fehler
}
