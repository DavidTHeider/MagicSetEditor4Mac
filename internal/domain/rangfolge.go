package domain

// **Wer im Streitfall recht behält** — Werkstück G4 von Phase 5.5
// (#1461).
//
// # Die Leiter, und warum sie so herum steht
//
// **Das Issue gibt sie wörtlich vor**: „unüberschreibbare Engine-Invarianten;
// versionierte Magic-Basis; Format; gesperrte Turnierparameter; zulässige
// Session-Hausregeln."
//
// **Von oben nach unten wird enger.** Ganz oben steht, was nie fällt; ganz
// unten, was zwei Menschen am Küchentisch miteinander ausmachen. Eine Stufe
// darf die darüber nicht antasten — und das ist der ganze Sinn: eine
// Hausregel, die eine Invariante aufhöbe, wäre keine Hausregel, sondern ein
// Loch.
//
// # Und der Streit wird vor „Bereit" entschieden
//
// **„Konflikte werden vor Bereit abgewiesen"** (#1461). Nicht während des
// Spiels und nicht beim ersten Zug, der daran scheitert: wer eine Partie
// aufsetzt, erfährt beim Aufsetzen, dass sie nicht geht.
//
// # Was ausdrücklich nicht dazugehört
//
// **Darstellungspräferenzen.** Das Issue nennt sie eigens: „Darstellungs-
// präferenzen stehen außerhalb der Fachregeln." Wie groß jemand seine Karten
// sieht, ist keine Regel und steht in keiner Stufe.

import (
	"fmt"
	"slices"
)

// Regelrang ist eine Stufe der Leiter — je kleiner, desto stärker.
type Regelrang int

const (
	// RangInvariante ist das, was kein Modus aufhebt.
	RangInvariante Regelrang = iota

	// RangBasis ist das versionierte Regelwerk des Spiels.
	RangBasis

	// RangFormat ist das Deckformat samt seinen Zahlen.
	RangFormat

	// RangTurnier sind die gesperrten Parameter einer Veranstaltung.
	RangTurnier

	// RangHausregel ist, was eine Runde für sich vereinbart.
	RangHausregel
)

// Regelraenge gibt alle fünf, von stark nach schwach.
func Regelraenge() []Regelrang {
	return []Regelrang{
		RangInvariante, RangBasis, RangFormat, RangTurnier, RangHausregel,
	}
}

// IstRegelrang sagt, ob ein Wert einer der fünf ist.
func IstRegelrang(rang Regelrang) bool {
	return slices.Contains(Regelraenge(), rang)
}

// Name gibt den Rang in Worten — für Meldungen, die ein Mensch liest.
func (r Regelrang) Name() string {
	switch r {
	case RangInvariante:
		return "eine Invariante der Engine"
	case RangBasis:
		return "das Regelwerk des Spiels"
	case RangFormat:
		return "das Format"
	case RangTurnier:
		return "ein Turnierparameter"
	case RangHausregel:
		return "eine Hausregel"
	}

	return "etwas Unbekanntes"
}

// Regelkonflikt ist ein Streit zwischen zwei Stufen.
type Regelkonflikt struct {
	// Staerker und Schwaecher sind die beiden Stufen.
	Staerker   Regelrang
	Schwaecher Regelrang

	// Gegenstand ist, worum es geht — `startleben`, `sichtbarkeit`.
	Gegenstand string
}

// Fehler gibt den Konflikt als Fehler.
//
// **Mit beiden Stufen im Satz.** „Das geht nicht" sagt niemandem, was zu tun
// ist; „eine Hausregel kann das Regelwerk des Spiels nicht ändern" schon.
func (k Regelkonflikt) Fehler() error {
	return fmt.Errorf(
		"%w: %s kann %s nicht ändern — es geht um %q",
		ErrValidation, k.Schwaecher.Name(), k.Staerker.Name(), k.Gegenstand)
}
