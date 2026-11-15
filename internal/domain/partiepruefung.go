package domain

// **Ob eine Partie so anfangen darf** — Werkstück G1 von Phase 5.5
// (#1458).
//
// # Eine eigene Datei, und der Schnitt folgt der Frage
//
// **Nebenan steht, *was* eine Partie ist** — jedes Feld mit seinem Grund.
// Hier steht, **was daran unmöglich ist**: ein Tisch ohne Spiel, ein Sitz
// ohne Wirt, eine Marke unter null, eine Schicht, die es nicht gibt.
// `partie.go` stand mit beidem bei 405 Zeilen (§2) — die Dauereffekte aus
// `M10` (#1612)
// waren die Zeilen, die darüber hinausgingen.

import (
	"fmt"
	"strings"
)

// Pruefe sagt, ob diese Partie so anfangen darf.
func (p Partie) Pruefe() error {
	if strings.TrimSpace(p.Spiel) == "" {
		return fmt.Errorf(
			"%w: eine Partie gehört zu einem Spiel — sonst gibt es keine "+
				"Regeln, gegen die sich etwas prüfen ließe", ErrValidation)
	}

	if fehler := p.pruefeRegelwerk(); fehler != nil {
		return fehler
	}

	if fehler := p.Regelwerk.Pruefe(); fehler != nil {
		return fehler
	}

	if fehler := p.pruefeSitze(); fehler != nil {
		return fehler
	}

	if fehler := p.pruefeMarken(); fehler != nil {
		return fehler
	}

	if fehler := p.pruefeDauereffekte(); fehler != nil {
		return fehler
	}

	if fehler := p.pruefeZuschauer(); fehler != nil {
		return fehler
	}

	return p.pruefeGroesse()
}

// pruefeGroesse sagt, ob diese Partie noch eine Partiedatei wäre — #2065.
//
// **Dieselbe Grenze wie die Einfuhr** ([PartieaustauschHoechstens]): eine
// Partie, die man am Tisch beginnen kann, muss man auch ausführen und wieder
// einlesen können. Aus zwei Decks mit je 147.720 Karten — `deck aus-set` mit
// der Menge 40 über ein Set mit 3.693 Karten — entstand eine, die das nicht
// mehr war: sie sprengte beim Lesen den Speicher des Treibers, die Liste der
// Partien zeigte einen Panic, und danach hing das Beenden.
//
// **Gemessen wird die Datei und nicht eine Kartenzahl**: was eine Partie
// wiegt, hängt an mehr als ihren Karten, und eine Zahl daneben liefe von der
// Grenze weg, die sie schützen soll.
func (p Partie) pruefeGroesse() error {
	laenge := len(FormePartie(p))
	if laenge <= PartieaustauschHoechstens {
		return nil
	}

	return fmt.Errorf(
		"%w: diese Partie wäre als Datei %d Bytes groß — ausführen und wieder "+
			"einlesen lässt sich höchstens %d. Die Decks sind zusammen zu groß "+
			"für einen Tisch (%d Objekte)",
		ErrValidation, laenge, PartieaustauschHoechstens, len(p.Objekte))
}

// pruefeDauereffekte weist eine Schicht ab, die es nicht gibt — CR 613.1.
//
// **Fremde Daten sind der Normalfall** (§8): eine Partie kommt aus dem
// Bestand, und dort kann eine Fassung stehen, die eine Schicht kannte, die
// diese nicht kennt. Ein Effekt, der in **keiner** Schicht liegt, wirkt nie
// und ließe sich auch nie finden — das ist ein Fehler und kein Befund.
func (p Partie) pruefeDauereffekte() error {
	for _, effekt := range p.Dauereffekte {
		if !istSchicht(effekt.Schicht) {
			return fmt.Errorf(
				"%w: %q ist keine der zehn Schichten aus CR 613.1",
				ErrValidation, effekt.Schicht)
		}
	}

	return nil
}

