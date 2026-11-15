package mcp

import (
	"context"
	"log/slog"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ein Ruf wird angemeldet, ausgeführt und wieder abgemeldet** — eine eigene
// Datei, seit `server.go` mit dem Auffangen eines Zusammenbruchs (BK-05b-MCP,
// #2113) über 400 Zeilen ging (§2).

// rufe führt ein Werkzeug aus — und antwortet nicht, wenn es abgebrochen
// wurde (`abbruch.go`).
func (s *Server) rufe(ctx context.Context, anfrage Anfrage) (Antwort, bool) {
	lauf, beenden := s.beginneLauf(ctx, anfrage.ID)

	return s.vollende(lauf, beenden, anfrage)
}

// vollende führt einen angemeldeten Ruf aus und meldet ihn wieder ab.
//
// **Angemeldet wird vorher und getrennt davon**: über stdio läuft ein Ruf
// nebenher, und ein Abbruch, der vor seiner Anmeldung einträfe, ginge an
// einem Ruf vorbei, den es noch nicht gab — gefunden, als ein Test zehn
// Minuten auf ihn wartete.
//
// **Und ein Werkzeug, das zusammenbricht, bekommt genau eine Fehlerantwort**
// (BK-05b-MCP, #2113). Über stdio läuft der Ruf in einer eigenen Goroutine,
// und eine Panik dort beendete den ganzen Prozess; über HTTP fing `net/http`
// sie, schickte aber keine Antwort und ließ den Ruf als laufend stehen. Der
// Grund geht ins Protokoll und nicht ans Gegenüber — er kann sagen, was im
// Bestand steht.
func (s *Server) vollende(lauf context.Context, beenden func() bool, anfrage Anfrage) (antwort Antwort, antworten bool) {
	defer domain.Auffangen(func(fehler error) {
		slog.ErrorContext(lauf, "ein Werkzeug ist zusammengebrochen", "grund", fehler.Error())

		antwort, antworten = Antwort{}, false
		if !beenden() {
			antwort, antworten = fehlerAuf(anfrage, codeInnererFehler, "das Werkzeug ist zusammengebrochen"), true
		}
	})

	antwort = s.fuehreAus(s.mitMelder(lauf, fortschrittsmarke(anfrage.Parameter)), anfrage)

	if beenden() {
		return Antwort{}, false
	}

	return antwort, true
}
