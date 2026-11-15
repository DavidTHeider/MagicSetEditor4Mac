package domain_test

// **Welche Bindungen einen Lauscher bekamen** — Befund 3 des Zweitblicks über
// Phase G (#2272). Die angesagten Adressen stehen in derselben Reihenfolge;
// wer sie paart, paart sie mit dieser Liste und nicht mit der gewählten.

import (
	"reflect"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestGelungeneBindungen(t *testing.T) {
	t.Parallel()

	schleife := domain.Bindung{Port: 8080}
	netz := domain.Bindung{Adresse: "192.0.2.10", Port: 8080}
	zweites := domain.Bindung{Adresse: "198.51.100.7", Port: 8080}

	faelle := []struct {
		name      string
		bindungen []domain.Bindung
		fehlende  []domain.Lauschbefund
		erwartet  []domain.Bindung
	}{
		{"ohne Ausfall alle", []domain.Bindung{schleife, netz}, nil, []domain.Bindung{schleife, netz}},
		{
			"die erste fällt aus — die leere Adresse heißt die Schleife",
			[]domain.Bindung{schleife, netz, zweites},
			[]domain.Lauschbefund{{Adresse: domain.Schleifenadresse}},
			[]domain.Bindung{netz, zweites},
		},
		{
			"eine mittlere fällt aus",
			[]domain.Bindung{schleife, netz, zweites},
			[]domain.Lauschbefund{{Adresse: "192.0.2.10"}},
			[]domain.Bindung{schleife, zweites},
		},
		{
			"ein Befund nimmt nur eine von zweien mit derselben Adresse",
			[]domain.Bindung{netz, netz},
			[]domain.Lauschbefund{{Adresse: "192.0.2.10"}},
			[]domain.Bindung{netz},
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			if gelungen := domain.Gelungene(fall.bindungen, fall.fehlende); !reflect.DeepEqual(gelungen, fall.erwartet) {
				t.Errorf("gelungen sind %+v, erwartet %+v", gelungen, fall.erwartet)
			}
		})
	}
}
