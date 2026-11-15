package vorschauweg_test

// **Ein Bild unter seinem Namen** — AP-12
// (#2369),
// Befund B-24.
//
// Zieht man ein Bild aus dem Fenster in den Finder, legt WebKit es als Datei
// ab und nennt sie nach dem letzten Teil der Adresse — gemessen am
// 2026-10-17 mit `ap12ziehquelle09`: „474dccaa9edc0be9.png“, die Marke. Unter
// `<Pfad><marke>/<name>.png` liefert der Weg dasselbe Bild, und die Datei
// heißt wie die Karte.

import (
	"net/http"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/vorschauweg"
)

func TestEinBildKommtAuchUnterEinemNamenHinterDerMarke(t *testing.T) {
	t.Parallel()

	var fach vorschauweg.Fach

	adresse := fach.Hinterlege([]byte("ein Bild"))
	mitNamen := strings.TrimSuffix(adresse, ".png") + "/Langtext.png"

	schreiben := hole(t, &fach, mitNamen)

	if schreiben.Code != http.StatusOK {
		t.Fatalf("Code %d unter %q", schreiben.Code, mitNamen)
	}

	if schreiben.Body.String() != "ein Bild" {
		t.Errorf("abgeholt %q", schreiben.Body.String())
	}

	if art := schreiben.Header().Get("Content-Type"); art != "image/png" {
		t.Errorf("Content-Type %q", art)
	}
}

// **Der Name sucht nichts** — eine unbekannte Marke bleibt 404, gleich, was
// dahinter steht; und ein Name allein ist keine Marke.
func TestEinNameOhneBekannteMarkeGibtVierNullVier(t *testing.T) {
	t.Parallel()

	var fach vorschauweg.Fach

	adresse := fach.Hinterlege([]byte("ein Bild"))
	marke := strings.TrimSuffix(strings.TrimPrefix(adresse, vorschauweg.Pfad), ".png")

	for _, falsch := range []string{
		vorschauweg.Pfad + "unbekannt/" + marke + ".png",
		vorschauweg.Pfad + "Langtext.png",
		vorschauweg.Pfad + "/Langtext.png",
	} {
		if code := hole(t, &fach, falsch).Code; code != http.StatusNotFound {
			t.Errorf("Code %d unter %q, erwartet 404", code, falsch)
		}
	}
}
