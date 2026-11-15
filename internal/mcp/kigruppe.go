package mcp

// **Die atomare Gruppe über MCP** — Werkstück A13.1b von Phase 5.3
// (#1452).
//
// # Warum es dieses Werkzeug neben `schreiben.ki.vorschlag` gibt
//
// **Weil eine Schleife über das andere etwas anderes tut.** Wer drei
// Vorschläge einzeln übernimmt, hat drei Transaktionen; bricht die dritte ab,
// stehen zwei Karten geändert da. Das Issue nennt genau das: „entweder
// vollständig … oder klarer Konflikt **ohne Teilbestand**."
//
// # Und warum es ein Modell das wissen muss
//
// **Weil die Erklärung die einzige Stelle ist, an der es steht.** Ein Modell
// sieht die Werkzeugliste und sonst nichts; steht dort nicht, dass dieses
// Werkzeug alles-oder-nichts ist, nimmt es das andere in einer Schleife.

import (
	"context"
	"encoding/json"
)

// KIGruppenwahlsicht ist ein Vorschlag mit den Feldern, die gelten sollen.
type KIGruppenwahlsicht struct {
	Kennung string   `json:"kennung"`
	Felder  []string `json:"felder"`
}

// KIGruppenbezug ist, was für eine Gruppenübernahme gebraucht wird.
type KIGruppenbezug struct {
	Wahl []KIGruppenwahlsicht `json:"wahl"`
}

// KIGruppenkartensicht nennt eine Karte und ihre geschriebenen Felder.
type KIGruppenkartensicht struct {
	KarteID int64    `json:"karteId"`
	Felder  []string `json:"felder"`
}

// KIGruppensicht sagt, was aus der Gruppe wurde.
type KIGruppensicht struct {
	Karten int                    `json:"karten"`
	Felder int                    `json:"felder"`
	Namen  []KIGruppenkartensicht `json:"namen"`
}

// kigruppenwerkzeug gibt das Werkzeug für die atomare Übernahme.
func kigruppenwerkzeug(werkstatt Preiswerkstatt) Werkzeug {
	return Werkzeug{
		Name: "schreiben.ki.gruppe", Art: Schreibend,
		Erklaerung: "Übernimmt MEHRERE abgelegte Vorschläge als EINE " +
			"Gruppe — GANZ ODER GAR NICHT. Scheitert ein einziger Schritt " +
			"(der Vorschlag ist weg, die Frist abgelaufen, die Karte hat " +
			"sich geändert, ein Feldname stimmt nicht), wird KEINE Karte " +
			"geschrieben und KEIN Vorschlag verbraucht — ein zweiter " +
			"Versuch ist möglich. ÄNDERT Karten. " +
			"Benutze dies statt »schreiben.ki.vorschlag« in einer Schleife: " +
			"die wäre eine Transaktion je Karte, und ein Abbruch in der " +
			"Mitte ließe die halbe Gruppe geschrieben stehen. " +
			"Dieselbe Karte zweimal wird abgewiesen: die zweite Zeile " +
			"überschriebe die erste, ohne dass es jemand sieht.",
		Schema: objekt(map[string]any{
			"wahl": objektliste(
				"Die Vorschläge samt den Feldern, die aus ihnen gelten sollen.",
				objekt(map[string]any{
					"kennung": text("Die Kennung des Auftrags, wie " +
						"»schreiben.ki.batch« sie je Ziel nennt."),
					"felder": textliste(
						"Die Feldnamen, die übernommen werden sollen — "+
							"LEER ist ein Fehler und nicht »alles«.",
						"Ein Feldname aus dem Vorschlag."),
				}, "kennung", "felder")),
		}, "wahl"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var bezug KIGruppenbezug

			if err := lies(roh, &bezug); err != nil {
				return "", err
			}

			sicht, fehler := werkstatt.KIGruppe(ctx, bezug)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(mitLeererGruppe(sicht))
		},
	}
}

// mitLeererGruppe macht aus `null` eine leere Liste.
func mitLeererGruppe(sicht KIGruppensicht) KIGruppensicht {
	if sicht.Namen == nil {
		sicht.Namen = []KIGruppenkartensicht{}
	}

	for i, zeile := range sicht.Namen {
		if zeile.Felder == nil {
			sicht.Namen[i].Felder = []string{}
		}
	}

	return sicht
}
