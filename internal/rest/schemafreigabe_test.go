package rest_test

// **Jede Anfrage hält die Schemafreigabe** — Zweitblick über Phase 2 von
// 0.9.0, Befunde 1 und 2 (#2323). Die Fassungswache stand in der
// Schreibwache: hinter dem Nachweis, der schon `letzte_nutzung` schrieb, nicht
// vor den beiden Wechseln, und nicht im Lauf eines gelösten Auftrags. Und sie
// fragte nur, statt die Antwort zu halten — eine Wanderung konnte zwischen
// Frage und Schreiben festschreiben.

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// auftragsdienstMit ist [auftragsdienst] mit einer Schemafreigabe.
func auftragsdienstMit(t *testing.T, bogen *langsamerBogen, freigabe func(context.Context) (func(), error)) http.Handler {
	t.Helper()

	jetzt := time.Date(2026, time.September, 29, 12, 0, 0, 0, time.UTC)
	naehte := rest.Naehte{Leser: williger{}, Ausfuhr: bogen, Auftraege: rest.NeuesAuftragsbuch(time.Now)}

	return rest.NeuerDienst(probemarke, func() time.Time { return jetzt }, naehte).
		MitFortschrittsnaht(func(ctx context.Context, melder domain.Fortschrittsmelder) context.Context {
			return context.WithValue(ctx, auftragsmelder{}, melder)
		}).MitSchemafreigabe(freigabe).Behandler()
}

// freigabezaehler zählt geholte und gehaltene Freigaben.
type freigabezaehler struct {
	geholt, gehalten atomic.Int32
	nein             error
}

func (z *freigabezaehler) halten(context.Context) (func(), error) {
	if z.nein != nil {
		return nil, z.nein
	}

	z.geholt.Add(1)
	z.gehalten.Add(1)

	return func() { z.gehalten.Add(-1) }, nil
}

// **Ohne Freigabe kommt keine Anfrage an** — lesend, schreibend, die zwei
// Wechsel, und auch ohne Nachweis: gefragt wird vor ihm.
func TestOhneSchemafreigabeKommtKeineAnfrageAn(t *testing.T) {
	t.Parallel()

	zaehler := &freigabezaehler{nein: domain.MitGrund(domain.ErrBelegt, errors.New("gewandert"), "gleich noch einmal")}
	dienst := neuerDienst(t).MitSchemafreigabe(zaehler.halten)

	faelle := []struct {
		ziel endpunkt
		wie  gestalt
	}{
		{endpunkt{name: "lesen", methode: http.MethodGet, weg: "/sets"}, gueltig},
		{endpunkt{name: "schreiben", methode: http.MethodPost, weg: "/sets/1/karten", rumpf: `{}`}, gueltig},
		{endpunkt{name: "Profil wechseln", methode: http.MethodPut, weg: "/profil", rumpf: `{"profilID":"2"}`}, gueltig},
		{endpunkt{name: "Nutzer wechseln", methode: http.MethodPut, weg: "/nutzer", rumpf: `{"nutzerID":"2"}`}, gueltig},
		{endpunkt{name: "ohne Nachweis", methode: http.MethodGet, weg: "/sets"}, ohneMarke},
	}

	for _, fall := range faelle {
		antwort := ruf(t, dienst, fall.ziel, fall.wie)

		if antwort.Code != http.StatusServiceUnavailable || alsBefund(t, antwort).Code != "belegt" {
			t.Errorf("%s: %d %s, erwartet 503 belegt", fall.ziel.name, antwort.Code, antwort.Body.String())
		}
	}
}

// **Gehalten wird bis zum Ende der Anfrage** — die Schreibwache läuft
// darunter, und danach ist nichts mehr gehalten.
func TestDieSchemafreigabeReichtUeberDieAnfrage(t *testing.T) {
	t.Parallel()

	zaehler := &freigabezaehler{}

	var waehrend atomic.Int32

	dienst := neuerDienst(t).MitSchemafreigabe(zaehler.halten).MitSchreibwache(func(context.Context) error {
		waehrend.Store(zaehler.gehalten.Load())

		return nil
	})

	ruf(t, dienst, endpunkt{name: "schreiben", methode: http.MethodPost, weg: "/sets/1/karten", rumpf: `{}`}, gueltig)

	if waehrend.Load() != 1 || zaehler.gehalten.Load() != 0 || zaehler.geholt.Load() != 1 {
		t.Errorf("während %d, danach %d, geholt %d", waehrend.Load(), zaehler.gehalten.Load(), zaehler.geholt.Load())
	}
}

// **Und ein gelöster Auftrag holt sie für seinen Lauf** — er läuft, nachdem
// die Anfrage längst vorbei ist.
func TestEinAuftragHaeltDieSchemafreigabeWaehrendErLaeuft(t *testing.T) {
	t.Parallel()

	zaehler := &freigabezaehler{}
	bogen := neuerBogen()
	dienst := auftragsdienstMit(t, bogen, zaehler.halten)

	angenommen := auftragsruf(t, dienst, http.MethodPost, "/sets/3/bogen", `{"ziel":"/tmp/b.pdf"}`, true)
	if angenommen.Code != http.StatusAccepted {
		t.Fatalf("POST: %d %s", angenommen.Code, angenommen.Body.String())
	}

	auftrag := alsAuftrag(t, angenommen)

	<-bogen.gemeldet

	if zaehler.gehalten.Load() != 1 {
		t.Errorf("während der Auftrag läuft, gehalten %d", zaehler.gehalten.Load())
	}

	close(bogen.weiter)
	bisZumStand(t, dienst, auftrag.ID, "fertig")

	if zaehler.gehalten.Load() != 0 {
		t.Errorf("nach dem Auftrag gehalten %d", zaehler.gehalten.Load())
	}
}

