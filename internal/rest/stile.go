// Die Stilwahl und die Stiloptionen über REST
// (#1107).

package rest

import "net/http"

// stile nennt die wählbaren Stile eines Sets.
func (d *Dienst) stile(schreiben http.ResponseWriter, anfrage *http.Request) {
	setID, gut := kennung(schreiben, anfrage, "setID")
	if !gut {
		return
	}

	angebote, fehler := d.naehte.Stile.Stile(anfrage.Context(), setID)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"stile": ohneNil(angebote)})
}

// stilfolgen sagt, was ein Wechsel kostet — **bevor** er stattfindet.
func (d *Dienst) stilfolgen(schreiben http.ResponseWriter, anfrage *http.Request) {
	setID, gut := kennung(schreiben, anfrage, "setID")
	if !gut {
		return
	}

	folgen, fehler := d.naehte.Stile.Folgen(
		anfrage.Context(), setID, anfrage.PathValue("paket"))
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{
		"ungedeuteteFelder": ohneNil(folgen.UngedeuteteFelder),
		"verloreneOptionen": ohneNil(folgen.VerloreneOptionen),
		"spielwechsel":      folgen.Spielwechsel,
	})
}

// stilWechseln setzt den Stil eines Sets.
func (d *Dienst) stilWechseln(schreiben http.ResponseWriter, anfrage *http.Request) {
	setID, gut := kennung(schreiben, anfrage, "setID")
	if !gut {
		return
	}

	var rumpf struct {
		Paket string `json:"paket"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	if fehler := d.naehte.Stile.Wechseln(
		anfrage.Context(), setID, rumpf.Paket); fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{})
}

// stiloptionen gibt die Optionen des gewählten Stils samt Werten.
func (d *Dienst) stiloptionen(schreiben http.ResponseWriter, anfrage *http.Request) {
	setID, gut := kennung(schreiben, anfrage, "setID")
	if !gut {
		return
	}

	optionen, fehler := d.naehte.Stile.Optionen(anfrage.Context(), setID)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"optionen": ohneNil(optionen)})
}

// stiloptionSetzen schreibt den Wert einer Stiloption.
func (d *Dienst) stiloptionSetzen(schreiben http.ResponseWriter, anfrage *http.Request) {
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

	if fehler := d.naehte.Stile.OptionSetzen(
		anfrage.Context(), setID, anfrage.PathValue("feld"), rumpf.Wert); fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{})
}

// stiloptionZuruecksetzen stellt eine Stiloption auf ihre Vorgabe zurück.
//
// **`DELETE` und kein leerer Wert an [Dienst.stiloptionSetzen]**: leer ist bei
// einer Textoption ein gültiger Wert, und zwei Bedeutungen an einem Weg wären
// später nicht mehr auseinanderzuhalten.
func (d *Dienst) stiloptionZuruecksetzen(schreiben http.ResponseWriter, anfrage *http.Request) {
	setID, gut := kennung(schreiben, anfrage, "setID")
	if !gut {
		return
	}

	if fehler := d.naehte.Stile.OptionZuruecksetzen(
		anfrage.Context(), setID, anfrage.PathValue("feld")); fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{})
}
