package app_test

import (
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Navigator und Inspector unter Darstellung, und Titel nach dem Stand** —
// UXR-F14 (#2029),
// Folge aus UXR-02 (#1910).
//
// # Die Quelle
//
// Apple, Human Interface Guidelines, „The menu bar" (nachgelesen am
// 2026-10-02 für UXR-02, Befund in #2029): ein Eintrag, der etwas ein- oder
// ausblendet, heißt nach dem Stand — „einblenden", wenn es zu ist,
// „ausblenden", wenn es steht. Und wo es nichts umzulegen gibt, ist er
// gegraut.
//
// **MSE2 kennt weder Navigator noch Inspector** (`src/gui/set/window.cpp`,
// Fassung f9d9356 vom 2021-07-06, führt seine Bereiche über Reiter): die
// Einträge sind **new**, die Kürzel ⌥⌘[ und ⌥⌘] sind die des Arbeitsplatzes
// (`rahmen/tastatur.ts`).

func navigatorkuerzel() *app.Kuerzel { return &app.Kuerzel{Taste: "[", Wahl: true} }
func inspectorkuerzel() *app.Kuerzel { return &app.Kuerzel{Taste: "]", Wahl: true} }

// teile stellt alle drei Fensterteile vorhanden und gleich offen hin.
func teile(offen bool) app.Darstellungsstand {
	teil := app.Fensterteil{Vorhanden: true, Offen: offen}

	return app.Darstellungsstand{Seitenleiste: teil, Navigator: teil, Inspector: teil}
}

// darstellungsmenue baut die Leiste in einem Stand und gibt das Menü
// *Darstellung* zurück.
func darstellungsmenue(t *testing.T, sprache string, stand app.Darstellungsstand) *app.Menue {
	t.Helper()

	texte := app.MenuetexteFuer(sprache)
	leiste := app.MenueleisteMitStand(app.Menuhandlungen{}, texte, nil, stand)

	menue := untermenueNamens(leiste, texte.Darstellung)
	if menue == nil {
		t.Fatalf("%s: kein Menü %q", sprache, texte.Darstellung)
	}

	return menue
}

func TestNavigatorUndInspectorStehenUnterDarstellungMitIhrenKuerzeln(t *testing.T) {
	t.Parallel()

	for _, sprache := range app.Sprachen {
		menue := darstellungsmenue(t, sprache, teile(true))

		for name, kuerzel := range map[string]*app.Kuerzel{
			"Seitenleiste": seitenleistenkuerzel(),
			"Navigator":    navigatorkuerzel(),
			"Inspector":    inspectorkuerzel(),
		} {
			if eintragMitKuerzel(menue, kuerzel) == nil {
				t.Errorf("%s: kein Eintrag für %s unter Darstellung", sprache, name)
			}
		}

		// **⌥⌘[ ist nicht ⌘[**: Zurück und Vorwärts liegen auf derselben
		// Taste ohne ⌥, und keiner der Einträge darf den anderen ersetzen.
		zurueck := eintragMitKuerzel(menue, app.Befehlstaste("["))
		navigator := eintragMitKuerzel(menue, navigatorkuerzel())

		if zurueck == nil || navigator == nil || zurueck == navigator {
			t.Errorf("%s: ⌘[ und ⌥⌘[ sind nicht zwei Einträge", sprache)
		}
	}
}

// **Der Titel folgt dem Stand** — in jeder Sprache, und die beiden Wörter
// unterscheiden sich.
func TestDieEintraegeHeissenNachDemStand(t *testing.T) {
	t.Parallel()

	for _, sprache := range app.Sprachen {
		texte := app.MenuetexteFuer(sprache)
		offen := darstellungsmenue(t, sprache, teile(true))
		zu := darstellungsmenue(t, sprache, teile(false))

		for _, fall := range []struct {
			kuerzel    *app.Kuerzel
			ausblenden string
			einblenden string
		}{
			{seitenleistenkuerzel(), texte.SeitenleisteAusblenden, texte.SeitenleisteEinblenden},
			{navigatorkuerzel(), texte.NavigatorAusblenden, texte.NavigatorEinblenden},
			{inspectorkuerzel(), texte.InspectorAusblenden, texte.InspectorEinblenden},
		} {
			if fall.ausblenden == "" || fall.einblenden == "" || fall.ausblenden == fall.einblenden {
				t.Fatalf("%s: »%s« und »%s« sind nicht zwei Wörter",
					sprache, fall.ausblenden, fall.einblenden)
			}

			if gesehen := eintragMitKuerzel(offen, fall.kuerzel).Beschriftung; gesehen != fall.ausblenden {
				t.Errorf("%s: wo es steht, heißt der Eintrag %q, erwartet %q", sprache, gesehen, fall.ausblenden)
			}

			if gesehen := eintragMitKuerzel(zu, fall.kuerzel).Beschriftung; gesehen != fall.einblenden {
				t.Errorf("%s: wo es zu ist, heißt der Eintrag %q, erwartet %q", sprache, gesehen, fall.einblenden)
			}
		}
	}
}

// **Die Wörter selbst** — damit ein Tippfehler im Katalog auffällt und nicht
// erst im Fenster.
func TestDieWoerterDerEintraege(t *testing.T) {
	t.Parallel()

	for sprache, erwartet := range map[string][6]string{
		"de": {
			"Seitenleiste einblenden", "Seitenleiste ausblenden",
			"Navigator einblenden", "Navigator ausblenden",
			"Inspector einblenden", "Inspector ausblenden",
		},
		"en": {
			"Show Sidebar", "Hide Sidebar",
			"Show Navigator", "Hide Navigator",
			"Show Inspector", "Hide Inspector",
		},
	} {
		texte := app.MenuetexteFuer(sprache)
		gesehen := [6]string{
			texte.SeitenleisteEinblenden, texte.SeitenleisteAusblenden,
			texte.NavigatorEinblenden, texte.NavigatorAusblenden,
			texte.InspectorEinblenden, texte.InspectorAusblenden,
		}

		if gesehen != erwartet {
			t.Errorf("%s: %q, erwartet %q", sprache, gesehen, erwartet)
		}
	}
}

// **Gegraut, wo es nichts umzulegen gibt** — je Teil für sich.
func TestEinEintragIstGegrautWoDasTeilFehlt(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		name    string
		stand   app.Darstellungsstand
		gegraut map[string]bool
	}{
		{
			name:    "alles da",
			stand:   teile(true),
			gegraut: map[string]bool{"seitenleiste": false, "navigator": false, "inspector": false},
		},
		{
			name:    "nichts da — der Anfang, das Wiederherstellungsfenster, die Sperre",
			stand:   app.Darstellungsstand{},
			gegraut: map[string]bool{"seitenleiste": true, "navigator": true, "inspector": true},
		},
		{
			name: "eine Seite ohne Arbeitsplatz",
			stand: app.Darstellungsstand{
				Seitenleiste: app.Fensterteil{Vorhanden: true, Offen: true},
			},
			gegraut: map[string]bool{"seitenleiste": false, "navigator": true, "inspector": true},
		},
		{
			name: "ein Arbeitsplatz ohne Inspector",
			stand: app.Darstellungsstand{
				Seitenleiste: app.Fensterteil{Vorhanden: true, Offen: true},
				Navigator:    app.Fensterteil{Vorhanden: true, Offen: false},
			},
			gegraut: map[string]bool{"seitenleiste": false, "navigator": false, "inspector": true},
		},
		{
			name: "die Leiste steht oben und lässt sich nicht einklappen",
			stand: app.Darstellungsstand{
				Seitenleiste: app.Fensterteil{Vorhanden: false, Offen: true},
				Navigator:    app.Fensterteil{Vorhanden: true, Offen: true},
				Inspector:    app.Fensterteil{Vorhanden: true, Offen: true},
			},
			gegraut: map[string]bool{"seitenleiste": true, "navigator": false, "inspector": false},
		},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			for _, sprache := range app.Sprachen {
				menue := darstellungsmenue(t, sprache, fall.stand)

				for name, kuerzel := range map[string]*app.Kuerzel{
					"seitenleiste": seitenleistenkuerzel(),
					"navigator":    navigatorkuerzel(),
					"inspector":    inspectorkuerzel(),
				} {
					if gesehen := eintragMitKuerzel(menue, kuerzel).Gegraut; gesehen != fall.gegraut[name] {
						t.Errorf("%s: %s gegraut=%v, erwartet %v", sprache, name, gesehen, fall.gegraut[name])
					}
				}
			}
		})
	}
}

