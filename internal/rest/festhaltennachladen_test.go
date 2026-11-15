package rest_test

// **Festgehalten ist festgehalten — auch wenn die Liste fehlt** — §7,
// Regel 3 (#1970).
//
// Der Rumpf eines Festhaltens ist eine Liste; der Grund, aus dem sie nach
// geglücktem Schreiben leer ist, steht deshalb im Kopf [rest.Nachladekopf].

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// nachladenderWeg hält fest und kann danach nicht nachlesen.
type nachladenderWeg struct{ williger }

func (nachladenderWeg) HandelFesthalten(
	context.Context, string, int64, int, int64,
) (rest.Festgehalten, error) {
	return rest.Festgehalten{
		Zusagen:        []rest.Handelszusage{},
		Nachladefehler: &rest.Befund{Code: "validierung", Meldung: "krumm"},
	}, nil
}

func TestEinFesthaltenOhneNachlesenNenntDenGrundImKopf(t *testing.T) {
	t.Parallel()

	dienst := rest.NeuerDienst(probemarke, time.Now, rest.Naehte{
		Leser: williger{}, Leitung: nachladenderWeg{},
	})

	antwort := ruf(t, dienst, endpunkt{
		methode: http.MethodPost, weg: "/handel/v-1/zusagen",
		rumpf: `{"exemplarId":7,"menge":1,"fassung":1}`,
	}, gueltig)

	switch {
	case antwort.Code != http.StatusCreated:
		t.Fatalf("Status %d: %s", antwort.Code, antwort.Body.String())
	case antwort.Header().Get(rest.Nachladekopf) != "validierung":
		t.Errorf("der Kopf sagt %q", antwort.Header().Get(rest.Nachladekopf))
	case antwort.Body.String() != "[]\n":
		t.Errorf("der Rumpf ist %q", antwort.Body.String())
	}

	// **Ohne Nachladefehler kein Kopf** — er ist eine Ausnahme und keine
	// leere Angabe an jeder Antwort.
	ohne := ruf(t, rest.NeuerDienst(probemarke, time.Now, rest.Naehte{
		Leser: williger{}, Leitung: williger{},
	}), endpunkt{
		methode: http.MethodPost, weg: "/handel/v-1/zusagen",
		rumpf: `{"exemplarId":7,"menge":1,"fassung":1}`,
	}, gueltig)

	if kopf := ohne.Header().Get(rest.Nachladekopf); ohne.Code != http.StatusCreated || kopf != "" {
		t.Errorf("ohne Nachladefehler: %d, %q", ohne.Code, kopf)
	}
}
