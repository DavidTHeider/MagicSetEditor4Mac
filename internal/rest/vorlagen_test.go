// **Eine leere Vorlagenliste kommt als Liste an und nicht als `null`** —
// #1291.
//
// Der Zweig, der `nil` zu `[]` macht: ein Spiel ohne Exportvorlage ist der
// Normalfall, nicht der Ausnahmefall (§3).
package rest_test

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// ohneVorlagen ist ein Spiel, für das es keine Exportvorlage gibt.
type ohneVorlagen struct{ williger }

func (ohneVorlagen) Vorlagen(context.Context, int64) ([]rest.Vorlagenzeile, []string, error) {
	return nil, nil, nil
}

// **Und sie kommt als leere Liste an.**
func TestEineLeereVorlagenlisteKommtAlsListeAn(t *testing.T) {
	t.Parallel()

	dienst := rest.NeuerDienst(probemarke,
		func() time.Time { return time.Date(2026, time.August, 30, 12, 0, 0, 0, time.UTC) },
		rest.Naehte{Vorlagen: ohneVorlagen{}})

	antwort := ruf(t, dienst, endpunkt{
		methode: http.MethodGet, weg: "/sets/1/vorlagen",
	}, gueltig)

	if antwort.Code != http.StatusOK {
		t.Fatalf("Status %d: %s", antwort.Code, antwort.Body.String())
	}

	if strings.Contains(antwort.Body.String(), "null") {
		t.Errorf("die Listen kamen als null an: %s", antwort.Body.String())
	}
}

// **Ein Teilerfolg beim Einlesen geht auch über den Dienst nicht verloren** —
// Befund 1 des Zweitblicks über Phase 4.4
// (#1291).
func TestEinTeilerfolgBeimEinlesenStehtInDerFehlerantwort(t *testing.T) {
	t.Parallel()

	dienst := rest.NeuerDienst(probemarke,
		func() time.Time { return time.Date(2026, time.August, 30, 12, 0, 0, 0, time.UTC) },
		rest.Naehte{Ausfuhr: halbeEinfuhr{}})

	antwort := ruf(t, dienst, endpunkt{
		methode: http.MethodPost, weg: "/mtgjson",
		rumpf: `{"pfad":"/tmp/kaputt.json","spiel":"magic"}`,
	}, gueltig)

	if antwort.Code < 400 {
		t.Fatalf("Status %d — der Fehler ging verloren", antwort.Code)
	}

	// **Die Kennung steht in der Fehlerantwort.** Sie ist die einzige Spur,
	// die ein Aufrufer nach einem Abbruch hat.
	if !strings.Contains(antwort.Body.String(), `"7"`) {
		t.Errorf("die Antwort nennt das angelegte Set nicht: %s", antwort.Body.String())
	}
}

// halbeEinfuhr legt ein Set an und scheitert danach.
type halbeEinfuhr struct{ williger }

func (halbeEinfuhr) MTGJSONLesen(context.Context, string, string) (
	angelegt []int64, befunde []string, fehler error,
) {
	return []int64{7}, []string{"eine Zeile war krumm"},
		errors.New("die Datei bricht ab")
}

// **Und über den Dienst ebenso** (Nachlauf zu #1291).
func TestEinTeilerfolgBeiDerUebernahmeStehtInDerFehlerantwort(t *testing.T) {
	t.Parallel()

	dienst := rest.NeuerDienst(probemarke,
		func() time.Time { return time.Date(2026, time.August, 30, 12, 0, 0, 0, time.UTC) },
		rest.Naehte{Pakete: halbeUebernahme{}})

	antwort := ruf(t, dienst, endpunkt{
		methode: http.MethodPost, weg: "/pakete/uebernommen", rumpf: `{"ort":"/tmp/data"}`,
	}, gueltig)

	if antwort.Code < 400 {
		t.Fatalf("Status %d — der Fehler ging verloren", antwort.Code)
	}

	if !strings.Contains(antwort.Body.String(), "magic") {
		t.Errorf("die Antwort nennt das übernommene Paket nicht: %s", antwort.Body.String())
	}
}

// halbeUebernahme kopiert ein Paket und scheitert beim Auffrischen.
type halbeUebernahme struct{ williger }

func (halbeUebernahme) PaketUebernehmen(context.Context, string) (rest.Uebernahme, error) {
	return rest.Uebernahme{
		Uebernommen: []rest.Paket{{Art: "mse-game", Name: "magic"}},
		Uebergangen: []string{},
		Befunde:     []string{},
	}, errors.New("der Katalog ist zu")
}
