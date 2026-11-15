package service_test

// **Derselbe Druck zweimal nennt die Zeile, die es schon gibt** — DOM-03,
// S04.NAT-01 und 03 (#1698).
//
// Der Bestand weist die zweite Zeile ab; bis hierher kam „diesen Druck führt
// die Sammlung schon" an, und wer die Menge erhöhen wollte, musste die Zeile
// suchen. Jetzt trägt die Absage sie mit — samt Fassung, damit das Erhöhen
// gegen den gelesenen Stand geht (§7, Regel 2).

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// eindeutigeAblage weist eine zweite gleiche Zeile ab, wie die Eindeutigkeit
// im Bestand — und kann beim Nachschlagen scheitern.
type eindeutigeAblage struct {
	*sammlungsdoppel

	nachschlagefehler error
	// vergessen lässt die Absage kommen, ohne dass die Zeile noch dasteht —
	// ein Löschen dazwischen.
	vergessen bool
}

func (a *eindeutigeAblage) Anlegen(ctx context.Context, exemplar *domain.Exemplar) (int64, error) {
	for _, zeile := range a.zeilen {
		if zeile.GleicheZeile(*exemplar) {
			if a.vergessen {
				a.zeilen = nil
			}

			return 0, domain.NewFieldError("", domain.SchluesselDoppeltesExemplar, "doppelt")
		}
	}

	return a.sammlungsdoppel.Anlegen(ctx, exemplar)
}

func (a *eindeutigeAblage) Fuer(ctx context.Context, printing string) ([]*domain.Exemplar, error) {
	if a.nachschlagefehler != nil {
		return nil, a.nachschlagefehler
	}

	return a.sammlungsdoppel.Fuer(ctx, printing)
}

func (a *eindeutigeAblage) NachDruck(ctx context.Context, druckID int64) ([]*domain.Exemplar, error) {
	if a.nachschlagefehler != nil {
		return nil, a.nachschlagefehler
	}

	return a.sammlungsdoppel.NachDruck(ctx, druckID)
}

func TestEinZweiterGleicherDruckNenntDieVorhandeneZeile(t *testing.T) {
	t.Parallel()

	for name, exemplar := range map[string]domain.Exemplar{
		"mit Alias": {Printing: "druck-1", Menge: 2, Sprache: "en"},
		"vorläufig": {DruckID: 7, Menge: 2, Sprache: "en"},
	} {
		ablage := &eindeutigeAblage{sammlungsdoppel: &sammlungsdoppel{}}
		pflege := service.NeueSammlungspflege(ablage)

		erste, fehler := pflege.Anlegen(t.Context(), exemplar)
		if fehler != nil {
			t.Fatal(fehler)
		}

		// **Auch ohne Angaben, die die Pflege erst setzt** — Zustand und
		// Sprache kommen geglättet an, wie beim ersten Mal.
		noch := exemplar
		noch.Menge = 1
		noch.Sprache = " en "

		_, fehler = pflege.Anlegen(t.Context(), noch)

		var dublette *domain.Dublette
		if !errors.As(fehler, &dublette) {
			t.Fatalf("%s: keine Dublette: %v", name, fehler)
		}

		if dublette.Vorhanden.ID != erste || dublette.Vorhanden.Menge != 2 || dublette.Vorhanden.Fassung != 1 {
			t.Errorf("%s: die Zeile %+v", name, dublette.Vorhanden)
		}

		if !errors.Is(fehler, domain.ErrDoppelt) {
			t.Errorf("%s: die Art ging verloren", name)
		}
	}
}

func TestOhneZeileBleibtEsBeiDerAbsage(t *testing.T) {
	t.Parallel()

	nachschlagen := errors.New("die Platte klemmt")

	for name, ablage := range map[string]*eindeutigeAblage{
		"das Nachschlagen scheitert": {sammlungsdoppel: &sammlungsdoppel{}, nachschlagefehler: nachschlagen},
		"die Zeile ist weg":          {sammlungsdoppel: &sammlungsdoppel{}, vergessen: true},
	} {
		pflege := service.NeueSammlungspflege(ablage)
		exemplar := domain.Exemplar{Printing: "druck-1", Menge: 1}

		if _, fehler := pflege.Anlegen(t.Context(), exemplar); fehler != nil {
			t.Fatal(fehler)
		}

		_, fehler := pflege.Anlegen(t.Context(), exemplar)

		var dublette *domain.Dublette
		if errors.As(fehler, &dublette) || !errors.Is(fehler, domain.ErrDoppelt) {
			t.Errorf("%s: %v", name, fehler)
		}

		// **Der Lesefehler verdrängt die Absage nicht** — sie sagt, was zu
		// tun ist; er gehört ins Protokoll.
		if errors.Is(fehler, nachschlagen) {
			t.Errorf("%s: der Lesefehler steht an der Stelle der Absage", name)
		}
	}
}

func TestEinAndererFehlerBleibtEinAndererFehler(t *testing.T) {
	t.Parallel()

	ablage := &eindeutigeAblage{sammlungsdoppel: &sammlungsdoppel{fehler: domain.ErrNotFound}}

	_, fehler := service.NeueSammlungspflege(ablage).Anlegen(t.Context(), domain.Exemplar{Printing: "druck-1", Menge: 1})
	if !errors.Is(fehler, domain.ErrNotFound) || errors.Is(fehler, domain.ErrDoppelt) {
		t.Errorf("%v", fehler)
	}
}
