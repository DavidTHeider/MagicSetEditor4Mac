package rest

// **Wer antwortet hier** — Werkstück O9 von Phase 5.2
// (#1433).
//
// # Hinter der Marke
//
// Wie `GET /gesundheit`, und aus demselben Grund (dort ausgeschrieben). Die
// Fassung eines Programms ist eine Auskunft über den Rechner, auf dem es
// läuft; wer sie ohne Nachweis bekäme, wüsste, welche Lücken dieser Stand hat.
//
// # Was hier **nicht** steht
//
// Die **Fähigkeiten**. Sie sind eine Aushandlung und keine Auskunft, und sie
// stehen im Kompatibilitätsmanifest von O6
// (#1430).

import (
	"context"
	"net/http"
)

// Metaauskunft ist die Naht zur Anwendung.
type Metaauskunft interface {
	Meta(ctx context.Context) (any, error)
}

// meta beantwortet `GET /meta`.
func (d *Dienst) meta(schreiben http.ResponseWriter, anfrage *http.Request) {
	angaben, fehler := d.naehte.Meta.Meta(anfrage.Context())
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, angaben)
}
