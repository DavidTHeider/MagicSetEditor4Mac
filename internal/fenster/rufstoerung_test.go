package fenster

// **Störungen am Rufbuch: eine Bindung verzögern oder fachlich scheitern
// lassen** — B03 (#2222).
//
// Die Zustände „laden" und „Fehler" jeder Seite lassen sich so am gebauten
// Programm herstellen, ohne eine Seite anzufassen.

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestEineVerzoegerteBindungAntwortetSpaeterUndUnveraendert(t *testing.T) {
	t.Parallel()

	buch := neuesRufbuch()
	buch.Stoeren([]Stoerung{{Bindung: "Sets", VerzoegerungMs: 150}})

	antwort := httptest.NewRecorder()
	beginn := time.Now()
	buch.Huelle(antwortend(t, http.StatusOK, `{"sets":[]}`)).ServeHTTP(antwort, bindungsruf(t, "Sets"))

	if dauer := time.Since(beginn); dauer < 150*time.Millisecond {
		t.Errorf("die Antwort kam nach %v, verzögert war um 150 ms", dauer)
	}

	if antwort.Body.String() != `{"sets":[]}` {
		t.Errorf("die Antwort war %q", antwort.Body.String())
	}
}

func TestEineGestoerteBindungScheitertFachlichOhneDurchzugehen(t *testing.T) {
	t.Parallel()

	buch := neuesRufbuch()
	buch.Stoeren([]Stoerung{{Bindung: "Sets", Fehler: true}})

	durch := false
	innen := http.HandlerFunc(func(http.ResponseWriter, *http.Request) { durch = true })
	antwort := httptest.NewRecorder()
	buch.Huelle(innen).ServeHTTP(antwort, bindungsruf(t, "Sets"))

	if durch {
		t.Error("der gestörte Ruf ging trotzdem zur Bindung durch")
	}

	var gelesen struct {
		Fehler *struct {
			Code       string `json:"code"`
			Schluessel string `json:"schluessel"`
		} `json:"fehler"`
	}
	if fehler := json.Unmarshal(antwort.Body.Bytes(), &gelesen); fehler != nil || gelesen.Fehler == nil {
		t.Fatalf("keine fachliche Fehlerantwort: %q (%v)", antwort.Body.String(), fehler)
	}

	if gelesen.Fehler.Code != "unbekannt" || gelesen.Fehler.Schluessel != "unbekannt" {
		t.Errorf("Fehler %+v", gelesen.Fehler)
	}

	if eintrag := buch.Abzug(false)["Sets"]; eintrag.Ausgaenge["gestoert"] != 1 {
		t.Errorf("das Rufbuch führt den Ruf als %+v, erwartet war gestört", eintrag.Ausgaenge)
	}
}

func TestEineStoerungTrifftNurIhreBindung(t *testing.T) {
	t.Parallel()

	buch := neuesRufbuch()
	buch.Stoeren([]Stoerung{{Bindung: "Sets", Fehler: true}})

	antwort := httptest.NewRecorder()
	buch.Huelle(antwortend(t, http.StatusOK, `{"decks":[]}`)).ServeHTTP(antwort, bindungsruf(t, "Decks"))

	if antwort.Body.String() != `{"decks":[]}` {
		t.Errorf("eine andere Bindung wurde mitgestört: %q", antwort.Body.String())
	}
}

func TestOhneStoerungenGehtAllesDurch(t *testing.T) {
	t.Parallel()

	buch := neuesRufbuch()
	buch.Stoeren([]Stoerung{{Bindung: "Sets", Fehler: true}})
	buch.Stoeren(nil)

	antwort := httptest.NewRecorder()
	buch.Huelle(antwortend(t, http.StatusOK, `{"sets":[]}`)).ServeHTTP(antwort, bindungsruf(t, "Sets"))

	if antwort.Body.String() != `{"sets":[]}` {
		t.Errorf("nach dem Aufheben blieb die Störung: %q", antwort.Body.String())
	}
}

// **Ein abgebrochener Ruf wartet die Verzögerung nicht ab** — sonst hielte
// eine überholte Vorschau die nächste auf.
func TestEinAbgebrochenerRufWartetNichtAufDieVerzoegerung(t *testing.T) {
	t.Parallel()

	buch := neuesRufbuch()
	buch.Stoeren([]Stoerung{{Bindung: "Sets", VerzoegerungMs: 5000}})

	anfrage := bindungsruf(t, "Sets")
	ctx, abbrechen := context.WithCancel(anfrage.Context())
	abbrechen()

	beginn := time.Now()
	buch.Huelle(antwortend(t, http.StatusOK, `{}`)).ServeHTTP(httptest.NewRecorder(), anfrage.WithContext(ctx))

	if dauer := time.Since(beginn); dauer > time.Second {
		t.Errorf("der abgebrochene Ruf wartete %v", dauer)
	}
}

// **Eine Panik in der Gestalt, die Wails schickt** — F03 (#2342). So prüft
// die Hüllenprobe, was das Fenster aus einer Bindungspanik macht, ohne einen
// echten Fehler in eine Bindung zu legen: 422, `kind: RuntimeError` und die
// Meldung mit Paketpfad (`transport_http.go`, `httpError`; `bindings.go`,
// `Call`, v3.0.0-beta.27).
func TestEineGestoerteBindungAntwortetWieEinePanik(t *testing.T) {
	t.Parallel()

	buch := neuesRufbuch()
	buch.Stoeren([]Stoerung{{Bindung: "Sets", Panik: true}})

	antwort := httptest.NewRecorder()
	buch.Huelle(antwortend(t, http.StatusOK, `{"sets":[]}`)).ServeHTTP(antwort, bindungsruf(t, "Sets"))

	var gelesen struct {
		Meldung string `json:"message"`
		Art     string `json:"kind"`
	}
	if fehler := json.Unmarshal(antwort.Body.Bytes(), &gelesen); fehler != nil {
		t.Fatalf("keine Panikantwort: %q (%v)", antwort.Body.String(), fehler)
	}

	if antwort.Code != http.StatusUnprocessableEntity || gelesen.Art != "RuntimeError" ||
		gelesen.Meldung != "github.com/DavidTHeider/MagicSetEditor4Mac/internal/app.Bindungen.Sets: panic: Probepanik" {
		t.Errorf("Status %d, Antwort %+v", antwort.Code, gelesen)
	}
}
