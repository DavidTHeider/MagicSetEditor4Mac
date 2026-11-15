package service_test

// **Die Aufsicht schickt eine Massenänderung durch den rechnenden Weg** —
// #2516. Wo der Feldschreiber rechnet, rechnet er auch hier; die Vorschau
// zeigt, was gespeichert würde, und die Setgrenze gilt wie bisher (#808).

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// rechnendeprobe rechnet wie ein Wertskript, das eckige Klammern setzt.
type rechnendeprobe struct {
	*feldprobe

	karten   []int64
	erwartet []domain.Feldvorwert
	angelegt []int64
}

func (r *rechnendeprobe) MassenSetzen(
	_ context.Context, karteIDs []int64, _, _ string, erwartet []domain.Feldvorwert,
) (string, error) {
	r.karten, r.erwartet = karteIDs, erwartet

	return "gerechnet7", nil
}

func (*rechnendeprobe) Gerechnet(_ context.Context, _ *domain.CardRow, _, wert string) string {
	return "[" + wert + "]"
}

func rechnendeAufsicht() (*service.Kartenaufsicht, *rechnendeprobe) {
	probe := &rechnendeprobe{feldprobe: &feldprobe{}}
	sicht := &kartensicht{karten: map[int64]int64{1: 10, 2: 10, 3: 20}}

	return service.NeueKartenaufsicht(sicht, &lebenszyklusprotokoll{}).MitFeldweg(probe, probe), probe
}

func TestDieAufsichtRechnetDieMassenaenderung(t *testing.T) {
	t.Parallel()

	t.Run("ohne Erwartung", func(t *testing.T) {
		t.Parallel()

		auge, probe := rechnendeAufsicht()

		vorgang, fehler := auge.FelderSetzen(t.Context(), 10, []int64{1, 2}, "name", "Glut")
		if fehler != nil || vorgang != "gerechnet7" || !slices.Equal(probe.karten, []int64{1, 2}) || probe.erwartet != nil {
			t.Errorf("%q, %v, Karten %v, Erwartung %v", vorgang, fehler, probe.karten, probe.erwartet)
		}

		if len(probe.massen) != 0 {
			t.Errorf("der rohe Weg schrieb %v", probe.massen)
		}
	})

	t.Run("mit Erwartung", func(t *testing.T) {
		t.Parallel()

		auge, probe := rechnendeAufsicht()
		erwartet := []domain.Feldvorwert{{KarteID: 1, Wert: "alt", GabEs: true}}

		vorgang, fehler := auge.FelderSetzenMitErwartung(t.Context(), 10, erwartet, "name", "Glut")
		if fehler != nil || vorgang != "gerechnet7" || !slices.Equal(probe.karten, []int64{1}) || len(probe.erwartet) != 1 {
			t.Errorf("%q, %v, Karten %v, Erwartung %v", vorgang, fehler, probe.karten, probe.erwartet)
		}
	})

	t.Run("eine fremde Karte kommt nicht durch", func(t *testing.T) {
		t.Parallel()

		auge, probe := rechnendeAufsicht()

		if _, fehler := auge.FelderSetzen(t.Context(), 10, []int64{1, 3}, "name", "Glut"); !errors.Is(fehler, domain.ErrNotFound) || probe.karten != nil {
			t.Errorf("%v, gerechnet für %v", fehler, probe.karten)
		}
	})
}

// **Die Vorschau zeigt das Gerechnete** — sonst stünde dort „Glut“, und
// gespeichert würde „[Glut]“.
func TestDieVorschauZeigtDenGerechnetenWert(t *testing.T) {
	t.Parallel()

	auge, _ := rechnendeAufsicht()

	vorschau, fehler := auge.FelderVorschau(t.Context(), 10, []int64{1}, "name", "Glut")
	if fehler != nil || len(vorschau.Zeilen) != 1 || vorschau.Zeilen[0].Neu != "[Glut]" {
		t.Errorf("%+v, %v", vorschau, fehler)
	}
}

// KarteGerechnetAnlegen legt eine Karte mit ihren gerechneten Werten an.
func (r *rechnendeprobe) KarteGerechnetAnlegen(_ context.Context, setID int64) (int64, error) {
	r.angelegt = append(r.angelegt, setID)

	return 42, nil
}

// **Eine neue Karte geht durch den rechnenden Weg** (#2516) — wo der
// Feldschreiber rechnet; sonst wie bisher an den Lebenszyklus.
func TestDieAufsichtLegtEineGerechneteKarteAn(t *testing.T) {
	t.Parallel()

	auge, probe := rechnendeAufsicht()

	if kennung, fehler := auge.KarteAnlegen(t.Context(), 10); fehler != nil || kennung != 42 || !slices.Equal(probe.angelegt, []int64{10}) {
		t.Errorf("%d, %v, angelegt %v", kennung, fehler, probe.angelegt)
	}
}
