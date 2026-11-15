// **Die Preise über die Rückschleife** — Werkstück 3 von Phase 3.4
// (#1165).
//
// # Der Vorbehalt reist in jeder Antwort mit
//
// Jeder Stand trägt **Quelle, Zeitpunkt und Alter in Stunden**. Die Quelle
// sagt über ihre Preise ausdrücklich, dass sie für nichts einsteht — eine Zahl
// ohne diese drei wäre eine Behauptung, für die niemand geradesteht.

package rest

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Preiswerkstatt ist, was der Dienst für die Preise braucht.
type Preiswerkstatt interface {
	PreiseHolen(ctx context.Context, printing string) ([]Preissicht, []string, error)
	Preise(ctx context.Context, printing string) ([]Preissicht, error)
	Preisverlauf(ctx context.Context, reihe Preisreihe) ([]Preissicht, error)
	Portfolio(ctx context.Context) ([]Portfoliozeilensicht, int, error)

	// Preisbulk holt eine ganze Preisdatei einer Quelle (A3, #1442).
	Preisbulk(ctx context.Context, quelle, umfang string) (Bulkbilanzsicht, error)

	// Marktangebote gibt die Posten eines Marktplatzes zu einem Druck
	// (A4, #1443).
	Marktangebote(ctx context.Context, wunsch Marktwunsch) (Marktsicht, error)

	// Anbieterstand nennt das Kontingent an einem Anbieterkonto (A5, #1444).
	Anbieterstand(ctx context.Context, dienst string) (Anbietersicht, error)

	// Fremdpreisverlauf holt den Verlauf eines Drucks bei einer Quelle
	// (A5, #1444).
	Fremdpreisverlauf(ctx context.Context, wunsch Verlaufsbezug) (Verlaufssicht, error)

	// KIModelle nennt, was über KI-Modelle feststeht (A6, #1445).
	KIModelle(ctx context.Context) (KIKatalogsicht, error)

	// KIVorschlag lässt ein Modell Felder einer Karte vorschlagen (A7, #1446).
	// KIBatch lässt ein Modell für mehrere Karten vorschlagen (A13.1, #1452).
	//
	// **Es schreibt an keiner Karte** — übernommen wird je Vorschlag.
	KIBatch(ctx context.Context, bezug KIBatchbezug) (KIBatchsicht, error)

	// KIGruppe übernimmt mehrere Vorschläge **als eine Gruppe** —
	// Werkstück A13.1b (#1452). Ganz oder gar nicht.
	KIGruppe(ctx context.Context, bezug KIGruppenbezug) (KIGruppensicht, error)

	// KIErkennung liest ein Kartenfoto — Werkstück A13.3 (#1452).
	KIErkennung(ctx context.Context, bezug KIErkennungsbezug) (any, error)

	// KIEinschaetzung lässt ein Modell eine Karte beurteilen (A13.2a, #1452).
	//
	// **Die Antwort ist `any`** — die Gestalt steht in `internal/app`.
	KIEinschaetzung(ctx context.Context, bezug KIEinschaetzungsbezug) (any, error)

	KIVorschlag(ctx context.Context, bezug KIVorschlagsbezug) (
		KIVorschlagssicht, error)

	// KIModellPruefen misst, was ein Modell kann (A9, #1448).
	KIModellPruefen(ctx context.Context, bezug KIProbenbezug) (
		KIProbensicht, error)

	// KIModellVergessen nimmt die Messungen eines Modells zurück (A9, #1448).
	KIModellVergessen(ctx context.Context, bezug KIProbenbezug) error

	// KIGeraetemodellFuerSitzung schaltet Apples Gerätemodell für diesen
	// Dienst ab oder an (AI-05, #2042) und gibt den neuen Stand.
	KIGeraetemodellFuerSitzung(ctx context.Context, an bool) (bool, error)
}

