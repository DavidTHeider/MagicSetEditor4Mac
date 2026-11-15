package domain

// **Gespeicherte Filter sind eine Einstellung** — UX-12.AK-03
// (#1684),
// Vermerk 0135, Punkt 7.
//
// Eine Liste aus Name, Sicht und Filter unter `filter.gespeichert`, auf der
// Ebene des Profils. **Damit überleben sie einen Neustart und reisen mit der
// Ausfuhr der Einstellungen**, ohne dass es dafür einen zweiten Weg gibt —
// die Ausfuhr nimmt, was eine Einstellung ist.

import (
	"encoding/json"
	"strings"
)

// EinstellungGespeicherteFilter ist der Schlüssel der Liste.
const EinstellungGespeicherteFilter = "filter.gespeichert"

// GespeicherterFilter ist ein Filter mit Namen, in einer Sicht.
type GespeicherterFilter struct {
	Name   string       `json:"name"`
	Sicht  Markenart    `json:"sicht"`
	Filter Markenfilter `json:"filter"`
}

// gespeicherterFilterRoh ist die Gestalt in der Einstellung — flach, damit
// sie sich von Hand schreiben lässt.
type gespeicherterFilterRoh struct {
	Name         string   `json:"name"`
	Sicht        string   `json:"sicht"`
	Marken       []string `json:"marken"`
	Verknuepfung string   `json:"verknuepfung"`
}

// GespeicherteFilterAus liest die Liste — leer heißt keine.
func GespeicherteFilterAus(wert string) ([]GespeicherterFilter, error) {
	if strings.TrimSpace(wert) == "" {
		return []GespeicherterFilter{}, nil
	}

	var roh []gespeicherterFilterRoh
	if json.Unmarshal([]byte(wert), &roh) != nil {
		return nil, NewFieldError("wert", SchluesselValidierung,
			"Die gespeicherten Filter sind keine Liste in JSON.")
	}

	filter := make([]GespeicherterFilter, 0, len(roh))
	gesehen := map[string]bool{}

	for _, eintrag := range roh {
		gelesen, fehler := gespeicherterFilterAus(eintrag)
		if fehler != nil {
			return nil, fehler
		}

		// **Derselbe Name in derselben Sicht ist einmal gemeint** — in zwei
		// Sichten sind es zwei Filter.
		schluessel := string(gelesen.Sicht) + "\x00" + strings.ToLower(gelesen.Name)
		if gesehen[schluessel] {
			return nil, NewFieldError("wert", SchluesselValidierung,
				"»%s« steht in der Sicht %s zweimal.", gelesen.Name, gelesen.Sicht)
		}

		gesehen[schluessel] = true
		filter = append(filter, gelesen)
	}

	return filter, nil
}

func gespeicherterFilterAus(roh gespeicherterFilterRoh) (GespeicherterFilter, error) {
	name := strings.Join(strings.Fields(roh.Name), " ")
	if name == "" {
		return GespeicherterFilter{}, NewFieldError("wert", SchluesselValidierung,
			"Ein gespeicherter Filter braucht einen Namen.")
	}

	sicht, fehler := PruefeMarkenart(roh.Sicht)
	if fehler != nil {
		return GespeicherterFilter{}, fehler
	}

	if len(roh.Marken) == 0 {
		return GespeicherterFilter{}, NewFieldError("wert", SchluesselValidierung,
			"»%s« nennt keine Marke — ein Filter ohne Marke zeigt alles.", name)
	}

	filter, fehler := NeuerMarkenfilter(roh.Marken, roh.Verknuepfung)
	if fehler != nil {
		return GespeicherterFilter{}, fehler
	}

	return GespeicherterFilter{Name: name, Sicht: sicht, Filter: filter}, nil
}

// PruefeGespeicherteFilter ist die Prüfung der Einstellung.
func PruefeGespeicherteFilter(wert string) error {
	_, fehler := GespeicherteFilterAus(wert)

	return fehler
}
