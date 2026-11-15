package rest

// **Die Metriken über den Dienst** — Werkstück O9 von Phase 5.2
// (#1433).
//
// # Der Authvertrag, ausdrücklich
//
// **Hinter der Marke**, wie `GET /gesundheit` und aus demselben Grund (dort
// ausgeschrieben). Zusätzlich hier: Zahlen über das, was jemand gewirkt hat,
// sind eine Auskunft über seine Arbeit — auch dann, wenn keine Kennung darin
// steht.
//
// # Kein Prometheus-Format
//
// **Und das ist eine Entscheidung, kein Versehen.** Ein Textformat mit
// `# HELP`-Zeilen wäre eine zweite Gestalt derselben Zahlen, und die
// Verpflichtung aus [0033] gälte dann für beide. Wer sie in Prometheus
// braucht, schreibt sechs Zeilen Umsetzer über diesen Weg; wer sie liest,
// bekommt dasselbe JSON wie an den anderen Flächen.

import (
	"context"
	"net/http"
)

// Metrikauskunft ist die Naht zur Anwendung.
type Metrikauskunft interface {
	Metriken(ctx context.Context) (any, error)
}

// metriken beantwortet `GET /metriken`.
func (d *Dienst) metriken(schreiben http.ResponseWriter, anfrage *http.Request) {
	stand, fehler := d.naehte.Metriken.Metriken(anfrage.Context())
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, stand)
}