// pruefeSitze hält die Sitze und den Wirt gegeneinander.
func (p Partie) pruefeSitze() error {
	// **Ein einzelner Sitz ist ein Aufbau und kein Spiel — und genau deshalb
	// erlaubt** (G3,
	// #1460).
	//
	// **CR 100.1** (Fassung 2026-08-19): die Regeln gelten für Spiele mit
	// **zwei oder mehr** Spielern. Ein Goldfish hat einen — er ist damit kein
	// Spiel nach den Regeln, und dieses Programm setzt an ihm auch keine
	// durch; was es überhaupt durchsetzt, sagt die Regelmatrix (G4, #1461).
	//
	// **Hier stand bis zum 2026-09-17 „mindestens zwei"**, mit genau dieser
	// Begründung — „mit einem ist es kein Spiel, sondern ein Aufbau". Der Satz
	// war richtig und die Folgerung falsch: #1460 verlangt „Solo Goldfish",
	// also genau diesen Aufbau. Wer allein übt, will kein Spiel; er will seine
	// Kurve sehen (§15, Punkt 4: keine stillschweigende Änderung — deshalb
	// steht sie hier).
	if len(p.Sitze) == 0 {
		return fmt.Errorf(
			"%w: eine Partie braucht mindestens einen Sitz", ErrValidation)
	}

	gesehen := make(map[string]bool, len(p.Sitze))

	for _, sitz := range p.Sitze {
		if strings.TrimSpace(sitz.Kennung) == "" {
			return fmt.Errorf(
				"%w: ein Sitz ohne Kennung lässt sich nicht ansprechen",
				ErrValidation)
		}

		if gesehen[sitz.Kennung] {
			return fmt.Errorf(
				"%w: die Sitzkennung %q kommt zweimal vor — dann gehörte "+
					"jeder Befehl zu beiden", ErrValidation, sitz.Kennung)
		}

		gesehen[sitz.Kennung] = true

		if strings.TrimSpace(sitz.Name) == "" {
			return fmt.Errorf(
				"%w: der Sitz %q hat keinen Namen — auf dem Tisch steht dann "+
					"nichts", ErrValidation, sitz.Kennung)
		}

		if sitz.Eingeladen != "" && sitz.Kennung == p.Wirt {
			return fmt.Errorf(
				"%w: der Wirt wird nicht eingeladen — sein Deck liegt beim "+
					"Anlegen", ErrValidation)
		}

		// **Wer eingeladen ist, bringt sein Deck mit** (P9b, #2377): bis
		// dahin steht sein Sitz leer da, und nur seiner.
		if sitz.Wartet() {
			continue
		}

		if fehler := sitz.Deck.Pruefe(); fehler != nil {
			return fmt.Errorf("der Sitz %q: %w", sitz.Kennung, fehler)
		}
	}

	if !gesehen[p.Wirt] {
		return fmt.Errorf(
			"%w: der Wirt %q hat keinen Sitz — dann wartet diese Partie auf "+
				"niemanden", ErrValidation, p.Wirt)
	}

	return nil
}

// pruefeZuschauer hält sie von den Sitzkennungen fern.
//
// **Zwei Namensräume wären einer zu viel.** Eine Sichtprüfung fragt „hat
// dieser Ruf einen Sitz?"; ein Zuschauer auf einer Sitzkennung bekäme darauf
// die falsche Antwort — und mit ihr eine fremde Hand.
func (p Partie) pruefeZuschauer() error {
	for _, zuschauer := range p.Zuschauer {
		if strings.TrimSpace(zuschauer.Kennung) == "" {
			return fmt.Errorf(
				"%w: ein Zuschauer ohne Kennung lässt sich nicht ansprechen",
				ErrValidation)
		}

		if p.SitzVon(zuschauer.Kennung) != nil {
			return fmt.Errorf(
				"%w: %q ist zugleich ein Sitz und ein Zuschauer — dann sähe "+
					"eine Sichtprüfung ihn für einen Spieler an",
				ErrValidation, zuschauer.Kennung)
		}
	}

	return nil
}

// pruefeRegelwerk hält Griff und eingefrorenes Profil zusammen.
//
// **Ein Griff, der auf ein anderes Profil zeigt als das eingefrorene, wäre
// eine Auskunft, die zweimal verschieden ausfällt** — das Fenster zeigte die
// eine, die Entscheidung fiele nach der anderen.
func (p Partie) pruefeRegelwerk() error {
	if fehler := p.Profil.Pruefe(); fehler != nil {
		return fehler
	}

	if p.Profil.Kennung != p.Regelwerk.Kennung ||
		p.Profil.Fassung != p.Regelwerk.Fassung {
		return fmt.Errorf(
			"%w: die Partie nennt das Regelwerk %q in der Fassung %d und "+
				"trägt das Profil %q in der Fassung %d",
			ErrValidation, p.Regelwerk.Kennung, p.Regelwerk.Fassung,
			p.Profil.Kennung, p.Profil.Fassung)
	}

	if p.Profil.Spiel != p.Spiel {
		return fmt.Errorf(
			"%w: die Partie spielt %q und trägt das Regelprofil von %q",
			ErrValidation, p.Spiel, p.Profil.Spiel)
	}

	return nil
}
