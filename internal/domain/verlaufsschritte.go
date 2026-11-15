package domain

// **Der Verlauf als Schritte** — Werkstück U6 von Phase 5.1
// (#1417).
//
// Der Bestand führt **Einträge**, je Karte und Feld einen; zurückgenommen
// wird aber ein **Schritt** — eine Massenänderung als Ganzes
// ([Verlaufseintrag.Gruppe]). Wer den Verlauf liest, will dieselbe Einheit
// sehen, die ⌘Z bewegt: dreihundert Zeilen für eine Änderung erklärten
// nichts.
//
// **Herkunft: new.** MSE2 hält seinen Verlauf im `ActionStack`
// (`src/util/action_stack.hpp`, Fassung f9d9356 vom 2021-07-06) — Aktionen mit
// Namen, nicht gespeichert, ohne Vorher und Nachher je Feld; eine Liste
// davon zeigt es nirgends.

// Verlaufsschritte fasst die Einträge zu Schritten zusammen — in der
// Reihenfolge des Verlaufs, der jüngste zuletzt.
//
// **Eine Gruppe ist ein Schritt, ein Eintrag ohne Gruppe einer für sich.**
// Der **erste** Eintrag einer Gruppe vertritt sie, so wie bei der Rücknahme
// ([Verlaufsschritt]), und `Umfang` zählt ihre Karten.
//
// **Bis #2396 war es der jüngste, und gezählt wurden Einträge.** Bei einer
// Massenänderung ist beides dasselbe — ein Feld, je Karte ein Eintrag. Seit
// der Schreibweg nachzieht, wer ein Feld liest, steht das bearbeitete Feld
// zuerst und seine Leser danach, an derselben Karte: der jüngste nannte das
// zuletzt nachgezogene Feld mit dessen Werten, und die Zählung machte aus
// einer Karte zwei.
func Verlaufsschritte(eintraege []Verlaufseintrag) []Verlaufsschritt {
	schritte := make([]Verlaufsschritt, 0, len(eintraege))
	stelle := map[string]int{}
	karten := map[string]map[int64]bool{}

	for _, eintrag := range eintraege {
		if eintrag.Gruppe != "" {
			if i, da := stelle[eintrag.Gruppe]; da {
				karten[eintrag.Gruppe][eintrag.KarteID] = true
				schritte[i].Umfang = len(karten[eintrag.Gruppe])

				continue
			}

			stelle[eintrag.Gruppe] = len(schritte)
			karten[eintrag.Gruppe] = map[int64]bool{eintrag.KarteID: true}
		}

		schritte = append(schritte, Verlaufsschritt{Verlaufseintrag: eintrag, Umfang: 1})
	}

	return schritte
}
