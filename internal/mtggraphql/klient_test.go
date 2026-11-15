package mtggraphql_test

// **Der Klient für die MTGJSON-GraphQL-Schnittstelle** — Werkstück 4 von
// Phase 4.8 (#1321).
//
// # Was hier geprüft wird und was nicht
//
// **Geprüft**: die Gestalt der Anfrage, die Kopfzeile, die Deutung der
// Antwort — gegen einen eigenen Testserver und gegen **eingecheckte
// Beispiele, die vom echten Dienst stammen** (`testdaten/`; abgerufen am
// 2026-09-02, ohne Zugangsmarke, deshalb nur die Fehlerfälle echt).
//
// **Nicht geprüft**: dass der echte Dienst Kartendaten so schickt. Es gibt
// hier keine Patreon-Marke — Vermerk 0056, Entscheidung 4. Was daraus folgt,
// steht in der Oberfläche: „nie gegen den echten Dienst gelaufen".

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mtggraphql"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// beispiel liest eine Datei aus `testdaten/`.
func beispiel(t *testing.T, name string) []byte {
	t.Helper()

	//nolint:gosec // der Name kommt aus dem Lauf, nicht von außen
	roh, err := os.ReadFile(filepath.Join("testdaten", name))
	if err != nil {
		t.Fatalf("%s lesen: %v", name, err)
	}

	return roh
}

// schreibe antwortet und meldet einen Schreibfehler.
func schreibe(t *testing.T, schreiber http.ResponseWriter, rumpf []byte) {
	t.Helper()

	if _, err := schreiber.Write(rumpf); err != nil {
		t.Errorf("antworten: %v", err)
	}
}

// klientGegen baut einen Klienten gegen einen Testserver.
func klientGegen(t *testing.T, marke string, bediener http.HandlerFunc) *mtggraphql.Klient {
	t.Helper()

	server := httptest.NewServer(bediener)
	t.Cleanup(server.Close)

	return mtggraphql.NeuerKlient(server.URL, netz.KlientUeber(server.Client()), marke)
}

// **Die Anfrage ist ein POST mit Abfrage und Variablen.**
//
// **Nachgemessen am 2026-09-02**: der Dienst nimmt `{"query": …}` als JSON
// im Rumpf. Wer die Abfrage in die Adresse schriebe, bekäme einen 400 mit
// „potential Cross-Site Request Forgery".
func TestDieAnfrageIstEinPostMitAbfrage(t *testing.T) {
	t.Parallel()

	var (
		methode string
		typ     string
		rumpf   []byte
	)

	klient := klientGegen(t, "geheim", func(schreiber http.ResponseWriter, gestellt *http.Request) {
		methode = gestellt.Method
		typ = gestellt.Header.Get("Content-Type")
		gelesen, lesefehler := io.ReadAll(gestellt.Body)
		if lesefehler != nil {
			t.Errorf("den Rumpf lesen: %v", lesefehler)
		}

		rumpf = gelesen

		schreibe(t, schreiber, beispiel(t, "karten.json"))
	})

	if _, err := klient.Karten(t.Context(), "Blitzschlag"); err != nil {
		t.Fatalf("fragen: %v", err)
	}

	if methode != http.MethodPost {
		t.Errorf("die Methode ist %q, erwartet POST", methode)
	}

	if !strings.Contains(typ, "application/json") {
		t.Errorf("der Inhaltstyp ist %q", typ)
	}

	for _, teil := range []string{`"query"`, "cards", "name_eq"} {
		if !strings.Contains(string(rumpf), teil) {
			t.Errorf("der Rumpf %q enthält %q nicht", rumpf, teil)
		}
	}
}

// **Die Marke reist als `authorization: Bearer …`.**
//
// **Nachgelesen am 2026-09-02** unter `https://mtgjson.com/mtggraphql/`:
// „authorization" mit „Bearer <Access Token>".
func TestDieMarkeReistAlsBearer(t *testing.T) {
	t.Parallel()

	var gesehen string

	klient := klientGegen(t, "geheim", func(schreiber http.ResponseWriter, gestellt *http.Request) {
		gesehen = gestellt.Header.Get("Authorization")

		schreibe(t, schreiber, beispiel(t, "karten.json"))
	})

	if _, err := klient.Karten(t.Context(), "Blitzschlag"); err != nil {
		t.Fatalf("fragen: %v", err)
	}

	if gesehen != "Bearer geheim" {
		t.Errorf("die Kopfzeile ist %q, erwartet »Bearer geheim«", gesehen)
	}
}

