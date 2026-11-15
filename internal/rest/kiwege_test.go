package rest_test

// **Die Wege der KI am Dienst** — Werkstück A7 von Phase 5.3
// (#1446).

import (
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// kidienst baut einen Dienst mit einer Preiswerkstatt.
func kidienst(t *testing.T, werkstatt rest.Preiswerkstatt) *rest.Dienst {
	t.Helper()

	jetzt := time.Date(2026, time.September, 11, 12, 0, 0, 0, time.UTC)

	return rest.NeuerDienst(probemarke, func() time.Time { return jetzt },
		rest.Naehte{Preise: werkstatt})
}

// **Der Vorschlag kommt an.**
func TestDerKIVorschlagKommtAmDienstAn(t *testing.T) {
	t.Parallel()

	antwort := ruf(t, kidienst(t, williger{}), endpunkt{
		methode: http.MethodPost, weg: "/karten/4/ki-vorschlag",
		rumpf: `{"alias":"beispiel:3b","auftrag":"Schreibe."}`,
	}, gueltig)

	if antwort.Code != http.StatusOK {
		t.Fatalf("der Status ist %d: %s", antwort.Code, antwort.Body.String())
	}

	if !strings.Contains(antwort.Body.String(), `"kennung":"ki-1-1"`) {
		t.Errorf("die Antwort ist %s", antwort.Body.String())
	}
}

// **Eine Kartenkennung, die keine Zahl ist, wird am Weg abgewiesen.**
//
// **Vor der Naht und nicht in ihr**: der Weg nennt eine Kennung, und was
// dort steht, ist entweder eine oder keine.
func TestEineUnlesbareKartenkennungWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	antwort := ruf(t, kidienst(t, williger{}), endpunkt{
		methode: http.MethodPost, weg: "/karten/vier/ki-vorschlag",
		rumpf: `{"alias":"beispiel:3b","auftrag":"x"}`,
	}, gueltig)

	if antwort.Code != http.StatusBadRequest {
		t.Fatalf("der Status ist %d: %s", antwort.Code, antwort.Body.String())
	}
}

// **Ein unlesbarer Rumpf ebenso.**
func TestEinUnlesbarerRumpfBeimKIVorschlagWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	antwort := ruf(t, kidienst(t, williger{}), endpunkt{
		methode: http.MethodPost, weg: "/karten/4/ki-vorschlag",
		rumpf: `{kaputt`,
	}, gueltig)

	if antwort.Code != http.StatusBadRequest {
		t.Fatalf("der Status ist %d: %s", antwort.Code, antwort.Body.String())
	}
}

// **Ein Fehler der Naht kommt als Fehler durch.**
func TestEinFehlerDerKINahtKommtDurch(t *testing.T) {
	t.Parallel()

	antwort := ruf(t, kidienst(t,
		unwilliger{fehler: errors.New("die Naht schweigt")}), endpunkt{
		methode: http.MethodPost, weg: "/karten/4/ki-vorschlag",
		rumpf: `{"alias":"beispiel:3b","auftrag":"x"}`,
	}, gueltig)

	if antwort.Code == http.StatusOK {
		t.Fatalf("der Fehler kam als Erfolg durch: %s", antwort.Body.String())
	}
}

// **Und die zwei Wege der Messung weisen einen unlesbaren Rumpf ab** (A9,
// #1448).
//
// **Beide und nicht einer**: sie lesen denselben Bezug, aber aus zwei
// Händen — und eine Zeile, die nur an einer steht, fiele an der anderen aus.
func TestEinUnlesbarerRumpfBeiDerMessungWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	for name, weg := range map[string]string{
		"messen":    "/ki/modelle/pruefen",
		"vergessen": "/ki/modelle/vergessen",
		// **Und der Sitzungsschalter des Gerätemodells** (AI-05, #2042).
		"sitzung": "/ki/geraetemodell/sitzung",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			antwort := ruf(t, kidienst(t, williger{}), endpunkt{
				methode: http.MethodPost, weg: weg, rumpf: `{kaputt`,
			}, gueltig)

			if antwort.Code != http.StatusBadRequest {
				t.Fatalf("der Status ist %d: %s", antwort.Code, antwort.Body.String())
			}
		})
	}
}

// **Ein Fehler der Naht kommt auch bei der Messung durch.**
func TestEinFehlerDerMessungKommtDurch(t *testing.T) {
	t.Parallel()

	for name, weg := range map[string]string{
		"messen":    "/ki/modelle/pruefen",
		"vergessen": "/ki/modelle/vergessen",
		"sitzung":   "/ki/geraetemodell/sitzung",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			antwort := ruf(t, kidienst(t,
				unwilliger{fehler: errors.New("die Naht schweigt")}), endpunkt{
				methode: http.MethodPost, weg: weg,
				rumpf: `{"anbieter":"openai","alias":"gpt-probe"}`,
			}, gueltig)

			if antwort.Code == http.StatusOK {
				t.Fatalf("der Fehler kam nicht durch: %s", antwort.Body.String())
			}
		})
	}
}
