package rest

// **Zurück ist nicht eine Handlung, sondern vier** — Werkstück O7 von Phase
// 5.2 (#1431).
//
// **Hinter der Marke**, wie jeder andere Weg: die Antworten sagen, was in
// diesem Bestand liegt und was eine Rückkehr daran unklar ließe.
//
// # Warum das Zurückspielen hier weiter fehlt und die Quittung nicht
//
// Der Dienst hält **eine** Verbindung über seine Laufzeit, und alle Nähte
// haben sie sich beim Aufbau geholt; ein Tausch der Datei ließe sie auf etwas
// zeigen, das es nicht mehr gibt (`docs/FAEHIGKEITEN.md`). Eine Quittung
// tauscht nichts — sie setzt eine Marke in derselben Datei.

import (
	"context"
	"net/http"
	"strconv"
)

// Ruecksprungauskunft ist die Naht zur Anwendung.
type Ruecksprungauskunft interface {
	Ruecksprungarten(ctx context.Context) (any, error)
	Ruecksprunglage(ctx context.Context, pfad, kennwort string) (any, error)
	Rueckkehren(ctx context.Context) (any, error)
	RueckkehrQuittieren(ctx context.Context, kennung int64) error
}

// ruecksprungarten beantwortet `GET /ruecksprung/arten`.
func (d *Dienst) ruecksprungarten(schreiben http.ResponseWriter, anfrage *http.Request) {
	arten, fehler := d.naehte.Ruecksprung.Ruecksprungarten(anfrage.Context())
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, arten)
}

// ruecksprunglage beantwortet `GET /ruecksprung/lage?datei=…`.
//
// **Der Pfad steht in der Abfrage und nicht im Weg.** Ein Dateipfad trägt
// Schrägstriche; als Wegstück müsste er kodiert werden, und dann liest ihn
// niemand mehr.
func (d *Dienst) ruecksprunglage(schreiben http.ResponseWriter, anfrage *http.Request) {
	pfad := anfrage.URL.Query().Get("datei")
	if pfad == "" {
		nein(schreiben, http.StatusBadRequest, "validierung",
			"datei fehlt — nenne die Sicherungsdatei als Abfragewert")

		return
	}

	lage, fehler := d.naehte.Ruecksprung.Ruecksprunglage(anfrage.Context(), pfad,
		anfrage.Header.Get(Kennwortkopf))
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, lage)
}

// rueckkehren beantwortet `GET /rueckkehren`.
func (d *Dienst) rueckkehren(schreiben http.ResponseWriter, anfrage *http.Request) {
	rueckkehren, fehler := d.naehte.Ruecksprung.Rueckkehren(anfrage.Context())
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, rueckkehren)
}

// rueckkehrQuittieren beantwortet `POST /rueckkehren/{id}/quittung`.
func (d *Dienst) rueckkehrQuittieren(schreiben http.ResponseWriter, anfrage *http.Request) {
	kennung, fehler := strconv.ParseInt(anfrage.PathValue("id"), 10, 64)
	if fehler != nil {
		nein(schreiben, http.StatusBadRequest, "validierung",
			"id ist keine Kennung")

		return
	}

	if quittierfehler := d.naehte.Ruecksprung.RueckkehrQuittieren(
		anfrage.Context(), kennung); quittierfehler != nil {
		fehlerhaft(schreiben, quittierfehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"quittiert": kennung})
}
