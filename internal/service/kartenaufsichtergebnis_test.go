package service_test

// **Die Aufsicht reicht weiter, was gespeichert ist** — #2535. Das Fenster
// bestätigt seinen Entwurf mit genau diesem Wert; die Setgrenze gilt dabei wie
// für jedes Setzen (#808).

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// ergebnisprobe sagt zusätzlich, was sie gespeichert hätte.
type ergebnisprobe struct{ *feldprobe }

func (e ergebnisprobe) FeldSetzenMitErgebnis(
	_ context.Context, karteID int64, _, wert string,
) (gespeichert string, bekannt bool, fehler error) {
	e.gesetzt = append(e.gesetzt, karteID)

	return "[" + wert + "]", true, nil
}

func TestDieAufsichtReichtDasGespeicherteWeiter(t *testing.T) {
	t.Parallel()

	sicht := &kartensicht{karten: map[int64]int64{1: 10, 3: 20}}

	t.Run("ein Schreiber, der es sagt", func(t *testing.T) {
		t.Parallel()

		felder := ergebnisprobe{&feldprobe{}}
		auge := service.NeueKartenaufsicht(sicht, &lebenszyklusprotokoll{}).MitFeldweg(felder, felder)

		gespeichert, bekannt, fehler := auge.FeldSetzenMitErgebnis(t.Context(), 10, 1, "name", "Glut")
		if fehler != nil || gespeichert != "[Glut]" || !bekannt {
			t.Errorf("%q, %v, %v — erwartet [Glut]", gespeichert, bekannt, fehler)
		}
	})

	// **Einer, der es nicht sagt, schreibt trotzdem** — und das Gespeicherte
	// bleibt offen, statt den übergebenen Wert zu behaupten.
	t.Run("ein Schreiber, der es nicht sagt", func(t *testing.T) {
		t.Parallel()

		felder := &feldprobe{}
		auge := service.NeueKartenaufsicht(sicht, &lebenszyklusprotokoll{}).MitFeldweg(felder, felder)

		gespeichert, bekannt, fehler := auge.FeldSetzenMitErgebnis(t.Context(), 10, 1, "name", "Glut")
		if fehler != nil || gespeichert != "" || bekannt || len(felder.gesetzt) != 1 {
			t.Errorf("%q, %v, %v, geschrieben %v", gespeichert, bekannt, fehler, felder.gesetzt)
		}
	})

	t.Run("eine fremde Karte kommt nicht durch", func(t *testing.T) {
		t.Parallel()

		felder := ergebnisprobe{&feldprobe{}}
		auge := service.NeueKartenaufsicht(sicht, &lebenszyklusprotokoll{}).MitFeldweg(felder, felder)

		_, _, fehler := auge.FeldSetzenMitErgebnis(t.Context(), 10, 3, "name", "Glut")
		if !errors.Is(fehler, domain.ErrNotFound) || len(felder.gesetzt) != 0 {
			t.Errorf("%v, geschrieben %v — erwartet ErrNotFound und nichts", fehler, felder.gesetzt)
		}
	})
}
