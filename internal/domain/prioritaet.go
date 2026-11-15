package domain

// **Wer wann handeln darf** — Werkstück G9 von Phase 5.5
// (#1592).
//
// # Die Gestalt, die ohne Stapel gilt
//
// **CR 500.2** (Fassung 2026-08-19): ein Abschnitt endet, wenn der Stapel
// leer ist **und** alle nacheinander passen. Einen Stapel gibt es hier nicht
// (CR 405 ist Phase 5.8) — er ist damit **immer** leer, und „alle passen
// nacheinander" ist das einzige, was einen Abschnitt beendet. Das ist keine
// Vereinfachung, sondern genau der Fall, den die Regel beschreibt.
//
// # Was hier ausdrücklich fehlt
//
// **Abkürzungen** (CR 732). Ein Zug hat zehn Abschnitte, die Priorität
// vergeben, und bei zwei Sitzen sind das **zwanzig** Passen, in denen nichts
// geschieht. Für Menschen ist das ohne Abkürzung keine Bedienung — sie
// braucht ein eigenes Werkstück, und bis dahin steht das als Grenze in der
// Regelmatrix.

import "fmt"

// Zugstand ist, wo eine Partie im Zug steht — CR 102.1, CR 500.1.
type Zugstand struct {
	// Nummer ist der wievielte Zug — eins beim ersten.
	//
	// **Null heißt: es geht noch nicht zugweise zu.** Eine Partie aus G1 bis
	// G8 ist ein Sandkasten ohne Zugbegriff, und die soll weiterlaufen.
	Nummer int

	// Aktiv ist die Sitzkennung dessen, dessen Zug es ist — CR 102.1.
	Aktiv string

	// Abschnitt ist die Stelle im Zug.
	Abschnitt Abschnitt

	// Prioritaet ist, wer gerade handeln darf — **leer heißt: niemand**
	// (CR 502.4, CR 514.3).
	Prioritaet string

	// Starter ist, wer den ersten Zug genommen hat — CR 103.1 (Fassung
	// 2026-09-25), und
	// CR 103.8a hängt daran: **er** überspringt sein erstes Ziehen.
	Starter string

	// Abwurf ist, wie viele Karten der aktive Spieler noch abwerfen muss —
	// CR 514.1.
	//
	// **Eine Pflicht und keine Handlung**: welche Karten weggehen, entscheidet
	// er, und bis er es getan hat, wartet der Aufräumschritt.
	Abwurf int

	// Kampf ist, wer wen angreift und wer wen blockt — CR 508, CR 509
	// (Fassung 2026-08-19), Werkstück M6
	// (#1608).
	//
	// **Er geht mit dem Zug auf null** — CR 511.3: nach dem Kampfende ist
	// niemand mehr Angreifer oder Blocker.
	Kampf Kampfstand

	// Zauber zählt, wie viele Zauber der aktive Spieler in diesem Zug
	// gewirkt hat — CR 731.2a, CR 731.2b (M20,
	// #1621).
	//
	// **Eine Zahl je Zug und keine Tabelle je Sitz**: CR 731.2a fragt nach
	// dem **aktiven Spieler** des vorigen Zuges, und einen anderen gibt es
	// je Zug nicht. **Sie geht mit dem Zug auf null**, denn gefragt ist
	// „während dieses Zuges".
	Zauber int

	// Laender ist, wie viele Länder der aktive Spieler in diesem Zug schon
	// gespielt hat — CR 305.2 (Fassung 2026-08-19), Werkstück M4
	// (#1606).
	//
	// # Eine Zahl und keine Tabelle je Sitz
	//
	// **Weil nur der aktive Spieler eines spielen kann** — CR 116.2a sagt
	// „in einer Hauptphase **seines** Zuges". Eine Tabelle je Sitz führte
	// elf Nullen mit, von denen keine je etwas anderes würde.
	//
	// **Sie geht mit dem Zug auf null.** CR 305.2a zählt, was „in diesem
	// Zug" gespielt wurde; ein Zähler, der bliebe, ließe niemanden je ein
	// zweites Land spielen.
	Laender int

	// Gepasst sind die, die seit der letzten Handlung gepasst haben, in der
	// Reihenfolge, in der sie es taten — CR 117.4.
	//
	// **Eine Liste und keine Zahl**: „alle nacheinander" ist eine Aussage
	// über **wer**, nicht über wie viele. Bei drei Sitzen dürfen nicht
	// zweimal dieselben zwei passen und den Abschnitt beenden.
	Gepasst []string
}

