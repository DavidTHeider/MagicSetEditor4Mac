// **Was ein `id_token` beweisen muss** — der Nachlauf zum Zweitblick über
// Phase 4.12 (#1351).
//
// # Warum eine eigene Datei
//
// `tausch_test.go` prüft den **Tausch** — Anfrage, Antwort, Fehlerwege des
// Endpunkts. Hier steht, was das zurückgegebene Token selbst hergeben muss.
// Der Schnitt folgt der Frage (§2).
//
// # Was der Zweitblick angestrichen hat
//
// **Geprüft wurde nur `sub` und ein `nonce`, wenn einer dastand.** OIDC Core
// 1.0 §3.1.3.7 verlangt mehr, und zwar unabhängig von der Signatur:
//
//   - Punkt 2: „The Issuer Identifier … MUST exactly match the value of the
//     iss (issuer) Claim."
//   - Punkt 3: „The Client MUST validate that the aud (audience) Claim
//     contains its client_id."
//   - Punkt 11: „If a nonce value was sent in the Authentication Request, a
//     nonce Claim MUST be present."
//
// **Der dritte Punkt ist der, der stillschweigend fehlte**: das `nonce` wurde
// nur geprüft, **wenn** eines dastand. Ein Token ohne `nonce` kam durch — und
// damit ließe sich eines wiedereinspielen, das für eine andere Anmeldung
// ausgestellt wurde.

package anmeldung_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/anmeldung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// abgewiesen fährt den OIDC-Weg mit einem gebauten Token und erwartet ein Nein.
//
// # Der `nonce` wird eingesetzt, wenn keiner dasteht
//
// **Sonst scheiterte jeder Lauf hier an ihm** — und sagte damit nichts über
// die Prüfung, die er meint (§10: ein Lauf, der scheitert, bevor er die Stelle
// erreicht, die er prüfen will). Genau das ist beim ersten Anlauf passiert:
// vier neue Läufe waren grün, ohne dass es `exp` oder `azp` gab.
//
// **Wer den `nonce` selbst prüfen will, setzt einen** — dann bleibt er stehen.
func abgewiesen(t *testing.T, ansprueche map[string]any) {
	t.Helper()

	var jeton string

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			antworte(t, schreiben, `{"id_token":"`+jeton+`"}`)
		}))
	t.Cleanup(gegenstelle.Close)

	anbieter := anmeldung.Google("kennung", "")
	anbieter.Tauschadresse = gegenstelle.URL + "/token"

	lauf := mussBeginnen(t, anbieter)

	if _, steht := ansprueche["nonce"]; !steht {
		ansprueche["nonce"] = nonceAus(t, lauf)
	}

	jeton = jetonMit(t, ansprueche)

	if _, fehler := lauf.Einloesen(t.Context(),
		anmeldung.RuferUeber(gegenstelle.Client()),
		"der-code"); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("Fehler %v, erwartet war ErrValidation", fehler)
	}
}

// **Ein Token ohne `nonce` wird abgewiesen** — OIDC Core §3.1.3.7, Punkt 11.
func TestEinIDTokenOhneNonceWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	abgewiesen(t, map[string]any{
		"sub": "1", "iss": anmeldung.AusstellerGoogle, "aud": "kennung",
		"exp": time.Now().Add(time.Hour).Unix(), "nonce": "",
	})
}

// **Ein Token für einen anderen Klienten wird abgewiesen** — Punkt 3.
//
// **Ohne diese Prüfung nähme dieses Programm ein Token an, das ein fremder
// Klient beim selben Anbieter für sich hat ausstellen lassen.**
func TestEinIDTokenFuerEinenAnderenKlientenWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	abgewiesen(t, map[string]any{
		"sub": "1", "iss": anmeldung.AusstellerGoogle, "aud": "ein anderer",
		"exp": time.Now().Add(time.Hour).Unix(),
	})
}

// **Ein Token von einem anderen Aussteller wird abgewiesen** — Punkt 2.
func TestEinIDTokenVonEinemAnderenAusstellerWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	abgewiesen(t, map[string]any{
		"sub": "1", "iss": "https://irgendwer.example", "aud": "kennung",
		"exp": time.Now().Add(time.Hour).Unix(),
	})
}

// **Und ein Token ohne `iss` ebenso** — „exactly match" ist mit nichts nicht
// zu erfüllen.
func TestEinIDTokenOhneAusstellerWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	abgewiesen(t, map[string]any{
		"sub": "1", "aud": "kennung",
		"exp": time.Now().Add(time.Hour).Unix(),
	})
}

