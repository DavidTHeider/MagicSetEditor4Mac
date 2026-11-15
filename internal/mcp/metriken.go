package mcp

// **Die Metriken über MCP** — Werkstück O9 von Phase 5.2
// (#1433).

import (
	"context"
	"encoding/json"
)

// Metrikauskunft ist die Naht zur Anwendung.
//
// **Die Antwort der Anwendung, durchgereicht** — seit OPS-05 (#1722), wie
// beim Dienst. Hier stand eine eigene Gestalt mit zwei Feldern, und sie
// wäre hinter der Anwendung zurückgeblieben, sobald diese mehr sagt: dieselbe
// Fähigkeit, zwei Antworten (§16).
type Metrikauskunft interface {
	Metriken(ctx context.Context) (any, error)
}

func metrikwerkzeug(naht Metrikauskunft) []Werkzeug {
	if naht == nil {
		return nil
	}

	return []Werkzeug{{
		Name: "lesen.metriken", Art: Lesend, Idempotent: true,
		Erklaerung: "Nennt, was dieser Prozess seit seinem Start gewirkt hat: " +
			"je Fläche, Wirkungsklasse und Ausgang eine Zahl. Die Zahl der " +
			"Reihen ist durch »grenze« nach oben beschränkt und wächst NICHT " +
			"mit der Zahl der Karten, Sets oder Profile — es sind drei " +
			"Aufzählungen und keine Kennungen. Gezählt werden geprüfte " +
			"Wirkungen (Datei, Netz, verbindlich), nicht jeder Ruf. " +
			"Dazu: Laufzeiten je Fläche und Art in Eimern (Millisekunden), " +
			"Verbindungen zum Bestand, Zwischenspeicher, Arbeitsspeicher " +
			"(Bytes), Erhebungszeitraum, Messaufwand und was nicht gemessen " +
			"ist; die Ablaufspur, wenn sie eingeschaltet ist. " +
			"ÄNDERT NICHTS.",
		Schema: objekt(map[string]any{}),
		Ruf: func(ctx context.Context, _ json.RawMessage) (string, error) {
			stand, fehler := naht.Metriken(ctx)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(stand)
		},
	}}
}
