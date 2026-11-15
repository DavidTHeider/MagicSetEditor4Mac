package msescript

import (
	"testing"
)

// **Eine eingebaute Funktion, von Hand gerufen, ordnet auch ohne Umgebung**
// (#651).
//
// `Ruf.Umgebung` trägt die Umgebung der Rufstelle, und sie **darf** leer sein:
// wer `ordneListe` unmittelbar ruft, hat keine. Dann bleibt die Bibliothek.
//
// **Der Ordnungsausdruck ruft deshalb eine Bibliotheksfunktion.** Ein Rumpf
// aus lauter eigenen Namen liefe auch über einer Umgebung ohne alles, und der
// Test bliebe grün, ohne den Fall zu enthalten (§10).
//
// Er steht im Paket, weil die Frage im Paket liegt: von außen ist ein `Ruf`
// ohne Umgebung nicht zu bauen.
func TestOrdneListeOhneUmgebung(t *testing.T) {
	t.Parallel()

	liste := Wert{Art: ArtAbbildung, Eintraege: []Eintrag{
		{Name: "0", Wert: Text("bb")},
		{Name: "1", Wert: Text("a")},
	}}

	rumpf, err := Zerteile("length(input)")
	if err != nil {
		t.Fatalf("der Ordnungsausdruck ließ sich nicht zerteilen: %v", err)
	}

	geordnet, err := ordneListe(t.Context(), Ruf{
		Stellung: []Wert{liste},
		Benannt: []Eintrag{{Name: "order_by", Wert: Wert{
			Art:      ArtFunktion,
			Funktion: &Funktionsteil{Rumpf: rumpf},
		}}},
	})
	if err != nil {
		t.Fatalf("unerwarteter Fehler: %v", err)
	}

	if geordnet.AlsText() != "abb" {
		t.Errorf("Ergebnis = %q, erwartet \"abb\"", geordnet.AlsText())
	}
}
