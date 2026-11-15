// **Der zweite Faktor: einrichten und entfernen** — Werkstück 7 von Phase
// 4.12 (#1351),
// nach Vermerk [0066](../../docs/entscheidungen/0066-nutzer-profile-und-die-vier-entscheidungen.md).
//
// # Was hier entschieden wird
//
// **Dass ein Faktor erst gilt, wenn er bewiesen ist.** Wer ihn beim Anzeigen
// des Musters einträgt und den Scan vergisst, verlangt danach einen Code, den
// niemand erzeugen kann — er hätte sich selbst ausgesperrt.
//
// **Und dass der Dienst den Sensor selbst fragt.** Eine Fläche, die „ja, war
// Touch ID" meldet, weist nichts nach; der Sensor hängt an einer Naht, und der
// Composition Root setzt sie (§2).
//
// # Was hier **nicht** steht
//
// **Kein RFC 6238 und kein Argon2id.** Beides steht in `internal/totp` und
// `internal/kennwort`; dieser Dienst ruft sie und rechnet nicht selbst.
//
// **Und nicht, wo ein Geheimnis liegt.** Das sagt der Composition Root über
// [Geheimnisse] — im Betrieb der Schlüsselbund, im Lauf ein Feld.

package service

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/kennwort"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/totp"
)

// Faktorablage ist die Naht zum Bestand.
type Faktorablage interface {
	Faktoren(ctx context.Context, nutzerID int64) ([]domain.Faktor, error)
	FaktorAnlegen(ctx context.Context, neu domain.Faktor) (domain.Faktor, error)
	FaktorLoeschen(ctx context.Context, nutzerID int64, art domain.Faktorart) error

	CodesSetzen(ctx context.Context, nutzerID int64, gehuetete []string) error
	FreieCodes(ctx context.Context, nutzerID int64) (
		[]domain.Wiederherstellungscode, error)
	CodeVerbrauchen(ctx context.Context, codeID int64) error

	// **Der Haushalt** (SEC-08 bis SEC-10, #2149–#2151, Vermerk 0178):
	// der Schritt, der schon galt, und das Budget der Fehlversuche — beide
	// im Bestand, denn vier Prozesse teilen ihn.
	SchrittVerbrauchen(ctx context.Context, nutzerID, schritt int64) (bool, error)
	FaktorversuchBeginnen(ctx context.Context, nutzerID int64, jetzt time.Time) (time.Time, error)
	FaktorversuchGelungen(ctx context.Context, nutzerID int64) error
}

// Geheimnisse ist die Naht dorthin, wo ein TOTP-Geheimnis liegt.
//
// **Der Dienst sagt „setze", nicht „schreib in den Schlüsselbund".** Wo es
// liegt, ist eine Frage der Plattform und gehört in den Composition Root —
// dieselbe Richtung wie bei den Zugangsdaten fremder Dienste (§2, §16).
type Geheimnisse interface {
	Setze(ctx context.Context, nutzerID int64, geheimnis string) error
	Hole(ctx context.Context, nutzerID int64) (string, error)
	Nimm(ctx context.Context, nutzerID int64) error
}

// Sensor ist die Naht zur Biometrie dieses Rechners.
type Sensor interface {
	// Vorhanden sagt, ob eine Abfrage überhaupt versucht werden kann —
	// **ohne etwas anzustoßen**.
	Vorhanden(ctx context.Context) bool

	// Bestaetige zeigt den Dialog und wartet auf den Sensor.
	Bestaetige(ctx context.Context, grund string) error
}

// Zweiterfaktor richtet zweite Faktoren ein und prüft sie.
type Zweiterfaktor struct {
	ablage Faktorablage
	bund   Geheimnisse
	sensor Sensor

	// uhr sagt, wie spät es ist — für das Prüffenster von TOTP.
	uhr func() time.Time

	// einstellungen sagt, was dieser Mensch für sich eingestellt hat —
	// Werkstück 8 (#1351). `nil` heißt: es gilt die Vorgabe.
	einstellungen *Nutzereinstellungen

	// huete hütet einen Wiederherstellungscode.
	//
	// **Ein Feld und kein fester Ruf**, aus demselben Grund wie [uhr]: der
	// Fehlerweg von [kennwort.Huete] ist von hier aus nicht erreichbar — ein
	// gruppierter Code hat vierzehn Zeichen und liegt damit sicher zwischen
	// `Mindestlaenge` und `Hoechstlaenge`. Ohne die Naht bliebe der Zweig
	// unbetreten und wäre eine Behauptung darüber, was im Ernstfall geschieht
	// (§3).
	huete func(string) (string, error)
}

