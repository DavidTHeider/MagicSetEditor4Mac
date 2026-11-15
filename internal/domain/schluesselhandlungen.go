package domain

// **Handlungen mit Schlüsselwort** — Werkstück M7 der Regelmaschine
// (#1609).
//
// # Welche von den neunundsechzig hier stehen
//
// **Die, die das Regelwerk allein beschreibt.** CR 701 führt 69 Handlungen
// (701.2 „Activate" bis 701.70 „Recruit"); die meisten sind Bausteine, aus
// denen ein **Kartentext** besteht, und dieser Text ist `M8`
// (#1610).
//
// **Hier stehen die, die ohne Text auskommen** — sie sagen nur, **wohin**
// etwas geht oder **was** ein Spieler tut:
//
//   - **Zerstören** (701.8a) — vom Schlachtfeld in den Friedhof des Eigners.
//   - **Opfern** (701.21a) — dasselbe, aber vom Beherrscher, und nur bei
//     einem Permanent, das er beherrscht.
//   - **Exilieren** (701.13a) — von überall ins Exil.
//   - **Neutralisieren** (701.6a) — vom Stapel in den Friedhof des Eigners,
//     ohne verrechnet zu werden.
//   - **Mahlen** (701.17a) — die obersten Karten der Bibliothek in den
//     Friedhof.
//   - **Kämpfen** (701.14a) — zwei Kreaturen teilen einander ihre Stärke zu.
//
// **Die Wahlen stehen daneben** in `schluesselwahlen.go`: Hellsehen
// (701.22a) und Überwachen (701.25a) verlangen eine Entscheidung, und die
// kommt als Nutzlast.
//
// # Und was daraus nicht folgt
//
// **Kein Kartentext ruft sie.** Sie stehen als Befehle da, damit ein Mensch
// sie am Tisch tut — so wie er heute eine Karte von Hand umlegt. Was eine
// Karte sagt, führt weiterhin niemand aus (CR 608.2c); das steht als Grenze
// in der Regelmatrix.

import "fmt"

// Zerstoeren legt ein Permanent in den Friedhof seines Eigners — CR 701.8a.
//
// **Vom Schlachtfeld und von nirgends sonst**: CR 701.8a spricht von einem
// Permanent, und Permanente liegen auf dem Schlachtfeld (CR 110.1). Was in
// einer anderen Zone liegt, lässt sich nicht zerstören — das ist ein Befund
// und kein Fehler (§8), und die Meldung sagt es.
// **Und Unzerstörbares wird nicht zerstört** — CR 702.12b (M14,
// #1616).
// **Eine Abweisung und kein stilles Nichts**: wer den Befehl stellt, soll
// erfahren, warum er nichts bewirkt hat (§8).
func (p Partie) Zerstoeren(objekt string) (Partie, error) {
	stelle, fehler := p.stelleInZone(objekt, ZoneTisch, "zerstören", "701.8a")
	if fehler != nil {
		return p, fehler
	}

	if p.HatFaehigkeit(objekt, SchluesselUnzerstoerbar) {
		return p, fmt.Errorf(
			"%w: %q ist unzerstörbar (CR 702.12b)", ErrValidation, objekt)
	}

	return p.inDenFriedhof(stelle), nil
}

// Opfern ist dasselbe, aber vom Beherrscher — CR 701.21a.
//
// **Wer opfert, muss beherrschen.** CR 701.21a sagt es ausdrücklich: niemand
// opfert etwas, das ihm nicht gehört. Das ist der einzige Unterschied zum
// Zerstören — und er ist der Grund, warum beide getrennt dastehen.
func (p Partie) Opfern(wer, objekt string) (Partie, error) {
	stelle, fehler := p.stelleInZone(objekt, ZoneTisch, "opfern", "701.21a")
	if fehler != nil {
		return p, fehler
	}

	if p.Objekte[stelle].Beherrscher != wer {
		return p, fmt.Errorf(
			"%w: %q beherrscht %q nicht und opfert es nicht (CR 701.21a)",
			ErrValidation, wer, objekt)
	}

	return p.inDenFriedhof(stelle), nil
}