// Preissicht ist ein Preisstand, wie diese Fläche ihn sieht.
type Preissicht struct {
	Printing string `json:"printing"`

	// DruckID ist die Zeile des Drucks in diesem Bestand (A2, #1441) — als
	// Zeichenkette, leer für keine. Erweiternd.
	DruckID string `json:"druckID"`

	// **Die sieben Teile der Reihe** (A3, #1442) — `markt`, `art`, `zustand`
	// und `sprache` sind **erweiternd** dazugekommen.
	Quelle   string `json:"quelle"`
	Markt    string `json:"markt"`
	Art      string `json:"art"`
	Waehrung string `json:"waehrung"`
	Finish   string `json:"finish"`
	Zustand  string `json:"zustand"`
	Sprache  string `json:"sprache"`

	// Lage sagt, was feststeht: `genannt`, `weggefallen` oder `ungeprueft`.
	Lage string `json:"lage"`

	// Wert ist leer, wo die Lage nicht `genannt` ist.
	Wert string `json:"wert"`

	// Beobachtet ist der Tag, den die Quelle nennt — leer, wo sie keinen nennt.
	Beobachtet   string `json:"beobachtet"`
	Erhoben      string `json:"erhoben"`
	AlterStunden int    `json:"alterStunden"`
}

// Preisreihe benennt, welche Reihe gemeint ist.
//
// **Ohne Vorgaben** (A3, #1442): „wenn nichts dasteht, dann Scryfall" wäre
// eine Preisstrategie, und das Issue verbietet sie ausdrücklich.
type Preisreihe struct {
	Printing string
	Quelle   string
	Markt    string
	Art      string
	Waehrung string
	Finish   string
	Zustand  string
	Sprache  string
}

// Portfoliozeilensicht ist die Summe einer Währung.
type Portfoliozeilensicht struct {
	Waehrung       string `json:"waehrung"`
	Wert           string `json:"wert"`
	Exemplare      int    `json:"exemplare"`
	AeltesterPreis string `json:"aeltesterPreis"`
	AlterStunden   int    `json:"alterStunden"`
	Quelle         string `json:"quelle"`
}

// preise gibt die abgelegten Stände eines Drucks.
func (d *Dienst) preise(schreiben http.ResponseWriter, anfrage *http.Request) {
	staende, fehler := d.naehte.Preise.Preise(anfrage.Context(),
		anfrage.PathValue("printing"))
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"staende": ohneNil(staende)})
}

// preisverlauf gibt eine Reihe in der Zeit.
//
// # Brechend, und zwar mit Absicht (§16)
//
// Der Weg hieß `GET /drucke/{printing}/preise/{waehrung}/{finish}`, und die
// drei fehlenden Teile der Reihe — Quelle, Markt, Preisart — ergänzte die
// Naht dahinter mit „Scryfall, Papier, Verkauf". Solange es eine Quelle gab,
// war das richtig geraten; mit MTGJSON daneben ist es eine **Preisstrategie**,
// und A3 (#1442) verbietet sie: „Quellenwahl explizit".
//
// **Als Abfrage und nicht im Weg**: sieben Teile in einem Pfad sind keine
// Adresse mehr, und `zustand` und `sprache` sind oft leer — ein Pfad mit zwei
// leeren Segmenten ist keiner.
func (d *Dienst) preisverlauf(schreiben http.ResponseWriter, anfrage *http.Request) {
	abfrage := anfrage.URL.Query()

	reihe := Preisreihe{
		Printing: anfrage.PathValue("printing"),
		Quelle:   abfrage.Get("quelle"),
		Markt:    abfrage.Get("markt"),
		Art:      abfrage.Get("art"),
		Waehrung: abfrage.Get("waehrung"),
		Finish:   abfrage.Get("finish"),
		Zustand:  abfrage.Get("zustand"),
		Sprache:  abfrage.Get("sprache"),
	}

	if fehlend := reihe.fehlendeTeile(); fehlend != "" {
		fehlerhaft(schreiben, fmt.Errorf(
			"%w: die Preisreihe nennt %s nicht — welche Reihe gemeint ist, "+
				"sagt der Aufrufer", domain.ErrValidation, fehlend))

		return
	}

	staende, fehler := d.naehte.Preise.Preisverlauf(anfrage.Context(), reihe)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"staende": ohneNil(staende)})
}

