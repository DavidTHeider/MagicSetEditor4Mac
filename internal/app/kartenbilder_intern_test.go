package app

// **Ein Bild, das sich nicht kodieren lässt, fehlt in der Liste** — Werkstück
// 2 von Phase 3.5
// (#1166).
//
// **Ein interner Lauf, weil die Kodiernaht intern ist.** Sie von außen
// erreichbar zu machen hieße, eine Naht für einen Test zu öffnen, die im
// Betrieb niemand braucht — und `png.Encode` scheitert an einem
// `bytes.Buffer` nicht von selbst.
//
// **Und die Liste bleibt heil**: die übrigen Bilder kommen an, dieses fehlt.
// Eine Liste, die wegen eines Kodierfehlers gar nichts zeigte, wäre der
// schlechtere Tausch (§8).

import (
	"image"
	"io"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestEinNichtKodierbaresBildFehltInDerListe(t *testing.T) {
	t.Parallel()

	anwendung := NeueApp("")
	anwendung.packen = func(io.Writer, image.Image) error { return errKodierung }
	anwendung.vorschaubilder = service.NeuesVorschaubildwerk(probenbild{})

	bilder := anwendung.bilderZu(
		[]*domain.CardRow{{
			ID: 1, Values: []domain.Value{{Field: "kunst", Text: "eins.png"}},
		}},
		map[string]bool{"1": true}, "kunst", "/wo/das/set.mse-set", "")

	if len(bilder) != 0 {
		t.Errorf("es kam ein Bild: %v", bilder)
	}
}

// probenbild gibt zu jedem Verweis ein winziges Bild.
type probenbild struct{}

func (probenbild) Bild(string) (image.Image, error) {
	return image.NewNRGBA(image.Rect(0, 0, 4, 4)), nil
}

// **Und die Gegenprobe: mit heiler Kodierung kommt es an.**
//
// Ohne sie hätte der Lauf darüber zwei Lesarten — die Kodierung scheitert,
// oder das Bild wäre ohnehin nicht gekommen (§10).
func TestMitHeilerKodierungKommtDasBildAn(t *testing.T) {
	t.Parallel()

	anwendung := NeueApp("")
	anwendung.vorschaubilder = service.NeuesVorschaubildwerk(probenbild{})

	bilder := anwendung.bilderZu(
		[]*domain.CardRow{{
			ID: 1, Values: []domain.Value{{Field: "kunst", Text: "eins.png"}},
		}},
		map[string]bool{"1": true}, "kunst", "/wo/das/set.mse-set", "")

	if len(bilder) != 1 {
		t.Errorf("es kam kein Bild: %v", bilder)
	}
}