// **Ohne Freigabe läuft auch ein Auftrag nicht** — sein Ergebnis ist der
// Fehler der Freigabe.
func TestEinAuftragOhneSchemafreigabeScheitert(t *testing.T) {
	t.Parallel()

	zaehler := &freigabezaehler{}
	bogen := neuerBogen()
	dienst := auftragsdienstMit(t, bogen, func(ctx context.Context) (func(), error) {
		// **Die Anfragen bekommen sie, der Lauf danach nicht** — dazwischen
		// hat eine neuere Fassung gewandert. Den Lauf erkennt die Probe an
		// seinem Melder.
		if ctx.Value(auftragsmelder{}) == nil {
			return zaehler.halten(ctx)
		}

		return nil, domain.MitGrund(domain.ErrBelegt, errors.New("gewandert"), "gleich noch einmal")
	})

	auftrag := alsAuftrag(t, auftragsruf(t, dienst, http.MethodPost, "/sets/3/bogen", `{"ziel":"/tmp/b.pdf"}`, true))

	if gescheitert := bisZumStand(t, dienst, auftrag.ID, "gescheitert"); gescheitert.Status != http.StatusServiceUnavailable {
		t.Errorf("der Auftrag lief ohne Freigabe: %+v", gescheitert)
	}

	select {
	case <-bogen.gemeldet:
		t.Error("der Bogen wurde gelegt")
	default:
	}
}

// **Die Freigabe ganz außen** — Nachlauf über Phase 2 von 0.9.0, Befund 3:
// `cmd/mse-rest` liest das gezeigte Profil, bevor die Anfrage den Dienst
// erreicht, und damit vor der Freigabe. [rest.Dienst.UnterFreigabe] setzt sie
// davor — und der Dienst innen holt sie dann nicht ein zweites Mal.
func TestDieFreigabeGehtVorDasProfil(t *testing.T) {
	t.Parallel()

	nein := &freigabezaehler{nein: domain.MitGrund(domain.ErrBelegt, errors.New("gewandert"), "gleich noch einmal")}
	gerufen := 0
	innen := http.HandlerFunc(func(http.ResponseWriter, *http.Request) { gerufen++ })

	antwort := httptest.NewRecorder()
	neuerDienst(t).MitSchemafreigabe(nein.halten).UnterFreigabe(innen).
		ServeHTTP(antwort, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/sets", http.NoBody))

	if antwort.Code != http.StatusServiceUnavailable || gerufen != 0 {
		t.Errorf("ohne Freigabe: %d, innen %d-mal gerufen", antwort.Code, gerufen)
	}

	zaehler := &freigabezaehler{}
	dienst := neuerDienst(t).MitSchemafreigabe(zaehler.halten)
	aussen := dienst.UnterFreigabe(dienst.Behandler())

	anfrage := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/sets", http.NoBody)
	gueltig(anfrage)

	aussen.ServeHTTP(httptest.NewRecorder(), anfrage)

	if zaehler.geholt.Load() != 1 || zaehler.gehalten.Load() != 0 {
		t.Errorf("geholt %d, danach gehalten %d — erwartet einmal und nichts", zaehler.geholt.Load(), zaehler.gehalten.Load())
	}
}

// **Ein gelöster Auftrag holt seine eigene** — auch wenn die Anfrage, die
// ihn annahm, unter der äußeren Freigabe lief: deren Kontext trägt er weiter.
func TestEinAuftragHoltSeineEigeneFreigabeAuchUnterDerAeusseren(t *testing.T) {
	t.Parallel()

	zaehler := &freigabezaehler{}
	bogen := neuerBogen()

	jetzt := time.Date(2026, time.September, 29, 12, 0, 0, 0, time.UTC)
	naehte := rest.Naehte{Leser: williger{}, Ausfuhr: bogen, Auftraege: rest.NeuesAuftragsbuch(time.Now)}
	dienst := rest.NeuerDienst(probemarke, func() time.Time { return jetzt }, naehte).MitSchemafreigabe(zaehler.halten)
	aussen := dienst.UnterFreigabe(dienst.Behandler())

	auftrag := alsAuftrag(t, auftragsruf(t, aussen, http.MethodPost, "/sets/3/bogen", `{"ziel":"/tmp/b.pdf"}`, true))

	<-bogen.gemeldet

	if zaehler.gehalten.Load() != 1 {
		t.Errorf("während der Auftrag läuft, gehalten %d", zaehler.gehalten.Load())
	}

	close(bogen.weiter)
	bisZumStand(t, aussen, auftrag.ID, "fertig")
}
