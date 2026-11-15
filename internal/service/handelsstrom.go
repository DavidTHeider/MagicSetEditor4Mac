package service

// **Was hinausgeht und was nachgeliefert wird** — Werkstück `T5` von Phase
// 5.6 (#1470).
//
// # Warum das Senden über einen Ausgang geht
//
// **Weil „abgeschickt" nicht „angekommen" heißt.** Eine Nachricht, die nur
// über die Leitung ginge, wäre nach einem Abriss weg — und niemand wüsste,
// ob die andere Seite sie hat. Sie liegt deshalb erst im Ausgang und geht
// dann hinaus; erst die Bestätigung räumt sie weg.
//
// # Und warum die Post ohne Ausgang nicht sendet
//
// **Weil ein stilles Nichts das Schlimmste wäre.** Eine Post ohne Ausgang
// kann empfangen — das braucht ihn nicht —, aber senden kann sie nicht, und
// das sagt sie, statt die Nachricht verschwinden zu lassen.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Ausgangsnachricht ist eine Nachricht, die noch nicht bestätigt ist.
type Ausgangsnachricht struct {
	NachrichtID string
	Vorgang     string
	Empfaenger  string

	// Folge ist die Sequenz dieser Nachricht — sie entscheidet, was bei
	// einer Wiederaufnahme nachzuliefern ist.
	Folge int64

	// Nachricht ist, was hinausgeht.
	Nachricht Handelsnachricht
}

// Ausgang ist, was die Post zum Senden braucht.
type Ausgang interface {
	Einreihen(ctx context.Context, zeile Ausgangsnachricht) error
	Offene(ctx context.Context, empfaenger string) ([]Ausgangsnachricht, error)
	Bestaetigen(ctx context.Context, nachrichtID string) error
}

// MitAusgang hängt den Ausgang an — **wahlfrei**, und ohne ihn wird nur
// empfangen.
func (p *Handelspost) MitAusgang(ausgang Ausgang) *Handelspost {
	p.ausgang = ausgang

	return p
}

// Senden legt eine Nachricht in den Ausgang.
//
// **Zweimal dasselbe ergibt eine Zeile** — der Ausgang ist idempotent, und
// wer nach einem Neustart noch einmal sendet, schickt nicht doppelt.
func (p *Handelspost) Senden(
	ctx context.Context, empfaenger string, nachricht Handelsnachricht,
) error {
	if p.ausgang == nil {
		return domain.MitGrund(domain.ErrValidation, nil,
			"Diese Post kann nur empfangen und nichts senden.")
	}

	return p.ausgang.Einreihen(ctx, Ausgangsnachricht{
		NachrichtID: nachricht.Umschlag.NachrichtID,
		Vorgang:     nachricht.Umschlag.Vorgang,
		Empfaenger:  empfaenger,
		Folge:       nachricht.Umschlag.Folge,
		Nachricht:   nachricht,
	})
}

// Bestaetigt räumt eine Nachricht weg, die angekommen ist.
func (p *Handelspost) Bestaetigt(ctx context.Context, nachrichtID string) error {
	if p.ausgang == nil {
		return domain.MitGrund(domain.ErrValidation, nil,
			"Diese Post kann nur empfangen und nichts senden.")
	}

	return p.ausgang.Bestaetigen(ctx, nachrichtID)
}

// Wiederaufnehmen sagt, was nach einem Abriss nachzuliefern ist.
//
// **Drei Rückgaben, und die mittlere ist kein Fehler.** „Es klemmt" heißt:
// beide Seiten haben ordentlich gearbeitet, und trotzdem stehen zwei
// Verläufe da. Wer daraus einen `error` machte, ließe jemanden die Klärung
// wegprüfen (siehe [domain.Nachzuliefern]).
func (p *Handelspost) Wiederaufnehmen(
	ctx context.Context, empfaenger string, meiner, seiner domain.Handelsfortschritt,
) (nachzuliefern []Ausgangsnachricht, klaerung bool, fehler error) {
	if p.ausgang == nil {
		return nil, false, domain.MitGrund(domain.ErrValidation, nil,
			"Diese Post kann nur empfangen und nichts senden.")
	}

	abFolge, klemmt := domain.Nachzuliefern(meiner, seiner)
	if klemmt {
		return nil, true, nil
	}

	offen, fehler := p.ausgang.Offene(ctx, empfaenger)
	if fehler != nil {
		return nil, false, fehler
	}

	// **Was schon bestätigt ist, geht nicht noch einmal hinaus.** Der
	// Ausgang weiß nur, was offen ist; welche Folge drüben angekommen ist,
	// sagt der ausgetauschte Fortschritt.
	nachzuliefern = make([]Ausgangsnachricht, 0, len(offen))

	for _, zeile := range offen {
		if zeile.Folge >= abFolge {
			nachzuliefern = append(nachzuliefern, zeile)
		}
	}

	return nachzuliefern, false, nil
}
