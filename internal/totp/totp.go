// Package totp rechnet zeitbasierte Einmalkennwörter nach RFC 6238.
//
// **Ein Blatt** (§2): es kennt `internal/domain` für seine Fehlerklassen und
// sonst nichts aus diesem Projekt. Werkstück 7 von Phase 4.12
// (#1351),
// entschieden in
// [0066](../../docs/entscheidungen/0066-nutzer-profile-und-die-vier-entscheidungen.md).
//
// # Warum selbst gerechnet und keine Bibliothek
//
// **Weil es fünfzig Zeilen sind und eine amtliche Probe dazu gibt.** RFC 6238
// nennt in Anhang B achtzehn Vektoren; eine Umsetzung, die sie alle trifft,
// ist dieselbe wie jede andere. Eine Abhängigkeit dafür wäre ein Vertrauen
// mehr, das niemand nachrechnet.
//
// # Was hier **nicht** steht
//
// **Wo ein Geheimnis liegt.** Es gehört in den Schlüsselbund
// (`internal/schluesselbund`) und nicht in den Bestand — die Bestandsdatei ist
// die, die man weitergibt, und wer das Geheimnis hat, macht jeden Code.
package totp

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1" //#nosec G505 -- RFC 6238 §1.2: HMAC-SHA-1 ist das Verfahren
	"crypto/sha256"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"hash"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Verfahren ist die Hashfunktion unter dem HMAC.
type Verfahren string

// Die drei Verfahren aus RFC 6238 §1.2.
//
// **SHA-1 ist der Normalfall**, und das ist keine Nachlässigkeit: die
// Authenticator-Apps lesen `algorithm=SHA1`, und die Sicherheit von HMAC-SHA-1
// hängt nicht an der Kollisionsresistenz von SHA-1 (RFC 6151 §2.3). Die
// anderen beiden stehen da, weil die RFC sie nennt und weil ihre Vektoren die
// Umsetzung mitprüfen.
const (
	SHA1   Verfahren = "SHA1"
	SHA256 Verfahren = "SHA256"
	SHA512 Verfahren = "SHA512"
)

// Die Vorgaben, die diese Anwendung setzt.
const (
	// Schritt ist X aus RFC 6238 §4.1 — dreißig Sekunden.
	Schritt = 30 * time.Second

	// Stellen ist, wie viele Ziffern ein Code hat, wenn niemand es sagt.
	Stellen = 6

	// Fenster ist, wie viele Schritte in jede Richtung noch gelten.
	//
	// **Eine Entscheidung dieses Prüfers und keine Vorgabe der RFC** (§5.2,
	// die es ausdrücklich offenlässt und zur Sparsamkeit rät). **Einer**:
	// das deckt eine um bis zu dreißig Sekunden abweichende Uhr ab und
	// verlängert die Frist eines abgefangenen Codes auf höchstens neunzig
	// Sekunden.
	Fenster = 1

	// Geheimnislaenge ist, wie viele Byte ein frisches Geheimnis wiegt.
	//
	// **Zwanzig**, die Länge eines SHA-1-Blocks und das, was RFC 4226 §4 R6
	// als Mindestmaß nennt; in base32 sind das die 32 Zeichen, die eine
	// Authenticator-App erwartet.
	Geheimnislaenge = 20

	// Aussteller steht in der Adresse für den Scanner.
	Aussteller = "MagicSetEditor4Mac"
)

// Zaehler ist alles, was eine Rechnung braucht.
//
// **Ein Verbund und keine fünf Argumente**: die Felder gehören zusammen, und
// wer sie einzeln reicht, vertauscht irgendwann zwei.
type Zaehler struct {
	// Geheimnis ist der rohe Schlüssel — **nicht** seine base32-Gestalt.
	Geheimnis []byte

	// Verfahren ist die Hashfunktion; leer heißt [SHA1].
	Verfahren Verfahren

	// Stellen ist die Zahl der Ziffern; null heißt [Stellen].
	Stellen int
}

// Code gibt die Ziffern, die zu diesem Zeitpunkt gehören.
//
// **Ohne Fehlerrückgabe**: jede Eingabe hat ein Ergebnis. Ein leeres
// Geheimnis gibt Ziffern, die zu nichts passen — und das ist richtig so, denn
// **ob** jemand ein Geheimnis hat, ist keine Frage an eine Rechnung.
func Code(zaehler Zaehler, jetzt time.Time) string {
	schritte := uint64(jetzt.Unix()) / uint64(Schritt.Seconds()) //#nosec G115 -- Zeit vor 1970 gibt es hier nicht

	roh := make([]byte, 8)
	binary.BigEndian.PutUint64(roh, schritte)

	rechner := hmac.New(bauer(zaehler.Verfahren), zaehler.Geheimnis)
	rechner.Write(roh)
	summe := rechner.Sum(nil)

	// **Die dynamische Kürzung aus RFC 4226 §5.3**: die untersten vier Bit
	// des letzten Byte sagen, wo die vier Byte beginnen, und das oberste Bit
	// davon fällt weg.
	ab := summe[len(summe)-1] & 0x0f
	zahl := binary.BigEndian.Uint32(summe[ab:ab+4]) & 0x7fffffff

	stellen := zaehler.Stellen
	if stellen <= 0 {
		stellen = Stellen
	}

	return fmt.Sprintf("%0*d", stellen, zahl%teiler(stellen))
}

// Stimmt sagt, ob dieser Code zu diesem Zeitpunkt gehört.
//
// **Der Vergleich dauert immer gleich lang**, aus demselben Grund wie in
// `internal/kennwort`: wer die Zeit misst, liest sonst Ziffer für Ziffer ab.
func Stimmt(zaehler Zaehler, code string, jetzt time.Time) error {
	_, fehler := PassenderSchritt(zaehler, code, jetzt)

	return fehler
}