// ZugBeginnen fängt den ersten Zug an — beim Enttappen, ohne Priorität.
//
// **Ohne Priorität, weil CR 502.4 es sagt**: im Enttappen bekommt sie
// niemand. Wer hier eine vergäbe, gäbe sie an einer Stelle, an der die Regel
// sie ausdrücklich verweigert.
func (p Partie) ZugBeginnen(aktiv string) Partie {
	p.Zug = Zugstand{
		Nummer: 1, Aktiv: aktiv, Starter: aktiv, Abschnitt: zugfolge[0],
	}

	return p.abschnittBetreten()
}

// Weiter geht in den nächsten Abschnitt — und in den nächsten Zug, wenn der
// Zug zu Ende ist.
//
// **Der nächste Sitz wird aktiv** (CR 102.1): reihum in der Reihenfolge der
// Sitze. Eine andere Reihenfolge gäbe es erst mit CR 103.1, und die gehört
// zum Anfang der Partie und nicht hierher.
func (p Partie) Weiter() Partie {
	return p.abschnittVerlassen().abschnittBetreten()
}

// abschnittVerlassen beendet den laufenden Abschnitt und stellt den nächsten.
//
// # Warum es diese Stelle gibt
//
// **Weil sie zweimal dastand** — hier und in der Schleife von
// [Partie.abschnittBetreten] —, und weil ab `M3` etwas hinzukommt, das an
// **beiden** geschehen muss: der Manavorrat leert sich, sobald ein Abschnitt
// endet (CR 500.5). Eine Zusage an zwei Stellen läuft auseinander, und die
// zweite fällt niemandem auf (§10).
func (p Partie) abschnittVerlassen() Partie {
	// **Ein zweiter Schadensschritt, wo jemand Erstschlag hat** — CR 510.4
	// (M6, #1608).
	//
	// **Er ist derselbe Schritt noch einmal**, und die Zugfolge kennt ihn
	// nicht: CR 510.4 setzt ihn **anstelle** des Übergangs ins Kampfende
	// ein. Hier steht er deshalb als Abschnitt, der sich nicht ändert.
	if p.Zug.Abschnitt.Schritt == SchrittKampfschaden &&
		p.zweiterSchadensschritt() {
		p.Zug.Gepasst = nil

		return p
	}

	// **Beim Verlassen und nicht beim Betreten** — CR 500.5 knüpft es an das
	// **Ende** eines Abschnitts. Der Unterschied ist messbar: wer beim Betreten leerte,
	// nähme dem Enttappschritt das Mana, das im Aufräumschritt davor noch
	// hätte ausgegeben werden dürfen.
	//
	// **Es ist eine zugbasierte Handlung** (CR 703.4q) und geht nicht über
	// den Stapel — niemand kann darauf antworten.
	p = p.Manaleeren()

	naechster, neuerZug := p.Zug.Abschnitt.Naechster()

	p.Zug.Abschnitt = naechster
	p.Zug.Gepasst = nil

	if neuerZug {
		// **Der Tag-Nacht-Wechsel prüft den vorigen Zug** — CR 502.2 mit
		// CR 731.2a/b (M20,
		// #1621).
		// **Vor dem Zurücksetzen des Zählers**: gefragt ist, was der
		// vorige aktive Spieler gewirkt hat.
		p.Tageszeit = p.naechsteTageszeit()

		p.Zug.Nummer++
		p.Zug.Aktiv = p.naechsterSitz(p.Zug.Aktiv)

		// **Und der Zauberzähler geht mit dem Zug auf null** — dieselbe
		// Bewegung wie beim Landzähler darunter.
		p.Zug.Zauber = 0

		// **Der Landzähler geht mit dem Zug auf null** — CR 305.2a zählt,
		// was **in diesem Zug** gespielt wurde (M4, #1606).
		p.Zug.Laender = 0

		// **Und der Kampf ebenso** — CR 511.3 (M6, #1608).
		p.Zug.Kampf = Kampfstand{}
	}

	return p
}

