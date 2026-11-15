package rest_test

// **Lange Datei-Läufe als Auftrag** — FMT-03.AK-02
// (#1692).
//
// Wer mit `Prefer: respond-async` (RFC 7240, Abschnitt 4.1) einen Bogen legt,
// ein Set einfährt oder einen Installer schreibt, bekommt `202` und den Ort
// des Auftrags; dort stehen Stand, Fortschritt und — wenn er fertig ist — die
// Antwort, die der Weg sonst gleich gegeben hätte. Abbrechen ist `DELETE`,
// wiederholen ein `POST` an `…/wiederholung`.

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

type auftragsmelder struct{}

// langsamerBogen meldet eine Karte von zweien und wartet dann auf `weiter`
// — oder auf den Abbruch.
type langsamerBogen struct {
	williger

	gemeldet chan struct{}
	weiter   chan struct{}
	panik    bool
}

func (b *langsamerBogen) Bogen(ctx context.Context, _ int64, _ string) (seiten, karten int, befunde []string, fehler error) {
	if b.panik {
		panic("der Bogen ist zerrissen")
	}

	if melder, gehoert := ctx.Value(auftragsmelder{}).(domain.Fortschrittsmelder); gehoert {
		melder(domain.Fortschritt{Erledigt: 1, Gesamt: 2, Einheit: domain.EinheitKarte})
	}

	b.gemeldet <- struct{}{}

	select {
	case <-b.weiter:
		return 1, 2, nil, nil
	case <-ctx.Done():
		return 0, 0, nil, ctx.Err()
	}
}

func auftragsdienst(t *testing.T, bogen *langsamerBogen) http.Handler {
	t.Helper()

	jetzt := time.Date(2026, time.September, 29, 12, 0, 0, 0, time.UTC)
	naehte := rest.Naehte{Leser: williger{}, Ausfuhr: bogen, Auftraege: rest.NeuesAuftragsbuch(time.Now)}

	return rest.NeuerDienst(probemarke, func() time.Time { return jetzt }, naehte).
		MitFortschrittsnaht(func(ctx context.Context, melder domain.Fortschrittsmelder) context.Context {
			return context.WithValue(ctx, auftragsmelder{}, melder)
		}).Behandler()
}

func auftragsruf(t *testing.T, dienst http.Handler, methode, weg, rumpf string, asynchron bool) *httptest.ResponseRecorder {
	t.Helper()

	anfrage := httptest.NewRequestWithContext(t.Context(), methode, weg, http.NoBody)

	if rumpf != "" {
		anfrage = httptest.NewRequestWithContext(t.Context(), methode, weg, strings.NewReader(rumpf))
	}

	gueltig(anfrage)

	if asynchron {
		anfrage.Header.Set("Prefer", "wait=5, respond-async")
	}

	antwort := httptest.NewRecorder()
	dienst.ServeHTTP(antwort, anfrage)

	return antwort
}

type auftragsdto struct {
	ID              string `json:"id"`
	Weg             string `json:"weg"`
	Stand           string `json:"stand"`
	Status          int    `json:"status"`
	WiederholungVon string `json:"wiederholungVon"`
	Fortschritt     struct {
		Erledigt int    `json:"erledigt"`
		Gesamt   int    `json:"gesamt"`
		Satz     string `json:"satz"`
	} `json:"fortschritt"`
}

func alsAuftrag(t *testing.T, antwort *httptest.ResponseRecorder) auftragsdto {
	t.Helper()

	var umschlag struct {
		Auftrag auftragsdto `json:"auftrag"`
	}

	if err := json.Unmarshal(antwort.Body.Bytes(), &umschlag); err != nil || umschlag.Auftrag.ID == "" {
		t.Fatalf("kein Auftrag: %v\n%s", err, antwort.Body.String())
	}

	return umschlag.Auftrag
}

