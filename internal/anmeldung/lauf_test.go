package anmeldung_test

// **Der Anmeldefluss** — Werkstück 4 von Phase 4.12
// (#1351).

import (
	"errors"
	"net/url"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/anmeldung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Adresse trägt `state`, PKCE und die Rückadresse.**
//
// # Was hier geprüft wird
//
// **Jeder Wert, den der Anbieter verlangt** — und für die vier ist das
// nachgelesen (Vermerk 0066): `code_challenge_method=S256` bei allen, weil
// GitHub nichts anderes nimmt; `state` bei allen, weil X es zur Pflicht macht.
func TestDieAdresseTraegtStateUndPKCE(t *testing.T) {
	t.Parallel()

	lauf, fehler := anmeldung.Beginne(anmeldung.GitHub("kennung", "geheim"),
		"http://127.0.0.1:7788/rueckruf")
	if fehler != nil {
		t.Fatalf("beginnen: %v", fehler)
	}

	adresse, zerlegefehler := url.Parse(lauf.Adresse)
	if zerlegefehler != nil {
		t.Fatalf("die Adresse zerlegen: %v", zerlegefehler)
	}

	werte := adresse.Query()

	gefordert := map[string]string{
		"client_id":             "kennung",
		"response_type":         "code",
		"redirect_uri":          "http://127.0.0.1:7788/rueckruf",
		"code_challenge_method": "S256",
	}

	for feld, erwartet := range gefordert {
		if gesehen := werte.Get(feld); gesehen != erwartet {
			t.Errorf("%s ist %q, erwartet war %q", feld, gesehen, erwartet)
		}
	}

	for _, feld := range []string{"state", "code_challenge"} {
		if werte.Get(feld) == "" {
			t.Errorf("%s fehlt", feld)
		}
	}
}

// **Zwei Läufe teilen nichts.**
//
// **Weder `state` noch den Prüfer.** Wären sie gleich, wäre `state` keine
// Sperre gegen einen untergeschobenen Rückruf und PKCE keine gegen einen
// abgefangenen Code.
func TestZweiLaeufeTeilenNichts(t *testing.T) {
	t.Parallel()

	erster := mussBeginnen(t, anmeldung.GitHub("k", "g"))
	zweiter := mussBeginnen(t, anmeldung.GitHub("k", "g"))

	if erster.Adresse == zweiter.Adresse {
		t.Error("zwei Läufe haben dieselbe Adresse")
	}
}

// **Ein OIDC-Anbieter bekommt `nonce` und `openid`.**
//
// Bei Microsoft ist `nonce` im hybrid flow **Pflicht**; hier läuft der reine
// Code-Fluss, und OIDC Core 3.1.3.7 verlangt die Prüfung nur, *wenn* der
// `nonce` im Token steht. Gesendet wird er trotzdem — sonst gibt es nichts zu
// prüfen.
func TestEinOIDCAnbieterBekommtNonceUndOpenid(t *testing.T) {
	t.Parallel()

	lauf, fehler := anmeldung.Beginne(anmeldung.Google("kennung", ""),
		"http://127.0.0.1:7788/rueckruf")
	if fehler != nil {
		t.Fatalf("beginnen: %v", fehler)
	}

	werte := werteAus(t, lauf)

	if werte.Get("nonce") == "" {
		t.Error("nonce fehlt")
	}

	if !strings.Contains(werte.Get("scope"), "openid") {
		t.Errorf("scope ist %q und nennt openid nicht", werte.Get("scope"))
	}
}

// **Ein Rückruf mit falschem `state` wird abgewiesen.**
//
// **Das ist der Kern der Sperre**: ohne diese Prüfung nähme die Anwendung
// einen Code an, den jemand anderes untergeschoben hat — und meldete den
// Angreifer als den Nutzer an.
func TestEinRueckrufMitFalschemStateWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	lauf := mussBeginnen(t, anmeldung.GitHub("k", "g"))

	fehler := lauf.PruefeRueckruf(url.Values{
		"code": {"c"}, "state": {"etwas anderes"},
	})

	if !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("Fehler %v, erwartet war ErrValidation", fehler)
	}
}

// **Und einer ohne Code ebenso.**
func TestEinRueckrufOhneCodeWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	lauf := mussBeginnen(t, anmeldung.GitHub("k", "g"))

	if fehler := lauf.PruefeRueckruf(url.Values{
		"state": {lauf.State},
	}); fehler == nil {
		t.Error("der Rückruf ohne Code kam durch")
	}
}

// **Eine Absage des Anbieters wird als solche gemeldet.**
//
// **Nicht als „etwas ging schief".** Wer im Browser auf „Abbrechen" drückt,
// hat eine Entscheidung getroffen; eine Fehlermeldung darüber wäre eine
// Falschauskunft (§8).
func TestEineAbsageDesAnbietersWirdAlsSolcheGemeldet(t *testing.T) {
	t.Parallel()

	lauf := mussBeginnen(t, anmeldung.GitHub("k", "g"))

	fehler := lauf.PruefeRueckruf(url.Values{
		"state":             {lauf.State},
		"error":             {"access_denied"},
		"error_description": {"the user canceled the authentication"},
	})

	if !errors.Is(fehler, anmeldung.ErrAbgesagt) {
		t.Errorf("Fehler %v, erwartet war ErrAbgesagt", fehler)
	}
}

