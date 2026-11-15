// **Die dritte Hälfte des Vertragstests** — dieselbe Anwendung, dieselbe Frage.
//
// Eine eigene Datei, weil `bindungenlesen_test.go` mit ihr über 400 Zeilen
// ginge (§2). Der Schnitt folgt der Frage: dort stehen die Bindungen an
// **einer** Karte oder einem Set, hier die, die **über** ein Set auswerten —
// Suche, Statistik, Verteilung — und die Sprache, die über allem steht.
package app_test

import (
	"context"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

func TestJedeAuswertendeBindungErreichtIhreAnwendung(t *testing.T) {
	stand := bindungsprobe(t)
	anwendung, bindungen := stand.Anwendung, stand.Bindungen
	probe, statistik := stand.Suche, stand.Statistik

	t.Run("Suchen", func(t *testing.T) {
		antwort := bindungen.Suchen(t.Context(), "feuer", "name", "1")
		if antwort.Fehler != nil {
			t.Errorf("%+v", antwort.Fehler)
		}

		// **Die Argumente in ihrer Reihenfolge** — der Durchreicher ist genau
		// die Stelle, an der zwei Zeichenketten die Plätze tauschen.
		if probe.frage.Text != "feuer" || probe.frage.Field != "name" ||
			probe.frage.SetID != 1 {
			t.Errorf("%+v", probe.frage)
		}
	})

	t.Run("Suchfelder", func(t *testing.T) {
		antwort := bindungen.Suchfelder(t.Context(), "3")
		if antwort.Fehler != nil {
			t.Fatalf("%+v", antwort.Fehler)
		}

		if probe.frage.SetID != 3 || len(antwort.Felder) != 1 {
			t.Errorf("%d / %+v", probe.frage.SetID, antwort.Felder)
		}
	})

	t.Run("Statistik", func(t *testing.T) {
		antwort := bindungen.Statistik(t.Context(), "3")
		if antwort.Fehler != nil {
			t.Fatalf("%+v", antwort.Fehler)
		}

		if statistik.setID != 3 || antwort.Statistik.Karten != 12 {
			t.Errorf("%d / %+v", statistik.setID, antwort.Statistik)
		}
	})

	t.Run("Verteilung", func(t *testing.T) {
		antwort := bindungen.Verteilung(t.Context(), "3", "typ")
		if antwort.Fehler != nil {
			t.Fatalf("%+v", antwort.Fehler)
		}

		if statistik.feld != "typ" || len(antwort.Verteilung) != 1 {
			t.Errorf("%q / %+v", statistik.feld, antwort.Verteilung)
		}
	})

	t.Run("SpracheSetzen", func(t *testing.T) {
		var gesetzt bool

		anwendung.MitMenue(func(_ context.Context, _ *app.Menue) { gesetzt = true })

		if antwort := bindungen.SpracheSetzen(t.Context(), "en"); antwort.Fehler != nil {
			t.Fatalf("%+v", antwort.Fehler)
		}

		if !gesetzt {
			t.Error("die Menüleiste wurde nicht getauscht")
		}
	})
}
