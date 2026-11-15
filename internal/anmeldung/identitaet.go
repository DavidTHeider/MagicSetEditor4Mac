// **Woher die Identität kommt** — Werkstück 4 von Phase 4.12
// (#1351).
//
// # Zwei Wege, und das ist der ganze Unterschied zwischen den vier Anbietern
//
// **OIDC** (Google, Microsoft) legt sie in ein `id_token`, das der
// Tauschendpunkt mitschickt. **OAuth pur** (GitHub, X) hat keines; dort kostet
// die Identität einen zweiten Ruf.

package anmeldung

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// ausOIDC liest die Identität aus dem `id_token`.
//
// # Die Signatur wird **nicht** geprüft, und das ist eine Entscheidung
//
// **OIDC Core 3.1.3.7 erlaubt es ausdrücklich**: „If the ID Token is received
// via direct communication between the Client and the Token Endpoint (which it
// is in this flow), the TLS server validation MAY be used to validate the
// issuer in place of checking the token signature."
//
// **Was das kostet, gehört danebengeschrieben**: die Identität hängt an der
// TLS-Kette des Tauschendpunkts. Wer ein gültiges Zertifikat dafür vorweisen
// kann, kann sie fälschen — er könnte dann allerdings ebenso den ganzen Tausch
// beantworten. Der Preis ist derselbe wie für den Fluss überhaupt, und dafür
// entfallen JWKS, Schlüsselwechsel und eine RSA-Prüfung.
//
// # Was **stattdessen** geprüft wird — und zwar alles davon
//
// Dieselbe Stelle der Spezifikation verlangt drei Dinge, die von der Signatur
// unabhängig sind:
//
//   - **Punkt 2 — der Aussteller.** „The Issuer Identifier … MUST exactly
//     match the value of the iss (issuer) Claim." Er steht je Anbieter in
//     [Anbieter.Aussteller] und ist **nachgelesen**, nicht erschlossen.
//   - **Punkt 3 — die Zielgruppe.** „The Client MUST validate that the aud
//     (audience) Claim contains its client_id." Ohne diese Prüfung nähme
//     dieses Programm ein Token an, das ein **fremder** Klient beim selben
//     Anbieter für sich hat ausstellen lassen.
//   - **Punkt 11 — der `nonce`.** „If a nonce value was sent in the
//     Authentication Request, a nonce Claim MUST be present." Ein „wenn
//     einer dasteht" ließe ein Token **ohne** `nonce` durch — genau die
//     Wiedereinspielung, gegen die er da ist.
//   - **Punkt 9 — der Ablauf.** „The current time MUST be before the time
//     represented by the exp Claim." Ein MUSS, das ohne Signaturprüfung
//     nachvollziehbar ist: `exp` steht im Rumpf, und die Uhr genügt.
//   - **Punkte 4 und 5 — mehrere Zielgruppen und `azp`.** „If the ID Token
//     contains multiple audiences, the Client SHOULD verify that an azp Claim
//     is present" und „If an azp Claim is present, the Client SHOULD verify
//     that its client_id is the Claim Value."
//
// **Die letzten beiden sind ein SOLLTE der Spezifikation und hier ein Muss.**
// „Mehrere Zielgruppen" heißt: dasselbe Token gilt auch woanders. Ohne `azp`
// sagt nichts, **für wen** es angefordert wurde — und dieses Programm hat
// keinen Grund, eines anzunehmen, das ein anderer geholt hat.
//
// **Die drei kamen am 2026-09-06 dazu**, angestrichen vom Nachlauf zum
// Zweitblick über Phase 4.12. `iat` steht bewusst nicht dabei: die
// Spezifikation lässt dort ausdrücklich Spielraum („The iat Claim can be used
// to reject tokens that were issued too far away from the current time,
// limiting the amount of time that nonces need to be stored"), und eine
// Grenze, die niemand nachgelesen hat, wäre eine erfundene.
func (l *Lauf) ausOIDC(jeton string) (domain.Fremdkennung, error) {
	if jeton == "" {
		return domain.Fremdkennung{}, fmt.Errorf("%w: %s gab kein id_token zurück",
			domain.ErrValidation, l.anbieter.Name)
	}

	teile := strings.Split(jeton, ".")
	if len(teile) != 3 {
		return domain.Fremdkennung{}, fmt.Errorf("%w: das id_token von %s hat %d Teile statt drei",
			domain.ErrValidation, l.anbieter.Name, len(teile))
	}

	roh, entpackfehler := base64.RawURLEncoding.DecodeString(teile[1])
	if entpackfehler != nil {
		// **Beide Fehler, nicht einer.** Die Klasse sagt der Oberfläche, wie
		// sie es zeigt; der Grund sagt dem Protokoll, was schiefging.
		// `errors.Is` findet danach beide.
		return domain.Fremdkennung{}, errors.Join(fmt.Errorf(
			"%w: der Rumpf des id_token ist kein base64url", domain.ErrValidation),
			entpackfehler)
	}

	var ansprueche struct {
		Subjekt     string    `json:"sub"`
		Name        string    `json:"name"`
		Nonce       string    `json:"nonce"`
		Aussteller  string    `json:"iss"`
		Zielgruppe  zielkreis `json:"aud"`
		Angefordert string    `json:"azp"`

		// Ablauf ist `exp` in Sekunden seit 1970 — **0 heißt: keiner steht
		// da**, und das ist eine Abweisung und keine Ewigkeit.
		Ablauf int64 `json:"exp"`
	}

	if err := json.Unmarshal(roh, &ansprueche); err != nil {
		return domain.Fremdkennung{}, errors.Join(fmt.Errorf(
			"%w: der Rumpf des id_token ist kein JSON", domain.ErrValidation), err)
	}

	if ansprueche.Aussteller != l.anbieter.Aussteller {
		return domain.Fremdkennung{}, fmt.Errorf(
			"%w: das id_token stammt nicht von %s", domain.ErrValidation, l.anbieter.Name)
	}

	if !ansprueche.Zielgruppe.enthaelt(l.anbieter.Kennung) {
		return domain.Fremdkennung{}, fmt.Errorf(
			"%w: das id_token ist für einen anderen Klienten ausgestellt",
			domain.ErrValidation)
	}

	// **Mehrere Zielgruppen brauchen ein `azp`, und es muss auf uns zeigen.**
	if len(ansprueche.Zielgruppe) > 1 &&
		ansprueche.Angefordert != l.anbieter.Kennung {
		return domain.Fremdkennung{}, fmt.Errorf(
			"%w: das id_token gilt für mehrere Klienten und nennt nicht diesen "+
				"als den, der es angefordert hat", domain.ErrValidation)
	}

	// **Abgelaufen ist abgelaufen** — und ohne `exp` gilt es als abgelaufen.
	// **`time.Now` unmittelbar und keine Naht davor.** Eine Uhr als Variable
	// wäre eine Naht, die kein Lauf benutzt: die beiden Zweige lassen sich
	// mit einem Zeitpunkt in der Vergangenheit und einem in der Zukunft
	// betreten, und eine Naht ohne Aufrufer ist keine Zusage (§3).
	if ansprueche.Ablauf == 0 ||
		!time.Now().Before(time.Unix(ansprueche.Ablauf, 0)) {
		return domain.Fremdkennung{}, fmt.Errorf(
			"%w: das id_token von %s ist abgelaufen",
			domain.ErrValidation, l.anbieter.Name)
	}

	// **Vorhanden **und** gleich.** Der erste Teil ist der, der gefehlt hat:
	// wir senden bei OIDC immer einen `nonce`, also muss auch einer
	// zurückkommen.
	if ansprueche.Nonce != l.nonce {
		return domain.Fremdkennung{}, fmt.Errorf(
			"%w: das id_token gehört zu einer anderen Anmeldung", domain.ErrValidation)
	}

	if ansprueche.Subjekt == "" {
		return domain.Fremdkennung{}, fmt.Errorf("%w: das id_token von %s nennt kein sub",
			domain.ErrValidation, l.anbieter.Name)
	}

	return domain.Fremdkennung{
		Anbieter:    l.anbieter.Name,
		Subjekt:     ansprueche.Subjekt,
		Anzeigename: ansprueche.Name,
	}, nil
}

