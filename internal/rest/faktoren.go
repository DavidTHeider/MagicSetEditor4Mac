// **Der zweite Faktor über REST** — Werkstück 7 von Phase 4.12
// (#1351).
//
// # Warum hier alles geht und über MCP nicht
//
// **Weil niemand mitliest.** Der Dienst lauscht auf `127.0.0.1`, das
// Geheimnis geht von einem Prozess zum anderen und landet in keinem Verlauf.
// Über MCP stünde es im Mitschnitt eines Sprachmodells — dort fehlen das
// Einrichten und das Bestätigen deshalb, und zwar begründet.
//
// **Der Weg trägt das Geheimnis trotzdem nie**: es steht im **Rumpf** der
// Antwort, und eine Anfragezeile landet in jedem Zugriffsprotokoll.

package rest

import "net/http"

// Faktor ist ein zweiter Faktor.
type Faktor struct {
	Art string `json:"art"`

	// Bezeichnung sagt, welches Gerät gemeint ist — leer erlaubt.
	Bezeichnung string `json:"bezeichnung"`
}

// Faktorstand ist, was ein Nutzer hat und was dieser Rechner kann.
type Faktorstand struct {
	Faktoren []Faktor `json:"faktoren"`

	// Biometrie sagt, ob sich Touch ID oder Face ID hier einrichten ließe.
	Biometrie bool `json:"biometrie"`

	// Biometrieart ist „Touch ID", „Face ID" oder leer.
	Biometrieart string `json:"biometrieart"`

	// Biometriegrund steht da, wenn es nicht geht.
	Biometriegrund string `json:"biometriegrund"`
}

// Faktoreinrichtung ist, was **einmal** zu sehen ist.
type Faktoreinrichtung struct {
	Adresse   string   `json:"adresse"`
	Geheimnis string   `json:"geheimnis"`
	Codes     []string `json:"codes"`
}

// faktoren gibt die zweiten Faktoren eines Nutzers.
func (d *Dienst) faktoren(schreiben http.ResponseWriter, anfrage *http.Request) {
	stand, fehler := d.naehte.Nutzer.Faktoren(
		anfrage.Context(), anfrage.PathValue("nutzerID"))
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	// **Nie `nil`** (§3): eine leere Liste ist eine Antwort.
	if stand.Faktoren == nil {
		stand.Faktoren = []Faktor{}
	}

	ja(schreiben, http.StatusOK, stand)
}

// faktorEinrichten richtet einen zweiten Faktor ein.
//
// **Ein Weg für beide Arten**, weil beide dasselbe bedeuten: „ich will einen
// zweiten Faktor". Was zurückkommt, unterscheidet sich — bei `totp` das
// Muster, bei `biometrie` nichts als ein `201`.
func (d *Dienst) faktorEinrichten(schreiben http.ResponseWriter, anfrage *http.Request) {
	var rumpf struct {
		Art string `json:"art"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	nutzerID := anfrage.PathValue("nutzerID")

	if rumpf.Art == "biometrie" {
		if fehler := d.naehte.Nutzer.BiometrieEinrichten(
			anfrage.Context(), nutzerID); fehler != nil {
			fehlerhaft(schreiben, fehler)

			return
		}

		ja(schreiben, http.StatusNoContent, nil)

		return
	}

	eingerichtet, fehler := d.naehte.Nutzer.FaktorEinrichten(anfrage.Context(), nutzerID)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	if eingerichtet.Codes == nil {
		eingerichtet.Codes = []string{}
	}

	ja(schreiben, http.StatusCreated, eingerichtet)
}

// faktorBestaetigen trägt den TOTP-Faktor ein.
//
// **`PUT` und nicht `POST`**: derselbe Code zweimal führt nicht zum selben
// Zustand — er gilt nur einmal —, aber der **Ort** ist derselbe: dieser eine
// Faktor dieses einen Nutzers.
func (d *Dienst) faktorBestaetigen(schreiben http.ResponseWriter, anfrage *http.Request) {
	var rumpf struct {
		Code string `json:"code"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	if fehler := d.naehte.Nutzer.FaktorBestaetigen(
		anfrage.Context(), anfrage.PathValue("nutzerID"), rumpf.Code); fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusNoContent, nil)
}

// faktorEntfernen nimmt einen zweiten Faktor heraus.
func (d *Dienst) faktorEntfernen(schreiben http.ResponseWriter, anfrage *http.Request) {
	if fehler := d.naehte.Nutzer.FaktorEntfernen(anfrage.Context(),
		anfrage.PathValue("nutzerID"), anfrage.PathValue("art")); fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusNoContent, nil)
}