// **Ohne Marke wird gar nicht erst gefragt.**
//
// Der Dienst antwortet ohne sie mit „Access denied" — eine Anfrage, deren
// Antwort feststeht, ist eine Anfrage zu viel (§16, Regel 4). Und die
// Meldung sagt, was fehlt, statt einen Serverfehler zu behaupten.
func TestOhneMarkeWirdNichtGefragt(t *testing.T) {
	t.Parallel()

	var anfragen int

	klient := klientGegen(t, "", func(schreiber http.ResponseWriter, _ *http.Request) {
		anfragen++

		schreibe(t, schreiber, beispiel(t, "karten.json"))
	})

	_, err := klient.Karten(t.Context(), "Blitzschlag")

	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("Fehler ist %v, erwartet eine Validierung", err)
	}

	if !strings.Contains(err.Error(), "Zugangsmarke") {
		t.Errorf("die Meldung %q sagt nicht, was fehlt", err)
	}

	if anfragen != 0 {
		t.Errorf("%d Anfragen ohne Marke, erwartet keine", anfragen)
	}
}

// **Ein Fehler im Rumpf ist ein Fehler — auch bei HTTP 200.**
//
// **Nachgemessen am 2026-09-02**: der Dienst antwortet auf eine Abfrage ohne
// Marke mit **Status 200**, `data: null` und einem `errors`-Feld. Wer nur den
// Statuscode liest, sieht einen Erfolg (§16: „`errors` im Antwortkörper als
// Befund und nicht als Erfolg").
func TestEinFehlerImRumpfIstEinFehler(t *testing.T) {
	t.Parallel()

	klient := klientGegen(t, "geheim", func(schreiber http.ResponseWriter, _ *http.Request) {
		schreiber.WriteHeader(http.StatusOK)

		schreibe(t, schreiber, beispiel(t, "zugriffverweigert.json"))
	})

	_, err := klient.Karten(t.Context(), "Blitzschlag")

	if err == nil {
		t.Fatal("HTTP 200 mit errors kam als Erfolg an")
	}

	if !strings.Contains(err.Error(), "Access denied") {
		t.Errorf("die Meldung %q gibt den Fehler des Dienstes nicht wieder", err)
	}
}

// **»Zugriff verweigert« ist kein Serverfehler.**
//
// **Nachgemessen am 2026-09-02**: der Dienst nennt dafür den Code
// `INTERNAL_SERVER_ERROR` — das ist irreführend, und ein Klient, der ihn
// weiterreichte, meldete „der Dienst ist kaputt", wo „deine Marke gilt
// nicht" gemeint ist (§8).
func TestZugriffVerweigertIstEinZugangsfehler(t *testing.T) {
	t.Parallel()

	klient := klientGegen(t, "falsch", func(schreiber http.ResponseWriter, _ *http.Request) {
		schreibe(t, schreiber, beispiel(t, "zugriffverweigert.json"))
	})

	_, err := klient.Karten(t.Context(), "Blitzschlag")

	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("Fehler ist %v, erwartet eine Validierung", err)
	}

	if strings.Contains(err.Error(), "INTERNAL_SERVER_ERROR") {
		t.Errorf("die Meldung reicht den irreführenden Code durch: %v", err)
	}
}

// **Eine Karte kommt durch.**
func TestEineKarteKommtDurch(t *testing.T) {
	t.Parallel()

	klient := klientGegen(t, "geheim", func(schreiber http.ResponseWriter, _ *http.Request) {
		schreibe(t, schreiber, beispiel(t, "karten.json"))
	})

	karten, err := klient.Karten(t.Context(), "Blitzschlag")
	if err != nil {
		t.Fatalf("fragen: %v", err)
	}

	if len(karten) != 1 {
		t.Fatalf("%d Karten, erwartet 1", len(karten))
	}

	eine := karten[0]

	for _, fall := range []struct{ feld, gelesen, erwartet string }{
		{"Kennung", eine.Kennung, "aaaa-bbbb-cccc-dddd"},
		{"Name", eine.Name, "Blitzschlag"},
		{"Manakosten", eine.Manakosten, "{R}"},
		{"Typzeile", eine.Typzeile, "Instant"},
		{"Seltenheit", eine.Seltenheit, "common"},
		{"Setcode", eine.Setcode, "PRB"},
		{"Nummer", eine.Nummer, "141"},
		{"Kuenstler", eine.Kuenstler, "Selbst gezeichnet"},
	} {
		if fall.gelesen != fall.erwartet {
			t.Errorf("%s ist %q, erwartet %q", fall.feld, fall.gelesen, fall.erwartet)
		}
	}
}

// **Ein Fehler in der Abfrage wird genannt.**
func TestEinAbfragefehlerWirdGenannt(t *testing.T) {
	t.Parallel()

	klient := klientGegen(t, "geheim", func(schreiber http.ResponseWriter, _ *http.Request) {
		schreiber.WriteHeader(http.StatusBadRequest)

		schreibe(t, schreiber, beispiel(t, "feldfehler.json"))
	})

	_, err := klient.Karten(t.Context(), "Blitzschlag")

	if err == nil {
		t.Fatal("angenommen, erwartet ein Fehler")
	}

	if !strings.Contains(err.Error(), "Unknown argument") {
		t.Errorf("die Meldung %q gibt den Fehler des Dienstes nicht wieder", err)
	}
}
