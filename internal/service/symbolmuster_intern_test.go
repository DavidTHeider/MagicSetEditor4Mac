package service

import "testing"

// **Die Muster aus dem Pack, jedes gegen einen Text, den es treffen soll.**
//
// Innen geprüft und nicht über den Planbauer: die Frage ist, ob ein Ausdruck
// gelesen wird, und dafür ist ein Kartenplan Zeremonie. Die Bauform steht in
// `symbolmuster.go`; hier stehen die Fälle, die es im Pack wirklich gibt.
func TestEinMusterTrifftSoWeitWieMSEEsTrifft(t *testing.T) {
	faelle := []struct {
		name, code, text string
		laenge           int
	}{
		{"eine Zeichenklasse", "[XYZ]", "Y1", 1},
		{"eine Gruppe", "([XYZ0-9])/W", "3/W2", 3},
		{"Vorausschau greift nicht", "10(?!/)", "10", 2},
		{"Vorausschau greift", "10(?!/)", "10/W", 0},
		{"der Auffangausdruck, Ziffer", "[0-9](?!/[WUBRGCSTQ2])|.", "7x", 1},
		{"der Auffangausdruck, Hybridziffer", "[0-9](?!/[WUBRGCSTQ2])|.", "7/W", 1},
		{"der Auffangausdruck, fremdes Zeichen", "[0-9](?!/[WUBRGCSTQ2])|.", "&", 1},
		{"der längste Zweig gewinnt", "[0-9]+(?!/[W])|.", "123x", 3},
		{"ein maskierter Strich", `L/(\+[XYZ0-9][0-9]*)`, "L/+12", 5},
		{"ein Strich in der Klasse", "[a|b]", "b", 1},
		{"trifft gar nicht", "[XYZ]", "1", 0},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			gelesen, err := muster(fall.code)
			if err != nil {
				t.Fatalf("Muster %q: %v", fall.code, err)
			}

			laenge, _, passt := gelesen.treffer(fall.text)
			if passt != (fall.laenge > 0) || laenge != fall.laenge {
				t.Errorf("%q deckt von %q %d Zeichen (passt: %v), erwartet %d",
					fall.code, fall.text, laenge, passt, fall.laenge)
			}
		})
	}
}

// **Ein Muster, das diese Zerlegung nicht trifft, wird abgelehnt.**
//
// Und zwar **beim zweiten Mal genauso**: das Ergebnis ist gemerkt, damit ein
// Fehlschlag nicht bei jeder Karte neu bezahlt wird.
func TestEinUnlesbaresMusterWirdAbgelehnt(t *testing.T) {
	faelle := []struct{ name, code string }{
		{"Vorausschau mitten im Zweig", "a(?!b)c"},
		{"Vorausschau ohne schließende Klammer", "a(?!b"},
		{"eine Klammer, die nicht aufgeht", "a(?!b))"},
		{"kein gültiger Ausdruck", "["},
		{"eine Vorausschau, die kein Ausdruck ist", "a(?![)"},
		{"eine Klammer, die offen bleibt", "a(?!(b)"},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			for versuch := range 2 {
				if _, err := muster(fall.code); err == nil {
					t.Fatalf("Versuch %d: %q wurde angenommen", versuch, fall.code)
				}
			}
		})
	}
}

// **Ein gelesenes Muster wird beim zweiten Mal nicht neu übersetzt.**
//
// Die Gegenprobe zum Merken: dieselbe Zeichenkette muss dieselbe Übersetzung
// liefern, nicht eine gleichwertige zweite.
func TestEinGelesenesMusterWirdGemerkt(t *testing.T) {
	erst, err := muster("[QRS]")
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	zweit, err := muster("[QRS]")
	if err != nil {
		t.Fatalf("erneut lesen: %v", err)
	}

	if erst != zweit {
		t.Error("das Muster wurde ein zweites Mal übersetzt")
	}
}
