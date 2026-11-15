package msescript

import "testing"

// **Von Hand gerufen, ohne Umgebung, zählt sie auch** — dieselbe Zusage wie
// `TestOrdneListeOhneUmgebung`: `Ruf.Umgebung` darf leer sein, dann bleibt die
// Bibliothek. Der Filter ruft deshalb eine Bibliotheksfunktion.
func TestZahlImSetOhneUmgebung(t *testing.T) {
	t.Parallel()

	karte := Wert{Art: ArtAbbildung, Kennung: "karte:1", Eintraege: []Eintrag{
		{Name: "name", Wert: Text("Anton")},
	}}
	set := Wert{Art: ArtAbbildung, Kennung: "set", Eintraege: []Eintrag{{Name: "0", Wert: karte}}}

	rumpf, err := Zerteile(`length(card.name) == 5`)
	if err != nil {
		t.Fatalf("zerteilen: %v", err)
	}

	gezaehlt, err := zahlImSet(t.Context(), Ruf{
		Stellung: []Wert{set},
		Benannt: []Eintrag{{Name: "filter", Wert: Wert{
			Art: ArtFunktion, Funktion: &Funktionsteil{Rumpf: rumpf},
		}}},
	}, set)
	if err != nil || gezaehlt.AlsText() != "1" {
		t.Errorf("gezählt %s (%v), erwartet 1", gezaehlt.AlsText(), err)
	}
}
