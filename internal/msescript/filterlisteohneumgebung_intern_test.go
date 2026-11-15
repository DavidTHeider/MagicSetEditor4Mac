package msescript

import (
	"testing"
)

// **Wer die eingebaute Funktion von Hand ruft, gibt keine Umgebung mit.**
//
// Dann bleibt die Bibliothek — ein Rumpf ohne eigene Namen siebt damit
// genauso. Dieselbe Regel wie bei [schluesselVon], und derselbe Grund: eine
// Funktion, die ohne Umgebung abstürzte, wäre in jedem Test ein Sonderfall
// (#730).
//
// **Der Test steht im Paket**, weil der Weg nur von hier erreichbar ist: über
// eine Quelle ausgewertet gibt es immer eine Umgebung.
func TestFilterListOhneUmgebung(t *testing.T) {
	t.Parallel()

	liste := Wert{Art: ArtAbbildung, Eintraege: []Eintrag{
		{Name: "0", Wert: Zahl(1)},
		{Name: "1", Wert: Zahl(9)},
	}}

	gesiebt, err := siebeListe(t.Context(), Ruf{
		Stellung: []Wert{liste},
		Benannt: []Eintrag{{Name: "filter", Wert: Wert{
			Art:      ArtFunktion,
			Funktion: &Funktionsteil{Rumpf: gelesenerAusdruckOderNil("input > 5")},
		}}},
	})
	if err != nil {
		t.Fatalf("sieben: %v", err)
	}

	if len(gesiebt.Eintraege) != 1 || gesiebt.Eintraege[0].Wert.AlsText() != "9" {
		t.Errorf("gesiebt wurde %+v", gesiebt.Eintraege)
	}
}

// gelesenerAusdruckOderNil zerteilt einen Ausdruck für den Rumpf.
func gelesenerAusdruckOderNil(quelle string) *Knoten {
	knoten, _ := Zerteile(quelle) //nolint:errcheck // ein unlesbarer Rumpf ist hier nil

	return knoten
}
