// **Ein MSE-Script-Ausdruck über die Rückschleife** —
// #1070.
//
// # Warum `POST` und nicht `GET`
//
// Ein Ausdruck ist kein Wegbestandteil: er trägt Anführungszeichen, Klammern
// und Zeilenumbrüche, und in einer Abfragezeichenkette müsste jeder Aufrufer
// ihn verschlüsseln. **Er ändert trotzdem nichts** — das steht am Werkzeug
// und in `docs/FAEHIGKEITEN.md`, nicht am Verb.

package rest

import (
	"context"
	"net/http"
)

// Skriptkonsole ist, was der Dienst zum Auswerten braucht.
type Skriptkonsole interface {
	Auswerten(ctx context.Context, karteID int64, ausdruck string) (Skriptergebnis, error)

	// Namen zählt auf, was ein Ausdruck an dieser Karte nennen darf
	// (#1164).
	Namen(ctx context.Context, karteID int64) ([]string, error)
}

// Skriptergebnis ist, was ein ausgewerteter Ausdruck zurückmeldet.
type Skriptergebnis struct {
	Quelltext string   `json:"quelltext"`
	Text      string   `json:"text"`
	Art       string   `json:"art"`
	Befunde   []string `json:"befunde"`
}

// skript wertet einen Ausdruck im Kontext einer Karte aus.
func (d *Dienst) skript(schreiben http.ResponseWriter, anfrage *http.Request) {
	karteID, gut := kennung(schreiben, anfrage, "karteID")
	if !gut {
		return
	}

	var rumpf struct {
		Ausdruck string `json:"ausdruck"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	ergebnis, fehler := d.naehte.Konsole.Auswerten(anfrage.Context(), karteID, rumpf.Ausdruck)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, ergebnis)
}

// skriptnamen zählt auf, was ein Ausdruck an dieser Karte nennen darf.
//
// **`GET` und nicht `POST`** — anders als beim Ausdruck darüber gibt es
// nichts zu übermitteln, und die Antwort hängt allein an der Karte.
func (d *Dienst) skriptnamen(schreiben http.ResponseWriter, anfrage *http.Request) {
	karteID, gut := kennung(schreiben, anfrage, "karteID")
	if !gut {
		return
	}

	namen, fehler := d.naehte.Konsole.Namen(anfrage.Context(), karteID)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	// **Eine leere Liste und kein `null`** — dieselbe Überlegung wie am
	// Fenster: wer die Antwort liest, soll auf eines prüfen müssen.
	if namen == nil {
		namen = []string{}
	}

	ja(schreiben, http.StatusOK, map[string]any{"namen": namen})
}