// PassenderSchritt gibt den Zeitschritt, zu dem dieser Code gehört — im
// Fenster um `jetzt`.
//
// **Der Schritt des Codes und nicht der der Uhr** (SEC-09, #2150): wer einen
// Code verbraucht, verbraucht genau diesen Schritt, und RFC 6238 §5.2 verlangt,
// dass derselbe danach nicht mehr gilt.
func PassenderSchritt(zaehler Zaehler, code string, jetzt time.Time) (int64, error) {
	stellen := zaehler.Stellen
	if stellen <= 0 {
		stellen = Stellen
	}

	if len(code) != stellen || strings.TrimFunc(code, istZiffer) != "" {
		return 0, fmt.Errorf("%w: ein Code hat %d Ziffern und sonst nichts",
			domain.ErrValidation, stellen)
	}

	for versatz := -Fenster; versatz <= Fenster; versatz++ {
		zeitpunkt := jetzt.Add(time.Duration(versatz) * Schritt)
		gerechnet := Code(zaehler, zeitpunkt)

		if subtle.ConstantTimeCompare([]byte(gerechnet), []byte(code)) == 1 {
			return zeitpunkt.Unix() / int64(Schritt.Seconds()), nil
		}
	}

	return 0, ErrStimmtNicht
}

// ErrStimmtNicht heißt: die Form ist in Ordnung, der Code nicht.
//
// **Ein eigener Fehler und keine Validierung** — dieselbe Trennung wie in
// `internal/kennwort`: „noch einmal ablesen" ist etwas anderes als „das ist
// kein Code".
var ErrStimmtNicht = fmt.Errorf("%w: der Code stimmt nicht", domain.ErrValidation)

// NeuesGeheimnis zieht ein Geheimnis.
//
// **Ohne Fehlerrückgabe**, weil `crypto/rand.Read` seit Go 1.24 keinen Fehler
// mehr gibt — dieselbe Stelle wie in `internal/anmeldung` und `internal/kennwort`.
func NeuesGeheimnis() []byte {
	roh := make([]byte, Geheimnislaenge)
	rand.Read(roh) //nolint:errcheck // gibt seit Go 1.24 keinen Fehler

	return roh
}

// AlsText gibt das Geheimnis in der Gestalt, die eine App erwartet.
//
// **Base32 ohne Auffüllen und in Großbuchstaben** — so steht es im
// Key-Uri-Format, und so nimmt es jeder Scanner entgegen.
func AlsText(geheimnis []byte) string {
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(geheimnis)
}

// GeheimnisAus liest die base32-Gestalt zurück.
func GeheimnisAus(text string) ([]byte, error) {
	roh, fehler := base32.StdEncoding.WithPadding(base32.NoPadding).
		DecodeString(strings.ToUpper(strings.TrimSpace(text)))
	if fehler != nil {
		return nil, fmt.Errorf(
			"%w: das Geheimnis ist kein base32", domain.ErrValidation)
	}

	return roh, nil
}

// Adresse gibt die `otpauth://`-Adresse für einen Scanner.
//
// **Nach dem Key-Uri-Format**, wie Google Authenticator, 1Password und Aegis
// es lesen: `otpauth://totp/<Aussteller>:<Konto>` mit `secret`, `issuer`,
// `algorithm`, `digits` und `period`.
func Adresse(zaehler Zaehler, konto string) string {
	verfahren := zaehler.Verfahren
	if verfahren == "" {
		verfahren = SHA1
	}

	stellen := zaehler.Stellen
	if stellen <= 0 {
		stellen = Stellen
	}

	werte := url.Values{}
	werte.Set("secret", AlsText(zaehler.Geheimnis))
	werte.Set("issuer", Aussteller)
	werte.Set("algorithm", string(verfahren))
	werte.Set("digits", strconv.Itoa(stellen))
	werte.Set("period", strconv.Itoa(int(Schritt.Seconds())))

	// **Der Pfad wird kodiert und nicht zusammengeklebt**: ein Konto darf ein
	// `/`, ein `?` und ein Leerzeichen tragen, und jedes davon zerbräche eine
	// gebaute Zeichenkette.
	adresse := url.URL{
		Scheme:   "otpauth",
		Host:     "totp",
		Path:     "/" + Aussteller + ":" + konto,
		RawQuery: werte.Encode(),
	}

	return adresse.String()
}

// bauer gibt die Hashfunktion zu einem Verfahren.
//
// **Der Rückfall ist SHA-1**, weil ein unbekanntes Verfahren aus einem
// gespeicherten Wert kommt und ein Absturz dort keine Antwort wäre (§8).
func bauer(verfahren Verfahren) func() hash.Hash {
	switch verfahren {
	case SHA256:
		return sha256.New
	case SHA512:
		return sha512.New
	}

	// **SHA-1 für alles andere, und das schließt den leeren Wert ein.** Ein
	// eigener Zweig für [SHA1] daneben wäre eine Zeile, die kein Lauf
	// betreten kann — dasselbe Ergebnis auf zwei Wegen (§3).
	return sha1.New
}

// teiler gibt die Zehnerpotenz zu einer Stellenzahl.
func teiler(stellen int) uint32 {
	ganz := uint32(1)
	for range stellen {
		ganz *= 10
	}

	return ganz
}

// istZiffer sagt, ob dieses Zeichen eine Ziffer ist — für `strings.TrimFunc`.
func istZiffer(zeichen rune) bool { return zeichen >= '0' && zeichen <= '9' }
