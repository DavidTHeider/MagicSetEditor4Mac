// **Drei Zustände, nicht zwei**
// (#1155).
//
// Ein Set ist da, ein Set ist nicht da — **oder es ist nicht zu sagen**. Der
// dritte ging bis zum 2026-08-26 als „nicht da" hinaus.
//
// **Eine eigene Datei**, seit `durchreichen_test.go` über die Zeilengrenze aus
// §2 ging. Der Schnitt folgt der Frage: dort steht, dass jeder Endpunkt seinen
// Fehler durchreicht — hier, dass **dieser** Endpunkt drei Antworten
// unterscheidet.
package rest_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// **Ein Fehler beim Nachsehen ist ein Fehler, kein „nicht gefunden".**
//
// # Was hier vorher stand
//
// „`setGibtEs` liest die Setköpfe. Scheitert das, gibt es keine Auskunft — und
// die Antwort ist dann »nicht gefunden« und nicht ein fröhliches 200."
//
// **Der erste Halbsatz stimmt, der zweite zog den falschen Schluss.** Es gibt
// drei Zustände, nicht zwei: das Set ist da, das Set ist nicht da, und **es
// ist nicht zu sagen**. Der dritte als „nicht da" auszugeben nimmt einem
// Aufrufer die Unterscheidung, die §8 gerade herstellen will — „der Datensatz
// fehlt" ist ein normaler Zustand, „der Bestand ist nicht lesbar" ein Ausfall.
//
// Angestrichen vom Gesamtreview vor 0.2.0
// (#1155).
//
// **Einordnung nach §16: brechend** — wer heute auf die 404 baut, bekommt eine
// 500. Zulässig, weil die Fläche in keiner Fassung ausgeliefert ist.
func TestEinFehlerBeimNachsehenIstEinFehler(t *testing.T) {
	t.Parallel()

	jetzt := time.Date(2026, time.August, 23, 12, 0, 0, 0, time.UTC)
	leserOhneKoepfe := kopflos{}

	dienst := rest.NeuerDienst(probemarke, func() time.Time { return jetzt },
		naehteMit(leserOhneKoepfe))

	antwort := ruf(t, dienst, endpunkt{methode: http.MethodGet, weg: "/sets/1/karten"}, gueltig)
	pruefeAbweisung(t, antwort, http.StatusInternalServerError, "unbekannt")
}

// **Und ein Set, das es wirklich nicht gibt, bleibt „nicht gefunden".**
//
// **Ohne diesen Fall wäre die Korrektur eine Behauptung**: eine Antwort, die
// **immer** 500 gäbe, bestünde den Test darüber ebenso — und nähme dieselbe
// Unterscheidung weg, nur andersherum (§10).
func TestEinSetDasEsNichtGibtBleibtNichtGefunden(t *testing.T) {
	t.Parallel()

	jetzt := time.Date(2026, time.August, 23, 12, 0, 0, 0, time.UTC)

	dienst := rest.NeuerDienst(probemarke, func() time.Time { return jetzt },
		naehteMit(leererLeser{}))

	antwort := ruf(t, dienst, endpunkt{methode: http.MethodGet, weg: "/sets/1/karten"}, gueltig)
	pruefeAbweisung(t, antwort, http.StatusNotFound, "nichtGefunden")
}

// leererLeser kennt keine Karten und **keine Sets** — ohne zu scheitern.
type leererLeser struct{ williger }

func (leererLeser) Karten(context.Context, int64) ([]*domain.CardRow, error) {
	return nil, nil
}

func (leererLeser) Setkoepfe(context.Context) ([]domain.SetHead, error) {
	return nil, nil
}

// kopflos gibt keine Karten und scheitert beim Lesen der Setköpfe.
type kopflos struct{ williger }

func (kopflos) Karten(context.Context, int64) ([]*domain.CardRow, error) {
	return nil, nil
}

func (kopflos) Setkoepfe(context.Context) ([]domain.SetHead, error) {
	return nil, errors.New("der Bestand antwortet nicht")
}

// naehteMit baut Nähte, in denen nur der Leser eigen ist.
//
// **Als Helfer, seit es zwei Fälle gibt**, die denselben Aufbau brauchen und
// sich in **einer** Naht unterscheiden. Zwei Aufzählungen von zwanzig Zeilen
// nebeneinander laufen auseinander, und dann prüft der zweite Fall etwas
// anderes als der erste.
func naehteMit(leser rest.Leser) rest.Naehte {
	return rest.Naehte{
		Leser:         leser,
		Schreiber:     williger{},
		Ausfuhr:       williger{},
		Abgleich:      williger{},
		Bilder:        williger{},
		Auskunft:      williger{},
		Sammeldaten:   williger{},
		Anlage:        williger{},
		Saetze:        williger{},
		Konsole:       williger{},
		Werkstatt:     williger{},
		Wortwahl:      nil,
		Woerter:       nil,
		Pack:          williger{},
		Angaben:       williger{},
		Stile:         williger{},
		Zusatz:        williger{},
		Pakete:        williger{},
		Zahlen:        williger{},
		Verlauf:       williger{},
		Einstellungen: williger{},
		Fassungen:     williger{},
		Anleger:       nil,
	}
}
