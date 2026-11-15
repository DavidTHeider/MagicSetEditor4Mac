package app_test

// **Der dritte Teil des Vertragstests** — dieselbe Anwendung, dieselbe
// Frage, und die Bindungen, die ein **Bild** zurückgeben.
//
// Eine eigene Datei, weil `bindungenlesen_test.go` mit ihnen über 400 Zeilen
// ginge (§2). Der Schnitt folgt der Frage: dort stehen die Bindungen über den
// Bestand, hier die, die etwas zeichnen — Vorschau, Probevorschau und die
// Wahlbilder einer Stiloption.

import "testing"

func TestJedeZeichnendeBindungErreichtIhreAnwendung(t *testing.T) {
	stand := bindungsprobe(t)
	bindungen := stand.Bindungen

	t.Run("Probevorschau", func(t *testing.T) {
		antwort := bindungen.Probevorschau(t.Context(), "1", "name", "Eisregen", 1)
		if antwort.Fehler != nil {
			t.Fatalf("%+v", antwort.Fehler)
		}

		if antwort.Vorschau == nil || antwort.Vorschau.Breite == 0 {
			t.Errorf("%+v", antwort.Vorschau)
		}
	})

	t.Run("Stilprobevorschau", func(t *testing.T) {
		antwort := bindungen.Stilprobevorschau(t.Context(), "1", "border color", "rot", 1)
		if antwort.Fehler != nil {
			t.Fatalf("%+v", antwort.Fehler)
		}

		if antwort.Vorschau == nil || antwort.Vorschau.Breite == 0 {
			t.Errorf("%+v", antwort.Vorschau)
		}
	})

	// **Die Wahlbilder einer Stiloption** (#1225) — der Planbauer dieser
	// Vorrichtung gibt keine, und die Bindung reicht das durch.
	t.Run("Stilwahlbilder", func(t *testing.T) {
		antwort := bindungen.Stilwahlbilder(t.Context(), "1", "tap symbol")
		if antwort.Fehler != nil {
			t.Fatalf("%+v", antwort.Fehler)
		}

		if antwort.Bilder == nil {
			t.Error("die Liste ist nil statt leer — über die Bindung käme `null` an")
		}
	})

	t.Run("Vorschau", func(t *testing.T) {
		antwort := bindungen.Vorschau(t.Context(), "1", 1)
		if antwort.Fehler != nil {
			t.Fatalf("%+v", antwort.Fehler)
		}

		if antwort.Vorschau == nil || antwort.Vorschau.Breite == 0 {
			t.Errorf("%+v", antwort.Vorschau)
		}
	})
}

// **Die Marken erreichen ihre Anwendung** — Werkstück 3 von Phase 3.5
// (#1208).
//
// **Hier und nicht bei den lesenden**: beide Wege **schreiben**, und die
// lesenden gibt es am Fenster nicht — die Marken reisen mit den Zeilen der
// Kartenliste.
func TestDieMarkenbindungenErreichenIhreAnwendung(t *testing.T) {
	bindungen := bindungsprobe(t).Bindungen

	if antwort := bindungen.MarkeSetzen(t.Context(), "1", "1", "fertig"); antwort.Fehler != nil {
		t.Errorf("MarkeSetzen: %+v", antwort.Fehler)
	}

	if antwort := bindungen.MarkeEntfernen(t.Context(), "1", "1", "fertig"); antwort.Fehler != nil {
		t.Errorf("MarkeEntfernen: %+v", antwort.Fehler)
	}
}
