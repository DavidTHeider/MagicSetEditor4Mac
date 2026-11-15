package mcp

// **Was ein gescheitertes Werkzeug dem Gegenüber sagt** — QA-09
// (#1740),
// Befund P3 in `docs/PENTEST.md`.
//
// # Der Befund
//
// **Der Dienst verbarg einen Fehler ohne Fachklasse, der Werkzeugserver
// reichte ihn durch** — samt SQLite-Meldung, Pfad unter dem Zuhause und
// Transportfehler einer Gegenstelle. Was das Modell liest, verlässt bei
// einem Anbieter im Netz den Rechner; §8 lässt Rohes nicht an eine Fläche.
// #1812 hatte den Fall der vollen Platte behoben, hier steht die Klasse.
//
// # Was durchkommt
//
// **Die Fachklassen, die auch `internal/rest/antwort.go` nennt** — ihr Satz
// ist für einen Menschen geschrieben, und das Modell lernt aus ihm, was es
// anders machen soll. **Dazu die Fehler aus `domain`, die keine Klasse
// tragen** und trotzdem nur einen Satz sind; ohne sie verlöre das Modell
// Antworten, die es heute bekommt. Alles andere wird zu einem Satz, und die
// Vorgangskennung daneben führt ins Protokoll.

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Gescheitert ist der Satz für einen Fehler ohne Fachklasse — derselbe wie
// beim Dienst.
const Gescheitert = "der Aufruf ist gescheitert"

// fachklassen sind die Fehler, deren Satz das Gegenüber lesen darf.
var fachklassen = []error{
	context.Canceled, context.DeadlineExceeded,
	domain.ErrBelegt, domain.ErrVoll, domain.ErrNotFound, domain.ErrConflict,
	domain.ErrErlaubnis, domain.ErrNurLesend, domain.ErrAbgeschaltet, domain.ErrValidation,

	// **Ohne Klasse und trotzdem nur ein Satz.**
	domain.ErrGesperrt, domain.ErrSicherung, domain.ErrSchluesselFehlt,
	domain.ErrZweiterFaktor, domain.ErrGedrosselt, domain.ErrDienstAus,
	domain.ErrEinloesungGescheitert, domain.ErrEinloesungUnbeantwortet,
	domain.ErrEinladungNichtEinloesbar,

	// **Ein fremder Dienst** (D02, #2233): sein Satz nennt den Wirt ohne
	// Abfrage, und der Grund steht im Protokoll.
	domain.ErrGegenstelle,
}

// fehlertext gibt, was ein Werkzeugfehler dem Gegenüber sagt.
func fehlertext(fehler error) string {
	// **Eine volle Platte heißt so** (#1812) — vor der Klassenfrage, denn
	// erst die Umdeutung macht aus dem Pfad einen Rat.
	fehler = domain.Platzmangel(fehler)

	for _, klasse := range fachklassen {
		if errors.Is(fehler, klasse) {
			return fehler.Error()
		}
	}

	// **Fehlerhafte Argumente sind ein Fehler des Aufrufers, kein Verrat.**
	// Ein Ruf mit unbrauchbarem JSON soll dem Gegenüber sagen, was es anders
	// machen muss — und ein Fehler von `encoding/json` trägt nur die Gestalt
	// der Eingabe, keinen Pfad und keine Meldung des Bestands. Dieselbe
	// Unterscheidung wie am Dienst, wo ein krummer Rumpf `validierung` heißt
	// (`internal/rest/endpunkte.go`).
	var (
		syntax  *json.SyntaxError
		typfall *json.UnmarshalTypeError
	)

	if errors.As(fehler, &syntax) || errors.As(fehler, &typfall) {
		return fehler.Error()
	}

	return Gescheitert
}
