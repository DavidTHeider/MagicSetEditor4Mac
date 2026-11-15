package rest

// **Ein Handel im Dienst** —
// #1637.
//
// # Warum der Schritt derselbe Weg bleibt
//
// **Weil der Handel die Ressource ist und der Schritt sein Zustand** — der
// Satz steht seit `T6` im Kopf von `handelsverbuchung.go`, und er trägt
// weiter: `POST /handel/{vorgang}/schritt` nimmt jetzt **zwölf** Schritte
// statt drei. Nach [0033](../../docs/entscheidungen/0033-was-ueber-fassungen-hinweg-zugesagt-wird.md)
// ist das **erweiternd**: ein Feld nimmt mehr Werte an, keiner fällt weg.
//
// # Und warum der Vorschlag einen eigenen Weg hat
//
// **Weil er kein Schritt ist.** Ein Schritt bewegt den Zustand; ein Vorschlag
// legt eine **neue Revision** an und macht beide Zustimmungen ungültig. Wer
// beides unter denselben Rumpf legte, hätte ein Feld `positionen`, das bei
// elf von zwölf Schritten leer bliebe.

import (
	"context"
	"net/http"
)

// Leitungsnaht ist, was der Dienst für die Schritte vor dem Abschluss
// braucht.
type Leitungsnaht interface {
	HandelEroeffnen(ctx context.Context, wunsch Handelswunsch) (
		Handelsvorgang, error)
	HandelVorschlagen(ctx context.Context, vorgang string,
		positionen []Handelsposition) (Handelsvorgang, error)
	HandelStand(ctx context.Context, vorgang string) (Handelsvorgang, error)

	// HandelProtokoll gibt die Zustandswechsel (TRADE-03.AK-03, #1717).
	HandelProtokoll(ctx context.Context, vorgang string) ([]Handelsprotokollzeile, error)

	// HandelPreisanhalt gibt, was die Seiten ungefähr wert sind — in der
	// Gestalt der App-Naht (TRADE-02b, #1951).
	HandelPreisanhalt(ctx context.Context, vorgang string) (any, error)

	// HandelVorgaenge nennt die Handel mit einem Gegenüber — UX-08.AK-04
	// (#1680).
	HandelVorgaenge(ctx context.Context, gegenueber string) (
		[]Handelsvorgang, error)

	HandelSchicken(ctx context.Context, vorgang, empfaenger string) (
		string, error)
	HandelPostbestaetigung(ctx context.Context, nachrichtID string) error
	HandelWiederaufnahme(ctx context.Context, vorgang, empfaenger string,
		seiner Handelsfortschritt) (Wiederaufnahme, error)

	// HandelFesthalten hält eine benannte Besitzzeile fest
	// (#1643).
	HandelFesthalten(ctx context.Context, vorgang string, exemplarID int64,
		menge int, fassung int64) (Festgehalten, error)

	// Handelszusagen gibt, was für einen Vorgang festgehalten ist.
	Handelszusagen(ctx context.Context, vorgang string) ([]Handelszusage, error)

	// Liste und Entwürfe (TRADE-01, #1715).
	HandelUebersicht(ctx context.Context, filter Handelsfilter) ([]Handelsuebersichtszeile, error)
	HandelEntwurfAendern(ctx context.Context, wunsch Handelswunsch) (Handelsvorgang, error)
	HandelKopieren(ctx context.Context, vorgang string) (Handelsvorgang, error)
	HandelEntwurfLoeschen(ctx context.Context, vorgang string) error
	HandelArchivieren(ctx context.Context, vorgang string, an bool) error
}

// Handelszusage ist eine festgehaltene Besitzzeile.
type Handelszusage struct {
	ExemplarID int64  `json:"exemplarId"`
	Menge      int    `json:"menge"`
	Fassung    int64  `json:"fassung"`
	Revision   int    `json:"revision"`
	Frist      string `json:"frist,omitempty"`
	Uebergeben bool   `json:"uebergeben"`
}

// Festgehalten ist die Antwort der Naht auf ein Festhalten: die Liste —
// und der Grund, aus dem sie nach geglücktem Schreiben fehlen kann
// (§7, Regel 3; #1970).
type Festgehalten struct {
	Zusagen        []Handelszusage
	Nachladefehler *Befund
}

// Nachladekopf trägt den Code eines Nachladefehlers, **wo der Rumpf eine
// Liste ist** und kein Feld daneben hat (#1970).
//
// **Im Kopf und nicht im Rumpf**, aus demselben Grund wie [Vorgangskopf]:
// ein Feld verlangte eine andere Gestalt der Antwort, und die wäre brechend
// (0033). Ein Kopf ist erweiternd — wer ihn nicht kennt, liest wie bisher.
const Nachladekopf = "X-Nachladefehler"

