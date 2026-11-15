package domain

// **Die zustandsbasierten Aktionen, die ein Objekt kosten** — Werkstück M5
// der Regelmaschine
// (#1607).
//
// # Was hier steht und was nebenan
//
// **`zustandsaktionen.go` führt die Runde und die Regeln, die einen Sitz
// kosten** (CR 704.5a–c). Hier stehen die, die ein **Objekt** treffen — und
// dafür braucht es, was `M2`
// (#1604)
// gebracht hat: Widerstandskraft, Kartentyp und die Zahlen auf der Karte.
//
// # In den Friedhof des **Eigners**
//
// **Jede dieser Regeln sagt „owner's graveyard"**, nicht „controller's".
// Außerhalb von Stapel und Schlachtfeld hat ein Objekt ohnehin keinen
// Beherrscher (CR 109.4), und deshalb wechselt er beim Ablegen mit.

const (
	// MarkeLoyalitaet sind Loyalitätsmarken — CR 122.1e (Fassung
	// 2026-08-19), gelesen von CR 704.5i.
	//
	// **Der Name ist der dieses Programms**, so wie jede Marke einen Namen
	// hat ([Spielobjekt.Zaehler]). Die CR nennt sie „loyalty counters";
	// welches Wort im Zähler steht, entscheidet, wer sie ablegt.
	MarkeLoyalitaet = "loyalität"

	// MarkeVerteidigung sind Verteidigungsmarken — CR 310.4b, gelesen von
	// CR 704.5v und CR 704.5w.
	MarkeVerteidigung = "verteidigung"
)

// objektaktionen wendet an, was ein Objekt kostet — und sagt, ob etwas
// geschah.
//
// **Eine Runde und keine Schleife**: die Wiederholung aus CR 704.3 führt
// [Partie.Zustandspruefung].
func (p Partie) objektaktionen() (Partie, bool) {
	geblieben := make([]Spielobjekt, 0, len(p.Objekte))

	etwas := false

	for _, objekt := range p.Objekte {
		ausgang := p.ausgangFuer(objekt)

		switch ausgang {
		case bleibt:
			geblieben = append(geblieben, objekt)

			continue

		// **CR 704.5d** — es hört auf zu existieren, und damit steht es in
		// keiner Zone mehr. Es wird nicht abgelegt, es wird **weggelassen**.
		case verschwindet:
			etwas = true

			continue

		case inDenFriedhof:
		}

		etwas = true

		// **CR 400.7** (M18, #1619): Marken, Zustand und Anhängsel bleiben
		// zurück. Was daran hing, löst `anhaengselaktionen` in derselben
		// Runde — CR 704.4 lässt sie gleichzeitig laufen.
		objekt = NeuesObjekt(objekt, p.Zug.Nummer)
		objekt.Zone = ZoneFriedhof
		objekt.Beherrscher = objekt.Eigner

		geblieben = append(geblieben, objekt)
	}

	if !etwas {
		return p, false
	}

	p.Objekte = geblieben

	return p, true
}

// zustandsausgang sagt, was mit einem Objekt geschieht.
type zustandsausgang int

const (
	bleibt zustandsausgang = iota
	verschwindet
	inDenFriedhof
)

