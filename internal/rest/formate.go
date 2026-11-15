// **Formate über den Dienst** — Werkstück 7 von Phase 4.6
// (#1293).
//
// Ein Format ist **Daten** (Vermerk 0061); dieser Dienst liefert keine mit.

package rest

import "net/http"

// formate nennt die Formate — `?spiel=` schränkt ein.
func (d *Dienst) formate(schreiben http.ResponseWriter, anfrage *http.Request) {
	liste, fehler := d.naehte.Decks.Formate(
		anfrage.Context(), anfrage.URL.Query().Get("spiel"))
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, liste)
}

// formatbauarten nennt die drei nachgelesenen Bauarten.
func (d *Dienst) formatbauarten(schreiben http.ResponseWriter, anfrage *http.Request) {
	liste, fehler := d.naehte.Decks.Bauarten(anfrage.Context())
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, liste)
}

// formatAnlegen legt ein Format an.
//
// **Zwei Wege in einem `POST`**: mit `bauart` kommen die nachgelesenen Zahlen
// mit, ohne sie stehen die eigenen im Rumpf.
func (d *Dienst) formatAnlegen(schreiben http.ResponseWriter, anfrage *http.Request) {
	var wunsch map[string]any

	if !gelesen(schreiben, anfrage, &wunsch) {
		return
	}

	// **Dieselbe Vorsicht wie beim Setzen**: ein leerer Rumpf ist nil, und
	// ein Zugriff darauf liest zwar, schreibt aber nicht.
	if wunsch == nil {
		wunsch = map[string]any{}
	}

	if bauart, ist := wunsch["bauart"].(string); ist && bauart != "" {
		d.ausBauart(schreiben, anfrage, wunsch, bauart)

		return
	}

	format, fehler := d.naehte.Decks.FormatSetzen(anfrage.Context(), wunsch)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusCreated, format)
}

// ausBauart legt ein Format mit den Zahlen einer Bauart an.
func (d *Dienst) ausBauart(
	schreiben http.ResponseWriter, anfrage *http.Request,
	wunsch map[string]any, bauart string,
) {
	text := func(name string) string {
		wert, ist := wunsch[name].(string)
		if !ist {
			return ""
		}

		return wert
	}

	format, fehler := d.naehte.Decks.FormatAusBauart(
		anfrage.Context(), bauart, text("name"), text("spiel"), text("scryfallname"))
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusCreated, format)
}

// formatSetzen schreibt ein Format fort.
//
// **`PUT` mit der Kennung im Weg** — sie steht dort und nicht im Rumpf, damit
// zwei Angaben nicht auseinanderlaufen können.
func (d *Dienst) formatSetzen(schreiben http.ResponseWriter, anfrage *http.Request) {
	formatID, gut := kennung(schreiben, anfrage, "formatID")
	if !gut {
		return
	}

	var wunsch map[string]any

	if !gelesen(schreiben, anfrage, &wunsch) {
		return
	}

	// **Ein leerer Rumpf gibt eine nil-Abbildung, und ein Schreibzugriff
	// darauf stürzt ab.** Gefunden von `TestJederEndpunktReichtSeinenFehlerDurch`
	// — genau dafür gibt es ihn.
	if wunsch == nil {
		wunsch = map[string]any{}
	}

	wunsch["id"] = float64(formatID)

	format, fehler := d.naehte.Decks.FormatSetzen(anfrage.Context(), wunsch)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, format)
}

// formatLoeschen nimmt ein Format heraus — die Decks bleiben.
func (d *Dienst) formatLoeschen(schreiben http.ResponseWriter, anfrage *http.Request) {
	formatID, gut := kennung(schreiben, anfrage, "formatID")
	if !gut {
		return
	}

	if fehler := d.naehte.Decks.FormatLoeschen(
		anfrage.Context(), formatID); fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{
		"hinweis": "Decks, die auf diesem Format standen, sind jetzt ungeprüft.",
	})
}
