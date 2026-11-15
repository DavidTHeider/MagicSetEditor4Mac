package app_test

import (
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Die Menüs stehen in der Ordnung von macOS** — UXR-02
// (#1910).
//
// # Die Quellen
//
// Apple, Human Interface Guidelines, „The menu bar" (nachgelesen am
// 2026-10-02 über `developer.apple.com/tutorials/data/design/
// human-interface-guidelines/the-menu-bar.json`): die Leiste führt Anwendung,
// Ablage, Bearbeiten, Format, Darstellung, **die eigenen Menüs der
// Anwendung**, Fenster und Hilfe; die eigenen stehen **zwischen Darstellung
// und Fenster** (Abschnitt *App-specific menus*).
//
// **MSE2 stellt seine eigenen ebenso hinter Bearbeiten**: jede Tafel fügt ihr
// Menü an Stelle 2 ein, nach Ablage und Bearbeiten
// (`mb->Insert(2, menuCard, …)`, `src/gui/set/cards_panel.cpp`, Zeile 208,
// Fassung f9d9356 vom 2021-07-06; ebenso `keywords_panel.cpp`,
// `set_info_panel.cpp`, `stats_panel.cpp`). Eine Darstellung kennt MSE2 nicht.
//
// **Bis hierher stand *Set* zwischen Ablage und Bearbeiten** — gegen beide.
// Wer ⌘C sucht, fand an der zweiten Stelle von links die Karten.
func TestDieMenuesStehenInDerOrdnungVonMacOS(t *testing.T) {
	t.Parallel()

	for _, sprache := range []string{"de", "en"} {
		texte := app.MenuetexteFuer(sprache)
		leiste := app.Menueleiste(app.Menuhandlungen{}, texte, nil)

		gesehen := make([]string, 0, len(leiste.Eintraege))
		for _, eintrag := range leiste.Eintraege {
			gesehen = append(gesehen, eintrag.Beschriftung)
		}

		erwartet := []string{
			texte.Anwendung, texte.Ablage, texte.Bearbeiten, texte.Darstellung,
			texte.Set, texte.Fenster, texte.Hilfe,
		}
		if !slices.Equal(gesehen, erwartet) {
			t.Errorf("%s: erwartet %v, gesehen %v", sprache, erwartet, gesehen)
		}
	}
}