// fehlendeTeile nennt, was eine Reihe nicht sagt — leer heißt vollständig.
//
// **`zustand` und `sprache` gehören nicht dazu**: leer ist dort eine Angabe
// („unbekannt"), und heute füllt sie keine Quelle.
func (r Preisreihe) fehlendeTeile() string {
	fehlend := []string{}

	for name, wert := range map[string]string{
		"quelle": r.Quelle, "markt": r.Markt, "art": r.Art,
		"waehrung": r.Waehrung, "finish": r.Finish,
	} {
		if strings.TrimSpace(wert) == "" {
			fehlend = append(fehlend, name)
		}
	}

	slices.Sort(fehlend)

	return strings.Join(fehlend, ", ")
}

// preisbulk holt eine ganze Preisdatei und legt ab, was darin steht.
//
// **Quelle und Umfang stehen im Weg** (§16): sie sind die Sache, nach der
// gefragt wird, und nicht ein Filter darauf. Eine Vorgabe gäbe es nicht —
// „Quellenwahl explizit" (A3, #1442).
func (d *Dienst) preisbulk(schreiben http.ResponseWriter, anfrage *http.Request) {
	bilanz, fehler := d.naehte.Preise.Preisbulk(anfrage.Context(),
		anfrage.PathValue("quelle"), anfrage.PathValue("umfang"))
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	bilanz.Befunde = ohneNil(bilanz.Befunde)

	ja(schreiben, http.StatusOK, bilanz)
}

// preiseHolen fragt die Quelle und legt die Stände ab.
func (d *Dienst) preiseHolen(schreiben http.ResponseWriter, anfrage *http.Request) {
	staende, befunde, fehler := d.naehte.Preise.PreiseHolen(anfrage.Context(),
		anfrage.PathValue("printing"))
	if fehler != nil {
		// **Die Befunde reisen auch im Fehlerfall mit** — sonst verschwände
		// „drei Preise waren nicht zu deuten" hinter dem einen Satz.
		fehlerhaftMit(schreiben, fehler, map[string]any{"befunde": ohneNil(befunde)})

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{
		"staende": ohneNil(staende), "befunde": ohneNil(befunde),
	})
}

// portfolio rechnet die Sammlung zusammen.
func (d *Dienst) portfolio(schreiben http.ResponseWriter, anfrage *http.Request) {
	zeilen, ohnePreis, fehler := d.naehte.Preise.Portfolio(anfrage.Context())
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{
		"zeilen": ohneNil(zeilen), "ohnePreis": ohnePreis,
	})
}

// marktangebote gibt die Posten eines Marktplatzes zu einem Druck.
//
// **Erweiternd** (§16): ein neuer Weg, der keinen alten ändert.
//
// **Nur GET** — hier wird gelesen. Was ein Konto tut, ist Sache von T11
// (#1476).
func (d *Dienst) marktangebote(schreiben http.ResponseWriter, anfrage *http.Request) {
	abfrage := anfrage.URL.Query()

	sicht, fehler := d.naehte.Preise.Marktangebote(anfrage.Context(), Marktwunsch{
		Quelle:    abfrage.Get("quelle"),
		Printing:  anfrage.PathValue("printing"),
		Sprache:   abfrage.Get("sprache"),
		Finish:    abfrage.Get("finish"),
		Zustaende: abfrage["zustand"],
	})
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, sicht)
}

// anbieterstand nennt das Kontingent an einem Anbieterkonto.
//
// **Erweiternd** (§16): ein neuer Weg, der keinen alten ändert.
func (d *Dienst) anbieterstand(schreiben http.ResponseWriter, anfrage *http.Request) {
	sicht, fehler := d.naehte.Preise.Anbieterstand(
		anfrage.Context(), anfrage.PathValue("dienst"))
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, sicht)
}

// fremdpreisverlauf holt den Verlauf eines Drucks bei einer Quelle.
//
// **`POST` und nicht `GET`**: er legt Preisstände ab, und ein `GET`, das
// schreibt, ist eine Falle für jeden Zwischenspeicher (§16).
func (d *Dienst) fremdpreisverlauf(schreiben http.ResponseWriter, anfrage *http.Request) {
	abfrage := anfrage.URL.Query()

	sicht, fehler := d.naehte.Preise.Fremdpreisverlauf(anfrage.Context(), Verlaufsbezug{
		Quelle:    abfrage.Get("quelle"),
		Printing:  anfrage.PathValue("printing"),
		Finish:    abfrage.Get("finish"),
		Zustaende: abfrage["zustand"],
	})
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, sicht)
}
