package fenster

import (
	"encoding/json"
	"testing"

	"github.com/wailsapp/wails/v3/pkg/application"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Eine Meldung trägt immer eine Liste** — keine, eine oder zwei Angaben.
func TestDieNutzlastIstImmerEineListe(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		daten    []any
		erwartet string
	}{
		"keine":                         {nil, `[]`},
		"eine":                          {[]any{"7"}, `["7"]`},
		"zwei":                          {[]any{"/wo", "kaputt"}, `["/wo","kaputt"]`},
		"eine Liste als einzige Angabe": {[]any{[]string{"a", "b"}}, `[["a","b"]]`},
	} {
		roh, err := json.Marshal(nutzlast(fall.daten))
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}

		if string(roh) != fall.erwartet {
			t.Errorf("%s: %s, erwartet %s", name, roh, fall.erwartet)
		}
	}
}

// **Ein Ablegen trägt Stelle und Pfade — auch ohne Ziel und ohne Pfade.**
func TestEinAblegenTraegtStelleUndPfade(t *testing.T) {
	t.Parallel()

	mitZiel := ablageAus([]string{"/a.png"}, &application.DropTargetDetails{X: 12, Y: 34})
	if mitZiel.X != 12 || mitZiel.Y != 34 || len(mitZiel.Pfade) != 1 {
		t.Errorf("mit Ziel: %+v", mitZiel)
	}

	ohne := ablageAus(nil, nil)

	roh, err := json.Marshal(ohne)
	if err != nil {
		t.Fatal(err)
	}

	if string(roh) != `{"x":0,"y":0,"pfade":[]}` {
		t.Errorf("ohne Ziel und Pfade: %s — `null` statt einer leeren Liste?", roh)
	}
}

// **Maximiert, normal und gemerkt** — UX-14.AK-01.
func TestDieFensteroptionenFolgenDemStand(t *testing.T) {
	t.Parallel()

	maximiert := fensteroptionen(app.Fensterstand{Maximiert: true})
	if maximiert.StartState != application.WindowStateMaximised ||
		maximiert.Width != 1280 || maximiert.Height != 860 {
		t.Errorf("maximiert: %+v", maximiert)
	}

	gemerkt := fensteroptionen(app.Fensterstand{Breite: 1440, Hoehe: 900})
	if gemerkt.StartState != application.WindowStateNormal ||
		gemerkt.Width != 1440 || gemerkt.Height != 900 {
		t.Errorf("gemerkt: %+v", gemerkt)
	}

	if !gemerkt.EnableFileDrop || gemerkt.MinWidth != 1024 || gemerkt.MinHeight != 700 {
		t.Errorf("Ablegen und Mindestmaß: %+v", gemerkt)
	}
}

// **Tab erreicht jeden Knopf** — gefunden am 2026-09-24 mit der WebKit-Probe
// (#1828). Ohne die macOS-Einstellung „Tastaturnavigation" springt Tab in
// WebKit nur zwischen Textfeldern und Auswahllisten: ein Knopf ohne
// `tabindex` ist für die Tastatur nicht da, und aus der Befehlspalette fiel
// der Fokus schon beim ersten Tab aus dem Dialog.
func TestTabErreichtAuchKnoepfeUndVerweise(t *testing.T) {
	t.Parallel()

	optionen := fensteroptionen(app.Fensterstand{})
	vorlieben := optionen.Mac.WebviewPreferences

	if !vorlieben.TabFocusesLinks.IsSet() || !vorlieben.TabFocusesLinks.Get() {
		t.Error("TabFocusesLinks ist nicht gesetzt — Tab überspringt Knöpfe")
	}
}
