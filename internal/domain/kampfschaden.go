package domain

// **Kampfschaden** — Werkstück M6 der Regelmaschine
// (#1608).
//
// # Zuerst zuteilen, dann austeilen
//
// **CR 510.1 und CR 510.2 sind zwei Schritte in einem.** Erst weist jede
// Kreatur ihren Schaden zu, **dann** wird alles **gleichzeitig** ausgeteilt —
// und dazwischen kommt niemand. Der Unterschied ist messbar: eine 2/2, die
// von einer 3/3 erschlagen wird, teilt ihre zwei Schaden **trotzdem** aus.
// Wer nacheinander austeilte, ließe sie sterben, bevor sie zurückschlägt.
//
// **Deshalb wird erst gerechnet und dann geschrieben.**

// zuteilung ist ein zugeteilter Schaden, bevor er ausgeteilt wird.
type zuteilung struct {
	// Objekt ist die Kennung des getroffenen Permanents — leer, wenn ein
	// Sitz getroffen wird.
	Objekt string

	// Sitz ist die Kennung des getroffenen Spielers — leer, wenn ein
	// Permanent getroffen wird.
	Sitz string

	Menge int

	// Von ist die Kennung der Kreatur, die austeilt — CR 903.10a (M19).
	//
	// **Nur der Kampfschaden hat sie**, und nur dort zählt sie: CR 903.10a
	// spricht ausdrücklich von **Kampfschaden** desselben Commanders.
	Von string
}

// kampfschaden teilt den Kampfschaden zu und aus — CR 510.1, CR 510.2.
//
// **Nicht exportiert**: er ist eine **zugbasierte Handlung** (CR 703.4m) und
// geschieht von selbst, sobald der Schritt beginnt. Ein Befehl dafür wäre
// eine Wahl, wo die Regeln keine lassen — und an einer Fläche stünde damit
// etwas, das niemand entscheiden darf (§16).
//
// **Und ohne Fehler.** Ein Wächter über den Schritt wäre ein Zweig, den kein
// Weg betritt: der einzige Aufrufer ist [Partie.kampfschadenAlsHandlung], und
// die läuft genau in diesem Schritt. Ein solcher Zweig ist ein Versprechen,
// das niemand hält (§10); die Bedingung steht dort, wo sie entschieden wird.
func (p Partie) kampfschaden() Partie {
	zuteilungen := p.kampfzuteilungen()

	objekte := make([]Spielobjekt, len(p.Objekte))
	copy(objekte, p.Objekte)

	sitze := make([]Sitz, len(p.Sitze))
	copy(sitze, p.Sitze)

	for _, eine := range zuteilungen {
		// **Lebensverknüpfung gibt Leben, gleich wen es trifft** —
		// CR 702.15b (M14,
		// #1616):
		// „Schaden aus einer Quelle mit Lebensverknüpfung", und ein
		// Permanent ist eine Quelle wie ein Spieler.
		if p.HatFaehigkeit(eine.Von, SchluesselLebensverknuepfung) {
			lebengeben(sitze, p.beherrscherVon(eine.Von), eine.Menge)
		}

		if eine.Sitz != "" {
			// **An einen Gegangenen wird nichts zugeteilt** — CR 800.4e
			// (M13, #1615).
			// Sein Sitz steht noch in der Liste, weil der Ausgang ihn
			// nennt; im Spiel ist er nicht mehr.
			if p.istDraussen(eine.Sitz) {
				continue
			}

			for stelle := range sitze {
				if sitze[stelle].Kennung == eine.Sitz {
					sitze[stelle].Leben -= eine.Menge

					// **Und wer dem Monarchen schadet, löst aus** —
					// CR 725.2 (M9, #1611). **Vor dem Commanderzähler
					// und nach dem Leben**: die Reihenfolge ist gleich,
					// beide lesen nur.
					p = p.monarchwechselDurchKampfschaden(eine)

					// **Und je Commander gezählt** — CR 903.10a (M19,
					// #1620).
					// **Zusätzlich und nicht anstelle**: das Leben sinkt
					// wie bei jedem anderen Schaden.
					p.commanderschadenZaehlen(&sitze[stelle], eine)
				}
			}

			continue
		}

		// **Und was verhindert wird, wird nicht markiert** — CR 615.1
		// (M11, #1613).
		nachVerhinderung := p.schadenVerhindern(eine.Objekt, eine.Menge)

		if nachVerhinderung == 0 {
			continue
		}

		for stelle := range objekte {
			if objekte[stelle].Kennung != eine.Objekt {
				continue
			}

			objekte[stelle].Schaden += nachVerhinderung

			// **Und Todesberührung merkt sich, dass sie da war** —
			// CR 702.2b mit CR 704.5h (M14): nicht wie viel, sondern
			// **dass** Schaden aus einer solchen Quelle angekommen ist.
			if p.HatFaehigkeit(eine.Von, SchluesselTodesberuehrung) {
				objekte[stelle].Todesberuehrt = true
			}
		}
	}

	p.Objekte = objekte
	p.Sitze = sitze
	p.Zug.Kampf.Schadensschritte++

	// **Und die Initiative wechselt je Beherrscher und nicht je Kreatur** —
	// CR 726.2 (M9, #1611). Deshalb steht sie **nach** der Schleife: sie
	// fragt „eine oder mehr Kreaturen eines Spielers", und das ist eine
	// Aussage über die ganze Zuteilung.
	return p.initiativeDurchKampfschaden(zuteilungen)
}

