package werkzeuge

// **Die Paketarten stehen in der Domäne wie in `msepaket`** — UX-13.AK-01
// (#1685). `domain` darf das Blatt nicht kennen (§2), also führt es die
// sieben ein zweites Mal; dieser Wächter hält beide Listen gleich.

import (
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

func TestDiePaketartenStehenInDerDomaeneWieImBlatt(t *testing.T) {
	t.Parallel()

	imBlatt := make([]string, 0, len(msepaket.Arten()))
	for _, art := range msepaket.Arten() {
		imBlatt = append(imBlatt, string(art))
	}

	inDerDomaene := domain.Paketendungen()
	slices.Sort(imBlatt)
	slices.Sort(inDerDomaene)

	if !slices.Equal(imBlatt, inDerDomaene) {
		t.Errorf("msepaket %v, domain %v", imBlatt, inDerDomaene)
	}
}
