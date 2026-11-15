package rest

// **Die KI am Dienst** — Werkstück A6 von Phase 5.3
// (#1445).
//
// **Eine eigene Datei**, weil `preise.go` mit ihr über die Zeilengrenze aus §2
// ginge. Der Schnitt folgt der Frage: dort steht, was etwas **kostet**, hier,
// was ein Modell **kann**.

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// KIKatalogsicht ist der Modellkatalog samt der Regel, die für ihn gilt.
type KIKatalogsicht struct {
	// NurLokal sagt, ob etwas diesen Rechner verlassen darf.
	NurLokal bool `json:"nurLokal"`

	// Einrichtung ist der Stand, der am weitesten kommt — einer der fünf
	// aus `app/kilaufzeit.go` (A7, #1446).
	//
	// **Der beste über alle Anbieter** (A8, #1447): welcher woran hängt,
	// steht je Zeile unter `anbieter`.
	Einrichtung string `json:"einrichtung"`

	// Anbieter sind die Quellen, jede mit ihrem eigenen Stand.
	Anbieter []KIAnbietersicht `json:"anbieter"`

	// Cloud sagt, was über die Auslagerung der **lokalen** Laufzeit
	// feststeht — und was nicht.
	Cloud KICloudsicht `json:"cloud"`

	// Befunde sind die Sätze, die über alle Anbieter hinweg gelten.
	Befunde []string `json:"befunde"`

	Modelle []KIModellsicht `json:"modelle"`

	// Vorrang nennt, wen ein Vorschlag mit Vorrang fragt (#2054).
	Vorrang []string `json:"vorrang"`
}

// KIAnbietersicht ist ein Anbieter samt dem, was über ihn feststeht.
type KIAnbietersicht struct {
	Name        string `json:"name"`
	Einrichtung string `json:"einrichtung"`
	Laufzeit    string `json:"laufzeit"`
	Adresse     string `json:"adresse"`

	// Ort ist `lokal` oder `cloud` — **wo dieser Anbieter rechnet** (#1539).
	// Eine Messung an einem Cloudanbieter kostet Geld, und wer das erst an
	// der Absage merkt, hat schon gerufen.
	Ort string `json:"ort"`

	Befunde []string `json:"befunde"`

	// Messbar sind die Modelle, die er **nennt**, ohne etwas über sie zu
	// sagen (A9, #1448).
	Messbar []string `json:"messbar"`

	// Gemessen sind die Modelle mit einer Messung im Bestand.
	Gemessen []string `json:"gemessen"`

	// SitzungAus sagt, dass er für diese Sitzung abgeschaltet ist — nur
	// Apples Gerätemodell kennt das (AI-05, #2042).
	SitzungAus bool `json:"sitzungAus"`
}

// KICloudsicht ist der Cloudstand samt Fundstelle.
//
// **Drei Stände und nicht zwei** (A7, #1446): „keine Datei" heißt ungewiss,
// denn die zweite dokumentierte Abschaltung steht in der Umgebung eines
// fremden Prozesses.
type KICloudsicht struct {
	Stand string `json:"stand"`
	Beleg string `json:"beleg"`
}

// KIVorschlagssicht ist ein Kartenvorschlag.
type KIVorschlagssicht struct {
	Kennung string `json:"kennung"`
	Modell  string `json:"modell"`

	Felder []KIFeldwertsicht `json:"felder"`

	// Befunde sind die Felder, die das Schema **nicht** durchgelassen hat.
	Befunde []string `json:"befunde"`

	Ablauf string `json:"ablauf"`

	// Uebernommen ist die Zahl der geschriebenen Felder — null, wenn nur
	// vorgeschlagen wurde.
	Uebernommen int `json:"uebernommen"`

	// Anbieter und Weg sagen, wer geantwortet hat und wie er drankam —
	// `gewaehlt`, `vorrang` oder `rueckfall` (AI-05, #2042).
	Anbieter string `json:"anbieter"`
	Weg      string `json:"weg"`

	// Ausgang ist leer, oder `abgelehnt`; Rueckfaelle nennt dann, wer es
	// versuchen darf.
	Ausgang     string   `json:"ausgang,omitempty"`
	Rueckfaelle []string `json:"rueckfaelle,omitempty"`
}

// KIFeldwertsicht ist ein vorgeschlagener Feldwert.
type KIFeldwertsicht struct {
	Feld  string   `json:"feld"`
	Werte []string `json:"werte"`

	// Bisher ist, was **heute** im Feld steht — leer heißt: nichts.
	//
	// **Damit sichtbar ist, was sich ändert** (A11, #1450). Ein Vorschlag,
	// der nur den neuen Wert zeigt, verlangt von jemandem, dass er den alten
	// im Kopf hat — auch an einer Fläche ohne Fenster.
	Bisher string `json:"bisher"`
}

