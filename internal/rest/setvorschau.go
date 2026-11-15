package rest

// **Ein Set vorab ansehen** — FMT-03.AK-01
// (#1692).
//
// `POST /sets/vorschau` sagt, was `POST /sets` mit der Datei täte, und
// schreibt nichts. **`POST`, obwohl nichts geschrieben wird** — der Pfad steht
// im Rumpf wie bei jedem Weg, der eine Datei liest, und der Pfadvertrag sieht
// ihn dort.

import (
	"context"
	"net/http"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Setvorschau sagt, was das Einfahren einer Datei täte.
type Setvorschau interface {
	SetVorschau(ctx context.Context, pfad string) (any, error)
}

func (d *Dienst) setvorschau(schreiben http.ResponseWriter, anfrage *http.Request) {
	var rumpf struct {
		Pfad string `json:"pfad"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	if rumpf.Pfad == "" {
		fehlerhaft(schreiben, domain.NewFieldError("pfad", domain.SchluesselValidierung,
			"Es wurde kein Pfad genannt."))

		return
	}

	vorschau, fehler := d.naehte.Setvorschau.SetVorschau(anfrage.Context(), rumpf.Pfad)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"vorschau": vorschau})
}
