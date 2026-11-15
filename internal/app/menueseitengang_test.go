package app_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Zurück und Vorwärts zwischen den Seiten** — UX-03.AK-02
// (#1675):
// ⌘[ und ⌘] unter *Darstellung* — die Kürzel des Finders für den vorigen und
// den nächsten Ordner. Das Fenster führt den Gang; der Eintrag meldet nur.

func TestZurueckUndVorwaertsStehenUnterDarstellung(t *testing.T) {
	t.Parallel()

	for _, sprache := range app.Sprachen {
		texte := app.MenuetexteFuer(sprache)
		leiste := app.Menueleiste(app.Menuhandlungen{}, texte, nil)

		darstellung := untermenueNamens(leiste, texte.Darstellung)
		if darstellung == nil {
			t.Fatalf("%s: kein Menü %q", sprache, texte.Darstellung)
		}

		for _, erwartet := range []struct {
			taste        string
			beschriftung string
		}{
			{"[", texte.SeiteZurueck},
			{"]", texte.SeiteVor},
		} {
			eintrag := eintragMitKuerzel(darstellung, app.Befehlstaste(erwartet.taste))
			if eintrag == nil {
				t.Fatalf("%s: kein Eintrag auf ⌘%s unter %q", sprache, erwartet.taste, texte.Darstellung)
			}

			if eintrag.Beschriftung == "" || eintrag.Beschriftung != erwartet.beschriftung {
				t.Errorf("%s: ⌘%s heißt %q, erwartet %q",
					sprache, erwartet.taste, eintrag.Beschriftung, erwartet.beschriftung)
			}
		}
	}
}

func TestZurueckUndVorwaertsRufenIhreHandlung(t *testing.T) {
	t.Parallel()

	gerufen := map[string]int{}
	leiste := app.Menueleiste(app.Menuhandlungen{
		SeiteZurueck: func() { gerufen["zurueck"]++ },
		SeiteVor:     func() { gerufen["vor"]++ },
	}, app.MenuetexteFuer("de"), nil)

	eintragMitKuerzel(leiste, app.Befehlstaste("[")).Handlung()
	eintragMitKuerzel(leiste, app.Befehlstaste("]")).Handlung()
	eintragMitKuerzel(leiste, app.Befehlstaste("]")).Handlung()

	if gerufen["zurueck"] != 1 || gerufen["vor"] != 2 {
		t.Errorf("gerufen: %v, erwartet zurueck=1 und vor=2", gerufen)
	}
}

// Die Beschriftungen heißen wie in den Systemeinstellungen — und in beiden
// Sprachen anders.
func TestZurueckUndVorwaertsFolgenDerSprache(t *testing.T) {
	t.Parallel()

	deutsch := app.MenuetexteFuer("de")
	englisch := app.MenuetexteFuer("en")

	if deutsch.SeiteZurueck != "Zurück" || deutsch.SeiteVor != "Vorwärts" {
		t.Errorf("deutsch: %q und %q", deutsch.SeiteZurueck, deutsch.SeiteVor)
	}

	if englisch.SeiteZurueck != "Back" || englisch.SeiteVor != "Forward" {
		t.Errorf("englisch: %q und %q", englisch.SeiteZurueck, englisch.SeiteVor)
	}
}

// untermenueNamens sucht ein Untermenü der Leiste beim Namen.
func untermenueNamens(leiste *app.Menue, name string) *app.Menue {
	for _, eintrag := range leiste.Eintraege {
		if eintrag.Untermenue != nil && eintrag.Beschriftung == name {
			return eintrag.Untermenue
		}
	}

	return nil
}
