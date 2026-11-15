package app

import "testing"

// **Eine Leiste ohne Fenstermenü** bekommt das Objektmenü ans Ende — der Fall
// entsteht aus [MenueleisteMitStand] nicht, die immer eines baut; geprüft wird
// hier, dass der Rückfall nicht ins Leere greift.
func TestEinObjektmenueOhneFenstermenueStehtAmEnde(t *testing.T) {
	t.Parallel()

	leiste := &Menue{Eintraege: []*Menueeintrag{{Beschriftung: "Ablage", Untermenue: &Menue{}}}}
	objektmenueEinsetzen(leiste, Objekthandlungen{
		Titel: "Karte", Art: "karte", Kennung: "7", Epoche: 1,
		Eintraege: []Objekthandlung{{Schluessel: "a", Name: "A"}},
	}, func(Objektausloesung) {})

	if len(leiste.Eintraege) != 2 || leiste.Eintraege[1].Beschriftung != "Karte" {
		t.Fatalf("die Leiste trägt %+v", leiste.Eintraege)
	}
}
