package domain

// **Der Stapel** — Werkstück M1 der Regelmaschine
// (#1603).
//
// # Die Zone gab es seit G1, und niemand legte etwas darauf
//
// **`ZoneStapel` steht seit dem Spielkern da** und ist seit G9 als geordnet
// geführt (CR 400.5). Was fehlte, war das **Verrechnen**: ein Objekt, das
// darauf liegt, blieb liegen, und CR 500.2 („der Stapel ist leer **und** alle
// passen nacheinander") hatte nur seine zweite Hälfte.
//
// # Was M2 dafür geliefert hat
//
// **Das Verrechnen braucht den Kartentyp.** CR 608.2n legt einen Spontan-
// oder Hexereizauber in den Friedhof, CR 608.3a lässt ein bleibendes Objekt
// auf das Schlachtfeld kommen — welches von beiden gilt, steht in der
// Typzeile. Vor `M2`
// (#1604)
// wäre die Entscheidung geraten gewesen.
//
// # Was hier ausdrücklich nicht geschieht
//
// **„Tu, was das Objekt sagt"** (CR 608.2c) — der Kern des Verrechnens — ist
// eine Fähigkeit, und Fähigkeiten gibt es nicht (`M8`,
// #1610).
// Was hier steht, ist der **Weg**: das Objekt verlässt den Stapel und landet,
// wo die Regel es hinlegt. Der Rest des Zaubers geschieht nicht, und das ist
// eine Grenze und keine Auslassung (§10).
//
// **Und niemand legt etwas darauf.** Das Wirken ist `M4`
// (#1606);
// bis dahin kommt ein Objekt über den Sandkastenbefehl `objekt.schieben`
// dorthin — von Hand, wie alles im Sandkasten.

// StapelOben gibt das oberste Objekt des Stapels — **nil heißt: er ist leer**.
//
// **Oben ist die kleinste Ordnung**, wie in jeder geordneten Zone dieses
// Programms (siehe [Spielobjekt.Ordnung]). CR 405.2 sagt, dass jedes neue
// Objekt **über** alle vorhandenen kommt; das ist dieselbe Aussage von der
// anderen Seite.
func (p Partie) StapelOben() *Spielobjekt {
	oberste := -1

	for stelle, objekt := range p.Objekte {
		if objekt.Zone != ZoneStapel {
			continue
		}

		if oberste < 0 || objekt.Ordnung < p.Objekte[oberste].Ordnung {
			oberste = stelle
		}
	}

	if oberste < 0 {
		return nil
	}

	return &p.Objekte[oberste]
}

// Verrechnen nimmt das oberste Objekt vom Stapel — CR 608.1.
//
// **Ein leerer Stapel gibt dieselbe Partie** und ist damit gefahrlos zu
// rufen: CR 405.5 lässt in diesem Fall den Abschnitt enden, und das
// entscheidet [Partie.Passen].
func (p Partie) Verrechnen() Partie {
	oben := p.StapelOben()
	if oben == nil {
		return p
	}

	kennung := oben.Kennung

	objekte := make([]Spielobjekt, len(p.Objekte))
	copy(objekte, p.Objekte)

	// **Ein Zauber, dessen Ziele alle unzulässig wurden, verpufft** —
	// CR 608.2b, Werkstück M4
	// (#1606).
	//
	// **Er wird dann nicht verrechnet und trotzdem abgelegt** — CR 608.2b
	// sagt beides in einem Satz. **Nicht CR 608.2m** — die handelt davon,
	// dass ein Zauber, der den Stapel verlässt, zu Ende verrechnet wird; er
	// verlässt den Stapel wie jeder andere. Der Unterschied ist, was er
	// unterwegs **tut** — und was ein Zauber tut, weiß ohne `M8` ohnehin
	// niemand. **Hier steht er trotzdem**, weil er den Weg bestimmt: ein
	// bleibender Zauber mit unzulässigem Ziel kommt **nicht** auf das
	// Schlachtfeld, sondern in den Friedhof.
	verpufft := !p.zieleNochZulaessig(*oben)

	p.Objekte = objekte

	for stelle := range p.Objekte {
		if p.Objekte[stelle].Kennung != kennung {
			continue
		}

		// **Eine Fähigkeit geht ihren eigenen Weg** — CR 603.3 mit
		// CR 608.2n (M9,
		// #1611):
		// sie ist keine Karte, sie tut, was das Regelwerk sagt, und danach
		// hört sie auf zu existieren.
		if p.Objekte[stelle].Art == ObjektFaehigkeit {
			return p.faehigkeitVerrechnen(stelle)
		}

		// **Und jetzt geschieht, was auf der Karte steht** — CR 608.2c
		// (M8, #1610).
		//
		// **Vor dem Zonenwechsel**: CR 608.2n legt den Zauber **als
		// letzten Teil** seiner Verrechnung ab, und ein Objekt, das schon
		// im Friedhof liegt, hätte seine Wirkungen verloren (CR 400.7).
		//
		// **Und nicht, wenn er verpufft** — CR 608.2b: dann wird er gar
		// nicht verrechnet.
		// **Gelesen wird jetzt und nicht beim Ankündigen** — CR 608.2c
		// (#1625).
		// Zwischen den beiden Augenblicken liegt mindestens eine
		// Prioritätsrunde, und was in die Rechnung eingeht, kann sich darin
		// geändert haben.
		if !verpufft {
			p = p.gedeutet(stelle)
			p = p.wirkungenAusfuehren(stelle)
		}

		// **Was den Stapel verlässt, ist ein neues Objekt** — CR 400.7
		// (M18, #1619).
		// Die Ziele sind damit weg, und das ist die Regel und nicht eine
		// Aufräumarbeit: sie galten **diesem** Zauber.
		p = p.zonenwechsel(stelle, p.abgelegtNach(stelle, verpufft))
		p.ohneBeherrscherImFriedhof(stelle)

		break
	}

	return p
}

