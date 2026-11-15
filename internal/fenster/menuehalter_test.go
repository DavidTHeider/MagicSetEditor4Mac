package fenster

// **Eine neue Leiste füllt dasselbe Menü** — B04
// (#2223).
//
// Bis dahin baute jede Meldung des Darstellungsstands ein neues
// `application.Menu`, und `Menu.Set` legte dafür ein neues `NSMenu` an. Das
// vorige gab Wails nie frei (`setApplicationMenu`, `application_darwin.go`,
// v3.0.0-beta.27: `m.applicationMenu` wird überschrieben), samt jedem Eintrag
// darunter und seinem Platz in Wails' Eintragstafel. Am nativen Fenster lagen
// nach fünf Minuten Bearbeiten 14.454 `MenuItem`, 11.677 `NSMenuItem` und
// 3.891 `NSMenu` im Speicher (`heap`), und der Prozess wuchs um 5 MB je
// Minute.

import (
	"testing"

	"github.com/wailsapp/wails/v3/pkg/application"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// eintraegeIn zählt die Einträge auf der obersten Ebene.
func eintraegeIn(menue *application.Menu) int {
	zahl := 0
	for menue.ItemAt(zahl) != nil {
		zahl++
	}

	return zahl
}

func TestEineNeueLeisteFuelltDasselbeMenue(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	fenster := &probehuelle{}
	mit, _ := probeumsetzer()

	verdrahten(anwendung, fenster, mit)

	for _, sprache := range []string{"en", "de", "en"} {
		if antwort := anwendung.SpracheSetzen(sprache); antwort.Fehler != nil {
			t.Fatalf("Sprache %s: %+v", sprache, antwort.Fehler)
		}
	}

	if len(fenster.leisten) < 3 {
		t.Fatalf("gesetzt wurden %d Leisten, erwartet wenigstens drei", len(fenster.leisten))
	}

	// **Zwei Menüs, und nie zweimal dasselbe hintereinander.** Mit einem
	// einzigen setzte Wails jede Leiste über dasselbe `NSMenu`, und AppKit
	// hängte »Vollbildmodus« nur beim ersten `setMainMenu:` an (Drehbuch
	// `vollbild`, 2026-10-09). Mehr als zwei hieße wieder: das vorige
	// bleibt liegen.
	verschiedene := map[*application.Menu]bool{}
	for stelle, leiste := range fenster.leisten {
		verschiedene[leiste] = true

		if stelle > 0 && leiste == fenster.leisten[stelle-1] {
			t.Errorf("Leiste %d ist dasselbe Menü wie die davor — AppKit hängte »Vollbildmodus« nicht wieder an", stelle)
		}
	}

	if len(verschiedene) != 2 {
		t.Errorf("gesetzt wurden %d verschiedene Menüs, erwartet zwei im Wechsel", len(verschiedene))
	}

	frisch := menueAus(app.Menueleiste(anwendung.Fensterhandlungen(), app.MenuetexteFuer("en"), nil), mit)
	if steht, will := eintraegeIn(fenster.leisten[len(fenster.leisten)-1]), eintraegeIn(frisch); steht != will {
		t.Errorf("das Menü trägt %d Einträge, eine frische Leiste %d — die alten blieben stehen", steht, will)
	}

	if fenster.leisten[len(fenster.leisten)-1].FindByLabel(app.MenuetexteFuer("en").Ablage) == nil {
		t.Error("die englische Leiste fehlt")
	}
}
