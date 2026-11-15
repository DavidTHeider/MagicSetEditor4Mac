package domain

// **Was von selbst geschieht** — Werkstück G9 von Phase 5.5
// (#1592).
//
// # Die Reihenfolge ist die Regel
//
// **CR 703.3** (Fassung 2026-08-19): beginnt ein Abschnitt, der zugbasierte
// Handlungen hat, laufen sie **zuerst** — vor den zustandsbasierten Aktionen,
// vor den wartenden Auslösern, **vor der Priorität**. Beides gibt es hier
// nicht (Phase 5.8); die Reihenfolge steht trotzdem, weil 5.8 sie braucht.
//
// # Was hier gebaut ist und was ausdrücklich nicht
//
// **CR 703.4 führt fünfzehn** zugbasierte Handlungen (703.4a–q, ohne `l` und
// `o` — nachgezählt am 2026-09-17). Dieses Werkstück baut **drei** davon plus
// das Enttappen:
//
//   - **Enttappen** (CR 502.3) — es steht in 703.4 nicht, obwohl 502.3 es
//     eine zugbasierte Handlung nennt; deshalb stehen beide Nummern da.
//   - **Ziehen** (CR 703.4d, CR 504.1, CR 121.1).
//   - **Abwerfen auf die Handgrenze** (CR 703.4n, CR 514.1).
//   - **Schaden entfernen** (CR 703.4p, CR 514.2) — und zwar **nur** die
//     eine Hälfte davon: die andere lässt „bis zum Ende des Zuges"-Effekte
//     enden, und die brauchen CR 611. Das ist Phase 5.8, und es steht hier
//     als Grenze und nicht als Auslassung (§10).
//
// **Nicht gebaut**: Phasing (CR 502.1, 703.4a — es braucht CR 702.26), die
// Tag/Nacht-Prüfung (703.4b/c), Saga-Marken (703.4f) und das Leeren des
// Manavorrats (CR 500.5, 703.4q) — Mana gibt es hier nicht.

import "fmt"

// zughandlungen führt die Handlungen des Abschnitts aus, in den gerade
// gewechselt wurde.
func (p Partie) zughandlungen() Partie {
	switch p.Zug.Abschnitt.Schritt {
	case SchrittEnttappen:
		return p.enttappen()

	case SchrittZiehen:
		return p.zugZiehen()

	case SchrittKampfschaden:
		return p.kampfschadenAlsHandlung()

	case SchrittAufraeumen:
		return p.aufraeumen()

	// **Zu Beginn des Endschritts zieht der Monarch** — CR 725.2 (M9,
	// #1611).
	case SchrittEnde:
		return p.monarchZiehtImEndschritt()
	}

	// **Und zu Beginn der ersten Hauptphase die Strahlung** — CR 728.1.
	if p.Zug.Abschnitt.Phase == PhaseHauptVorKampf {
		return p.strahlungInDerHauptphase()
	}

	return p
}

// kampfschadenAlsHandlung teilt den Kampfschaden aus — CR 703.4m.
//
// # Warum hier und nicht auf Zuruf
//
// **Weil CR 510.1 und CR 510.2 zugbasierte Handlungen sind**: sie geschehen,
// sobald der Schritt beginnt, und **niemand kommt dazwischen** (CR 510.2
// sagt es ausdrücklich). Ein Befehl „jetzt Schaden austeilen" wäre eine Wahl,
// wo die Regeln keine lassen.
//
// **Zweimal, wo jemand Erstschlag hat** (CR 510.4): der Schritt wiederholt
// sich, und beim zweiten Mal steht [Kampfstand.Schadensschritte] auf eins.
//
// # Warum hier kein Wächter gegen ein drittes Mal steht
//
// **Weil der Schritt höchstens zweimal betreten wird, und das entscheidet
// eine andere Stelle.** [Partie.abschnittBetreten] läuft je Abschnitt
// **einmal**; wiederholt wird der Schadensschritt allein von
// [Partie.abschnittVerlassen], und das tut es nur, solange
// [Partie.zweiterSchadensschritt] wahr ist — also genau einmal.
//
// **Ein Wächter hier wäre ein Zweig, den kein Weg betritt** (§10). Die
// Schranke steht dort, wo sie entschieden wird, und nicht zweimal.
func (p Partie) kampfschadenAlsHandlung() Partie {
	return p.kampfschaden()
}