// Zusagenwunsch ist, was im Rumpf eines Festhaltens steht.
type Zusagenwunsch struct {
	ExemplarID int64 `json:"exemplarId"`
	Menge      int   `json:"menge"`
	Fassung    int64 `json:"fassung"`
}

// Handelsposition ist eine Zeile eines Angebots.
type Handelsposition struct {
	Von      string `json:"von"`
	Printing string `json:"printing"`
	DruckID  int64  `json:"druckId,omitempty"`
	Menge    int    `json:"menge"`

	Guete   string `json:"guete"`
	Sprache string `json:"sprache"`
	Finish  string `json:"finish"`

	Abbild    string `json:"abbild,omitempty"`
	Bewertung int64  `json:"bewertung,omitempty"`
}

// Handelsvorgang ist **eine** Revision.
type Handelsvorgang struct {
	Vorgang  string `json:"vorgang"`
	Revision int    `json:"revision"`
	Abdruck  string `json:"abdruck"`
	Eltern   string `json:"eltern,omitempty"`

	Art       string   `json:"art"`
	Initiator string   `json:"initiator"`
	Parteien  []string `json:"parteien"`

	// Ich ist die Partei dieses Profils — die, deren Stimme Naechster nennt
	// (#2026). Leer, wo es keine Identität gibt.
	Ich     string `json:"ich"`
	Zustand string `json:"zustand"`

	Positionen       []Handelsposition `json:"positionen"`
	BewertungGeteilt bool              `json:"bewertungGeteilt"`
	Zustimmungen     []string          `json:"zustimmungen"`
	Ende             bool              `json:"ende"`

	// Beschluss ist der Schritt, dem die Stimmen gelten (TRADE-03); Lage,
	// Naechster, Moegliche und Phase sagen, wo er für dieses Profil steht
	// und was es jetzt tun darf (TRADE-02, #1716).
	Beschluss string   `json:"beschluss"`
	Lage      string   `json:"lage"`
	Naechster string   `json:"naechster"`
	Moegliche []string `json:"moegliche"`
	Phase     string   `json:"phase"`
}

// Handelswunsch ist, womit ein Vorgang eröffnet wird.
//
// **Ohne Initiator** — er kommt aus dem Ausweis dieses Rechners. Wer ihn
// nennen dürfte, eröffnete einen Handel im Namen eines anderen.
type Handelswunsch struct {
	Vorgang          string            `json:"vorgang"`
	Art              string            `json:"art"`
	Gegenueber       []string          `json:"gegenueber"`
	Positionen       []Handelsposition `json:"positionen"`
	BewertungGeteilt bool              `json:"bewertungGeteilt"`
}

// Vorschlagswunsch ist, was im Rumpf eines Vorschlags steht.
type Vorschlagswunsch struct {
	Positionen []Handelsposition `json:"positionen"`
}

// Versandwunsch ist, was im Rumpf eines Versands steht.
type Versandwunsch struct {
	// Empfaenger ist die Peerkennung des Gegenübers.
	Empfaenger string `json:"empfaenger"`
}

// Handelsfortschritt ist, was eine Seite als bestätigt kennt.
type Handelsfortschritt struct {
	Empfaenger string `json:"empfaenger"`
	Revision   int    `json:"revision"`
	Abdruck    string `json:"abdruck"`
	Folge      int64  `json:"folge"`
}

// Nachlieferung ist eine Nachricht, die noch einmal hinausmuss.
type Nachlieferung struct {
	NachrichtID string `json:"nachrichtId"`
	Vorgang     string `json:"vorgang"`
	Empfaenger  string `json:"empfaenger"`
	Folge       int64  `json:"folge"`
	Art         string `json:"art"`
}

// Wiederaufnahme sagt, was nach einem Abriss fehlt.
type Wiederaufnahme struct {
	Nachzuliefern []Nachlieferung `json:"nachzuliefern"`

	// Klaerung heißt: zwei Verläufe stehen da, und das lösen zwei Menschen.
	Klaerung bool `json:"klaerung"`
}

// Versandbeleg nennt die Nachricht, die nun im Ausgang liegt.
type Versandbeleg struct {
	NachrichtID string `json:"nachrichtId"`
}

