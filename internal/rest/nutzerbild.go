// **Das Profilbild im Dienst** — Werkstück 8 von Phase 4.12
// (#1351).
//
// # Warum ein Pfad und keine hochgeladene Datei
//
// **Weil der Dienst auf `127.0.0.1` läuft** und derselbe Rechner ist
// ([0046](../../docs/entscheidungen/0046-doch-ein-dienst-auf-der-rueckschleife.md)).
// Ein `multipart/form-data` schöbe dieselben Bytes durch einen zweiten
// Kanal, und die Anwendung schriebe sie wieder auf dieselbe Platte, von der
// sie kommen. Dieselbe Form hat schon
// `PUT /sets/{setID}/karten/{karteID}/felder/{feld}/bild`.
//
// # Und warum die Fassung im Rumpf steht
//
// **Weil sie eine Bedingung ist und keine Adresse.** Der Ort ist das Bild
// eines Nutzers; welchen Stand der Rufer gelesen hat, gehört zur Anfrage.
// Bei `DELETE` gibt es keinen Rumpf — dort steht sie in der Abfrage, wie
// der bestätigende Name beim Löschen eines Nutzers.

package rest

import "net/http"

// nutzerbildSetzen legt eine Bilddatei als Profilbild ab.
func (d *Dienst) nutzerbildSetzen(schreiben http.ResponseWriter, anfrage *http.Request) {
	var rumpf struct {
		Pfad    string `json:"pfad"`
		Fassung string `json:"fassung"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	geaendert, fehler := d.naehte.Nutzer.NutzerbildSetzen(anfrage.Context(),
		anfrage.PathValue("nutzerID"), rumpf.Fassung, rumpf.Pfad)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"nutzer": geaendert})
}

// nutzerbildEntfernen nimmt das Profilbild eines Nutzers wieder weg.
func (d *Dienst) nutzerbildEntfernen(schreiben http.ResponseWriter, anfrage *http.Request) {
	geaendert, fehler := d.naehte.Nutzer.NutzerbildEntfernen(anfrage.Context(),
		anfrage.PathValue("nutzerID"), anfrage.URL.Query().Get("fassung"))
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"nutzer": geaendert})
}

// profilbadgeSetzen legt eine Bilddatei als Abzeichen eines Profils ab.
//
// **`/profile/{id}/badge` und nicht unter dem Nutzer** — Werkstück 9 (#1351):
// ein Abzeichen gehört einem **Profil**, und der Weg sagt es. Dieselbe
// Überlegung wie bei `PUT /profile/{profilID}`.
func (d *Dienst) profilbadgeSetzen(schreiben http.ResponseWriter, anfrage *http.Request) {
	var rumpf struct {
		Pfad    string `json:"pfad"`
		Fassung string `json:"fassung"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	geaendert, fehler := d.naehte.Nutzer.ProfilbadgeSetzen(anfrage.Context(),
		anfrage.PathValue("profilID"), rumpf.Fassung, rumpf.Pfad)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"profil": geaendert})
}

// profilbadgeEntfernen nimmt das Abzeichen eines Profils wieder weg.
func (d *Dienst) profilbadgeEntfernen(
	schreiben http.ResponseWriter, anfrage *http.Request,
) {
	geaendert, fehler := d.naehte.Nutzer.ProfilbadgeEntfernen(anfrage.Context(),
		anfrage.PathValue("profilID"), anfrage.URL.Query().Get("fassung"))
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"profil": geaendert})
}
