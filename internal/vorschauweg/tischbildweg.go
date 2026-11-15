package vorschauweg

// **Der Weg zum eigenen Tischbild** — Werkstück G3 von Phase 5.5
// (#1460):
// „neutrale und rechtlich verwendbare Hintergründe **einschließlich Import**".
//
// # Warum ein Weg und keine Daten-URI
//
// **Weil ein Hintergrund groß ist.** G2 hat gemessen, was 182 KB als Base64
// über die Wails-Brücke kosten: **554 ms**
// (#1270).
// Ein Tischbild von 1920 Punkten wiegt ein Vielfaches davon, und es käme bei
// jedem Öffnen des Tisches erneut.

import (
	"net/http"
	"strings"
)

// Tischpfad ist der Anfang jeder Tischbildadresse.
const Tischpfad = "/tischbild/"

// Tischbildquelle gibt das abgelegte Tischbild eines Menschen.
//
// **Eine Schnittstelle und keine Ablage** — dieses Blatt kennt weder
// `service` noch Dateien (§2); es weiß nur, dass jemand auf eine Kennung hin
// Bytes liefert.
type Tischbildquelle interface {
	Tischbild(nutzerID string) ([]byte, bool)
}

// MitTischbildern hängt die Quelle an den Weg.
func (w *Weg) MitTischbildern(quelle Tischbildquelle) *Weg {
	w.tischbilder = quelle

	return w
}

// tischbild beantwortet den Abruf eines Tischbildes.
func (w *Weg) tischbild(schreiben http.ResponseWriter, anfrage *http.Request) {
	if w.tischbilder == nil {
		http.NotFound(schreiben, anfrage)

		return
	}

	eigenesBild(schreiben, anfrage, Tischpfad, w.tischbilder.Tischbild)
}

// eigenesBild beantwortet den Abruf eines Bildes, das ein Mensch mitgebracht
// hat — Tisch oder Fenster.
//
// **Eine Antwort für beide**: dieselbe Kennung, dasselbe PNG aus der eigenen
// Ablage, derselbe Grund gegen `immutable`.
func eigenesBild(
	schreiben http.ResponseWriter, anfrage *http.Request, vorsatz string,
	hole func(nutzerID string) ([]byte, bool),
) {
	kennung := strings.TrimSuffix(
		strings.TrimPrefix(anfrage.URL.Path, vorsatz), ".png")

	bild, da := hole(kennung)
	if !da {
		http.NotFound(schreiben, anfrage)

		return
	}

	schreiben.Header().Set("Content-Type", "image/png")

	// **`no-cache` und nicht `immutable`**: hier trägt dieselbe Adresse
	// morgen ein anderes Bild; wer sie für immer zwischenspeicherte, zeigte
	// nach einem Wechsel wochenlang das alte. Die Kartenbilder, deren Marke aus
	// dem Inhalt folgt, gehen seit E06 gar nicht mehr in den Speicher der
	// Webansicht (`no-store`).
	schreiben.Header().Set("Cache-Control", "no-cache")

	// #nosec G705 -- ein PNG aus der eigenen Ablage, mit `Content-Type:
	// image/png` ausgeliefert; es kommt aus einer Datei, die die Anwendung
	// selbst geschrieben hat
	//
	//nolint:errcheck // ein abgebrochener Abruf ist der Normalfall
	_, _ = schreiben.Write(bild)
}
