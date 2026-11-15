package rest

// **Einladen und annehmen im Dienst** — NET-01
// (#1705).
//
//	POST /peer/einladungen            eine Einladung ausstellen
//	POST /peer/einladungen/pruefung   sehen, was eine sagt — schreibt nichts
//	POST /peer/einladungen/annahme    sie annehmen, ganz oder gar nicht
//
// **Der Code steht im Rumpf und nie im Pfad.** Er nennt, wie dieser Rechner
// zu erreichen ist, und ein Pfad landet in Protokollen, in denen niemand ihn
// sucht. Deshalb ist auch die Prüfung ein `POST`, obwohl sie nichts schreibt.

import (
	"context"
	"net/http"
)

// Kontakteinladung ist eine ausgestellte Einladung, wie der Dienst sie nennt.
type Kontakteinladung struct {
	Code          string `json:"code"`
	Ablauf        string `json:"ablauf"`
	Fingerabdruck string `json:"fingerabdruck"`
}

// Einladungsvorschau ist, was eine Einladung sagt, bevor jemand annimmt.
type Einladungsvorschau struct {
	Kennung        string `json:"kennung"`
	Name           string `json:"name"`
	Zweck          string `json:"zweck"`
	Adresse        string `json:"adresse"`
	Ablauf         string `json:"ablauf"`
	Fingerabdruck  string `json:"fingerabdruck"`
	Vergleichscode string `json:"vergleichscode"`
	Lage           string `json:"lage"`
	Vorgaenger     string `json:"vorgaenger"`
	Angenommen     bool   `json:"angenommen"`
	Einloesbar     bool   `json:"einloesbar"`
}

// Einladungsannahme ist, was ein Mensch zu einer Einladung entschieden hat.
type Einladungsannahme struct {
	Code           string `json:"code"`
	Name           string `json:"name"`
	Vergleichscode string `json:"vergleichscode"`
	Freigeben      bool   `json:"freigeben"`

	// MeinName ist, wie dieses Profil sich beim Einladenden nennt — leer
	// heißt: ohne Namen (NET-01b).
	MeinName string `json:"meinName,omitempty"`
}

// Einloesemeldung sagt, ob sich dieses Profil beim Einladenden melden
// konnte (NET-01b) — `gemeldet`, `ohne` oder `gescheitert`.
type Einloesemeldung struct {
	Stand   string `json:"stand"`
	Grund   string `json:"grund,omitempty"`
	Meldung string `json:"meldung,omitempty"`
}

// einladungErstellen beantwortet `POST /peer/einladungen`.
func (d *Dienst) einladungErstellen(
	schreiben http.ResponseWriter, anfrage *http.Request,
) {
	// **`adresse` ist freiwillig** (NET-01b): wer sie nennt, lädt einlösbar
	// ein — sie ist, wo ein `mse horchen` dieses Bestands steht.
	var wunsch struct {
		Name    string `json:"name"`
		Zweck   string `json:"zweck"`
		Adresse string `json:"adresse"`
	}

	if !gelesen(schreiben, anfrage, &wunsch) {
		return
	}

	einladung, fehler := d.naehte.Kontakte.KontakteinladungErstellen(
		anfrage.Context(), wunsch.Name, wunsch.Zweck, wunsch.Adresse)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusCreated, einladung)
}

// einladungPruefen beantwortet `POST /peer/einladungen/pruefung`.
func (d *Dienst) einladungPruefen(
	schreiben http.ResponseWriter, anfrage *http.Request,
) {
	var wunsch struct {
		Code string `json:"code"`
	}

	if !gelesen(schreiben, anfrage, &wunsch) {
		return
	}

	vorschau, fehler := d.naehte.Kontakte.KontakteinladungPruefen(
		anfrage.Context(), wunsch.Code)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, vorschau)
}

// einladungAnnehmen beantwortet `POST /peer/einladungen/annahme`.
func (d *Dienst) einladungAnnehmen(
	schreiben http.ResponseWriter, anfrage *http.Request,
) {
	var wunsch Einladungsannahme

	if !gelesen(schreiben, anfrage, &wunsch) {
		return
	}

	d.kontakteAntworten(schreiben, func(ctx context.Context) (Peerkontakte, error) {
		return d.naehte.Kontakte.KontakteinladungAnnehmen(ctx, wunsch)
	}, anfrage)
}
