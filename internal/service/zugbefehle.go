package service

// **Die Befehle, die zum Zug gehören** — Werkstück G9 von Phase 5.5
// (#1592).
//
// # Warum sie Befehle sind und keine Bindungen
//
// **Weil sie damit an allen vier Flächen zugleich dastehen** (§16). Ein
// Spielbefehl geht über `PartieBefehl`, und den gibt es am Fenster, auf der
// Kommandozeile, über MCP und über REST. Eine eigene Bindung je Weg wäre
// vierfache Arbeit für dieselbe Fachregel.
//
// # Und warum fünf und nicht zwei
//
// **Weil ein Sitz erst mit ihnen ein Spieler ist.** Leben, Vergiftungsmarken
// und markierter Schaden stehen seit G9 am Modell (CR 119.1, CR 122.1f,
// CR 120.3) — ohne einen Weg, sie zu ändern, wären sie drei Felder, die
// niemand verstellt, und das ist ein Versprechen, das keiner einlöst (§10).

import (
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

const (
	// BefehlPassen gibt die Priorität weiter — CR 117.3d.
	//
	// **Der `#nosec` steht hier, weil `gosec` in „passen" ein Kennwort
	// sieht** (G101) — es ist der deutsche Name einer Spielhandlung. Und er
	// steht als `#nosec` und nicht als `//nolint`: `gosec` läuft eigenständig
	// und liest die Anweisung an golangci-lint nicht (§4).
	BefehlPassen = "zug.passen" // #nosec G101 -- eine Spielhandlung, kein Kennwort

	// BefehlAbwerfen erfüllt einen Schritt der Abwurfpflicht — CR 514.1.
	BefehlAbwerfen = "zug.abwerfen"

	// BefehlLeben ändert den Lebensstand eines Sitzes — CR 119.
	//
	// **Ändern und nicht setzen.** „Drei weniger" überlebt eine Antwort, die
	// sich mit einer anderen kreuzt; „siebzehn" überschreibt sie (§7,
	// Regel 2).
	BefehlLeben = "spieler.leben"

	// BefehlVergiftung ändert die Vergiftungsmarken — CR 122.1f.
	BefehlVergiftung = "spieler.vergiftung"

	// BefehlSchaden markiert Schaden an einem Objekt — CR 120.3.
	BefehlSchaden = "objekt.schaden"

	// BefehlAbkuerzen reicht durch, bis etwas geschieht — CR 732 (M1,
	// #1603).
	//
	// **Er spart zwanzig Passen je Zug** — die Zahl steht seit G9 als Grenze
	// in der Regelmatrix.
	BefehlAbkuerzen = "zug.abkuerzen"

	// BefehlZugende beendet den Zug vorzeitig — CR 724.1 (M1,
	// #1603).
	//
	// **Alles auf dem Stapel wird exiliert**, dann geht es geradewegs in den
	// Aufräumschritt.
	BefehlZugende = "zug.beenden"

	// BefehlMana legt Mana in den Vorrat eines Sitzes — CR 106.4 (M3,
	// #1605).
	//
	// **Er ist Buchführung wie [BefehlLeben]**, und aus demselben Grund: es
	// gibt keine Quelle, die von selbst Mana erzeugt, außer dem Grundland
	// ([BefehlLandmana]). Wer eine Manafähigkeit hat, die auf seiner Karte
	// **steht**, braucht `M8` — bis dahin trägt er sie hier nach.
	BefehlMana = "spieler.mana"

	// BefehlLandmana tappt ein Grundland für sein Mana — CR 305.6 (M3).
	//
	// **Anders als [BefehlMana] braucht er Priorität**: CR 605.3a erlaubt
	// eine Manafähigkeit, wenn jemand Priorität hat oder gerade bezahlt. Ein
	// Bezahlen gibt es ohne `M4` nicht — bleibt die Priorität.
	BefehlLandmana = "objekt.mana"

	// BefehlWirken kündigt einen Zauber an und legt ihn auf den Stapel —
	// CR 601.2 (M4,
	// #1606).
	//
	// **Er braucht Priorität** (CR 117.1) — im Gegensatz zu [BefehlMana]
	// daneben ist er keine Buchführung, sondern eine Handlung im Spiel.
	BefehlWirken = "objekt.wirken"

	// BefehlLandspielen legt ein Land aus der Hand auf das Schlachtfeld —
	// CR 116.2a (M4).
	//
	// **Eine Sonderhandlung**: sie geht nicht über den Stapel (CR 116.1),
	// und niemand kann darauf antworten.
	BefehlLandspielen = "objekt.land"

	// BefehlAngreifer erklärt die Angreifer — CR 508.1 (M6,
	// #1608).
	//
	// **Er braucht keine Priorität**: CR 508.1 ist eine **zugbasierte
	// Handlung** mit einer Wahl darin — der aktive Spieler erklärt, und
	// **danach** bekommt er Priorität (CR 508.2). Wer sie vorher verlangte,
	// drehte die Reihenfolge des Regelwerks um.
	BefehlAngreifer = "kampf.angreifer"

	// BefehlBlocker erklärt die Blocker — CR 509.1 (M6).
	//
	// **Ihn stellt der Verteidiger**, nicht der aktive Spieler — und auch er
	// braucht keine Priorität (CR 509.2 gibt sie danach).
	BefehlBlocker = "kampf.blocker"

	// BefehlAufgeben verlässt die Partie — CR 104.3a (M17,
	// #1618).
	//
	// **Jederzeit**, sagt die Regel: aufgeben braucht keine Priorität und
	// wartet auf keinen Abschnitt. Wer aufgibt, ist sofort draußen.
	BefehlAufgeben = "zug.aufgeben"
)

// Zugbefehle gibt die zwölf, die nicht über [Sandkasten.wirken] gehen.
func Zugbefehle() []string {
	return []string{
		BefehlPassen, BefehlAbkuerzen, BefehlZugende, BefehlAbwerfen,
		BefehlLeben, BefehlVergiftung, BefehlSchaden,
		BefehlMana, BefehlLandmana,
		BefehlWirken, BefehlLandspielen,
		BefehlAngreifer, BefehlBlocker,
		BefehlZerstoeren, BefehlOpfern, BefehlExilieren,
		BefehlNeutralisieren, BefehlKaempfen,
		BefehlMahlen, BefehlHellsehen, BefehlUeberwachen,
		BefehlMuenzwurf, BefehlWuerfeln,
		BefehlSpielstein, BefehlEmblem,
		BefehlBeherrschen, BefehlUmwandeln, BefehlKopieren,
		BefehlHeimholen, BefehlLegende, BefehlReihung,
		BefehlMonarch, BefehlInitiative, BefehlStrahlung,
		BefehlAufgeben,
	}
}

// brauchtPrioritaet sagt, ob dieser Befehl an der Reihe sein muss.
//
// # Die Trennung, und warum sie dort verläuft
//
// **Was am Tisch geschieht, braucht Priorität** — eine Karte umlegen, tappen,
// mischen, hineinsehen. Das sind Handlungen im Spiel, und CR 117.1 gibt sie
// dem, der gerade dran ist.
//
// **Was aufgeschrieben wird, braucht keine.** Der Lebensstand, die
// Vergiftungsmarken und der markierte Schaden sind in diesem Programm
// Buchführung: dieses Werkstück baut keine Quelle, die sie von selbst ändert,
// und ein Mensch, der seinen Stand nachträgt, wartet dafür nicht auf seinen
// Zug.
//
// **Und der Abwurf steht dazwischen**: er geschieht im Aufräumschritt, in dem
// niemand Priorität hat (CR 514.3) — er braucht keine und ist trotzdem nur
// dem erlaubt, der am Zug ist. Das prüft [domain.Partie.Abwerfen].
//
// **Aufgeben braucht sie auch nicht**, und zwar aus einem anderen Grund:
// CR 104.3a sagt ausdrücklich „jederzeit". Wer nicht mehr mag, wartet nicht
// auf seinen Zug.
func brauchtPrioritaet(art string) bool {
	switch art {
	case BefehlAbwerfen, BefehlLeben, BefehlVergiftung, BefehlSchaden,
		BefehlMana, BefehlAngreifer, BefehlBlocker,
		// **Ein Wurf braucht keine Priorität** (M15, #1617): er ändert
		// nichts am Tisch. Was er bewirkt, steht auf einer Karte — und die
		// zu wirken braucht sie sehr wohl.
		BefehlMuenzwurf, BefehlWuerfeln,
		// **Die Legendenwahl braucht keine Priorität** (M12, #1614): eine
		// zustandsbasierte Aktion wird abgearbeitet, **bevor** jemand
		// Priorität bekommt (CR 704.3). Wer auf seinen Zug warten müsste,
		// hielte die Partie an — und CR 704.5j trifft jeden Sitz.
		BefehlLegende,

		// **Und die Reihung erst recht nicht** (M9, #1611): solange sie
		// offen ist, hat **niemand** Priorität — CR 603.3 legt die
		// Fähigkeiten, bevor sie das nächste Mal jemand bekäme. Ein Befehl,
		// der sie verlangte, wäre nie ausführbar.
		BefehlReihung,

		// **Das Heimholen braucht keine Priorität** (M19, #1620): CR 903.9a
		// hängt an einer zustandsbasierten Prüfung, und die läuft, bevor
		// jemand Priorität bekommt (CR 704.3).
		BefehlHeimholen,

		// **Die Bezeichnungen brauchen keine Priorität** (M20, #1621):
		// was sie vergibt, sind eingebaute ausgelöste Fähigkeiten
		// (CR 725.2, CR 726.2, CR 728.1), und die laufen ohne Priorität.
		// Bis es sie gibt, ist das Vergeben Buchführung.
		BefehlMonarch, BefehlInitiative, BefehlStrahlung,
		BefehlAufgeben,

		// **Der Beginn braucht keine Priorität** (#1839): er liegt vor dem
		// ersten Zug (CR 103.8 steht nach CR 103.5), und dort hat sie
		// niemand.
		BefehlStarthand, BefehlMulligan, BefehlUnterlegen, BefehlBehalten:
		return false
	}

	return true
}

// istZugbefehl sagt, ob dieser Befehl den Zugstand oder einen Spieler meint.
func istZugbefehl(art string) bool {
	return art == BefehlPassen || art == BefehlAbkuerzen ||
		art == BefehlZugende || art == BefehlAbwerfen ||
		art == BefehlLeben || art == BefehlVergiftung ||
		art == BefehlMana || art == BefehlAufgeben
}

// zugbefehl legt die sechs aus, die kein Objekt meinen.
func (s *Sandkasten) zugbefehl(
	partie domain.Partie, befehl domain.Spielbefehl,
) ([]domain.Ereignis, domain.Partie, error) {
	switch befehl.Art {
	case BefehlPassen:
		danach, fehler := partie.Passen(befehl.Absender)

		return ereignisAus(danach, fehler, befehl)

	case BefehlAbkuerzen:
		danach, fehler := partie.Abkuerzen(befehl.Absender)

		return ereignisAus(danach, fehler, befehl)

	// **Der Zug endet, und niemand fragt nach Priorität** — CR 724.1c sagt
	// ausdrücklich, dass dabei keine vergeben wird. Wer den Befehl stellen
	// darf, entscheidet `brauchtPrioritaet` daneben.
	case BefehlZugende:
		return ereignisAus(partie.ZugBeenden(), nil, befehl)

	case BefehlAbwerfen:
		objekt, _ := befehl.Nutzlast["objekt"].(string) //nolint:errcheck // siehe [Sandkasten.objektstelle]

		danach, fehler := partie.Abwerfen(befehl.Absender, objekt)

		return ereignisAus(danach, fehler, befehl)

	case BefehlAufgeben:
		return s.aufgeben(partie, befehl)

	case BefehlMana:
		return s.manabefehl(partie, befehl)
	}

	return s.markenbefehl(partie, befehl)
}

// markenbefehl ändert Leben oder Vergiftungsmarken eines Sitzes.
//
// **Um wie viel, sagt der Befehl** — und der Sitz ist der Absender. Wer den
// Stand eines anderen änderte, änderte etwas, das nicht ihm gehört; wer
// einem anderen etwas zufügt, sagt es ihm, und der trägt es ein.
func (s *Sandkasten) markenbefehl(
	partie domain.Partie, befehl domain.Spielbefehl,
) ([]domain.Ereignis, domain.Partie, error) {
	um, fehler := ganzeZahl(befehl, "um")
	if fehler != nil {
		return nil, domain.Partie{}, fehler
	}

	danach := s.kopie(partie)

	stelle := -1

	for lauf := range danach.Sitze {
		if danach.Sitze[lauf].Kennung == befehl.Absender {
			stelle = lauf
		}
	}

	if stelle < 0 {
		return nil, domain.Partie{}, fmt.Errorf(
			"%w: %q sitzt nicht an diesem Tisch", domain.ErrValidation,
			befehl.Absender)
	}

	if befehl.Art == BefehlLeben {
		danach.Sitze[stelle].Leben += um
	} else {
		danach.Sitze[stelle].Vergiftung += um
	}

	if fehler := danach.Pruefe(); fehler != nil {
		return nil, domain.Partie{}, fehler
	}

	ereignis := domain.Ereignis{
		Art: befehl.Art,
		Oeffentlich: map[string]any{
			"sitz": befehl.Absender, "um": float64(um),
			"stand": standNach(danach.Sitze[stelle], befehl.Art),
		},
	}

	// **Was daraus folgt, prüft [Sandkasten.Auslegen]** — CR 704.3, und
	// zwar für **jeden** Befehl. Bis `M5`
	// (#1607)
	// stand die Prüfung hier, also an genau einem Weg; ein neuer Befehl
	// hätte sie vergessen können.
	return []domain.Ereignis{ereignis}, danach, nil
}

// schadenbefehl markiert Schaden an einem Objekt — CR 120.3.
//
// **Er liegt am Objekt und bleibt bis zum Aufräumschritt** (CR 514.2), und
// genau deshalb ist er kein Zähler: eine Marke überlebt den Zug.
func (s *Sandkasten) schadenbefehl(
	danach *domain.Partie, stelle int, befehl domain.Spielbefehl,
) (map[string]any, error) {
	um, fehler := ganzeZahl(befehl, "um")
	if fehler != nil {
		return nil, fehler
	}

	danach.Objekte[stelle].Schaden += um

	return map[string]any{
		"um":      float64(um),
		"schaden": float64(danach.Objekte[stelle].Schaden),
	}, nil
}

// standNach gibt den Stand, den dieser Befehl geändert hat.
func standNach(sitz domain.Sitz, art string) float64 {
	if art == BefehlLeben {
		return float64(sitz.Leben)
	}

	return float64(sitz.Vergiftung)
}

// ganzeZahl liest eine Zahl aus der Nutzlast.
//
// **Über `float64`, weil JSON keine ganzen Zahlen kennt** — dieselbe Stelle
// wie beim Zähler, und derselbe Grund.
func ganzeZahl(befehl domain.Spielbefehl, feld string) (int, error) {
	roh, da := befehl.Nutzlast[feld]
	if !da {
		return 0, fmt.Errorf(
			"%w: der Befehl %q nennt kein %q", domain.ErrValidation,
			befehl.Art, feld)
	}

	zahl, passt := roh.(float64)
	if !passt {
		return 0, fmt.Errorf(
			"%w: %q ist keine Zahl", domain.ErrValidation, feld)
	}

	return int(zahl), nil
}

// aufgeben lässt einen Sitz die Partie verlassen — CR 104.3a.
//
// **Wer nicht am Tisch sitzt, gibt nicht auf.** Ein Zuschauer hat nichts
// aufzugeben, und ein Sitz, den es nicht gibt, ist ein Irrtum und kein Wunsch
// (§8) — [domain.Partie.Scheidetaus] ließe ihn still durchgehen, und der
// Absender bekäme ein „geschehen" für nichts.
func (s *Sandkasten) aufgeben(
	partie domain.Partie, befehl domain.Spielbefehl,
) ([]domain.Ereignis, domain.Partie, error) {
	if partie.SitzVon(befehl.Absender) == nil {
		return nil, domain.Partie{}, fmt.Errorf(
			"%w: %q sitzt nicht an diesem Tisch", domain.ErrValidation,
			befehl.Absender)
	}

	danach := partie.Scheidetaus(befehl.Absender, domain.GrundAufgabe)

	return []domain.Ereignis{{
		Art: befehl.Art,
		Oeffentlich: map[string]any{
			"sitz":          befehl.Absender,
			"grund":         string(domain.GrundAufgabe),
			"sieger":        danach.Ausgang.Sieger,
			"unentschieden": danach.Ausgang.Unentschieden,
			"stand":         string(danach.Stand),
		},
	}}, danach, nil
}
