package fenster

import (
	"strconv"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// nachtragen bringt die Befehle des geltenden Menüs auf eine Leiste derselben
// Gestalt — Beschriftung, Gegrautes und Handlung (#2312).
//
// # Warum nicht neu bauen
//
// **Wails gibt die Beschriftung eines neuen Eintrags nie frei**: `newMenuItem`
// gibt nur `tooltip` zurück, nicht `label` (`menuitem_darwin.go`, Zeilen
// 15–42, v3.0.0-beta.27). Jeder Neubau ließ so eine Zeichenkette je Eintrag
// liegen, und gebaut wurde bei jedem Wechsel des Darstellungsstands und jedem
// aufgeschlagenen Set — am nativen Fenster rund 650 KB in sechs Minuten
// Bearbeiten. `setMenuItemLabel` gibt seine Zeichenkette frei, und
// `setMenuItemDisabled` braucht keine.
//
// **Geändert wird nur, was sich ändert**: jeder Ruf an Wails geht über den
// Hauptfaden (`dispatch_sync`).
func (h *menuehalter) nachtragen(leiste *app.Menue, mit umsetzer) {
	for stelle, eintrag := range befehleIn(leiste, nil) {
		befehl := h.befehle[stelle]
		if befehl.Label() != eintrag.Beschriftung {
			befehl.SetLabel(eintrag.Beschriftung)
		}

		if befehl.Enabled() == eintrag.Gegraut {
			befehl.SetEnabled(!eintrag.Gegraut)
		}

		mit.binden(befehl, eintrag.Handlung)
	}
}

// befehleIn sammelt die Befehle einer Leiste in der Reihenfolge, in der
// [fuellen] sie anlegt.
func befehleIn(leiste *app.Menue, befehle []*app.Menueeintrag) []*app.Menueeintrag {
	for _, eintrag := range leiste.Eintraege {
		switch {
		case eintrag.Untermenue != nil:
			befehle = befehleIn(eintrag.Untermenue, befehle)
		case eintrag.Rolle != app.KeineMenuerolle, eintrag.Trenner:
		default:
			befehle = append(befehle, eintrag)
		}
	}

	return befehle
}

// gestaltVon beschreibt, was sich an einer Leiste nur durch einen Neubau
// ändern lässt: die Folge der Einträge, ihre Art, Untermenüs mit Titel und
// Rolle, Rollen mit Beschriftung und die Kürzel. **Nicht darin**: die
// Beschriftung eines Befehls, sein Gegrautes und seine Handlung —
// [menuehalter.nachtragen] setzt sie an Ort und Stelle.
func gestaltVon(leiste *app.Menue) string {
	var gestalt strings.Builder

	gestaltSchreiben(&gestalt, leiste)

	return gestalt.String()
}

func gestaltSchreiben(gestalt *strings.Builder, leiste *app.Menue) {
	for _, eintrag := range leiste.Eintraege {
		switch {
		case eintrag.Untermenue != nil:
			gestalt.WriteString("U" + strconv.Quote(eintrag.Beschriftung) + strconv.Itoa(int(eintrag.Rolle)) + "[")
			gestaltSchreiben(gestalt, eintrag.Untermenue)
			gestalt.WriteString("]")
		case eintrag.Rolle != app.KeineMenuerolle:
			gestalt.WriteString("R" + strconv.Itoa(int(eintrag.Rolle)) + strconv.Quote(eintrag.Beschriftung))
		case eintrag.Trenner:
			gestalt.WriteString("-")
		default:
			gestalt.WriteString("B")
			if eintrag.Kuerzel != nil {
				gestalt.WriteString(strconv.Quote(kuerzeltext(eintrag.Kuerzel)))
			}
		}
	}
}
