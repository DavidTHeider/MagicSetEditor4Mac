package netz_test

// **Wie oft der Speicher antwortet** — OPS-05.AK-01
// (#1722):
// „Metriken umfassen … Cachetreffer". Gezählt wird über den ganzen Prozess,
// wie beim Messwerk in `service` — und deshalb läuft dieser Test **nicht
// parallel**: er misst den Unterschied, und ein Nachbar, der gleichzeitig
// holte, zählte mit.

import (
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

func TestDerSpeicherZaehltTrefferUndGaenge(t *testing.T) {
	vorher := netz.Speicherzahlen()

	jetzt := time.Date(2026, time.September, 26, 12, 0, 0, 0, time.UTC)
	speicher := netz.NeuerSpeicher(t.TempDir(), time.Hour, func() time.Time { return jetzt })

	gut := func() ([]byte, error) { return []byte(`{}`), nil }
	schlecht := func() ([]byte, error) { return nil, errors.New("kein Netz") }

	for _, schritt := range []func() ([]byte, error){gut, gut, gut} {
		if _, _, fehler := speicher.Hole(t.Context(), "a", schritt); fehler != nil {
			t.Fatal(fehler)
		}
	}

	jetzt = jetzt.Add(2 * time.Hour)
	if _, _, fehler := speicher.Hole(t.Context(), "a", schlecht); fehler != nil {
		t.Fatal(fehler)
	}

	if _, _, fehler := speicher.Hole(t.Context(), "b", schlecht); fehler == nil {
		t.Fatal("ohne Stand und ohne Netz kam keine Klage")
	}

	nachher := netz.Speicherzahlen()
	if nachher.Treffer-vorher.Treffer != 2 || nachher.Geholt-vorher.Geholt != 1 ||
		nachher.Veraltet-vorher.Veraltet != 1 || nachher.Fehlgeschlagen-vorher.Fehlgeschlagen != 1 {
		t.Errorf("vorher %+v, nachher %+v", vorher, nachher)
	}
}