// kampfzuteilungen rechnet aus, wer wem wie viel zuteilt — CR 510.1a–d.
//
// **Gerechnet und nicht geschrieben**: der Stand, gegen den gerechnet wird,
// ist der **vor** dem Austeilen. Genau das verlangt CR 510.2.
func (p Partie) kampfzuteilungen() []zuteilung {
	var alle []zuteilung

	for _, angriff := range p.Zug.Kampf.Angriffe {
		if p.teiltInDiesemSchrittAus(angriff.Kreatur) {
			alle = append(alle, p.angreiferzuteilung(angriff)...)
		}
	}

	for _, block := range p.Zug.Kampf.Bloecke {
		if p.teiltInDiesemSchrittAus(block.Kreatur) {
			alle = append(alle, p.blockerzuteilung(block)...)
		}
	}

	return alle
}

// angreiferzuteilung gibt, was ein Angreifer austeilt — CR 510.1b, CR 510.1c.
func (p Partie) angreiferzuteilung(angriff Angriff) []zuteilung {
	staerke := p.kampfstaerke(angriff.Kreatur)
	if staerke <= 0 {
		// **Null oder weniger teilt gar nichts zu** — CR 510.1a sagt es
		// ausdrücklich, und der Unterschied ist sichtbar: ein Eintrag mit
		// null Schaden zählte als Schaden (CR 704.5g).
		return nil
	}

	blocker := p.Zug.Kampf.blocker(angriff.Kreatur)
	if len(blocker) == 0 {
		// **Ungeblockt trifft, was angegriffen wurde** — CR 510.1b.
		return []zuteilung{p.zielzuteilung(
			angriff.Ziel, staerke, angriff.Kreatur)}
	}

	// **Geblockt trifft die Blocker** — CR 510.1c.
	//
	// **Der erste bekommt alles, wenn es mehrere sind.** CR 509.2 lässt den
	// Angreifer eine Reihenfolge wählen, und CR 510.1c teilt in dieser
	// Reihenfolge zu; eine Wahl gibt es am Tisch nicht (Grenze in der
	// Matrix), also gilt die Reihenfolge der Erklärung.
	//
	// **Und Trampelschaden gibt den Rest weiter** — CR 702.19b (M14,
	// #1616).
	if !p.HatFaehigkeit(angriff.Kreatur, SchluesselTrampelschaden) {
		return []zuteilung{{
			Objekt: blocker[0], Menge: staerke, Von: angriff.Kreatur,
		}}
	}

	return p.trampelzuteilung(angriff, blocker[0], staerke)
}

