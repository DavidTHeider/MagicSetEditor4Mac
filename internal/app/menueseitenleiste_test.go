package app_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Die Seitenleiste in der Menüleiste** — Werkstück U14 von Phase 5.1
// (#1490):
// ⌥⌘S wie im Finder und in Mail, unter *Darstellung*, damit das Kürzel dort
// sichtbar ist (§6: bedient, nicht nachgeahmt).

func seitenleistenkuerzel() *app.Kuerzel {
	return &app.Kuerzel{Taste: "s", Wahl: true}
}

func TestDieSeitenleisteLiegtAufWahlBefehlS(t *testing.T) {
	t.Parallel()

	for _, sprache := range app.Sprachen {
		leiste := app.Menueleiste(app.Menuhandlungen{}, app.MenuetexteFuer(sprache), nil)

		eintrag := eintragMitKuerzel(leiste, seitenleistenkuerzel())
		if eintrag == nil {
			t.Fatalf("%s: kein Eintrag auf ⌥⌘S", sprache)
		}

		if eintrag.Beschriftung == "" {
			t.Errorf("%s: der Eintrag auf ⌥⌘S trägt keine Beschriftung", sprache)
		}
	}

	// **Kein Zusammenstoß mit ⌘S und ⇧⌘S** — beide sichern.
	leiste := app.Menueleiste(app.Menuhandlungen{}, app.MenuetexteFuer("de"), nil)
	if sichern := eintragMitKuerzel(leiste, app.Befehlstaste("s")); sichern == nil ||
		sichern.Beschriftung == eintragMitKuerzel(leiste, seitenleistenkuerzel()).Beschriftung {
		t.Error("⌘S und ⌥⌘S zeigen auf denselben Eintrag")
	}
}

func TestDieSeitenleisteRuftIhreHandlung(t *testing.T) {
	t.Parallel()

	gerufen := 0
	leiste := app.Menueleiste(app.Menuhandlungen{
		SeitenleisteUmschalten: func() { gerufen++ },
	}, app.MenuetexteFuer("de"), nil)

	eintrag := eintragMitKuerzel(leiste, seitenleistenkuerzel())
	if eintrag == nil {
		t.Fatal("kein Eintrag auf ⌥⌘S")
	}

	eintrag.Handlung()

	if gerufen != 1 {
		t.Errorf("die Handlung wurde %d-mal gerufen, erwartet war einmal", gerufen)
	}
}

// Das Menü heißt in beiden Sprachen anders — und wie unter macOS.
func TestDasDarstellungsmenueFolgtDerSprache(t *testing.T) {
	t.Parallel()

	deutsch := app.MenuetexteFuer("de")
	englisch := app.MenuetexteFuer("en")

	if deutsch.Darstellung != "Darstellung" || englisch.Darstellung != "View" {
		t.Errorf("Darstellung heißt %q und %q", deutsch.Darstellung, englisch.Darstellung)
	}

	if deutsch.SeitenleisteAusblenden == "" || deutsch.SeitenleisteAusblenden == englisch.SeitenleisteAusblenden {
		t.Errorf("Seitenleiste ausblenden heißt %q und %q",
			deutsch.SeitenleisteAusblenden, englisch.SeitenleisteAusblenden)
	}
}
