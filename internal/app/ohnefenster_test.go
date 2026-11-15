package app_test

// **Ohne Fenster sagt die Anwendung es, statt abzustürzen** — ENG-02
// (#1727).
//
// Bis `internal/fenster` die Nähte setzt, gelten die aus `ohnefenster.go`:
// eine Meldung verhallt, ein Dialog und die Zwischenablage antworten mit
// einem Fehler. Unter Wails 2 wäre an derselben Stelle ein `nil`-Kontext in
// die Laufzeit gegangen.

import (
	"context"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// kopierablage ist eine Kartenablage, die kopiert und sonst nichts kann —
// das Einfügen scheitert an der Zwischenablage, bevor es sie fragt.
type kopierablage struct{ app.Ablagequelle }

func (kopierablage) Kopieren(context.Context, int64, []int64) (string, error) {
	return "abgelegt", nil
}

func TestOhneFensterSagtDieAnwendungEsStattAbzustuerzen(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	anwendung.MitKartenablage(kopierablage{})

	// **Die Meldung verhallt** — ein Menüeintrag ohne Fenster tut nichts, und
	// das hält nichts auf.
	anwendung.Fensterhandlungen().SetSichern()

	if antwort := anwendung.BestandSichern(context.Background(), "", false); antwort.Fehler == nil {
		t.Error("ein Dialog ohne Fenster kam durch")
	}

	if antwort := anwendung.AblageEinfuegen(context.Background(), "1"); antwort.Fehler == nil {
		t.Error("eine Zwischenablage ohne Fenster wurde gelesen")
	}

	if antwort := anwendung.AblageKopieren(context.Background(), "1", []string{"2"}); antwort.Fehler == nil {
		t.Error("eine Zwischenablage ohne Fenster wurde beschrieben")
	}
}