// ausNutzerruf holt sie mit einem zweiten Ruf.
func (l *Lauf) ausNutzerruf(ctx context.Context, rufer *Rufer, zugangswort string) (
	domain.Fremdkennung, error,
) {
	anfrage, baufehler := http.NewRequestWithContext(ctx, http.MethodGet,
		l.anbieter.Nutzeradresse, http.NoBody)
	if baufehler != nil {
		return domain.Fremdkennung{}, fmt.Errorf("die Nutzerabfrage bauen: %w", baufehler)
	}

	anfrage.Header.Set("Authorization", "Bearer "+zugangswort)
	anfrage.Header.Set("Accept", "application/json")
	anfrage.Header.Set("User-Agent", kenner)

	antwort, ruffehler := rufer.unten.Do(anfrage)
	if ruffehler != nil {
		return domain.Fremdkennung{}, fmt.Errorf("%s antwortet nicht: %w", l.anbieter.Name, ruffehler)
	}

	defer func() { _ = antwort.Body.Close() }() //nolint:errcheck // gelesen ist gelesen

	if antwort.StatusCode < 200 || antwort.StatusCode >= 300 {
		return domain.Fremdkennung{}, fmt.Errorf("%w: %s weist die Nutzerabfrage ab (%d)",
			domain.ErrValidation, l.anbieter.Name, antwort.StatusCode)
	}

	roh, lesefehler := io.ReadAll(io.LimitReader(antwort.Body, hoechstens))
	if lesefehler != nil {
		return domain.Fremdkennung{}, fmt.Errorf("die Nutzerabfrage von %s lesen: %w",
			l.anbieter.Name, lesefehler)
	}

	return l.deuteNutzer(roh)
}

