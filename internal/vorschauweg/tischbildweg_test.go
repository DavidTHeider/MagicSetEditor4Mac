package vorschauweg_test

// **Der Weg zum eigenen Tischbild** — Werkstück G3 von Phase 5.5
// (#1460).

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/vorschauweg"
)

// tischquelle ist ein Doppel für die Ablage.
type tischquelle struct {
	bilder  map[string][]byte
	gefragt []string
}

func (q *tischquelle) Tischbild(nutzerID string) ([]byte, bool) {
	q.gefragt = append(q.gefragt, nutzerID)
	bild, da := q.bilder[nutzerID]

	return bild, da
}

func abruf(t *testing.T, weg http.Handler, pfad string) *httptest.ResponseRecorder {
	t.Helper()

	anfrage := httptest.NewRequestWithContext(
		t.Context(), http.MethodGet, pfad, http.NoBody)

	blatt := httptest.NewRecorder()
	weg.ServeHTTP(blatt, anfrage)

	return blatt
}

func TestEinTischbildKommtUnterSeinerKennung(t *testing.T) {
	t.Parallel()

	quelle := &tischquelle{bilder: map[string][]byte{"7": []byte("PNG")}}
	weg := vorschauweg.Neu(&vorschauweg.Fach{}).MitTischbildern(quelle)

	blatt := abruf(t, weg, "/tischbild/7.png")

	switch {
	case blatt.Code != http.StatusOK:
		t.Fatalf("Code %d, erwartet 200", blatt.Code)
	case blatt.Body.String() != "PNG":
		t.Errorf("Inhalt %q", blatt.Body.String())
	case blatt.Header().Get("Content-Type") != "image/png":
		t.Errorf("Typ %q", blatt.Header().Get("Content-Type"))
	}
}

// **Dieselbe Adresse trägt morgen ein anderes Bild** — anders als bei den
// Kartenbildern, deren Marke aus dem Inhalt folgt. Wer hier `immutable`
// setzte, zeigte nach einem Wechsel wochenlang das alte.
func TestEinTischbildWirdNichtFuerImmerZwischengespeichert(t *testing.T) {
	t.Parallel()

	quelle := &tischquelle{bilder: map[string][]byte{"7": []byte("PNG")}}
	weg := vorschauweg.Neu(&vorschauweg.Fach{}).MitTischbildern(quelle)

	kopf := abruf(t, weg, "/tischbild/7.png").Header().Get("Cache-Control")
	if kopf != "no-cache" {
		t.Errorf("Cache-Control %q, erwartet no-cache", kopf)
	}
}

func TestOhneTischbildKommtVierhundertvier(t *testing.T) {
	t.Parallel()

	quelle := &tischquelle{bilder: map[string][]byte{}}
	weg := vorschauweg.Neu(&vorschauweg.Fach{}).MitTischbildern(quelle)

	if blatt := abruf(t, weg, "/tischbild/7.png"); blatt.Code != http.StatusNotFound {
		t.Errorf("Code %d, erwartet 404", blatt.Code)
	}
}

// **Ohne Quelle gibt es keinen Sturz** — dieselbe Zusage wie beim Spielpfad:
// wer die Anwendung ohne Tischbilder fährt, bekommt 404 und keinen Absturz.
func TestOhneQuelleKommtVierhundertvier(t *testing.T) {
	t.Parallel()

	weg := vorschauweg.Neu(&vorschauweg.Fach{})

	if blatt := abruf(t, weg, "/tischbild/7.png"); blatt.Code != http.StatusNotFound {
		t.Errorf("Code %d, erwartet 404", blatt.Code)
	}
}

// **Die Kennung wird aus dem Weg gelesen und nicht geraten.**
func TestDieKennungKommtAusDemWeg(t *testing.T) {
	t.Parallel()

	quelle := &tischquelle{bilder: map[string][]byte{}}
	weg := vorschauweg.Neu(&vorschauweg.Fach{}).MitTischbildern(quelle)

	abruf(t, weg, "/tischbild/42.png")

	if len(quelle.gefragt) != 1 || quelle.gefragt[0] != "42" {
		t.Errorf("gefragt %v, erwartet [42]", quelle.gefragt)
	}
}
