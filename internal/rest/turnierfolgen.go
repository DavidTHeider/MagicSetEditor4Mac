// **Die Folgen einer Berichtigung** — DOM-04.AK-02
// (#1699).
//
// Einordnung nach §16: **erweiternd** — ein Weg kommt hinzu.

package rest

import "net/http"

// turnierFolgen sagt, was eine Berichtigung verschöbe — derselbe Rumpf wie
// beim Eintragen, und **ohne ein Feld** heißt es: zurücknehmen.
func (d *Dienst) turnierFolgen(schreiben http.ResponseWriter, anfrage *http.Request) {
	turnierID, paarungID, stand, gut := ergebnisanfrage(schreiben, anfrage)
	if !gut {
		return
	}

	folgen, fehler := d.naehte.Turniere.TurnierKorrekturfolgen(
		anfrage.Context(), turnierID, paarungID, stand)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, folgen)
}

// ergebnisanfrage liest Paarung, Turnier und Spielstand einer Anfrage — für
// das Eintragen und für die Frage nach den Folgen.
//
// **Kein einziges Feld genannt heißt: zurücknehmen.** Ein Rumpf mit Nullen
// heißt 0–0–0, und das ist ein anderer Satz — deshalb Zeiger.
func ergebnisanfrage(schreiben http.ResponseWriter, anfrage *http.Request) (
	turnierID, paarungID int64, stand map[string]any, gut bool,
) {
	if paarungID, gut = kennung(schreiben, anfrage, "paarungID"); !gut {
		return 0, 0, nil, false
	}

	if turnierID, gut = kennung(schreiben, anfrage, "turnierID"); !gut {
		return 0, 0, nil, false
	}

	var wunsch struct {
		SpieleA       *int `json:"spieleA"`
		SpieleB       *int `json:"spieleB"`
		Unentschieden *int `json:"unentschieden"`
	}

	if !gelesen(schreiben, anfrage, &wunsch) {
		return 0, 0, nil, false
	}

	if wunsch.SpieleA != nil || wunsch.SpieleB != nil || wunsch.Unentschieden != nil {
		stand = map[string]any{
			"spieleA":       zahlOder(wunsch.SpieleA),
			"spieleB":       zahlOder(wunsch.SpieleB),
			"unentschieden": zahlOder(wunsch.Unentschieden),
		}
	}

	return turnierID, paarungID, stand, true
}
