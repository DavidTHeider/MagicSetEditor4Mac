package service_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Bindet kein Sitz ein Profil, sitzt das gezeigte am Sitz des Wirts** —
// UXR-F21 (#2041),
// Vermerk 0164.
func TestDerWirtSitztImGezeigtenProfil(t *testing.T) {
	t.Parallel()

	zwei := func() []service.Sitzwunsch {
		return []service.Sitzwunsch{{Kennung: "ada"}, {Kennung: "bo"}}
	}

	faelle := []struct {
		name     string
		wuensche []service.Sitzwunsch
		wirt     string
		gezeigt  int64
		erwartet []int64
	}{
		{"ohne Bindung sitzt das gezeigte Profil am Wirt", zwei(), "bo", 7, []int64{0, 7}},
		{"der Wirt am ersten Sitz", zwei(), "ada", 7, []int64{7, 0}},
		{
			"eine gewählte Bindung bleibt, wie sie ist",
			[]service.Sitzwunsch{{Kennung: "ada"}, {Kennung: "bo", ProfilID: 3}},
			"ada", 7,
			[]int64{0, 3},
		},
		{"ohne gezeigtes Profil bindet nichts", zwei(), "ada", 0, []int64{0, 0}},
		{"ein Wirt, der nicht am Tisch sitzt, bindet nichts", zwei(), "cem", 7, []int64{0, 0}},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			gebunden := service.WirtImGezeigtenProfil(fall.wuensche, fall.wirt, fall.gezeigt)

			gesehen := make([]int64, 0, len(gebunden))
			for _, wunsch := range gebunden {
				gesehen = append(gesehen, wunsch.ProfilID)
			}

			if len(gesehen) != len(fall.erwartet) {
				t.Fatalf("%v, erwartet %v", gesehen, fall.erwartet)
			}

			for stelle := range gesehen {
				if gesehen[stelle] != fall.erwartet[stelle] {
					t.Errorf("%v, erwartet %v", gesehen, fall.erwartet)
				}
			}
		})
	}
}

// **Die Wünsche des Aufrufers bleiben unberührt** — die Vorgabe gibt eine
// Abschrift, sonst sähe ein zweiter Bau mit denselben Wünschen eine Bindung,
// die niemand gewählt hat.
func TestDieVorgabeSchreibtNichtInDieWuensche(t *testing.T) {
	t.Parallel()

	wuensche := []service.Sitzwunsch{{Kennung: "ada"}, {Kennung: "bo"}}
	_ = service.WirtImGezeigtenProfil(wuensche, "ada", 7)

	if wuensche[0].ProfilID != 0 {
		t.Errorf("die Wünsche tragen jetzt %d", wuensche[0].ProfilID)
	}
}