// enttappen enttappt die Permanenten des aktiven Spielers — CR 502.3.
//
// **Seine und nur seine.** CR 502.3 sagt, dass gewöhnlich **alle** seine
// Permanenten enttappen; Effekte, die welche festhalten, gibt es hier nicht.
//
// **Und nur auf dem Tisch**: getappt ist ein Zustand von Permanenten
// (CR 701.26a), und Permanente liegen auf dem Schlachtfeld (CR 110.1).
func (p Partie) enttappen() Partie {
	objekte := make([]Spielobjekt, len(p.Objekte))
	copy(objekte, p.Objekte)

	for stelle := range objekte {
		if objekte[stelle].Beherrscher == p.Zug.Aktiv &&
			objekte[stelle].Zone == ZoneTisch {
			objekte[stelle].Getappt = false
		}
	}

	p.Objekte = objekte

	return p
}

// zugZiehen lässt den aktiven Spieler eine Karte ziehen — CR 504.1.
//
// **Wer anfängt, überspringt sein erstes Ziehen** — CR 103.8a, und zwar
// ausdrücklich für das Spiel **zu zweit**. Bei mehr als zwei zieht auch der
// Starter; bei einem Sitz gibt es keinen Gegner, und die Regel greift nicht.
//
// **Eine leere Bibliothek hält nichts an** — sie wird vermerkt. Wer daraus
// ziehen wollte, verliert bei der nächsten zustandsbasierten Prüfung
// (CR 704.5b, [Partie.Zustandspruefung]); der Zug läuft bis dahin weiter.
func (p Partie) zugZiehen() Partie {
	if p.Zug.Nummer == 1 && p.Zug.Aktiv == p.Zug.Starter && len(p.Sitze) == 2 {
		return p
	}

	return p.ziehtEine(p.Zug.Aktiv)
}

// ziehtEine legt die oberste Karte der Bibliothek auf die Hand — CR 121.1.
//
// **Oben ist die kleinste Ordnung**: `Sandkasten.platzFuer` gibt für „oben"
// die Eins, und der Aufbau legt die Bibliothek mit 1 beginnend.
func (p Partie) ziehtEine(wer string) Partie {
	oberste := -1

	for stelle, objekt := range p.Objekte {
		if objekt.Beherrscher != wer || objekt.Zone != ZoneBibliothek {
			continue
		}

		if oberste < 0 || objekt.Ordnung < p.Objekte[oberste].Ordnung {
			oberste = stelle
		}
	}

	// **Ins Leere gezogen** — CR 704.5b: nicht der Verlust, sondern sein
	// Grund. Was daraus wird, entscheidet [Partie.Zustandspruefung].
	if oberste < 0 {
		return p.merkeLeergezogen(wer)
	}

	// **Eine gezogene Karte ist ein neues Objekt** — CR 400.7 (M18, #1619).
	p = p.zonenwechsel(oberste, ZoneHand)

	return p
}

