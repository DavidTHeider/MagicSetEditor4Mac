package service_test

// **Was eine empfangene Revision behaupten darf** — TRADE-03.AK-01 und
// NAT-01 (#1717):
// an der Post, nicht nur in der Domäne — und es wird dann **nichts**
// geschrieben.

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestEineEmpfangeneRevisionBehauptetNichtsFremdes(t *testing.T) {
	t.Parallel()

	angenommen := einNachfolger()
	angenommen.Zustand = domain.HandelAngenommen
	angenommen.Zustimmungen = []string{"bich", dasGegenueber}

	verbucht := einErstesVomGegenueber()
	verbucht.Zustand = domain.HandelLokalVerbucht

	for _, fall := range []struct {
		name    string
		kennt   bool
		angebot domain.Handelsangebot
	}{
		{"ein Vorschlag, der schon angenommen sein will", true, angenommen},
		{"eine erste Revision, die schon verbucht sein will", false, verbucht},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			briefkasten := &briefkastenprobe{}
			stand := standprobe{stand: derHeutigeStand()}

			if !fall.kennt {
				stand = standprobe{klagt: domain.MitGrund(domain.ErrNotFound, nil, "gibt es nicht")}
			}

			post := service.NeueHandelspost(briefkasten, stand, kontaktnachschlagprobe{kontakt: einHandelnder()})

			if _, fehler := post.Empfangen(context.Background(), dasGegenueber,
				einePost(fall.angebot, nil)); !errors.Is(fehler, domain.ErrValidation) {
				t.Errorf("Fehler ist %v", fehler)
			}

			if len(briefkasten.angebote) != 0 {
				t.Errorf("geschrieben wurden %d Angebote", len(briefkasten.angebote))
			}
		})
	}
}
