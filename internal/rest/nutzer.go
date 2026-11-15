// **Nutzer und Profile über REST** — Werkstück 1 von Phase 4.12
// (#1351).

package rest

import "net/http"

// Nutzer ist ein Nutzer samt seinen Profilen, wie der Dienst ihn ausgibt.
type Nutzer struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Bild string `json:"bild"`

	// **Die Angaben aus Werkstück 8** (#1351) — alle freiwillig, dazu die
	// Fassung, die beim Ändern zurückzugeben ist.
	Anrede       string `json:"anrede"`
	Namenszusatz string `json:"namenszusatz"`
	Vorname      string `json:"vorname"`
	Weiterenamen string `json:"weiterenamen"`
	Nachname     string `json:"nachname"`
	Slogan       string `json:"slogan"`
	Kurzebio     string `json:"kurzebio"`
	Langebio     string `json:"langebio"`
	Fassung      string `json:"fassung"`

	// **Woran die Kontoart zu erkennen ist** — Werkstück 6 (#1351).
	//
	// **Das Kennwort steht hier nicht.** Es kommt im Rumpf herein und geht
	// nie zurück — auch nicht gehütet: ein Hash, den jeder abholen kann, ist
	// ein Hash, den jeder offline durchprobieren kann.
	Email   string `json:"email"`
	Telefon string `json:"telefon"`
	Anonym  bool   `json:"anonym"`

	Profile []Profil `json:"profile"`
}

// Profil ist ein Profil.
type Profil struct {
	ID       string `json:"id"`
	NutzerID string `json:"nutzerID"`
	Name     string `json:"name"`

	// **Die freiwilligen Angaben** — Werkstück 9 (#1351), Punkte 10 bis 13.
	Subslogan         string `json:"subslogan"`
	Kurzebeschreibung string `json:"kurzebeschreibung"`
	Langebeschreibung string `json:"langebeschreibung"`
	Badge             string `json:"badge"`

	// **Ein technisches Profil und sein Ablauf** — Werkstück 10 (#1351).
	Technisch bool   `json:"technisch"`
	Ablauf    string `json:"ablauf"`

	// Fassung ist die, die beim Ändern zurückzugeben ist — Werkstück 8
	// (#1351).
	Fassung string `json:"fassung"`
}

// Arbeitskontext sagt, wer gerade in welchem Profil als was arbeitet —
// Werkstück O1 von Phase 5.2 (#1425).
type Arbeitskontext struct {
	NutzerID  string `json:"nutzerID"`
	ProfilID  string `json:"profilID"`
	EignerID  string `json:"eignerID"`
	Rolle     string `json:"rolle"`
	NurLesend bool   `json:"nurLesend"`
	Modus     string `json:"modus"`
	Revision  int64  `json:"revision"`
}

// Profilangaben ist, was ein Aufruf an einem Profil ändern darf.
//
// **Die Kennung steht im Weg und nicht im Rumpf** (`json:"-"`) — dieselbe
// Bauform wie bei [Nutzerangaben]: der Ort ist das Profil.
type Profilangaben struct {
	ProfilID string `json:"-"`
	Name     string `json:"name"`

	Subslogan         string `json:"subslogan"`
	Kurzebeschreibung string `json:"kurzebeschreibung"`
	Langebeschreibung string `json:"langebeschreibung"`
	Badge             string `json:"badge"`

	// **Technisch und Ablauf** — Werkstück 10 (#1351).
	Technisch bool   `json:"technisch"`
	Ablauf    string `json:"ablauf"`

	Fassung string `json:"fassung"`
}

// Nutzerangaben ist, was ein Aufruf an einem Nutzer ändern darf.
type Nutzerangaben struct {
	NutzerID     string `json:"-"`
	Name         string `json:"name"`
	Bild         string `json:"bild"`
	Anrede       string `json:"anrede"`
	Namenszusatz string `json:"namenszusatz"`
	Vorname      string `json:"vorname"`
	Weiterenamen string `json:"weiterenamen"`
	Nachname     string `json:"nachname"`
	Slogan       string `json:"slogan"`
	Kurzebio     string `json:"kurzebio"`
	Langebio     string `json:"langebio"`

	// Fassung ist die zuletzt gelesene.
	//
	// **Ohne sie ginge eine Änderung verloren** (§7, Regel 2 über die
	// Flächen hinweg).
	Fassung string `json:"fassung"`
}

// nutzer gibt alle Nutzer samt ihren Profilen.
func (d *Dienst) nutzer(schreiben http.ResponseWriter, anfrage *http.Request) {
	gelesen, fehler := d.naehte.Nutzer.Nutzer(anfrage.Context())
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	// **Nie `nil`** (§3): eine leere Liste ist eine Antwort, `null` ist keine.
	if gelesen == nil {
		gelesen = []Nutzer{}
	}

	ja(schreiben, http.StatusOK, map[string]any{"nutzer": gelesen})
}

// nutzerAnlegen legt einen Nutzer samt seinem ersten Profil an.
func (d *Dienst) nutzerAnlegen(schreiben http.ResponseWriter, anfrage *http.Request) {
	var rumpf struct {
		Name string `json:"name"`
		Bild string `json:"bild"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	angelegt, fehler := d.naehte.Nutzer.NutzerAnlegen(
		anfrage.Context(), rumpf.Name, rumpf.Bild)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusCreated, map[string]any{"nutzer": angelegt})
}

// profilAnlegen legt ein weiteres Profil für einen Nutzer an.
//
// **`POST /nutzer/{nutzerID}/profile`** — das Profil hängt am Nutzer, und der
// Pfad sagt es. Ein `POST /profile` mit dem Nutzer im Rumpf ließe die
// Zugehörigkeit dort verschwinden, wo sie niemand sieht.
func (d *Dienst) profilAnlegen(schreiben http.ResponseWriter, anfrage *http.Request) {
	var rumpf struct {
		Name string `json:"name"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	angelegt, fehler := d.naehte.Nutzer.ProfilAnlegen(
		anfrage.Context(), anfrage.PathValue("nutzerID"), rumpf.Name)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusCreated, map[string]any{"profil": angelegt})
}