// nutzerantwort deckt beide Gestalten, die ein zweiter Ruf schickt.
//
// **Ein Verbund für GitHub und X.** GitHub antwortet flach mit `id` als
// **Zahl**, X packt seines in `data` und führt `id` als **Zeichenkette** —
// deshalb `json.Number`, das beides nimmt.
type nutzerantwort struct {
	ID    json.Number `json:"id"`
	Name  string      `json:"name"`
	Login string      `json:"login"`

	Daten *struct {
		ID         json.Number `json:"id"`
		Name       string      `json:"name"`
		Nutzername string      `json:"username"`
	} `json:"data"`
}

// deuteNutzer macht aus der Antwort eine Kennung.
func (l *Lauf) deuteNutzer(roh []byte) (domain.Fremdkennung, error) {
	var gelesen nutzerantwort

	if err := json.Unmarshal(roh, &gelesen); err != nil {
		return domain.Fremdkennung{}, errors.Join(fmt.Errorf("%w: %s antwortete nicht mit JSON",
			domain.ErrValidation, l.anbieter.Name), err)
	}

	subjekt, name := gelesen.ID.String(), erster(gelesen.Name, gelesen.Login)

	if gelesen.Daten != nil {
		subjekt = gelesen.Daten.ID.String()
		name = erster(gelesen.Daten.Name, gelesen.Daten.Nutzername)
	}

	if subjekt == "" || subjekt == "0" {
		return domain.Fremdkennung{}, fmt.Errorf("%w: %s nennt keine Kennung für diesen Menschen",
			domain.ErrValidation, l.anbieter.Name)
	}

	return domain.Fremdkennung{Anbieter: l.anbieter.Name, Subjekt: subjekt, Anzeigename: name}, nil
}

// erster gibt den ersten nichtleeren Wert.
func erster(werte ...string) string {
	for _, wert := range werte {
		if strings.TrimSpace(wert) != "" {
			return wert
		}
	}

	return ""
}

// zielkreis ist der `aud`-Anspruch — eine Zeichenkette **oder** eine Liste.
//
// **Beides**, und zwar nach OIDC Core §2: „The aud value is an array of
// case-sensitive strings … In the special case when the ID Token has one
// audience, the aud value MAY be a single case-sensitive string." Ein Typ, der
// nur eines von beiden nähme, wiese die Hälfte der Anbieter ab.
type zielkreis []string

// UnmarshalJSON liest beide Gestalten.
func (z *zielkreis) UnmarshalJSON(roh []byte) error {
	var einer string
	if err := json.Unmarshal(roh, &einer); err == nil {
		*z = zielkreis{einer}

		return nil
	}

	var mehrere []string
	if err := json.Unmarshal(roh, &mehrere); err != nil {
		return fmt.Errorf("%w: das Feld aud ist weder Zeichenkette noch Liste",
			domain.ErrValidation)
	}

	*z = mehrere

	return nil
}

// enthaelt sagt, ob die eigene Kennung dabei ist.
func (z zielkreis) enthaelt(kennung string) bool {
	return slices.Contains(z, kennung)
}
