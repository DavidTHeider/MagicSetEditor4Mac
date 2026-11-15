package anmeldung_test

// **Der Tausch von Code gegen Identität** — Werkstück 4 von Phase 4.12
// (#1351).
//
// # Gegen einen selbst geschriebenen Anbieter (§16)
//
// **Eingecheckt und selbst geschrieben**, nicht gegen den echten Dienst: ein
// Lauf gegen GitHub bräuchte eine registrierte Anwendung, ein Geheimnis und
// einen Menschen, der im Browser klickt. Was hier geprüft wird, ist der
// **Weg** — dass PKCE mitgeht, dass das Geheimnis dort steht, wo es hingehört,
// und dass aus der Antwort eine Identität wird.

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/anmeldung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Der OAuth-Weg holt die Identität aus einem zweiten Ruf.**
func TestDerOAuthWegHoltDieIdentitaetAusEinemZweitenRuf(t *testing.T) {
	t.Parallel()

	var gesehen url.Values

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, anfrage *http.Request) {
			if anfrage.URL.Path == "/token" {
				if err := anfrage.ParseForm(); err != nil {
					t.Errorf("das Formular lesen: %v", err)
				}

				gesehen = anfrage.PostForm

				antworte(t, schreiben, `{"access_token":"abc","token_type":"bearer"}`)

				return
			}

			// **Und der zweite Ruf trägt das Zugangswort.**
			if anfrage.Header.Get("Authorization") != "Bearer abc" {
				t.Errorf("der zweite Ruf trug %q", anfrage.Header.Get("Authorization"))
			}

			antworte(t, schreiben, `{"id":4711,"login":"wer","name":"Wer Auch Immer"}`)
		}))
	t.Cleanup(gegenstelle.Close)

	anbieter := anmeldung.GitHub("kennung", "geheim")
	anbieter.Tauschadresse = gegenstelle.URL + "/token"
	anbieter.Nutzeradresse = gegenstelle.URL + "/user"

	lauf := mussBeginnen(t, anbieter)

	kennung, fehler := lauf.Einloesen(t.Context(), anmeldung.RuferUeber(gegenstelle.Client()), "der-code")
	if fehler != nil {
		t.Fatalf("einlösen: %v", fehler)
	}

	if kennung.Subjekt != "4711" || kennung.Anbieter != "github" {
		t.Errorf("Identität %+v", kennung)
	}

	if kennung.Anzeigename != "Wer Auch Immer" {
		t.Errorf("Anzeigename %q", kennung.Anzeigename)
	}

	// **Und der Tausch trug, was er tragen muss.**
	for feld, erwartet := range map[string]string{
		"grant_type":    "authorization_code",
		"code":          "der-code",
		"client_id":     "kennung",
		"client_secret": "geheim",
	} {
		if gesehen.Get(feld) != erwartet {
			t.Errorf("%s ist %q, erwartet war %q", feld, gesehen.Get(feld), erwartet)
		}
	}

	if gesehen.Get("code_verifier") == "" {
		t.Error("der Prüfer fehlt — PKCE ginge ins Leere")
	}
}

