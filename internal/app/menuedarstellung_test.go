package app_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Darstellung zurücksetzen, aus der Menüleiste** — UX-09.AK-03
// (#1681):
// *„ein sicherer Standard kann auch bei unlesbarer Konfiguration
// wiederhergestellt werden."* Die Menüleiste zeichnet macOS — sie bleibt
// lesbar, gleich was im Fenster steht.

// eintragNamens sucht einen Eintrag über seine Beschriftung, in jeder Tiefe.
func eintragNamens(leiste *app.Menue, beschriftung string) *app.Menueeintrag {
	for _, eintrag := range leiste.Eintraege {
		if eintrag.Beschriftung == beschriftung && eintrag.Untermenue == nil {
			return eintrag
		}

		if eintrag.Untermenue != nil {
			if gefunden := eintragNamens(eintrag.Untermenue, beschriftung); gefunden != nil {
				return gefunden
			}
		}
	}

	return nil
}

func TestDieDarstellungLaesstSichAusDerMenueleisteZuruecksetzen(t *testing.T) {
	t.Parallel()

	for _, sprache := range app.Sprachen {
		texte := app.MenuetexteFuer(sprache)
		gerufen := 0
		leiste := app.Menueleiste(app.Menuhandlungen{
			DarstellungZuruecksetzen: func() { gerufen++ },
		}, texte, nil)

		eintrag := eintragNamens(leiste, texte.DarstellungZuruecksetzen)
		if texte.DarstellungZuruecksetzen == "" || eintrag == nil {
			t.Fatalf("%s: kein Eintrag »%s«", sprache, texte.DarstellungZuruecksetzen)
		}

		eintrag.Handlung()

		if gerufen != 1 {
			t.Errorf("%s: %d-mal gerufen, erwartet einmal", sprache, gerufen)
		}
	}

	if app.MenuetexteFuer("de").DarstellungZuruecksetzen ==
		app.MenuetexteFuer("en").DarstellungZuruecksetzen {
		t.Error("der Eintrag folgt der Sprache nicht")
	}
}

// **Der Eintrag meldet, das Fenster entscheidet** — wie jeder andere.
func TestDasZuruecksetzenDerDarstellungIstEinEreignis(t *testing.T) {
	t.Parallel()

	if app.EreignisDarstellungZuruecksetzen != "menue:darstellung-zuruecksetzen" {
		t.Errorf("Ereignis %q", app.EreignisDarstellungZuruecksetzen)
	}

	if app.NeueApp("").Fensterhandlungen().DarstellungZuruecksetzen == nil {
		t.Error("das Fenster bekommt keine Handlung dafür")
	}
}
