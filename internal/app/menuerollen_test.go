package app_test

// **Die Menüs von macOS in der Sprache des Fensters** —
// #255.
//
// Unter Wails 2 kamen Anwendungs-, Bearbeiten- und Fenstermenü als Rollen,
// und ihre Beschriftungen standen englisch in Objective-C (Vermerk 0030).
// Wails 3 baut einen Eintrag mit Rolle und lässt ihn beschriften, ohne dass er
// seinen Selektor verliert. Die Leiste führt deshalb **Einträge** mit Rolle —
// beschriftet aus dem Menükatalog — und keine ganzen Rollenmenüs mehr.
//
// **Was hier nicht steht**: dass macOS die Einträge wirklich bedient. Das
// zeigt nur das Fenster (`docs/WAILS3.md`).

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// rollenIn gibt die Rollen der Einträge eines Menüs, in ihrer Reihenfolge —
// Trennlinien als `-1`, Befehle ohne Rolle als `0`.
func rollenIn(menue *app.Menue) []app.Menuerolle {
	rollen := make([]app.Menuerolle, 0, len(menue.Eintraege))

	for _, eintrag := range menue.Eintraege {
		if eintrag.Trenner {
			rollen = append(rollen, -1)

			continue
		}

		rollen = append(rollen, eintrag.Rolle)
	}

	return rollen
}

func gleicheRollen(a, b []app.Menuerolle) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

// untermenue sucht ein Untermenü der Leiste nach seiner Beschriftung.
func untermenue(leiste *app.Menue, beschriftung string) *app.Menueeintrag {
	for _, eintrag := range leiste.Eintraege {
		if eintrag.Untermenue != nil && eintrag.Beschriftung == beschriftung {
			return eintrag
		}
	}

	return nil
}

// **Das Anwendungsmenü steht vorn, und es ist ein Baum** — mit „Über" und den
// Einstellungen, wo macOS sie erwartet.
func TestDasAnwendungsmenueStehtVornUndTraegtUeberUndEinstellungen(t *testing.T) {
	for _, sprache := range app.Sprachen {
		t.Run(sprache, func(t *testing.T) {
			texte := app.MenuetexteFuer(sprache)
			gerufen := map[string]int{}
			leiste := app.Menueleiste(app.Menuhandlungen{
				UeberOeffnen:         func() { gerufen["ueber"]++ },
				EinstellungenOeffnen: func() { gerufen["einstellungen"]++ },
			}, texte, nil)

			vorn := leiste.Eintraege[0]
			if vorn.Untermenue == nil || vorn.Beschriftung != texte.Anwendung {
				t.Fatalf("vorn steht %+v — erwartet das Anwendungsmenü %q", vorn, texte.Anwendung)
			}

			erwartet := []app.Menuerolle{
				app.KeineMenuerolle, -1, app.KeineMenuerolle, -1, app.RolleDienste, -1,
				app.RolleAusblenden, app.RolleAndereAusblenden, app.RolleAlleEinblenden, -1,
				app.RolleBeenden,
			}
			if rollen := rollenIn(vorn.Untermenue); !gleicheRollen(rollen, erwartet) {
				t.Errorf("das Anwendungsmenü führt %v, erwartet %v", rollen, erwartet)
			}

			eintraege := vorn.Untermenue.Eintraege
			if eintraege[0].Beschriftung != texte.Ueber {
				t.Errorf("zuerst steht %q, erwartet %q", eintraege[0].Beschriftung, texte.Ueber)
			}

			if eintraege[2].Beschriftung != texte.Einstellungen ||
				!gleichesKuerzel(eintraege[2].Kuerzel, app.Befehlstaste(",")) {
				t.Errorf("an dritter Stelle steht %+v, erwartet %q auf ⌘,", eintraege[2], texte.Einstellungen)
			}

			eintraege[0].Handlung()
			eintraege[2].Handlung()

			if gerufen["ueber"] != 1 || gerufen["einstellungen"] != 1 {
				t.Errorf("gerufen wurde %v", gerufen)
			}
		})
	}
}