// Exilieren schickt ein Objekt ins Exil — CR 701.13a.
//
// **Von überall.** CR 701.13a nennt keine Zone: was im Exil landet, kam von
// dort, wo es lag. Das ist die einzige dieser Handlungen, die keine
// Zonenbedingung hat.
func (p Partie) Exilieren(objekt string) (Partie, error) {
	stelle, fehler := p.objektstelle(objekt)
	if fehler != nil {
		return p, fehler
	}

	p = p.zonenwechsel(stelle, ZoneExil)

	// **Außerhalb von Stapel und Schlachtfeld hat ein Objekt keinen
	// Beherrscher** — CR 109.4. Der Eigner ist, was bleibt.
	p.Objekte[stelle].Beherrscher = p.Objekte[stelle].Eigner

	return p, nil
}

// Neutralisieren nimmt einen Zauber vom Stapel — CR 701.6a.
//
// **Er wird nicht verrechnet, und nichts von ihm geschieht.** Was hier
// trägt, ist der **Weg**: vom Stapel in den Friedhof des Eigners. Was der
// Zauber getan hätte, führt ohnehin niemand aus (CR 608.2c).
func (p Partie) Neutralisieren(objekt string) (Partie, error) {
	stelle, fehler := p.stelleInZone(
		objekt, ZoneStapel, "neutralisieren", "701.6a")
	if fehler != nil {
		return p, fehler
	}

	return p.inDenFriedhof(stelle), nil
}

// Mahlen legt die obersten Karten der Bibliothek in den Friedhof — CR 701.17a.
//
// **So viele, wie da sind.** CR 701.17a sagt „so viele Karten"; sind weniger
// da, gehen die da sind — und das ist **kein Fehler**. Ob daraus ein Verlust
// wird, entscheidet die zustandsbasierte Aktion beim nächsten Ziehen
// (CR 704.5b).
func (p Partie) Mahlen(wer string, wie int) (Partie, error) {
	if wie < 0 {
		return p, fmt.Errorf("%w: %d Karten mahlen geht nicht (CR 701.17a)",
			ErrValidation, wie)
	}

	if !p.sitzSteht(wer) {
		return p, fmt.Errorf("%w: den Sitz %q gibt es in dieser Partie nicht",
			ErrNotFound, wer)
	}

	for range wie {
		stelle := obersteDerBibliothek(p.Objekte, wer)
		if stelle < 0 {
			break
		}

		// **Über den einen Weg** — CR 400.7 (M18, #1619).
		p = p.zonenwechsel(stelle, ZoneFriedhof)
		p.Objekte[stelle].Beherrscher = p.Objekte[stelle].Eigner
	}

	return p, nil
}

// Kaempfen lässt zwei Kreaturen einander ihre Stärke zuteilen — CR 701.14a.
//
// **Gleichzeitig**, wie im Kampf (CR 510.2): beide rechnen gegen den Stand
// **vor** dem Austeilen, und eine sterbende Kreatur teilt trotzdem aus.
//
// **Es ist kein Kampfschaden** (CR 701.14b): er kommt von keiner der beiden
// Kreaturen als Angreifer oder Blocker. Für dieses Programm ist der
// Unterschied heute keiner — er wird es, sobald ein Effekt danach fragt
// (`M8`).
func (p Partie) Kaempfen(erste, zweite string) (Partie, error) {
	if erste == zweite {
		return p, fmt.Errorf(
			"%w: %q kämpft nicht gegen sich selbst (CR 701.14a)",
			ErrValidation, erste)
	}

	links, fehler := p.kampffaehig(erste)
	if fehler != nil {
		return p, fehler
	}

	rechts, fehler := p.kampffaehig(zweite)
	if fehler != nil {
		return p, fehler
	}

	// **Erst rechnen, dann schreiben** — der Grund steht bei
	// [Partie.kampfschaden].
	linksStaerke := p.kampfstaerke(erste)
	rechtsStaerke := p.kampfstaerke(zweite)

	objekte := make([]Spielobjekt, len(p.Objekte))
	copy(objekte, p.Objekte)

	if rechtsStaerke > 0 {
		objekte[links].Schaden += rechtsStaerke
	}

	if linksStaerke > 0 {
		objekte[rechts].Schaden += linksStaerke
	}

	p.Objekte = objekte

	return p, nil
}