// abschnittBetreten führt die zugbasierten Handlungen aus und gibt die
// Priorität — und geht weiter, wo niemand handeln darf.
//
// # Warum ein Abschnitt ohne Priorität nicht stehenbleibt
//
// **Weil ihn nichts weiterbrächte.** Ein Abschnitt endet, wenn alle
// nacheinander passen (CR 500.2); im Enttappen (CR 502.4) und im Aufräumen
// (CR 514.3) passt niemand, weil niemand Priorität hat. Er tut also, was er
// zu tun hat, und gibt weiter — und **genau das beschreibt CR 502.4**: was
// im Enttappen auslöst, wartet, bis das nächste Mal jemand Priorität bekäme.
//
// **Die eine Ausnahme ist die Abwurfpflicht** (CR 514.1): sie ist eine Wahl
// des Spielers, und bis er sie getroffen hat, steht der Aufräumschritt.
//
// **Die Schleife ist begrenzt**, und nicht durch eine Zahl: jeder Durchlauf
// geht einen Abschnitt weiter, und zehn der zwölf vergeben Priorität — nach
// höchstens zwei Durchläufen steht sie.
func (p Partie) abschnittBetreten() Partie {
	for {
		p = p.zughandlungen()

		// **Vor der Priorität, nicht danach** — CR 703.3 und CR 704.3
		// (Fassung 2026-08-19): erst die zugbasierten Handlungen, dann die
		// zustandsbasierten Aktionen, dann bekommt jemand Priorität. Wer sie
		// danach prüfte, gäbe einem Sitz Priorität, der die Partie schon
		// verlassen hat (M17, #1618).
		p = p.Zustandspruefung()

		// **Und danach, was ausgelöst hat** — CR 603.3 (M9,
		// #1611):
		// die wartenden Fähigkeiten kommen auf den Stapel, bevor jemand
		// Priorität bekommt.
		p = p.ausloeserAufDenStapel()

		p.Zug.Prioritaet = p.prioritaetFuer(p.Zug.Abschnitt, p.Zug.Aktiv)

		// **Gefragt wird der Abschnitt und nicht der Name.** Ein leerer
		// aktiver Spieler — eine Partie ohne Sitze, wie sie nur aus einem von
		// Hand geänderten Bestand kommt — gäbe eine leere Priorität, und wer
		// **darauf** weiterginge, drehte sich für immer. Gemessen am
		// 2026-09-17: der Lauf hing, statt rot zu werden.
		if p.Zug.Abschnitt.VergibtPrioritaet() || p.Zug.Abwurf > 0 {
			return p
		}

		p = p.abschnittVerlassen()
	}
}

// Passen gibt die Priorität weiter — und beendet den Abschnitt, wenn alle
// nacheinander gepasst haben.
//
// **Wer nicht an der Reihe ist, passt nicht.** Ein Passen des Falschen
// verschöbe die Reihenfolge, und danach wäre „alle nacheinander" eine andere
// Aussage (CR 117.4).
func (p Partie) Passen(wer string) (Partie, error) {
	if p.Zug.Prioritaet == "" {
		return p, fmt.Errorf(
			"%w: in diesem Abschnitt hat niemand Priorität — gepasst wird, "+
				"wo jemand handeln dürfte", ErrValidation)
	}

	if wer != p.Zug.Prioritaet {
		return p, fmt.Errorf(
			"%w: %q ist nicht an der Reihe — die Priorität hat %q",
			ErrValidation, wer, p.Zug.Prioritaet)
	}

	gepasst := make([]string, 0, len(p.Zug.Gepasst)+1)
	gepasst = append(gepasst, p.Zug.Gepasst...)
	gepasst = append(gepasst, wer)

	p.Zug.Gepasst = gepasst

	// **Alle nacheinander** — CR 117.4: so viele Passen hintereinander, wie
	// es Sitze gibt. Bei einem Sitz ist das eines, und genau deshalb steht
	// hier die Sitzzahl und keine Zwei (Solo Goldfish, #1460).
	//
	// **Und nur die, die noch dabei sind** — CR 800.4 (M13,
	// #1615):
	// eine Partie zu dritt läuft weiter, wenn einer geht, und wer gegangen
	// ist, passt nicht mehr. Wer ihn mitzählte, wartete auf eine Antwort,
	// die niemand geben kann.
	if len(gepasst) >= p.NochImSpiel() {
		return p.nachDemPassen(), nil
	}

	p.Zug.Prioritaet = p.naechsterSitz(wer)

	return p, nil
}

