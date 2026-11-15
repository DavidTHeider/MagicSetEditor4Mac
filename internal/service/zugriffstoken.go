// **Zugriffstoken** — Werkstück 10 von Phase 4.12
// (#1351),
// Punkt 13 der Kommentierung vom 2026-09-05.
//
// # Die Gestalt eines Tokens
//
// `mse_<kennung>_<geheimnis>` — die Kennung vorn, das Geheimnis hinten.
//
// **Warum die Kennung mitreist**: ohne sie müsste jede Prüfung **jede**
// gehütete Form durchprobieren. Mit ihr ist es ein Nachschlagen und **ein**
// Vergleich. Dieselbe Bauform benutzt GitHub für seine Zugriffstoken, und aus
// demselben Grund.
//
// **Die Kennung ist kein Geheimnis** — sie steht in jeder Liste. Was schützt,
// ist der hintere Teil.
//
// **Und der Trenner kommt im Geheimnis nicht vor** — das ist keine Hoffnung,
// sondern die Wahl des Alphabets: [crypto/rand.Text] gibt Base32 nach RFC
// 4648, also `A`–`Z` und `2`–`7`. Ein Unterstrich ist darin unmöglich.
//
// Der Satz steht hier, weil der erste Anlauf `base64.RawURLEncoding` nahm —
// und dessen Alphabet enthält `_`. Ein Trennen an jedem Unterstrich zerlegte
// das Geheimnis mit, **aber nur manchmal**: je nachdem, welche Bytes gezogen
// wurden. Angestrichen hat es `TestEinTokenWirdWiedererkannt`; von Hand wäre
// es ein Fehlschlag in etwa jedem vierten Fall gewesen und damit einer, den
// man für Flackern hält.
//
// # Warum SHA-256 und nicht Argon2id
//
// **Weil ein Token kein Kennwort ist.** Argon2id ist teuer, damit ein
// **geratenes** Kennwort teuer wird: Menschen wählen kurze, und ein Angreifer
// probiert Wörterbücher. Ein Token hier ist ein Zufallswert aus 32 Byte —
// 256 Bit —, und gegen 2²⁵⁶ Möglichkeiten hilft keine Streckung; sie kostet
// nur den, der prüft.
//
// **Was Argon2id trotzdem leistet und SHA-256 nicht**: es schützt schwache
// Geheimnisse. Deshalb bleibt das Kennwort (Werkstück 6) dort, wo es ist —
// die Entscheidung hier gilt für **zufällige** Geheimnisse und für nichts
// sonst.
//
// **Und der Vergleich läuft in konstanter Zeit** (`subtle`): ein Vergleich,
// der beim ersten falschen Byte abbricht, verrät seine Länge und Stelle.

package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Tokenmarke steht vor jeder Kennung.
//
// **Damit ein Token als solches erkennbar ist.** Wer eines versehentlich in
// eine Datei schreibt, soll es beim Durchsehen wiederfinden — und ein
// Suchmuster über einer Ablage findet es auch.
const Tokenmarke = "mse"

// Tokenablage ist die Naht zum Bestand.
type Tokenablage interface {
	TokenAnlegen(ctx context.Context, neu domain.Zugriffstoken, gehuetet string) (
		domain.Zugriffstoken, error)
	Token(ctx context.Context, profilID int64) ([]domain.Zugriffstoken, error)
	TokenMitKennung(ctx context.Context, tokenID int64) (
		domain.Zugriffstoken, domain.Gehuetetes, error)

	// **Rotation, letzte Nutzung und Sammelwiderruf** — Werkstück O2 von
	// Phase 5.2 (#1426).
	TokenRotieren(ctx context.Context, tokenID int64, gehuetet, vorgaengerBis string) (
		domain.Zugriffstoken, error)
	TokenNutzungVermerken(ctx context.Context, tokenID int64, zeitpunkt string) error
	TokenAlleLoeschen(ctx context.Context, profilID int64) (int64, error)
	TokenAendern(ctx context.Context, geaendert domain.Zugriffstoken) (
		domain.Zugriffstoken, error)
	TokenLoeschen(ctx context.Context, tokenID int64) error
}

// Tokenverwaltung legt Token an und prüft sie.
type Tokenverwaltung struct {
	ablage Tokenablage
	uhr    func() time.Time
}

