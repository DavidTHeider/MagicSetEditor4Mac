package peer

// **Die zwei letzten Schritte und die Sitzung, die daraus entsteht** —
// Werkstück N1 von Phase 5.4
// (#1454).

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Sitzung ist, worauf sich beide Seiten geeinigt haben.
type Sitzung struct {
	// Kennung ist aus beiden Zufallszahlen gerechnet.
	//
	// **Beide Seiten kommen auf dieselbe, und keine hat sie gewählt.**
	// Eine Kennung, die eine Seite bestimmt, ist eine Kennung, die eine
	// Seite wiederverwenden kann.
	Kennung string

	// Gegenueber ist die Kennung am anderen Ende — **der Nachweis**.
	Gegenueber Kennung

	// Fingerabdruck ist [Gegenueber] in einer Form, die zwei Menschen
	// vergleichen. Er wird **aus der Kennung gerechnet**, nicht aus der
	// Nachricht gelesen.
	Fingerabdruck string

	// SelbstgewaehlterName ist, wie das Gegenüber sich nennt.
	//
	// **Er belegt nichts.** Das Issue sagt es wörtlich: „Anzeigenamen
	// allein sind kein Identitätsnachweis" — das Feld heißt so, damit
	// niemand ihn für einen hält.
	SelbstgewaehlterName string

	Zweck      domain.Zweck
	Kanal      string
	MeineRolle string
	SeineRolle string
}

// Schliessen nimmt die Annahme an und gibt den Abschluss für die Gegenseite.
func (a Anbahnung) Schliessen(roh []byte) (Sitzung, []byte, error) {
	huelle, fehler := Auspacke(roh)
	if fehler != nil {
		return Sitzung{}, nil, fehler
	}

	if huelle.Art != ArtAnnahme {
		return Sitzung{}, nil, fmt.Errorf(
			"%w: hier steht eine Annahme und kein %q", domain.ErrValidation,
			huelle.Art)
	}

	gegenueber := Kennung(huelle.Absender)

	if befund := pruefeGegenueber(
		a.ich.kennung, gegenueber, a.wunsch.Erwartet); befund != nil {
		return Sitzung{}, nil, befund
	}

	var annahme Annahme

	if befund := json.Unmarshal(huelle.Rumpf, &annahme); befund != nil {
		return Sitzung{}, nil, domain.MitGrund(
			domain.ErrValidation, befund, "der Rumpf ist keine Annahme")
	}

	if befund := a.pruefeAnnahme(huelle, annahme); befund != nil {
		return Sitzung{}, nil, befund
	}

	abschluss, fehler := a.ich.verpackeSchritt(
		dieZwecke[a.wunsch.Zweck], ArtAbschluss, a.nonce, 2, Abschluss{
			Fassung:   Handschlagfassung,
			DeinNonce: annahme.MeinNonce,
			MeinNonce: a.nonce,
		})
	if fehler != nil {
		return Sitzung{}, nil, fehler
	}

	return Sitzung{
		Kennung: sitzungskennung(
			a.wunsch.Zweck, a.wunsch.Kanal, a.ich.kennung, gegenueber,
			a.nonce, annahme.MeinNonce),
		Gegenueber:           gegenueber,
		Fingerabdruck:        Fingerabdruck(gegenueber),
		SelbstgewaehlterName: annahme.Name,
		Zweck:                a.wunsch.Zweck,
		Kanal:                a.wunsch.Kanal,
		MeineRolle:           a.wunsch.Rolle,
		SeineRolle:           annahme.Rolle,
	}, abschluss, nil
}

