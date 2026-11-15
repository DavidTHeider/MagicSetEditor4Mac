// **Der Probelauf über den Dienst** — UXR-F7
// (#2015).
//
// Zwei Wege, die **lesen**:
//
//	POST /sets/{setID}/bogen/probelauf     {"ziel"?}
//	POST /sets/{setID}/vorlagen/probelauf  {"paket", "ziel"?, "wahl"?}
//
// Sie rechnen, was `POST /sets/{setID}/bogen` und `POST /sets/{setID}/vorlagen`
// täten, und schreiben nichts. **`POST`, weil die Wahl in den Rumpf gehört** —
// wie bei `/ersetzung/vorschau`. Das Ziel ist freiwillig: ist es genannt, wird
// es nur angesehen, und liegt dort schon etwas, antwortet der Probelauf mit
// demselben Nein wie der schreibende Weg.
//
// Einordnung nach §16: **erweiternd** — zwei Wege kommen hinzu, keiner ändert
// seine Bedeutung.

package rest

import (
	"context"
	"net/http"
)

// Bogenprobelauf sagt, wie ein Druckbogen würde.
type Bogenprobelauf struct {
	Karten          int `json:"karten"`
	Seiten          int `json:"seiten"`
	Spalten         int `json:"spalten"`
	Zeilen          int `json:"zeilen"`
	JeSeite         int `json:"jeSeite"`
	AufLetzterSeite int `json:"aufLetzterSeite"`
	Bildbreite      int `json:"bildbreite"`

	// Maße in Millimetern.
	Seitenbreite float64 `json:"seitenbreite"`
	Seitenhoehe  float64 `json:"seitenhoehe"`
	Kartenbreite float64 `json:"kartenbreite"`
	Kartenhoehe  float64 `json:"kartenhoehe"`
	Anschnitt    float64 `json:"anschnitt"`
	Rand         float64 `json:"rand"`

	DPI           int  `json:"dpi"`
	Schnittmarken bool `json:"schnittmarken"`
}

// Vorlagenprobelauf sagt, was eine Exportvorlage ergäbe.
type Vorlagenprobelauf struct {
	Karten        int      `json:"karten"`
	Dateiname     string   `json:"dateiname"`
	Bytes         int      `json:"bytes"`
	Beigaben      int      `json:"beigaben"`
	BeigabenBytes int      `json:"beigabenBytes"`
	Verzeichnis   string   `json:"verzeichnis"`
	Geschaetzt    bool     `json:"geschaetzt"`
	Befunde       []string `json:"befunde"`
}

// Probelauf rechnet den Druckbogen und die Exportvorlage durch, ohne zu
// schreiben.
//
// **Eine eigene Naht** und kein Anbau an [Ausfuhr] und [Vorlagenausgabe]: die
// zwei Wege lesen, die beiden anderen schreiben, und eine Naht, die beides
// trägt, verbirgt, wer was darf.
type Probelauf interface {
	BogenProbelauf(ctx context.Context, setID int64, ziel string) (Bogenprobelauf, error)

	VorlageProbelauf(ctx context.Context, setID int64, paket, ziel string,
		wahl map[string]string) (Vorlagenprobelauf, error)
}

// bogenprobelauf rechnet den Druckbogen eines Sets durch.
func (d *Dienst) bogenprobelauf(schreiben http.ResponseWriter, anfrage *http.Request) {
	setID, gut := kennung(schreiben, anfrage, "setID")
	if !gut {
		return
	}

	var rumpf struct {
		Ziel string `json:"ziel"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	bogen, fehler := d.naehte.Probelauf.BogenProbelauf(anfrage.Context(), setID, rumpf.Ziel)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, struct {
		Probelauf bool `json:"probelauf"`
		Bogenprobelauf
	}{true, bogen})
}

// vorlagenprobelauf wertet eine Exportvorlage über einem Set aus.
func (d *Dienst) vorlagenprobelauf(schreiben http.ResponseWriter, anfrage *http.Request) {
	setID, gut := kennung(schreiben, anfrage, "setID")
	if !gut {
		return
	}

	var rumpf struct {
		Paket string            `json:"paket"`
		Ziel  string            `json:"ziel"`
		Wahl  map[string]string `json:"wahl"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	vorlage, fehler := d.naehte.Probelauf.VorlageProbelauf(
		anfrage.Context(), setID, rumpf.Paket, rumpf.Ziel, rumpf.Wahl)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	vorlage.Befunde = ohneLeeretexte(vorlage.Befunde)

	ja(schreiben, http.StatusOK, struct {
		Probelauf bool `json:"probelauf"`
		Vorlagenprobelauf
	}{true, vorlage})
}
