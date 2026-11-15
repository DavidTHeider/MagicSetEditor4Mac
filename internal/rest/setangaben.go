// Die Setangaben, die Vorbehalte und die mitgelieferten Sets über REST
// (#1107).
//
// **Drei Fähigkeiten des Fensters, die kein Fenster brauchen.** Ein Set hat
// Titel, Setcode und Copyright; wer hundert Sets aus einem Skript anlegt, will
// sie setzen können, ohne eine Lasche aufzuschlagen.

package rest

import (
	"net/http"
)

// angaben gibt die Angabenfelder eines Sets samt ihren Werten.
func (d *Dienst) angaben(schreiben http.ResponseWriter, anfrage *http.Request) {
	setID, gut := kennung(schreiben, anfrage, "setID")
	if !gut {
		return
	}

	felder, werte, fehler := d.naehte.Angaben.Angaben(anfrage.Context(), setID)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	// **Beide Listen nie `null`** — dieselbe Zusage wie überall hier.
	ja(schreiben, http.StatusOK, map[string]any{
		"felder": ohneNil(felder),
		"werte":  ohneNil(werte),
	})
}

// angabeSetzen schreibt eine Angabe und nennt die vorige.
func (d *Dienst) angabeSetzen(schreiben http.ResponseWriter, anfrage *http.Request) {
	setID, gut := kennung(schreiben, anfrage, "setID")
	if !gut {
		return
	}

	var rumpf struct {
		Wert string `json:"wert"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	vorher, hatte, fehler := d.naehte.Angaben.Setzen(
		anfrage.Context(), setID, anfrage.PathValue("feld"), rumpf.Wert)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	// **Der vorige Wert reist mit**, damit sich die Änderung zurücknehmen
	// lässt — dieselbe Zusage wie beim Kartenfeld.
	ja(schreiben, http.StatusOK, map[string]any{"vorher": vorher, "hatte": hatte})
}

// vorbehalte gibt, was beim Einfahren nicht gedeutet wurde.
func (d *Dienst) vorbehalte(schreiben http.ResponseWriter, anfrage *http.Request) {
	setID, gut := kennung(schreiben, anfrage, "setID")
	if !gut {
		return
	}

	bloecke, fehler := d.naehte.Angaben.Vorbehalte(anfrage.Context(), setID)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"vorbehalte": ohneNil(bloecke)})
}

// beispielsets nennt die mitgelieferten Sets samt Pfad.
func (d *Dienst) beispielsets(schreiben http.ResponseWriter, anfrage *http.Request) {
	sets, fehler := d.naehte.Angaben.Beispielsets()
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	// **Ohne Kontext**: sie liest ein Verzeichnis im Bündel und fragt den
	// Bestand nichts. Der Parameter bleibt trotzdem stehen, weil `HandleFunc`
	// ihn verlangt.
	_ = anfrage

	ja(schreiben, http.StatusOK, map[string]any{"sets": ohneNil(sets)})
}

// ohneNil gibt eine Liste, die im JSON nie `null` wird.
//
// **Einmal für alle Listen dieses Dienstes.** Ein `nil` aus einem
// Anwendungsfall wird sonst zu `null`, und ein Aufrufer bekommt statt einer
// leeren Liste einen Wert, den er nicht durchlaufen kann (§3, „Zwischen den
// Prüfebenen").
func ohneNil[T any](liste []T) []T {
	if liste == nil {
		return []T{}
	}

	return liste
}