// **Ein richtiger Rückruf geht durch.**
func TestEinRichtigerRueckrufGehtDurch(t *testing.T) {
	t.Parallel()

	lauf := mussBeginnen(t, anmeldung.GitHub("k", "g"))

	if fehler := lauf.PruefeRueckruf(url.Values{
		"code": {"c"}, "state": {lauf.State},
	}); fehler != nil {
		t.Errorf("abgewiesen: %v", fehler)
	}
}

// **Ein Anbieter ohne Kennung wird abgewiesen.**
//
// **Und das ist der Normalfall**, nicht der Sonderfall: dieses Projekt liefert
// keine Kennungen mit — wer sich anmelden will, trägt seine eigenen ein
// (Vermerk 0066). Ohne sie ist der Lauf gar nicht zu beginnen.
func TestEinAnbieterOhneKennungWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	if _, fehler := anmeldung.Beginne(anmeldung.GitHub("", ""),
		"http://127.0.0.1:1/r"); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("Fehler %v, erwartet war ErrValidation", fehler)
	}
}

// **Und GitHub ohne Geheimnis ebenso** — belegt am 2026-09-05.
//
// GitHubs Tauschendpunkt führt `client_secret` als **Required**, auch mit
// PKCE. Google nennt es für Desktop-Klienten *optional*, Microsoft verbietet
// es für öffentliche Klienten ausdrücklich. Wer das nicht je Anbieter führt,
// baut einen Fluss, der bei einem von dreien scheitert — und zwar erst am
// Tauschendpunkt, wo die Meldung keinem etwas sagt.
func TestGitHubOhneGeheimnisWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	if _, fehler := anmeldung.Beginne(anmeldung.GitHub("kennung", ""),
		"http://127.0.0.1:1/r"); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("Fehler %v, erwartet war ErrValidation", fehler)
	}
}

// **Microsoft mit Geheimnis wird abgewiesen.**
//
// Die andere Richtung, und sie ist ebenso belegt: „Public clients … must not
// use secrets or certificates when redeeming an authorization code."
func TestMicrosoftMitGeheimnisWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	if _, fehler := anmeldung.Beginne(anmeldung.Microsoft("kennung", "geheim"),
		"http://127.0.0.1:1/r"); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("Fehler %v, erwartet war ErrValidation", fehler)
	}
}

// **Ein unbekannter Anbieter wird abgewiesen.**
func TestEinUnbekannterAnbieterWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	fremd := anmeldung.Anbieter{Name: "apple", Kennung: "k"}

	if _, fehler := anmeldung.Beginne(fremd,
		"http://127.0.0.1:1/r"); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("Fehler %v, erwartet war ErrValidation", fehler)
	}
}

// **Ohne Rückadresse gibt es keinen Weg zurück.**
func TestOhneRueckadresseWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	if _, fehler := anmeldung.Beginne(anmeldung.GitHub("k", "g"),
		"  "); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("Fehler %v, erwartet war ErrValidation", fehler)
	}
}

// **Ein Fehler des Anbieters, der keine Absage ist, ist eine Störung.**
//
// **Und die beiden dürfen nicht zusammenfallen**: `access_denied` heißt „ich
// will nicht", `server_error` heißt „bei mir ging etwas schief". Wer beides
// gleich behandelt, sagt dem einen, er habe abgebrochen, und dem anderen, er
// solle es noch einmal versuchen.
func TestEinAnbieterfehlerIstKeineAbsage(t *testing.T) {
	t.Parallel()

	lauf := mussBeginnen(t, anmeldung.GitHub("k", "g"))

	fehler := lauf.PruefeRueckruf(url.Values{
		"state": {lauf.State}, "error": {"server_error"},
	})

	if errors.Is(fehler, anmeldung.ErrAbgesagt) {
		t.Error("eine Störung kam als Absage an")
	}

	if !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("Fehler %v, erwartet war ErrValidation", fehler)
	}
}

// **X gibt seinen Anbieter her** — und `state` ist dort Pflicht.
func TestXGibtSeinenAnbieter(t *testing.T) {
	t.Parallel()

	lauf, fehler := anmeldung.Beginne(anmeldung.X("kennung", ""),
		"http://127.0.0.1:1/r")
	if fehler != nil {
		t.Fatalf("beginnen: %v", fehler)
	}

	if werteAus(t, lauf).Get("state") == "" {
		t.Error("state fehlt — bei X ist er Pflicht")
	}

	if !strings.Contains(werteAus(t, lauf).Get("scope"), "users.read") {
		t.Errorf("scope ist %q", werteAus(t, lauf).Get("scope"))
	}
}

// mussBeginnen beginnt einen Lauf und bricht ab, wenn es nicht geht.
//
// **Und nicht `_`**: ein verworfener Fehler hier hieße, dass der Test mit
// einem `nil`-Lauf weiterläuft und an einer ganz anderen Stelle abstürzt —
// die Meldung stünde dann über der falschen Zeile.
func mussBeginnen(t *testing.T, anbieter anmeldung.Anbieter) *anmeldung.Lauf {
	t.Helper()

	lauf, fehler := anmeldung.Beginne(anbieter, "http://127.0.0.1:1/r")
	if fehler != nil {
		t.Fatalf("beginnen: %v", fehler)
	}

	return lauf
}

// werteAus zerlegt die Anmeldeadresse eines Laufs.
func werteAus(t *testing.T, lauf *anmeldung.Lauf) url.Values {
	t.Helper()

	adresse, fehler := url.Parse(lauf.Adresse)
	if fehler != nil {
		t.Fatalf("die Adresse zerlegen: %v", fehler)
	}

	return adresse.Query()
}
