// **Turniere über den Dienst** — Werkstück 5 von Phase 4.7
// (#1294).
//
// Einordnung nach §16: **erweiternd** — neun Wege kommen hinzu, keiner ändert
// einen bestehenden.

package rest

import (
	"context"
	"net/http"
)

// Turnierwerkstatt ist, was der Dienst für Turniere braucht.
//
// **Gleichlautend mit `mcp.Turniere`**, und das ist Absicht: dieselbe
// Fähigkeit, dieselbe Naht (§16).
type Turnierwerkstatt interface {
	Turniere(ctx context.Context, spiel string) (any, error)
	Turnier(ctx context.Context, turnierID int64) (any, error)
	TurnierRangliste(ctx context.Context, turnierID int64) (any, error)
	TurnierPaarungsvorschlag(ctx context.Context, turnierID int64) (any, error)

	TurnierAnlegen(ctx context.Context, name, spiel, notiz string,
		einzelspiel bool, formatID int64) (any, error)
	TurnierKopieren(ctx context.Context, turnierID int64, name string,
		mitTeilnehmenden bool) (any, error)
	TurnierKorrekturfolgen(ctx context.Context, turnierID, paarungID int64,
		stand map[string]any) (any, error)
	TurnierAbschliessen(ctx context.Context, turnierID int64) error
	TurnierWiedereroeffnen(ctx context.Context, turnierID int64) error
	TurnierProtokoll(ctx context.Context, turnierID int64) (any, error)
	TurnierAusgabe(ctx context.Context, turnierID int64) (any, error)
	TurnierBericht(ctx context.Context, turnierID int64) (any, error)
	TurnierEinlesen(ctx context.Context, inhalt string) (any, error)
	TurnierLoeschen(ctx context.Context, turnierID int64) error
	TurnierAnmelden(ctx context.Context, turnierID int64, name string,
		deckID int64) (any, error)
	TurnierAussteigen(ctx context.Context, turnierID, teilnehmerID int64,
		ausgestiegen bool) error
	TurnierRundeLegen(ctx context.Context, turnierID int64) (any, error)
	TurnierErgebnisSetzen(ctx context.Context, turnierID, paarungID int64,
		stand map[string]any) error
}

// turniere nennt die Turniere — `?spiel=` schränkt ein.
func (d *Dienst) turniere(schreiben http.ResponseWriter, anfrage *http.Request) {
	liste, fehler := d.naehte.Turniere.Turniere(
		anfrage.Context(), anfrage.URL.Query().Get("spiel"))
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, liste)
}

// turnier gibt ein Turnier samt Teilnehmern und Verlauf.
func (d *Dienst) turnier(schreiben http.ResponseWriter, anfrage *http.Request) {
	d.zumTurnier(schreiben, anfrage, d.naehte.Turniere.Turnier)
}

// turnierRangliste rechnet den Stand.
func (d *Dienst) turnierRangliste(schreiben http.ResponseWriter, anfrage *http.Request) {
	d.zumTurnier(schreiben, anfrage, d.naehte.Turniere.TurnierRangliste)
}

// turnierVorschlag zeigt die nächste Runde, **ohne sie zu legen**.
//
// **`GET`, weil nichts entsteht.** Der Unterschied zu `POST /runden` ist der
// ganze Punkt: hinsehen, bevor man Tatsachen schafft.
func (d *Dienst) turnierVorschlag(schreiben http.ResponseWriter, anfrage *http.Request) {
	d.zumTurnier(schreiben, anfrage, d.naehte.Turniere.TurnierPaarungsvorschlag)
}

// zumTurnier bedient die Wege, die nur eine Turnierkennung nehmen.
func (d *Dienst) zumTurnier(
	schreiben http.ResponseWriter, anfrage *http.Request,
	arbeit func(context.Context, int64) (any, error),
) {
	turnierID, gut := kennung(schreiben, anfrage, "turnierID")
	if !gut {
		return
	}

	ergebnis, fehler := arbeit(anfrage.Context(), turnierID)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, ergebnis)
}

