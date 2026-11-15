package rest

// **Der Tauschabgleich im Dienst** — Werkstück `T4` von Phase 5.6
// (#1469).
//
// # Warum `POST` und nicht `GET`
//
// **Weil der fremde Stand mitkommt.** Er ist zwei Listen lang und gehört
// nicht in eine Adresse; ein `GET` mit hundert Drucken im Fragezeichenteil
// wäre eine Adresse, die Zwischenspeicher und Protokolle mitschreiben.
//
// **Geschrieben wird dabei nichts** — der Abgleich rechnet und antwortet. Das
// steht hier, weil `POST` sonst „es hat sich etwas geändert" hieße.

import (
	"context"
	"net/http"
)

// Tauschnaht ist, was der Dienst für den Abgleich braucht.
type Tauschnaht interface {
	Tauschabgleich(
		ctx context.Context, kennung string, stand Fremdstand, behalten, hoechstzahl int,
	) (Tauschabgleich, error)
}

// Fremdstand ist, was von einem Gegenüber angekommen ist.
type Fremdstand struct {
	Angebot  []Sichtkarte `json:"angebot"`
	Wuensche []Sichtkarte `json:"wuensche"`
}

// Tauschkandidat ist ein Treffer samt seiner Erklärung.
type Tauschkandidat struct {
	Wunsch  string `json:"wunsch"`
	Angebot string `json:"angebot"`
	Menge   int    `json:"menge"`
	Fehlend int    `json:"fehlend"`
	Rang    string `json:"rang"`
	Art     string `json:"art"`
}

// Tauschseite ist, was eine Seite gibt.
type Tauschseite struct {
	Zeilen   []Tauschkandidat `json:"zeilen"`
	Stueck   int              `json:"stueck"`
	Gedeckt  int              `json:"gedeckt"`
	Offen    int              `json:"offen"`
	Wert     int64            `json:"wert"`
	Waehrung string           `json:"waehrung"`
	Bepreist bool             `json:"bepreist"`
}

// Tauschvorschlag ist der Vorschlag für beide Seiten.
type Tauschvorschlag struct {
	Gebe    Tauschseite `json:"gebe"`
	Bekomme Tauschseite `json:"bekomme"`
}

// Tauschabgleich ist das Ergebnis.
type Tauschabgleich struct {
	Kennung      string           `json:"kennung"`
	Gebe         []Tauschkandidat `json:"gebe"`
	Bekomme      []Tauschkandidat `json:"bekomme"`
	Vorschlag    Tauschvorschlag  `json:"vorschlag"`
	HatVorschlag bool             `json:"hatVorschlag"`
	Befunde      []string         `json:"befunde"`
}

// tauschabgleichZeigen rechnet den Abgleich gegen einen fremden Stand.
func (d *Dienst) tauschabgleichZeigen(
	schreiben http.ResponseWriter, anfrage *http.Request,
) {
	var stand Fremdstand

	if !gelesen(schreiben, anfrage, &stand) {
		return
	}

	abgleich, fehler := d.naehte.Tausch.Tauschabgleich(anfrage.Context(),
		anfrage.PathValue("kennung"), stand,
		zahlAus(anfrage.URL.Query().Get("behalten")),
		zahlAus(anfrage.URL.Query().Get("hoechstzahl")))
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, abgleich)
}
