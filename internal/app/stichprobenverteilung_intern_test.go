//go:build fremdpakete

// Die Stichprobe erreicht das **Ende** ihrer Gruppe
// (#536).
package app

import "testing"

// TestDieStichprobeErreichtDasEnde ist die Gegenprobe zum Befund des
// Zweitblicks.
//
// **Bei 33 Karten und zwölf Plätzen** war der ganzzahlige Schritt 2, und die
// Stichprobe endete bei Stelle 22 — die letzten zehn Karten waren
// unerreichbar. Genau diese Größen hat die Planeswalker-Gruppe von `Red`.
func TestDieStichprobeErreichtDasEnde(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name    string
		karten  int
		anzahl  int
		erwarte int
	}{
		{"33 Karten, 12 Plätze — die Planeswalker von Red", 33, 12, 32},
		{"29 Karten, 12 Plätze", 29, 12, 28},
		{"3761 Karten, 168 Plätze", 3761, 168, 3760},
		{"gleich viele wie Plätze", 12, 12, 11},
		{"weniger als Plätze", 5, 12, 4},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			karten := make([]KarteDTO, fall.karten)
			for i := range karten {
				karten[i] = KarteDTO{ID: kennung(i)}
			}

			gewaehlt := verteilteStichprobe(karten, fall.anzahl)
			if len(gewaehlt) == 0 {
				t.Fatal("leere Stichprobe")
			}

			if gewaehlt[0].ID != kennung(0) {
				t.Errorf("erste gezogene Karte %q, erwartet %q",
					gewaehlt[0].ID, kennung(0))
			}

			letzte := gewaehlt[len(gewaehlt)-1].ID
			if letzte != kennung(fall.erwarte) {
				t.Errorf("letzte gezogene Karte %q, erwartet %q — das Ende der "+
					"Gruppe wird nicht erreicht", letzte, kennung(fall.erwarte))
			}
		})
	}
}

// TestDieStichprobeZiehtOhneWiederholung hält fest, dass keine Karte doppelt
// kommt.
//
// **Ohne diese Zusage täuschte die Zahl.** Ein Lauf über „zwölf Karten", von
// denen drei dieselbe sind, misst neun.
func TestDieStichprobeZiehtOhneWiederholung(t *testing.T) {
	t.Parallel()

	karten := make([]KarteDTO, 33)
	for i := range karten {
		karten[i] = KarteDTO{ID: kennung(i)}
	}

	gewaehlt := verteilteStichprobe(karten, 12)

	gesehen := map[string]bool{}
	for i := range gewaehlt {
		if gesehen[gewaehlt[i].ID] {
			t.Errorf("Karte %q kommt zweimal vor", gewaehlt[i].ID)
		}

		gesehen[gewaehlt[i].ID] = true
	}

	if len(gewaehlt) != 12 {
		t.Errorf("%d Karten gezogen, erwartet 12", len(gewaehlt))
	}
}

// kennung macht aus einer Stelle eine Kartenkennung.
func kennung(stelle int) string {
	return string(rune('A'+stelle/26)) + string(rune('a'+stelle%26))
}