// KIVorschlagsbezug ist, was für einen Vorschlag gebraucht wird.
type KIVorschlagsbezug struct {
	KarteID int64 `json:"karteId"`

	// Anbieter ist die Quelle — `ollama` oder `anthropic` (A8, #1447).
	Anbieter string `json:"anbieter"`

	Alias   string `json:"alias"`
	Auftrag string `json:"auftrag"`

	// Uebernehmen sagt, ob der Vorschlag geschrieben werden soll.
	//
	// **Ohne ihn wird nichts geschrieben** — er ist die ausdrückliche
	// Entscheidung, die am Fenster der zweite Klick ist.
	Uebernehmen bool `json:"uebernehmen"`

	// Felder nennt, **was** übernommen wird — leer heißt: alles, was das
	// Schema durchgelassen hat (A11, #1450).
	//
	// # Warum hier leer »alles« heißt und am Fenster nicht
	//
	// **Weil hier niemand den Vorschlag gesehen hat.** Am Fenster steht er
	// vor jemandem, der je Feld entscheidet; hier ist die Entscheidung
	// `uebernehmen` selbst — und sie ist ausdrücklich. Wer enger wählen
	// will, nennt die Felder.
	Felder []string `json:"felder,omitempty"`

	// Belegfrage geht an den Suchindex — Werkstück A13.4 von Phase 5.3
	// (#1452).
	//
	// **Leer heißt: gar nicht fragen.** Wo sie steht, bekommt das Modell
	// dieselben Belege wie eine Einschätzung — aus demselben Index, mit
	// demselben Manifest —, und die Antwort führt sie unter `belege`.
	Belegfrage string `json:"belegfrage,omitempty"`

	// Freigegeben nennt die **privaten** Korpora, die dabei mitgesucht
	// werden dürfen. Leer heißt: keine.
	Freigegeben []string `json:"freigegeben,omitempty"`

	// Nur nennt, **wonach** gefragt wird — leer heißt: alle Felder. Mit
	// `anbieter: "vorrang"` machen Name und Flavortext eine kleine Aufgabe
	// für Apples Gerätemodell (AI-05, #2042).
	Nur []string `json:"nur,omitempty"`
}

// KIModellsicht ist ein Modell.
type KIModellsicht struct {
	Anbieter string `json:"anbieter"`
	Alias    string `json:"alias"`
	Laufzeit string `json:"laufzeit"`
	Ort      string `json:"ort"`

	Faehigkeiten []KIBelegsicht `json:"faehigkeiten"`
	Herkunft     []KIGliedsicht `json:"herkunft"`
	Bindung      string         `json:"bindung"`

	// Kontextfenster und KontextBeleg wie am Fenster (AI-01.AK-02, #1742):
	// 0 heißt nicht angegeben.
	Kontextfenster int    `json:"kontextfenster"`
	KontextBeleg   string `json:"kontextbeleg"`
}

// KIBelegsicht ist eine belegte Fähigkeit.
type KIBelegsicht struct {
	Faehigkeit string `json:"faehigkeit"`
	Datum      string `json:"datum"`
	Beleg      string `json:"beleg"`
}

// KIGliedsicht ist ein Glied der Herkunftskette.
type KIGliedsicht struct {
	Art      string `json:"art"`
	Kennung  string `json:"kennung"`
	Aus      string `json:"aus"`
	Quelle   string `json:"quelle"`
	Revision string `json:"revision"`
	Lizenz   string `json:"lizenz"`
}

// Anbietersicht ist das Kontingent an einem Anbieterkonto.
type Anbietersicht struct {
	Dienst     string `json:"dienst"`
	Gesamt     int64  `json:"gesamt"`
	Rest       int64  `json:"rest"`
	Verbraucht int64  `json:"verbraucht"`

	// Ueberzugssatz ist Text — Geld je Kredit (0093).
	Ueberzugssatz string `json:"ueberzugssatz"`

	// HoechstensAltMinuten sagt, wie alt die Zahl beim Anbieter sein kann.
	HoechstensAltMinuten int `json:"hoechstensAltMinuten"`
}

// Verlaufsbezug ist, wonach ein Verlaufsbezug fragt.
type Verlaufsbezug struct {
	Quelle    string
	Printing  string
	Finish    string
	Zustaende []string
}

// Verlaufssicht ist, was ein Verlaufsbezug ergeben hat.
type Verlaufssicht struct {
	Quelle     string   `json:"quelle"`
	Fassung    string   `json:"fassung"`
	Zugeordnet bool     `json:"zugeordnet"`
	Staende    int      `json:"staende"`
	Befunde    []string `json:"befunde"`
}

// Marktwunsch ist, wonach ein Marktabruf fragt.
type Marktwunsch struct {
	Quelle    string
	Printing  string
	Sprache   string
	Finish    string
	Zustaende []string
}

// Marktsicht ist, was ein Marktabruf ergeben hat.
type Marktsicht struct {
	Quelle     string          `json:"quelle"`
	Fassung    string          `json:"fassung"`
	Zugeordnet bool            `json:"zugeordnet"`
	Angebote   []Angebotssicht `json:"angebote"`
	Befunde    []string        `json:"befunde"`
}