// NeueTokenverwaltung baut sie über einer Ablage.
func NeueTokenverwaltung(ablage Tokenablage) *Tokenverwaltung {
	return &Tokenverwaltung{ablage: ablage, uhr: time.Now}
}

// MitUhr setzt die Uhr — für Läufe, die einen festen Zeitpunkt brauchen.
func (v *Tokenverwaltung) MitUhr(uhr func() time.Time) *Tokenverwaltung {
	v.uhr = uhr

	return v
}

// Anlegen legt ein Token an und gibt seinen Klartext **einmal**.
//
// **Danach gibt es ihn nicht mehr.** Der Bestand trägt die gehütete Form; wer
// den Klartext verliert, legt ein neues an. Dasselbe tun die
// Wiederherstellungscodes aus Werkstück 7 und aus demselben Grund: ein
// Geheimnis, das sich nachschlagen lässt, ist keines.
func (v *Tokenverwaltung) Anlegen(
	ctx context.Context, neu domain.Zugriffstoken,
) (domain.Zugriffstoken, string, error) {
	neu.Angelegt = v.uhr().UTC().Format(time.RFC3339)
	geheimnis := rand.Text()

	angelegt, ablagefehler := v.ablage.TokenAnlegen(ctx, neu, huete(geheimnis))
	if ablagefehler != nil {
		return domain.Zugriffstoken{}, "", ablagefehler
	}

	return angelegt, Tokenmarke + "_" + strconv.FormatInt(angelegt.ID, 10) +
		"_" + geheimnis, nil
}

// Alle gibt die Token eines Profils — ohne ihre Geheimnisse.
func (v *Tokenverwaltung) Alle(ctx context.Context, profilID int64) (
	[]domain.Zugriffstoken, error,
) {
	return v.ablage.Token(ctx, profilID)
}

// Aendern schreibt Bezeichnung, Scope und Ablauf fort.
//
// **Das Profil kommt aus dem Bestand und nicht vom Rufer.** Ein Token wechselt
// sein Profil nicht; wer es mitschickte, könnte es — und ein Token, das
// plötzlich woanders arbeitet, ist eine Rechteerweiterung und kein Umbenennen.
// Dieselbe Überlegung wie bei `App.ProfilAendern` und dem Eigner eines Profils.
//
// **Und dasselbe gilt für das Geheimnis**: [Tokenablage.TokenAendern] rührt es
// nicht an. Wer umbenennt, entzieht niemandem den Zugang.
func (v *Tokenverwaltung) Aendern(
	ctx context.Context, geaendert domain.Zugriffstoken,
) (domain.Zugriffstoken, error) {
	vorher, _, fehler := v.ablage.TokenMitKennung(ctx, geaendert.ID)
	if fehler != nil {
		return domain.Zugriffstoken{}, fehler
	}

	geaendert.ProfilID = vorher.ProfilID
	geaendert.Angelegt = vorher.Angelegt

	// **Ein Token wechselt seinen Sitz nicht** (#1819): die Bindung entsteht
	// mit ihm, und wer es umbenennt, rückt das Werkzeug dahinter nicht an
	// einen anderen Sitz.
	geaendert.Sitzbindung = vorher.Sitzbindung

	return v.ablage.TokenAendern(ctx, geaendert)
}

// Loeschen nimmt ein Token heraus.
func (v *Tokenverwaltung) Loeschen(ctx context.Context, tokenID int64) error {
	return v.ablage.TokenLoeschen(ctx, tokenID)
}

// ProfilDes sagt, welchem Profil ein Token gehört.
//
// **Die Frage vor jedem Ruf mit einer Tokenkennung**
// (#1817):
// wer ein Token ändert, rotiert oder löscht, braucht die Rolle an **dessen**
// Profil. Das gezeigte ist ein anderes, sobald jemand die Kennung eines
// fremden Tokens nennt.
func (v *Tokenverwaltung) ProfilDes(ctx context.Context, tokenID int64) (int64, error) {
	eines, _, fehler := v.ablage.TokenMitKennung(ctx, tokenID)

	return eines.ProfilID, fehler
}