// NeuerZweiterfaktor baut ihn über seinen drei Nähten.
func NeuerZweiterfaktor(
	ablage Faktorablage, bund Geheimnisse, sensor Sensor,
) *Zweiterfaktor {
	return &Zweiterfaktor{
		ablage: ablage, bund: bund, sensor: sensor,
		uhr: time.Now, huete: kennwort.Huete,
	}
}

// MitEinstellungen hängt die Einstellungen des Nutzers an.
//
// **Als eigener Ruf**, aus demselben Grund wie [Nutzerverwaltung.MitFaktor]:
// beide hängen an derselben Ablage, und ein gemeinsamer Aufruf schriebe die
// Reihenfolge im Composition Root fest.
func (z *Zweiterfaktor) MitEinstellungen(einstellungen *Nutzereinstellungen) *Zweiterfaktor {
	z.einstellungen = einstellungen

	return z
}

// MitUhr setzt die Uhr — für Läufe, die einen festen Zeitpunkt brauchen.
func (z *Zweiterfaktor) MitUhr(uhr func() time.Time) *Zweiterfaktor {
	z.uhr = uhr

	return z
}

// Einrichtung ist, was jemand **einmal** zu sehen bekommt.
//
// **Einmal, und das ist die ganze Aussage**: das Geheimnis und die Codes
// stehen danach nirgends im Klartext. Wer sie wegklickt, richtet neu ein.
type Einrichtung struct {
	// Adresse ist die `otpauth://`-Adresse für den Scanner.
	Adresse string

	// Geheimnis ist dieselbe Angabe zum Abtippen — für den, der keine Kamera
	// hat oder keinen QR-Code will.
	Geheimnis string

	// Codes sind die Wiederherstellungscodes im Klartext.
	Codes []string
}

// Faktoren gibt die zweiten Faktoren eines Nutzers.
func (z *Zweiterfaktor) Faktoren(ctx context.Context, nutzerID int64) (
	[]domain.Faktor, error,
) {
	return z.ablage.Faktoren(ctx, nutzerID)
}

// TOTPEinrichten zieht ein Geheimnis und gibt zurück, was einmal zu sehen ist.
//
// **Der Faktor entsteht dabei noch nicht** — siehe [Zweiterfaktor.TOTPBestaetigen].
func (z *Zweiterfaktor) TOTPEinrichten(ctx context.Context, nutzerID int64) (
	Einrichtung, error,
) {
	geheim := totp.NeuesGeheimnis()

	if fehler := z.bund.Setze(ctx, nutzerID, totp.AlsText(geheim)); fehler != nil {
		return Einrichtung{}, fehler
	}

	codes, gehuetete, fehler := neueCodes(z.huete)
	if fehler != nil {
		return Einrichtung{}, fehler
	}

	if setzfehler := z.ablage.CodesSetzen(ctx, nutzerID, gehuetete); setzfehler != nil {
		return Einrichtung{}, setzfehler
	}

	return Einrichtung{
		Adresse:   totp.Adresse(totp.Zaehler{Geheimnis: geheim}, kontoname(nutzerID)),
		Geheimnis: totp.AlsText(geheim),
		Codes:     codes,
	}, nil
}

// TOTPBestaetigen trägt den Faktor ein, wenn ein Code aus ihm stimmt.
//
// **Im Budget der Fehlversuche**, wie jeder Codeweg ([imBudget], Zweitblick
// über F, Befund F2) — und **in einem Schreiben mit seinem Schritt**
// (Befund F1): getrennt angelegt und verbraucht, galt der Code der
// Bestätigung dazwischen, und nach einem Fehler stand der Faktor ohne
// verbrauchten Schritt da. Auch dieser Code gilt nur einmal (RFC 6238 §5.2).
func (z *Zweiterfaktor) TOTPBestaetigen(
	ctx context.Context, nutzerID int64, code string,
) error {
	_, fehler := imBudget(ctx, z, nutzerID, func() (struct{}, error) {
		zaehler, fehler := z.zaehlerFuer(ctx, nutzerID)
		if fehler != nil {
			return struct{}{}, fehler
		}

		schritt, stimmtnicht := totp.PassenderSchritt(zaehler, code, z.uhr())
		if stimmtnicht != nil {
			return struct{}{}, stimmtnicht
		}

		_, anlegefehler := z.ablage.FaktorAnlegen(ctx, domain.Faktor{
			NutzerID: nutzerID, Art: domain.FaktorTOTP, LetzterSchritt: schritt,
		})

		return struct{}{}, anlegefehler
	})

	return fehler
}