// bisZumStand fragt, bis der Auftrag den Stand erreicht — auf die Wirkung
// gewartet, nicht auf den Ruf (§3).
func bisZumStand(t *testing.T, dienst http.Handler, id, stand string) auftragsdto {
	t.Helper()

	for range 400 {
		if gelesen := alsAuftrag(t, auftragsruf(t, dienst, http.MethodGet, "/auftraege/"+id, "", false)); gelesen.Stand == stand {
			return gelesen
		}

		time.Sleep(5 * time.Millisecond)
	}

	t.Fatalf("der Auftrag %s erreicht %q nicht", id, stand)

	return auftragsdto{}
}

func neuerBogen() *langsamerBogen {
	return &langsamerBogen{gemeldet: make(chan struct{}, 4), weiter: make(chan struct{})}
}

func TestEinBogenLaeuftAlsAuftrag(t *testing.T) {
	bogen := neuerBogen()
	dienst := auftragsdienst(t, bogen)

	angenommen := auftragsruf(t, dienst, http.MethodPost, "/sets/3/bogen", `{"ziel":"/tmp/b.pdf"}`, true)
	if angenommen.Code != http.StatusAccepted || angenommen.Header().Get("Preference-Applied") != "respond-async" {
		t.Fatalf("POST: %d %v (%s)", angenommen.Code, angenommen.Header(), angenommen.Body.String())
	}

	auftrag := alsAuftrag(t, angenommen)
	if angenommen.Header().Get("Location") != "/auftraege/"+auftrag.ID || auftrag.Weg != "POST /sets/{setID}/bogen" {
		t.Errorf("Ort %q, Weg %q", angenommen.Header().Get("Location"), auftrag.Weg)
	}

	<-bogen.gemeldet

	laufend := bisZumStand(t, dienst, auftrag.ID, "laeuft")
	if laufend.Fortschritt.Erledigt != 1 || laufend.Fortschritt.Gesamt != 2 || laufend.Fortschritt.Satz != "1 von 2 Karten" {
		t.Errorf("Fortschritt: %+v", laufend.Fortschritt)
	}

	if frueh := auftragsruf(t, dienst, http.MethodGet, "/auftraege/"+auftrag.ID+"/ergebnis", "", false); frueh.Code != http.StatusConflict {
		t.Errorf("Ergebnis vor dem Ende: %d", frueh.Code)
	}

	close(bogen.weiter)

	if fertig := bisZumStand(t, dienst, auftrag.ID, "fertig"); fertig.Status != http.StatusCreated {
		t.Errorf("fertig: %+v", fertig)
	}

	// **Das Ergebnis ist die Antwort, die der Weg gleich gegeben hätte.**
	ergebnis := auftragsruf(t, dienst, http.MethodGet, "/auftraege/"+auftrag.ID+"/ergebnis", "", false)
	if ergebnis.Code != http.StatusCreated || !strings.Contains(ergebnis.Body.String(), `"seiten":1`) ||
		!strings.HasPrefix(ergebnis.Header().Get("Content-Type"), "application/json") {
		t.Errorf("Ergebnis: %d %s", ergebnis.Code, ergebnis.Body.String())
	}

	liste := auftragsruf(t, dienst, http.MethodGet, "/auftraege", "", false)
	if liste.Code != http.StatusOK || !strings.Contains(liste.Body.String(), auftrag.ID) {
		t.Errorf("Liste: %d %s", liste.Code, liste.Body.String())
	}
}

