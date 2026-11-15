package rest

// **Jeder Ruf wird gemessen** — OPS-05.AK-01/02
// (#1722).
//
// # Mit dem Muster und nicht mit dem Pfad
//
// **Der Pfad trüge Kennungen** — `/sets/abc123/karten` —, und damit
// Nutzinhalte, die AK-02 aus jeder Aufzeichnung heraushält. Das Muster
// (`GET /sets/{setID}/karten`) ist eines von endlich vielen; es kommt vom
// Wegweiser selbst, der dieselbe Anfrage gleich darauf zustellt.
//
// # Ganz außen
//
// **Vor Herkunft, Marke und Pfadvertrag**: auch ein abgewiesener Ruf ist ein
// Ruf, und die Fehlerrate, nach der das Issue fragt, bestünde ohne ihn nur
// aus den Fehlern, die hinter der Marke entstehen.

import (
	"context"
	"net/http"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// UnbekannterWeg steht für eine Anfrage, die keinen Weg trifft — **ein**
// Wert, damit ein Aufrufer mit erfundenen Pfaden die Spur nicht füllt.
const UnbekannterWeg = "unbekannt"

// Messung ist ein gemessener Ruf.
type Messung struct {
	Weg        string
	Schreibend bool
	Dauer      time.Duration
	Ausgang    domain.Rufausgang
}

// MitMessung setzt die Naht, die jeden Ruf aufnimmt — ohne sie misst niemand.
func (d *Dienst) MitMessung(messen func(context.Context, Messung)) *Dienst {
	d.messen = messen

	return d
}

// AusgangNachStatus ordnet einen Status einem Ausgang zu.
func AusgangNachStatus(status int) domain.Rufausgang {
	switch {
	case status >= http.StatusInternalServerError:
		return domain.RufGescheitert
	case status >= http.StatusBadRequest:
		return domain.RufAbgewiesen
	default:
		return domain.RufGelungen
	}
}

// mitMessung misst jeden Ruf, wenn eine Naht gesetzt ist.
func (d *Dienst) mitMessung(wegweiser *http.ServeMux, weiter http.Handler) http.Handler {
	if d.messen == nil {
		return weiter
	}

	return http.HandlerFunc(func(schreiben http.ResponseWriter, anfrage *http.Request) {
		beginn := time.Now()
		mitschrift := &statusmitschrift{ResponseWriter: schreiben, status: http.StatusOK}

		weiter.ServeHTTP(mitschrift, anfrage)

		weg := UnbekannterWeg
		if _, muster := wegweiser.Handler(anfrage); muster != "" {
			weg = muster
		}

		d.messen(anfrage.Context(), Messung{
			Weg:        weg,
			Schreibend: anfrage.Method != http.MethodGet && anfrage.Method != http.MethodHead,
			Dauer:      time.Since(beginn),
			Ausgang:    AusgangNachStatus(mitschrift.status),
		})
	})
}

// statusmitschrift merkt sich den Status, den der Behandler schreibt.
type statusmitschrift struct {
	http.ResponseWriter
	status int
}

// Unwrap lässt `http.ResponseController` hindurch (FMT-03.AK-03).
func (m *statusmitschrift) Unwrap() http.ResponseWriter { return m.ResponseWriter }

func (m *statusmitschrift) WriteHeader(status int) {
	m.status = status
	m.ResponseWriter.WriteHeader(status)
}