// trampelzuteilung teilt tödlich auf den Blocker und den Rest weiter —
// CR 702.19b.
//
// **„Tödlich" heißt: was noch fehlt** — die Widerstandskraft abzüglich des
// Schadens, der schon markiert ist (CR 702.19b nennt es „lethal damage", und
// CR 704.5g sagt, wann Schaden tödlich ist). **Mit Todesberührung genügt
// eines** (CR 702.2c).
//
// **Und unter eins geht es nicht**: ein Blocker, der schon tödlichen Schaden
// trägt, bekommt trotzdem einen — sonst träte der Angreifer an ihm vorbei,
// ohne ihn zu berühren.
func (p Partie) trampelzuteilung(
	angriff Angriff, blocker string, staerke int,
) []zuteilung {
	toedlich := 1

	if !p.HatFaehigkeit(angriff.Kreatur, SchluesselTodesberuehrung) {
		toedlich = max(1, p.fehlenderSchaden(blocker))
	}

	if toedlich >= staerke {
		return []zuteilung{{
			Objekt: blocker, Menge: staerke, Von: angriff.Kreatur,
		}}
	}

	return []zuteilung{
		{Objekt: blocker, Menge: toedlich, Von: angriff.Kreatur},
		p.zielzuteilung(angriff.Ziel, staerke-toedlich, angriff.Kreatur),
	}
}

// fehlenderSchaden sagt, wie viel diesem Permanent noch zum Tod fehlt —
// CR 704.5g.
func (p Partie) fehlenderSchaden(kennung string) int {
	// **Den Blocker gibt es** — [Partie.blockPruefen] hat ihn auf dem
	// Schlachtfeld gefunden, bevor er in den Kampfstand kam. Ein Zweig für
	// „nicht da" wäre einer, den kein Weg betritt (§10).
	fehlt := 1

	for _, objekt := range p.Objekte {
		if objekt.Kennung == kennung {
			_, widerstand := p.Werte(objekt)
			fehlt = widerstand.Zahl - objekt.Schaden
		}
	}

	return fehlt
}

// zielzuteilung sagt, ob ein Ziel ein Sitz oder ein Permanent ist.
func (p Partie) zielzuteilung(ziel string, menge int, von string) zuteilung {
	for _, sitz := range p.Sitze {
		if sitz.Kennung == ziel {
			return zuteilung{Sitz: ziel, Menge: menge, Von: von}
		}
	}

	return zuteilung{Objekt: ziel, Menge: menge, Von: von}
}

// blockerzuteilung gibt, was ein Blocker austeilt — CR 510.1d.
func (p Partie) blockerzuteilung(block Block) []zuteilung {
	staerke := p.kampfstaerke(block.Kreatur)
	if staerke <= 0 {
		return nil
	}

	// **Ein Blocker trifft den, den er blockt** — CR 510.1d. Blockt er
	// mehrere, bekommt der erste alles; der Grund steht oben.
	return []zuteilung{{
		Objekt: block.Angreifer, Menge: staerke, Von: block.Kreatur,
	}}
}

// kampfstaerke gibt die Stärke einer Kreatur **mit Marken** — CR 122.1a.
//
// **Null, wo keine dasteht**: eine Kreatur ohne lesbare Stärke teilt nichts
// zu, statt dass jemand eine erfindet (§10).
func (p Partie) kampfstaerke(kennung string) int {
	for _, objekt := range p.Objekte {
		if objekt.Kennung != kennung || objekt.Zone != ZoneTisch {
			continue
		}

		staerke, _ := p.Werte(objekt)

		if !staerke.Steht || staerke.Stern {
			return 0
		}

		return staerke.Zahl
	}

	return 0
}

// beherrscherVon gibt den Beherrscher eines Objekts.
//
// **Die Quelle gibt es** — sie ist eine Kreatur aus dem Kampfstand, und der
// wird beim Erklären geprüft (derselbe Grund wie bei [Partie.fehlenderSchaden]).
func (p Partie) beherrscherVon(kennung string) string {
	wer := ""

	for _, objekt := range p.Objekte {
		if objekt.Kennung == kennung {
			wer = objekt.Beherrscher
		}
	}

	return wer
}

// lebengeben trägt einem Sitz Leben nach — CR 119.3.
func lebengeben(sitze []Sitz, wer string, um int) {
	for stelle := range sitze {
		if sitze[stelle].Kennung == wer {
			sitze[stelle].Leben += um
		}
	}
}