// handelEroeffnen legt einen Vorgang an.
func (d *Dienst) handelEroeffnen(
	schreiben http.ResponseWriter, anfrage *http.Request,
) {
	var wunsch Handelswunsch

	if !gelesen(schreiben, anfrage, &wunsch) {
		return
	}

	vorgang, fehler := d.naehte.Leitung.HandelEroeffnen(
		anfrage.Context(), wunsch)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusCreated, vorgang)
}

// handelVorschlagen trägt die nächste Revision ein.
func (d *Dienst) handelVorschlagen(
	schreiben http.ResponseWriter, anfrage *http.Request,
) {
	var wunsch Vorschlagswunsch

	if !gelesen(schreiben, anfrage, &wunsch) {
		return
	}

	vorgang, fehler := d.naehte.Leitung.HandelVorschlagen(anfrage.Context(),
		anfrage.PathValue("vorgang"), wunsch.Positionen)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, vorgang)
}

// handelsstandZeigen gibt die letzte Revision eines Vorgangs.
func (d *Dienst) handelsstandZeigen(
	schreiben http.ResponseWriter, anfrage *http.Request,
) {
	vorgang, fehler := d.naehte.Leitung.HandelStand(anfrage.Context(),
		anfrage.PathValue("vorgang"))
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, vorgang)
}

// handelsvorgaengeZeigen nennt die Handel mit einem Gegenüber — je Vorgang
// die letzte Revision, zuletzt geänderte zuerst (UX-08.AK-04).
func (d *Dienst) handelsvorgaengeZeigen(
	schreiben http.ResponseWriter, anfrage *http.Request,
) {
	vorgaenge, fehler := d.naehte.Leitung.HandelVorgaenge(anfrage.Context(),
		anfrage.PathValue("kennung"))
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"vorgaenge": ohneNil(vorgaenge)})
}

// handelVersenden legt den heutigen Stand in den Ausgang.
func (d *Dienst) handelVersenden(
	schreiben http.ResponseWriter, anfrage *http.Request,
) {
	var wunsch Versandwunsch

	if !gelesen(schreiben, anfrage, &wunsch) {
		return
	}

	nachrichtID, fehler := d.naehte.Leitung.HandelSchicken(anfrage.Context(),
		anfrage.PathValue("vorgang"), wunsch.Empfaenger)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusCreated, Versandbeleg{NachrichtID: nachrichtID})
}

// handelspostBestaetigen räumt eine Nachricht weg, die angekommen ist.
func (d *Dienst) handelspostBestaetigen(
	schreiben http.ResponseWriter, anfrage *http.Request,
) {
	nachrichtID := anfrage.PathValue("nachrichtId")

	if fehler := d.naehte.Leitung.HandelPostbestaetigung(
		anfrage.Context(), nachrichtID); fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, Versandbeleg{NachrichtID: nachrichtID})
}

// handelWiederaufnehmen sagt, was nach einem Abriss noch einmal hinausmuss.
func (d *Dienst) handelWiederaufnehmen(
	schreiben http.ResponseWriter, anfrage *http.Request,
) {
	var seiner Handelsfortschritt

	if !gelesen(schreiben, anfrage, &seiner) {
		return
	}

	offen, fehler := d.naehte.Leitung.HandelWiederaufnahme(anfrage.Context(),
		anfrage.PathValue("vorgang"), seiner.Empfaenger, seiner)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, offen)
}

// handelFesthalten hält eine benannte Besitzzeile für einen Vorgang fest.
func (d *Dienst) handelFesthalten(
	schreiben http.ResponseWriter, anfrage *http.Request,
) {
	var wunsch Zusagenwunsch

	if !gelesen(schreiben, anfrage, &wunsch) {
		return
	}

	festgehalten, fehler := d.naehte.Leitung.HandelFesthalten(anfrage.Context(),
		anfrage.PathValue("vorgang"), wunsch.ExemplarID, wunsch.Menge,
		wunsch.Fassung)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	// **Festgehalten ist festgehalten** (§7, Regel 3; #1970): 201, und der
	// Kopf sagt, warum die Liste leer ist.
	if nachladen := festgehalten.Nachladefehler; nachladen != nil {
		schreiben.Header().Set(Nachladekopf, nachladen.Code)
	}

	ja(schreiben, http.StatusCreated, festgehalten.Zusagen)
}

// handelszusagenZeigen gibt, was für einen Vorgang festgehalten ist.
func (d *Dienst) handelszusagenZeigen(
	schreiben http.ResponseWriter, anfrage *http.Request,
) {
	zusagen, fehler := d.naehte.Leitung.Handelszusagen(
		anfrage.Context(), anfrage.PathValue("vorgang"))
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, zusagen)
}
