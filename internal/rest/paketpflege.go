// **Pakete einbinden, entfernen und aus MSE2 übernehmen** —
// #1291.
//
// **Eine eigene Datei**, weil `bestandswege.go` sonst über die Zeilengrenze
// liefe (§2).

package rest

import "net/http"

// paketInstallieren bindet ein Paket ein, das schon auf der Platte liegt.
//
// **`PUT /pakete/eingebunden` und nicht `POST /pakete`** — jener **bezieht**
// aus dem Netz. Zwei Fähigkeiten hinter einer Adresse wären eine Zusage, die
// niemand lesen kann (#891).
//
// Einordnung nach §16: **erweiternd** — ein Weg kommt hinzu, keiner ändert
// seine Bedeutung.
func (d *Dienst) paketInstallieren(schreiben http.ResponseWriter, anfrage *http.Request) {
	var rumpf struct {
		Quelle string `json:"quelle"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	paket, fehler := d.naehte.Pakete.PaketInstallieren(anfrage.Context(), rumpf.Quelle)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusCreated, map[string]any{"paket": paket})
}

// paketAnlegen legt ein neues Paket aus einem Gerüst an (#1350).
//
// **`POST /pakete/neu` und nicht `POST /pakete`** — jener **bezieht** aus dem
// Netz, dieser **schreibt** eines auf die Platte. Zwei Fähigkeiten unter einem
// Verb wären zwei Verträge in einem Pfad; derselbe Grund, aus dem
// `PUT /pakete/eingebunden` eine eigene Adresse hat.
func (d *Dienst) paketAnlegen(schreiben http.ResponseWriter, anfrage *http.Request) {
	var rumpf struct {
		Art      string `json:"art"`
		Name     string `json:"name"`
		Vollname string `json:"vollname"`
		Spiel    string `json:"spiel"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	paket, fehler := d.naehte.Pakete.PaketAnlegen(
		anfrage.Context(), rumpf.Art, rumpf.Name, rumpf.Vollname, rumpf.Spiel)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusCreated, map[string]any{"paket": paket})
}

// paketEntfernen nimmt ein Paket von der Platte und aus dem Katalog.
//
// **Der Pfad steht in der Abfrage**, wie der Name beim Löschen eines Sets: ein
// `DELETE` hat keinen Rumpf (#1218). Er ist die **Bestätigung** — dieselbe
// Wache wie über MCP, und aus demselben Grund: der Dienst wird über die
// Leitung angesprochen, und wer ruft, sieht seinen Aufruf nicht.
func (d *Dienst) paketEntfernen(schreiben http.ResponseWriter, anfrage *http.Request) {
	art := anfrage.PathValue("art")
	name := anfrage.PathValue("name")

	// **Ein leerer Pfad wird durchgereicht und nicht hier abgewiesen** —
	// dieselbe Überlegung wie beim Namen eines gelöschten Sets (#1218): dass
	// ein leerer nie passt, ist die Antwort der Naht und nicht die dieser
	// Fläche. Eine zweite Prüfung hier gäbe eine zweite Meldung für dieselbe
	// Sache.
	pfad := anfrage.URL.Query().Get("pfad")

	if fehler := d.naehte.Pakete.PaketEntfernen(
		anfrage.Context(), art, name, pfad); fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"entfernt": art + "/" + name})
}

// paketUebernehmen holt alle Vorlagenpakete aus einem Verzeichnis.
func (d *Dienst) paketUebernehmen(schreiben http.ResponseWriter, anfrage *http.Request) {
	var rumpf struct {
		Ort string `json:"ort"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	uebernahme, fehler := d.naehte.Pakete.PaketUebernehmen(anfrage.Context(), rumpf.Ort)

	// **Der Teilerfolg geht mit dem Fehler hinaus** — dieselbe Klasse wie beim
	// MTG-JSON-Einlesen (Befund 1 des Zweitblicks über Phase 4.4): die Pakete
	// **liegen** schon im Verzeichnis, wenn das Auffrischen des Katalogs
	// scheitert (§7 Regel 3, #1041).
	if fehler != nil {
		fehlerhaftMit(schreiben, fehler, uebernahme)

		return
	}

	ja(schreiben, http.StatusCreated, uebernahme)
}

// mse2Orte nennt die üblichen Orte einer MSE2-Installation.
//
// **Vorgeschlagen, nicht gelesen** (#55): was dort liegt, erfährt die
// Anwendung erst über `POST /pakete/uebernommen`.
func (d *Dienst) mse2Orte(schreiben http.ResponseWriter, _ *http.Request) {
	orte := d.naehte.Pakete.MSE2Orte()
	if orte == nil {
		orte = []string{}
	}

	ja(schreiben, http.StatusOK, map[string]any{"orte": orte})
}

// kartenbildSetzen liest eine Bilddatei und schreibt sie in ein Kartenfeld.
//
// **`PUT` und nicht `POST`**: es ersetzt den Wert eines benannten Feldes —
// derselbe Zug wie `PUT /sets/{setID}/karten/{karteID}/felder/{feld}`, nur mit
// einer Datei statt einem Text.
//
// Einordnung nach §16: **erweiternd** — ein Weg kommt hinzu, keiner ändert
// seine Bedeutung.
func (d *Dienst) kartenbildSetzen(schreiben http.ResponseWriter, anfrage *http.Request) {
	setID, gut := kennung(schreiben, anfrage, "setID")
	if !gut {
		return
	}

	karteID, gut := kennung(schreiben, anfrage, "karteID")
	if !gut {
		return
	}

	var rumpf struct {
		Pfad string `json:"pfad"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	masse, fehler := d.naehte.Kartenbilder.AusDatei(
		anfrage.Context(), setID, karteID, anfrage.PathValue("feld"), rumpf.Pfad)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, masse)
}
