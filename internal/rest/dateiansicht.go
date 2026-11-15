package rest

// **Eine Datei ansehen über den Dienst** — UX-13
// (#1685).
//
// **Lesend und ohne Bilddaten**: wer den Dienst fragt, bekommt, was über die
// Datei bekannt ist, ihren Text und die Einträge eines Pakets — das Bild
// selbst hat er ohnehin, wenn er an den Pfad darf. **Im Finder zeigen und
// Öffnen gibt es hier nicht**: der Dienst öffnete auf dem Rechner, auf dem er
// läuft, und nicht bei dem, der fragt.

import (
	"context"
	"net/http"
)

// Dateiansicht ist, was der Weg dieser Datei braucht.
type Dateiansicht interface {
	DateiAnsehen(ctx context.Context, pfad string) (any, error)
}

// dateiAnsehen beantwortet `GET /dateien/ansicht?pfad=…`.
func (d *Dienst) dateiAnsehen(schreiben http.ResponseWriter, anfrage *http.Request) {
	pfad, gut := pflichttext(schreiben, "pfad", anfrage.URL.Query().Get("pfad"))
	if !gut {
		return
	}

	ansicht, fehler := d.naehte.Dateien.DateiAnsehen(anfrage.Context(), pfad)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, ansicht)
}
