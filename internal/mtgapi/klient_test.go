package mtgapi_test

// **Der Klient für magicthegathering.io** — Werkstück 3 von Phase 4.8.
//
// **Gegen einen eigenen Testserver**, nicht gegen den echten Dienst: hier
// wird die **Gestalt der Anfrage** geprüft — Kopfzeilen, Blätterung, der
// Umgang mit der Grenze. Was der echte Dienst antwortet, prüft
// `netzlauf_test.go` hinter seinem Build-Tag (§16, §3).

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mtgapi"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// schreibe antwortet und meldet einen Schreibfehler.
//
// **Ein `_, _ =` genügt hier nicht**: `errcheck` läuft mit `check-blank`, und
// das ist die richtige Einstellung — ein Testserver, der still nicht
// antwortet, macht den Lauf grün und die Zusage leer.
func schreibe(t *testing.T, schreiber http.ResponseWriter, rumpf []byte) {
	t.Helper()

	if _, err := schreiber.Write(rumpf); err != nil {
		t.Errorf("antworten: %v", err)
	}
}

// klientGegen baut einen Klienten gegen einen Testserver.
func klientGegen(t *testing.T, bediener http.HandlerFunc) *mtgapi.Klient {
	t.Helper()

	server := httptest.NewServer(bediener)
	t.Cleanup(server.Close)

	return mtgapi.NeuerKlient(server.URL, netz.KlientUeber(server.Client()))
}

// **Die Anfrage nennt das Programm** (§16, Regel 2).
func TestDieAnfrageNenntDasProgramm(t *testing.T) {
	t.Parallel()

	var gesehen http.Header

	klient := klientGegen(t, func(schreiber http.ResponseWriter, gestellt *http.Request) {
		gesehen = gestellt.Header.Clone()
		schreibe(t, schreiber, []byte(`{"cards":[]}`))
	})

	if _, _, err := klient.Karten(t.Context(), "name=Blitz"); err != nil {
		t.Fatalf("fragen: %v", err)
	}

	if kenner := gesehen.Get("User-Agent"); !strings.Contains(kenner, "MagicSetEditor4Mac") {
		t.Errorf("der User-Agent ist %q", kenner)
	}
}

// **Die Blätterung wird gesetzt, und die Seitengröße ist die erlaubte.**
//
// **Nachgelesen am 2026-09-02**: `pageSize` hat Vorgabe **und** Höchstwert
// 100. Wer mehr verlangte, bekäme nicht mehr — und behauptete es in seinem
// eigenen Code.
func TestDieBlaetterungWirdGesetzt(t *testing.T) {
	t.Parallel()

	var gesehen string

	klient := klientGegen(t, func(schreiber http.ResponseWriter, gestellt *http.Request) {
		gesehen = gestellt.URL.RawQuery
		schreibe(t, schreiber, []byte(`{"cards":[]}`))
	})

	if _, _, err := klient.Karten(t.Context(), "name=Blitz"); err != nil {
		t.Fatalf("fragen: %v", err)
	}

	for _, teil := range []string{"name=Blitz", "pageSize=100"} {
		if !strings.Contains(gesehen, teil) {
			t.Errorf("die Abfrage %q enthält %q nicht", gesehen, teil)
		}
	}
}

// **Ein 403 ist die Grenze und wird nicht wiederholt** (§16, Regel 4).
//
// **Nachgelesen am 2026-09-02**: magicthegathering.io antwortet über der
// Grenze mit **403 Forbidden** und „Rate Limit Exceeded" — nicht mit 429.
// Ein Klient, der das als „verboten" deutete, sagte etwas Falsches; einer,
// der es wiederholte, verstieße gegen die Bedingungen.
func TestEinDreiundvierzigIstDieGrenze(t *testing.T) {
	t.Parallel()

	var anfragen int

	klient := klientGegen(t, func(schreiber http.ResponseWriter, _ *http.Request) {
		anfragen++

		schreiber.Header().Set("Ratelimit-Limit", "5000")
		schreiber.Header().Set("Ratelimit-Remaining", "0")
		schreiber.WriteHeader(http.StatusForbidden)
		schreibe(t, schreiber, []byte("Rate Limit Exceeded"))
	})

	_, _, err := klient.Karten(t.Context(), "name=Blitz")

	if !errors.Is(err, domain.ErrGedrosselt) {
		t.Fatalf("Fehler ist %v, erwartet ErrGedrosselt", err)
	}

	if anfragen != 1 {
		t.Errorf("%d Anfragen, erwartet genau eine — ein 429 wird nicht wiederholt", anfragen)
	}
}

