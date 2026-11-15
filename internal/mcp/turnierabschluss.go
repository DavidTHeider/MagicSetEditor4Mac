// **Abschluss und Protokoll eines Turniers** — DOM-04.AK-04
// (#1699),
// Vermerk 0139.
//
// Einordnung nach §16: **erweiternd** — drei Werkzeuge kommen hinzu.

package mcp

import (
	"context"
	"encoding/json"
)

// turnierkennung ist das eine Feld dieser drei Werkzeuge.
func turnierkennung() map[string]any {
	return objekt(map[string]any{
		"turnierID": zahl("Die Kennung des Turniers aus »lesen.turniere«."),
	}, "turnierID")
}

// liesTurnierkennung liest es.
func liesTurnierkennung(roh json.RawMessage) (int64, error) {
	var wahl struct {
		TurnierID int64 `json:"turnierID"`
	}

	return wahl.TurnierID, lies(roh, &wahl)
}

// schreibenAbschliessen schließt ein Turnier ab.
func schreibenAbschliessen(naht Turniere) Werkzeug {
	return Werkzeug{
		Name: "schreiben.turnier.abschliessen", Art: Schreibend,
		Erklaerung: "Schließt ein Turnier ab — nur, wenn es mindestens eine Runde " +
			"hat und jede Paarung ein Ergebnis. Danach nimmt es KEINE Änderung " +
			"mehr an: kein Anmelden, keine Runde, kein Ergebnis — bis es wieder " +
			"eröffnet wird. Das Protokoll hält beides fest.",
		Schema: turnierkennung(),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			turnierID, err := liesTurnierkennung(roh)
			if err != nil {
				return "", err
			}

			if fehler := naht.TurnierAbschliessen(ctx, turnierID); fehler != nil {
				return "", fehler
			}

			return "abgeschlossen", nil
		},
	}
}

// schreibenWiedereroeffnen nimmt den Abschluss zurück.
func schreibenWiedereroeffnen(naht Turniere) Werkzeug {
	return Werkzeug{
		Name: "schreiben.turnier.wiedereroeffnen", Art: Schreibend,
		Erklaerung: "Eröffnet ein abgeschlossenes Turnier wieder — danach lassen " +
			"sich Teilnehmende, Runden und Ergebnisse wieder ändern. Ein offenes " +
			"Turnier weist es ab.",
		Schema: turnierkennung(),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			turnierID, err := liesTurnierkennung(roh)
			if err != nil {
				return "", err
			}

			if fehler := naht.TurnierWiedereroeffnen(ctx, turnierID); fehler != nil {
				return "", fehler
			}

			return "wiedereröffnet", nil
		},
	}
}

// lesenProtokoll gibt, was an einem Turnier geschah.
func lesenProtokoll(naht Turniere) Werkzeug {
	return Werkzeug{
		Name: "lesen.turnier.protokoll", Art: Lesend,
		Erklaerung: "Nennt, was an einem Turnier geschah, in der Reihenfolge, in " +
			"der es geschah: angelegt, angemeldet, aus- und eingestiegen, Runde " +
			"gelegt, Ergebnis eingetragen oder zurückgenommen, abgeschlossen, " +
			"wiedereröffnet. »angaben« ist ein Name, eine Rundennummer oder ein " +
			"Spielstand a:b:u. ÄNDERT NICHTS.",
		Schema: turnierkennung(),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			turnierID, err := liesTurnierkennung(roh)
			if err != nil {
				return "", err
			}

			protokoll, fehler := naht.TurnierProtokoll(ctx, turnierID)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(protokoll)
		},
	}
}
