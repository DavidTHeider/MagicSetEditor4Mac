package fenster

import (
	"github.com/wailsapp/wails/v3/pkg/application"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// EreignisAbgelegt ist das Ereignis, mit dem Go dem Fenster abgelegte Dateien
// meldet — mit ihren **absoluten** Pfaden und der Stelle, an der losgelassen
// wurde (#246).
//
// **Ein eigenes Ereignis, weil Wails 3 das Ablegen anders meldet als Wails
// 2.** Dort rief die Laufzeit im Fenster einen Rückruf mit `x`, `y` und den
// Pfaden; hier meldet sie es an **Go** (`WindowFilesDropped`), und nur, wenn
// über einem Element mit `data-file-drop-target` losgelassen wurde. Das
// Fenster trägt die Marke an seiner Wurzel und hört auf dieses Ereignis —
// und bekommt damit dieselben drei Angaben wie vorher (`frontend/src/fenster.ts`).
const EreignisAbgelegt = "datei:abgelegt"

// Ablage ist, was das Fenster über ein Ablegen erfährt.
type Ablage struct {
	X     int      `json:"x"`
	Y     int      `json:"y"`
	Pfade []string `json:"pfade"`
}

// ablageAus liest ein Ablegen aus Wails' Ereignis.
//
// **Eine leere Liste und kein `nil`** — im Fenster käme sonst `null` an.
func ablageAus(pfade []string, ziel *application.DropTargetDetails) Ablage {
	ablage := Ablage{Pfade: append([]string{}, pfade...)}
	if ziel != nil {
		ablage.X, ablage.Y = ziel.X, ziel.Y
	}

	return ablage
}

// nutzlast macht aus den Werten eines Ereignisses **immer eine Liste**.
//
// **Warum immer**: Wails 3 gibt keinen Wert als `null`, einen als ihn selbst
// und mehrere als Liste weiter (`EventManager.Emit`, v3.0.0-beta.23). Das
// Fenster hörte unter Wails 2 mit `(...daten)` und bekam jeden Wert einzeln;
// ein einzelner Wert, der selbst eine Liste ist, wäre auf dem neuen Weg nicht
// von zwei Werten zu unterscheiden. Eine Liste, auch eine leere, ist es immer.
func nutzlast(daten []any) []any {
	return append([]any{}, daten...)
}

// fensteroptionen übersetzt den gemerkten Fensterstand in Wails' Angabe —
// UX-14.AK-01 (#1686).
//
// **Gelesen wird in `internal/app`, entschieden dort** — hier steht nur, was
// Wails daraus macht (§2). **Maximiert und nicht Vollbild**: Vollbild nähme
// die Menüleiste mit, und diese Anwendung bedient sie (§6).
func fensteroptionen(stand app.Fensterstand) application.WebviewWindowOptions {
	optionen := application.WebviewWindowOptions{
		Name:  "haupt",
		Title: "MagicSetEditor4Mac",

		// **Die eingebaute Größe, und sie gilt bei `normal`** — bei
		// `maximiert` überschreibt sie der Startzustand, bei `zuletzt` die
		// gemerkte Zahl, sofern es eine gibt.
		Width:  mass(stand.Breite, 1280),
		Height: mass(stand.Hoehe, 860),

		// **1.024 × 700 seit Phase 5.1** (Vermerk 0076, #1412): das kleinste
		// Fenster, in dem ein Arbeitsplatz mit eingeklapptem Navigator eine
		// M15-Karte samt Inspector ganz zeigt.
		MinWidth:  mindestbreite,
		MinHeight: mindesthoehe,

		// **Ohne diese Zeile gibt es keine Pfade** (#246) — eine Webansicht
		// sieht beim Ablegen nur einen Namen; Wails hebt das für die eigene
		// Anwendung auf und meldet die absoluten Pfade an Go.
		EnableFileDrop: true,

		// **Tab erreicht jeden Knopf und jeden Verweis** — nicht nur Textfelder
		// und Auswahllisten. Ohne die macOS-Einstellung „Tastaturnavigation"
		// übersprang WebKit jeden Knopf ohne `tabindex`: die Anwendung war ohne
		// Maus nicht zu bedienen (§6), und aus der Befehlspalette fiel der Fokus
		// beim ersten Tab aus dem Dialog, weil Mantines Fokusfalle die Knöpfe
		// für erreichbar hielt. Gefunden mit der WebKit-Probe (#1828); wer die
		// Einstellung eingeschaltet hat, bemerkt keinen Unterschied.
		Mac: application.MacWindow{
			WebviewPreferences: application.MacWebviewPreferences{
				TabFocusesLinks: application.Enabled,
			},
		},
	}

	if stand.Maximiert {
		optionen.StartState = application.WindowStateMaximised
	}

	return optionen
}

// mass gibt die gemerkte Zahl — oder die eingebaute, wo keine steht.
func mass(gemerkt, eingebaut int) int {
	if gemerkt == 0 {
		return eingebaut
	}

	return gemerkt
}
