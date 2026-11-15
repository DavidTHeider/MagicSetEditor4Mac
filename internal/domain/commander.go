package domain

// **Die Commandervariante** — Werkstück M19 der Regelmaschine
// (#1620).
//
// # Warum sie die benannte Ausnahme ist
//
// **Weil #1593 die Varianten aus CR 900ff. ausgrenzt und diese eine
// ausdrücklich aufnimmt.** Sie ist die meistgespielte, und der Bestand kennt
// sie längst: CR 903.5a/b/e stehen seit Phase 4 in den Formatgrößen.
//
// # Vier Regeln, und jede hat ihre Nummer
//
//   - **Vierzig Leben** (CR 903.7). Nicht CR 103.4 — das ist der Grundfall
//     mit zwanzig, und die Nummer stand so in einem früheren Entwurf.
//   - **Die Steuer** (CR 903.8): zwei generisches Mana mehr je vorherigem
//     Wirken aus der Kommandozone.
//   - **Einundzwanzig Kampfschaden** desselben Commanders (CR 903.10a) —
//     eine zustandsbasierte Aktion.
//   - **Der Weg zurück** (CR 903.9a): der Eigner **darf** ihn aus Friedhof
//     oder Exil in die Kommandozone legen. Ein Dürfen und kein Müssen, und
//     damit eine **Wahl** — dieselbe Gestalt wie die Legendenregel.

import "fmt"

// Commanderfassung ist die Fassung des Commanderprofils.
//
// **Eine eigene Zahl neben [Grundprofilfassung]**: zwei Profile mit
// verschiedenen Zusagen, und jede hat ihre eigene Geschichte.
const Commanderfassung = 1

// Commanderleben ist, womit ein Commanderspieler anfängt — CR 903.7.
const Commanderleben = 40

// Commanderschadensgrenze ist, ab wann derselbe Commander tötet —
// CR 903.10a.
//
// **Kampfschaden**, und zwar über die **ganze Partie** gezählt und nicht je
// Zug — deshalb steht die Zahl am Sitz und nicht am Zugstand.
const Commanderschadensgrenze = 21

// Commandersteuer gibt, wie viel generisches Mana ein Wirken aus der
// Kommandozone zusätzlich kostet — CR 903.8.
//
// **Zwei je vorherigem Mal**, und das erste ist umsonst. Die Zahl zählt
// **Wirken aus der Kommandozone** und nicht jedes Wirken: wer seinen
// Commander aus der Hand wirkt, zahlt nichts dazu.
func Commandersteuer(schonGewirkt int) int {
	return 2 * schonGewirkt
}

// Commanderkennung ist der Name des Commanderprofils.
const Commanderkennung = "commander"

// ProfilZu gibt das Regelprofil zu einer Kennung.
//
// # Warum eine Stelle und kein `switch` beim Aufrufer
//
// **Weil ein Profil eingefroren wird** (G4,
// #1461):
// es entsteht **einmal**, beim Aufbau der Partie, und jede zweite Stelle,
// die es wählte, liefe mit dieser auseinander.
//
// **Was es nicht kennt, ist der Sandkasten** — und das ist die ehrliche
// Vorgabe: ein Profil, das niemand kennt, setzt nichts durch, und das sagt
// die Matrix. Wer eine dritte Variante baut, trägt sie hier ein.
func ProfilZu(kennung, spiel string) Regelprofil {
	if kennung == Commanderkennung {
		return Commanderprofil(spiel)
	}

	return Grundprofil(spiel)
}

// Commanderprofil gibt das Regelprofil einer Commanderpartie — CR 903.
func Commanderprofil(spiel string) Regelprofil {
	grund := Grundprofil(spiel)

	return Regelprofil{
		Kennung: Commanderkennung,
		Fassung: Commanderfassung,
		Spiel:   spiel,
		Wirksam: Grundprofilquelle,
		Quelle: Regelquelle{
			Dokument: DokumentCR,
			Fassung:  Grundprofilquelle,
			Nummern:  []string{"903", "408"},
		},

		// **Zwei bis vier** — dieselbe Spanne wie im Grundprofil. CR 903.2
		// nennt drei bis fünf als das Übliche und verbietet zwei nicht.
		Spieler: Spielerzahl{Mindestens: 1, Hoechstens: 4},

		// **Vierzig** — CR 903.7.
		Startleben: Commanderleben,

		// **Dieselbe Matrix wie das Grundprofil, plus der Commanderbereich**
		// — was der Kern durchsetzt, hängt an den Werkstücken und nicht an
		// der Variante.
		//
		// **Der Eintrag wird ersetzt und nicht angehängt**: das Grundprofil
		// füllt jeden Bereich ohne Aussage mit `manuell`, und
		// [Regelmatrix.Eintrag] nimmt den **ersten**. Ein angehängter
		// stünde dahinter und wäre nie zu sehen — gefunden am 2026-09-17
		// an einem roten Lauf, der „im Commanderprofil steht manuell"
		// sagte.
		Matrix: mitCommander(grund.Matrix),

		ErlaubteHausregeln: Hausregelarten(),
	}
}

