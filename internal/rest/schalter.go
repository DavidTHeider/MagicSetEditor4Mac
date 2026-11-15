package rest

// **Bereiche und Funktionen abschalten, am Dienst** — OPS-10
// (#1908).
//
// **Hinter der Marke**: wer sich nicht ausweist, erfährt nicht, was hier
// abgeschaltet ist — er bekommt dieselbe Abweisung wie bei jedem Weg.

import (
	"context"
	"net/http"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Schalternaht weist einen Weg ab, dessen Bereich abgeschaltet ist — `nil`
// heißt durchlassen. Der Weg kommt ohne Platzhalternamen ([domain.RESTWeg]).
type Schalternaht func(ctx context.Context, weg string) error

// Schalterauskunft ist, was `GET /schalter` braucht.
type Schalterauskunft interface {
	Schalter(ctx context.Context) ([]Schalter, error)
}

// Schalter ist ein Schalter mit seinem wirksamen Stand.
type Schalter struct {
	Kennung string `json:"kennung"`
	Art     string `json:"art"`
	Reife   string `json:"reife"`
	Vorgabe string `json:"vorgabe"`
	An      bool   `json:"an"`
}

// MitSchaltern setzt die Naht, die jeden Weg vor seiner Arbeit fragt.
func (d *Dienst) MitSchaltern(naht Schalternaht) *Dienst {
	d.schalter = naht

	return d
}

// mitSchaltern fragt die Naht mit dem Muster, das die Anfrage trifft.
func (d *Dienst) mitSchaltern(weg *http.ServeMux, weiter http.Handler) http.Handler {
	return http.HandlerFunc(func(schreiben http.ResponseWriter, anfrage *http.Request) {
		if d.schalter != nil {
			_, muster := weg.Handler(anfrage)
			if fehler := d.schalter(anfrage.Context(), domain.RESTWeg(muster)); fehler != nil {
				fehlerhaft(schreiben, fehler)

				return
			}
		}

		weiter.ServeHTTP(schreiben, anfrage)
	})
}

func (d *Dienst) schalterSchauen(schreiben http.ResponseWriter, anfrage *http.Request) {
	stand, fehler := d.naehte.Schalter.Schalter(anfrage.Context())
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, stand)
}
