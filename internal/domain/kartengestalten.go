package domain

// **Spielsteine, Embleme, Kontrolle und Seiten** — Werkstück M12 der
// Regelmaschine
// (#1614).
//
// # Vier Dinge, die keine Karte brauchen
//
//   - **Einen Spielstein erzeugen** (CR 111.1, CR 701.7a) — er vertritt
//     keine Karte, und seine Eigenschaften nennt der, der ihn erzeugt.
//   - **Ein Emblem erzeugen** (CR 114.1) — es liegt in der Kommandozone und
//     gehört dem, der es bekommt (CR 114.2).
//   - **Die Kontrolle wechseln** (CR 110.2, CR 613.1b) — und zwar **ohne**
//     das Objekt zu bewegen und **ohne** es zu enttappen.
//   - **Eine Seite umdrehen** (CR 701.27a) — bei einer Karte, die zwei hat
//     (CR 712.9).
//
// # Und eine Warnung an den nächsten Leser
//
// **Die Nummern hier sind nachgeschlagen und waren beim ersten Anlauf
// falsch.** `CR 711` heißt **Leveler Cards**; das Umwandeln steht in
// CR 701.27, die zweiseitige Karte in CR 712. Der erste Entwurf zitierte
// durchweg `711.2` — plausibel, weil 712 danebensteht, und trotzdem eine
// Regel über etwas ganz anderes. Eine erfundene Nummer ist schlimmer als
// keine (§11).
//
// # Und was hier nicht steht
//
// **Jede Gestalt, die eine Fähigkeit ist.** Saga (CR 714), Class (716),
// Attraction (717), Prototype (718), Case (719), Omen (720), Station (721)
// und das Verlies (309) sind Kartentexte mit Zählern und Auslösern; sie
// brauchen `M8` (#1610)
// und `M9` (#1611).
// Das steht als Grenze in der Regelmatrix.

import "fmt"

// Spielstein erzeugt einen Spielstein auf dem Schlachtfeld — CR 701.7a.
//
// **Die Eigenschaften kommen mit**, in denselben Feldnamen wie in einem
// MSE-Paket — der Grund steht bei [Spielobjekt.Felder].
//
// **Er gehört dem, der ihn erzeugt, und der beherrscht ihn** — CR 111.2 sagt
// beides. Und er kommt mit dem laufenden Zug (CR 302.6): ein frisch
// erzeugter Spielstein greift nicht sofort an.
func (p Partie) Spielstein(wer, kennung string, felder map[string]string) (Partie, error) {
	if fehler := p.neuesObjektPruefen(wer, kennung); fehler != nil {
		return p, fehler
	}

	if len(felder) == 0 {
		return p, fmt.Errorf(
			"%w: ein Spielstein ohne Eigenschaften ist keiner (CR 701.7a)",
			ErrValidation)
	}

	eigene := make(map[string]string, len(felder))
	for name, wert := range felder {
		eigene[name] = wert
	}

	p.Objekte = append(append([]Spielobjekt{}, p.Objekte...), Spielobjekt{
		Kennung: kennung, Art: ObjektSpielstein,
		Beherrscher: wer, Eigner: wer,
		Zone: ZoneTisch, Felder: eigene, Seit: p.Zug.Nummer,
	})

	return p, nil
}

// Emblem erzeugt ein Emblem in der Kommandozone — CR 114.1, CR 114.2.
//
// **Es gehört dem, der es bekommt, und er beherrscht es** — CR 114.2 sagt
// beides ausdrücklich.
//
// **Ohne Eigenschaften**: CR 114.3 sagt es ausdrücklich — keine Typen, keine
// Manakosten, keine Farbe; was ein Emblem hat, sind **Fähigkeiten**. Eine
// Fähigkeit steht in einem Kartentext, und der ist `M8`; was hier entsteht,
// ist die Marke.
func (p Partie) Emblem(wer, kennung string) (Partie, error) {
	if fehler := p.neuesObjektPruefen(wer, kennung); fehler != nil {
		return p, fehler
	}

	p.Objekte = append(append([]Spielobjekt{}, p.Objekte...), Spielobjekt{
		Kennung: kennung, Art: ObjektEmblem,
		Beherrscher: wer, Eigner: wer,
		Zone: ZoneKommando, Seit: p.Zug.Nummer,
	})

	return p, nil
}

// neuesObjektPruefen prüft Sitz und Kennung eines neuen Objekts.
func (p Partie) neuesObjektPruefen(wer, kennung string) error {
	if !p.sitzSteht(wer) {
		return fmt.Errorf("%w: den Sitz %q gibt es in dieser Partie nicht",
			ErrNotFound, wer)
	}

	if kennung == "" {
		return fmt.Errorf("%w: ein Objekt ohne Kennung gibt es nicht",
			ErrValidation)
	}

	for _, objekt := range p.Objekte {
		if objekt.Kennung == kennung {
			return fmt.Errorf(
				"%w: die Kennung %q ist in dieser Partie schon vergeben",
				ErrConflict, kennung)
		}
	}

	return nil
}

