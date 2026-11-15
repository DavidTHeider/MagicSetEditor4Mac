package rest

// **Die Gesundheit über den Dienst** — Werkstück O9 von Phase 5.2
// (#1433).
//
// # Zwei Wege und nicht einer
//
//   - **`GET /gesundheit`** gibt die ganze Auskunft als JSON — für einen
//     Menschen und für ein Werkzeug, das den Grund lesen will.
//   - **`GET /bereit`** gibt **nur den Statuscode**: 200, wenn er arbeiten
//     kann, 503 sonst. Das ist die Form, die eine Sonde liest, und sie ist an
//     den anderen drei Flächen sinnlos — dort gibt es keinen Statuscode.
//
// # Der Authvertrag, ausdrücklich
//
// **Beide stehen hinter der Marke**, wie jeder andere Weg. Der Dienst hört nur
// auf `127.0.0.1` und weist ab, was nach einem Browser aussieht; es gibt
// keinen Orchestrator, der eine Sonde ohne Nachweis bräuchte. Und ein offener
// Weg wäre die einzige Stelle, an der ein lokaler Prozess **ohne** Nachweis
// etwas über diesen Bestand erführe — „migriert gerade" ist eine Auskunft.
//
// Wer das ändern will, ändert diesen Absatz und trägt den Weg in
// `docs/AUTORISIERUNG.md` als ausdrücklich offenen ein.

import (
	"context"
	"net/http"
)

// Gesundheitsauskunft ist die Naht zur Anwendung.
type Gesundheitsauskunft interface {
	Gesundheit(ctx context.Context) (any, bool, error)
}

// gesundheit beantwortet `GET /gesundheit`.
func (d *Dienst) gesundheit(schreiben http.ResponseWriter, anfrage *http.Request) {
	zustand, _, fehler := d.naehte.Gesundheit.Gesundheit(anfrage.Context())
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, zustand)
}

// bereit beantwortet `GET /bereit` — **mit dem Statuscode als Antwort**.
//
// **503 und nicht 500**: „ich kann gerade nicht" ist keine Störung, sondern
// ein Zustand, der vergeht. Ein Rumpf steht trotzdem da, damit ein Mensch den
// Grund sieht, ohne einen zweiten Weg zu rufen.
func (d *Dienst) bereit(schreiben http.ResponseWriter, anfrage *http.Request) {
	zustand, bereit, fehler := d.naehte.Gesundheit.Gesundheit(anfrage.Context())
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	if !bereit {
		ja(schreiben, http.StatusServiceUnavailable, zustand)

		return
	}

	ja(schreiben, http.StatusOK, zustand)
}
