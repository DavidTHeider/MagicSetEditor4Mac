package fenster

// **Ein Fenster, das auf den Bildschirm passt** — UX-14.AK-04
// (#1686):
// *„Gespeicherte Positionen werden bei Monitorwechsel plausibilisiert."*
// Gemerkt wird die Größe; wer sie auf einem großen Bildschirm gezogen hat und
// ihn abzieht, bekommt sie auf dem kleinen geklemmt statt über den Rand.

import "testing"

func TestDasFensterPasstAufDenBildschirm(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		breite, hoehe               int
		flaecheBreite, flaecheHoehe int
		wirdBreite, wirdHoehe       int
		geaendert                   bool
	}{
		"passt schon":      {1280, 860, 1512, 944, 1280, 860, false},
		"zu breit":         {2400, 860, 1512, 944, 1512, 860, true},
		"zu hoch":          {1280, 1300, 1512, 944, 1280, 944, true},
		"beides":           {2560, 1440, 1512, 944, 1512, 944, true},
		"Fläche unbekannt": {2560, 1440, 0, 0, 2560, 1440, false},
		"genau so groß":    {1512, 944, 1512, 944, 1512, 944, false},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			breite, hoehe, geaendert := passend(
				fall.breite, fall.hoehe, fall.flaecheBreite, fall.flaecheHoehe)

			if breite != fall.wirdBreite || hoehe != fall.wirdHoehe ||
				geaendert != fall.geaendert {
				t.Errorf("es kam %d × %d (%v)", breite, hoehe, geaendert)
			}
		})
	}
}

// **Die Arbeitsfläche kommt in Bildpunkten, das Fenster misst in Punkten** —
// nachgelesen in Wails v3.0.0-beta.23 (`pkg/application/screen_darwin.go`,
// `cScreenToScreen`): `getScreenForWindow` nimmt `visibleFrame` mal
// `backingScaleFactor` und teilt nicht zurück. Ohne die Umrechnung griff die
// Klemmung auf Retina nie — angestrichen vom Zweitblick über Phase 6.3 (S2).
func TestDieArbeitsflaecheKommtInPunkten(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		breite, hoehe         int
		faktor                float32
		wirdBreite, wirdHoehe int
	}{
		"Retina":           {2940, 1840, 2, 1470, 920},
		"ohne Skalierung":  {1920, 1080, 1, 1920, 1080},
		"Faktor unbekannt": {1920, 1080, 0, 1920, 1080},
		"krummer Faktor":   {3024, 1890, 1.5, 2016, 1260},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			breite, hoehe := inPunkten(fall.breite, fall.hoehe, fall.faktor)
			if breite != fall.wirdBreite || hoehe != fall.wirdHoehe {
				t.Errorf("es kam %d × %d", breite, hoehe)
			}
		})
	}
}

// **Und nicht unter das Mindestmaß** (G7 desselben Zweitblicks): Wails senkt
// es sonst still mit `SetSize`, und das Fenster ließe sich kleiner ziehen,
// als die Oberfläche trägt.
func TestDieKlemmungBleibtAmMindestmass(t *testing.T) {
	t.Parallel()

	breite, hoehe, geaendert := passend(1280, 860, 900, 600)
	if breite != mindestbreite || hoehe != mindesthoehe || !geaendert {
		t.Errorf("es kam %d × %d (%v)", breite, hoehe, geaendert)
	}
}
