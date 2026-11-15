package mcp

// **Der Pfadvertrag des Werkzeugservers** — OPS-02.AK-04
// (#1719).
//
// **Vor dem Werkzeug und nicht in ihm** — dieselbe Regel wie beim Dienst
// (`internal/rest/pfadvertrag.go`): an den Namen aus `domain.IstPfadfeld`,
// in der Ordnung der Namen. Gesetzt wird er, wo der Server über HTTP
// spricht; über stdio spricht nur, wer den Prozess selbst gestartet hat.
//
// **In jeder Schreibweise, die der Dekodierer zuordnet** (#2073) — für die
// Pfadfelder wie für die Namensfelder: `"Datei"` landet im selben Feld wie
// `"datei"`. Und unter den [Werkzeug.Pfadfelder], die ein Werkzeug nennt.

import (
	"context"
	"encoding/json"
	"maps"
	"slices"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// MitPfadvertrag setzt den Vertrag — ohne ihn urteilt niemand.
func (s *Server) MitPfadvertrag(vertrag func(ctx context.Context, pfad string) error) *Server {
	s.pfadvertrag = vertrag

	return s
}

// pfadeVerwehrt sagt, woran ein Pfad in den Argumenten scheitert — `nil`,
// wenn keiner scheitert oder niemand urteilt.
func (s *Server) pfadeVerwehrt(ctx context.Context, werkzeug Werkzeug, roh json.RawMessage) error {
	if s.pfadvertrag == nil {
		return nil
	}

	for _, pfad := range domain.PfadeIn(ohneNamensfelder(roh, werkzeug.Namensfelder), werkzeug.Pfadfelder...) {
		if fehler := s.pfadvertrag(ctx, pfad); fehler != nil {
			return fehler
		}
	}

	return nil
}

// ohneNamensfelder nimmt die Felder heraus, die einen Namen tragen — **in
// jeder Schreibweise**, denn der Dekodierer legt sie alle in dasselbe Feld.
// Was kein Objekt ist, bleibt, wie es ist, und sagt dann selbst, was an ihm
// falsch ist.
func ohneNamensfelder(roh json.RawMessage, namen []string) json.RawMessage {
	var felder map[string]json.RawMessage
	if len(namen) == 0 || json.Unmarshal(roh, &felder) != nil {
		return roh
	}

	maps.DeleteFunc(felder, func(name string, _ json.RawMessage) bool {
		return slices.ContainsFunc(namen, func(namensfeld string) bool { return strings.EqualFold(name, namensfeld) })
	})

	ohne, _ := json.Marshal(felder) //nolint:errcheck,errchkjson // eine Abbildung roher Werte lässt sich immer schreiben

	return ohne
}