// BiometrieEinrichten trägt den biometrischen Faktor ein.
//
// **Einmal den Finger, bevor er gilt** — dieselbe Überlegung wie beim
// zweistufigen TOTP: sonst trüge jemand einen Faktor ein, den er nicht
// bedienen kann.
func (z *Zweiterfaktor) BiometrieEinrichten(ctx context.Context, nutzerID int64) error {
	if !z.sensor.Vorhanden(ctx) {
		return domain.NewFieldError("art", domain.SchluesselValidierung,
			"Dieser Rechner hat keine Biometrie, die sich einrichten ließe.")
	}

	if fehler := z.sensor.Bestaetige(ctx, grundEinrichten); fehler != nil {
		return fehler
	}

	_, anlegefehler := z.ablage.FaktorAnlegen(ctx, domain.Faktor{
		NutzerID: nutzerID, Art: domain.FaktorBiometrie,
	})

	return anlegefehler
}

// Entfernen nimmt einen zweiten Faktor heraus — samt seinem Geheimnis.
func (z *Zweiterfaktor) Entfernen(
	ctx context.Context, nutzerID int64, art domain.Faktorart,
) error {
	if fehler := z.ablage.FaktorLoeschen(ctx, nutzerID, art); fehler != nil {
		return fehler
	}

	// **Nur TOTP hat eines.** Biometrie hängt am Rechner und nicht an einem
	// Wert, den jemand aufbewahrt.
	if art != domain.FaktorTOTP {
		return nil
	}

	return z.bund.Nimm(ctx, nutzerID)
}

// zaehlerFuer holt das Geheimnis eines Nutzers und baut den Zähler.
func (z *Zweiterfaktor) zaehlerFuer(ctx context.Context, nutzerID int64) (
	totp.Zaehler, error,
) {
	text, fehler := z.bund.Hole(ctx, nutzerID)
	if fehler != nil {
		return totp.Zaehler{}, fehler
	}

	geheim, lesefehler := totp.GeheimnisAus(text)
	if lesefehler != nil {
		return totp.Zaehler{}, lesefehler
	}

	return totp.Zaehler{Geheimnis: geheim}, nil
}

// Die Sätze, die im Dialog des Sensors stehen.
//
// **Sie stehen hier und nicht im Sprachkatalog**, und das ist eine offene
// Stelle: macOS zeigt sie, nicht das Fenster, und die Fassade reicht keine
// Übersetzung bis hierher. Deutsch ist die Ausgangssprache (§9).
const (
	grundEinrichten = "um Touch ID als zweiten Faktor einzurichten"
	grundPruefen    = "um dich als Eigentümer dieses Kontos auszuweisen"
)

// Codelaenge ist, wie viele Zeichen ein Wiederherstellungscode trägt.
//
// **Zwölf aus base32, in drei Gruppen zu vier** — sechzig Bit. Das ist keine
// Zusage über Rechenaufwand, sondern die Antwort darauf, dass er sich
// abschreiben lassen muss: eine Zeile, die niemand tippt, ist kein Rückweg.
const Codelaenge = 12

// neueCodes zieht die Wiederherstellungscodes samt ihrer gehüteten Form.
func neueCodes(huete func(string) (string, error)) (
	klartext, gehuetete []string, fehler error,
) {
	for range domain.Wiederherstellungscodes {
		roh := make([]byte, Codelaenge)
		rand.Read(roh) //nolint:errcheck // gibt seit Go 1.24 keinen Fehler

		code := gruppiert(strings.ToLower(
			base32.StdEncoding.WithPadding(base32.NoPadding).
				EncodeToString(roh)[:Codelaenge]))

		gehuetet, huetefehler := huete(code)
		if huetefehler != nil {
			return nil, nil, huetefehler
		}

		klartext = append(klartext, code)
		gehuetete = append(gehuetete, gehuetet)
	}

	return klartext, gehuetete, nil
}

// gruppiert setzt Bindestriche nach je vier Zeichen.
func gruppiert(code string) string {
	teile := make([]string, 0, len(code)/4)

	for ab := 0; ab < len(code); ab += 4 {
		teile = append(teile, code[ab:min(ab+4, len(code))])
	}

	return strings.Join(teile, "-")
}

// kontoname ist der Name, unter dem der Faktor in der App erscheint.
//
// **Die Kennung und nicht der Anzeigename**: der ändert sich, und eine
// Authenticator-App zeigt danach etwas anderes als der Bestand.
func kontoname(nutzerID int64) string { return fmt.Sprintf("Nutzer %d", nutzerID) }

// istNichtGefunden hält die Unterscheidung an einer Stelle.
func istNichtGefunden(fehler error) bool { return errors.Is(fehler, domain.ErrNotFound) }
