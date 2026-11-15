package domain

// **Was mit dem Zug geschieht, wenn jemand geht** — CR 800.4a und CR 800.4j
// (Fassung 2026-09-25),
// #1971.
//
// # Warum eine eigene Stelle
//
// **Die Reihenfolge, die einen Gegangenen überspringt, gibt es seit M13**
// ([Partie.naechsterSitz]) — sie greift, sobald jemand passt oder ein Zug
// endet. Was fehlte, ist der **Augenblick des Gehens**: hielt der Gehende die
// Priorität, behielt er sie; stand er in der Liste derer, die gepasst haben,
// zählte sein Passen weiter; war er am Zug, bekam er im nächsten Schritt
// wieder die Priorität. In einer Partie zu dritt konnte danach niemand mehr
// passen. `FuzzZugfolge` hat es beim Bau von QA-06b nachgestellt (#1967).
//
// **new** — MSE2 spielt nicht (§11); nachgelesen ist in der CR.

// zugNachDemGehen hält den Zug in Gang, nachdem Sitze gegangen sind.
//
// **Nur, solange die Partie läuft.** Bei zweien endet sie mit dem Gehen
// (CR 104.2a), und eine beendete Partie vergibt keine Priorität mehr.
func (p Partie) zugNachDemGehen() Partie {
	if p.Stand == StandBeendet || p.Zug.Nummer == 0 {
		return p
	}

	// **Gepasst haben nur, die noch im Spiel sind** — CR 117.4 meint mit
	// „alle" die Spieler im Spiel. Ein Passen des Gehenden zählte sonst für
	// einen, der nicht mehr gefragt werden kann, und die Runde endete, bevor
	// der letzte Verbliebene gepasst hat.
	gepasst := make([]string, 0, len(p.Zug.Gepasst))

	for _, wer := range p.Zug.Gepasst {
		if !p.istDraussen(wer) {
			gepasst = append(gepasst, wer)
		}
	}

	p.Zug.Gepasst = gepasst

	// **Wer mit der Priorität geht, gibt sie an den Nächsten in
	// Zugreihenfolge, der noch im Spiel ist** — CR 800.4a (Fassung
	// 2026-09-25).
	if p.Zug.Prioritaet != "" && p.istDraussen(p.Zug.Prioritaet) {
		p.Zug.Prioritaet = p.naechsterSitz(p.Zug.Prioritaet)
	}

	// **Und eine Abwurfpflicht geht mit dem, der sie hatte.** Geht der
	// aktive Spieler in seinem Zug, läuft der Zug ohne ihn zu Ende —
	// CR 800.4j (Fassung 2026-09-25). Seine Hand hat mit ihm das Spiel
	// verlassen (CR 800.4a), und abwerfen kann nur, wer am Zug ist; der
	// Aufräumschritt, der auf ihn wartete, stünde für immer.
	if p.Zug.Abwurf > 0 && p.istDraussen(p.Zug.Aktiv) {
		p.Zug.Abwurf = 0

		return p.Weiter()
	}

	return p
}
