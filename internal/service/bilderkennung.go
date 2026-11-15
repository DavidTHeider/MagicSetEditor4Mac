package service

// **Aus einem Kartenbild werden Kandidaten** — Werkstück A13.3 von Phase 5.3
// (#1452).
//
// # Die Arbeitsteilung, und sie ist der Kern
//
// **Das Modell liest. Die Suche findet. Ein Mensch wählt.**
//
// Das Issue sagt es in einem Satz: „Vision/Kartenerkennung liefert mögliche
// Identitäten mit Unsicherheit und **expliziter Zuordnung über A2**;
// Menge/Exemplar wird **nicht automatisch** angelegt."
//
// Daraus folgt der Aufbau: was das Modell sagt, sind [domain.Beobachtung]en —
// gelesener Text, erkanntes Set, beschriebene Illustration. Daraus wird ein
// **Suchbegriff**, und die Kandidaten kommen aus **derselben** Suche, die
// eine getippte Kartenliste benutzt. Eine Bilderkennung kann damit keinen
// Druck nennen, den die Suche nicht auch gefunden hätte.
//
// # Warum hier nichts geschrieben wird
//
// **Weil das Issue es ausschließt** — und weil es auch ohne die Zeile richtig
// wäre: „Counterspell" ohne Setangabe passt auf 72 Drucke (nachgelesen am
// 2026-08-30), und welchen jemand in der Hand hält, weiß nur er.

