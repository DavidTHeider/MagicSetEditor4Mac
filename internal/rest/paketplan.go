package rest

// **Plan, Ersetzen und Zurück** — DOM-01.AK-02 und AK-03
// (#1696).
//
// **Eine eigene Datei**, weil `paketpflege.go` einbindet, anlegt, übernimmt
// und entfernt; hier steht, was **vor** einer Änderung zu sehen ist und wie
// eine Fassung gegen eine andere getauscht wird.

import "net/http"

// Paketplan ist, was eine Paketänderung bewirken würde.
//
// **Listen ohne `omitempty`** (§3): „kein Set" und „kein Konflikt" sind
// Aussagen, `null` ist keine.
type Paketplan struct {
	Vorgang   string          `json:"vorgang"`
	Paket     string          `json:"paket"`
	Quelle    string          `json:"quelle"`
	Von       string          `json:"von"`
	Zu        string          `json:"zu"`
	Geaendert bool            `json:"geaendert"`
	Moeglich  bool            `json:"moeglich"`
	Sets      []string        `json:"sets"`
	Konflikte []Paketkonflikt `json:"konflikte"`
}

// Paketkonflikt ist ein Bedarf, der nach der Änderung nicht mehr erfüllt
// wäre.
type Paketkonflikt struct {
	Paket    string `json:"paket"`
	Braucht  string `json:"braucht"`
	Art      string `json:"art"`
	Verlangt string `json:"verlangt"`
	Da       string `json:"da"`
}

// paketPlan sagt vor einer Änderung, was sie bewirkt.
//
// **Ein `GET` mit Abfrage**: der Plan ändert nichts, und ein Rumpf an einem
// `GET` wäre eine Zusage, die kein Zwischenstück hält. Einordnung nach §16:
// **erweiternd**.
func (d *Dienst) paketPlan(schreiben http.ResponseWriter, anfrage *http.Request) {
	abfrage := anfrage.URL.Query()

	plan, fehler := d.naehte.Pakete.PaketPlan(anfrage.Context(), abfrage.Get("vorgang"),
		abfrage.Get("art"), abfrage.Get("name"), abfrage.Get("quelle"))
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"plan": plan})
}

// paketErsetzen legt eine neue Fassung über ein eingebundenes Paket.
//
// **`PUT /pakete/ersetzt` und nicht `PUT /pakete/eingebunden`** — jener
// weist ein vorhandenes Paket ab, und das ist seine Zusage. Wer ersetzt,
// sagt es mit der Adresse.
func (d *Dienst) paketErsetzen(schreiben http.ResponseWriter, anfrage *http.Request) {
	var rumpf struct {
		Quelle     string `json:"quelle"`
		Bestaetigt bool   `json:"bestaetigt"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	paket, fehler := d.naehte.Pakete.PaketErsetzen(anfrage.Context(), rumpf.Quelle, rumpf.Bestaetigt)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"paket": paket})
}

// paketZurueck tauscht die aktuelle Fassung gegen die vorige.
func (d *Dienst) paketZurueck(schreiben http.ResponseWriter, anfrage *http.Request) {
	paket, fehler := d.naehte.Pakete.PaketZurueck(
		anfrage.Context(), anfrage.PathValue("art"), anfrage.PathValue("name"))
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"paket": paket})
}

// paketBedarf nennt, was einem Set an Paketen fehlt oder zu alt ist.
//
// **Am Set**: der Bedarf ist eine Frage an ein Set, nicht an den Katalog —
// `GET /pakete` und der Plan fragen den Katalog (DOM-01.AK-04). Einordnung
// nach §16: **erweiternd**.
func (d *Dienst) paketBedarf(schreiben http.ResponseWriter, anfrage *http.Request) {
	setID, gut := kennung(schreiben, anfrage, "setID")
	if !gut {
		return
	}

	befunde, fehler := d.naehte.Pakete.PaketBedarf(anfrage.Context(), setID)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"befunde": ohneNil(befunde)})
}
