package domain_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Eine Behaltemenge von null heißt die Vorgabe eins** — UXR-F18
// (#2033).
//
// Bei „Dubletten" zeigt eine Sicht, was über der Behaltemenge liegt. Mit null
// ist jedes Stück eine Dublette, und „Dubletten" gäbe die ganze Sammlung frei
// — wer das will, nimmt den Umfang dafür. Bis hierhin schickte das Fenster
// null, die Kommandozeile ohne Angabe null und der Dienst ohne Feld null;
// `mse dubletten` und die Ordnung der Sammlung nahmen eins. **Die Vorgabe
// steht jetzt an einer Stelle**, und jede Fläche gibt für dieselbe Eingabe
// dieselbe Antwort (§16).
func TestEineBehaltemengeVonNullHeisstEins(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name     string
		freigabe domain.Sammlungsfreigabe
		erwartet int
	}{
		{
			"Dubletten ohne Angabe behalten eins",
			domain.Sammlungsfreigabe{Umfang: domain.UmfangDubletten},
			1,
		},
		{
			"Dubletten mit Angabe behalten sie",
			domain.Sammlungsfreigabe{Umfang: domain.UmfangDubletten, Behalten: 3},
			3,
		},
		{
			"ein anderer Umfang bekommt keine Behaltemenge",
			domain.Sammlungsfreigabe{Umfang: domain.UmfangSammlung},
			0,
		},
		{
			"eine negative Angabe bleibt, damit die Prüfung sie anstreicht",
			domain.Sammlungsfreigabe{Umfang: domain.UmfangDubletten, Behalten: -1},
			-1,
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			if behalten := fall.freigabe.MitVorgaben().Behalten; behalten != fall.erwartet {
				t.Errorf("Behalten %d, erwartet %d", behalten, fall.erwartet)
			}
		})
	}
}