// **Ein Teil, das ein anderes verdeckt, behält seinen Titel.** Liegt eine
// Schublade über dem Arbeitsplatz, ist der Eintrag gegraut — und er sagt
// weiter, wie die Spalte darunter steht.
func TestEinVerdecktesTeilBehaeltSeinenTitel(t *testing.T) {
	t.Parallel()

	texte := app.MenuetexteFuer("de")
	menue := darstellungsmenue(t, "de", app.Darstellungsstand{
		Navigator: app.Fensterteil{Vorhanden: false, Offen: true},
	})

	eintrag := eintragMitKuerzel(menue, navigatorkuerzel())
	if !eintrag.Gegraut || eintrag.Beschriftung != texte.NavigatorAusblenden {
		t.Errorf("gegraut=%v, Titel %q", eintrag.Gegraut, eintrag.Beschriftung)
	}
}

// **Kein anderer Eintrag der Leiste wird gegraut** — gegraut sind genau die
// drei, die vom Stand des Fensters abhängen. Ein vierter wäre ein Eintrag, der
// aus Versehen nichts mehr täte.
func TestNurDieDreiEintraegeDerDarstellungKoennenGegrautSein(t *testing.T) {
	t.Parallel()

	for _, sprache := range app.Sprachen {
		texte := app.MenuetexteFuer(sprache)
		leiste := app.MenueleisteMitStand(app.Menuhandlungen{}, texte, nil, app.Darstellungsstand{})

		var gegraut []string

		var gehen func(*app.Menue)

		gehen = func(menue *app.Menue) {
			for _, eintrag := range menue.Eintraege {
				if eintrag.Gegraut {
					gegraut = append(gegraut, eintrag.Beschriftung)
				}

				if eintrag.Untermenue != nil {
					gehen(eintrag.Untermenue)
				}
			}
		}

		gehen(leiste)

		erwartet := []string{
			texte.SeitenleisteEinblenden, texte.NavigatorEinblenden, texte.InspectorEinblenden,
		}
		if !slices.Equal(gegraut, erwartet) {
			t.Errorf("%s: gegraut sind %q, erwartet %q", sprache, gegraut, erwartet)
		}
	}
}

