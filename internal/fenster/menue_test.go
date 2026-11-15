package fenster

import (
	"fmt"
	"testing"

	"github.com/wailsapp/wails/v3/pkg/application"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// probeumsetzer schreibt eine Rolle als Eintrag hin und merkt sich, welche
// Handlung an welchem Eintrag hängt.
func probeumsetzer() (mit umsetzer, gebunden map[*application.MenuItem]func()) {
	gebunden = map[*application.MenuItem]func(){}

	mit = umsetzer{
		rollen: func(ziel *application.Menu, rolle app.Menuerolle, beschriftung string) {
			ziel.Add(fmt.Sprintf("Rolle %d: %s", rolle, beschriftung))
		},
		binden: func(befehl *application.MenuItem, handlung func()) { gebunden[befehl] = handlung },
	}

	return mit, gebunden
}

// eintrag sucht einen Eintrag nach seiner Beschriftung, auch in Untermenüs.
// menueAus baut eine Leiste in ein frisches Menü — der Weg der Hülle, ohne
// dass ein Test ein Fenster bräuchte.
func menueAus(leiste *app.Menue, mit umsetzer) *application.Menu {
	menue, _ := (&menuehalter{}).erneuern(leiste, mit)

	return menue
}

func eintrag(menue *application.Menu, beschriftung string) *application.MenuItem {
	if treffer := menue.FindByLabel(beschriftung); treffer != nil {
		return treffer
	}

	return nil
}

// **Die ganze Leiste erreicht Wails** — Rollen an ihrer Stelle, Untermenüs,
// Trennlinien, Kürzel und die Handlung hinter jedem Befehl.
func TestDieLeisteErreichtWailsWieSieGebautIst(t *testing.T) {
	t.Parallel()

	gerufen := map[string]int{}
	handlungen := app.Menuhandlungen{
		SetSichern:             func() { gerufen["sichern"]++ },
		EinstellungenOeffnen:   func() { gerufen["einstellungen"]++ },
		SeitenleisteUmschalten: func() { gerufen["seitenleiste"]++ },
		NavigatorUmschalten:    func() { gerufen["navigator"]++ },
		InspectorUmschalten:    func() { gerufen["inspector"]++ },
	}

	// **Alles da und offen**: so heißen die drei Einträge der Darstellung
	// „ausblenden", und keiner ist gegraut (UXR-F14, #2029).
	da := app.Fensterteil{Vorhanden: true, Offen: true}
	stand := app.Darstellungsstand{Seitenleiste: da, Navigator: da, Inspector: da}

	texte := app.MenuetexteFuer("de")
	mit, gebunden := probeumsetzer()
	leiste := menueAus(app.MenueleisteMitStand(handlungen, texte, nil, stand), mit)

	// **Die Menüs an ihrer Stelle** — Anwendung zuerst, der Setzweig
	// zwischen Darstellung und Fenster, wie macOS die eigenen Menüs einer
	// Anwendung führt (UXR-02, #1910; bis dahin stand er vor Bearbeiten), die
	// Hilfe zuletzt; so stehen sie in `Menueleiste`.
	for stelle, erwartet := range map[int]string{
		0: texte.Anwendung,
		1: texte.Ablage,
		2: texte.Bearbeiten,
		3: texte.Darstellung,
		4: texte.Set,
		5: texte.Fenster,
		6: texte.Hilfe,
	} {
		if gefunden := leiste.ItemAt(stelle); gefunden == nil || gefunden.Label() != erwartet {
			t.Errorf("an Stelle %d steht %v, erwartet %q", stelle, gefunden, erwartet)
		}
	}

	// **Ein Eintrag mit Rolle kommt mit seiner Beschriftung an** (#255), und
	// das Fenstermenü trägt die Rolle, an der macOS die Fenster führt.
	if bearbeiten := leiste.ItemAt(2).GetSubmenu(); bearbeiten.ItemAt(0) == nil ||
		bearbeiten.ItemAt(0).Label() != fmt.Sprintf("Rolle %d: %s", app.RolleWiderrufen, texte.Widerrufen) {
		t.Errorf("Bearbeiten beginnt mit %v", bearbeiten.ItemAt(0))
	}

	if fenster := leiste.FindByRole(application.WindowMenu); fenster == nil || fenster.Label() != texte.Fenster {
		t.Errorf("das Fenstermenü trägt seine Rolle nicht: %v", fenster)
	}

	for beschriftung, kuerzel := range map[string]string{
		texte.Sichern:                "Cmd+S",
		texte.SichernUnter:           "Cmd+Shift+S",
		texte.Einstellungen:          "Cmd+,",
		texte.SeitenleisteAusblenden: "Cmd+Option+S",
		texte.NavigatorAusblenden:    "Cmd+Option+[",
		texte.InspectorAusblenden:    "Cmd+Option+]",
		texte.SeiteZurueck:           "Cmd+[",
		texte.SeiteVor:               "Cmd+]",
		texte.Suchen:                 "",
	} {
		befehl := eintrag(leiste, beschriftung)
		if befehl == nil {
			t.Errorf("%q fehlt", beschriftung)

			continue
		}

		if befehl.GetAccelerator() != kuerzel {
			t.Errorf("%q trägt %q, erwartet %q", beschriftung, befehl.GetAccelerator(), kuerzel)
		}
	}

	for beschriftung, schluessel := range map[string]string{
		texte.Sichern: "sichern", texte.Einstellungen: "einstellungen",
		texte.SeitenleisteAusblenden: "seitenleiste", texte.NavigatorAusblenden: "navigator",
		texte.InspectorAusblenden: "inspector",
	} {
		handlung := gebunden[eintrag(leiste, beschriftung)]
		if handlung == nil {
			t.Fatalf("an %q hängt keine Handlung", beschriftung)
		}

		handlung()

		if gerufen[schluessel] != 1 {
			t.Errorf("%q rief %d-mal, erwartet einmal", beschriftung, gerufen[schluessel])
		}
	}

	// **Die Trennlinien der Ablage kommen mit.**
	ablage := leiste.ItemAt(1).GetSubmenu()
	trenner := 0

	for stelle := 0; ablage.ItemAt(stelle) != nil; stelle++ {
		if ablage.ItemAt(stelle).IsSeparator() {
			trenner++
		}
	}

	// **Drei, seit die Einstellungen im Anwendungsmenü stehen** (#255).
	if trenner != 3 {
		t.Errorf("die Ablage hat %d Trennlinien, erwartet drei", trenner)
	}
}

// **Gegraut kommt gegraut an** — UXR-F14
// (#2029).
//
// Das Modell kennt `Gegraut`, Wails kennt `Enabled`; die Hülle verbindet beide.
// Wails schaltet das Selbstaktivieren der Einträge ab (`setAutoenablesItems:NO`,
// `menu_darwin.go`, v3.0.0-beta.23): ein Eintrag ist, was **diese** Anwendung
// ihm sagt, und macOS fragt nicht nach. Was macOS daraus zeichnet, zeigt nur
// das Fenster (`scripts/huellenprobe/darstellung.mjs`).
func TestEinGegrauterEintragErreichtWailsGegraut(t *testing.T) {
	t.Parallel()

	texte := app.MenuetexteFuer("de")
	mit, _ := probeumsetzer()

	// **Im Anfangsstand fehlt alles**: keine Seitenleiste, kein Arbeitsplatz.
	anfang := menueAus(app.Menueleiste(app.Menuhandlungen{}, texte, nil), mit)

	for _, beschriftung := range []string{
		texte.SeitenleisteEinblenden, texte.NavigatorEinblenden, texte.InspectorEinblenden,
	} {
		if befehl := eintrag(anfang, beschriftung); befehl == nil || befehl.Enabled() {
			t.Errorf("%q fehlt oder ist nicht gegraut: %v", beschriftung, befehl)
		}
	}

	// **Alles andere bleibt bedienbar** — ein gegrauter Eintrag zu viel wäre
	// ein Kürzel, das still nichts mehr täte.
	for _, beschriftung := range []string{
		texte.Sichern, texte.Einstellungen, texte.SeiteZurueck, texte.SeiteVor,
		texte.DarstellungZuruecksetzen, texte.Suchen,
	} {
		if befehl := eintrag(anfang, beschriftung); befehl == nil || !befehl.Enabled() {
			t.Errorf("%q fehlt oder ist gegraut: %v", beschriftung, befehl)
		}
	}

	// **Und mit allen Teilen steht keiner der drei mehr gegraut da.**
	da := app.Fensterteil{Vorhanden: true, Offen: false}
	stand := app.Darstellungsstand{Seitenleiste: da, Navigator: da, Inspector: da}
	mit, _ = probeumsetzer()
	bedienbar := menueAus(app.MenueleisteMitStand(app.Menuhandlungen{}, texte, nil, stand), mit)

	for _, beschriftung := range []string{
		texte.SeitenleisteEinblenden, texte.NavigatorEinblenden, texte.InspectorEinblenden,
	} {
		if befehl := eintrag(bedienbar, beschriftung); befehl == nil || !befehl.Enabled() {
			t.Errorf("%q fehlt oder ist gegraut: %v", beschriftung, befehl)
		}
	}
}

// **Jede Rolle der Leiste hat in Wails eine Rolle, die wirkt** (#255).
//
// Zwei Rollen aus Wails' eigenen Menüs tun auf macOS nichts — `UnHide` und
// `Front` haben weder Selektor noch Handlung. Eine Zuordnung dorthin sähe aus
// wie ein Menüeintrag und wäre keiner.
func TestJedeRolleHatInWailsEineWirksameRolle(t *testing.T) {
	t.Parallel()

	tot := map[application.Role]string{
		application.NoRole:             "keine Rolle",
		application.UnHide:             "UnHide hat keinen Selektor",
		application.Front:              "Front hat keinen Selektor",
		application.PasteAndMatchStyle: "PasteAndMatchStyle hat keinen Selektor",
	}

	for _, rolle := range app.Menuerollen() {
		wails, da := wailsrollen[rolle]
		if !da {
			t.Errorf("die Rolle %v hat keine Entsprechung in Wails", rolle)

			continue
		}

		if grund, ist := tot[wails]; ist {
			t.Errorf("die Rolle %v geht an eine, die nichts tut: %s", rolle, grund)
		}
	}

	if len(wailsrollen) != len(app.Menuerollen()) {
		t.Errorf("%d Zuordnungen für %d Rollen", len(wailsrollen), len(app.Menuerollen()))
	}
}

// **Jede Kombination wird so geschrieben, wie Wails sie liest.**
func TestKuerzeltextSchreibtJedeKombination(t *testing.T) {
	t.Parallel()

	for kuerzel, erwartet := range map[app.Kuerzel]string{
		{Taste: "s"}:                             "CmdOrCtrl+s",
		{Taste: "s", Umschalt: true}:             "CmdOrCtrl+Shift+s",
		{Taste: "s", Wahl: true}:                 "CmdOrCtrl+OptionOrAlt+s",
		{Taste: "s", Umschalt: true, Wahl: true}: "CmdOrCtrl+Shift+OptionOrAlt+s",
		{Taste: "page down", OhneBefehl: true}:   "page down",
	} {
		if text := kuerzeltext(&kuerzel); text != erwartet {
			t.Errorf("%+v: %q, erwartet %q", kuerzel, text, erwartet)
		}
	}
}
