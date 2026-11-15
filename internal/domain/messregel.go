// **Wie die Ablaufspur und das Leistungsprofil eingestellt sind** — OPS-05
// (#1722),
// Marke `new`: MSE2 misst sich nicht.
//
// # Aus, bis jemand sie einschaltet (AK-03)
//
// **Die Spur ist aus**, und wer sie einschaltet, bestimmt drei Grenzen: wie
// viele Rufe sie aufnimmt (Abtastung in Prozent), wie viele sie höchstens
// hält (ein Ring, der älteste geht) und wie lange (Minuten). Alle drei haben
// eine Obergrenze; ein unlesbarer Wert fällt auf die Vorgabe und nie auf
// „unbegrenzt". Dazu die Zahl der Leistungsprofile, die liegen bleiben.
//
// # Die Spur ab dem nächsten Start, die Profile sofort
//
// **Die Spur wie das Protokoll** (OPS-04): gelesen wird beim Start des
// Prozesses, der misst — Dienst und Werkzeugserver —, und jeder hält seine
// eigene. Die Zahl der Profile liest der Ruf, der eines schreibt.

package domain

import (
	"strconv"
	"time"
)

// Die Schlüssel der Messeinstellungen.
const (
	MessungSpur         = "messung.spur"
	MessungAbtastung    = "messung.abtastung"
	MessungSpurgrenze   = "messung.spurgrenze"
	MessungAufbewahrung = "messung.aufbewahrung"
	MessungProfile      = "messung.profile"
)

// Die Grenzen der Messeinstellungen.
const (
	abtastungVorgabe           = 100
	spurgrenzeKleinste         = 10
	spurgrenzeVorgabe          = 1000
	spurgrenzeHoechstens       = 10000
	spuraufbewahrungVorgabe    = 60
	spuraufbewahrungHoechstens = 1440
	profileVorgabe             = 3
	profileHoechstens          = 20
)

// Messschluessel gibt die Schlüssel in Anzeigereihenfolge.
func Messschluessel() []string {
	return []string{MessungSpur, MessungAbtastung, MessungSpurgrenze, MessungAufbewahrung, MessungProfile}
}

// messeinstellungen sind die Zeilen für die Reihe der Einstellungen.
func messeinstellungen() []einstellung {
	neu := func(schluessel string, werte []string, vorgabe string, pruefung func(string) error) einstellung {
		wirkung := WirkungNeustart
		if schluessel == MessungProfile {
			wirkung = WirkungSofort
		}

		return einstellung{
			Schluessel: schluessel, Wirkung: wirkung, Bereich: BereichAnwendung,
			Werte: werte, Vorgabe: vorgabe, Pruefung: pruefung,
		}
	}

	return []einstellung{
		neu(MessungSpur, []string{ProtokollAus, ProtokollAn}, ProtokollAus, nil),
		neu(MessungAbtastung, nil, strconv.Itoa(abtastungVorgabe), zahlZwischen(1, abtastungVorgabe)),
		neu(MessungSpurgrenze, nil, strconv.Itoa(spurgrenzeVorgabe),
			zahlZwischen(spurgrenzeKleinste, spurgrenzeHoechstens)),
		neu(MessungAufbewahrung, nil, strconv.Itoa(spuraufbewahrungVorgabe),
			zahlZwischen(1, spuraufbewahrungHoechstens)),
		neu(MessungProfile, nil, strconv.Itoa(profileVorgabe), zahlZwischen(1, profileHoechstens)),
	}
}

// Messregel ist, was aus den Einstellungen für die Messung folgt.
type Messregel struct {
	Spur bool

	// Abtastung ist der Anteil der Rufe in Prozent, den die Spur aufnimmt.
	Abtastung int

	// Spurgrenze ist die Zahl der Einträge, die die Spur höchstens hält.
	Spurgrenze int

	// Aufbewahrung ist, wie lange ein Eintrag bleibt.
	Aufbewahrung time.Duration

	// Profile ist die Zahl der Leistungsprofile, die liegen bleiben.
	Profile int
}

// MessregelAus macht aus den wirksamen Einstellungen eine Regel.
func MessregelAus(werte map[string]string) Messregel {
	zahl := func(schluessel string, kleinste, groesste, vorgabe int) int {
		gelesen, fehler := strconv.Atoi(werte[schluessel])
		if fehler != nil || gelesen < kleinste || gelesen > groesste {
			return vorgabe
		}

		return gelesen
	}

	return Messregel{
		Spur:       werte[MessungSpur] == ProtokollAn,
		Abtastung:  zahl(MessungAbtastung, 1, abtastungVorgabe, abtastungVorgabe),
		Spurgrenze: zahl(MessungSpurgrenze, spurgrenzeKleinste, spurgrenzeHoechstens, spurgrenzeVorgabe),
		Aufbewahrung: time.Duration(zahl(MessungAufbewahrung, 1, spuraufbewahrungHoechstens,
			spuraufbewahrungVorgabe)) * time.Minute,
		Profile: zahl(MessungProfile, 1, profileHoechstens, profileVorgabe),
	}
}
