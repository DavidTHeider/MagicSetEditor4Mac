package app_test

import (
	"bytes"
	"image/png"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Naht zum Vorschauweg** —
// #1270.
//
// Was hier zugesagt wird, ist nicht die Geschwindigkeit — die misst
// `vorschauzerlegung_bench_intern_test.go` und der Lauf im Fenster. Und es ist
// auch nicht das Fach: Verdrängen, Nebenläufigkeit und der 404 stehen bei
// `internal/vorschauweg`, wo sie hingehören.
//
// **Hier steht die Naht**: die Antwort auf `Vorschau` trägt eine **Adresse**,
// und sie führt zu genau dem Bild, über das die Antwort spricht.

// Die Adresse aus der Vorschau führt zu einem lesbaren PNG.
func TestDieVorschauadresseFuehrtZumBild(t *testing.T) {
	t.Parallel()

	karte := karteprobe{karte: &domain.CardRow{
		ID: 1, Values: []domain.Value{{Field: "name", Text: "Feuerball"}},
	}}

	anwendung := editor(&schreibprobe{}, karte)

	antwort := anwendung.Vorschau(t.Context(), "1", 1)
	if antwort.Fehler != nil {
		t.Fatalf("Befund: %+v", antwort.Fehler)
	}

	if !strings.HasPrefix(antwort.Vorschau.Bild, app.Bildwegpfad) {
		t.Fatalf("Bild = %q — das ist keine Adresse unter %q",
			antwort.Vorschau.Bild, app.Bildwegpfad)
	}

	bild, err := png.Decode(bytes.NewReader(bildbytes(t, anwendung, antwort.Vorschau.Bild)))
	if err != nil {
		t.Fatalf("das abgeholte Bild lesen: %v", err)
	}

	// **Die Maße aus der Antwort gelten auch für die Bytes.** Sonst hätte die
	// Adresse ein anderes Bild geholt als das, über das die Antwort spricht —
	// und genau das ist der Fehler, den eine falsch gemerkte Marke machte.
	if bild.Bounds().Dx() != antwort.Vorschau.Breite ||
		bild.Bounds().Dy() != antwort.Vorschau.Hoehe {
		t.Errorf("abgeholt %dx%d, angesagt %dx%d",
			bild.Bounds().Dx(), bild.Bounds().Dy(),
			antwort.Vorschau.Breite, antwort.Vorschau.Hoehe)
	}
}

// **Zwei Vorschauen bekommen zwei Adressen** — und jede führt zu ihrem Bild.
//
// Ohne diese Zusage zeigte das Fenster beim schnellen Durchgehen der
// Kartenliste die vorletzte Karte: der Browser nähme für dieselbe Adresse
// seinen zwischengespeicherten Stand, und die Antwort trägt `immutable`.
func TestJedeVorschauBekommtEineEigeneAdresse(t *testing.T) {
	t.Parallel()

	karte := karteprobe{karte: &domain.CardRow{
		ID: 1, Values: []domain.Value{{Field: "name", Text: "Feuerball"}},
	}}

	anwendung := editor(&schreibprobe{}, karte)

	klein := anwendung.Vorschau(t.Context(), "1", 1).Vorschau
	gross := anwendung.Vorschau(t.Context(), "1", 2).Vorschau

	if klein.Bild == gross.Bild {
		t.Fatalf("beide Vorschauen unter %q — dann zeigt der Browser die alte", klein.Bild)
	}

	kleinbild, err := png.Decode(bytes.NewReader(bildbytes(t, anwendung, klein.Bild)))
	if err != nil {
		t.Fatalf("die kleine Vorschau lesen: %v", err)
	}

	grossbild, err := png.Decode(bytes.NewReader(bildbytes(t, anwendung, gross.Bild)))
	if err != nil {
		t.Fatalf("die große Vorschau lesen: %v", err)
	}

	if grossbild.Bounds().Dx() <= kleinbild.Bounds().Dx() {
		t.Errorf("die doppelte Vergrößerung gibt %d px, die einfache %d",
			grossbild.Bounds().Dx(), kleinbild.Bounds().Dx())
	}
}

// bildbytes holt ein Bild über den **echten** Weg.
//
// **Über den Weg und nicht über das Fach.** Ein Test, der die Bytes direkt aus
// der Ablage nähme, ginge an dem vorbei, was das Fenster tut — und die
// Zerlegung der Adresse in eine Marke ist genau die Stelle, an der es klemmen
// kann.
//
// Ein Gegenstück steht in `vorschaubildabruf_intern_test.go`.
func bildbytes(t *testing.T, anwendung *app.App, adresse string) []byte {
	t.Helper()

	schreiben := httptest.NewRecorder()
	app.Bildweg(anwendung).ServeHTTP(schreiben,
		httptest.NewRequestWithContext(probekontext(t), http.MethodGet, adresse, http.NoBody))

	if schreiben.Code != http.StatusOK {
		t.Fatalf("Bild %q abholen: Code %d", adresse, schreiben.Code)
	}

	if art := schreiben.Header().Get("Content-Type"); art != "image/png" {
		t.Errorf("Content-Type %q — ein Browser zeigt das nicht als Bild", art)
	}

	return schreiben.Body.Bytes()
}
