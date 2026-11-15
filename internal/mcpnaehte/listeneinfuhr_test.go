package mcpnaehte

// **MCP bekommt dieselben Angaben einer Listenzeile wie Fenster und REST** —
// Befund 9 des Zweitblicks über Phase G, zweiter Teil (G01,
// #2248).
//
// Die Zeile trägt seit G01, ob es mehr Drucke gibt als eine Seite, und ihren
// Ausfall als übersetzbaren Fehler. Das DTO von `internal/mcp` übernahm
// beides nicht: ein Sprachmodell las nur den deutschen Befund.

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

func TestDieListenzeileTraegtWeitereUndAusfall(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	dienst := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.RawQuery, "Ausfall") {
			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		karte := `{"object":"card","id":"00000000-0000-4000-8000-000000000001","name":"Viele","lang":"en",` +
			`"layout":"normal","set":"probe","set_name":"Probesatz","collector_number":"1","rarity":"common","type_line":"Instant"}`
		_, _ = w.Write([]byte(`{"object":"list","has_more":true,"total_cards":200,"data":[` + karte + `]}`)) //nolint:errcheck // ein Testserver
	}))
	t.Cleanup(dienst.Close)

	vorher := app.MitScryfalladresse(dienst.URL)
	t.Cleanup(func() { app.MitScryfalladresse(vorher) })

	deutung, err := listeneinfuhr{db: frischerBestand(t)}.Deuten(probekontext(t), "1 Viele\n1 Ausfall")
	if err != nil {
		t.Fatalf("Deuten: %v", err)
	}

	viele, ausfall := deutung.Zeilen[0], deutung.Zeilen[1]
	if !viele.Weitere || viele.Fehler != nil {
		t.Errorf("Viele: weitere %v, Fehler %+v", viele.Weitere, viele.Fehler)
	}

	if ausfall.Fehler == nil || ausfall.Fehler.Schluessel == "" || ausfall.Fehler.Meldung == "" {
		t.Errorf("Ausfall: Fehler %+v", ausfall.Fehler)
	}
}
