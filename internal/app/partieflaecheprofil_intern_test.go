package app

// **Die Fläche ohne Fenster kennt die Profile** — UXR-F21
// (#2041).
//
// # Wogegen hier geprüft wird
//
// **Dass `PartieflaecheUeber` dieselbe Auskunft gibt wie das Fenster**, wenn
// ein Sitz an ein Profil gebunden wird, das es nicht gibt: die Partie
// entsteht, und die Antwort nennt die Bindung (G8, §8). Bis dahin hatte die
// Fläche weder Verwaltung noch Kenner, [App.profilbefunde] gab nichts, und
// eine Bindung ins Leere sah aus wie keine Bindung.
//
// **Die Fälle des Befunds selbst** stehen in `partieprofil_intern_test.go`
// gegen Doppel; hier läuft die Verdrahtung gegen einen echten Bestand — dort
// ist sie, was ein Doppel nicht prüft.

import "testing"

func TestDieFlaecheOhneFensterNenntEinVerschwundenesProfil(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		profilID string
		befunde  int
	}{
		"ohne Bindung":                  {profilID: "", befunde: 0},
		"ein Profil, das es gibt":       {profilID: "1", befunde: 0},
		"ein Profil, das es nicht gibt": {profilID: "987654", befunde: 1},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			db := bestandFuerDenRiegel(t)
			deckID := einDeckImBestand(t, db)

			antwort := PartieflaecheUeber(db).PartieAusWuenschen(
				riegelkontext(t), "magic", "anna", []Sitzwunsch{
					{Kennung: "anna", Name: "Anna", DeckID: deckID},
					{
						Kennung: "bert", Name: "Bert", DeckID: deckID,
						ProfilID: fall.profilID,
					},
				})

			if antwort.Fehler != nil {
				t.Fatalf("PartieAusWuenschen: %+v", antwort.Fehler)
			}

			// **Die Partie ist da, gleich was die Bindung sagt** — ein
			// Befund ist kein Fehler (§8).
			if antwort.Partie.ID == "" || len(antwort.Partie.Sitze) != 2 {
				t.Errorf("die Partie fehlt: %+v", antwort.Partie)
			}

			if len(antwort.Befunde) != fall.befunde {
				t.Errorf("%d Befunde %q, erwartet %d",
					len(antwort.Befunde), antwort.Befunde, fall.befunde)
			}
		})
	}
}
