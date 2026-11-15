package rest_test

// **Die Prüfung eines Sets läuft als Auftrag** — Befund 11 des Zweitblicks
// über Phase 3 von 0.9.0 (PERF02, #2352): sie lief nur im Ruf, und ein Set,
// dessen Prüfung die Schreibfrist des Dienstes überdauert, bekam seine
// Antwort nie.

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// langsamePruefung meldet eine Karte von dreien und wartet auf `weiter`.
type langsamePruefung struct {
	gemeldet chan struct{}
	weiter   chan struct{}
}

func (p langsamePruefung) Problemstand(context.Context, int64) (rest.Problemstand, error) {
	return rest.Problemstand{Urteil: "ungeprueft"}, nil
}

func (p langsamePruefung) ProblemePruefen(ctx context.Context, _ int64) (rest.Problemstand, error) {
	if melder, gehoert := ctx.Value(auftragsmelder{}).(domain.Fortschrittsmelder); gehoert {
		melder(domain.Fortschritt{Erledigt: 1, Gesamt: 3, Einheit: domain.EinheitKarte})
	}

	p.gemeldet <- struct{}{}

	select {
	case <-p.weiter:
		return rest.Problemstand{Urteil: "aktuell", Sicher: true, Geprueft: 3, Gesamt: 3}, nil
	case <-ctx.Done():
		return rest.Problemstand{}, ctx.Err()
	}
}

func TestDiePruefungEinesSetsLaeuftAlsAuftrag(t *testing.T) {
	pruefung := langsamePruefung{gemeldet: make(chan struct{}, 1), weiter: make(chan struct{})}
	dienst := rest.NeuerDienst(probemarke, time.Now, rest.Naehte{
		Leser: williger{}, Probleme: pruefung, Auftraege: rest.NeuesAuftragsbuch(time.Now),
	}).MitFortschrittsnaht(func(ctx context.Context, melder domain.Fortschrittsmelder) context.Context {
		return context.WithValue(ctx, auftragsmelder{}, melder)
	}).Behandler()

	angenommen := auftragsruf(t, dienst, http.MethodPost, "/sets/3/probleme", "", true)
	if angenommen.Code != http.StatusAccepted {
		t.Fatalf("POST: %d (%s)", angenommen.Code, angenommen.Body.String())
	}

	auftrag := alsAuftrag(t, angenommen)
	<-pruefung.gemeldet

	if laufend := bisZumStand(t, dienst, auftrag.ID, "laeuft"); laufend.Fortschritt.Erledigt != 1 || laufend.Fortschritt.Gesamt != 3 {
		t.Errorf("Fortschritt %+v", laufend.Fortschritt)
	}

	close(pruefung.weiter)

	if fertig := bisZumStand(t, dienst, auftrag.ID, "fertig"); fertig.Status != http.StatusOK {
		t.Errorf("fertig mit %d", fertig.Status)
	}
}
