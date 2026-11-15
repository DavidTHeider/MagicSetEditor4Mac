package rest_test

// **Eine Datei ansehen braucht einen Pfad** — UX-13
// (#1685).

import (
	"net/http"
	"testing"
)

func TestOhnePfadGibtEsKeineAnsicht(t *testing.T) {
	t.Parallel()

	gezaehlt := &zaehlender{}

	antwort := ruf(t, dienstMit(t, gezaehlt), endpunkt{
		name: "eine Datei ansehen", methode: http.MethodGet, weg: "/dateien/ansicht",
	}, gueltig)

	if antwort.Code != http.StatusBadRequest {
		t.Errorf("Status %d: %s", antwort.Code, antwort.Body.String())
	}

	// **Und die Naht wurde gar nicht erst gefragt.**
	if gezaehlt.rufe != 0 {
		t.Errorf("die Naht wurde %d-mal gerufen", gezaehlt.rufe)
	}
}
