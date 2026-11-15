// Package kennwort hütet Kennwörter und erkennt sie wieder.
//
// **Werkstück 6 von Phase 4.12**
// (#1351).
//
// # Ein Blatt (§2)
//
// Es kennt `internal/domain` für die Fehlerklassen und sonst nichts aus diesem
// Projekt. Was ein Nutzer ist, weiß es nicht.
//
// # Argon2id, und die Zahlen sind nachgelesen
//
// **OWASP Password Storage Cheat Sheet, gelesen am 2026-09-05**: „Use Argon2id
// with a minimum configuration of 19 MiB of memory, an iteration count of 2,
// and 1 degree of parallelism." Das ist das **Mindestmaß**, und es ist
// genommen — nicht mehr, weil dieses Programm auf dem Rechner eines Menschen
// läuft, der gerade wartet, und nicht weniger, weil weniger unter der
// Empfehlung läge.
//
// # Kein Pepper, und das ist eine Entscheidung
//
// Dieselbe Quelle nennt ihn: ein Pepper liegt **nicht** neben dem Hash,
// sondern getrennt — hier wäre das der Schlüsselbund. Er ist trotzdem nicht
// gebaut, und der Grund gehört danebengeschrieben: **er sperrt den Eigentümer
// aus**. Wer seinen Bestand sichert und auf einem neuen Mac zurückspielt, hat
// den Schlüsselbund des alten nicht — und sein eigenes Kennwort wäre wertlos.
//
// **Gegen einen mitgelesenen Bestand hilft die Verschlüsselung** (Werkstück
// 14, Vermerk 0066), und die ist die Antwort auf diese Frage. Ein Pepper wäre
// die halbe.
package kennwort

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// ErrStimmtNicht heißt: die Form ist in Ordnung, das Kennwort nicht.
//
// **Ein eigener Fehler und keine Validierung.** „Stimmt nicht" heißt noch
// einmal tippen; „kaputte Form" heißt, im Bestand ist etwas faul (§8).
var ErrStimmtNicht = errors.New("das Kennwort stimmt nicht")

// Die Zahlen aus dem OWASP-Mindestmaß.
const (
	// Speicher ist `m` in Kibibyte — 19 MiB.
	Speicher uint32 = 19456

	// Durchgaenge ist `t`.
	Durchgaenge uint32 = 2

	// Faeden ist `p`.
	Faeden uint8 = 1

	// Laenge ist, wie viele Bytes der Hash hat.
	Laenge uint32 = 32

	// Salzlaenge ist, wie viele Bytes der Salzwert hat.
	Salzlaenge = 16
)

// Mindestlaenge ist, wie kurz ein Kennwort sein darf.
//
// **Zwölf Zeichen und keine Zeichenklassen.** Eine Regel „Groß, klein, Ziffer,
// Sonderzeichen" erzeugt `Passwort1!` und verbietet eine Wortfolge, die um
// Größenordnungen besser ist. Länge ist das Maß, das trägt.
const Mindestlaenge = 12

// Hoechstlaenge ist, wie lang eines sein darf.
//
// **Nicht aus Prinzip, sondern weil Hüten Arbeit kostet**: Argon2id rechnet
// über die ganze Eingabe, und ein Kennwort von einem Megabyte wäre ein Weg,
// diesen Rechner zu beschäftigen.
const Hoechstlaenge = 1024

// Huete gibt die Form, die im Bestand liegt.
func Huete(klartext string) (string, error) {
	if len(klartext) < Mindestlaenge {
		return "", fmt.Errorf("%w: ein Kennwort hat mindestens %d Zeichen",
			domain.ErrValidation, Mindestlaenge)
	}

	if len(klartext) > Hoechstlaenge {
		return "", fmt.Errorf("%w: ein Kennwort hat höchstens %d Zeichen",
			domain.ErrValidation, Hoechstlaenge)
	}

	salz := make([]byte, Salzlaenge)

	// **Ohne Fehlerprüfung, und das ist nachgelesen**: `crypto/rand.Read`
	// „never returns an error, and always fills b entirely" — es bricht das
	// Programm ab, wenn das Betriebssystem versagt. Ein `if err != nil`
	// daneben wäre ein Zweig, den kein Lauf betreten kann (§3).
	_, _ = rand.Read(salz) //nolint:errcheck // siehe oben

	hash := argon2.IDKey([]byte(klartext), salz, Durchgaenge, Speicher, Faeden, Laenge)

	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, Speicher, Durchgaenge, Faeden,
		base64.RawStdEncoding.EncodeToString(salz),
		base64.RawStdEncoding.EncodeToString(hash)), nil
}

// Stimmt sagt, ob dieses Kennwort zu dieser Form gehört.
func Stimmt(gehuetet, klartext string) error {
	salz, hash, zahlen, fehler := zerlege(gehuetet)
	if fehler != nil {
		return fehler
	}

	// **`uint32` ohne Sorge**: [langGenug] hat die Länge auf höchstens
	// [Laengstermischwert] beschränkt, bevor dieser Wert hierher kam.
	gerechnet := argon2.IDKey([]byte(klartext), salz,
		zahlen.durchgaenge, zahlen.speicher, zahlen.faeden,
		uint32(len(hash))) //#nosec G115 -- durch langGenug auf 1024 begrenzt

	// **Ein Vergleich, der immer gleich lange dauert.** Ein `bytes.Equal`
	// bricht beim ersten Unterschied ab, und wer die Zeit misst, liest den
	// Hash Byte für Byte ab.
	if subtle.ConstantTimeCompare(gerechnet, hash) != 1 {
		return ErrStimmtNicht
	}

	return nil
}