// mitCommander ersetzt den Commanderbereich einer Matrix.
func mitCommander(matrix Regelmatrix) Regelmatrix {
	eintraege := make([]Matrixeintrag, 0, len(matrix.Eintraege))

	for _, eintrag := range matrix.Eintraege {
		if eintrag.Bereich != BereichCommander {
			eintraege = append(eintraege, eintrag)
		}
	}

	return Regelmatrix{Eintraege: append(eintraege, commandereintrag())}
}

// commandereintrag gibt, was die Variante durchsetzt.
func commandereintrag() Matrixeintrag {
	return Matrixeintrag{
		Bereich: BereichCommander,
		Grad:    UnterstuetzungAutomatisch,
		Quelle:  quelle("903", "408"),
		Grenzen: []string{
			// **Wer der Commander ist, sagt der Sitz** — CR 903.3 nennt
			// eine legendäre Kreaturenkarte; ob sie das ist, prüft niemand,
			// weil es eine Frage an den Deckbau ist (CR 903.5).
			"ob die genannte Karte legendär ist, prüft niemand (CR 903.3)",

			// **Die Farbidentität ist eine Deckbauregel** — CR 903.4.
			"die Farbidentität des Decks prüft niemand (CR 903.4)",

			// **Und was ein Commander tut, steht auf ihm** — CR 903.6
			// nennt die Fähigkeit, die manche haben.
			"kein Commander hat eine Fähigkeit (CR 903.6, M8)",

			// **Der Weg zurück ist ein Dürfen** — CR 903.9a, und er wird
			// gestellt statt zu geschehen.
			"der Weg in die Kommandozone wird gestellt, nicht genommen " +
				"(CR 903.9a)",
		},
	}
}

// Heimholen legt einen Commander aus Friedhof oder Exil in die
// Kommandozone — CR 903.9a.
//
// **Ein Dürfen und kein Müssen.** CR 903.9a sagt „may", und deshalb ist es
// ein Befehl und keine zustandsbasierte Aktion, die von selbst liefe.
//
// **Der Eigner entscheidet** — nicht, wer es beherrscht hat.
func (p Partie) Heimholen(wer, objekt string) (Partie, error) {
	stelle, fehler := p.objektstelle(objekt)
	if fehler != nil {
		return p, fehler
	}

	if !p.istCommander(p.Objekte[stelle]) {
		return p, fmt.Errorf(
			"%w: %q ist kein Commander (CR 903.3)", ErrValidation, objekt)
	}

	if p.Objekte[stelle].Eigner != wer {
		return p, fmt.Errorf(
			"%w: %q gehört %q nicht (CR 903.9a)",
			ErrValidation, objekt, wer)
	}

	if zone := p.Objekte[stelle].Zone; zone != ZoneFriedhof &&
		zone != ZoneExil {
		return p, fmt.Errorf(
			"%w: %q liegt in %q — CR 903.9a nennt Friedhof und Exil",
			ErrValidation, objekt, zone)
	}

	p = p.zonenwechsel(stelle, ZoneKommando)
	p.Objekte[stelle].Beherrscher = p.Objekte[stelle].Eigner

	return p, nil
}

// istCommander sagt, ob dieses Objekt der Commander seines Eigners ist —
// CR 903.3.
//
// **Die Bezeichnung ist kein Merkmal** (CR 903.3): sie steht am Sitz und
// nicht an der Karte. Deshalb wird der Sitz gefragt und nicht die Typzeile.
func (p Partie) istCommander(objekt Spielobjekt) bool {
	if objekt.Karte == "" {
		return false
	}

	for _, sitz := range p.Sitze {
		if sitz.Kennung == objekt.Eigner {
			return sitz.Commander == objekt.Karte
		}
	}

	return false
}

// commanderschlag sagt, ob ein Sitz an Commanderschaden ausscheidet —
// CR 903.10a.
//
// **Desselben Commanders**, und deshalb je Kennung gezählt: zwei Commander
// mit je zwanzig sind vierzig und trotzdem keine einundzwanzig.
func commanderschlag(sitz Sitz, profil Regelprofil) bool {
	if profil.Kennung != Commanderkennung {
		return false
	}

	for _, schaden := range sitz.Commanderschaden {
		if schaden >= Commanderschadensgrenze {
			return true
		}
	}

	return false
}