// abgelegtNach sagt, wohin ein verrechnetes Objekt geht.
func (p Partie) abgelegtNach(stelle int, verpufft bool) Zone {
	if verpufft || p.gehtInDenFriedhof(p.Objekte[stelle]) {
		return ZoneFriedhof
	}

	return ZoneTisch
}

// ohneBeherrscherImFriedhof nimmt dem Objekt den Beherrscher, wo es keinen
// haben kann — CR 109.4.
//
// **Auf dem Schlachtfeld bleibt er**, und zwar der, der gewirkt hat
// (CR 608.3a): unter dessen Kontrolle kommt das Permanent.
func (p Partie) ohneBeherrscherImFriedhof(stelle int) {
	if p.Objekte[stelle].Zone == ZoneFriedhof {
		p.Objekte[stelle].Beherrscher = p.Objekte[stelle].Eigner
	}
}

// gehtInDenFriedhof sagt, ob ein verrechnetes Objekt in den Friedhof geht
// statt auf das Schlachtfeld.
//
// # Die beiden Fälle
//
// **CR 608.2n**: als letzter Teil des Verrechnens geht ein Spontan- oder
// Hexereizauber in den Friedhof seines **Eigners** — und deshalb wechselt
// dabei der Beherrscher: außerhalb von Stapel und Schlachtfeld hat ein Objekt
// keinen (CR 109.4), und der Eigner ist, was bleibt.
//
// **CR 608.3a**: ein bleibendes Objekt ohne Ziel wird ein Permanent und kommt
// auf das Schlachtfeld, unter die Kontrolle dessen, der es gewirkt hat — also
// unter die, die es schon hat.
//
// # Und der Fall, den die CR nicht kennt
//
// **Ein Objekt ohne lesbare Typzeile** — ein Spielstein (CR 111) oder eine
// Karte, deren Abzug keine Felder trägt. Es geht auf das Schlachtfeld, weil
// CR 608.3a der allgemeine Fall ist; **eine erfundene Typzeile wäre
// schlimmer** (§10). Wo das falsch ist, sieht man es: ein Spontanzauber
// bliebe liegen statt in den Friedhof zu gehen.
func (p Partie) gehtInDenFriedhof(objekt Spielobjekt) bool {
	eigen, da := p.Eigenschaften(objekt)

	return da &&
		(eigen.Typzeile.Ist(TypSpontan) || eigen.Typzeile.Ist(TypHexerei))
}

// wirkungenAusfuehren führt aus, was auf einem Zauber steht — CR 608.2c.
//
// **Eine nach der anderen, in der Reihenfolge des Textes.** CR 608.2c sagt,
// dass die Anweisungen der Reihe nach befolgt werden; wer sie umstellte,
// zöge eine Karte, die er erst mahlen wollte.
//
// **Was nicht geschehen kann, wird ausgelassen** — CR 101.2, und
// [Partie.Anwenden] sagt es. **Ohne Abbruch**: die übrigen Anweisungen
// gelten weiter, denn CR 101.2 lässt die eine aus und nicht den Rest.
func (p Partie) wirkungenAusfuehren(stelle int) Partie {
	for _, wirkung := range p.Objekte[stelle].Wirkungen {
		p, _ = p.Anwenden(wirkung)
	}

	return p
}
