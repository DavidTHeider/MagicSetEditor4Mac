// **Die Ablage bekommt Öffnen, Schließen und Drucken** —
// #891,
// Werkstück 2 von Phase Y.
//
// # Was vorher war
//
// Die Ablage hatte vier Einträge: Sichern, Sichern unter, Bestand sichern,
// Einstellungen. **⌘O, ⌘W und ⌘P taten nichts** — sie waren nicht belegt, und
// unter macOS sind alle drei gewohnte Bedeutungen.
package app_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Jedes Kürzel liegt auf seinem Eintrag und ruft seine Handlung.**
//
// # Warum das eine Tabelle ist und nicht vier Läufe
//
// Die Frage ist für alle dieselbe: liegt der Eintrag auf dem gewohnten Kürzel,
// und ruft er genau eine Handlung? Vier Läufe wären vier Kopien derselben
// Zusicherung — und die eine, die jemand beim Kopieren vergisst, fällt nicht
// auf.
func TestDieAblageBelegtDieGewohntenKuerzel(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		kuerzel *app.Kuerzel
		setzen  func(*app.Menuhandlungen, func())
	}{
		"Neu auf ⌘N": {
			app.Befehlstaste("n"),
			func(h *app.Menuhandlungen, f func()) { h.SetNeu = f },
		},
		"Öffnen auf ⌘O": {
			app.Befehlstaste("o"),
			func(h *app.Menuhandlungen, f func()) { h.SetOeffnen = f },
		},
		"Schließen auf ⌘W": {
			app.Befehlstaste("w"),
			func(h *app.Menuhandlungen, f func()) { h.SetSchliessen = f },
		},
		"Drucken auf ⌘P": {
			app.Befehlstaste("p"),
			func(h *app.Menuhandlungen, f func()) { h.Drucken = f },
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			gerufen := 0

			var handlungen app.Menuhandlungen

			fall.setzen(&handlungen, func() { gerufen++ })

			eintrag := eintragMitKuerzel(
				app.Menueleiste(handlungen, app.MenuetexteFuer("de"), nil), fall.kuerzel)
			if eintrag == nil {
				t.Fatalf("kein Eintrag auf dem Kürzel gefunden")
			}

			if eintrag.Beschriftung == "" {
				t.Error("der Eintrag trägt keine Beschriftung")
			}

			eintrag.Handlung()

			if gerufen != 1 {
				t.Errorf("die Handlung wurde %d-mal gerufen", gerufen)
			}
		})
	}
}

// **Und es gibt ein Hilfemenü.**
//
// macOS erwartet es rechts außen; wer dort nichts findet, sucht die Hilfe
// nirgends sonst (§6).
func TestDieLeisteFuehrtEinHilfemenue(t *testing.T) {
	t.Parallel()

	gerufen := 0

	leiste := app.Menueleiste(
		app.Menuhandlungen{HilfeOeffnen: func() { gerufen++ }},
		app.MenuetexteFuer("de"), nil)

	hilfe := untermenueMitNamen(leiste, app.MenuetexteFuer("de").Hilfe)
	if hilfe == nil {
		t.Fatal("die Leiste führt kein Hilfemenü")
	}

	if len(hilfe.Untermenue.Eintraege) == 0 {
		t.Fatal("das Hilfemenü ist leer")
	}

	eintrag := hilfe.Untermenue.Eintraege[0]
	eintrag.Handlung()

	if gerufen != 1 {
		t.Errorf("die Handlung wurde %d-mal gerufen", gerufen)
	}
}

// untermenueMitNamen sucht ein Untermenü an seiner Beschriftung.
func untermenueMitNamen(leiste *app.Menue, name string) *app.Menueeintrag {
	for _, eintrag := range leiste.Eintraege {
		if eintrag.Beschriftung == name && eintrag.Untermenue != nil {
			return eintrag
		}
	}

	return nil
}

// **Das Untermenü „Zuletzt benutzt" führt die jüngsten Sets** (#891).
//
// # Warum es überhaupt eines gibt
//
// MSE2 hat es (`src/gui/set/window.cpp`, Fassung f9d9356 vom 2021-07-06,
// Zeile 76: „recent files go here"), und macOS erwartet es in der Ablage.
func TestDieAblageFuehrtDieJuengstenSets(t *testing.T) {
	t.Parallel()

	var geoeffnet string

	leiste := app.Menueleiste(
		app.Menuhandlungen{SetZuletzt: func(setID string) { geoeffnet = setID }},
		app.MenuetexteFuer("de"),
		[]app.Zuletzteintrag{{ID: "7", Name: "Probeset"}, {ID: "9", Name: "Zweites"}},
	)

	ablage := untermenueMitNamen(leiste, app.MenuetexteFuer("de").Ablage)
	if ablage == nil {
		t.Fatal("die Leiste führt keine Ablage")
	}

	zuletzt := untermenueMitNamen(&app.Menue{Eintraege: ablage.Untermenue.Eintraege},
		app.MenuetexteFuer("de").Zuletzt)
	if zuletzt == nil {
		t.Fatal("die Ablage führt kein Untermenü »Zuletzt benutzt«")
	}

	if len(zuletzt.Untermenue.Eintraege) != 2 {
		t.Fatalf("das Untermenü führt %d Einträge", len(zuletzt.Untermenue.Eintraege))
	}

	if zuletzt.Untermenue.Eintraege[0].Beschriftung != "Probeset" {
		t.Errorf("der erste Eintrag heißt %q", zuletzt.Untermenue.Eintraege[0].Beschriftung)
	}

	// **Jeder Eintrag trägt seine eigene Kennung.** Ein Rückruf, der sie aus
	// der Schleifenvariablen läse, gäbe bei jedem Klick die letzte.
	zweiter := zuletzt.Untermenue.Eintraege[1]
	zweiter.Handlung()

	if geoeffnet != "9" {
		t.Errorf("geöffnet wurde %q, erwartet war »9«", geoeffnet)
	}
}

// **Ohne Sets steht das Untermenü nicht da.**
//
// Ein leeres Untermenü ist ein Versprechen, das beim Aufklappen
// zurückgenommen wird.
func TestOhneSetsFehltDasUntermenue(t *testing.T) {
	t.Parallel()

	leiste := app.Menueleiste(app.Menuhandlungen{}, app.MenuetexteFuer("de"), nil)

	ablage := untermenueMitNamen(leiste, app.MenuetexteFuer("de").Ablage)
	if ablage == nil {
		t.Fatal("die Leiste führt keine Ablage")
	}

	if untermenueMitNamen(&app.Menue{Eintraege: ablage.Untermenue.Eintraege},
		app.MenuetexteFuer("de").Zuletzt) != nil {
		t.Error("das Untermenü steht ohne Sets da")
	}
}