// Pruefe sagt, zu welchem Profil ein Klartext gehört — oder warum nicht.
//
// # Eine Meldung für alle Fehlschläge
//
// **Weil jede Unterscheidung eine Auskunft an den wäre, der rät.** „Diese
// Kennung gibt es nicht" sagte ihm, welche es gibt; „abgelaufen" sagte ihm,
// dass er beinahe richtig lag. Dieselbe Überlegung wie bei der Anmeldung
// (Werkstück 6).
//
// **Der Scope ist die Ausnahme, und zwar mit Grund**: wer ein gültiges Token
// hat, ist kein Fremder. Ihm zu sagen, dass sein Token nur lesen darf, ist
// eine Auskunft an einen Berechtigten — und ohne sie suchte er den Fehler bei
// seiner Anfrage.
func (v *Tokenverwaltung) Pruefe(
	ctx context.Context, klartext string, verlangt domain.Scope,
) (domain.Zugriffstoken, error) {
	kennung, geheimnis, zerlegt := zerlegeToken(klartext)
	if !zerlegt {
		return domain.Zugriffstoken{}, abgewiesenesToken()
	}

	eines, gehuetet, fehler := v.ablage.TokenMitKennung(ctx, kennung)
	if fehler != nil {
		return domain.Zugriffstoken{}, abgewiesenesToken()
	}

	jetzt := v.uhr()

	if !gehuetetPasst(huete(geheimnis), gehuetet, eines.VorgaengerGilt(jetzt)) {
		return domain.Zugriffstoken{}, abgewiesenesToken()
	}

	if eines.Abgelaufen(jetzt) {
		return domain.Zugriffstoken{}, abgewiesenesToken()
	}

	if !eines.Scope.Reicht(verlangt) {
		return domain.Zugriffstoken{}, fmt.Errorf(
			"%w: dieses Token darf %s, verlangt ist %s",
			domain.ErrErlaubnis, eines.Scope, verlangt)
	}

	// **Die Nutzung wird vermerkt, nachdem alles gepasst hat** — ein
	// abgewiesenes Token hinterlässt keine Spur, die aussähe, als sei es
	// benutzt worden. Und ein Vermerk, der nicht geht, ist ein Fehler des
	// Bestands und kein stilles Weiter (§8).
	if fehler := v.ablage.TokenNutzungVermerken(ctx, eines.ID,
		jetzt.UTC().Format(time.RFC3339)); fehler != nil {
		return domain.Zugriffstoken{}, fmt.Errorf("die Nutzung des Tokens vermerken: %w", fehler)
	}

	return eines, nil
}

// gehuetetPasst vergleicht in konstanter Zeit — gegen das geltende Geheimnis
// und, solange die Frist läuft, gegen das vorige.
//
// **Beide Vergleiche laufen immer**, auch wenn der erste schon passt: sonst
// verriete die Laufzeit, welches der beiden getroffen hat.
func gehuetetPasst(gehuetet string, abgelegt domain.Gehuetetes, vorgaengerGilt bool) bool {
	aktuell := subtle.ConstantTimeCompare([]byte(gehuetet), []byte(abgelegt.Aktuell)) == 1
	vorgaenger := subtle.ConstantTimeCompare([]byte(gehuetet), []byte(abgelegt.Vorgaenger)) == 1

	return aktuell || (vorgaengerGilt && vorgaenger)
}

// abgewiesenesToken ist der eine Satz für jeden Fehlschlag.
func abgewiesenesToken() error {
	return fmt.Errorf("%w: dieses Token gilt nicht", domain.ErrValidation)
}

// huete gibt die Form, die im Bestand steht.
func huete(geheimnis string) string {
	summe := sha256.Sum256([]byte(geheimnis))

	return hex.EncodeToString(summe[:])
}

// zerlegeToken zieht Kennung und Geheimnis aus einem Klartext.
//
// **`SplitN` mit 3 und nicht `Split`**, obwohl das Alphabet oben den Trenner
// ausschließt. Es kostet nichts, und es hält auch dann, wenn jemand später ein
// anderes Alphabet wählt — die Zusage steht dann an **einer** Stelle statt an
// zweien, die zusammenpassen müssen.
func zerlegeToken(klartext string) (kennung int64, geheimnis string, zerlegt bool) {
	teile := strings.SplitN(strings.TrimSpace(klartext), "_", 3)
	if len(teile) != 3 || teile[0] != Tokenmarke {
		return 0, "", false
	}

	gelesen, err := strconv.ParseInt(teile[1], 10, 64)
	if err != nil || gelesen <= 0 {
		return 0, "", false
	}

	return gelesen, teile[2], true
}
