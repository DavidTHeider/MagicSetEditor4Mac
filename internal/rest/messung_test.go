package rest_test

// **Jeder Ruf wird gemessen, mit seinem Muster und nicht mit seinem Pfad** —
// OPS-05.AK-01/02
// (#1722).
// Der Pfad trüge Kennungen, und damit Nutzinhalte; das Muster ist eines von
// endlich vielen.

import (
	"context"
	"net/http"
	"sync"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

type messprotokoll struct {
	sperre    sync.Mutex
	messungen []rest.Messung
}

func (p *messprotokoll) nimm(_ context.Context, messung rest.Messung) {
	p.sperre.Lock()
	defer p.sperre.Unlock()

	p.messungen = append(p.messungen, messung)
}

func TestJederRufWirdMitSeinemMusterGemessen(t *testing.T) {
	t.Parallel()

	protokoll := &messprotokoll{}
	dienst := rest.NeuerDienst(probemarke, festeUhr, rest.Naehte{Leser: williger{}, Grenzen: williger{}}).
		MitMessung(protokoll.nimm)

	ruf(t, dienst, endpunkt{methode: http.MethodGet, weg: "/sets/1/karten"}, gueltig)
	ruf(t, dienst, endpunkt{methode: http.MethodGet, weg: "/sets"}, ohneMarke)
	ruf(t, dienst, endpunkt{methode: http.MethodGet, weg: "/gibt/es/nicht"}, gueltig)
	ruf(t, dienst, endpunkt{methode: http.MethodPost, weg: "/pflege", rumpf: "{}"}, gueltig)

	if len(protokoll.messungen) != 4 {
		t.Fatalf("%d Messungen: %+v", len(protokoll.messungen), protokoll.messungen)
	}

	erwartet := []struct {
		weg        string
		schreibend bool
		ausgang    domain.Rufausgang
	}{
		{"GET /sets/{setID}/karten", false, domain.RufGelungen},
		{"GET /sets", false, domain.RufAbgewiesen},
		{rest.UnbekannterWeg, false, domain.RufAbgewiesen},
		{"POST /pflege", true, ""},
	}

	for stelle, soll := range erwartet {
		ist := protokoll.messungen[stelle]
		if ist.Weg != soll.weg || ist.Schreibend != soll.schreibend ||
			(soll.ausgang != "" && ist.Ausgang != soll.ausgang) || ist.Dauer < 0 {
			t.Errorf("%d: %+v, erwartet %+v", stelle, ist, soll)
		}
	}
}

// **Ohne Messnaht misst niemand** — und nichts stürzt ab.
func TestOhneMessnahtLaeuftDerDienstWieVorher(t *testing.T) {
	t.Parallel()

	dienst := rest.NeuerDienst(probemarke, festeUhr, rest.Naehte{Leser: williger{}, Grenzen: williger{}})
	if antwort := ruf(t, dienst, endpunkt{methode: http.MethodGet, weg: "/sets"}, gueltig); antwort.Code != http.StatusOK {
		t.Errorf("%d", antwort.Code)
	}
}

// **Der Ausgang nach dem Status**: unter 400 gelungen, 4xx abgewiesen — der
// Aufrufer müsste anders fragen —, ab 500 gescheitert.
func TestDerAusgangFolgtDemStatus(t *testing.T) {
	t.Parallel()

	for status, erwartet := range map[int]domain.Rufausgang{
		http.StatusOK: domain.RufGelungen, http.StatusNoContent: domain.RufGelungen,
		http.StatusNotFound: domain.RufAbgewiesen, http.StatusTooManyRequests: domain.RufAbgewiesen,
		http.StatusInternalServerError: domain.RufGescheitert, http.StatusBadGateway: domain.RufGescheitert,
	} {
		if ist := rest.AusgangNachStatus(status); ist != erwartet {
			t.Errorf("%d: %q", status, ist)
		}
	}
}
