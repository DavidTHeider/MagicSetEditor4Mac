//go:build darwin

package fenster

// **Der Teilen-Dialog von macOS** — NET-04
// (#1708),
// entschieden in
// [0155](../../docs/entscheidungen/0155-teilen-ist-eine-kopie.md).
//
// **`NSSharingServicePicker` und nicht eine eigene Liste** (§6: nativ und
// nicht nachgebaut): welche Ziele es gibt — Mail, Nachrichten, AirDrop,
// Notizen und was sonst eingerichtet ist —, weiß das System, und es fragt
// selbst nach Empfänger und Bestätigung. Nicht eingerichtete Ziele zeigt es
// gar nicht erst.
//
// **Ohne Lauf, und das steht in `.covignore`**: der Dialog braucht ein
// echtes Fenster. Was entscheidet — was geteilt wird, wie es heißt, was
// mitgeht —, steht in `internal/app` und `internal/domain` und wird dort
// gemessen; die Verdrahtung hält `verdrahtung_test.go`.

import (
	"errors"

	"github.com/tmc/apple/appkit"
	"github.com/tmc/apple/corefoundation"
	"github.com/tmc/apple/foundation"
	"github.com/tmc/apple/objc"
	"github.com/tmc/apple/objectivec"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// errOhneFenster heißt: es gibt kein Fenster, an dem der Dialog hängen könnte.
var errOhneFenster = errors.New("kein Fenster für den Teilen-Dialog")

// systemTeilen zeigt den Dialog am Fenster — für einen Text oder eine Datei.
//
// **Oben in der Mitte verankert**, wo die Leiste steht: dort erwartet man
// ein Menü, das zu einer Handlung gehört.
func systemTeilen(fenster application.Window, pfad, text string) error {
	if fenster == nil {
		return errOhneFenster
	}

	return application.InvokeSyncWithError(func() error {
		zeiger := fenster.NativeWindow()
		if zeiger == nil {
			return errOhneFenster
		}

		ansicht := appkit.NSWindowFromID(objc.ID(uintptr(zeiger))).ContentView()

		var eintrag objectivec.IObject = foundation.NewURLFileURLWithPath(pfad)
		if text != "" {
			eintrag = foundation.NewStringWithString(text)
		}

		waehler := appkit.NewSharingServicePickerWithItems(foundation.NewArrayWithObject(eintrag))

		rahmen := ansicht.Bounds()
		anker := corefoundation.CGRect{
			Origin: corefoundation.CGPoint{X: rahmen.Size.Width / 2, Y: rahmen.Size.Height - 8},
			Size:   corefoundation.CGSize{Width: 1, Height: 1},
		}

		waehler.ShowRelativeToRectOfViewPreferredEdge(anker, ansicht, foundation.NSMinYEdge)

		return nil
	})
}