// gezeigtesProfil sagt, in welchem Profil gearbeitet wird.
//
// **Jede andere Antwort dieses Dienstes gilt für dieses Profil** — Sets,
// Decks, Sammlung, Einstellungen. Ohne diese Auskunft wüsste ein Skript
// nicht, wessen Daten es gerade liest.
func (d *Dienst) gezeigtesProfil(schreiben http.ResponseWriter, anfrage *http.Request) {
	gezeigt, fehler := d.naehte.Nutzer.GezeigtesProfil(anfrage.Context())
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"profil": gezeigt})
}

// profilWechseln stellt das Profil um, in dem gearbeitet wird.
//
// **`PUT` und nicht `POST`**: derselbe Ruf zweimal führt zum selben Zustand,
// und beim zweiten Mal wird gar nicht geschrieben
// (`service.Nutzerverwaltung.Wechseln`).
//
// **200 und nicht 201**: es entsteht nichts. Ein Zeiger wird umgestellt.
func (d *Dienst) profilWechseln(schreiben http.ResponseWriter, anfrage *http.Request) {
	var rumpf struct {
		ProfilID string `json:"profilID"`

		// **Der zweite Faktor** — Werkstück 7 (#1351): der Code aus der
		// Authenticator-App oder ein Wiederherstellungscode. **Im Rumpf und
		// nicht im Weg**, wie das Kennwort: eine Anfragezeile landet in jedem
		// Zugriffsprotokoll.
		Nachweis string `json:"nachweis"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	gewechselt, fehler := d.naehte.Nutzer.ProfilWechseln(
		anfrage.Context(), rumpf.ProfilID, rumpf.Nachweis)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"profil": gewechselt})
}

// werArbeitet nennt den Menschen, der gerade arbeitet.
//
// **Ein eigener Weg neben `GET /profil`** — Werkstück 17 von Phase 4.12
// (#1400).
// Jener nennt das Profil samt seinem **Eigner**; wer in ein fremdes gewechselt
// ist, ist das nicht.
func (d *Dienst) werArbeitet(schreiben http.ResponseWriter, anfrage *http.Request) {
	nutzerID, fehler := d.naehte.Nutzer.WerArbeitet(anfrage.Context())
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"nutzerID": nutzerID})
}

// arbeitskontext sagt, wer gerade in welchem Profil als was arbeitet —
// Werkstück O1 von Phase 5.2
// (#1425).
//
// **Ein Zug**: `GET /profil` und `GET /werarbeitet` nacheinander sind zwei
// Stände; hier kommt der Mensch mit seinem Profil, seiner Rolle und einer
// Revision, an der ein Stand vom nächsten zu unterscheiden ist.
func (d *Dienst) arbeitskontext(schreiben http.ResponseWriter, anfrage *http.Request) {
	kontext, fehler := d.naehte.Nutzer.Arbeitskontext(anfrage.Context())
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"kontext": kontext})
}

// nutzerWechseln stellt um, wer arbeitet.
//
// **`PUT /nutzer` und nicht `PUT /nutzer/{id}`** — derselbe Grund wie bei
// `PUT /profil`: der Weg benennt den **Zeiger**, nicht den Nutzer.
// `PUT /nutzer/{id}` läse sich wie „ändere diesen Nutzer", und das ist ein
// anderer Weg, den es schon gibt.
//
// **200 und nicht 201**: es entsteht nichts.
func (d *Dienst) nutzerWechseln(schreiben http.ResponseWriter, anfrage *http.Request) {
	var rumpf struct {
		NutzerID string `json:"nutzerID"`

		// **Der zweite Faktor DIESES Nutzers** — nicht der eigene. Im Rumpf
		// und nicht im Weg, wie überall: eine Anfragezeile landet in jedem
		// Zugriffsprotokoll.
		Nachweis string `json:"nachweis"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	gewechselt, fehler := d.naehte.Nutzer.NutzerWechseln(
		anfrage.Context(), rumpf.NutzerID, rumpf.Nachweis)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"profil": gewechselt})
}

// anbieterkonten gibt die Anbieterkonten eines Nutzers.
func (d *Dienst) anbieterkonten(schreiben http.ResponseWriter, anfrage *http.Request) {
	gelesen, fehler := d.naehte.Nutzer.Anbieterkonten(
		anfrage.Context(), anfrage.PathValue("nutzerID"))
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	// **Nie `nil`** (§3): eine leere Liste ist eine Antwort, `null` ist keine.
	if gelesen == nil {
		gelesen = []Konto{}
	}

	ja(schreiben, http.StatusOK, map[string]any{"konten": gelesen})
}

// kontoLoesen nimmt eine fremde Kennung von einem Nutzer.
//
// **`DELETE` und der Anbieter im Weg**: gelöscht wird eine benannte
// Verbindung, und ein `DELETE` hat keinen Rumpf — dieselbe Form wie beim
// Löschen eines Sets (#1218).
//
// **204 und kein Rumpf**: es bleibt nichts zu zeigen.
func (d *Dienst) kontoLoesen(schreiben http.ResponseWriter, anfrage *http.Request) {
	if fehler := d.naehte.Nutzer.KontoLoesen(anfrage.Context(),
		anfrage.PathValue("nutzerID"), anfrage.PathValue("anbieter")); fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	schreiben.WriteHeader(http.StatusNoContent)
}
