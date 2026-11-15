package rest_test

// **Die Ränder der Recherchewege** — Werkstück A16 von Phase 5.3
// (#1515).
//
// **Zwei Ränder, und beide gehören dem Dienst und nicht dem Fachkern**: eine
// Kartenkennung im Weg, die keine Zahl ist, und ein Rumpf, der kein JSON ist.
// Beide werden abgewiesen, **bevor** die Naht gerufen wird — sonst stünde ein
// halber Auftrag im Fachkern.

import (
	"net/http"
	"testing"
)

// **Eine Kartenkennung, die keine Zahl ist, wird am Weg abgewiesen.**
func TestEineUnlesbareKarteFindetNichtsAehnliches(t *testing.T) {
	t.Parallel()

	gezaehlt := &zaehlender{}

	antwort := ruf(t, dienstMit(t, gezaehlt), endpunkt{
		name: "Ähnliche Karten", methode: http.MethodGet,
		weg: "/recherche/aehnlich/keinezahl",
	}, gueltig)

	if antwort.Code != http.StatusBadRequest {
		t.Fatalf("die unlesbare Karte gab %d", antwort.Code)
	}

	if gezaehlt.wieOft() != 0 {
		t.Errorf("die Naht wurde %dmal gerufen", gezaehlt.wieOft())
	}
}

// **Ein unlesbarer Rumpf stellt kein Kontextpaket zusammen.**
func TestEinUnlesbarerRumpfStelltKeinPaketZusammen(t *testing.T) {
	t.Parallel()

	gezaehlt := &zaehlender{}

	antwort := ruf(t, dienstMit(t, gezaehlt), endpunkt{
		name: "Kontextpaket", methode: http.MethodPost,
		weg: "/kontextpaket", rumpf: `{`,
	}, gueltig)

	if antwort.Code != http.StatusBadRequest {
		t.Fatalf("der kaputte Rumpf gab %d", antwort.Code)
	}

	if gezaehlt.wieOft() != 0 {
		t.Errorf("die Naht wurde %dmal gerufen", gezaehlt.wieOft())
	}
}