// **Ein gegrauter Eintrag trägt seine Handlung weiter** — die Hülle graut ihn,
// sie nimmt ihm nichts. Und die Leiste ohne Angabe ist die im Anfangsstand.
func TestDieLeisteOhneStandIstDieImAnfangsstand(t *testing.T) {
	t.Parallel()

	texte := app.MenuetexteFuer("de")
	gerufen := 0
	handlungen := app.Menuhandlungen{NavigatorUmschalten: func() { gerufen++ }}

	ohne := app.Menueleiste(handlungen, texte, nil)
	mit := app.MenueleisteMitStand(handlungen, texte, nil, app.Darstellungsstand{})

	eintrag := eintragMitKuerzel(ohne, navigatorkuerzel())
	if !eintrag.Gegraut || eintrag.Handlung == nil {
		t.Fatalf("gegraut=%v, Handlung=%v", eintrag.Gegraut, eintrag.Handlung != nil)
	}

	eintrag.Handlung()

	if gerufen != 1 {
		t.Errorf("%d-mal gerufen, erwartet einmal", gerufen)
	}

	if a, b := eintragMitKuerzel(ohne, navigatorkuerzel()), eintragMitKuerzel(mit, navigatorkuerzel()); a.Beschriftung != b.Beschriftung || a.Gegraut != b.Gegraut {
		t.Error("Menueleiste und MenueleisteMitStand ohne Stand sind nicht dasselbe")
	}
}

func TestNavigatorUndInspectorRufenIhreHandlung(t *testing.T) {
	t.Parallel()

	gerufen := map[string]int{}
	leiste := app.MenueleisteMitStand(app.Menuhandlungen{
		NavigatorUmschalten: func() { gerufen["navigator"]++ },
		InspectorUmschalten: func() { gerufen["inspector"]++ },
	}, app.MenuetexteFuer("de"), nil, teile(true))

	eintragMitKuerzel(leiste, navigatorkuerzel()).Handlung()
	eintragMitKuerzel(leiste, inspectorkuerzel()).Handlung()
	eintragMitKuerzel(leiste, inspectorkuerzel()).Handlung()

	if gerufen["navigator"] != 1 || gerufen["inspector"] != 2 {
		t.Errorf("gerufen: %v, erwartet navigator=1 und inspector=2", gerufen)
	}
}

// **Ohne Handlung stürzt nichts ab** — dieselbe Zusage wie bei jedem Eintrag.
func TestNavigatorUndInspectorOhneHandlungBleibenKlickbar(t *testing.T) {
	t.Parallel()

	leiste := app.MenueleisteMitStand(app.Menuhandlungen{}, app.MenuetexteFuer("de"), nil, teile(true))

	eintragMitKuerzel(leiste, navigatorkuerzel()).Handlung()
	eintragMitKuerzel(leiste, inspectorkuerzel()).Handlung()
}

// **Die Reihenfolge unter Darstellung**: die drei Teile zuerst, dann der Gang
// der Seiten, dann das Zurücksetzen — durch Trennlinien getrennt.
func TestDieDarstellungFuehrtIhreEintraegeInDieserReihenfolge(t *testing.T) {
	t.Parallel()

	texte := app.MenuetexteFuer("de")
	menue := darstellungsmenue(t, "de", teile(true))

	var gesehen []string

	for _, eintrag := range menue.Eintraege {
		if eintrag.Trenner {
			gesehen = append(gesehen, "—")

			continue
		}

		gesehen = append(gesehen, eintrag.Beschriftung)
	}

	erwartet := []string{
		texte.SeitenleisteAusblenden, texte.NavigatorAusblenden, texte.InspectorAusblenden,
		"—", texte.SeiteZurueck, texte.SeiteVor,
		"—", texte.DarstellungZuruecksetzen,
	}
	if !slices.Equal(gesehen, erwartet) {
		t.Errorf("gesehen %q, erwartet %q", gesehen, erwartet)
	}
}
