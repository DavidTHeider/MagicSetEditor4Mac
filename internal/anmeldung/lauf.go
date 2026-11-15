// **Ein Anmeldelauf** — Werkstück 4 von Phase 4.12
// (#1351).
//
// # Was ein Lauf ist
//
// **Ein einzelner Anmeldeversuch**, samt allem, was zwischen dem Öffnen des
// Browsers und dem Rückruf zusammengehalten werden muss: `state`, der
// PKCE-Prüfer und — bei OIDC — der `nonce`. Wer sie irgendwo anders hielte,
// müsste sie zusammensuchen, und genau das ist die Lücke, gegen die sie da
// sind.

package anmeldung

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Lauf ist ein Anmeldeversuch.
type Lauf struct {
	// Adresse ist, wohin der Browser geht.
	Adresse string

	// State ist der Wert, den der Rückruf tragen muss.
	State string

	anbieter     Anbieter
	rueckadresse string
	pruefer      string
	nonce        string
}

// Beginne baut einen Lauf.
//
// # Warum die Rückadresse hereingereicht wird
//
// **Weil sie den Port kennt, und der steht erst, wenn der Horcher steht.**
// Google verlangt für Desktop-Anwendungen ausdrücklich einen Rückruf auf die
// Schleifenadresse — „http://127.0.0.1:port oder http://[::1]:port. Query your
// platform for the relevant loopback IP address and start an HTTP listener on
// a random available port" — und **eigene Schemata sind dort nicht mehr
// erlaubt**: „Custom URI schemes are no longer supported due to the risk of app
// impersonation" (nachgelesen am 2026-09-05).
func Beginne(anbieter Anbieter, rueckadresse string) (*Lauf, error) {
	if err := anbieter.pruefe(); err != nil {
		return nil, err
	}

	if strings.TrimSpace(rueckadresse) == "" {
		return nil, fmt.Errorf(
			"%w: ohne Rückadresse gibt es keinen Weg zurück", domain.ErrValidation)
	}

	lauf := &Lauf{
		State:        zufall(),
		anbieter:     anbieter,
		rueckadresse: rueckadresse,
		pruefer:      zufall(),
	}

	if anbieter.Gestalt == OIDC {
		lauf.nonce = zufall()
	}

	lauf.Adresse = lauf.adresse()

	return lauf, nil
}

// adresse baut die Anmeldeadresse.
func (l *Lauf) adresse() string {
	werte := url.Values{
		"client_id":     {l.anbieter.Kennung},
		"response_type": {"code"},
		"redirect_uri":  {l.rueckadresse},
		"scope":         {strings.Join(l.anbieter.Bereiche, " ")},
		"state":         {l.State},

		// **`S256` und nicht `plain`.** Google und X nähmen beides, GitHub
		// nimmt **nur** `S256` — und ein Fluss, der je Anbieter ein anderes
		// Verfahren nähme, hätte einen Zweig, den niemand prüft.
		"code_challenge":        {herausforderung(l.pruefer)},
		"code_challenge_method": {"S256"},
	}

	if l.nonce != "" {
		werte.Set("nonce", l.nonce)
	}

	return l.anbieter.Anmeldeadresse + "?" + werte.Encode()
}

// PruefeRueckruf sagt, ob dieser Rückruf zu diesem Lauf gehört.
//
// # Drei Antworten, nicht zwei
//
// **Eine Absage ist keine Störung** ([ErrAbgesagt]): wer im Browser auf
// „Abbrechen" drückt, hat entschieden. **Ein falscher `state` ist ein
// Angriffsversuch** und keine Panne — ohne diese Prüfung nähme die Anwendung
// einen Code an, den jemand anderes untergeschoben hat, und meldete den
// Angreifer als den Nutzer an.
func (l *Lauf) PruefeRueckruf(werte url.Values) error {
	// **Der `state` zuerst.** Eine Fehlermeldung aus einem Rückruf, der nicht
	// zu diesem Lauf gehört, ist die Meldung eines Fremden.
	if werte.Get("state") != l.State {
		return fmt.Errorf(
			"%w: dieser Rückruf gehört nicht zu dieser Anmeldung",
			domain.ErrValidation)
	}

	if absage := werte.Get("error"); absage != "" {
		if absage == "access_denied" {
			return fmt.Errorf("%w: %s", ErrAbgesagt, werte.Get("error_description"))
		}

		return fmt.Errorf("%w: %s meldet %q: %s", domain.ErrValidation,
			l.anbieter.Name, absage, werte.Get("error_description"))
	}

	if werte.Get("code") == "" {
		return fmt.Errorf(
			"%w: der Rückruf trägt keinen Code", domain.ErrValidation)
	}

	return nil
}

// zufall gibt 32 zufällige Bytes als URL-sichere Zeichenkette.
//
// **32 Bytes**, weil PKCE (RFC 7636 §4.1) für den Prüfer 43 bis 128 Zeichen
// verlangt — 32 Bytes werden als base64url zu 43 Zeichen, der Untergrenze.
//
// **Ohne Fehlerrückgabe, und das ist nachgelesen**: `crypto/rand.Read` „never
// returns an error, and always fills b entirely" — es „crashes the program
// irrecoverably", wenn das Betriebssystem versagt. Ein `if err != nil` daneben
// wäre ein Zweig, den kein Lauf betreten kann, und §3 nennt genau das einen
// Fehlerpfad, der eine Behauptung ist.
func zufall() string {
	roh := make([]byte, 32)
	_, _ = rand.Read(roh) //nolint:errcheck // gibt nie einen Fehler, siehe oben

	return base64.RawURLEncoding.EncodeToString(roh)
}

// herausforderung rechnet den `code_challenge` aus dem Prüfer.
func herausforderung(pruefer string) string {
	summe := sha256.Sum256([]byte(pruefer))

	return base64.RawURLEncoding.EncodeToString(summe[:])
}
