package service

// **Welcher Zug sich zurücknehmen lässt — und mit welchem Grund nicht** —
// P8 (#2376), Vermerk 0197. Je Liste ein Fall, dazu die Umkehr der vier, die
// darauf stehen.

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func eines(art string, oeffentlich map[string]any) domain.Ereignis {
	if oeffentlich == nil {
		oeffentlich = map[string]any{}
	}

	if _, da := oeffentlich["sitz"]; !da {
		oeffentlich["sitz"] = "s1"
	}

	return domain.Ereignis{Folge: 7, Art: art, Oeffentlich: oeffentlich}
}

func TestUmkehrKehrtDieListeUm(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		ereignis domain.Ereignis
		art      string
		nutzlast map[string]any
	}{
		"Tappen wird Enttappen": {
			eines(BefehlTappen, map[string]any{"objekt": "o1"}),
			BefehlEnttappen,
			map[string]any{"objekt": "o1"},
		},
		"Enttappen wird Tappen": {
			eines(BefehlEnttappen, map[string]any{"objekt": "o1"}),
			BefehlTappen,
			map[string]any{"objekt": "o1"},
		},
		"Leben mit dem Gegenwert": {
			eines(BefehlLeben, map[string]any{"um": float64(-3)}),
			BefehlLeben,
			map[string]any{"um": float64(3)},
		},
		"Vergiftung mit dem Gegenwert": {
			eines(BefehlVergiftung, map[string]any{"um": float64(2)}),
			BefehlVergiftung,
			map[string]any{"um": float64(-2)},
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			umkehr, folge, fehler := umkehrVon([]domain.Ereignis{fall.ereignis}, "s1")
			if fehler != nil || folge != 7 || umkehr.Art != fall.art {
				t.Fatalf("es kam %+v, %d, %v", umkehr, folge, fehler)
			}

			for schluessel, wert := range fall.nutzlast {
				if umkehr.Nutzlast[schluessel] != wert {
					t.Errorf("%s ist %v statt %v", schluessel, umkehr.Nutzlast[schluessel], wert)
				}
			}
		})
	}
}

func TestUmkehrWeistMitGrundAb(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		letzte []domain.Ereignis
		art    error
		grund  string
	}{
		"nichts geschehen":    {nil, domain.ErrValidation, "noch nichts geschehen"},
		"fremder Folgezug":    {[]domain.Ereignis{eines(BefehlTappen, map[string]any{"sitz": "s2"})}, domain.ErrConflict, "anderer Sitz"},
		"eine Rücknahme":      {[]domain.Ereignis{eines(BefehlZuruecknehmen, nil)}, domain.ErrValidation, "Rücknahme nimmt man nicht"},
		"mehrere Folgen":      {[]domain.Ereignis{eines(BefehlLeben, nil), eines("partie.ende", nil)}, domain.ErrValidation, "mehr als eine Folge"},
		"Zufall":              {[]domain.Ereignis{eines(BefehlMuenzwurf, nil)}, domain.ErrValidation, "Zufall"},
		"Aufgedecktes":        {[]domain.Ereignis{eines(BefehlUmdrehen, nil)}, domain.ErrValidation, "aufgedeckt"},
		"private Ergänzung":   {[]domain.Ereignis{{Art: BefehlTappen, Oeffentlich: map[string]any{"sitz": "s1"}, Privat: []domain.Ergaenzung{{Empfaenger: "s1"}}}}, domain.ErrValidation, "aufgedeckt"},
		"Abschluss":           {[]domain.Ereignis{eines(BefehlPassen, nil)}, domain.ErrValidation, "Wort weitergegeben"},
		"nicht auf der Liste": {[]domain.Ereignis{eines(BefehlZaehler, nil)}, domain.ErrValidation, "nicht auf der Liste"},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, _, fehler := umkehrVon(fall.letzte, "s1")
			if !errors.Is(fehler, fall.art) || !strings.Contains(fehler.Error(), fall.grund) {
				t.Errorf("es kam %v", fehler)
			}
		})
	}
}
