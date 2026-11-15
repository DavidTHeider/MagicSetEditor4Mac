package service_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// TestEineVolleSpurKopiertNicht hält Zweitblick über Phase 6.4, Befund 15:
// der Kopf der Spur versprach einen Ring, und eine volle Spur kopierte bei
// **jedem** aufgenommenen Ruf alle Einträge — bei 10.000 unter der Sperre,
// die jeder Ruf jeder Fläche nimmt.
//
// **Gemessen an der Zuteilung und nicht an der Zeit**: ein Ruf durch eine
// volle Spur teilt nicht mehr zu als einer ohne Spur.
func TestEineVolleSpurKopiertNicht(t *testing.T) {
	service.LaufzeitwerkZuruecksetzen()
	t.Cleanup(service.LaufzeitwerkZuruecksetzen)

	ctx := service.MitVorgang(t.Context(), "0a1b2c3d")
	erfasse := func() { service.ErfasseRuf(ctx, ruf("GET /sets")) }

	ohne := testing.AllocsPerRun(50, erfasse)

	service.SpurEinrichten(domain.MessregelAus(map[string]string{
		domain.MessungSpur: "an", domain.MessungSpurgrenze: "10",
	}))

	for range 20 {
		erfasse()
	}

	if voll := testing.AllocsPerRun(50, erfasse); voll > ohne {
		t.Errorf("ein Ruf durch die volle Spur teilt %.0f zu, ohne Spur %.0f", voll, ohne)
	}

	// **Und der Ring hält, was er soll**: die zehn jüngsten, der Reihe nach.
	stand := service.Spurstand()
	if len(stand.Punkte) != 10 || stand.Verdraengt != stand.Aufgenommen-10 {
		t.Errorf("%d Punkte, %d verdrängt von %d", len(stand.Punkte), stand.Verdraengt, stand.Aufgenommen)
	}

	for i := 1; i < len(stand.Punkte); i++ {
		if stand.Punkte[i].Zeit.Before(stand.Punkte[i-1].Zeit) {
			t.Errorf("Punkt %d ist älter als sein Vorgänger", i)
		}
	}
}