// **Der OIDC-Weg nimmt sie aus dem `id_token`.**
//
// # Und prüft die Signatur **nicht**
//
// **Das ist erlaubt und es ist eine Entscheidung.** OIDC Core 3.1.3.7: „If the
// ID Token is received via direct communication between the Client and the
// Token Endpoint (which it is in this flow), the TLS server validation MAY be
// used to validate the issuer in place of checking the token signature."
//
// **Was das kostet**: die Identität hängt an der TLS-Kette des
// Tauschendpunkts. Wer ein gültiges Zertifikat für `oauth2.googleapis.com`
// vorweisen kann, kann sie fälschen — er könnte dann allerdings ebenso den
// Tausch selbst beantworten. Der Preis ist derselbe wie für den ganzen Fluss.
func TestDerOIDCWegNimmtDieIdentitaetAusDemIDToken(t *testing.T) {
	t.Parallel()

	var jeton string

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			antworte(t, schreiben,
				`{"access_token":"abc","id_token":"`+jeton+`"}`)
		}))
	t.Cleanup(gegenstelle.Close)

	anbieter := anmeldung.Google("kennung", "")
	anbieter.Tauschadresse = gegenstelle.URL + "/token"

	lauf := mussBeginnen(t, anbieter)

	// **Aussteller und Zielgruppe gehören dazu** — seit dem Nachlauf zum
	// Zweitblick über Phase 4.12 prüft `ausOIDC` beide (§3.1.3.7). Ein Token
	// ohne sie ist keines, das ein Anbieter je ausstellte.
	jeton = jetonMit(t, map[string]any{
		"sub": "1234567890", "name": "Wer Auch Immer",
		"iss": anmeldung.AusstellerGoogle, "aud": "kennung",
		"exp": time.Now().Add(time.Hour).Unix(), "nonce": nonceAus(t, lauf),
	})

	kennung, fehler := lauf.Einloesen(t.Context(), anmeldung.RuferUeber(gegenstelle.Client()), "der-code")
	if fehler != nil {
		t.Fatalf("einlösen: %v", fehler)
	}

	if kennung.Subjekt != "1234567890" || kennung.Anbieter != "google" {
		t.Errorf("Identität %+v", kennung)
	}
}

// **Ein `nonce`, der nicht stimmt, wird abgewiesen.**
//
// OIDC Core 3.1.3.7: „If present in the ID Token, Clients MUST verify that the
// nonce Claim Value is equal to the value of the nonce parameter sent in the
// Authentication Request." **Ohne diese Prüfung ließe sich ein Token
// wiedereinspielen**, das für eine andere Anmeldung ausgestellt wurde.
func TestEinFalscherNonceWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			antworte(t, schreiben, `{"id_token":"`+
				jetonMit(t, map[string]any{"sub": "1", "nonce": "ein anderer"})+`"}`)
		}))
	t.Cleanup(gegenstelle.Close)

	anbieter := anmeldung.Google("kennung", "")
	anbieter.Tauschadresse = gegenstelle.URL + "/token"

	lauf := mussBeginnen(t, anbieter)

	if _, fehler := lauf.Einloesen(t.Context(), anmeldung.RuferUeber(gegenstelle.Client()),
		"der-code"); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("Fehler %v, erwartet war ErrValidation", fehler)
	}
}

// jetonMit baut ein `id_token` ohne Signatur.
//
// **Ohne Signatur, und das ist kein Mangel des Tests**: geprüft wird sie auch
// im Betrieb nicht — siehe [TestDerOIDCWegNimmtDieIdentitaetAusDemIDToken].
// Ein signiertes Token hier wäre eine Vorrichtung, die mehr kann als der
// Betrieb, und das ist die andere Richtung derselben Falle (§3).
func jetonMit(t *testing.T, ansprueche map[string]any) string {
	t.Helper()

	roh, err := json.Marshal(ansprueche)
	if err != nil {
		t.Fatalf("die Ansprüche schreiben: %v", err)
	}

	kopf := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none"}`))

	return kopf + "." + base64.RawURLEncoding.EncodeToString(roh) + ".x"
}

// nonceAus liest den `nonce` aus der Anmeldeadresse eines Laufs.
func nonceAus(t *testing.T, lauf *anmeldung.Lauf) string {
	t.Helper()

	adresse, err := url.Parse(lauf.Adresse)
	if err != nil {
		t.Fatalf("die Adresse zerlegen: %v", err)
	}

	gesehen := adresse.Query().Get("nonce")
	if gesehen == "" {
		t.Fatal("die Adresse trägt keinen nonce")
	}

	return gesehen
}

// antworte schreibt einen Rumpf und meldet, wenn es nicht geht.
//
// **Dieselbe Form wie in `internal/mtgapi`, und aus demselben Grund**: ein
// Testserver, der still nicht antwortet, macht den Lauf grün und die Zusage
// leer.
func antworte(t *testing.T, schreiben http.ResponseWriter, rumpf string) {
	t.Helper()

	schreiben.Header().Set("Content-Type", "application/json")

	if _, fehler := schreiben.Write([]byte(rumpf)); fehler != nil {
		t.Errorf("antworten: %v", fehler)
	}
}