// kampffaehig prüft, ob ein Objekt eine Kreatur auf dem Schlachtfeld ist.
func (p Partie) kampffaehig(kennung string) (int, error) {
	stelle, fehler := p.stelleInZone(kennung, ZoneTisch, "kämpfen", "701.14a")
	if fehler != nil {
		return 0, fehler
	}

	eigen, da := p.Eigenschaften(p.Objekte[stelle])
	if !da || !eigen.Typzeile.Ist(TypKreatur) {
		return 0, fmt.Errorf("%w: %q ist keine Kreatur (CR 701.14a)",
			ErrValidation, kennung)
	}

	return stelle, nil
}

// inDenFriedhof legt ein Objekt in den Friedhof seines Eigners.
//
// **Und der Beherrscher geht mit** — CR 109.4: außerhalb von Stapel und
// Schlachtfeld hat ein Objekt keinen; der Eigner ist, was bleibt.
func (p Partie) inDenFriedhof(stelle int) Partie {
	// **Und vorher die Frage, ob es wirklich der Friedhof ist** — CR 614.1a
	// (M11, #1613).
	// **Ein ersetztes Ereignis geschieht nie** (CR 614.6): wer hier zuerst
	// ablegte und danach ins Exil schöbe, hätte den Friedhof berührt.
	ziel := p.friedhofErsetzen(p.Objekte[stelle].Kennung)

	// **Über den einen Weg** — CR 400.7 (M18,
	// #1619):
	// Marken, Zustand, markierter Schaden und was daran hing, bleiben
	// zurück.
	p = p.zonenwechsel(stelle, ziel)

	// **Außerhalb von Schlachtfeld und Stapel gibt es keinen Beherrscher**
	// — CR 109.4. Der Eigner ist, was bleibt.
	p.Objekte[stelle].Beherrscher = p.Objekte[stelle].Eigner

	return p
}

// objektstelle findet ein Objekt an seiner Kennung.
func (p Partie) objektstelle(kennung string) (int, error) {
	for stelle := range p.Objekte {
		if p.Objekte[stelle].Kennung == kennung {
			return stelle, nil
		}
	}

	return 0, fmt.Errorf("%w: das Objekt %q gibt es in dieser Partie nicht",
		ErrNotFound, kennung)
}

// stelleInZone findet ein Objekt und prüft, wo es liegt.
//
// **Die Zone steht in der Meldung**, und die Regelnummer auch: wer eine
// Karte von der Hand zerstören will, soll lesen, woran es lag (§8).
func (p Partie) stelleInZone(
	kennung string, zone Zone, handlung, regel string,
) (int, error) {
	stelle, fehler := p.objektstelle(kennung)
	if fehler != nil {
		return 0, fehler
	}

	if p.Objekte[stelle].Zone != zone {
		return 0, fmt.Errorf(
			"%w: %q liegt in %q und lässt sich dort nicht %s (CR %s)",
			ErrValidation, kennung, p.Objekte[stelle].Zone, handlung, regel)
	}

	return stelle, nil
}

// sitzSteht sagt, ob es diesen Sitz gibt.
func (p Partie) sitzSteht(kennung string) bool {
	for _, sitz := range p.Sitze {
		if sitz.Kennung == kennung {
			return true
		}
	}

	return false
}

// obersteDerBibliothek gibt die Stelle der obersten Karte — minus eins, wo
// keine mehr liegt.
func obersteDerBibliothek(objekte []Spielobjekt, wer string) int {
	oberste := -1

	for stelle := range objekte {
		if objekte[stelle].Zone != ZoneBibliothek ||
			objekte[stelle].Beherrscher != wer {
			continue
		}

		if oberste < 0 ||
			objekte[stelle].Ordnung < objekte[oberste].Ordnung {
			oberste = stelle
		}
	}

	return oberste
}
