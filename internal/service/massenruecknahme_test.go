package service_test

// **Was eine Massenänderung überschrieb, sagt ihr Bericht** — DOM-03.AK-03
// (#1698).
//
// # Warum
//
// Ein Lagerort für fünfzig Zeilen überschreibt fünfzig Lagerorte, und die
// Sammlung hat keinen Verlauf wie ein Set. **Eine Rücknahme braucht die alten
// Werte und die neue Fassung** — und beides kennt nur der Schreibvorgang
// selbst. Deshalb steht es im Bericht: wer zurücknimmt, schreibt die alten
// Werte gegen die neue Fassung, und hat seither jemand anderes geschrieben,
// wird das zum Befund statt zum Überschreiben (§7, Regel 2).

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestDerBerichtNenntDenStandVorher(t *testing.T) {
	t.Parallel()

	ablage, zeilen := dreiZeilen(t)
	ordner := "Ordner Rot"

	// **Eine Zeile mit veralteter Fassung** kommt nicht durch — und steht
	// deshalb auch nicht unter »vorher«.
	zeilen[2].Fassung = 9

	bericht, fehler := service.NeueMassenpflege(ablage).Aendern(t.Context(),
		service.Massenaenderung{Zeilen: zeilen, Lagerort: &ordner})
	if fehler != nil {
		t.Fatalf("Aendern: %v", fehler)
	}

	if len(bericht.Vorher) != 2 {
		t.Fatalf("vorher %+v, erwartet zwei Zeilen", bericht.Vorher)
	}

	for stelle, stand := range bericht.Vorher {
		if stand.ID != zeilen[stelle].ID || stand.Fassung != 2 || stand.Lagerort != "Kiste 4" ||
			stand.Menge != 1 || stand.Wunsch {
			t.Errorf("vorher %d ist %+v", stelle, stand)
		}
	}

	if bericht.Vorher[1].Zustand != domain.ZustandPlayed {
		t.Errorf("der Zustand vorher ist %q", bericht.Vorher[1].Zustand)
	}

	// **Und die Rücknahme geht gegen die neue Fassung durch.**
	kiste := "Kiste 4"
	zurueck, fehler := service.NeueMassenpflege(ablage).Aendern(t.Context(), service.Massenaenderung{
		Zeilen: []service.Massenzeile{
			{ID: bericht.Vorher[0].ID, Fassung: bericht.Vorher[0].Fassung},
			{ID: bericht.Vorher[1].ID, Fassung: bericht.Vorher[1].Fassung},
		},
		Lagerort: &kiste,
	})
	if fehler != nil || zurueck.Geaendert != 2 {
		t.Errorf("Rücknahme %+v (%v)", zurueck, fehler)
	}
}
