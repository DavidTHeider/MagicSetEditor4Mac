package fenster

// **Eine Leiste derselben Gestalt wird nachgetragen, nicht neu gebaut** —
// #2312 (#2312).
//
// # Der Anlass
//
// Wails gibt die Beschriftung eines neuen Eintrags nie frei: `newMenuItem`
// liest `label` und gibt nur `tooltip` zurück (`menuitem_darwin.go`, Zeilen
// 15–42, v3.0.0-beta.27), und `newMenuItemImpl` reicht `C.CString(item.label)`
// hinein. Jeder Neubau der Leiste ließ damit eine Zeichenkette je Eintrag
// liegen. Gemessen am nativen Fenster mit `MallocStackLogging`: 21.008
// Zuteilungen aus `_Cfunc__Cmalloc` in sechs Minuten Bearbeiten, rund
// 650 KB — ein Drittel dessen, was der Anwendungsprozess in dieser Zeit
// wuchs.
//
// **Gebaut wird die Leiste bei jedem Wechsel des Darstellungsstands und
// jedem aufgeschlagenen Set** ([app.App.DarstellungsstandMelden],
// `SetBenutzt`). Dabei ändern sich Beschriftungen („einblenden" statt
// „ausblenden", ein anderes Set in „Zuletzt benutzt"), das Gegraute und die
// Handlung — die Gestalt fast nie. `setMenuItemLabel` gibt seine
// Zeichenkette frei, `setMenuItemDisabled` braucht keine.
//
// # Was er nicht hält
//
// **Ob AppKit danach zeichnet, was die Einträge sagen.** Das zeigt nur das
// Fenster (`scripts/huellenprobe/darstellung.mjs`).

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

func TestEinNeuerStandSetztKeinNeuesMenue(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	fenster := &probehuelle{}
	mit, gebunden := probeumsetzer()

	verdrahten(anwendung, fenster, mit)

	da := app.Fensterteil{Vorhanden: true, Offen: true}
	zu := app.Fensterteil{Vorhanden: true}

	for _, stand := range []app.Darstellungsstand{
		{Seitenleiste: da, Navigator: da, Inspector: da},
		{Seitenleiste: zu, Navigator: da},
		{Seitenleiste: da, Navigator: zu, Inspector: zu},
	} {
		if antwort := anwendung.DarstellungsstandMelden(stand); antwort.Fehler != nil {
			t.Fatalf("Stand %+v: %+v", stand, antwort.Fehler)
		}
	}

	if len(fenster.leisten) != 1 {
		t.Fatalf("gesetzt wurden %d Leisten, erwartet nur die erste — drei Stände derselben Gestalt", len(fenster.leisten))
	}

	// **Und die eine Leiste sagt den letzten Stand**: die Seitenleiste ist
	// offen, Navigator und Inspector sind zu, und nichts ist gegraut.
	texte := app.MenuetexteFuer("de")
	leiste := fenster.leisten[0]

	for beschriftung, offen := range map[string]bool{
		texte.SeitenleisteAusblenden: true,
		texte.NavigatorEinblenden:    true,
		texte.InspectorEinblenden:    true,
	} {
		befehl := eintrag(leiste, beschriftung)
		if befehl == nil {
			t.Errorf("%q steht nicht in der Leiste", beschriftung)

			continue
		}

		if befehl.Enabled() != offen {
			t.Errorf("%q ist bedienbar: %v, erwartet %v", beschriftung, befehl.Enabled(), offen)
		}

		if gebunden[befehl] == nil {
			t.Errorf("an %q hängt keine Handlung", beschriftung)
		}
	}

	for _, alt := range []string{texte.SeitenleisteEinblenden, texte.NavigatorAusblenden, texte.InspectorAusblenden} {
		if eintrag(leiste, alt) != nil {
			t.Errorf("%q steht noch in der Leiste", alt)
		}
	}
}

// **Die Handlung kommt mit** — ein anderes Set an derselben Stelle von
// „Zuletzt benutzt" öffnet das andere Set, nicht das von vorhin.
func TestEinNachgetragenerEintragOeffnetDasNeueSet(t *testing.T) {
	t.Parallel()

	var geoeffnet []string
	handlungen := app.Menuhandlungen{SetZuletzt: func(id string) { geoeffnet = append(geoeffnet, id) }}
	texte := app.MenuetexteFuer("de")
	mit, gebunden := probeumsetzer()
	halter := &menuehalter{}

	erstes, neu := halter.erneuern(app.Menueleiste(handlungen, texte, []app.Zuletzteintrag{{ID: "1", Name: "Alpha"}}), mit)
	if !neu {
		t.Fatal("die erste Leiste gilt nicht als neu")
	}

	zweites, neu := halter.erneuern(app.Menueleiste(handlungen, texte, []app.Zuletzteintrag{{ID: "2", Name: "Beta"}}), mit)
	if neu || zweites != erstes {
		t.Fatalf("eine Leiste derselben Gestalt wurde neu gebaut (neu: %v)", neu)
	}

	befehl := eintrag(zweites, "Beta")
	if befehl == nil || eintrag(zweites, "Alpha") != nil {
		t.Fatalf("die Beschriftung wurde nicht nachgetragen: Beta %v", befehl)
	}

	gebunden[befehl]()

	if len(geoeffnet) != 1 || geoeffnet[0] != "2" {
		t.Errorf("geöffnet wurde %v, erwartet das Set 2", geoeffnet)
	}
}

// **Eine andere Gestalt wird neu gebaut** — ein Eintrag mehr in „Zuletzt
// benutzt" lässt sich nicht nachtragen.
func TestEineAndereGestaltBautNeu(t *testing.T) {
	t.Parallel()

	texte := app.MenuetexteFuer("de")
	mit, _ := probeumsetzer()
	halter := &menuehalter{}

	eins := []app.Zuletzteintrag{{ID: "1", Name: "Alpha"}}
	zwei := []app.Zuletzteintrag{{ID: "1", Name: "Alpha"}, {ID: "2", Name: "Beta"}}

	erstes, _ := halter.erneuern(app.Menueleiste(app.Menuhandlungen{}, texte, eins), mit)

	zweites, neu := halter.erneuern(app.Menueleiste(app.Menuhandlungen{}, texte, zwei), mit)
	if !neu || zweites == erstes {
		t.Fatalf("eine Leiste mit einem Eintrag mehr wurde nachgetragen (neu: %v)", neu)
	}

	if eintrag(zweites, "Beta") == nil {
		t.Error("der neue Eintrag fehlt")
	}

	// **Und danach trägt der Halter die neue Gestalt nach**, nicht die alte.
	drittes, neu := halter.erneuern(app.Menueleiste(app.Menuhandlungen{}, texte,
		[]app.Zuletzteintrag{{ID: "3", Name: "Gamma"}, {ID: "2", Name: "Beta"}}), mit)
	if neu || drittes != zweites || eintrag(drittes, "Gamma") == nil {
		t.Errorf("die zweite Gestalt wurde nicht nachgetragen (neu: %v)", neu)
	}
}
