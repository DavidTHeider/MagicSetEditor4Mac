package mcp

// **Wie Kommandozeile und Werkzeugserver eingerichtet sind** — OPS-09.AK-03
// (#1907).
//
// **Nur lesend**: eingerichtet und widerrufen wird am Fenster — ein Modell,
// das in `~/.zprofile` schriebe, schriebe ins Zuhause eines Menschen, der
// nicht gefragt wurde (dieselbe Entscheidung wie bei OPS-03).

import (
	"context"
	"encoding/json"
)

// Schnittstellenauskunft ist, was das Werkzeug dieser Datei braucht.
type Schnittstellenauskunft interface {
	Schnittstellen(ctx context.Context) ([]Schnittstelle, error)
}

// Schnittstelle ist ein Ziel samt Stand.
type Schnittstelle struct {
	Name   string   `json:"name"`
	Stand  string   `json:"stand"`
	Datei  string   `json:"datei"`
	Art    string   `json:"art"`
	Zeilen []string `json:"zeilen"`
}

func schnittstellenwerkzeug(quelle Schnittstellenauskunft) []Werkzeug {
	if quelle == nil {
		return nil
	}

	return []Werkzeug{{
		Name: "lesen.schnittstellen", Art: Lesend, Idempotent: true,
		Erklaerung: "Sagt, ob die Kommandozeile im PATH der Anmeldeshell steht und ob " +
			"mse-mcp bei Codex eingetragen ist — je mit Stand (offen, eingerichtet, " +
			"veraltet, fremd, nichtGefunden, unnoetig), Datei und den Zeilen, die " +
			"einrichten hinzufügt. ÄNDERT NICHTS: eingerichtet wird am Fenster.",
		Schema: objekt(nil),
		Ruf: func(ctx context.Context, _ json.RawMessage) (string, error) {
			stand, err := quelle.Schnittstellen(ctx)
			if err != nil {
				return "", err
			}

			return alsJSON(stand)
		},
	}}
}
