package domain

// **Der Beleg der Gegenseite** —
// #1650.
//
// # Warum es das gibt
//
// **Weil `beidseitig` bis zum 2026-09-19 eine Behauptung war.** Die Spalte
// sagt „auch die Gegenseite hat unterschrieben"; gesetzt wurde sie davon,
// dass **hier** jemand für beide Parteien bestätigt hat. Ihr eigener
// Kommentar schloss genau das aus: „zwei Unterschriften vorzutäuschen, weil
// zwei Parteien dastehen, wäre die Zusage, die dieses Werkstück ausdrücklich
// nicht gibt."
//
// **Und der Weg dafür stand gebaut da, ohne Aufrufer**: [Belegnachricht] —
// „der Beleg, der an die Gegenseite geht" — wurde in keinem Produktivweg je
// gefüllt. Derselbe Befund, der Phase 5.6 viermal getragen hat.

import (
	"slices"
	"time"
)

// Belegurkunde ist ein unterschriebener Abschluss, wie er hinausgeht.
//
// **Der ganze Beleg und nicht nur seine Form.** Wer die Form als Zeichenkette
// schickte, könnte die Unterschrift prüfen und nichts weiter: was
// unterschrieben wurde, stünde nur noch als Text da. Mit dem Beleg lässt sich
// die Form **nachrechnen** — und danach die Frage stellen, die zählt: gehört
// das zu **meinem** Handel?
type Belegurkunde struct {
	Beleg Handelsbeleg

	// Unterschrift ist die des Absenders über [Handelsbeleg.Unterschriftsform],
	// hexadezimal.
	Unterschrift string
}

// Oeffentlich gibt den Beleg ohne das, was einen Fremden nichts angeht.
//
// **Dieselbe Grenze wie in [Handelsbeleg.Unterschriftsform]**, nur an der
// anderen Stelle: jene **rechnet** ohne Bestandskennung, Lagerort und Notiz,
// diese **schickt** ohne sie. Wer nur die erste hätte, unterschriebe das
// Öffentliche und versendete das Private — und niemand sähe es, weil die
// Unterschrift trotzdem passt.
func (b Handelsbeleg) Oeffentlich() Handelsbeleg {
	offen := b
	offen.Zeilen = make([]Journalzeile, 0, len(b.Zeilen))

	for _, zeile := range b.Zeilen {
		offen.Zeilen = append(offen.Zeilen, Journalzeile{
			Art: zeile.Art, Anlass: zeile.Anlass, Vorgang: zeile.Vorgang,
			Printing: zeile.Printing, Menge: zeile.Menge,
			Zustand: zeile.Zustand, Sprache: zeile.Sprache,
			Finish: zeile.Finish, Gebucht: zeile.Gebucht,
		})
	}

	return offen
}

// Fremdbeleg ist, was von einer geprüften Urkunde im Bestand bleibt.
type Fremdbeleg struct {
	Vorgang string

	// Von ist die Peerkennung dessen, der unterschrieben hat.
	//
	// **Sie kommt aus der Leitung und nicht aus der Nutzlast** — wer spricht,
	// sagt die Sitzung; was er über sich behauptet, ist eine Behauptung.
	Von string

	Abdruck      string
	Unterschrift string

	Empfangen time.Time
}

// PasstZu sagt, ob diese Urkunde denselben Handel beschreibt wie der eigene
// Beleg — **gespiegelt**.
//
// # Warum die Unterschrift hier keine Rolle spielt
//
// **Weil sie eine andere Frage beantwortet.** Ob sie echt ist, sagt
// `peer.Pruefe` gegen die Kennung des Absenders; ob sie **hierher** gehört,
// sagt niemand sonst. Eine echte Unterschrift unter einem fremden Handel ist
// echt und trotzdem wertlos.
//
// # Und warum gespiegelt
//
// **Weil ein Handel zwei Seiten hat.** Was hier abgeht, kommt drüben an. Ein
// Gegenbeleg mit denselben Zeilen in derselben Richtung beschriebe einen
// Handel, bei dem beide dasselbe weggeben — und er käme durch, wenn man nur
// Vorgang und Abdruck prüfte.
func (u Belegurkunde) PasstZu(eigener Handelsbeleg) error {
	if abweichung := u.Abweichung(eigener); abweichung != nil {
		return abweichung
	}

	return nil
}

