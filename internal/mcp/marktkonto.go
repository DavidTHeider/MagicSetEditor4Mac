// **Das eigene Konto über MCP** — Werkstück `T11` von Phase 5.6
// (#1476).
//
// # Ein lesendes Werkzeug und kein zweites
//
// **Ein Modell sieht nach und kauft nicht.** Angebote anlegen, Warenkorb und
// Kauf kennt der Anbieter; über diese Fläche gibt es keines davon — und das
// ist eine Entscheidung
// ([0122](../../docs/entscheidungen/0122-ein-fremdes-konto-wird-gelesen.md),
// [0106](../../docs/entscheidungen/0106-drei-vertrauensgrenzen-und-nur-eine-wird-gebaut.md)).
//
// # Und warum es trotzdem an zwei Schaltern hängt
//
// **Weil auch Lesen hinausgeht.** Der Dienst muss an sein und die Kontoschau
// ebenso; beide fangen aus an, und beide schaltet ein Mensch am Fenster.

package mcp

import (
	"context"
	"encoding/json"
)

// Kontoauskunft ist die Naht zum eigenen Konto.
type Kontoauskunft interface {
	Marktkonto(ctx context.Context, quelle string, seite, menge int) (
		Kontostand, error)
}

// Eigenangebot ist ein Posten, den das eigene Konto anbietet.
type Eigenangebot struct {
	ID          int64  `json:"id"`
	BlueprintID int64  `json:"blueprintID"`
	Name        string `json:"name"`
	Menge       int    `json:"menge"`
	PreisCent   int64  `json:"preisCent"`
	Waehrung    string `json:"waehrung"`
	Eigennotiz  string `json:"eigennotiz"`
}

// Marktbestellung ist ein Kauf, an dem das eigene Konto beteiligt ist.
type Marktbestellung struct {
	ID   int64  `json:"id"`
	Code string `json:"code"`

	// Zustand ist einer der acht des Anbieters — unübersetzt.
	Zustand string `json:"zustand"`

	UeberNullkonto bool   `json:"ueberNullkonto"`
	Rolle          string `json:"rolle"`
	Stueck         int    `json:"stueck"`
	Bezahlt        string `json:"bezahlt"`

	// BewegtBestand sagt, ob diese Bestellung nach der Regel des Anbieters
	// ein Stück aus dem eigenen Lager genommen hat — **eine Auskunft und
	// keine Buchung**.
	BewegtBestand bool `json:"bewegtBestand"`
}

// Kontostand ist die Übersicht über das eigene Konto.
type Kontostand struct {
	Name         string            `json:"name"`
	NutzerID     int64             `json:"nutzerID"`
	Angebote     []Eigenangebot    `json:"angebote"`
	Bestellungen []Marktbestellung `json:"bestellungen"`
	Befunde      []string          `json:"befunde"`
}

// marktkontowerkzeug gibt das eine lesende — oder keines.
func marktkontowerkzeug(auskunft Kontoauskunft) []Werkzeug {
	if auskunft == nil {
		return nil
	}

	return []Werkzeug{{
		Name: "lesen.marktkonto", Art: Lesend, Idempotent: true,
		Erklaerung: "Nennt, was im EIGENEN Konto bei einem Marktplatz liegt: " +
			"die eigenen Angebote und die eigenen Bestellungen. LIEST NUR — " +
			"es wird nichts angeboten, nichts gekauft und nichts versandt; " +
			"diese Wege gibt es hier gar nicht. »bewegtBestand« sagt, ob eine " +
			"Bestellung nach der Regel des Anbieters ein Stück aus dem Lager " +
			"genommen hat; GEBUCHT wird daraus nichts. Der Weg hängt an zwei " +
			"Schaltern, die beide AUS beginnen: dem Dienst und der " +
			"Kontoschau — beide schaltet ein Mensch am Fenster. »preisCent« " +
			"steht in der KLEINSTEN Einheit der Währung — bei EUR Cent, beim " +
			"Yen ganze Yen.",
		Schema: objekt(map[string]any{
			"quelle": text("Der Marktplatz — heute »cardtrader«."),
			"seite":  zahl("Die Seite der Bestellungen, ab 1. Null nimmt die Vorgabe des Anbieters."),
			"menge":  zahl("Wie viele Bestellungen je Seite, 1 bis 100. Null nimmt die Vorgabe."),
		}, "quelle"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var argumente struct {
				Quelle string `json:"quelle"`
				Seite  int    `json:"seite"`
				Menge  int    `json:"menge"`
			}

			if err := lies(roh, &argumente); err != nil {
				return "", err
			}

			stand, err := auskunft.Marktkonto(ctx, argumente.Quelle,
				argumente.Seite, argumente.Menge)
			if err != nil {
				return "", err
			}

			return alsJSON(stand)
		},
	}}
}
