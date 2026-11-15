package vorschauweg_test

// **Der Weg zum eigenen Fensterbild** — UX-09.AK-01
// (#1681).

import (
	"net/http"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/vorschauweg"
)

// fensterquelle ist ein Doppel für die Ablage.
type fensterquelle struct {
	bilder map[string][]byte
}

func (q *fensterquelle) Fensterbild(nutzerID string) ([]byte, bool) {
	bild, da := q.bilder[nutzerID]

	return bild, da
}

func TestEinFensterbildKommtUnterSeinerKennung(t *testing.T) {
	t.Parallel()

	weg := vorschauweg.Neu(&vorschauweg.Fach{}).MitFensterbildern(
		&fensterquelle{bilder: map[string][]byte{"7": []byte("PNG")}})

	blatt := abruf(t, weg, "/fensterbild/7.png")

	switch {
	case blatt.Code != http.StatusOK:
		t.Fatalf("Code %d, erwartet 200", blatt.Code)
	case blatt.Body.String() != "PNG":
		t.Errorf("Inhalt %q", blatt.Body.String())
	case blatt.Header().Get("Content-Type") != "image/png":
		t.Errorf("Typ %q", blatt.Header().Get("Content-Type"))
	case blatt.Header().Get("Cache-Control") != "no-cache":
		t.Errorf("Cache-Control %q", blatt.Header().Get("Cache-Control"))
	}
}

// **Tisch und Fenster sind zwei Fächer**: unter der Adresse des einen kommt
// nie das Bild des anderen.
func TestFensterbildUndTischbildKommenAusIhremFach(t *testing.T) {
	t.Parallel()

	weg := vorschauweg.Neu(&vorschauweg.Fach{}).
		MitFensterbildern(&fensterquelle{bilder: map[string][]byte{"7": []byte("FENSTER")}}).
		MitTischbildern(&tischquelle{bilder: map[string][]byte{"7": []byte("TISCH")}})

	if inhalt := abruf(t, weg, "/fensterbild/7.png").Body.String(); inhalt != "FENSTER" {
		t.Errorf("Fenster %q", inhalt)
	}

	if inhalt := abruf(t, weg, "/tischbild/7.png").Body.String(); inhalt != "TISCH" {
		t.Errorf("Tisch %q", inhalt)
	}
}

func TestOhneFensterbildOderQuelleKommtVierhundertvier(t *testing.T) {
	t.Parallel()

	ohneQuelle := vorschauweg.Neu(&vorschauweg.Fach{})
	if blatt := abruf(t, ohneQuelle, "/fensterbild/7.png"); blatt.Code != http.StatusNotFound {
		t.Errorf("ohne Quelle: Code %d", blatt.Code)
	}

	leer := vorschauweg.Neu(&vorschauweg.Fach{}).MitFensterbildern(
		&fensterquelle{bilder: map[string][]byte{}})
	if blatt := abruf(t, leer, "/fensterbild/7.png"); blatt.Code != http.StatusNotFound {
		t.Errorf("ohne Bild: Code %d", blatt.Code)
	}
}

// **Die Weiche kennt den Vorsatz** — sonst ginge `/fensterbild/…` an die
// Oberfläche weiter und käme als `index.html` zurück.
func TestDieWeicheNimmtFensterbilder(t *testing.T) {
	t.Parallel()

	weg := vorschauweg.Neu(&vorschauweg.Fach{}).MitFensterbildern(
		&fensterquelle{bilder: map[string][]byte{"7": []byte("PNG")}})
	weiter := http.HandlerFunc(func(schreiben http.ResponseWriter, _ *http.Request) {
		schreiben.WriteHeader(http.StatusTeapot)
	})

	if blatt := abruf(t, weg.Vor(weiter), "/fensterbild/7.png"); blatt.Code != http.StatusOK {
		t.Errorf("Code %d, erwartet 200", blatt.Code)
	}
}
