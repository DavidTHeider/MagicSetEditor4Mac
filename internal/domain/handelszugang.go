package domain

// **Was die empfangende Seite verbuchen kann** —
// #2495.
//
// # Warum das eine Regel ist und nicht zwei
//
// **Weil sonst zwei Stellen dasselbe verschieden sagten.** Die Verbuchung
// bucht den Zugang aus der Zeile des Tauschs und prüft ihn dort
// (`service/handelsverbuchung.go`); eine Zeile ohne Sprache — „Wunsch aus
// einem Deck eintragen" legte solche an — ließ sich bis dahin anbieten,
// annehmen und übergeben und scheiterte erst danach. **Die Prüfung davor ist
// jetzt dieselbe Zeile mit derselben Validierung**, nur früher gefragt: beim
// Schritt, der auf eine Einigung zuführt.
//
// # Was sie nicht prüft
//
// **Ob die Zeile zum Bestand der gebenden Seite passt** — das sagt das
// Festhalten (#2492). Hier geht es um die andere Seite: ob sie buchen kann,
// was ankommt.

import "strconv"

// Zugangszeile ist die Zeile, mit der die empfangende Seite diese Position
// bucht.
//
// **Eine Druckzeile ohne Kennung geht nicht mit**: `DruckID` gilt nur im
// Bestand, der sie vergeben hat, und drüben ist sie keine.
func (p Handelsposition) Zugangszeile(vorgang string) Journalzeile {
	return Journalzeile{
		Art: JournalZugang, Anlass: AnlassTausch,
		Vorgang: vorgang, Printing: p.Printing,
		Menge: p.Menge, Zustand: Zustand(p.Guete),
		Sprache: p.Sprache, Finish: Finish(p.Finish),
	}
}

// PruefeVerbuchbar sagt, ob jede Zeile eines Tauschs auf der Seite, die sie
// bekommt, verbucht werden kann — oder welche nicht.
//
// **Am Feld und mit der Zeile** (§6, §8): der Satz nennt, was fehlt, der
// Schlüssel trägt die Nummer, und das Fenster sagt, wo zu ändern ist.
func (a Handelsangebot) PruefeVerbuchbar() error {
	if !a.Art.BewegtBestand() {
		return nil
	}

	for stelle, position := range a.Positionen {
		grund := verbuchungsbefund(position.Zugangszeile(a.Vorgang).AlsExemplar())
		if grund == "" {
			continue
		}

		zeile := strconv.Itoa(stelle + 1)

		return NewFieldError("positionen", SchluesselZeileNichtVerbuchbar,
			"Die Zeile %s lässt sich so nicht verbuchen — %s", zeile, grund).
			MitWerten(map[string]string{"zeile": zeile})
	}

	return nil
}

// verbuchungsbefund nennt, was an einem empfangenen Exemplar fehlt.
//
// **Der Zustand dazu**: [Exemplar.Validate] lässt ihn leer, weil jeder andere
// Weg vorher [Exemplar.Normalisieren] ruft — der Zugang tut das nicht und
// trüge ein Exemplar ohne Zustand ein. Eine Zeile ohne ihn ließe sich
// außerdem nie festhalten: kein Exemplar im Bestand ist ohne Zustand.
func verbuchungsbefund(exemplar Exemplar) string {
	if befunde := exemplar.Validate(); len(befunde) > 0 {
		return befunde[0].Message
	}

	if exemplar.Zustand == "" {
		return "Ein Druck hat einen Zustand."
	}

	return ""
}
