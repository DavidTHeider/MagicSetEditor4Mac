// **Wo die eigenen Dateien der Anwendung liegen** — Bilder und gezeichnete
// Symbole.
//
// **Eine eigene Datei und trotzdem Composition Root.** `app.go` stand an der
// Zeilengrenze aus §2; der Schnitt folgt der Frage: dort **die Anwendung**,
// hier die **Ablagen** — und sie sind eine Frage, nicht zwei: beide schreiben
// in dasselbe Verzeichnis, und genau das ist die Aussage.

package app

import (
	"log/slog"
	"path/filepath"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// ablagen richtet Bild- und Symbolablage ein und gibt die Bildablage.
//
// # Eine Wurzel, zwei Ablagen
//
// **Eine Ablage, drei Nutzer**: sie sagt, wo die Bilder liegen, die jemand in
// dieser Anwendung eingesetzt hat
// ([0029](../../docs/entscheidungen/0029-bilder-liegen-neben-dem-set.md)) —
// `BildSetzen` legt hinein, der Planbauer sucht darin, der Schreiber gibt sie
// weiter. Eine zweite wäre ein zweiter Ort für dieselben Dateien.
//
// **Und die Symbolablage schreibt in dasselbe Verzeichnis** (#890), weil
// `Beiwerk` beim Sichern alles von dort ins geschriebene Paket kopiert. Eine
// eigene Wurzel käme nie mit — und der Fehler fiele erst auf, wenn jemand die
// Datei woanders öffnet.
//
// # Ein Fehler hält den Start nicht auf
//
// Derselbe Grund wie beim Paketverzeichnis: wer nie ein Bild einsetzt und nie
// ein Symbol zeichnet, braucht sie nicht, und `Ablegen` meldet sich, wenn es
// so weit ist.
//
// **Er verschwindet aber auch nicht**
// (#857):
// unter `_` verworfen, fände, wer später ein Bild einsetzt und eine
// unverständliche Meldung bekommt, nirgends den Grund.
func (a *App) ablagen() *service.Bildablage {
	wurzel, err := a.bilderort()
	if err != nil {
		slog.Default().Error("das Bilderverzeichnis ließ sich nicht bestimmen",
			"grund", err.Error())
	}

	ablage := service.NeueBildablage(wurzel)

	a.bildablage = ablage
	a.symbole = service.NeueSymbolablage(ablage)

	a.bildvorratAnlegen()

	return ablage
}

// bildvorratAnlegen baut den Vorrat für Bildvorschläge (A12, #1451).
//
// # Warum im Zwischenspeicher und nicht neben dem Bestand
//
// **Weil ein Vorschlag kein Bestand ist.** Was unter `Application Support`
// liegt, geht in die Sicherung und in jedes Paket, das jemand weitergibt —
// und das Issue verlangt das Gegenteil: „temporäre Assets nie ungefragt ins
// Set/Backup/Peerpaket."
//
// `os.UserCacheDir` ist unter macOS `~/Library/Caches`. Was dort liegt, darf
// das System jederzeit wegräumen, und genau das ist die Zusage, die ein
// Vorschlag mit Frist ohnehin gibt.
//
// **Ohne Vorrat läuft alles andere weiter.** Wo sich kein
// Zwischenspeicherverzeichnis ermitteln lässt, gibt es keine Bildvorschläge
// — und jede Bindung sagt das, statt zu stolpern.
func (a *App) bildvorratAnlegen() {
	zwischen, err := zwischenspeicherort()
	if err != nil {
		slog.Default().Error("der Zwischenspeicher ließ sich nicht bestimmen",
			"grund", err.Error())

		return
	}

	vorrat, err := service.NeuerBildvorrat(
		filepath.Join(zwischen, "MagicSetEditor4Mac", "bildvorschlaege"),
		service.UeblicheBildgrenzen(), time.Now)
	if err != nil {
		slog.Default().Error("der Bildvorrat ließ sich nicht anlegen",
			"grund", err.Error())

		return
	}

	a.bildvorrat = vorrat
}