func TestEinAuftragLaesstSichAbbrechenUndWiederholen(t *testing.T) {
	bogen := neuerBogen()
	dienst := auftragsdienst(t, bogen)

	auftrag := alsAuftrag(t, auftragsruf(t, dienst, http.MethodPost, "/sets/3/bogen", `{"ziel":"/tmp/b.pdf"}`, true))
	<-bogen.gemeldet

	if abbruch := auftragsruf(t, dienst, http.MethodDelete, "/auftraege/"+auftrag.ID, "", false); abbruch.Code != http.StatusAccepted {
		t.Fatalf("DELETE: %d %s", abbruch.Code, abbruch.Body.String())
	}

	if abgebrochen := bisZumStand(t, dienst, auftrag.ID, "abgebrochen"); abgebrochen.Status != 499 {
		t.Errorf("abgebrochen: %+v", abgebrochen)
	}

	// **Eine Wiederholung ist ein neuer Auftrag** mit derselben Anfrage —
	// und sie sagt, woher sie kommt.
	wieder := auftragsruf(t, dienst, http.MethodPost, "/auftraege/"+auftrag.ID+"/wiederholung", "", false)
	if wieder.Code != http.StatusAccepted {
		t.Fatalf("Wiederholung: %d %s", wieder.Code, wieder.Body.String())
	}

	zweiter := alsAuftrag(t, wieder)
	if zweiter.ID == auftrag.ID || zweiter.WiederholungVon != auftrag.ID {
		t.Errorf("zweiter: %+v", zweiter)
	}

	<-bogen.gemeldet
	close(bogen.weiter)
	bisZumStand(t, dienst, zweiter.ID, "fertig")

	// **Ein beendeter Auftrag wird mit `DELETE` vergessen.**
	if weg := auftragsruf(t, dienst, http.MethodDelete, "/auftraege/"+zweiter.ID, "", false); weg.Code != http.StatusNoContent {
		t.Errorf("vergessen: %d", weg.Code)
	}

	if weg := auftragsruf(t, dienst, http.MethodGet, "/auftraege/"+zweiter.ID, "", false); weg.Code != http.StatusNotFound {
		t.Errorf("danach: %d", weg.Code)
	}
}

// **Ein laufender Auftrag wird nicht wiederholt** — zwei gleiche Läufe
// zugleich wären zwei Schreiber an einem Ziel.
func TestEinLaufenderAuftragWirdNichtWiederholt(t *testing.T) {
	bogen := neuerBogen()
	dienst := auftragsdienst(t, bogen)

	auftrag := alsAuftrag(t, auftragsruf(t, dienst, http.MethodPost, "/sets/3/bogen", `{"ziel":"/tmp/b.pdf"}`, true))
	<-bogen.gemeldet

	if wieder := auftragsruf(t, dienst, http.MethodPost, "/auftraege/"+auftrag.ID+"/wiederholung", "", false); wieder.Code != http.StatusConflict {
		t.Errorf("Wiederholung: %d", wieder.Code)
	}

	close(bogen.weiter)
	bisZumStand(t, dienst, auftrag.ID, "fertig")
}

// **Ein Weg, der kein Auftrag sein kann, antwortet gleich** — die Bitte ist
// eine Bitte (RFC 7240) und kein Befehl.
func TestEinKurzerWegAntwortetTrotzBitteGleich(t *testing.T) {
	dienst := auftragsdienst(t, neuerBogen())

	antwort := auftragsruf(t, dienst, http.MethodGet, "/sets", "", true)
	if antwort.Code != http.StatusOK || antwort.Header().Get("Preference-Applied") != "" {
		t.Errorf("GET /sets: %d %v", antwort.Code, antwort.Header())
	}
}

func TestEinUnbekannterAuftragIstNichtDa(t *testing.T) {
	dienst := auftragsdienst(t, neuerBogen())

	for _, weg := range []struct{ methode, pfad string }{
		{http.MethodGet, "/auftraege/gibtsnicht"},
		{http.MethodGet, "/auftraege/gibtsnicht/ergebnis"},
		{http.MethodDelete, "/auftraege/gibtsnicht"},
		{http.MethodPost, "/auftraege/gibtsnicht/wiederholung"},
	} {
		if antwort := auftragsruf(t, dienst, weg.methode, weg.pfad, "", false); antwort.Code != http.StatusNotFound {
			t.Errorf("%s %s: %d", weg.methode, weg.pfad, antwort.Code)
		}
	}
}

