package domain

// **Ein bekanntes Gegenüber und was es darf** — Werkstück N1 von Phase 5.4
// (#1454).
//
// # Warum Kontakt, Profil und Nutzer drei Begriffe sind
//
// **Weil sie drei verschiedene Fragen beantworten.** Ein Nutzer meldet sich an
// diesem Rechner an, ein Profil ist der Arbeitsstand, unter dem er arbeitet,
// und ein Kontakt ist jemand **anderes** — auf einem anderen Rechner, mit
// einem eigenen Schlüssel. Das Issue trennt sie ausdrücklich; eine gemeinsame
// Kennung machte aus zweien einen.

import (
	"fmt"
	"slices"
	"time"
)

// Kontaktstand sagt, wie weit ein Gegenüber geprüft ist.
type Kontaktstand string

// Die drei Stände.
const (
	// KontaktUnbestaetigt heißt: wir haben uns gesehen, mehr nicht.
	//
	// **Das ist der Ausgangszustand jeder Begegnung**, und er entsteht von
	// selbst. Er gestattet nichts außer der Paarung.
	KontaktUnbestaetigt Kontaktstand = "unbestaetigt"

	// KontaktBestaetigt heißt: ein Mensch hat den Fingerabdruck verglichen.
	//
	// **Ein Mensch, nicht das Programm.** Das Issue verlangt es: „Gegenüber
	// per Fingerprint/Bestätigung prüfen; Anzeigenamen allein sind kein
	// Identitätsnachweis."
	KontaktBestaetigt Kontaktstand = "bestaetigt"

	// KontaktGesperrt heißt: von hier kommt nichts mehr durch.
	KontaktGesperrt Kontaktstand = "gesperrt"
)

// DieKontaktstaende ist die Aufzählung in fester Reihenfolge.
var DieKontaktstaende = []Kontaktstand{
	KontaktUnbestaetigt, KontaktBestaetigt, KontaktGesperrt,
}

// Peerkontakt ist ein Gegenüber, das dieses Profil kennt.
type Peerkontakt struct {
	// Kennung ist der Nachweis — sie **ist** der öffentliche Schlüssel.
	Kennung string

	// Fingerabdruck ist die Lesehilfe für den Vergleich durch Menschen.
	Fingerabdruck string

	// SelbstgewaehlterName ist, wie das Gegenüber sich nennt — **kein
	// Nachweis**.
	SelbstgewaehlterName string

	// EigenerName ist, wie dieses Profil ihn nennt.
	//
	// **Er steht hier, damit der selbstgewählte nicht gebraucht wird.** Zwei
	// gleichnamige Peers sind sonst nicht zu unterscheiden — die Abnahme des
	// Issues verlangt genau das.
	EigenerName string

	Stand   Kontaktstand
	Gesehen time.Time

	// Erstmals ist die **erste** Begegnung — der Nullwert heißt: nicht
	// bekannt (T3, #1468).
	//
	// **Eine andere Frage als [Peerkontakt.Gesehen]**: „seit wann kenne ich
	// den?" gegen „wann zuletzt?". Für einen Kontakt, der vor `T3` entstanden
	// ist, weiß dieses Programm sie nicht — und ein geratenes Datum sähe aus
	// wie ein gemessenes (§8).
	Erstmals time.Time

	// Notiz ist, was dieses Profil sich zu dem Gegenüber notiert hat.
	//
	// **Lokal und nur lokal.** Sie geht über keine Leitung: eine Sichtkarte
	// trägt sie nicht, und eine Freigabe erst recht nicht.
	Notiz string

	// Adresse ist, wo dieses Gegenüber erreichbar ist — **von Hand
	// eingetragen und leer, solange niemand eine kennt**.
	//
	// **Sie ist kein Nachweis und keine Auffindbarkeit**, sondern beides
	// nicht: wer an ihr sitzt, sagt erst der Handschlag, und ob dort jemand
	// sitzt, sagt erst ein Versuch. Der Grund steht bei [PruefeAdresse].
	Adresse string

	// Rechte sind die Zwecke, die dieses Profil freigegeben hat.
	Rechte []Zweck

	// Vorgaenger ist die Kennung eines Kontakts, dessen neuer Schlüssel
	// dieser sein will — leer, wenn niemand das behauptet (NET-01, #1705).
	//
	// **Eine Behauptung, kein Erbe**: Rechte, Stand und Name des Vorgängers
	// gehen nicht über. Sie steht hier, damit „Schlüssel geändert" ein
	// eigener Zustand ist ([Peerkontakt.Lage]).
	Vorgaenger string
}

// Darf sagt, ob dieser Kontakt einen Zweck öffnen darf.
//
// # Die eine Regel, um die es geht
//
// **Eine Paarung darf jeder, der nicht gesperrt ist — und sonst nichts.** Ohne
// den ersten Halbsatz gäbe es keinen Weg, überhaupt Kontakt aufzunehmen: die
// Paarung **ist** das Kennenlernen. Ohne den zweiten wäre sie eine
// Generalvollmacht, und das Issue sagt ausdrücklich: „Pairing allein gestattet
// keinen Import oder Handel."
func (k Peerkontakt) Darf(zweck Zweck) bool {
	if k.Stand == KontaktGesperrt {
		return false
	}

	if zweck == ZweckPaarung {
		return true
	}

	return k.Stand == KontaktBestaetigt && slices.Contains(k.Rechte, zweck)
}

// Validate hält den Kontakt gegen das, was einer sein kann.
func (k Peerkontakt) Validate() error {
	if k.Kennung == "" {
		return fmt.Errorf(
			"%w: ein Kontakt ohne Kennung ist keiner", ErrValidation)
	}

	if !slices.Contains(DieKontaktstaende, k.Stand) {
		return fmt.Errorf(
			"%w: %q ist kein Stand eines Kontakts", ErrValidation, k.Stand)
	}

	if fehler := PruefeAdresse(k.Adresse); fehler != nil {
		return fehler
	}

	for _, zweck := range k.Rechte {
		if fehler := PruefeZweck(zweck); fehler != nil {
			return fehler
		}

		// **Ein Recht an einem ungeprüften Gegenüber gibt es nicht.**
		// Sonst hinge die Freigabe an einem Namen, den sich jeder geben
		// kann — und der Vergleich des Fingerabdrucks wäre Zierde.
		if k.Stand != KontaktBestaetigt {
			return fmt.Errorf(
				"%w: %q setzt ein bestätigtes Gegenüber voraus",
				ErrValidation, zweck)
		}
	}

	return nil
}
