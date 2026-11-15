package mcp

// **Ein Kartenbild lesen, über MCP** — Werkstück A13.3 von Phase 5.3
// (#1452).
//
// # Warum das Werkzeug „lesen" heißt und trotzdem Geld kostet
//
// **Weil es an keiner Karte und an keiner Sammlung schreibt.** Es bucht einen
// Kostenposten und schließt ihn wieder — wie die Einschätzung nebenan. Die
// Erklärung sagt beides, damit ein Modell nicht überrascht wird.

import (
	"context"
	"encoding/json"
)

// KIBeobachtungsicht ist eine gelesene Stelle.
type KIBeobachtungsicht struct {
	Art    string `json:"art"`
	Wert   string `json:"wert"`
	Lesbar bool   `json:"lesbar"`
}

// KIBildkandidatsicht ist ein Druck, der zu den Beobachtungen passt.
type KIBildkandidatsicht struct {
	Kennung    string `json:"kennung"`
	Name       string `json:"name"`
	Setcode    string `json:"setcode"`
	Setname    string `json:"setname"`
	Nummer     string `json:"nummer"`
	Seltenheit string `json:"seltenheit"`

	// Gestuetzt sind die Evidenzarten, die **diesen** Kandidaten stützen.
	Gestuetzt []string `json:"gestuetzt"`
}

// KIErkennungsbezug ist, was für eine Erkennung gebraucht wird.
type KIErkennungsbezug struct {
	Pfad     string `json:"pfad"`
	Anbieter string `json:"anbieter"`
	Alias    string `json:"alias"`
}

// KIErkennungssicht ist, was aus dem Bild wurde.
type KIErkennungssicht struct {
	Beobachtungen []KIBeobachtungsicht `json:"beobachtungen"`
	NichtGelesen  []string             `json:"nichtGelesen"`

	Befunde []KIBelegbefundsicht `json:"befunde"`

	Kandidaten []KIBildkandidatsicht `json:"kandidaten"`

	Drucke    int  `json:"drucke"`
	Weitere   bool `json:"weitere"`
	Eindeutig bool `json:"eindeutig"`

	Alter    string `json:"alter"`
	Anbieter string `json:"anbieter"`
	Modell   string `json:"modell"`
}

// kierkennungswerkzeug gibt das Werkzeug für die Kartenerkennung.
func kierkennungswerkzeug(werkstatt Preiswerkstatt) Werkzeug {
	return Werkzeug{
		Name: "schreiben.ki.erkennung", Art: Schreibend,
		Erklaerung: "Lässt ein Modell ein Kartenfoto LESEN und nennt die " +
			"Drucke, die dazu passen. An der KARTE und an der SAMMLUNG " +
			"ändert es NICHTS — es legt KEIN Exemplar an und bindet KEINEN " +
			"Druck; wer das will, nimmt danach »schreiben.exemplar«. " +
			"ÄNDERT den Bestand trotzdem: es bucht einen Kostenposten und " +
			"schließt ihn wieder. " +
			"Das Modell muss »vision« BELEGT haben. " +
			"Die Antwort trennt zwei Dinge: »beobachtungen« ist, was das " +
			"Modell GELESEN hat (je Stelle, mit »lesbar«), »kandidaten« " +
			"sind die Drucke, die eine SUCHE dazu gefunden hat. " +
			"»drucke« ist die Zahl der passenden — das ist die " +
			"UNSICHERHEIT, und »weitere« sagt, dass es noch mehr gibt. " +
			"Es gibt KEINE Prozentangabe: welcher Druck es ist, entscheidet " +
			"ein Mensch.",
		Schema: objekt(map[string]any{
			"pfad": text("Die Bilddatei auf DIESEM Rechner. Eine Adresse " +
				"wird nicht geholt."),
			"anbieter": text("Der Anbieter, wie »lesen.ki.modelle« ihn nennt."),
			"alias":    text("Das Modell — es muss »vision« belegt haben."),
		}, "pfad", "anbieter", "alias"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var bezug KIErkennungsbezug

			if err := lies(roh, &bezug); err != nil {
				return "", err
			}

			sicht, fehler := werkstatt.KIErkennung(ctx, bezug)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(mitLeerenErkennungslisten(sicht))
		},
	}
}

// mitLeerenErkennungslisten macht aus `null` eine leere Liste.
func mitLeerenErkennungslisten(sicht KIErkennungssicht) KIErkennungssicht {
	if sicht.Beobachtungen == nil {
		sicht.Beobachtungen = []KIBeobachtungsicht{}
	}

	if sicht.NichtGelesen == nil {
		sicht.NichtGelesen = []string{}
	}

	if sicht.Befunde == nil {
		sicht.Befunde = []KIBelegbefundsicht{}
	}

	if sicht.Kandidaten == nil {
		sicht.Kandidaten = []KIBildkandidatsicht{}
	}

	for i, kandidat := range sicht.Kandidaten {
		if kandidat.Gestuetzt == nil {
			sicht.Kandidaten[i].Gestuetzt = []string{}
		}
	}

	return sicht
}
