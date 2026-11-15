package rest

// **Freigegebene Sammlungssichten im Dienst** — Werkstück `T3` von Phase 5.6
// (#1468).
//
// # Warum eine Freigabe eine Ressource ist und kein Befehl
//
// **Weil es sie gibt oder nicht gibt.** `POST` erteilt, `DELETE` nimmt
// zurück, `GET` sagt, was steht — dieselbe Form wie bei der Sperre eines
// Kontakts. Ein `POST /freigaben/widerrufen` wäre ein Befehl, und der ließe
// sich nicht zweimal ausführen, ohne dass jemand über die zweite Antwort
// nachdenkt.

import (
	"context"
	"net/http"
)

// Sichtfreigabenaht ist, was der Dienst für die Sichten braucht.
type Sichtfreigabenaht interface {
	Sammlungsfreigaben(
		ctx context.Context, kennung string,
	) (Sammlungsfreigaben, error)
	SammlungFreigeben(
		ctx context.Context, freigabe Sammlungsfreigabe,
	) (Sammlungsfreigaben, error)
	SammlungsfreigabeWiderrufen(
		ctx context.Context, kennung, zweck, umfang string,
	) (Sammlungsfreigaben, error)
	Freigabesicht(
		ctx context.Context, kennung, zweck string,
	) ([]Sichtkarte, error)
}

// Sammlungsfreigabe ist eine erteilte Sicht.
type Sammlungsfreigabe struct {
	Empfaenger string   `json:"empfaenger"`
	Umfang     string   `json:"umfang"`
	Zweck      string   `json:"zweck"`
	Auswahl    []string `json:"auswahl"`
	Behalten   int      `json:"behalten"`
	Ablauf     string   `json:"ablauf"`
	Erteilt    string   `json:"erteilt"`
}

// Sammlungsfreigaben ist die Liste samt den möglichen Umfängen.
type Sammlungsfreigaben struct {
	Freigaben []Sammlungsfreigabe `json:"freigaben"`
	Umfaenge  []string            `json:"umfaenge"`

	// Nachladefehler: erteilt oder widerrufen **ist**, nur die Liste danach
	// war nicht zu lesen (§7, Regel 3; #1970). Erweiternd (0033).
	Nachladefehler *Befund `json:"nachladefehler,omitempty"`
}

// Sichtkarte ist eine Zeile, wie ein Gegenüber sie sähe.
//
// **Sechs Felder und eine Referenz** — und keines mehr. Was hier stünde,
// stünde am Draht: kein Erwerbspreis, keine Notiz, kein Lagerort, keine
// Fassung, keine Bestandskennung.
type Sichtkarte struct {
	Referenz    string `json:"referenz"`
	Printing    string `json:"printing"`
	Menge       int    `json:"menge"`
	Zustand     string `json:"zustand"`
	Sprache     string `json:"sprache"`
	Finish      string `json:"finish"`
	Bezeichnung string `json:"bezeichnung"`
}

// freigabenZeigen beantwortet `GET /peer/kontakte/{kennung}/freigaben`.
func (d *Dienst) freigabenZeigen(
	schreiben http.ResponseWriter, anfrage *http.Request,
) {
	d.freigabenAntworten(schreiben, func(ctx context.Context) (
		Sammlungsfreigaben, error,
	) {
		return d.naehte.Sichtfreigaben.Sammlungsfreigaben(
			ctx, anfrage.PathValue("kennung"))
	}, anfrage)
}

// freigabeErteilen beantwortet `POST /peer/kontakte/{kennung}/freigaben`.
func (d *Dienst) freigabeErteilen(
	schreiben http.ResponseWriter, anfrage *http.Request,
) {
	var wunsch Sammlungsfreigabe

	if !gelesen(schreiben, anfrage, &wunsch) {
		return
	}

	// **Der Empfänger steht im Weg und nicht im Rumpf.** Zwei Stellen für
	// dieselbe Angabe liefen auseinander, und welche gälte, entschiede der
	// Code — nicht der Aufrufer.
	wunsch.Empfaenger = anfrage.PathValue("kennung")

	d.freigabenAntworten(schreiben, func(ctx context.Context) (
		Sammlungsfreigaben, error,
	) {
		return d.naehte.Sichtfreigaben.SammlungFreigeben(ctx, wunsch)
	}, anfrage)
}

// freigabeWiderrufen beantwortet
// `DELETE /peer/kontakte/{kennung}/freigaben/{zweck}/{umfang}`.
func (d *Dienst) freigabeWiderrufen(
	schreiben http.ResponseWriter, anfrage *http.Request,
) {
	d.freigabenAntworten(schreiben, func(ctx context.Context) (
		Sammlungsfreigaben, error,
	) {
		return d.naehte.Sichtfreigaben.SammlungsfreigabeWiderrufen(ctx,
			anfrage.PathValue("kennung"), anfrage.PathValue("zweck"),
			anfrage.PathValue("umfang"))
	}, anfrage)
}

// freigabesichtZeigen beantwortet `GET /peer/kontakte/{kennung}/sicht`.
//
// **Der Zweck steht in der Abfrage**: dieselbe Sammlung sieht für „teilen"
// anders aus als für „handel", und beides ist eine Sicht auf dieselbe
// Ressource.
func (d *Dienst) freigabesichtZeigen(
	schreiben http.ResponseWriter, anfrage *http.Request,
) {
	sicht, fehler := d.naehte.Sichtfreigaben.Freigabesicht(anfrage.Context(),
		anfrage.PathValue("kennung"), anfrage.URL.Query().Get("zweck"))
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"zeilen": sicht})
}

// freigabenAntworten fährt einen der drei Wege und schreibt die Liste.
func (d *Dienst) freigabenAntworten(
	schreiben http.ResponseWriter,
	fahre func(context.Context) (Sammlungsfreigaben, error),
	anfrage *http.Request,
) {
	freigaben, fehler := fahre(anfrage.Context())
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, freigaben)
}