// Die Schranken, innerhalb derer eine gehütete Form gelesen wird.
//
// # Warum es sie gibt — und warum es keine Formprüfung ist
//
// **Weil `argon2.IDKey` tut, was dasteht.** Eine Zeile mit `m=99999999`
// verlangt hundert Gigabyte, und `t=0` oder `p=0` bringt die Bibliothek zum
// **Absturz** — nachgestellt am 2026-09-05 in
// `TestUnmoeglicheZahlenWerdenAbgewiesen`. Das ist kein Angriff von außen: es
// genügt eine Bestandsdatei, die jemand von Hand geändert hat oder die von
// einem anderen Werkzeug stammt (§8: fremde Daten sind der Normalfall).
//
// **Und die Hashlänge geht in eine `uint32`** — angestrichen von gosec
// (G115). Die Antwort darauf ist eine Schranke und keine Ausnahme im Wächter:
// ein Wert, den dieses Programm nie geschrieben hat, wird abgewiesen, statt
// stillschweigend umzulaufen.
//
// **Großzügig und nicht knapp**: sie sind ein Riegel gegen das Unmögliche,
// keine Zusage über die Güte. [Speicher] und [Durchgaenge] dürfen wachsen,
// ohne dass alte Kennwörter wertlos werden — das ist der Grund, warum die
// Form ihre Zahlen überhaupt mitträgt.
const (
	// Hoechstspeicher ist ein Gibibyte in Kibibyte.
	Hoechstspeicher = 1 << 20

	// Hoechstdurchgaenge ist die Obergrenze für `t`.
	Hoechstdurchgaenge = 64

	// Kuerzestersalzwert und Kuerzesterhash weisen ab, was zu wenig ist, um
	// je von diesem Programm zu stammen.
	Kuerzestersalzwert = 8
	Kuerzesterhash     = 16

	// Laengstermischwert ist die Obergrenze für Salzwert und Hash.
	Laengstermischwert = 1024
)

// zahlenwerte sind die Parameter, mit denen ein Hash entstand.
type zahlenwerte struct {
	speicher    uint32
	durchgaenge uint32
	faeden      uint8
}

// tragbar sagt, ob mit diesen Zahlen gerechnet werden darf.
func (z zahlenwerte) tragbar() bool {
	return z.speicher > 0 && z.speicher <= Hoechstspeicher &&
		z.durchgaenge > 0 && z.durchgaenge <= Hoechstdurchgaenge &&
		z.faeden > 0
}

// zerlege liest die gehütete Form — Salzwert, Hash und die Zahlen dazu.
//
// **Sie trägt ihre eigenen Zahlen**, und darum lässt sich das Mindestmaß
// später anheben, ohne die alten Kennwörter wertlos zu machen: gerechnet wird
// mit dem, was dasteht.
func zerlege(gehuetet string) (
	salz, hash []byte, zahlen zahlenwerte, fehler error,
) {
	teile := strings.Split(gehuetet, "$")

	// **Sechs Teile**: der leere vor dem ersten `$`, dann Verfahren, Fassung,
	// Zahlen, Salzwert, Hash.
	if len(teile) != 6 || teile[1] != "argon2id" {
		return nil, nil, zahlenwerte{}, fmt.Errorf(
			"%w: die gehütete Form ist keine von argon2id", domain.ErrValidation)
	}

	var fassung int

	if _, err := fmt.Sscanf(teile[2], "v=%d", &fassung); err != nil ||
		fassung != argon2.Version {
		return nil, nil, zahlenwerte{}, fmt.Errorf(
			"%w: die gehütete Form nennt eine andere Fassung als %d",
			domain.ErrValidation, argon2.Version)
	}

	if _, err := fmt.Sscanf(teile[3], "m=%d,t=%d,p=%d",
		&zahlen.speicher, &zahlen.durchgaenge, &zahlen.faeden); err != nil ||
		!zahlen.tragbar() {
		return nil, nil, zahlenwerte{}, fmt.Errorf(
			"%w: die gehütete Form nennt keine brauchbaren Zahlen", domain.ErrValidation)
	}

	salz, salzfehler := base64.RawStdEncoding.DecodeString(teile[4])
	if salzfehler != nil || !langGenug(salz, Kuerzestersalzwert) {
		return nil, nil, zahlenwerte{}, fmt.Errorf(
			"%w: der Salzwert ist keiner von hier", domain.ErrValidation)
	}

	hash, hashfehler := base64.RawStdEncoding.DecodeString(teile[5])
	if hashfehler != nil || !langGenug(hash, Kuerzesterhash) {
		return nil, nil, zahlenwerte{}, fmt.Errorf(
			"%w: der Hash ist keiner von hier", domain.ErrValidation)
	}

	return salz, hash, zahlen, nil
}

// langGenug sagt, ob dieser Wert in den Schranken liegt.
//
// **Die Obergrenze ist die, die `uint32(len(…))` sicher macht** — und sie
// steht hier und nicht an der Rechnung, damit ein Wert, der sie reißt, gar
// nicht erst dorthin kommt.
func langGenug(wert []byte, mindestens int) bool {
	return len(wert) >= mindestens && len(wert) <= Laengstermischwert
}