import (
	"context"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Erkennungsgrenze ist, wie viele Kandidaten höchstens genannt werden.
//
// **Dieselbe wie bei der Listendeutung** ([Deutungsgrenze]): mehr Zeilen
// helfen niemandem bei der Auswahl, und die Zahl der **passenden** Drucke
// steht ohnehin daneben.
const Erkennungsgrenze = 10

// Bildkandidat ist ein Druck, der zu den Beobachtungen passt.
type Bildkandidat struct {
	Treffer Suchtreffer

	// Gestuetzt sind die Evidenzarten, die **diesen** Kandidaten stützen.
	//
	// **Je Kandidat und nicht einmal für alle**: der Name passt auf jeden,
	// das Set nur auf einige — und genau daran sieht man, was die Auswahl
	// noch offenlässt.
	Gestuetzt []domain.Evidenzart
}

// Bilddeutung ist, was aus einem Kartenbild zu machen wäre.
//
// **Ohne Bindung und ohne Menge.** Sie ist eine Auskunft, kein Vorgang.
type Bilddeutung struct {
	// Erkennung ist, was das Modell gelesen hat — roh und getrennt.
	Erkennung domain.Bilderkennung

	// Befunde sind die Beanstandungen an der Auskunft des Modells.
	Befunde []domain.Belegbefund

	// Kandidaten sind die Drucke, die passen — höchstens [Erkennungsgrenze].
	Kandidaten []Bildkandidat

	// Drucke ist die Zahl der passenden Drucke, **auch wenn die Liste kürzer
	// ist**.
	//
	// **Das ist die Unsicherheit**, und sie ist gezählt statt geschätzt: eine
	// Karte, auf die 72 Drucke passen, ist nicht zu 14 % erkannt — sie ist
	// erkannt, und der **Druck** ist offen.
	Drucke int

	// Weitere sagt, dass Scryfall **noch mehr** hat, als gezählt wurde.
	//
	// **Ohne sie wäre `Drucke` eine Behauptung**: die Suche gibt eine Seite,
	// und `Weitere` ist die Auskunft, dass es dahinter weitergeht
	// ([scryfall.Treffer], das nicht von selbst nachlädt). Wer sie wegließe,
	// nennte 175 als „die Zahl", wo sie „mindestens" heißt.
	Weitere bool

	// Alter ist das Alter der Suchantwort — sie darf aus dem
	// Zwischenspeicher kommen (§16, Regel 5).
	Alter string

	// Modell ist, wer das Bild gelesen hat.
	Modell string
}

// Eindeutig sagt, ob genau ein Druck passt.
//
// **Und selbst dann wird nichts angelegt.** Das Issue lässt hier keinen
// Spielraum: „Menge/Exemplar wird nicht automatisch angelegt." Dieser Wert
// sagt der Oberfläche, dass sie **einen** Vorschlag hervorheben darf — nicht,
// dass sie ihn nehmen soll.
func (b Bilddeutung) Eindeutig() bool {
	return b.Drucke == 1 && len(b.Kandidaten) == 1 && !b.Weitere
}

// Bilderkenner macht aus einem Bild Kandidaten.
type Bilderkenner struct {
	suche Fremddrucksuche
}

// NeuerBilderkenner baut ihn.
func NeuerBilderkenner(suche Fremddrucksuche) *Bilderkenner {
	return &Bilderkenner{suche: suche}
}

// Deuten prüft die Beobachtungen und sucht die passenden Drucke.
//
// **Es fragt kein Modell** — was das Modell gesagt hat, kommt herein. Wer es
// fragt, ist `internal/app`; hier steht, was daraus wird.
//
// **Und es weist eine unbrauchbare Auskunft ab, statt zu suchen.** Ein Bild
// ohne eine einzige gelesene Stelle ergibt keinen Suchbegriff, und eine Suche
// über die leere Zeichenkette fände alles.
func (b *Bilderkenner) Deuten(
	ctx context.Context, erkennung domain.Bilderkennung,
) (Bilddeutung, error) {
	deutung := Bilddeutung{
		Erkennung:  erkennung,
		Befunde:    domain.PruefeBilderkennung(erkennung),
		Kandidaten: []Bildkandidat{},
	}

	begriff := erkennung.Suchbegriff()
	if begriff == "" {
		// **Kein Fehler, sondern eine Antwort.** „Ich konnte nichts lesen"
		// ist eine Auskunft über das Bild — und die Befunde sagen, warum.
		return deutung, nil
	}

	treffer, alter, _, fehler := b.suche.Druckesuche(ctx, begriff)
	if fehler != nil {
		return Bilddeutung{}, fehler
	}

	deutung.Drucke = len(treffer.Karten)
	deutung.Weitere = treffer.Weitere
	deutung.Alter = alter.String()

	for i, karte := range treffer.Karten {
		if i >= Erkennungsgrenze {
			break
		}

		deutung.Kandidaten = append(deutung.Kandidaten, Bildkandidat{
			Treffer:   alsSuchtreffer(karte),
			Gestuetzt: gestuetztVon(erkennung, alsSuchtreffer(karte)),
		})
	}

	return deutung, nil
}

// gestuetztVon sagt, welche Beobachtungen zu **diesem** Druck passen.
//
// # Warum das hier entschieden wird und nicht vom Modell
//
// **Weil es nachprüfbar ist.** Ob der gelesene Setcode zu einem Druck gehört,
// steht im Druck; ein Modell, das „passt gut" sagte, sagte etwas über sich
// selbst. Deshalb vergleicht diese Funktion Zeichenketten und nichts weiter.
//
// **Und deshalb steht das Kunstwerk nicht darin**: eine Bildbeschreibung
// lässt sich gegen keinen Druck halten.
func gestuetztVon(
	erkennung domain.Bilderkennung, treffer Suchtreffer,
) []domain.Evidenzart {
	gestuetzt := make([]domain.Evidenzart, 0, 3)

	for _, beobachtung := range erkennung.Beobachtungen {
		if !beobachtung.Lesbar || beobachtung.Wert == "" {
			continue
		}

		if stuetzt(beobachtung, treffer) {
			gestuetzt = append(gestuetzt, beobachtung.Art)
		}
	}

	return gestuetzt
}

// stuetzt hält **eine** Beobachtung gegen einen Druck.
//
// **Nicht `passt`**: den Namen trägt in diesem Paket schon der Filter der
// Recherche, und zwei Bedeutungen unter einem Namen sind eine Falle für den
// nächsten Leser.
func stuetzt(beobachtung domain.Beobachtung, treffer Suchtreffer) bool {
	// **Ohne Rücksicht auf Groß- und Kleinschreibung**: ein Modell liest,
	// was dasteht — und auf einer Karte steht „LIGHTNING BOLT" in
	// Kapitälchen, während Scryfall „Lightning Bolt" führt.
	switch beobachtung.Art {
	case domain.EvidenzName:
		return strings.EqualFold(beobachtung.Wert, treffer.Name)
	case domain.EvidenzSetcode:
		return strings.EqualFold(beobachtung.Wert, treffer.Setcode)
	case domain.EvidenzNummer:
		return beobachtung.Wert == treffer.Nummer
	case domain.EvidenzTypzeile:
		return strings.EqualFold(beobachtung.Wert, treffer.Typzeile)
	case domain.EvidenzLayout, domain.EvidenzKunstwerk:
		// **Sie stützen keinen Kandidaten.** Eine Bauart passt auf Tausende,
		// eine Bildbeschreibung lässt sich gegen nichts halten — beide stehen
		// in der Antwort, damit jemand sieht, was das Modell gesehen hat.
		return false
	}

	return false
}
