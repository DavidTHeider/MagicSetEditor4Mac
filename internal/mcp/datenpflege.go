package mcp

// **Die Datenpflege über MCP — lesend, und das ist die Entscheidung** —
// Werkstück O8 von Phase 5.2
// (#1432).
//
// # Warum die Vorschau da ist und die Ausführung nicht
//
// **Die Analyse ändert nichts.** Sie sagt, was im Bestand herumliegt, und
// genau das ist eine Frage, die ein Modell für jemanden beantworten kann.
//
// **Die Ausführung löscht.** Die Klasse `entscheidung` heißt „ein Mensch
// entscheidet" — und ein Werkzeug, das sie ausführen könnte, machte aus der
// Klasse eine Beschriftung ([0044](../../docs/entscheidungen/0044-was-ein-werkzeug-schreiben-darf.md)).
// Dazu kommt: die Vorschau **ist** die Zustimmung, und zustimmen kann nur
// jemand, der da ist.

import (
	"context"
	"encoding/json"
)

// Pflegebefund ist eine Auffälligkeit im Bestand.
type Pflegebefund struct {
	Art       string   `json:"art"`
	Klasse    string   `json:"klasse"`
	Anzahl    int      `json:"anzahl"`
	Revision  string   `json:"revision"`
	Beispiele []string `json:"beispiele"`
}

// Pflegevorschau ist die Vorschau samt ihrer Frist.
type Pflegevorschau struct {
	Erstellt   string         `json:"erstellt"`
	Gueltigbis string         `json:"gueltigbis"`
	Befunde    []Pflegebefund `json:"befunde"`
}

// Pflegeauskunft ist die Naht zur Anwendung.
type Pflegeauskunft interface {
	Pflegevorschau(ctx context.Context) (Pflegevorschau, error)
}

func pflegewerkzeug(naht Pflegeauskunft) []Werkzeug {
	if naht == nil {
		return nil
	}

	return []Werkzeug{{
		Name: "lesen.pflegevorschau", Art: Lesend, Idempotent: true,
		Erklaerung: "Sieht nach, was im Bestand herumliegt — verwaiste Zeilen, " +
			"abgelaufene Betriebsreste, mögliche Dubletten. **Ändert nichts.** " +
			"»klasse« ist »sicher« (zeigt auf nichts und hält nichts fest), " +
			"»entscheidung« (sieht nach einem Rest aus und KANN ein Nachweis " +
			"sein) oder »nurInformieren« (wird nicht angefasst — gleiche " +
			"Printingdaten sind KEINE Dubletten, und eine offene Rückkehr ist " +
			"ein Nachweis). »revision« ist die Prüfsumme über die betroffenen " +
			"Zeilen. AUSGEFÜHRT WIRD HIER NICHTS: das Löschen gehört an eine " +
			"Fläche, an der ein Mensch sitzt.",
		Schema: objekt(map[string]any{}),
		Ruf: func(ctx context.Context, _ json.RawMessage) (string, error) {
			vorschau, fehler := naht.Pflegevorschau(ctx)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(vorschau)
		},
	}}
}
