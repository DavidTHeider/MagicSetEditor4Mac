// Package anmeldung führt einen Menschen über einen fremden Anbieter herein.
//
// **Anmeldung über einen Anbieter** — Werkstück 4 von Phase 4.12
// (#1351),
// nach Vermerk [0066](../../docs/entscheidungen/0066-nutzer-profile-und-die-vier-entscheidungen.md).
//
// # Ein Blatt (§2)
//
// Dieses Paket kennt `internal/domain` und sonst nichts aus dem Projekt. Es
// weiß, wie ein OAuth-Fluss geht — nicht, wer ein Nutzer ist und wo er liegt.
//
// # Drei Schichten, und die dritte ist die, die man weglässt (§16)
//
// **Transport** (`http.Client`, Frist, Rückruf), **fremdes DTO** (die Gestalt,
// die der Anbieter schickt) und **Zuordnung** auf `domain`. Wer den fremden
// Verbund weiterreicht, hat die Struktur eines Anbieters in seiner Fachlogik.
//
// # Zwei Gestalten, vier Anbieter
//
// | Gestalt | Wer | Die Identität kommt |
// | --- | --- | --- |
// | **OIDC** | Google, Microsoft | aus dem `id_token` |
// | **OAuth pur** | GitHub, X | aus einem **zweiten Ruf** |
//
// # Und die Kennungen liefert dieses Projekt nicht mit
//
// **Wer sich anmelden will, registriert seine eigene Anwendung** beim
// Anbieter und trägt Kennung und — wo nötig — Geheimnis ein. Ein
// mitgeliefertes Geheimnis wäre keines: es stünde in einem öffentlichen
// Repositorium und in jedem ausgelieferten Bündel.
//
// **Wo das Geheimnis liegt, entscheidet nicht dieses Paket** — es bekommt es
// gereicht. Im Betrieb kommt es aus dem Schlüsselbund
// ([0056](../../docs/entscheidungen/0056-drei-weitere-kartendienste.md)), und
// zwar aus demselben Grund wie jedes andere: die Bestandsdatei ist die, die
// man weitergibt.
package anmeldung