// **Und `aud` darf eine Liste sein** — die Spezifikation lässt beides zu:
// „The aud value is an array of case-sensitive strings … In the special case
// when the ID Token has one audience, the aud value MAY be a single
// case-sensitive string."
func TestEineListeAlsAudienceGehtDurch(t *testing.T) {
	t.Parallel()

	var jeton string

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			antworte(t, schreiben, `{"id_token":"`+jeton+`"}`)
		}))
	t.Cleanup(gegenstelle.Close)

	anbieter := anmeldung.Google("kennung", "")
	anbieter.Tauschadresse = gegenstelle.URL + "/token"

	lauf := mussBeginnen(t, anbieter)

	jeton = jetonMit(t, map[string]any{
		"sub": "1", "iss": anmeldung.AusstellerGoogle,
		"aud": []string{"ein anderer", "kennung"}, "azp": "kennung",
		"exp": time.Now().Add(time.Hour).Unix(), "nonce": nonceAus(t, lauf),
	})

	if _, fehler := lauf.Einloesen(t.Context(),
		anmeldung.RuferUeber(gegenstelle.Client()), "der-code"); fehler != nil {
		t.Errorf("eine Liste als aud wurde abgewiesen: %v", fehler)
	}
}

// **Ein Token ohne `sub` wird abgewiesen** — ohne Subjekt gibt es keine
// Kennung, und die ist der ganze Zweck des Tausches.
//
// **Der Lauf stand vorher nie an dieser Stelle**: bis zum Nachlauf zum
// Zweitblick über Phase 4.12 kam er nie so weit, weil `iss` und `aud` gar
// nicht geprüft wurden und der erste Fall darüber schon nein sagte.
func TestEinIDTokenOhneSubjektWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	var jeton string

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			antworte(t, schreiben, `{"id_token":"`+jeton+`"}`)
		}))
	t.Cleanup(gegenstelle.Close)

	anbieter := anmeldung.Google("kennung", "")
	anbieter.Tauschadresse = gegenstelle.URL + "/token"

	lauf := mussBeginnen(t, anbieter)

	jeton = jetonMit(t, map[string]any{
		"iss": anmeldung.AusstellerGoogle, "aud": "kennung",
		"exp": time.Now().Add(time.Hour).Unix(), "nonce": nonceAus(t, lauf),
	})

	if _, fehler := lauf.Einloesen(t.Context(),
		anmeldung.RuferUeber(gegenstelle.Client()),
		"der-code"); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("Fehler %v, erwartet war ErrValidation", fehler)
	}
}

// **Und ein `aud`, das weder Zeichenkette noch Liste ist, ebenso.**
//
// **Ein Anbieter schickt so etwas nicht** — aber `aud` kommt von außen, und
// fremde Daten sind der Normalfall (§8). Ohne diesen Zweig verschwände eine
// Zahl still zu „keine Zielgruppe".
func TestEinAudienceAusEtwasAnderemWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	abgewiesen(t, map[string]any{
		"sub": "1", "iss": anmeldung.AusstellerGoogle, "aud": 42,
		"exp": time.Now().Add(time.Hour).Unix(),
	})
}

// **Ein abgelaufenes Token wird abgewiesen** — OIDC Core §3.1.3.7, Punkt 9:
// „The current time MUST be before the time represented by the exp Claim."
//
// **Ein MUSS, und ohne Signaturprüfung nachvollziehbar**: `exp` steht im
// Rumpf, und die Uhr dieses Rechners genügt. Angestrichen vom Nachlauf zum
// Zweitblick über Phase 4.12.
func TestEinAbgelaufenesIDTokenWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	abgewiesen(t, map[string]any{
		"sub": "1", "iss": anmeldung.AusstellerGoogle, "aud": "kennung",
		"exp": time.Now().Add(-time.Minute).Unix(),
	})
}

// **Und ein Token ohne `exp` ebenso** — „MUST be before" ist mit nichts nicht
// zu erfüllen.
func TestEinIDTokenOhneAblaufWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	abgewiesen(t, map[string]any{
		"sub": "1", "iss": anmeldung.AusstellerGoogle, "aud": "kennung",
	})
}

// **Mehrere Zielgruppen ohne `azp` werden abgewiesen** — §3.1.3.7, Punkte 4
// und 5.
//
// **Ein SOLLTE der Spezifikation, und hier ein Muss.** „Mehrere Zielgruppen"
// heißt: dasselbe Token gilt auch woanders. Ohne `azp` sagt nichts, **für
// wen** es ausgestellt wurde — und dieses Programm hat keinen Grund, ein Token
// anzunehmen, das ein anderer angefordert hat.
func TestMehrereZielgruppenOhneAzpWerdenAbgewiesen(t *testing.T) {
	t.Parallel()

	abgewiesen(t, map[string]any{
		"sub": "1", "iss": anmeldung.AusstellerGoogle,
		"aud": []string{"ein anderer", "kennung"},
		"exp": time.Now().Add(time.Hour).Unix(),
	})
}

// **Und ein `azp`, das auf jemand anderen zeigt, ebenso.**
func TestEinFremdesAzpWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	abgewiesen(t, map[string]any{
		"sub": "1", "iss": anmeldung.AusstellerGoogle,
		"aud": []string{"ein anderer", "kennung"}, "azp": "ein anderer",
		"exp": time.Now().Add(time.Hour).Unix(),
	})
}