// Beherrschen wechselt die Kontrolle über ein Permanent — CR 110.2 mit
// CR 613.1b.
//
// **Zwei Nummern, weil es zwei Aussagen sind**: CR 110.2 sagt, dass jedes
// Permanent einen Beherrscher hat und wer es per Vorgabe ist; CR 613.1b
// sagt, dass ein Effekt ihn in Schicht 2 ändert. **CR 110.2a ist es
// nicht** — die spricht davon, unter wessen Kontrolle etwas **ins Spiel
// kommt**, und das ist eine andere Frage.
//
// # Drei Dinge, die dabei **nicht** geschehen
//
// **Es bewegt sich nicht.** Ein Kontrollwechsel ist kein Zonenwechsel; das
// Objekt bleibt, wo es liegt, und bleibt dasselbe Objekt (CR 400.7 greift
// nicht, weil nichts die Zone wechselt).
//
// **Es enttappt nicht.** Getappt ist ein Zustand des Permanents (CR 110.5),
// und ein neuer Beherrscher ändert daran nichts.
//
// **Aber es verlässt den Kampf** — CR 506.4 nennt den Kontrollwechsel
// ausdrücklich: ein Permanent, dessen Beherrscher wechselt, ist danach weder
// Angreifer noch Blocker.
//
// **Und der Zug wird neu gestempelt** (CR 302.6): dort hängt das Angreifen
// daran, dass die Kreatur seit Beginn des jüngsten Zuges ihres Beherrschers
// **durchgehend** unter dessen Kontrolle steht. Nach einem Wechsel ist das
// nicht mehr wahr, und sie greift in diesem Zug nicht mehr an.
func (p Partie) Beherrschen(objekt, wer string) (Partie, error) {
	if !p.sitzSteht(wer) {
		return p, fmt.Errorf("%w: den Sitz %q gibt es in dieser Partie nicht",
			ErrNotFound, wer)
	}

	stelle, fehler := p.stelleInZone(
		objekt, ZoneTisch, "beherrschen", "110.2")
	if fehler != nil {
		return p, fehler
	}

	if p.Objekte[stelle].Beherrscher == wer {
		return p, fmt.Errorf(
			"%w: %q beherrscht %q schon (CR 110.2)",
			ErrValidation, wer, objekt)
	}

	objekte := make([]Spielobjekt, len(p.Objekte))
	copy(objekte, p.Objekte)

	objekte[stelle].Beherrscher = wer
	objekte[stelle].Seit = p.Zug.Nummer

	p.Objekte = objekte

	return p.ausDemKampf(objekt), nil
}

// ausDemKampf nimmt ein Permanent aus dem Kampf — CR 506.4.
func (p Partie) ausDemKampf(objekt string) Partie {
	angriffe := make([]Angriff, 0, len(p.Zug.Kampf.Angriffe))

	for _, angriff := range p.Zug.Kampf.Angriffe {
		if angriff.Kreatur != objekt {
			angriffe = append(angriffe, angriff)
		}
	}

	bloecke := make([]Block, 0, len(p.Zug.Kampf.Bloecke))

	for _, block := range p.Zug.Kampf.Bloecke {
		if block.Kreatur != objekt && block.Angreifer != objekt {
			bloecke = append(bloecke, block)
		}
	}

	p.Zug.Kampf.Angriffe = angriffe
	p.Zug.Kampf.Bloecke = bloecke

	return p
}

// Umwandeln dreht eine mehrseitige Karte auf ihre andere Seite — CR 701.27a.
//
// **Nur, wo es eine zweite gibt** — CR 712.9: umwandeln können nur
// Permanente, die eine zweiseitige Karte oder ein zweiseitiger Spielstein
// sind. Eine einseitige umzudrehen wäre eine Handlung ohne Gegenstand, und
// die Meldung sagt es (CR 701.27c).
//
// **Das Permanent bleibt dasselbe Objekt**: CR 701.27a lässt es umdrehen und
// nicht die Zone wechseln — die Marken und der Zustand bleiben, weil
// CR 400.7 gar nicht greift.
//
// **Und es ist etwas anderes als Umdrehen** (CR 701.27b): eine Karte verdeckt
// zu legen ist dieselbe Handbewegung und eine andere Spielhandlung.
func (p Partie) Umwandeln(objekt string) (Partie, error) {
	stelle, fehler := p.stelleInZone(
		objekt, ZoneTisch, "umwandeln", "701.27a")
	if fehler != nil {
		return p, fehler
	}

	seiten := p.seitenzahl(p.Objekte[stelle])
	if seiten < 2 {
		return p, fmt.Errorf(
			"%w: %q hat nur eine Seite und wandelt sich nicht um (CR 712.9)",
			ErrValidation, objekt)
	}

	objekte := make([]Spielobjekt, len(p.Objekte))
	copy(objekte, p.Objekte)

	objekte[stelle].Seite = (objekte[stelle].Seite + 1) % seiten

	p.Objekte = objekte

	return p, nil
}

// seitenzahl gibt, wie viele Seiten die Karte eines Objekts hat.
func (p Partie) seitenzahl(objekt Spielobjekt) int {
	for _, sitz := range p.Sitze {
		if sitz.Kennung != objekt.Eigner {
			continue
		}

		for _, karte := range sitz.Deck.Karten {
			if karte.Kennung == objekt.Karte {
				return len(karte.Seiten)
			}
		}
	}

	return 0
}