// pruefeAnnahme hält die Antwort gegen das eigene Angebot.
func (a Anbahnung) pruefeAnnahme(huelle Huelle, annahme Annahme) error {
	if befund := pruefeFassung(annahme.Fassung); befund != nil {
		return befund
	}

	if annahme.Zweck != a.wunsch.Zweck {
		return fmt.Errorf(
			"%w: angeboten war %q, bestätigt wurde %q", domain.ErrValidation,
			a.wunsch.Zweck, annahme.Zweck)
	}

	if annahme.Rolle == "" {
		return fmt.Errorf(
			"%w: die Gegenseite nennt keine Rolle", domain.ErrValidation)
	}

	if befund := pruefeKanal(annahme.Kanal, a.wunsch.Kanal); befund != nil {
		return befund
	}

	// **Hier steht der Beleg, dass gerade jemand da ist.** Die Gegenseite
	// hat meine Zahl unterschrieben; ein Mitschnitt aus einem früheren Lauf
	// trüge eine andere.
	if annahme.DeinNonce != a.nonce {
		return fmt.Errorf(
			"%w: die Antwort kennt meine Zufallszahl nicht",
			domain.ErrValidation)
	}

	if huelle.Sitzung != a.nonce {
		return fmt.Errorf(
			"%w: die Antwort gehört zu einer anderen Sitzung",
			domain.ErrValidation)
	}

	return pruefeZufallszahl(annahme.MeinNonce)
}

// Abschliessen nimmt den dritten Schritt an.
func (e Entgegennahme) Abschliessen(roh []byte) (Sitzung, error) {
	huelle, fehler := Auspacke(roh)
	if fehler != nil {
		return Sitzung{}, fehler
	}

	if huelle.Art != ArtAbschluss {
		return Sitzung{}, fmt.Errorf(
			"%w: hier steht ein Abschluss und kein %q", domain.ErrValidation,
			huelle.Art)
	}

	// **Derselbe wie im Angebot und nicht irgendeiner.** Ohne diesen Satz
	// schlösse ein Dritter die Sitzung ab, die zwei andere begonnen haben.
	if Kennung(huelle.Absender) != e.gegenueber {
		return Sitzung{}, fmt.Errorf(
			"%w: den Handschlag hat %s begonnen und %s schließt ihn",
			domain.ErrValidation, Fingerabdruck(e.gegenueber),
			Fingerabdruck(Kennung(huelle.Absender)))
	}

	var abschluss Abschluss

	if befund := json.Unmarshal(huelle.Rumpf, &abschluss); befund != nil {
		return Sitzung{}, domain.MitGrund(
			domain.ErrValidation, befund, "der Rumpf ist kein Abschluss")
	}

	if befund := pruefeFassung(abschluss.Fassung); befund != nil {
		return Sitzung{}, befund
	}

	if abschluss.DeinNonce != e.nonceAnnehmer ||
		abschluss.MeinNonce != e.nonceAnbieter {
		return Sitzung{}, fmt.Errorf(
			"%w: der Abschluss kennt die Zufallszahlen dieser Sitzung nicht",
			domain.ErrValidation)
	}

	return Sitzung{
		Kennung: sitzungskennung(
			e.zweck, e.kanal, e.gegenueber, e.ich.kennung,
			e.nonceAnbieter, e.nonceAnnehmer),
		Gegenueber:           e.gegenueber,
		Fingerabdruck:        Fingerabdruck(e.gegenueber),
		SelbstgewaehlterName: e.seinName,
		Zweck:                e.zweck,
		Kanal:                e.kanal,
		MeineRolle:           e.meineRolle,
		SeineRolle:           e.seineRolle,
	}, nil
}

// sitzungskennung rechnet die Kennung, auf die beide Seiten kommen.
//
// **Die Reihenfolge ist die des Ablaufs** — Anbieter zuerst —, und deshalb
// kennen beide Seiten sie, ohne sie auszuhandeln.
//
// **Gerahmt wie die Unterschriftsform** (`<Länge>:<Bytes>`): eine Kennung, die
// aus aneinandergehängten Feldern entstünde, ließe sich von zwei verschiedenen
// Sitzungen treffen.
func sitzungskennung(
	zweck domain.Zweck, kanal string, anbieter, annehmer Kennung,
	nonceAnbieter, nonceAnnehmer string,
) string {
	var werk bytes.Buffer

	for _, teil := range []string{
		"mse4mac/sitzung/1",
		string(zweck),
		kanal,
		string(anbieter),
		string(annehmer),
		nonceAnbieter,
		nonceAnnehmer,
	} {
		fmt.Fprintf(&werk, "%d:%s", len(teil), teil)
	}

	summe := sha256.Sum256(werk.Bytes())

	return hex.EncodeToString(summe[:])
}
