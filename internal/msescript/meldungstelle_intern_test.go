package msescript

// **Eine Meldung ohne Umgebung geht verloren, und hält nichts an** —
// Werkstück A von Phase 4.9
// (#1324).
//
// # Warum das ein interner Test ist
//
// **Weil es von außen nicht geht.** `AuswertenQuelle` reicht immer eine
// Umgebung; der Fall entsteht nur, wenn jemand eine eingebaute Funktion **von
// Hand** ruft — genau der Fall, den der Kopf von [Ruf] als erlaubt führt.
// Ein Zweig, den kein Test betritt, ist keine Zusage (§3).

import (
	"testing"
)

func TestEineMeldungOhneUmgebungGehtVerloren(t *testing.T) {
	t.Parallel()

	wert, err := meldeHinweis(t.Context(), Ruf{Stellung: []Wert{Text("hallo")}})
	if err != nil {
		t.Fatalf("melden: %v", err)
	}

	if wert.AlsText() != "hallo" {
		t.Errorf("herauskam %q, erwartet %q", wert.AlsText(), "hallo")
	}
}
