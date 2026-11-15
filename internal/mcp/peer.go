package mcp

// **Die eigene Peeridentität als Werkzeug** — Werkstück N1 von Phase 5.4
// (#1454).
//
// # Warum nur das Lesen
//
// **Weil das Anlegen und das Ablösen Handlungen sind, die ein Mensch
// verantwortet** ([0044](../../docs/entscheidungen/0044-was-ein-werkzeug-schreiben-darf.md)).
// Eine Rotation verlangt von **jedem** bekannten Kontakt eine neue Prüfung;
// ein Modell, das sie auslöste, entwertete Bestätigungen, die es nicht
// gegeben hat.
//
// **Das Lesen dagegen gehört hierher**: ein Agent, der einen Austausch
// vorbereitet, muss sagen können, wer dieses Programm ist.

import (
	"context"
	"encoding/json"
)

// Peeridentitaet ist die Antwort von `lesen.peeridentitaet`.
type Peeridentitaet struct {
	Stand               string   `json:"stand"`
	Kennung             string   `json:"kennung"`
	Fingerabdruck       string   `json:"fingerabdruck"`
	Seit                string   `json:"seit"`
	Vorher              string   `json:"vorher"`
	VorherFingerabdruck string   `json:"vorherFingerabdruck"`
	Befunde             []string `json:"befunde"`
}

// Peerauskunft ist die Naht zur Anwendung.
type Peerauskunft interface {
	Peeridentitaet(ctx context.Context) (Peeridentitaet, error)
}

func peerwerkzeug(naht Peerauskunft) []Werkzeug {
	if naht == nil {
		return nil
	}

	return []Werkzeug{{
		Name: "lesen.peeridentitaet", Art: Lesend, Idempotent: true,
		Erklaerung: "Nennt die Peeridentität DIESES Profils: »kennung« als " +
			"öffentliche Kennung, »fingerabdruck« als die Form, die zwei " +
			"Menschen vergleichen, »seit« als Zeitpunkt und »vorher« samt " +
			"Abdruck als die vor der letzten Ablösung. ÄNDERT NICHTS. " +
			"»stand« ist einer von vier: »ohne« (es gibt noch keine), " +
			"»bereit«, »verloren« (die Kennung ist vermerkt, der Schlüssel " +
			"fehlt — meist eine kopierte Bestandsdatei) und »fremd« " +
			"(Schlüsselbund und Bestand widersprechen sich). Bei »verloren« " +
			"und »fremd« legt dieses Werkzeug NICHTS an; das ist eine " +
			"Handlung für einen Menschen, und »befunde« sagt, welche.",
		Schema: objekt(map[string]any{}),
		Ruf: func(ctx context.Context, _ json.RawMessage) (string, error) {
			identitaet, fehler := naht.Peeridentitaet(ctx)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(identitaet)
		},
	}}
}