// turnierAnlegen legt ein Turnier an.
func (d *Dienst) turnierAnlegen(schreiben http.ResponseWriter, anfrage *http.Request) {
	var wunsch struct {
		Name        string `json:"name"`
		Spiel       string `json:"spiel"`
		Notiz       string `json:"notiz"`
		Einzelspiel bool   `json:"einzelspiel"`

		// FormatID ist wahlfrei — null heißt: keines (DOM-04, #1699).
		FormatID int64 `json:"formatID"`
	}

	if !gelesen(schreiben, anfrage, &wunsch) {
		return
	}

	turnier, fehler := d.naehte.Turniere.TurnierAnlegen(anfrage.Context(),
		wunsch.Name, wunsch.Spiel, wunsch.Notiz, wunsch.Einzelspiel, wunsch.FormatID)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusCreated, turnier)
}

// turnierLoeschen nimmt ein Turnier samt seinem Verlauf heraus.
func (d *Dienst) turnierLoeschen(schreiben http.ResponseWriter, anfrage *http.Request) {
	turnierID, gut := kennung(schreiben, anfrage, "turnierID")
	if !gut {
		return
	}

	if fehler := d.naehte.Turniere.TurnierLoeschen(
		anfrage.Context(), turnierID); fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	schreiben.WriteHeader(http.StatusNoContent)
}

// turnierAnmelden trägt einen Teilnehmer ein.
func (d *Dienst) turnierAnmelden(schreiben http.ResponseWriter, anfrage *http.Request) {
	turnierID, gut := kennung(schreiben, anfrage, "turnierID")
	if !gut {
		return
	}

	var wunsch struct {
		Name   string `json:"name"`
		DeckID int64  `json:"deckID"`
	}

	if !gelesen(schreiben, anfrage, &wunsch) {
		return
	}

	teilnehmer, fehler := d.naehte.Turniere.TurnierAnmelden(
		anfrage.Context(), turnierID, wunsch.Name, wunsch.DeckID)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusCreated, teilnehmer)
}

// turnierAussteigen meldet einen Teilnehmer ab — oder wieder an —,
// **innerhalb seines Turniers**.
//
// **Die Kennung des Turniers wird ausgewertet und nicht nur getragen.**
// Stünde sie nur im Weg, änderte eine Anfrage unter Turnier A jemanden aus
// Turnier B.
func (d *Dienst) turnierAussteigen(schreiben http.ResponseWriter, anfrage *http.Request) {
	turnierID, gut := kennung(schreiben, anfrage, "turnierID")
	if !gut {
		return
	}

	teilnehmerID, gutes := kennung(schreiben, anfrage, "teilnehmerID")
	if !gutes {
		return
	}

	var wunsch struct {
		Ausgestiegen bool `json:"ausgestiegen"`
	}

	if !gelesen(schreiben, anfrage, &wunsch) {
		return
	}

	if fehler := d.naehte.Turniere.TurnierAussteigen(anfrage.Context(),
		turnierID, teilnehmerID, wunsch.Ausgestiegen); fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	schreiben.WriteHeader(http.StatusNoContent)
}

// turnierRunde paart die nächste Runde und legt sie an.
func (d *Dienst) turnierRunde(schreiben http.ResponseWriter, anfrage *http.Request) {
	turnierID, gut := kennung(schreiben, anfrage, "turnierID")
	if !gut {
		return
	}

	runde, fehler := d.naehte.Turniere.TurnierRundeLegen(anfrage.Context(), turnierID)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusCreated, runde)
}

// turnierErgebnis trägt ein Ergebnis ein — **ein leerer Rumpf nimmt es
// zurück**.
//
// **`PUT` und nicht `POST`**: dieselbe Paarung, ein neuer Wert. Berichtigen
// ist am Turniertisch der Normalfall, und ein zweites `POST` legte sonst eine
// zweite Paarung nahe.
func (d *Dienst) turnierErgebnis(schreiben http.ResponseWriter, anfrage *http.Request) {
	// **Dieselbe Lesart wie die Frage nach den Folgen** (`ergebnisanfrage`).
	turnierID, paarungID, stand, gut := ergebnisanfrage(schreiben, anfrage)
	if !gut {
		return
	}

	if fehler := d.naehte.Turniere.TurnierErgebnisSetzen(
		anfrage.Context(), turnierID, paarungID, stand); fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	schreiben.WriteHeader(http.StatusNoContent)
}

// zahlOder macht aus einem fehlenden Zeiger die Null.
func zahlOder(wert *int) int {
	if wert == nil {
		return 0
	}

	return *wert
}
