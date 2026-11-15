package rest_test

// **Eine Ausgabe gehört dem, der sie erzeugt hat** — SEC-P6
// (#2002),
// Befund P6 in `docs/PENTEST.md`.
//
// # Wie bei den Aufträgen
//
// **Der Principal, der eine Ausgabe anstößt, holt sie ab** — kein anderer,
// auch nicht die Prozessmarke. Ein anderer bekommt `404` wie für einen Namen,
// den es nicht gibt (`auftraege.go`, `eigener`): ein `403` verriete, dass es
// ihn gibt. Bis #2002 war der gewürfelte Name das einzige Recht.

import (
	"net/http"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// automat ist der Nachweis des Tokens in diesem Lauf.
const automat = "Bearer mse_3_ABCDEFGH"

// besitzdienst ist der Übertragungsdienst mit einer Tokenwache daneben.
func besitzdienst(ablage *ablageprobe) http.Handler {
	return rest.NeuerDienst(probemarke, time.Now, rest.Naehte{
		Leser: williger{}, Ausfuhr: &merkendeAusfuhr{}, Setvorschau: williger{}, Dateiablage: ablage,
		Auftraege: rest.NeuesAuftragsbuch(time.Now),
	}).MitGrenzen(domain.Vorgabegrenzen()).
		MitTokenwache(&tokenwache{gilt: "mse_3_ABCDEFGH", profilID: 42}).Behandler()
}

func TestEineAusgabeGehoertIhremErzeuger(t *testing.T) {
	ablage := neueAblageprobe()
	dienst := besitzdienst(ablage)

	ort := schicke(t, dienst, http.MethodPost, "/sets/3/bogen", "application/json", []byte(`{}`)).
		Header().Get("Location")
	if ort == "" {
		t.Fatal("die Ausgabe nennt keinen Ort")
	}

	if fremd := schicke(t, dienst, http.MethodGet, ort, "", nil, "Authorization", automat); fremd.Code != http.StatusNotFound {
		t.Errorf("ein anderer Principal holte ab: %d %q", fremd.Code, fremd.Body.String())
	}

	if fremd := schicke(t, dienst, http.MethodDelete, ort, "", nil, "Authorization", automat); fremd.Code != http.StatusNotFound {
		t.Errorf("ein anderer Principal vergaß: %d", fremd.Code)
	}

	// **Die Gegenprobe**: der Erzeuger holt ab — die Ausgabe liegt also noch.
	if eigen := schicke(t, dienst, http.MethodGet, ort, "", nil); eigen.Code != http.StatusOK {
		t.Errorf("der Erzeuger bekam %d", eigen.Code)
	}

	// **Und umgekehrt**: was das Token anstößt, holt die Prozessmarke nicht.
	vomAutomaten := schicke(t, dienst, http.MethodPost, "/sets/3/bogen", "application/json",
		[]byte(`{}`), "Authorization", automat).Header().Get("Location")

	if marke := schicke(t, dienst, http.MethodGet, vomAutomaten, "", nil); marke.Code != http.StatusNotFound {
		t.Errorf("die Prozessmarke holte die Ausgabe des Tokens ab: %d", marke.Code)
	}

	if eigen := schicke(t, dienst, http.MethodGet, vomAutomaten, "", nil, "Authorization", automat); eigen.Code != http.StatusOK {
		t.Errorf("das Token bekam seine eigene Ausgabe nicht: %d", eigen.Code)
	}
}

// **Auch als Auftrag** — der Lauf trägt den Principal dessen, der ihn
// anstieß, und die Ausgabe darin gehört ihm.
func TestEineAusgabeAusEinemAuftragGehoertSeinemErzeuger(t *testing.T) {
	dienst := besitzdienst(neueAblageprobe())

	auftrag := alsAuftrag(t, schicke(t, dienst, http.MethodPost, "/sets/3/bogen", "application/json",
		[]byte(`{}`), "Prefer", "respond-async", "Authorization", automat))

	fertigFuerDenAutomaten(t, dienst, auftrag.ID)

	ort := schicke(t, dienst, http.MethodGet, "/auftraege/"+auftrag.ID+"/ergebnis", "", nil,
		"Authorization", automat).Header().Get("Location")

	if marke := schicke(t, dienst, http.MethodGet, ort, "", nil); marke.Code != http.StatusNotFound {
		t.Errorf("die Prozessmarke holte die Ausgabe eines fremden Auftrags ab: %d", marke.Code)
	}

	if eigen := schicke(t, dienst, http.MethodGet, ort, "", nil, "Authorization", automat); eigen.Code != http.StatusOK {
		t.Errorf("der Erzeuger bekam %d", eigen.Code)
	}
}

// fertigFuerDenAutomaten fragt wie [bisZumStand], aber mit dem Nachweis des
// Automaten — für die Prozessmarke ist sein Auftrag ein fremder (`404`).
func fertigFuerDenAutomaten(t *testing.T, dienst http.Handler, id string) {
	t.Helper()

	for range 400 {
		if alsAuftrag(t, schicke(t, dienst, http.MethodGet, "/auftraege/"+id, "", nil,
			"Authorization", automat)).Stand == "fertig" {
			return
		}

		time.Sleep(5 * time.Millisecond)
	}

	t.Fatalf("der Auftrag %s wird nicht fertig", id)
}