// prioritaetFuer sagt, wer in diesem Abschnitt anfängt — oder niemand.
//
// **CR 117.3a**: der aktive Spieler bekommt sie zu Beginn der meisten
// Abschnitte. „Die meisten" sind die zehn, die sie überhaupt vergeben
// ([Abschnitt.VergibtPrioritaet]).
func (p Partie) prioritaetFuer(abschnitt Abschnitt, aktiv string) string {
	if !abschnitt.VergibtPrioritaet() {
		return ""
	}

	// **Solange jemand seine Auslöser ordnet, bekommt niemand Priorität** —
	// CR 603.3 (M9, #1611):
	// gelegt wird, **bevor** das nächste Mal jemand Priorität bekäme. Ein
	// Sitz, der dazwischen handeln dürfte, handelte vor einer Fähigkeit, die
	// schon ausgelöst hat.
	if p.Ausloeserreihung != "" {
		return ""
	}

	// **Ein Zug ohne aktiven Spieler** — CR 800.4j (Fassung 2026-09-25): ist
	// er gegangen, bekommt die Priorität statt seiner der Nächste in
	// Zugreihenfolge (#1971).
	if p.istDraussen(aktiv) {
		return p.naechsterSitz(aktiv)
	}

	return aktiv
}

// naechsterSitz gibt den Nächsten in Zugreihenfolge — reihum.
//
// **Eine unbekannte Kennung bekommt den ersten Sitz.** Sie kann nur aus einem
// Bestand kommen, den jemand von Hand geändert hat; eine Partie, die daran
// hängenbliebe, wäre unbenutzbar (§8).
func (p Partie) naechsterSitz(wer string) string {
	if len(p.Sitze) == 0 {
		return ""
	}

	anfang := 0
	for stelle, sitz := range p.Sitze {
		if sitz.Kennung == wer {
			anfang = stelle + 1

			break
		}
	}

	// **Wer gegangen ist, wird übersprungen** — CR 800.4 (M13,
	// #1615).
	//
	// **Die härteste Stelle der Mehrspielerregeln**: eine Reihenfolge, die
	// einen Gegangenen mitzählt, gibt ihm Priorität, und niemand kann sie
	// ihm nehmen — die Partie hängt.
	//
	// **Die Schleife geht höchstens einmal herum.** Sind alle draußen,
	// bleibt der erste Sitz; das ist kein Spiel mehr, und wer dort eine
	// leere Kennung gäbe, ließe den Aufrufer raten (§8).
	for um := range len(p.Sitze) {
		naechster := p.Sitze[(anfang+um)%len(p.Sitze)]
		if !p.istDraussen(naechster.Kennung) {
			return naechster.Kennung
		}
	}

	return p.Sitze[0].Kennung
}

// nachDemPassen entscheidet, was geschieht, wenn alle passen — CR 405.5.
//
// # Die Regel hat zwei Hälften, und G9 hatte eine
//
// **CR 500.2**: ein Abschnitt endet, wenn der Stapel **leer** ist **und** alle
// nacheinander passen. Ohne Stapel war er immer leer, und G9
// (#1592)
// hat deshalb bei jedem vollständigen Passen weitergeschaltet. Seit `M1`
// (#1603)
// gibt es die erste Hälfte.
//
// **Und danach bekommt der aktive Spieler Priorität** — CR 117.3b, nicht der,
// der zuletzt gepasst hat und nicht der, der gewirkt hat. Das Issue nannte
// dafür CR 608.2k; die Nummer war falsch und ist berichtigt (§11).
func (p Partie) nachDemPassen() Partie {
	if p.StapelOben() == nil {
		return p.Weiter()
	}

	p = p.Verrechnen()
	p.Zug.Gepasst = nil

	// **Vor der Priorität die zustandsbasierten Aktionen** — CR 704.3,
	// dieselbe Reihenfolge wie beim Betreten eines Abschnitts.
	p = p.Zustandspruefung()

	// **Und danach, was ausgelöst hat** — CR 603.3 (M9,
	// #1611):
	// die wartenden Fähigkeiten kommen auf den Stapel, bevor jemand
	// Priorität bekommt. **Nach den zustandsbasierten Aktionen**, denn
	// CR 704.3 nennt sie zuerst.
	p = p.ausloeserAufDenStapel()

	p.Zug.Prioritaet = p.prioritaetFuer(p.Zug.Abschnitt, p.Zug.Aktiv)

	return p
}
