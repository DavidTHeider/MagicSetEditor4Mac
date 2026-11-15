package app_test

// **Die Decks eines anderen Profils zur Wahl, an der Fassade** — UX-08.AK-04
// (#1680).
//
// **Dieselben Ausgänge wie die Übernahme**: ohne Naht, krumme Kennung, ein
// Nein der Naht — und der Erfolg. Die Liste ist nie `null`.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

func TestJederAusgangDerDeckwahlMeldetSich(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		anwendung func() *app.App
		profil    string
	}{
		"ohne Naht": {
			anwendung: func() *app.App { return app.NeueApp("") },
			profil:    "2",
		},
		"krumme Profilkennung": {
			anwendung: func() *app.App {
				mit := app.NeueApp("")
				mit.MitDeckuebernahme(&uebernahmedoppel{})

				return mit
			},
			profil: "keine Zahl",
		},
		"ein Nein der Naht": {
			anwendung: func() *app.App {
				mit := app.NeueApp("")
				mit.MitDeckuebernahme(&uebernahmedoppel{
					fehler: errors.New("dafür fehlt dir die Rolle"),
				})

				return mit
			},
			profil: "2",
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			antwort := fall.anwendung().DeckeZurUebernahme(t.Context(), fall.profil)

			switch {
			case antwort.Fehler == nil:
				t.Fatal("es kam kein Fehler")
			case antwort.Decks == nil:
				t.Error("die Liste kam als null")
			}
		})
	}
}

func TestDieDeckwahlNenntDieDecksVonDrueben(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	anwendung.MitDeckuebernahme(&uebernahmedoppel{})

	antwort := anwendung.DeckeZurUebernahme(t.Context(), "2")

	switch {
	case antwort.Fehler != nil:
		t.Fatalf("Fehler: %+v", antwort.Fehler)
	case len(antwort.Decks) != 1 || antwort.Decks[0].Name != "Nachtlauf" ||
		antwort.Decks[0].ID != "7":
		t.Errorf("es kam %+v", antwort.Decks)
	}
}

// **Und der Flächenweg reicht sie durch** — für MCP und Kommandozeile.
func TestDieDeckwahlUeberDieFlaeche(t *testing.T) {
	t.Parallel()

	// **Im frischen Bestand gibt es kein fremdes Profil**: das Nein kommt
	// von hinter der Fläche, und die Liste ist trotzdem nicht `null`.
	antwort := app.NutzerflaecheUeber(frischerBestand(t)).
		DeckeZurUebernahme(probekontext(t), "2")

	switch {
	case antwort.Fehler == nil:
		t.Error("die Wahl in einem Profil ohne Freigabe kam durch")
	case antwort.Decks == nil:
		t.Error("die Liste kam als null")
	}
}
