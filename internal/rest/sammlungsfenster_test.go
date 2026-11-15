// **Der Ausschnitt der Sammlung über die Rückschleife** — Werkstück
// #1500.
//
// # Warum dieselbe Adresse
//
// `GET /sammlung` gab bisher alles. Ein zweiter Weg für dieselbe Frage wären
// zwei Verträge über eine Fläche, die über Fassungen hinweg zusagt (§16) —
// also **erweitert** sie: wer nichts angibt, bekommt weiter alles; wer eine
// Sicht oder einen Ausschnitt nennt, bekommt ihn samt beider Zahlen.
//
// Geprüft wird hier beides, und zwar an der Naht: dass die Angaben aus der
// Abfrage dort **ankommen**, und dass eine, die keine Zahl ist, als null
// ankommt statt als Fehler.
package rest_test

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// merkendesFenster hält fest, womit die Naht gerufen wurde.
type merkendesFenster struct {
	williger

	gerufen                        bool
	suche, ausfuehrung, sortierung string
	ab, anzahl                     int
	fehler                         error
}

func (m *merkendesFenster) Sammlungsfenster(
	_ context.Context, suche, ausfuehrung, sortierung string, ab, anzahl int,
) (fenster []rest.Exemplarsicht, zahlen domain.Fensterzahlen, fehler error) {
	m.gerufen = true
	m.suche, m.ausfuehrung, m.sortierung = suche, ausfuehrung, sortierung
	m.ab, m.anzahl = ab, anzahl

	if m.fehler != nil {
		return nil, domain.Fensterzahlen{}, m.fehler
	}

	return []rest.Exemplarsicht{{ID: 3, Printing: "probe-druck", Menge: 1}},
		domain.Fensterzahlen{Gesamt: 1, Insgesamt: 42, Stueck: 99}, nil
}

// dienstMitFenster baut einen Dienst über einem merkenden Fenster.
func dienstMitFenster(t *testing.T, doppel *merkendesFenster) *rest.Dienst {
	t.Helper()

	jetzt := time.Date(2026, time.September, 9, 12, 0, 0, 0, time.UTC)

	return rest.NeuerDienst(probemarke, func() time.Time { return jetzt },
		rest.Naehte{Sammlung: doppel})
}

// **Die Sicht und der Ausschnitt kommen an der Naht an.**
func TestDieSichtDerSammlungKommtAnDerNahtAn(t *testing.T) {
	t.Parallel()

	doppel := &merkendesFenster{}

	antwort := ruf(t, dienstMitFenster(t, doppel), endpunkt{
		methode: http.MethodGet,
		weg:     "/sammlung?suche=drache&ausfuehrung=foil&sortierung=name&ab=10&anzahl=5",
	}, gueltig)

	if antwort.Code != http.StatusOK {
		t.Fatalf("Status %d: %s", antwort.Code, antwort.Body.String())
	}

	if doppel.suche != "drache" || doppel.ausfuehrung != "foil" ||
		doppel.sortierung != "name" || doppel.ab != 10 || doppel.anzahl != 5 {
		t.Errorf("an der Naht kamen an: %q %q %q %d %d",
			doppel.suche, doppel.ausfuehrung, doppel.sortierung, doppel.ab, doppel.anzahl)
	}

	// **Beide Zahlen stehen da** — ohne sie wüsste niemand, wovon der
	// Ausschnitt ein Ausschnitt ist.
	for _, nennt := range []string{`"gesamt":1`, `"insgesamt":42`, `"stueck":99`} {
		if !strings.Contains(antwort.Body.String(), nennt) {
			t.Errorf("die Antwort nennt %s nicht: %s", nennt, antwort.Body.String())
		}
	}
}

// **Was keine Zahl ist, ist null** — und nicht ein Fehler.
//
// Ein `ab=neulich` ist ein Aufrufer, der sich vertan hat; ihn mit einem 400 zu
// beantworten wäre vertretbar, aber die Fläche gibt dann lieber den Anfang der
// Liste als gar nichts. **Eine negative Zahl ebenso**: sie zählte sonst
// rückwärts in einen Ausschnitt, den es nicht gibt.
func TestEineAbfrageOhneZahlZaehltAbNull(t *testing.T) {
	t.Parallel()

	for name, weg := range map[string]string{
		"keine Zahl":      "/sammlung?ab=neulich&anzahl=bald",
		"negativ":         "/sammlung?ab=-3&anzahl=-7",
		"leer, aber da":   "/sammlung?ab=&anzahl=",
		"zu groß für int": "/sammlung?ab=99999999999999999999&anzahl=1",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			doppel := &merkendesFenster{}

			antwort := ruf(t, dienstMitFenster(t, doppel), endpunkt{
				methode: http.MethodGet, weg: weg,
			}, gueltig)

			if antwort.Code != http.StatusOK {
				t.Fatalf("Status %d: %s", antwort.Code, antwort.Body.String())
			}

			if doppel.ab != 0 {
				t.Errorf("ab kam als %d an", doppel.ab)
			}
		})
	}
}

// **Ohne Sicht bleibt es die ganze Sammlung** — der Vertrag von vorher.
func TestOhneSichtKommtDieGanzeSammlung(t *testing.T) {
	t.Parallel()

	doppel := &merkendesFenster{}

	antwort := ruf(t, dienstMitFenster(t, doppel), endpunkt{
		methode: http.MethodGet, weg: "/sammlung",
	}, gueltig)

	if antwort.Code != http.StatusOK {
		t.Fatalf("Status %d: %s", antwort.Code, antwort.Body.String())
	}

	if doppel.gerufen {
		t.Error("das Fenster wurde gerufen, obwohl niemand einen Ausschnitt wollte")
	}

	if strings.Contains(antwort.Body.String(), `"gesamt"`) {
		t.Errorf("die Antwort trägt eine Zahl, die es vorher nicht gab: %s",
			antwort.Body.String())
	}
}

// **Ein Fehler des Fensters kommt heraus.**
func TestEinFehlerDesSammlungsfenstersKommtHeraus(t *testing.T) {
	t.Parallel()

	doppel := &merkendesFenster{fehler: errors.New("die Naht schweigt")}

	antwort := ruf(t, dienstMitFenster(t, doppel), endpunkt{
		methode: http.MethodGet, weg: "/sammlung?suche=drache",
	}, gueltig)

	if antwort.Code < http.StatusBadRequest {
		t.Errorf("endete mit %d: %s", antwort.Code, antwort.Body.String())
	}
}