// **Ein 403 ohne diese Meldung ist etwas anderes.**
//
// Derselbe Statuscode kann „du darfst hier nicht" heißen. Wer beides
// zusammenwürfe, riete dem Nutzenden zu warten, wo Warten nichts hilft (§8).
func TestEinDreiundvierzigOhneGrenzeIstKeineDrossel(t *testing.T) {
	t.Parallel()

	klient := klientGegen(t, func(schreiber http.ResponseWriter, _ *http.Request) {
		schreiber.WriteHeader(http.StatusForbidden)
		schreibe(t, schreiber, []byte("Nope"))
	})

	_, _, err := klient.Karten(t.Context(), "name=Blitz")

	if err == nil {
		t.Fatal("angenommen, erwartet ein Fehler")
	}

	if errors.Is(err, domain.ErrGedrosselt) {
		t.Errorf("ein fremdes 403 kam als Drosselung an: %v", err)
	}
}

// **Was übrig ist, wird gelesen und weitergereicht.**
//
// `Ratelimit-Remaining` sagt, wie viele Anfragen noch gehen. Ein Programm,
// das sie liest und wegwirft, hat sie nicht gelesen.
func TestDerRestwertKommtAn(t *testing.T) {
	t.Parallel()

	klient := klientGegen(t, func(schreiber http.ResponseWriter, _ *http.Request) {
		schreiber.Header().Set("Ratelimit-Limit", "5000")
		schreiber.Header().Set("Ratelimit-Remaining", "4711")
		schreibe(t, schreiber, []byte(`{"cards":[]}`))
	})

	_, stand, err := klient.Karten(t.Context(), "name=Blitz")
	if err != nil {
		t.Fatalf("fragen: %v", err)
	}

	if stand.Grenze != 5000 || stand.Uebrig != 4711 {
		t.Errorf("Stand ist %+v, erwartet 5000/4711", stand)
	}
}

// **Fehlt der Kopf, ist der Stand unbekannt — nicht null.**
//
// „Null übrig" hieße „gleich ist Schluss"; „unbekannt" heißt „der Dienst hat
// nichts gesagt". Wer beides zusammenwürfe, drosselte grundlos.
func TestOhneKopfIstDerStandUnbekannt(t *testing.T) {
	t.Parallel()

	klient := klientGegen(t, func(schreiber http.ResponseWriter, _ *http.Request) {
		schreibe(t, schreiber, []byte(`{"cards":[]}`))
	})

	_, stand, err := klient.Karten(t.Context(), "name=Blitz")
	if err != nil {
		t.Fatalf("fragen: %v", err)
	}

	if stand.Bekannt {
		t.Errorf("ohne Kopf gilt der Stand als bekannt: %+v", stand)
	}
}

// **Eine Karte kommt durch den ganzen Weg.**
func TestEineKarteKommtDurch(t *testing.T) {
	t.Parallel()

	klient := klientGegen(t, func(schreiber http.ResponseWriter, _ *http.Request) {
		schreibe(t, schreiber, beispiel(t, "karte.json"))
	})

	karten, _, err := klient.Karten(t.Context(), "name=Blitzschlag")
	if err != nil {
		t.Fatalf("fragen: %v", err)
	}

	if len(karten) != 1 || karten[0].Name != "Blitzschlag" {
		t.Errorf("gelesen: %+v", karten)
	}
}

// **Ein 404 ist »nichts gefunden« und kein Ausfall.**
func TestEinVierhundertvierIstNichtsGefunden(t *testing.T) {
	t.Parallel()

	klient := klientGegen(t, func(schreiber http.ResponseWriter, _ *http.Request) {
		schreiber.WriteHeader(http.StatusNotFound)
	})

	if _, _, err := klient.Karten(t.Context(), "name=Gibtsnicht"); !errors.Is(
		err, domain.ErrNotFound) {
		t.Errorf("Fehler ist %v, erwartet ErrNotFound", err)
	}
}
