package service

// **Vier Gestalten, die keine Karte brauchen** — Werkstück M12 der
// Regelmaschine
// (#1614).
//
// # Warum sie hier stehen und nicht bei den Handlungen
//
// **Weil sie nicht alle eine Handlung mit Schlüsselwort sind.** CR 701
// führt neunundsechzig Handlungen; ein Spielstein zu erzeugen ist eine davon
// (701.7a) und eine Karte umzuwandeln auch (701.27a) — ein Emblem zu
// erzeugen und die Kontrolle zu wechseln sind es **nicht**: die stehen in
// CR 114 und CR 110. Was alle vier eint, ist die Frage: **welche Gestalt hat
// das, was auf dem Tisch liegt?**

import (
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

const (
	// BefehlSpielstein erzeugt einen Spielstein — CR 111.1, CR 701.7a.
	//
	// **Die Eigenschaften kommen mit**, weil kein Deck sie hergibt: ein
	// Spielstein vertritt keine Karte, und sein Erzeuger nennt sie.
	BefehlSpielstein = "spieler.spielstein"

	// BefehlEmblem erzeugt ein Emblem in der Kommandozone — CR 114.1.
	//
	// **Ohne Eigenschaften**: ein Emblem hat nur Fähigkeiten (CR 114.1),
	// und eine Fähigkeit ist ein Kartentext (`M8`).
	BefehlEmblem = "spieler.emblem"

	// BefehlBeherrschen wechselt die Kontrolle — CR 110.2, CR 613.1b.
	//
	// **Er nennt, wer künftig beherrscht, und nicht nur „der Absender"**:
	// am Tisch schenkt man auch etwas her, und die Regel kennt die
	// Richtung nicht.
	BefehlBeherrschen = "objekt.beherrschen"

	// BefehlUmwandeln dreht eine zweiseitige Karte um — CR 701.27a.
	BefehlUmwandeln = "objekt.umwandeln"

	// BefehlKopieren macht aus einem Objekt die Kopie eines anderen —
	// CR 707.2 (M18,
	// #1619).
	BefehlKopieren = "objekt.kopieren"

	// BefehlHeimholen legt einen Commander in die Kommandozone —
	// CR 903.9a (M19,
	// #1620).
	//
	// **Ein Dürfen und kein Müssen**, und deshalb ein Befehl: CR 903.9a
	// sagt „may", und der Eigner entscheidet.
	BefehlHeimholen = "spieler.heimholen"

	// Die **Bezeichnungen** aus CR 725, CR 726 und CR 728 — Werkstück M20
	// (#1621).
	//
	// **Sie werden gestellt und geschehen nicht.** Was sie vergibt, sind
	// eingebaute ausgelöste Fähigkeiten (CR 725.2, CR 726.2, CR 728.1) —
	// **ohne Quelle**, sagt die CR ausdrücklich —, und ausgelöste
	// Fähigkeiten gibt es erst mit `M9`.
	BefehlMonarch    = "spieler.monarch"
	BefehlInitiative = "spieler.initiative"
	BefehlStrahlung  = "spieler.strahlung"

	// BefehlLegende beantwortet die Frage aus CR 704.5j.
	//
	// **Der einzige Befehl, der eine zustandsbasierte Aktion abschließt.**
	// Die übrigen vierundzwanzig geschehen; diese wartet auf eine Wahl —
	// der Grund steht im Kopf von `internal/domain/legendenregel.go`.
	BefehlLegende = "spieler.legende"

	// BefehlReihung beantwortet die Frage aus CR 603.3b — Werkstück M9
	// (#1611).
	//
	// **Der zweite Befehl, der eine Regel abschließt, statt eine Handlung
	// zu sein.** Wie die Legendenwahl steht er, weil das Regelwerk an
	// dieser Stelle eine **Wahl** verlangt: beherrscht ein Sitz mehrere
	// Fähigkeiten, die zugleich ausgelöst haben, nennt **er** die
	// Reihenfolge, in der sie auf den Stapel kommen.
	BefehlReihung = "spieler.reihung"
)

// istGestaltbefehl sagt, ob dieser Befehl eine der vier Gestalten meint.
func istGestaltbefehl(art string) bool {
	switch art {
	case BefehlSpielstein, BefehlEmblem, BefehlBeherrschen, BefehlUmwandeln,
		BefehlKopieren, BefehlHeimholen, BefehlLegende, BefehlReihung,
		BefehlMonarch, BefehlInitiative, BefehlStrahlung:
		return true
	}

	return false
}

// gestaltbefehl legt die vier aus.
func (s *Sandkasten) gestaltbefehl(
	partie domain.Partie, befehl domain.Spielbefehl,
) ([]domain.Ereignis, domain.Partie, error) {
	switch befehl.Art {
	case BefehlSpielstein:
		return s.spielsteinbefehl(partie, befehl)

	case BefehlEmblem:
		return s.emblembefehl(partie, befehl)

	case BefehlMonarch, BefehlInitiative, BefehlStrahlung:
		return s.bezeichnungsbefehl(partie, befehl)

	case BefehlReihung:
		return s.reihungsbefehl(partie, befehl)
	}

	return s.gestaltwechsel(partie, befehl)
}

// spielsteinbefehl erzeugt einen Spielstein — CR 701.7a.
func (s *Sandkasten) spielsteinbefehl(
	partie domain.Partie, befehl domain.Spielbefehl,
) ([]domain.Ereignis, domain.Partie, error) {
	kennung := neueKennung(partie, befehl, "spielstein")

	felder, fehler := felderAus(befehl, "felder")
	if fehler != nil {
		return nil, domain.Partie{}, fehler
	}

	danach, gescheitert := partie.Spielstein(befehl.Absender, kennung, felder)

	return handlungsereignis(danach, gescheitert, befehl,
		map[string]any{"objekt": kennung})
}

// emblembefehl erzeugt ein Emblem — CR 114.1.
func (s *Sandkasten) emblembefehl(
	partie domain.Partie, befehl domain.Spielbefehl,
) ([]domain.Ereignis, domain.Partie, error) {
	kennung := neueKennung(partie, befehl, "emblem")

	danach, gescheitert := partie.Emblem(befehl.Absender, kennung)

	return handlungsereignis(danach, gescheitert, befehl,
		map[string]any{"objekt": kennung})
}

// reihungsbefehl beantwortet die Frage aus CR 603.3b.
//
// **Die Nutzlast nennt die Fähigkeit und nicht ein Objekt** — eine wartende
// Fähigkeit liegt nirgends ([domain.Partie.Ausgeloest]), und ihre Kennung
// gilt nur, solange sie wartet. Deshalb geht sie **nicht** durch
// [Sandkasten.gestaltwechsel], der ein Objekt auf dem Tisch erwartet.
func (s *Sandkasten) reihungsbefehl(
	partie domain.Partie, befehl domain.Spielbefehl,
) ([]domain.Ereignis, domain.Partie, error) {
	kennung, fehler := zeichenkette(befehl, "kennung")
	if fehler != nil {
		return nil, domain.Partie{}, fehler
	}

	danach, gescheitert := partie.Reihung(befehl.Absender, kennung)

	return handlungsereignis(danach, gescheitert, befehl,
		map[string]any{"kennung": kennung})
}

// gestaltwechsel legt die zwei aus, die ein Objekt ändern.
func (s *Sandkasten) gestaltwechsel(
	partie domain.Partie, befehl domain.Spielbefehl,
) ([]domain.Ereignis, domain.Partie, error) {
	objekt, fehler := zeichenkette(befehl, "objekt")
	if fehler != nil {
		return nil, domain.Partie{}, fehler
	}

	if befehl.Art == BefehlUmwandeln {
		danach, gescheitert := partie.Umwandeln(objekt)

		return handlungsereignis(danach, gescheitert, befehl,
			map[string]any{"objekt": objekt})
	}

	if befehl.Art == BefehlKopieren {
		vorlage, fehlend := zeichenkette(befehl, "vorlage")
		if fehlend != nil {
			return nil, domain.Partie{}, fehlend
		}

		danach, gescheitert := partie.Kopieren(objekt, vorlage)

		return handlungsereignis(danach, gescheitert, befehl,
			map[string]any{"objekt": objekt, "vorlage": vorlage})
	}

	// **Die Legendenwahl und das Heimholen nennen keinen zweiten Sitz** —
	// CR 704.5j lässt den Beherrscher wählen, CR 903.9a den Eigner, und
	// beides ist der Absender.
	if befehl.Art == BefehlLegende {
		danach, gescheitert := partie.Legende(befehl.Absender, objekt)

		return handlungsereignis(danach, gescheitert, befehl,
			map[string]any{"objekt": objekt})
	}

	if befehl.Art == BefehlHeimholen {
		danach, gescheitert := partie.Heimholen(befehl.Absender, objekt)

		return handlungsereignis(danach, gescheitert, befehl,
			map[string]any{"objekt": objekt})
	}

	wer, fehler := zeichenkette(befehl, "wer")
	if fehler != nil {
		return nil, domain.Partie{}, fehler
	}

	danach, gescheitert := partie.Beherrschen(objekt, wer)

	return handlungsereignis(danach, gescheitert, befehl,
		map[string]any{"objekt": objekt, "wer": wer})
}

// zeichenkette liest ein Pflichtfeld der Nutzlast.
//
// **Fehlend und falsch getippt sind derselbe Fall** — derselbe Grund wie bei
// [Sandkasten.objektstelle]: zwei Meldungen für denselben Irrtum wären zwei
// Wege zu derselben Stelle.
func zeichenkette(befehl domain.Spielbefehl, feld string) (string, error) {
	wert, _ := befehl.Nutzlast[feld].(string) //nolint:errcheck // siehe darüber

	if wert == "" {
		return "", fmt.Errorf("%w: %q braucht %q in der Nutzlast",
			domain.ErrValidation, befehl.Art, feld)
	}

	return wert, nil
}

// neueKennung gibt die Kennung eines neuen Objekts — **die genannte, oder
// die nächste freie** (#2480): das Fenster sieht nur Sichtkennungen und kann
// nicht wissen, welche Kennung frei ist; der Kern kennt sie.
func neueKennung(partie domain.Partie, befehl domain.Spielbefehl, wofuer string) string {
	if genannt, _ := befehl.Nutzlast["kennung"].(string); genannt != "" { //nolint:errcheck // ohne Angabe vergibt der Kern
		return genannt
	}

	return partie.FreieKennung(befehl.Absender, wofuer)
}

// felderAus liest die gedruckten Eigenschaften aus der Nutzlast.
//
// **Zeichenketten und nichts anderes.** In einem MSE-Paket steht jedes Feld
// als Text — auch `power` und `toughness`, denn dort steht auch `*`. Wer
// eine Zahl schickte, bekäme aus JSON eine Fließkommazahl, und `2` würde zu
// `2.000000`; das ist ein **Befund** und kein Fehler (§8), und die Meldung
// nennt das Feld.
func felderAus(
	befehl domain.Spielbefehl, feld string,
) (map[string]string, error) {
	roh, steht := befehl.Nutzlast[feld].(map[string]any)
	if !steht || len(roh) == 0 {
		return nil, fmt.Errorf(
			"%w: %q braucht die gedruckten Eigenschaften in %q",
			domain.ErrValidation, befehl.Art, feld)
	}

	felder := make(map[string]string, len(roh))

	for name, wert := range roh {
		text, ist := wert.(string)
		if !ist {
			return nil, fmt.Errorf(
				"%w: in %q steht bei %q %v statt eines Textes",
				domain.ErrValidation, feld, name, wert)
		}

		felder[name] = text
	}

	return felder, nil
}

// bezeichnungsbefehl vergibt eine Bezeichnung — CR 725, CR 726, CR 728.
//
// **Der Absender bekommt sie.** Am Tisch nimmt sich niemand die Krone eines
// anderen; was sie vergibt, ist eine Fähigkeit (CR 725.2), und bis es die
// gibt, sagt der, der sie bekommt, dass er sie bekommt.
func (s *Sandkasten) bezeichnungsbefehl(
	partie domain.Partie, befehl domain.Spielbefehl,
) ([]domain.Ereignis, domain.Partie, error) {
	if befehl.Art == BefehlStrahlung {
		um, fehler := ganzeZahl(befehl, "um")
		if fehler != nil {
			return nil, domain.Partie{}, fehler
		}

		danach, gescheitert := partie.Strahlung(befehl.Absender, um)

		return handlungsereignis(danach, gescheitert, befehl,
			map[string]any{"um": float64(um)})
	}

	vergeben := domain.Partie.Monarchwerden
	if befehl.Art == BefehlInitiative {
		vergeben = domain.Partie.Initiativenehmen
	}

	danach, gescheitert := vergeben(partie, befehl.Absender)

	return handlungsereignis(danach, gescheitert, befehl, nil)
}
