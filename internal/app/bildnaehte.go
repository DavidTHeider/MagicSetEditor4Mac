package app

// **Die Nähte um Bilder** — Symbolablage, Bildablage, PNG-Kodierung und das
// Fach der Vorschauen.
//
// **Eine eigene Datei und ein eingebetteter Verbund**, seit `anwendung.go`
// mit der Zuordnungsnaht (U9, #1420) über die Zeilengrenze aus §2 ging —
// dieselbe Form wie `uebersichtsnaehte`. Eingebettet heißt: `a.symbole`
// bleibt `a.symbole`; der Schnitt ist einer der **Datei**, nicht der Naht.

import (
	"image"
	"io"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/vorschauweg"
)

type bildnaehte struct {
	// symbole legt gezeichnete `.mse-symbol`-Dateien ab
	// (#890).
	symbole Symbolablage

	// bildablage legt Bilder ab, die jemand eingesetzt hat
	// (#246).
	//
	// **Eine Naht und kein Pfad**: wo das Benutzerverzeichnis liegt, weiß
	// diese Schicht — aber ein Test hat keines.
	bildablage Bildablage

	// packen ist die Naht zur PNG-Kodierung.
	//
	// Als Feld und aus demselben Grund wie `oeffnen`: mit einem
	// `bytes.Buffer` kann `png.Encode` nicht scheitern, und der Fehlerzweig
	// wäre eine Behauptung darüber, was im Ernstfall geschieht — im Betrieb
	// schreibt irgendwann jemand auf etwas, das voll ist.
	packen func(io.Writer, image.Image) error

	// bildfach hält die gerasterten Vorschauen für die Assetleitung
	// (#1270).
	//
	// **Nicht als Zeiger**: es ist ein Teil dieser Anwendung und keine Naht,
	// die jemand austauscht. Sein Nullwert ist gebrauchsfertig.
	bildfach vorschauweg.Fach

	// bildvorrat hält erzeugte Bilder, bis jemand sie übernimmt oder
	// verwirft (A12, #1451) — nil heißt: es gibt keine Bildvorschläge.
	//
	// **Als Zeiger und mit einer Prüfung an jeder Bindung**: sein Ort liegt
	// im Zwischenspeicher des Benutzers, und den gibt es nicht überall. Wo
	// er fehlt, sagt die Bindung es, statt beim ersten Vorschlag zu
	// stolpern.
	bildvorrat *service.Bildvorrat

	// teilen hält die Teilungen dieser Sitzung — `teilen.go` (NET-04, #1708).
	teilen teilfelder
}
