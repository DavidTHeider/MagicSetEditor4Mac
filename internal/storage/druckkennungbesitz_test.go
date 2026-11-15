package storage_test

// **Ein Konflikt um eine Druckkennung nennt den Druck, dem sie gehört** — B05
// (#2224):
// ohne die Prüfung in `kennungAnhaengen` griffe der eindeutige Index und
// meldete einen Konflikt ohne Namen. Die Zufallsmutation ließ sie weg, und
// der Test darüber (`TestEineFremdeDruckkennungIstEinKonflikt`) blieb grün —
// er fragte nur die Art des Fehlers.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestEinKonfliktUmEineDruckkennungNenntIhrenDruck(t *testing.T) {
	t.Parallel()

	ablage := storage.NeueDrucke(frischeDatenbank(t))
	kennung := scryfallkennung("e3285e6b-3e79-4d7c-bf96-d920f973b122")
	alpha := &domain.Druck{Name: "Bolt", Setcode: "lea", Kennungen: []domain.Druckkennung{kennung}}
	beta := &domain.Druck{Name: "Bolt", Setcode: "leb", Kennungen: []domain.Druckkennung{{System: domain.SystemMTGJSON, Semantik: domain.SemantikDruck, Wert: "5f8287b1-5bb6-5f4c-ad17-316a40d5bb0c"}}}

	for _, druck := range []*domain.Druck{alpha, beta} {
		if err := ablage.Sichern(probekontext(t), druck); err != nil {
			t.Fatalf("sichern: %v", err)
		}
	}

	beta.Kennungen = append(beta.Kennungen, kennung)

	err := ablage.Sichern(probekontext(t), beta)
	if err == nil || !strings.Contains(err.Error(), "gehört schon dem Druck") {
		t.Errorf("der Konflikt nennt den Druck nicht, dem die Kennung gehört: %v", err)
	}
}