// zielzoneFuer sagt, wohin ein Objekt gehört — und ob es überhaupt muss.
//
// # Die vier Regeln in ihrer Reihenfolge
//
//   - **CR 704.5d**: ein Spielstein außerhalb des Schlachtfelds hört auf zu
//     existieren. Er bekommt [ZoneHeimatlos] und wird danach entfernt.
//   - **CR 704.5f**: Widerstandskraft null oder weniger — in den Friedhof.
//   - **CR 704.5g**: markierter Schaden ab der Widerstandskraft — zerstört.
//   - **CR 704.5h**: Schaden aus einer Quelle mit Todesberührung, gleich wie
//     viel (M14, #1616).
//   - **CR 704.5i**: ein Planeswalker mit Loyalität null.
//   - **CR 704.5v/w**: eine Kampfkarte mit Verteidigung null.
//
// **Die Reihenfolge ist keine Rangfolge** (CR 704.4) — sie steht so, damit
// die Antwort nicht davon abhängt, in welcher Reihenfolge die Objekte liegen.
func (p Partie) ausgangFuer(objekt Spielobjekt) zustandsausgang {
	// **CR 704.5d** — und nur außerhalb des Schlachtfelds.
	if objekt.Art == ObjektSpielstein && objekt.Zone != ZoneTisch {
		return verschwindet
	}

	// **Alles Weitere gilt für Permanente** (CR 110.1): sie liegen auf dem
	// Schlachtfeld.
	if objekt.Zone != ZoneTisch {
		return bleibt
	}

	eigen, da := p.Eigenschaften(objekt)
	if !da {
		return bleibt
	}

	switch {
	case p.stirbtAnWiderstandskraft(objekt):
		return inDenFriedhof

	// **CR 704.5i** — der Planeswalker mit Loyalität null.
	case eigen.Typzeile.Ist(TypPlaneswalker) &&
		markeBeiNull(objekt, MarkeLoyalitaet):
		return inDenFriedhof

	// **CR 704.5v und CR 704.5w** — die Kampfkarte mit Verteidigung null.
	// Beide Regeln sagen dasselbe; sie unterscheiden sich nur darin, ob eine
	// Fähigkeit noch auf dem Stapel wartet, und die gibt es hier nicht.
	case eigen.Typzeile.Ist(TypKampf) &&
		markeBeiNull(objekt, MarkeVerteidigung):
		return inDenFriedhof
	}

	return bleibt
}

// stirbtAnWiderstandskraft prüft CR 704.5f und CR 704.5g.
//
// **Ohne Widerstandskraft geschieht nichts** — CR 208.3: eine Nichtkreatur
// hat keine, und ein Land stirbt an keinem Schaden.
func (p Partie) stirbtAnWiderstandskraft(objekt Spielobjekt) bool {
	_, widerstand := p.Werte(objekt)
	if !widerstand.Steht {
		return false
	}

	// **CR 704.5f**: null oder weniger.
	//
	// **Unzerstörbarkeit hilft hier nicht**, und das ist keine Auslassung:
	// CR 704.5f legt das Permanent in den Friedhof und **zerstört** es
	// nicht — CR 702.12b spricht nur vom Zerstören.
	if widerstand.Zahl <= 0 {
		return true
	}

	// **Unzerstörbares stirbt an keinem Schaden** — CR 702.12b nennt
	// CR 704.5g ausdrücklich (M14,
	// #1616).
	// **Und an keiner Todesberührung**: CR 704.5h zerstört ebenfalls.
	if p.HatFaehigkeit(objekt.Kennung, SchluesselUnzerstoerbar) {
		return false
	}

	// **CR 704.5h**: Schaden aus einer Quelle mit Todesberührung, gleich wie
	// viel — solange die Widerstandskraft über null liegt (M14).
	if objekt.Todesberuehrt {
		return true
	}

	// **CR 704.5g**: markierter Schaden **ab** der Widerstandskraft. Der
	// Schaden bleibt bis zum Aufräumschritt liegen (CR 514.2), und deshalb
	// entscheidet er zwischendurch über Leben und Tod.
	return objekt.Schaden >= widerstand.Zahl
}

// markeBeiNull sagt, ob diese Marke **dasteht** und bei null ist.
//
// # Warum „dasteht" dazugehört
//
// **Weil niemand sie ablegt.** CR 306.5b lässt einen Planeswalker mit so
// vielen Loyalitätsmarken hineinkommen, wie aufgedruckt sind — das ist ein
// **Ersetzungseffekt** (CR 614.1c) und braucht `M11`
// (#1613).
// Ohne diese Bedingung stürbe jeder Planeswalker in dem Augenblick, in dem er
// den Tisch berührt — eine Regel, die greift, wo es die Größe noch gar nicht
// gibt (derselbe Fall wie beim Startleben in `M17`).
func markeBeiNull(objekt Spielobjekt, art string) bool {
	anzahl, dasteht := objekt.Zaehler[art]

	return dasteht && anzahl <= 0
}
