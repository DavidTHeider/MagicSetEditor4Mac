package app

// **Das eigene Tischbild** — Werkstück G3 von Phase 5.5
// (#1460).

import (
	"image"
	"image/color"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// tischprobe gibt ein winziges Bild.
//
// **Eine eigene und nicht `einfarbig()` aus `abnahmesicherung_test.go`**: die
// steht im Paket `app_test`, dieser Lauf im Paket `app` — er braucht
// `tischbilddatei`, und die ist nicht exportiert.
func tischprobe() image.Image {
	bild := image.NewNRGBA(image.Rect(0, 0, 4, 3))
	bild.SetNRGBA(0, 0, color.NRGBA{R: 200, A: 255})

	return bild
}

// **Beide Seiten nennen denselben Dateinamen.**
//
// `service` bildet ihn beim Ablegen, `app` beim Lesen — und keine der beiden
// gibt ihn der anderen. Liefen sie auseinander, legte die Anwendung ein Bild
// ab, das sie nie wiederfände; auffallen würde es niemandem, weil „kein Bild"
// eine gültige Antwort ist (§8).
func TestBeideSeitenNennenDenselbenDateinamen(t *testing.T) {
	t.Parallel()

	ablage := service.NeueBildablage(t.TempDir())

	ausService, fehler := ablage.TischbildAblegen(7, tischprobe())
	if fehler != nil {
		t.Fatalf("ablegen: %v", fehler)
	}

	if ausApp := tischbilddatei(7); ausApp != ausService {
		t.Errorf("app nennt %q, service %q", ausApp, ausService)
	}
}

// **Die Adresse trägt die Kennung und den Vorsatz.**
func TestDieTischbildadresseTraegtKennungUndVorsatz(t *testing.T) {
	t.Parallel()

	adresse := Tischbildpfad + "7.png"

	switch {
	case !strings.HasPrefix(adresse, "/tischbild/"):
		t.Errorf("Vorsatz fehlt: %q", adresse)
	case !strings.HasSuffix(adresse, "/7.png"):
		t.Errorf("Kennung fehlt: %q", adresse)
	}
}
