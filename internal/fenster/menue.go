package fenster

import (
	"github.com/wailsapp/wails/v3/pkg/application"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// umsetzer sind die zwei Griffe, die Wails zum Umsetzen braucht.
//
// **Zwei Nähte und keine festen Rufe**: Wails 3 baut das Anwendungsmenü aus
// dem Namen der laufenden Anwendung (`NewAppMenu` liest `globalApplication`),
// und ohne eine gibt es nur einen Absturz. Und einen Rückruf, den `OnClick`
// hinterlegt hat, gibt es nicht wieder heraus — `handleClick` ist nicht
// exportiert. Die Läufe setzen deshalb eigene ein und sehen so, **dass** und
// **wo** eine Rolle steht und welcher Eintrag welche Handlung trägt; was macOS
// daraus macht, zeigt nur das Fenster.
type umsetzer struct {
	rollen func(ziel *application.Menu, rolle app.Menuerolle, beschriftung string)
	binden func(befehl *application.MenuItem, handlung func())
}

// wailsrollen ordnet jeder Rolle der Leiste die Rolle in Wails zu, die auf
// macOS **wirkt** (#255).
//
// # Warum nicht die Namen, die Wails' eigene Menüs benutzen
//
// **Weil zwei davon nichts tun.** Wails' Anwendungsmenü führt „Show All" als
// `UnHide` und sein Fenstermenü „Bring All to Front" als `Front`; beide Rollen
// haben keinen Selektor und keine Handlung. Wirksam sind `ShowAll`
// (`unhideAllApplications:`) und `BringAllToFront` (`arrangeInFront:`) —
// nachgelesen in `menuitem_selectors_darwin.go` und `menuitem_roles.go`,
// v3.0.0-beta.23.
var wailsrollen = map[app.Menuerolle]application.Role{
	app.RolleDienste:          application.ServicesMenu,
	app.RolleAusblenden:       application.Hide,
	app.RolleAndereAusblenden: application.HideOthers,
	app.RolleAlleEinblenden:   application.ShowAll,
	app.RolleBeenden:          application.Quit,
	app.RolleWiderrufen:       application.Undo,
	app.RolleWiederholen:      application.Redo,
	app.RolleAusschneiden:     application.Cut,
	app.RolleKopieren:         application.Copy,
	app.RolleEinsetzen:        application.Paste,
	app.RolleLoeschen:         application.Delete,
	app.RolleAllesAuswaehlen:  application.SelectAll,
	app.RolleSprechenStarten:  application.StartSpeaking,
	app.RolleSprechenStoppen:  application.StopSpeaking,
	app.RolleMinimieren:       application.Minimise,
	app.RolleZoomen:           application.Zoom,
	app.RolleAlleNachVorne:    application.BringAllToFront,
}

// ohneKuerzel sind die Rollen, deren Kürzel aus Wails nicht übernommen wird.
//
// **Löschen trägt dort die Rücktaste** — ein Kürzel ohne ⌘, das jeder
// Tastendruck im Textfeld auslöste, wenn macOS es dem Menü gäbe. macOS selbst
// führt „Löschen" ohne Kürzel.
var ohneKuerzel = map[app.Menuerolle]bool{app.RolleLoeschen: true}

// menuehalter hält **ein** Menü über alle Leisten — B04 (#2223).
//
// **Ein neues `application.Menu` je Leiste ließ das vorige liegen**: `Menu.Set`
// legt für ein neues Menü ein neues `NSMenu` an und gibt das vorige nicht frei
// (`setApplicationMenu`, `application_darwin.go`, v3.0.0-beta.27), und jeder
// Eintrag behielt seinen Platz in Wails' Eintragstafel. Gemessen am nativen
// Fenster: 5 MB je Minute beim Bearbeiten, weil jeder Seitenwechsel den
// Darstellungsstand und damit die Leiste ändert.
//
// **`Destroy` und nicht `Clear`**: `Clear` nimmt die Einträge nur aus der
// Tafel; erst `Destroy` gibt auch ihr `NSMenuItem` frei. Neu gebaut wird das
// `NSMenu` danach von Wails selbst, auf dem Hauptfaden (`Menu.Update`).
//
// **Zwei Menüs im Wechsel und nicht eines**: mit einem einzigen bekam
// `setMainMenu:` jedes Mal dasselbe `NSMenu`, und AppKit hängte
// »Vollbildmodus« nur beim ersten Mal an die Darstellung (Drehbuch
// `vollbild`, 2026-10-09). Gefüllt wird immer das, das gerade nicht gilt;
// mehr als zwei liegen nie da.
//
// **Und gebaut wird nur, wenn sich die Gestalt ändert** (#2312,
// [menuehalter.nachtragen]).
type menuehalter struct {
	menues    [2]*application.Menu
	naechstes int

	// gestalt ist die der geltenden Leiste, befehle sind ihre Befehle in der
	// Reihenfolge des Baums.
	gestalt string
	befehle []*application.MenuItem
}

// erneuern trägt eine Leiste derselben Gestalt in das geltende Menü nach —
// sonst leert es das Menü, das gerade nicht gilt, und füllt es neu. **Neu
// heißt: die Hülle muss es setzen.**
func (h *menuehalter) erneuern(leiste *app.Menue, mit umsetzer) (*application.Menu, bool) {
	gestalt := gestaltVon(leiste)
	if geltend := h.menues[1-h.naechstes]; geltend != nil && gestalt == h.gestalt {
		h.nachtragen(leiste, mit)

		return geltend, false
	}

	ziel := h.menues[h.naechstes]
	if ziel == nil {
		ziel = application.NewMenu()
		h.menues[h.naechstes] = ziel
	} else {
		ziel.Destroy()
	}

	h.naechstes = 1 - h.naechstes
	h.gestalt = gestalt
	h.befehle = fuellen(ziel, leiste, mit, nil)

	return ziel, true
}

// fuellen hängt die Einträge eines Modells an ein Wails-Menü und gibt die
// Befehle zurück, an `befehle` angehängt.
func fuellen(
	ziel *application.Menu, leiste *app.Menue, mit umsetzer, befehle []*application.MenuItem,
) []*application.MenuItem {
	for _, eintrag := range leiste.Eintraege {
		switch {
		case eintrag.Untermenue != nil:
			befehle = fuellen(untermenueAus(ziel, eintrag), eintrag.Untermenue, mit, befehle)
		case eintrag.Rolle != app.KeineMenuerolle:
			mit.rollen(ziel, eintrag.Rolle, eintrag.Beschriftung)
		case eintrag.Trenner:
			ziel.AddSeparator()
		default:
			befehl := ziel.Add(eintrag.Beschriftung)
			if eintrag.Kuerzel != nil {
				befehl.SetAccelerator(kuerzeltext(eintrag.Kuerzel))
			}

			// **Gegraut wird hier gesagt, nicht erst gefragt**: Wails schaltet
			// das Selbstaktivieren der Einträge ab (`menu_darwin.go`, Zeile 30,
			// v3.0.0-beta.23), macOS fragt also nicht nach, und ein Eintrag
			// ist, was er beim Bauen gesagt bekam (UXR-F14, #2029). Das Kürzel
			// bleibt dran; ein grauer Eintrag löst es nicht aus.
			if eintrag.Gegraut {
				befehl.SetEnabled(false)
			}

			mit.binden(befehl, eintrag.Handlung)
			befehle = append(befehle, befehl)
		}
	}

	return befehle
}

// untermenueAus hängt ein Untermenü an und gibt es zum Füllen zurück — mit
// seiner Rolle, wenn es eine trägt: am Fenstermenü führt macOS die offenen
// Fenster (`menu_darwin.go`, `addWindowsMenu`, v3.0.0-beta.23).
func untermenueAus(ziel *application.Menu, eintrag *app.Menueeintrag) *application.Menu {
	unter := application.NewSubMenuItem(eintrag.Beschriftung)
	if eintrag.Rolle == app.RolleFenstermenue {
		unter.SetRole(application.WindowMenu)
	}

	ziel.Append(application.NewMenuFromItems(unter))

	return unter.GetSubmenu()
}

// kuerzeltext schreibt ein Kürzel so, wie Wails 3 es liest —
// `CmdOrCtrl+Shift+S`.
//
// **Die Reihenfolge ist gleichgültig**: Wails sortiert die Umschalter beim
// Lesen (`keys.go`, `parseAccelerator`, v3.0.0-beta.23).
func kuerzeltext(kuerzel *app.Kuerzel) string {
	text := "CmdOrCtrl+"
	if kuerzel.OhneBefehl {
		text = ""
	}
	if kuerzel.Umschalt {
		text += "Shift+"
	}

	if kuerzel.Wahl {
		text += "OptionOrAlt+"
	}

	return text + kuerzel.Taste
}