// **Bearbeiten und Fenster sind Bäume aus Einträgen mit Rolle** — und das
// Fenstermenü trägt selbst eine Rolle: dort führt macOS die offenen Fenster.
func TestBearbeitenUndFensterSindBaeumeAusRollen(t *testing.T) {
	for _, sprache := range app.Sprachen {
		t.Run(sprache, func(t *testing.T) {
			texte := app.MenuetexteFuer(sprache)
			leiste := app.Menueleiste(app.Menuhandlungen{}, texte, nil)

			bearbeiten := untermenue(leiste, texte.Bearbeiten)
			if bearbeiten == nil {
				t.Fatalf("die Leiste führt kein Menü %q", texte.Bearbeiten)
			}

			erwartet := []app.Menuerolle{
				app.RolleWiderrufen, app.RolleWiederholen, -1,
				app.RolleAusschneiden, app.RolleKopieren, app.RolleEinsetzen,
				app.RolleLoeschen, app.RolleAllesAuswaehlen, -1,
				app.KeineMenuerolle,
			}
			if rollen := rollenIn(bearbeiten.Untermenue); !gleicheRollen(rollen, erwartet) {
				t.Errorf("Bearbeiten führt %v, erwartet %v", rollen, erwartet)
			}

			sprachausgabe := untermenue(bearbeiten.Untermenue, texte.Sprachausgabe)
			if sprachausgabe == nil ||
				!gleicheRollen(rollenIn(sprachausgabe.Untermenue),
					[]app.Menuerolle{app.RolleSprechenStarten, app.RolleSprechenStoppen}) {
				t.Errorf("die Sprachausgabe fehlt oder führt anderes: %+v", sprachausgabe)
			}

			fenster := untermenue(leiste, texte.Fenster)
			if fenster == nil || fenster.Rolle != app.RolleFenstermenue {
				t.Fatalf("das Fenstermenü fehlt oder trägt keine Rolle: %+v", fenster)
			}

			if rollen := rollenIn(fenster.Untermenue); !gleicheRollen(rollen,
				[]app.Menuerolle{app.RolleMinimieren, app.RolleZoomen, -1, app.RolleAlleNachVorne}) {
				t.Errorf("das Fenstermenü führt %v", rollen)
			}
		})
	}
}

// **Jeder Eintrag mit Rolle ist beschriftet — und in jeder Sprache anders**,
// wo die Sprachen sich unterscheiden. Die Rolle stellt die Handlung, die
// Beschriftung stellt diese Anwendung.
func TestJederEintragMitRolleIstInDerSpracheDesFenstersBeschriftet(t *testing.T) {
	deutsch := app.Menueleiste(app.Menuhandlungen{}, app.MenuetexteFuer("de"), nil)
	englisch := app.Menueleiste(app.Menuhandlungen{}, app.MenuetexteFuer("en"), nil)

	for _, rolle := range app.Menuerollen() {
		de := suchen(deutsch, func(e *app.Menueeintrag) bool { return e.Rolle == rolle })
		en := suchen(englisch, func(e *app.Menueeintrag) bool { return e.Rolle == rolle })

		if de == nil || en == nil {
			t.Errorf("die Rolle %v steht nicht in beiden Leisten", rolle)

			continue
		}

		if de.Beschriftung == "" || en.Beschriftung == "" {
			t.Errorf("die Rolle %v ist nicht beschriftet: %q, %q", rolle, de.Beschriftung, en.Beschriftung)
		}
	}

	if widerrufen := suchen(deutsch, func(e *app.Menueeintrag) bool {
		return e.Rolle == app.RolleWiderrufen
	}); widerrufen == nil || widerrufen.Beschriftung != "Widerrufen" {
		t.Errorf("⌘Z heißt auf Deutsch %+v — erwartet „Widerrufen\" wie überall unter macOS", widerrufen)
	}
}

// **„Über" steht einmal** — im Anwendungsmenü, und nicht mehr in der Hilfe.
func TestUeberStehtNurImAnwendungsmenue(t *testing.T) {
	texte := app.MenuetexteFuer("de")
	leiste := app.Menueleiste(app.Menuhandlungen{}, texte, nil)

	hilfe := untermenue(leiste, texte.Hilfe)
	if hilfe == nil {
		t.Fatal("die Leiste führt keine Hilfe")
	}

	for _, eintrag := range hilfe.Untermenue.Eintraege {
		if eintrag.Beschriftung == texte.Ueber {
			t.Error("„Über\" steht noch in der Hilfe")
		}
	}

	ablage := untermenue(leiste, texte.Ablage)
	if ablage != nil && suchen(ablage.Untermenue, func(e *app.Menueeintrag) bool {
		return e.Beschriftung == texte.Einstellungen
	}) != nil {
		t.Error("die Einstellungen stehen noch in der Ablage")
	}
}
