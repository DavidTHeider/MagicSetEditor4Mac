package msescript

import "testing"

// **Wer `position` von Hand ruft, gibt keine Umgebung mit.** Dann bleibt die
// Bibliothek, und die Ordnung wird bei jeder Frage neu gebaut — dieselbe Regel
// wie bei `sort_list` ([schluesselVon]). Der Test steht im Paket, weil über
// eine Quelle ausgewertet immer eine Umgebung da ist.
func TestPositionOhneUmgebung(t *testing.T) {
	t.Parallel()

	karte := func(kennung, name string) Wert {
		return Wert{
			Art: ArtAbbildung, Kennung: kennung,
			Eintraege: []Eintrag{{Name: "name", Wert: Text(name)}},
		}
	}

	hinten, vorn := karte("karte:1", "b"), karte("karte:2", "a")
	set := Wert{
		Art: ArtAbbildung, Kennung: "set",
		Eintraege: []Eintrag{{Name: "0", Wert: hinten}, {Name: "1", Wert: vorn}},
	}

	stelle, err := stelleVon(t.Context(), Ruf{
		Benannt: []Eintrag{
			{Name: "of", Wert: hinten},
			{Name: "in", Wert: set},
			{Name: "order_by", Wert: Wert{
				Art:      ArtFunktion,
				Funktion: &Funktionsteil{Rumpf: gelesenerAusdruckOderNil("card.name")},
			}},
		},
	})
	if err != nil {
		t.Fatalf("position: %v", err)
	}

	if stelle.AlsText() != "1" {
		t.Errorf("Stelle %s, erwartet 1", stelle.AlsText())
	}
}
