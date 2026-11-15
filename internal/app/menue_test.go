package app_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// Die Einstellungen liegen auf ⌘, — das ist unter macOS keine Geschmacksfrage,
// sondern die Stelle, an der Nutzende sie suchen (AGENTS.md §6).
func TestEinstellungenLiegenAufBefehlKomma(t *testing.T) {
	leiste := app.Menueleiste(app.Menuhandlungen{EinstellungenOeffnen: func() {}},
		app.MenuetexteFuer("de"), nil)

	eintrag := eintragMitKuerzel(leiste, app.Befehlstaste(","))
	if eintrag == nil {
		t.Fatal("kein Eintrag auf ⌘, gefunden")
	}
	if eintrag.Beschriftung == "" {
		t.Error("der Eintrag auf ⌘, trägt keine Beschriftung")
	}
}

func TestEinstellungenRufenIhreHandlung(t *testing.T) {
	gerufen := 0
	leiste := app.Menueleiste(app.Menuhandlungen{
		EinstellungenOeffnen: func() { gerufen++ },
	}, app.MenuetexteFuer("de"), nil)

	eintrag := eintragMitKuerzel(leiste, app.Befehlstaste(","))
	if eintrag == nil {
		t.Fatal("kein Eintrag auf ⌘, gefunden")
	}

	eintrag.Handlung()

	if gerufen != 1 {
		t.Errorf("die Handlung wurde %d-mal gerufen, erwartet war einmal", gerufen)
	}
}

// Eine fehlende Handlung darf die Anwendung nicht beim Klick zum Absturz
// bringen. Ein `nil`-Rückruf ist ein Programmierfehler, aber er soll sich beim
// Bauen der Leiste zeigen und nicht erst, wenn jemand das Menü benutzt.
func TestMenueleisteOhneHandlungBleibtKlickbar(t *testing.T) {
	leiste := app.Menueleiste(app.Menuhandlungen{}, app.MenuetexteFuer("de"), nil)

	eintrag := eintragMitKuerzel(leiste, app.Befehlstaste(","))
	if eintrag == nil {
		t.Fatal("kein Eintrag auf ⌘, gefunden")
	}

	eintrag.Handlung()
}

// ------------------------------------------------------------------ Suchhilfen

func eintragMitKuerzel(leiste *app.Menue, kuerzel *app.Kuerzel) *app.Menueeintrag {
	return suchen(leiste, func(eintrag *app.Menueeintrag) bool {
		return gleichesKuerzel(eintrag.Kuerzel, kuerzel)
	})
}

func gleichesKuerzel(a, b *app.Kuerzel) bool {
	return a != nil && b != nil && *a == *b
}

// suchen geht die Leiste samt Untermenüs durch.
func suchen(leiste *app.Menue, passt func(*app.Menueeintrag) bool) *app.Menueeintrag {
	if leiste == nil {
		return nil
	}

	for _, eintrag := range leiste.Eintraege {
		if passt(eintrag) {
			return eintrag
		}
		if treffer := suchen(eintrag.Untermenue, passt); treffer != nil {
			return treffer
		}
	}

	return nil
}

// **Die Beschriftungen folgen der Sprache.**
func TestDieMenueleisteFolgtDerSprache(t *testing.T) {
	deutsch := app.Menueleiste(app.Menuhandlungen{}, app.MenuetexteFuer("de"), nil)
	englisch := app.Menueleiste(app.Menuhandlungen{}, app.MenuetexteFuer("en"), nil)

	deutscherEintrag := eintragMitKuerzel(deutsch, app.Befehlstaste(","))
	englischerEintrag := eintragMitKuerzel(englisch, app.Befehlstaste(","))

	if deutscherEintrag == nil || englischerEintrag == nil {
		t.Fatal("kein Eintrag auf ⌘, gefunden")
	}

	if deutscherEintrag.Beschriftung == englischerEintrag.Beschriftung {
		t.Errorf("beide Sprachen zeigen %q", deutscherEintrag.Beschriftung)
	}
}