// mitCommandersteuer legt die Steuer auf die Kosten — CR 903.8.
//
// **Nur aus der Kommandozone**: CR 903.8 nennt sie ausdrücklich. Wer seinen
// Commander aus der Hand wirkt, zahlt die gedruckten Kosten.
//
// **Und nur in einer Commanderpartie** — sonst gibt es keinen Commander.
func (p Partie) mitCommandersteuer(
	kosten Manakosten, wer string, objekt Spielobjekt,
) Manakosten {
	um := p.commandersteuerFuer(wer, objekt)
	if um == 0 {
		return kosten
	}

	// **Ein zusätzliches Symbol und keine geänderte Zahl.** Ein `{2}{R}`
	// mit vier Steuer wird `{2}{R}{4}` und nicht `{6}{R}`: die gedruckten
	// Kosten bleiben stehen, und was dazukommt, steht daneben. CR 601.2f
	// rechnet beides zusammen, und [Manavorrat.Bezahlbar] tut dasselbe.
	kosten.Symbole = append(append([]Manasymbol{}, kosten.Symbole...),
		Manasymbol{
			Roh:       fmt.Sprintf("{%d}", um),
			Zahl:      true,
			Generisch: um,
		})

	return kosten
}

// steuerZaehlen merkt sich ein Wirken aus der Kommandozone — CR 903.8.
func (p Partie) steuerZaehlen(wer string, objekt Spielobjekt) Partie {
	if !p.ausDerKommandozone(wer, objekt) {
		return p
	}

	sitze := make([]Sitz, len(p.Sitze))
	copy(sitze, p.Sitze)

	for stelle := range sitze {
		if sitze[stelle].Kennung == wer {
			sitze[stelle].Commandersteuer++
		}
	}

	p.Sitze = sitze

	return p
}

// commandersteuerFuer gibt die Steuer dieses Wirkens — CR 903.8.
//
// **Null heißt: sie gilt nicht**, und das ist der Normalfall — das erste Mal
// kostet nichts dazu, und außerhalb einer Commanderpartie gibt es sie gar
// nicht.
//
// **Eine Schleife und kein zweiter Griff an den Sitz**: `istCommander` hat
// ihn schon gefunden, und ein zweiter Blick brächte einen Zweig für „es gibt
// ihn doch nicht", den kein Weg betritt (§10).
func (p Partie) commandersteuerFuer(wer string, objekt Spielobjekt) int {
	if p.Profil.Kennung != Commanderkennung || objekt.Zone != ZoneKommando ||
		objekt.Eigner != wer || objekt.Karte == "" {
		return 0
	}

	for _, sitz := range p.Sitze {
		if sitz.Kennung == wer && sitz.Commander == objekt.Karte {
			return Commandersteuer(sitz.Commandersteuer)
		}
	}

	// **Was sonst in der Kommandozone liegt, kostet nichts** — ein Emblem
	// (CR 114.1) liegt dort und wird nie gewirkt, und eine Karte, die nicht
	// der Commander dieses Sitzes ist, zahlt seine Steuer nicht.
	return 0
}

// ausDerKommandozone sagt, ob dieses Wirken die Steuer auslöst — CR 903.8.
func (p Partie) ausDerKommandozone(wer string, objekt Spielobjekt) bool {
	return p.Profil.Kennung == Commanderkennung &&
		objekt.Zone == ZoneKommando &&
		objekt.Eigner == wer &&
		p.istCommander(objekt)
}

// commanderschadenZaehlen merkt sich Kampfschaden je Commander — CR 903.10a.
//
// **Nur in einer Commanderpartie und nur von einem Commander.** Sonst
// wüchse an jedem Sitz eine Tabelle, die niemand liest.
func (p Partie) commanderschadenZaehlen(sitz *Sitz, eine zuteilung) {
	if p.Profil.Kennung != Commanderkennung || eine.Von == "" {
		return
	}

	// **Eine Schleife und kein Nachschlagen**: eine Zuteilung entsteht nur,
	// wo eine Kreatur Stärke hat — also gibt es sie. Ein Zweig für „es gibt
	// sie nicht" wäre einer, den kein Weg betritt (§10).
	for _, quelle := range p.Objekte {
		if quelle.Kennung != eine.Von || !p.istCommander(quelle) {
			continue
		}

		if sitz.Commanderschaden == nil {
			sitz.Commanderschaden = map[string]int{}
		}

		sitz.Commanderschaden[eine.Von] += eine.Menge
	}
}
