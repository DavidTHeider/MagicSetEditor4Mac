package rest

// **Die Probleme eines ganzen Sets über REST** — PERF02
// (#2352).
//
// `GET` nennt den letzten Stand und kostet wie ein Lesen; `POST` prüft erst
// jede Karte und kostet wie ein Lauf über viele Stücke (`domain.Kostentafel`).
// **Ein `POST`, das nichts schreibt**: die Prüfung zeichnet Karten.
//
// **Im Ruf oder als Auftrag** (Zweitblick über Phase 3, Befund 11): die
// Schreibfrist des Dienstes ist fünf Minuten (`cmd/mse-rest/dienst.go`), und
// ein Set mit 3.788 Karten hält eine — ein größeres vielleicht nicht, und
// dann bekäme der Rufer statt der Antwort eine abgerissene Verbindung. Mit
// `Prefer: respond-async` läuft sie nebenher, mit Fortschritt je Karte und
// Abbruch über `DELETE /auftraege/{auftragID}` ([auftragswege]).

import (
	"context"
	"net/http"
)

// Probleme ist die Naht der Problemprüfung.
type Probleme interface {
	Problemstand(ctx context.Context, setID int64) (Problemstand, error)
	ProblemePruefen(ctx context.Context, setID int64) (Problemstand, error)
}

// Kartenproblem sind die Befunde einer Karte.
type Kartenproblem struct {
	KarteID string   `json:"karteID"`
	Befunde []string `json:"befunde"`
}

// Problemstand ist der Stand der Prüfung eines Sets: `urteil` sagt, wie weit
// er gilt, und nur bei `sicher` ist `zahl` eine Zahl, die gilt.
type Problemstand struct {
	Urteil   string          `json:"urteil"`
	Zahl     int             `json:"zahl"`
	Sicher   bool            `json:"sicher"`
	Geprueft int             `json:"geprueft"`
	Gesamt   int             `json:"gesamt"`
	Beendet  string          `json:"beendet"`
	Alter    int             `json:"alter"`
	Karten   []Kartenproblem `json:"karten"`
}

// problemstand nennt den letzten Stand — ohne zu prüfen.
func (d *Dienst) problemstand(schreiben http.ResponseWriter, anfrage *http.Request) {
	d.problemeMit(schreiben, anfrage, d.naehte.Probleme.Problemstand)
}

// problemePruefen prüft jede Karte und nennt den Stand danach.
func (d *Dienst) problemePruefen(schreiben http.ResponseWriter, anfrage *http.Request) {
	d.problemeMit(schreiben, anfrage, d.naehte.Probleme.ProblemePruefen)
}

func (d *Dienst) problemeMit(
	schreiben http.ResponseWriter, anfrage *http.Request,
	lesen func(context.Context, int64) (Problemstand, error),
) {
	setID, gut := kennung(schreiben, anfrage, "setID")
	if !gut {
		return
	}

	stand, fehler := lesen(anfrage.Context(), setID)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	stand.Karten = ohneNil(stand.Karten)
	ja(schreiben, http.StatusOK, stand)
}