// Angebotssicht ist ein Posten auf einem fremden Marktplatz.
type Angebotssicht struct {
	Quelle  string `json:"quelle"`
	Kennung string `json:"kennung"`

	// Wert ist der Preis als Text — `"2.17"`. **Nicht als Zahl**: eine
	// Gleitkommazahl im JSON verlöre Cent (0093).
	Wert     string `json:"wert"`
	Waehrung string `json:"waehrung"`

	Menge   int `json:"menge"`
	Buendel int `json:"buendel"`

	Finish  string `json:"finish"`
	Zustand string `json:"zustand"`
	Sprache string `json:"sprache"`

	Signiert   bool `json:"signiert"`
	Veraendert bool `json:"veraendert"`
	Gradiert   bool `json:"gradiert"`
	ImUrlaub   bool `json:"imUrlaub"`

	Verkaeufer string `json:"verkaeufer"`
	Land       string `json:"land"`
}

// Bulkbilanzsicht ist, was aus einem Bulkbezug wurde.
type Bulkbilanzsicht struct {
	Quelle  string   `json:"quelle"`
	Umfang  string   `json:"umfang"`
	Drucke  int      `json:"drucke"`
	Staende int      `json:"staende"`
	Befunde []string `json:"befunde"`
}

// kimodelle nennt, was über KI-Modelle feststeht.
//
// **Erweiternd** (§16) und lesend: der Katalog fragt nichts im Netz.
func (d *Dienst) kimodelle(schreiben http.ResponseWriter, anfrage *http.Request) {
	sicht, fehler := d.naehte.Preise.KIModelle(anfrage.Context())
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, sicht)
}

// kivorschlag lässt ein Modell Felder einer Karte vorschlagen.
//
// **Erweiternd** (§16): ein neuer Weg neben den bestehenden. Er **schreibt
// nur**, wenn der Rumpf es ausdrücklich verlangt — deshalb `POST` und nicht
// `GET`: auch ohne `uebernehmen` fragt er eine Laufzeit und kostet Zeit und
// Strom, und das ist keine Leseoperation im Sinne dieser Fläche.

// KIProbensicht ist, was eine Messung ergeben hat (A9, #1448).
type KIProbensicht struct {
	Proben []KIProbesicht `json:"proben"`
}

// KIProbesicht ist eine gemessene Fähigkeit.
type KIProbesicht struct {
	Anbieter   string `json:"anbieter"`
	Alias      string `json:"alias"`
	Faehigkeit string `json:"faehigkeit"`
	Datum      string `json:"datum"`
	Beleg      string `json:"beleg"`

	// Modell ist, was wirklich geantwortet hat — nicht der Alias.
	Modell string `json:"modell"`
}

// KIProbenbezug nennt, was gemessen werden soll.
type KIProbenbezug struct {
	Anbieter string `json:"anbieter"`
	Alias    string `json:"alias"`
}

// kimodellpruefen misst, was ein Modell kann (A9, #1448).
//
// **`POST` und nicht `GET`**: es kostet Geld und legt eine Aussage im
// Bestand ab. Ein `GET`, der beides tut, wäre ein Weg, den ein
// Zwischenspeicher wiederholen dürfte.
func (d *Dienst) kimodellpruefen(schreiben http.ResponseWriter, anfrage *http.Request) {
	var bezug KIProbenbezug

	if !gelesen(schreiben, anfrage, &bezug) {
		return
	}

	sicht, rufefehler := d.naehte.Preise.KIModellPruefen(anfrage.Context(), bezug)
	if rufefehler != nil {
		fehlerhaft(schreiben, rufefehler)

		return
	}

	ja(schreiben, http.StatusOK, sicht)
}

// kimodellvergessen nimmt die Messungen eines Modells zurück (A9, #1448).
func (d *Dienst) kimodellvergessen(
	schreiben http.ResponseWriter, anfrage *http.Request,
) {
	var bezug KIProbenbezug

	if !gelesen(schreiben, anfrage, &bezug) {
		return
	}

	if rufefehler := d.naehte.Preise.KIModellVergessen(
		anfrage.Context(), bezug); rufefehler != nil {
		fehlerhaft(schreiben, rufefehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"vergessen": true})
}

func (d *Dienst) kivorschlag(schreiben http.ResponseWriter, anfrage *http.Request) {
	kennung, fehler := strconv.ParseInt(anfrage.PathValue("karteID"), 10, 64)
	if fehler != nil {
		fehlerhaft(schreiben, fmt.Errorf("%w: die Kartenkennung %q ist keine Zahl",
			domain.ErrValidation, anfrage.PathValue("karteID")))

		return
	}

	var bezug KIVorschlagsbezug

	if !gelesen(schreiben, anfrage, &bezug) {
		return
	}

	bezug.KarteID = kennung

	sicht, rufefehler := d.naehte.Preise.KIVorschlag(anfrage.Context(), bezug)
	if rufefehler != nil {
		fehlerhaft(schreiben, rufefehler)

		return
	}

	ja(schreiben, http.StatusOK, sicht)
}
