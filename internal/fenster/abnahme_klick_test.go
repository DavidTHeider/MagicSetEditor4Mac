package fenster

// **Ein Doppelklick und ein Klick mit Umschalter** — P4
// (#2372):
// am Tisch löst ein Klick die Primär-, ein Doppelklick die Zweithandlung aus,
// und ⌘-Klick wählt. Ein Beleg am nativen Fenster braucht dafür Ereignisse,
// wie AppKit sie bei einer echten Maus liefert: **die Klickzahl zählt mit**
// (`clickCount` 1, dann 2 — daraus liest WebKit `detail` und `dblclick`),
// und der Umschalter steht an jedem Ereignis.

import (
	"slices"
	"testing"
)

func TestEinDoppelklickZaehltDieKlicksMit(t *testing.T) {
	t.Parallel()

	schritte, fehler := klickschritte(Ort{X: 5, Y: 6}, 2, nil)
	if fehler != nil {
		t.Fatal(fehler)
	}

	ort := Ort{X: 5, Y: 6}
	erwartet := []mausschritt{
		{art: artMausRunter, ort: ort, klicks: 1},
		{art: artMausHoch, ort: ort, klicks: 1},
		{art: artMausRunter, ort: ort, klicks: 2},
		{art: artMausHoch, ort: ort, klicks: 2},
	}

	if !slices.Equal(schritte, erwartet) {
		t.Errorf("Schritte %+v, erwartet %+v", schritte, erwartet)
	}
}

func TestEinKlickTraegtSeinenUmschalter(t *testing.T) {
	t.Parallel()

	schritte, fehler := klickschritte(Ort{X: 1, Y: 2}, 0, []string{"befehl", "umschalt"})
	if fehler != nil {
		t.Fatal(fehler)
	}

	if len(schritte) != 2 {
		t.Fatalf("ein Klick hat %d Schritte: %+v", len(schritte), schritte)
	}

	for _, schritt := range schritte {
		if schritt.klicks != 1 || schritt.schalter != schalterBefehl|schalterUmschalt {
			t.Errorf("Schritt %+v", schritt)
		}
	}
}

func TestEinKlickbefehlLiestAnzahlUndUmschalter(t *testing.T) {
	t.Parallel()

	befehl, fehler := befehlLesen([]byte(`{"was":"klicke","x":1,"y":2,"anzahl":2,"mit":["befehl"]}`))
	if fehler != nil || befehl.Anzahl != 2 || !slices.Equal(befehl.Mit, []string{"befehl"}) {
		t.Fatalf("%+v, %v", befehl, fehler)
	}
}

func TestEinKlickWeistUnbekanntesAb(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		anzahl int
		mit    []string
	}{
		"ein unbekannter Umschalter": {anzahl: 1, mit: []string{"hyper"}},
		"vier Klicks":                {anzahl: 4},
		"weniger als keiner":         {anzahl: -1},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, fehler := klickschritte(Ort{}, fall.anzahl, fall.mit); fehler == nil {
				t.Errorf("hingenommen: %+v", fall)
			}
		})
	}
}
