package mcp

// **Einen laufenden Ruf abbrechen** — FMT-03.AK-02
// (#1692).
//
// Nachgelesen am 2026-09-29 über Context7, MCP-Spezifikation Fassung
// 2026-07-28, `basic/patterns/cancellation`: über stdio schickt der Aufrufer
// `notifications/cancelled` mit der `requestId`; der Server soll die Arbeit
// anhalten, freigeben, was sie hielt, und **nicht antworten**. Ist der Ruf
// unbekannt oder schon fertig, darf er die Meldung übergehen. Über HTTP ist
// das Schließen des Ereignisstroms der Abbruch — dort endet der Kontext der
// Anfrage, und dieselbe Kette hält an.
//
// **Was angehalten wird, entscheidet der Fachkern**: er fragt zwischen zwei
// Stücken (`service.Schritt`) und lässt nur Ganzes stehen (AK-04).

import (
	"bytes"
	"context"
	"encoding/json"
)

// laufenderRuf ist ein Werkzeugruf, der noch arbeitet.
type laufenderRuf struct {
	abbrechen   context.CancelFunc
	abgebrochen bool
}

// beginneLauf meldet einen Ruf an. Zurück kommen sein Kontext und das Ende,
// das sagt, ob ihn jemand abgebrochen hat.
func (s *Server) beginneLauf(ctx context.Context, id json.RawMessage) (lauf context.Context, beenden func() bool) {
	lauf, abbrechen := context.WithCancel(ctx)
	schluessel := kennungsschluessel(id)
	eintrag := &laufenderRuf{abbrechen: abbrechen}

	s.laufsperre.Lock()
	if s.laufende == nil {
		s.laufende = map[string]*laufenderRuf{}
	}

	s.laufende[schluessel] = eintrag
	s.laufsperre.Unlock()

	return lauf, func() bool {
		s.laufsperre.Lock()
		defer s.laufsperre.Unlock()

		// **Nur den eigenen Eintrag**: zwei Rufe über HTTP dürfen dieselbe
		// Kennung tragen, und der zweite soll abbrechbar bleiben.
		if s.laufende[schluessel] == eintrag {
			delete(s.laufende, schluessel)
		}

		abbrechen()

		return eintrag.abgebrochen
	}
}

// abbruchHoeren bricht den genannten Ruf ab — einen unbekannten übergeht es.
func (s *Server) abbruchHoeren(parameter json.RawMessage) {
	var gelesen struct {
		Kennung json.RawMessage `json:"requestId"`
	}

	if json.Unmarshal(parameter, &gelesen) != nil || len(gelesen.Kennung) == 0 {
		return
	}

	s.laufsperre.Lock()
	defer s.laufsperre.Unlock()

	if eintrag := s.laufende[kennungsschluessel(gelesen.Kennung)]; eintrag != nil {
		eintrag.abgebrochen = true
		eintrag.abbrechen()
	}
}

// kennungsschluessel macht aus einer Kennung einen Schlüssel — ohne
// Leerraum darum, damit `"a"` und ` "a" ` derselbe Ruf sind.
func kennungsschluessel(id json.RawMessage) string {
	return string(bytes.TrimSpace(id))
}
