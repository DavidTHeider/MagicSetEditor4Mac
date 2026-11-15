package domain

import (
	"slices"
	"time"
)

// **Wie alt die letzte Sicherung ist, und wann der Plan wieder sichert** —
// OPS01 (#2347),
// Befund A-06 des Audits 0.9.0, Marke `new`: MSE2 sichert keinen Bestand,
// es schreibt Dateien.

// SicherungAltNach ist, ab wann eine Sicherung als alt gilt — das Issue sagt
// „über 14 Tage", und vierzehn Tage genau sind es noch nicht.
const SicherungAltNach = 14 * 24 * time.Hour

// sicherungsuhrspiel ist, wie weit ein Vermerk nach der Uhr dieses Rechners
// liegen darf. **Eine Minute und nicht null**: zwei Prozesse lesen dieselbe
// Uhr nicht im selben Augenblick, und ein Vermerk aus dem Nachbarprozess wäre
// sonst für einen Herzschlag „unstimmig".
const sicherungsuhrspiel = time.Minute

// Sicherungsstufe ordnet das Alter ein.
type Sicherungsstufe string

// Die vier Stufen.
const (
	// SicherungsstufeUnbekannt: kein Vermerk. **Unbekannt heißt unbekannt** —
	// ein Bestand, der nie über diese Fassung gesichert wurde, bekommt keinen
	// erfundenen Zeitpunkt, etwa den seiner Anlage.
	SicherungsstufeUnbekannt Sicherungsstufe = "unbekannt"

	// SicherungsstufeFrisch: höchstens [SicherungAltNach] her.
	SicherungsstufeFrisch Sicherungsstufe = "frisch"

	// SicherungsstufeAlt: länger her — die Erinnerung.
	SicherungsstufeAlt Sicherungsstufe = "alt"

	// SicherungsstufeUnstimmig: der Vermerk liegt nach der Uhr. Die Uhr wurde
	// zurückgestellt, oder der Bestand kam von einem Rechner mit falscher
	// Uhr; **ein Alter lässt sich dann nicht sagen**, und „frisch" wäre eine
	// Zusage, die niemand prüfen kann.
	SicherungsstufeUnstimmig Sicherungsstufe = "unstimmig"
)

// Sicherungsstand ist das Alter der letzten Sicherung.
type Sicherungsstand struct {
	Stufe Sicherungsstufe

	// Letzter ist der jüngste Vermerk — leer bei [SicherungsstufeUnbekannt].
	Letzter Sicherungsvermerk

	// Tage sind die vollen Tage seit [Sicherungsvermerk.Gesichert]; null,
	// wo sich kein Alter sagen lässt.
	Tage int
}

// SicherungsstandBei ordnet den letzten Vermerk zum Zeitpunkt `jetzt` ein.
func SicherungsstandBei(letzter Sicherungsvermerk, bekannt bool, jetzt time.Time) Sicherungsstand {
	if !bekannt {
		return Sicherungsstand{Stufe: SicherungsstufeUnbekannt}
	}

	alter := jetzt.Sub(letzter.Gesichert)

	switch {
	case alter < -sicherungsuhrspiel:
		return Sicherungsstand{Stufe: SicherungsstufeUnstimmig, Letzter: letzter}
	case alter > SicherungAltNach:
		return Sicherungsstand{Stufe: SicherungsstufeAlt, Letzter: letzter, Tage: int(alter / (24 * time.Hour))}
	default:
		return Sicherungsstand{Stufe: SicherungsstufeFrisch, Letzter: letzter, Tage: max(0, int(alter/(24*time.Hour)))}
	}
}

// SicherungsplanTakt ist der Abstand des Plans — **wöchentlich**, wie das
// Zielartboard `Betrieb09` es nennt.
const SicherungsplanTakt = 7 * 24 * time.Hour

// SicherungsplanBehalten ist, wie viele seiner eigenen Sicherungen der Plan
// stehen lässt.
const SicherungsplanBehalten = 7

// SicherungsplanFaellig sagt, ob der Plan sichern soll — gemessen an seinem
// **eigenen** letzten Vermerk und nicht an einer Sicherung von Hand: der Plan
// ist eine Reihe in einem Ordner, und eine Sicherung anderswohin ergänzt
// diese Reihe nicht.
//
// **Ein Vermerk nach der Uhr hält ihn nicht auf**: sonst sicherte ein Plan,
// dessen Uhr einmal um ein Jahr vorging, ein Jahr lang nicht.
func SicherungsplanFaellig(letzter Sicherungsvermerk, bekannt bool, jetzt time.Time) bool {
	if !bekannt {
		return true
	}

	alter := jetzt.Sub(letzter.Gesichert)

	return alter >= SicherungsplanTakt || alter < -sicherungsuhrspiel
}

// SicherungsplanAusduennen gibt die eigenen Sicherungen, die gehen dürfen —
// alle außer den [SicherungsplanBehalten] jüngsten, die jüngste zuerst.
//
// **„Eigene" entscheidet der Aufrufer**, und nur er kann es: eine Datei ist
// eigen, wenn ein Vermerk sie wiedererkennt — Name, Größe und Prüfsumme.
// Hier steht nur die Reihenfolge.
func SicherungsplanAusduennen(eigene []Sicherungsvermerk) []Sicherungsvermerk {
	if len(eigene) <= SicherungsplanBehalten {
		return nil
	}

	geordnet := slices.Clone(eigene)
	slices.SortStableFunc(geordnet, func(a, b Sicherungsvermerk) int {
		return b.Gesichert.Compare(a.Gesichert)
	})

	return geordnet[SicherungsplanBehalten:]
}