// Abweichung ist [Belegurkunde.PasstZu] mit dem konkreten Typ — `nil`, wenn
// die Urkunde passt (#2500).
//
// **Zwei Wege und nicht einer**, weil ein `*Belegabweichung` als `error`
// zurückgegeben nie `nil` wäre: wer den Mangel braucht, fragt hier, wer nur
// „passt es?" fragt, dort.
func (u Belegurkunde) Abweichung(eigener Handelsbeleg) *Belegabweichung {
	switch {
	case u.Beleg.Vorgang != eigener.Vorgang:
		return abweichung(MangelVorgang,
			"der Beleg gehört zum Vorgang %q und nicht zu %q",
			u.Beleg.Vorgang, eigener.Vorgang)

	case u.Beleg.Abdruck != eigener.Abdruck:
		return abweichung(MangelRevision,
			"der Beleg gehört zur Revision %q und nicht zu %q",
			u.Beleg.Abdruck, eigener.Abdruck)

	case !gleicheParteien(u.Beleg.Parteien, eigener.Parteien):
		return abweichung(MangelParteien,
			"der Beleg nennt die Parteien %v und nicht %v",
			u.Beleg.Parteien, eigener.Parteien)

	case len(u.Beleg.Zeilen) != len(eigener.Zeilen):
		return abweichung(MangelZeilen,
			"der Beleg trägt %d Zeilen und meiner %d",
			len(u.Beleg.Zeilen), len(eigener.Zeilen))
	}

	return spiegeln(u.Beleg.Zeilen, eigener.Zeilen)
}

// gleicheParteien vergleicht als **Menge** und nicht als Reihenfolge.
//
// **Weil die Reihenfolge der Parteien nichts bedeutet.** Wer sie hier
// erzwänge, wiese einen richtigen Beleg ab, weil das andere Programm anders
// sortiert — und niemand fände den Grund.
func gleicheParteien(dort, hier []string) bool {
	if len(dort) != len(hier) {
		return false
	}

	seine := slices.Clone(dort)
	meine := slices.Clone(hier)

	slices.Sort(seine)
	slices.Sort(meine)

	return slices.Equal(seine, meine)
}

// spiegeln prüft, dass zu jeder eigenen Zeile eine entgegengesetzte steht.
func spiegeln(dort, hier []Journalzeile) *Belegabweichung {
	offen := slices.Clone(dort)

	for _, meine := range hier {
		gegen, gilt := gegenrichtung(meine.Art)
		if !gilt {
			return abweichung(MangelZeilen,
				"eine Zeile der Art %q gehört in keinen Handelsbeleg",
				meine.Art)
		}

		stelle := slices.IndexFunc(offen, func(seine Journalzeile) bool {
			return seine.Art == gegen &&
				seine.Printing == meine.Printing &&
				seine.Menge == meine.Menge &&
				seine.Zustand == meine.Zustand &&
				seine.Sprache == meine.Sprache &&
				seine.Finish == meine.Finish
		})

		if stelle < 0 {
			return abweichung(MangelZeilen,
				"zu meiner Zeile %s %d×%s steht drüben keine entgegengesetzte",
				meine.Art, meine.Menge, meine.Printing)
		}

		offen = slices.Delete(offen, stelle, stelle+1)
	}

	return nil
}

// gegenrichtung dreht eine Journalart um — und sagt, ob es eine war.
//
// **Die zweite Rückgabe ist kein Zierat.** [Journalart] ist eine
// Zeichenkette, und was in einem Beleg von außen ankommt, ist eine
// Behauptung: eine Zeile der Art `"korrektur"` hat keine Gegenrichtung, und
// wer sie stillschweigend als Abgang läse, ließe sie durch.
func gegenrichtung(art Journalart) (Journalart, bool) {
	switch art {
	case JournalAbgang:
		return JournalZugang, true
	case JournalZugang:
		return JournalAbgang, true
	default:
		return "", false
	}
}
