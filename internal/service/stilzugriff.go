// Wie der Stilwähler an ein Stilpaket kommt.
//
// **Eine eigene Datei**, weil `stilwahl.go` mit ihr über 400 Zeilen ginge (§2).
// Der Schnitt folgt der Frage: dort steht, **welche** Stile es gibt und was ein
// Wechsel kostet; hier, **wie** einer geöffnet wird.

package service

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// stil öffnet ein Stilpaket aus den Paketwurzeln.
//
// # Die Kurzform wird aufgelöst
//
// `stylesheet:` trägt in einem echten `.mse-set` **nicht** den Paketnamen,
// sondern den Namen ohne das Spiel davor — die Regel steht samt Beleg an
// [stilpaketnamen]. Bis zum 2026-08-23 suchte diese Stelle nur den rohen
// Wert, und damit fand die **Stilauswahl** kein Paket, während der Renderer
// daneben eines fand: er ging über [Stilspeicher.ausPaket], das die Regel
// befolgt
// (#951).
//
// **Die Regel steht an einer Stelle, und beide Verbraucher rufen sie.** Sie
// hier ein zweites Mal auszuschreiben hieße, zwei Fassungen derselben Regel zu
// führen — und die eine liefe der anderen davon.
func (w *Stilwaehler) stil(paket, spiel string) (msepaket.Stil, error) {
	if strings.TrimSpace(paket) == "" {
		return msepaket.Stil{}, domain.NewFieldError("stil", domain.SchluesselPflichtfeld,
			"Es ist kein Stil genannt.")
	}

	for _, kandidat := range stilpaketnamen(paket, spiel) {
		if stil, gefunden := w.unterNamen(kandidat); gefunden {
			return stil, nil
		}
	}

	return msepaket.Stil{}, fmt.Errorf("%w: das Stilpaket %q liegt in keinem Paketverzeichnis",
		domain.ErrNotFound, paket+"."+string(msepaket.ArtStil))
}

// unterNamen sucht **einen** Paketnamen in allen Wurzeln.
func (w *Stilwaehler) unterNamen(name string) (msepaket.Stil, bool) {
	datei := name + "." + string(msepaket.ArtStil)

	for _, wurzel := range w.wurzeln {
		geoeffnet, err := msepaket.Oeffne(filepath.Join(wurzel, datei))
		if err != nil {
			continue
		}

		stil, _ := geoeffnet.Stilmodell()
		_ = geoeffnet.Schliessen() //nolint:errcheck // das Modell ist vollständig kopiert

		return stil, true
	}

	return msepaket.Stil{}, false
}

// Stilpaket gibt ein Stilpaket im Spiel eines Sets — **leer heißt der Stil
// des Sets**, ein Name den eigenen Stil einer Karte. Die Ausfuhr baut daraus
// `set.stylesheet` und `card.stylesheet` (QA-04-F3,
// #2039).
func (w *Stilwaehler) Stilpaket(ctx context.Context, setID int64, paket string) (msepaket.Stil, error) {
	kopf, err := w.kopf(ctx, setID)
	if err != nil {
		return msepaket.Stil{}, err
	}

	if paket == "" {
		paket = kopf.Style
	}

	return w.stil(paket, kopf.Game)
}