// aufraeumen nimmt den Schaden weg und stellt die Abwurfpflicht fest.
//
// **Zwei Handlungen, und die zweite ist eine Wahl.** CR 514.1 sagt, dass der
// aktive Spieler auf die Handgrenze abwirft — **welche** Karten, entscheidet
// er. Ein Programm, das selbst welche wegwürfe, träfe eine Entscheidung, die
// ihm nicht gehört; deshalb steht hier eine **Pflicht**, und der Abschnitt
// wartet, bis sie erfüllt ist.
//
// **Der Schaden geht in jedem Fall weg** (CR 514.2) — er hängt an keiner
// Wahl.
func (p Partie) aufraeumen() Partie {
	objekte := make([]Spielobjekt, len(p.Objekte))
	copy(objekte, p.Objekte)

	for stelle := range objekte {
		objekte[stelle].Schaden = 0

		// **Und mit ihm die Todesberührung** — CR 704.5h fragt nach
		// Schaden „seit der letzten Prüfung", und im Aufräumschritt geht
		// jeder weg (CR 514.2, M14).
		objekte[stelle].Todesberuehrt = false
	}

	p.Objekte = objekte
	p.Zug.Abwurf = max(0, p.handzahl(p.Zug.Aktiv)-Handgrenze)

	// **Und die zweite Hälfte von CR 514.2** — `G9` hat sie ausdrücklich
	// offen gelassen, und `M10`
	// (#1612)
	// baut sie: was „bis zum Ende des Zuges" gilt, endet hier.
	// **Und die Ersetzungen ebenso** — CR 514.2 mit CR 614.5 (M11,
	// #1613):
	// was mit dem Zug endet, endet, und was bleibt, bekommt seine
	// Gelegenheit zurück.
	return p.ohneEffekteBisZugende().ohneErsetzungenBisZugende()
}

// handzahl zählt, was dieser Sitz auf der Hand hält.
func (p Partie) handzahl(wer string) int {
	anzahl := 0

	for _, objekt := range p.Objekte {
		if objekt.Beherrscher == wer && objekt.Zone == ZoneHand {
			anzahl++
		}
	}

	return anzahl
}

// Abwerfen erfüllt einen Schritt der Abwurfpflicht — CR 514.1.
//
// **Die Wahl gehört dem Spieler**, und deshalb nennt er die Karte. Erfüllt
// ist die Pflicht, wenn keine mehr offen ist; dann geht der Zug weiter.
func (p Partie) Abwerfen(wer, objekt string) (Partie, error) {
	if p.Zug.Abwurf == 0 {
		return p, fmt.Errorf(
			"%w: es ist nichts abzuwerfen — die Hand ist nicht über der "+
				"Grenze von %d", ErrValidation, Handgrenze)
	}

	if wer != p.Zug.Aktiv {
		return p, fmt.Errorf(
			"%w: %q wirft nicht ab — im Aufräumschritt tut es, wer am Zug ist",
			ErrValidation, wer)
	}

	stelle := -1

	for lauf, eines := range p.Objekte {
		if eines.Kennung == objekt && eines.Beherrscher == wer &&
			eines.Zone == ZoneHand {
			stelle = lauf

			break
		}
	}

	if stelle < 0 {
		return p, fmt.Errorf(
			"%w: %q liegt nicht auf der Hand von %q", ErrValidation, objekt, wer)
	}

	// **Und eine abgeworfene ebenso** — CR 400.7 (M18, #1619).
	p = p.zonenwechsel(stelle, ZoneFriedhof)
	p.Zug.Abwurf--

	if p.Zug.Abwurf > 0 {
		return p, nil
	}

	return p.Weiter(), nil
}

// merkeLeergezogen vermerkt den Griff in eine leere Bibliothek — CR 704.5b.
//
// **Nur an einem Sitz, der einer ist.** Wer nicht am Tisch sitzt, zieht auch
// nicht; ein Vermerk an niemandem wäre ein Zustand ohne Träger.
func (p Partie) merkeLeergezogen(wer string) Partie {
	for stelle := range p.Sitze {
		if p.Sitze[stelle].Kennung != wer || p.Sitze[stelle].Leergezogen {
			continue
		}

		sitze := make([]Sitz, len(p.Sitze))
		copy(sitze, p.Sitze)
		sitze[stelle].Leergezogen = true
		p.Sitze = sitze

		return p
	}

	return p
}
