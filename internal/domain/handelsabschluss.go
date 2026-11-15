package domain

// **Was ein Handel im Bestand bewirkt** — Werkstück `T6` von Phase 5.6
// (#1471).
//
// # Warum das hier steht und nicht im Bestand
//
// **Weil es die Sprache ist und nicht die Ablage.** Ein Abschluss ist ein
// Vorgang zwischen zwei Menschen — Angebot, Beleg, Unterschrift, und was er
// bewegt. Dass eine Datenbank ihn in einer Transaktion schreibt, ist die
// Antwort auf eine andere Frage.
//
// **Und weil sonst drei Schichten dreimal dasselbe erklärten.** Stünde er in
// `storage`, dürfte `service` ihn nicht nennen (§2) und bräuchte einen
// eigenen Verbund daneben, dazu eine Übersetzung, die nichts übersetzt. Genau
// das ist bei `Ausgangsnachricht` geschehen — und der Preis war ein
// Ausgangsweg, für den es keinen Aufrufer gibt.

import "strings"

// Belegnachricht ist der Beleg, der an die Gegenseite geht.
//
// **Ohne Vorgang**: der steht im [Handelsbeleg], und zweimal dieselbe Angabe
// läuft irgendwann auseinander.
type Belegnachricht struct {
	NachrichtID string
	Empfaenger  string

	// Nachricht ist Umschlag und Nutzlast, fertig verpackt.
	Nachricht string
}

// Handelsabschluss ist, was ein bestätigter Handel im Bestand bewirkt.
type Handelsabschluss struct {
	// Angebot ist die Revision, die abgeschlossen wird.
	// Vorher ist der Stand, von dem aus verbucht wird — geschrieben wird
	// nur, wenn er noch gilt (Befund Z1 des Zweitblicks über Phase 6.8).
	// Eine leere Revision heißt: es gab keinen, und `Angebot` wird angelegt.
	Vorher Handelsangebot

	Angebot Handelsangebot

	// Beleg trägt die Zeilen — Abgänge und Zugänge in **einer** Liste.
	//
	// **Eine und nicht zwei**: der Beleg geht so hinaus, wie er hier steht.
	// Zwei Listen, aus denen erst beim Schreiben eine wird, hätten zwei
	// Reihenfolgen — und die unterschriebene wäre nicht die gebuchte.
	Beleg Handelsbeleg

	// Unterschrift ist **vorher** gerechnet.
	//
	// **Weil in einer offenen Transaktion nichts warten darf**: eine
	// Unterschrift kommt aus dem Schlüsselbund, und der fragt zurück.
	Unterschrift string

	// Vermerk hält die Nachricht fest, die den Abschluss ausgelöst hat —
	// `nil`, wenn keine dahintersteht.
	Vermerk *Inboxeintrag

	// Ausgang ist der Beleg, der hinausgeht — `nil`, wenn keiner hinausgeht.
	Ausgang *Belegnachricht
}

// Pruefe sagt, ob dieser Abschluss überhaupt einer ist.
//
// **Vor jeder Transaktion.** Was ohne die Datenbank zu entscheiden ist, wird
// ohne sie entschieden — sonst hält eine offene Transaktion auf, was gar
// nicht erst hätte beginnen müssen.
func (h Handelsabschluss) Pruefe() error {
	switch {
	case h.Beleg.Vorgang != h.Angebot.Vorgang:
		return MitGrund(ErrValidation, nil,
			"der Beleg gehört zu einem anderen Vorgang als das Angebot")

	// **Der Abdruck wird nachgerechnet** — er ist die Zusage, dass
	// unterschrieben wurde, was gebucht wird.
	case h.Beleg.Abdruck != h.Angebot.Abdruck():
		return MitGrund(ErrConflict, nil,
			"der Beleg nennt einen anderen Stand als das Angebot")

	case strings.TrimSpace(h.Unterschrift) == "":
		return MitGrund(ErrValidation, nil,
			"ein Abschluss ohne Unterschrift ist keiner")

	// **Ein unverbuchter Zustand ist kein Abschluss.** Er ist das, was davor
	// kommt — und wer ihn hier schriebe, machte aus einer offenen Übergabe
	// eine erledigte.
	case h.Angebot.Zustand != HandelLokalVerbucht &&
		h.Angebot.Zustand != HandelBeidseitigVerbucht:
		return MitGrund(ErrConflict, nil,
			"ein Abschluss schreibt einen verbuchten Zustand und keinen anderen")

	case len(h.Beleg.Zeilen) == 0:
		return MitGrund(ErrValidation, nil,
			"ein Abschluss ohne Bewegung ist keiner")
	}

	for _, zeile := range h.Beleg.Zeilen {
		if fehler := zeile.Pruefe(); fehler != nil {
			return fehler
		}

		// **Ein Abgang nennt die Zeile, aus der er kommt.** Ohne sie wüsste
		// der Bestand nicht, wovon er abziehen soll — und ein Zugang braucht
		// sie nicht, weil er seine erst anlegt.
		if zeile.Art == JournalAbgang && zeile.ExemplarID == 0 {
			return MitGrund(ErrValidation, nil,
				"ein Abgang nennt die Besitzzeile, aus der er kommt")
		}
	}

	return nil
}
