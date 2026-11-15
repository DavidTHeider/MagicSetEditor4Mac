// **Die Stände der bezogenen Bäume** — G04
// (#2251).
//
// Dieselbe Auskunft und dieselben Schritte wie im Fenster (§16). Einordnung
// nach §16: **erweiternd** — vier Wege kommen hinzu, keiner ändert seine
// Bedeutung.

package rest

import "net/http"

// paketBezuege nennt die bezogenen Bäume — mit `?fragen=1` samt dem neuesten
// Stand ihrer Quelle. **Ohne die Frage geht nichts ins Netz.**
func (d *Dienst) paketBezuege(schreiben http.ResponseWriter, anfrage *http.Request) {
	bezuege, fehler := d.naehte.Pakete.PaketBezuege(anfrage.Context(), anfrage.URL.Query().Get("fragen") == "1")
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"bezuege": ohneNil(bezuege)})
}

// paketBezugVorschau sagt, was eine Aktualisierung auf `?stand=` täte.
func (d *Dienst) paketBezugVorschau(schreiben http.ResponseWriter, anfrage *http.Request) {
	vorschau, fehler := d.naehte.Pakete.PaketBezugVorschau(
		anfrage.Context(), anfrage.PathValue("name"), anfrage.URL.Query().Get("stand"))
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"vorschau": vorschau})
}

// paketBezugAktualisieren bezieht einen Commit an die Stelle eines Baums.
func (d *Dienst) paketBezugAktualisieren(schreiben http.ResponseWriter, anfrage *http.Request) {
	var rumpf struct {
		Stand      string `json:"stand"`
		Bestaetigt bool   `json:"bestaetigt"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	bezug, fehler := d.naehte.Pakete.PaketBezugAktualisieren(
		anfrage.Context(), anfrage.PathValue("name"), rumpf.Stand, rumpf.Bestaetigt)
	if fehler != nil {
		fehlerhaftMit(schreiben, fehler, teilbezug(bezug))

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"bezug": bezug})
}

// paketBezugZurueck tauscht einen Baum mit seinem aufgehobenen Stand.
func (d *Dienst) paketBezugZurueck(schreiben http.ResponseWriter, anfrage *http.Request) {
	bezug, fehler := d.naehte.Pakete.PaketBezugZurueck(anfrage.Context(), anfrage.PathValue("name"))
	if fehler != nil {
		fehlerhaftMit(schreiben, fehler, teilbezug(bezug))

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"bezug": bezug})
}

// teilbezug nennt einen Baum, der trotz Fehler liegt — ein gescheitertes
// Vermerken nimmt den Tausch nicht zurück (§7, Regel 3) —, oder nichts.
func teilbezug(bezug Bezug) any {
	if bezug.Ort == "" {
		return nil
	}

	return map[string]any{"bezug": bezug}
}