// **Ein Auftrag, der zusammenbricht, reißt den Dienst nicht mit** — er läuft
// in einer eigenen Goroutine, und dort fängt `net/http` nichts mehr.
func TestEinZusammenbrechenderAuftragScheitert(t *testing.T) {
	bogen := neuerBogen()
	bogen.panik = true
	dienst := auftragsdienst(t, bogen)

	auftrag := alsAuftrag(t, auftragsruf(t, dienst, http.MethodPost, "/sets/3/bogen", `{"ziel":"/tmp/b.pdf"}`, true))

	if gescheitert := bisZumStand(t, dienst, auftrag.ID, "gescheitert"); gescheitert.Status != http.StatusInternalServerError {
		t.Errorf("gescheitert: %+v", gescheitert)
	}
}

// **Mehr als vier laufende Aufträge nimmt der Dienst nicht an** — jeder hält
// Speicher für seine Antwort, und die Fachläufe dahinter gehen ohnehin einer
// nach dem anderen.
func TestZuVieleAuftraegeWerdenAbgewiesen(t *testing.T) {
	bogen := neuerBogen()
	bogen.gemeldet = make(chan struct{}, 8)
	dienst := auftragsdienst(t, bogen)

	for range 4 {
		alsAuftrag(t, auftragsruf(t, dienst, http.MethodPost, "/sets/3/bogen", `{"ziel":"/tmp/b.pdf"}`, true))
	}

	fuenfter := auftragsruf(t, dienst, http.MethodPost, "/sets/3/bogen", `{"ziel":"/tmp/b.pdf"}`, true)
	if fuenfter.Code != http.StatusTooManyRequests || fuenfter.Header().Get("Retry-After") == "" {
		t.Errorf("fünfter: %d %v", fuenfter.Code, fuenfter.Header())
	}

	close(bogen.weiter)
}

// **Ein Rumpf über der Grenze wird auch als Auftrag abgewiesen.**
func TestEinZuGrosserAuftragsrumpfWirdAbgewiesen(t *testing.T) {
	naehte := rest.Naehte{Leser: williger{}, Ausfuhr: neuerBogen(), Auftraege: rest.NeuesAuftragsbuch(time.Now)}
	grenzen := domain.Vorgabegrenzen()
	grenzen.Antwort = 200

	dienst := rest.NeuerDienst(probemarke, time.Now, naehte).MitGrenzen(grenzen).Behandler()

	antwort := auftragsruf(t, dienst, http.MethodPost, "/sets/3/bogen",
		`{"ziel":"/tmp/`+strings.Repeat("x", 300)+`.pdf"}`, true)
	if antwort.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("Status %d: %s", antwort.Code, antwort.Body.String())
	}
}

// **Eine Wiederholung geht ein zweites Mal durch den Pfadvertrag** — er kann
// sich seit dem ersten Lauf geändert haben.
func TestEineWiederholungFragtDenPfadvertragNeu(t *testing.T) {
	bogen := neuerBogen()
	close(bogen.weiter)

	gesperrt := false
	naehte := rest.Naehte{Leser: williger{}, Ausfuhr: bogen, Auftraege: rest.NeuesAuftragsbuch(time.Now)}
	dienst := rest.NeuerDienst(probemarke, time.Now, naehte).
		MitPfadvertrag(func(_ context.Context, pfad string) error {
			if gesperrt {
				return domain.NewFieldError("pfad", domain.SchluesselValidierung, "%s ist gesperrt", pfad)
			}

			return nil
		}).Behandler()

	auftrag := alsAuftrag(t, auftragsruf(t, dienst, http.MethodPost, "/sets/3/bogen", `{"ziel":"/tmp/b.pdf"}`, true))
	<-bogen.gemeldet
	bisZumStand(t, dienst, auftrag.ID, "fertig")

	gesperrt = true
	zweiter := alsAuftrag(t, auftragsruf(t, dienst, http.MethodPost, "/auftraege/"+auftrag.ID+"/wiederholung", "", false))

	if gescheitert := bisZumStand(t, dienst, zweiter.ID, "gescheitert"); gescheitert.Status != http.StatusBadRequest {
		t.Errorf("zweiter: %+v", gescheitert)
	}
}
