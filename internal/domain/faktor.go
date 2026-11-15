// **Der zweite Faktor** — Werkstück 7 von Phase 4.12
// (#1351),
// nach Vermerk [0066](../../docs/entscheidungen/0066-nutzer-profile-und-die-vier-entscheidungen.md).
//
// # Zwei Arten, und eine davon geht überall
//
// **TOTP ist der Weg, der immer geht** — auf einem Mac ohne Touch ID, unter
// Windows und unter Linux (§12). **Biometrie ist der bequeme**, und sie hängt
// am Rechner: wer seinen Bestand auf einen anderen Mac zurückspielt, hat sie
// dort nicht.
//
// # Was hier **nicht** steht
//
// **Das Geheimnis.** Es liegt im Schlüsselbund, aus demselben Grund wie die
// Zugangsdaten fremder Dienste: die Bestandsdatei ist die, die man weitergibt,
// und wer das TOTP-Geheimnis hat, macht jeden Code. Hier steht nur, **dass**
// jemand diesen Faktor hat.

package domain

import "strings"

// Faktorart sagt, welcher Weg gemeint ist.
type Faktorart string

// Die zwei Arten.
const (
	// FaktorTOTP ist ein Code aus einer Authenticator-App (RFC 6238).
	FaktorTOTP Faktorart = "totp"

	// FaktorBiometrie ist Touch ID oder Face ID auf **diesem** Rechner.
	FaktorBiometrie Faktorart = "biometrie"
)

// Faktorarten gibt die beiden — für eine Oberfläche, die sie anbietet.
//
// **Eine Funktion und keine Veränderliche**: eine Liste, die jemand von außen
// ändern kann, ist keine Aufzählung.
func Faktorarten() []Faktorart { return []Faktorart{FaktorTOTP, FaktorBiometrie} }

// Hoechstbezeichnung ist, wie lang der Name eines Faktors sein darf.
const Hoechstbezeichnung = 100

// Faktor ist ein zweiter Faktor eines Nutzers.
type Faktor struct {
	ID       int64
	NutzerID int64

	// Art ist [FaktorTOTP] oder [FaktorBiometrie].
	Art Faktorart

	// Bezeichnung sagt, welches Gerät gemeint ist — „iPhone", „dieser Mac".
	//
	// **Freiwillig**: wer nur einen Faktor hat, braucht keinen Namen dafür.
	Bezeichnung string

	// LetzterSchritt ist der zuletzt angenommene TOTP-Schritt — beim
	// Anlegen der des Bestätigungscodes, damit Faktor und Verbrauch **ein**
	// Schreiben sind (SEC-09, Vermerk 0178). Null heißt: noch keiner.
	LetzterSchritt int64
}

// Validate prüft einen Faktor.
func (f Faktor) Validate() error {
	if f.NutzerID == 0 {
		return NewFieldError("nutzerID", SchluesselPflichtfeld,
			"Ein zweiter Faktor gehört zu einem Nutzer.")
	}

	// **Die Art wird gegen die Aufzählung geprüft und nicht gegen sich
	// selbst.** Der `CHECK` in der Wanderung ist die zweite Reihe: er greift,
	// wo niemand durch diesen Weg geht, und er nennt kein Feld.
	bekannt := false

	for _, art := range Faktorarten() {
		bekannt = bekannt || art == f.Art
	}

	if !bekannt {
		return NewFieldError("art", SchluesselValidierung,
			"%q ist keine Art eines zweiten Faktors.", string(f.Art))
	}

	if f.Bezeichnung != strings.TrimSpace(f.Bezeichnung) {
		return NewFieldError("bezeichnung", SchluesselValidierung,
			"Eine Bezeichnung trägt keine Leerzeichen am Rand; %q geht nicht.",
			f.Bezeichnung)
	}

	if len([]rune(f.Bezeichnung)) > Hoechstbezeichnung {
		return NewFieldError("bezeichnung", SchluesselValidierung,
			"Eine Bezeichnung hat höchstens %d Zeichen.", Hoechstbezeichnung)
	}

	return nil
}

// Wiederherstellungscode ist ein gehüteter Code samt seiner Kennung.
//
// # Warum er in der Domäne steht und nicht im Bestand
//
// **Weil `service` ihn zurückbekommt.** Ein Typ, den nur `storage` kennt,
// ließe sich in einer Naht in `service` nicht nennen — und die Schnittstelle
// gehört dorthin, wo sie gebraucht wird (§2).
//
// **Die gehütete Form reist mit**, weil der Prüfer sie braucht: er hält jeden
// freien Code gegen die Eingabe und verbraucht den, der passt. Der Klartext
// steht nirgends — er wird einmal gezeigt und nie wieder.
type Wiederherstellungscode struct {
	ID       int64
	Gehuetet string
}

// Wiederherstellungscodes ist, wie viele ein Nutzer beim Einrichten bekommt.
//
// **Zehn**, die Zahl, die GitHub, Google und 1Password ausgeben. Sie ist keine
// Vorgabe irgendeiner Norm — sie ist die Antwort darauf, wie oft jemand sein
// Gerät verliert, bevor er neue holt.
const Wiederherstellungscodes = 10

// FehlenderFaktor sagt, dass hier ein zweiter Faktor verlangt wird.
//
// # Warum ein eigener Fehler und keine Abweisung
//
// **Weil die Fläche etwas anderes tun muss als bei einem falschen Kennwort.**
// „Kennung oder Kennwort passen nicht zusammen" heißt: noch einmal tippen.
// Dies hier heißt: **frag nach dem Code** — und dafür muss die Fläche es
// unterscheiden können (§8).
func FehlenderFaktor(art Faktorart) error {
	return NewFieldError("faktor", SchluesselZweiterFaktor,
		"Dieser Nutzer verlangt einen zweiten Faktor.").
		MitWerten(map[string]string{"art": string(art)})
}