import (
	"errors"
	"fmt"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// ErrAbgesagt heißt: der Mensch hat im Browser abgebrochen.
//
// **Ein eigener Fehler und keine Meldung.** „Abbrechen" ist eine Entscheidung;
// sie als Fehler zu zeigen wäre eine Falschauskunft (§8).
var ErrAbgesagt = errors.New("die Anmeldung wurde abgebrochen")

// Gestalt sagt, woher die Identität kommt.
type Gestalt int

const (
	// OIDC nimmt sie aus dem `id_token` des Tauschendpunkts.
	OIDC Gestalt = iota + 1

	// OAuth nimmt sie aus einem zweiten Ruf an eine Nutzerabfrage.
	OAuth
)

// Anbieter ist ein Anmeldeweg, wie ihn ein Anbieter beschreibt.
//
// **Ein Verbund und keine Schnittstelle.** Was die vier unterscheidet, sind
// Adressen und zwei Schalter — vier Umsetzungen wären vier Sonderfälle für
// dieselbe Sache.
type Anbieter struct {
	// Name ist der Name aus [domain.Anbieter].
	Name string

	// Anmeldeadresse und Tauschadresse sind die beiden Endpunkte.
	Anmeldeadresse string
	Tauschadresse  string

	// Nutzeradresse ist der zweite Ruf — nur bei [OAuth].
	Nutzeradresse string

	// Bereiche sind die `scope`-Werte.
	Bereiche []string

	Gestalt Gestalt

	// Aussteller ist der `iss`, den ein `id_token` tragen muss — leer bei
	// [OAuth], denn dort gibt es keines.
	//
	// **Er wird verglichen und nicht erschlossen**: OIDC Core §3.1.3.7,
	// Punkt 2, verlangt „MUST exactly match", und bei Microsoft weicht er von
	// der Anmeldeadresse ab. Woher die Werte stammen, steht bei
	// [AusstellerGoogle].
	Aussteller string

	// Kennung ist die `client_id`, Geheimnis das `client_secret`.
	Kennung   string
	Geheimnis string

	// GeheimnisPflicht sagt, ob der Tauschendpunkt eines verlangt.
	//
	// **Je Anbieter nachgelesen und nicht geraten** (2026-09-05): GitHub führt
	// `client_secret` als *Required*, auch mit PKCE. Google nennt es für
	// Desktop-Klienten *optional*. Microsoft **verbietet** es für öffentliche
	// Klienten: „Public clients, which include native applications and single
	// page apps, must not use secrets or certificates when redeeming an
	// authorization code."
	GeheimnisPflicht bool

	// GeheimnisVerboten ist die andere Richtung — Microsoft.
	GeheimnisVerboten bool
}

// GitHub gibt den Anbieter GitHub.
//
// **OAuth pur**: die Identität kommt aus `GET /user`, denn GitHub gibt kein
// `id_token`. **Und das Geheimnis ist Pflicht** — nachgelesen am 2026-09-05.
func GitHub(kennung, geheimnis string) Anbieter {
	return Anbieter{
		Name:             "github",
		Anmeldeadresse:   "https://github.com/login/oauth/authorize",
		Tauschadresse:    "https://github.com/login/oauth/access_token",
		Nutzeradresse:    "https://api.github.com/user",
		Bereiche:         []string{"read:user"},
		Gestalt:          OAuth,
		Kennung:          kennung,
		Geheimnis:        geheimnis,
		GeheimnisPflicht: true,
	}
}

// Die Aussteller der beiden OIDC-Anbieter.
//
// **Am 2026-09-06 aus ihren Entdeckungsdokumenten gelesen**, nicht erinnert
// (§1): `https://accounts.google.com/.well-known/openid-configuration` und
// `https://login.microsoftonline.com/consumers/v2.0/.well-known/openid-configuration`,
// je das Feld `issuer`.
//
// **Microsofts Aussteller nennt eine GUID und nicht `consumers`.** Der
// Anmeldeweg heißt `/consumers/…`, der Aussteller im Token trägt den
// Mandanten für persönliche Konten aus. Wer ihn aus der Adresse erschlösse,
// bekäme einen Wert, den kein Token je trägt — und wiese jede Anmeldung ab.
const (
	AusstellerGoogle = "https://accounts.google.com"

	AusstellerMicrosoft = "https://login.microsoftonline.com/" +
		"9188040d-6c67-4c5b-b112-36a304b66dad/v2.0"
)

// Google gibt den Anbieter Google.
//
// **OIDC**, und das Geheimnis ist für Desktop-Klienten *optional* — deshalb
// weder Pflicht noch verboten.
func Google(kennung, geheimnis string) Anbieter {
	return Anbieter{
		Name:           "google",
		Anmeldeadresse: "https://accounts.google.com/o/oauth2/v2/auth",
		Tauschadresse:  "https://oauth2.googleapis.com/token",
		Bereiche:       []string{"openid", "profile"},
		Gestalt:        OIDC,
		Aussteller:     AusstellerGoogle,
		Kennung:        kennung,
		Geheimnis:      geheimnis,
	}
}

// Microsoft gibt den Anbieter Microsoft für persönliche Konten.
//
// **`consumers` und nicht `common`**: das Issue nennt ausdrücklich das
// persönliche Konto. `common` ließe auch Organisationskonten zu, und wer eines
// benutzt, meldete sich mit einer Kennung an, die sein Arbeitgeber abschalten
// kann.
func Microsoft(kennung, geheimnis string) Anbieter {
	return Anbieter{
		Name:           "microsoft",
		Anmeldeadresse: "https://login.microsoftonline.com/consumers/oauth2/v2.0/authorize",
		Tauschadresse:  "https://login.microsoftonline.com/consumers/oauth2/v2.0/token",
		Bereiche:       []string{"openid", "profile"},
		Gestalt:        OIDC,
		Aussteller:     AusstellerMicrosoft,
		Kennung:        kennung,
		Geheimnis:      geheimnis,

		// **Verboten, nicht bloß entbehrlich** — siehe [Anbieter.GeheimnisPflicht].
		GeheimnisVerboten: true,
	}
}

// X gibt den Anbieter X.
//
// **OAuth pur**, und `state` ist dort **Pflicht** — hier ohnehin bei allen.
//
// **Was nicht geprüft ist**, gehört danebengeschrieben (Vermerk 0066): die
// freie Stufe der X-API führt 100 Reads im Monat, und ob eine Nutzerabfrage im
// Nutzerkontext dagegen zählt, ist **nicht gemessen**.
func X(kennung, geheimnis string) Anbieter {
	return Anbieter{
		Name:           "x",
		Anmeldeadresse: "https://twitter.com/i/oauth2/authorize",
		Tauschadresse:  "https://api.twitter.com/2/oauth2/token",
		Nutzeradresse:  "https://api.twitter.com/2/users/me",
		Bereiche:       []string{"users.read", "tweet.read"},
		Gestalt:        OAuth,
		Kennung:        kennung,
		Geheimnis:      geheimnis,
	}
}

// Startbereit sagt, ob ein Anmeldeweg mit diesem Anbieter beginnen kann —
// **dieselbe Prüfung wie in [Beginne]** (Nachlauf über ID-05, Befund 4). Die
// Oberfläche nennt einen Anbieter „eingerichtet“ erst, wenn sie besteht;
// vorher genügte eine `client_id`, und GitHub ohne Geheimnis hieß eingerichtet
// und scheiterte beim ersten Klick.
func (a Anbieter) Startbereit() error { return a.pruefe() }

// pruefe sagt, ob dieser Anbieter benutzbar ist.
func (a Anbieter) pruefe() error {
	if !domain.KenntAnbieter(a.Name) {
		return fmt.Errorf("%w: %q ist keiner der Anbieter %v",
			domain.ErrValidation, a.Name, domain.Anbieter())
	}

	if strings.TrimSpace(a.Kennung) == "" {
		return fmt.Errorf(
			"%w: für %s ist keine Kennung eingetragen — dieses Programm liefert "+
				"keine mit, wer sich anmelden will, registriert seine eigene Anwendung",
			domain.ErrValidation, a.Name)
	}

	if a.GeheimnisPflicht && strings.TrimSpace(a.Geheimnis) == "" {
		return fmt.Errorf(
			"%w: %s verlangt beim Tausch ein Geheimnis, auch mit PKCE",
			domain.ErrValidation, a.Name)
	}

	if a.GeheimnisVerboten && strings.TrimSpace(a.Geheimnis) != "" {
		return fmt.Errorf(
			"%w: %s weist ein Geheimnis von einem öffentlichen Klienten ab",
			domain.ErrValidation, a.Name)
	}

	return nil
}
