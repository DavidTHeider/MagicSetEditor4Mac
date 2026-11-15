package service_test

// **Die Fehlerwege der Handelspost** — Werkstück `T5` von Phase 5.6
// (#1470).
//
// **Eine eigene Datei**, weil `handelspost_test.go` sonst über die
// Zeilengrenze aus §2 ginge. Der Schnitt folgt der Frage: dort, was geschieht;
// hier, was nicht durchkommt.
//
// # Warum jeder einzeln geprüft wird
//
// **Weil vier Wege scheitern können und jeder für sich.** Ein Doppel, das
// überall klagt, misst nur den ersten — und der Rest stünde als ungeprüfte
// Zusage da (§10).

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestJederWegDerPostKannScheitern(t *testing.T) {
	t.Parallel()

	gestolpert := errors.New("der Bestand ist zu")

	for name, stelle := range map[string]func() *service.Handelspost{
		"der Briefkasten beim Nachschlagen": func() *service.Handelspost {
			return service.NeueHandelspost(
				&briefkastenprobe{klagt: gestolpert},
				standprobe{stand: derHeutigeStand()},
				kontaktnachschlagprobe{kontakt: einHandelnder()})
		},

		"der Kontakt": func() *service.Handelspost {
			return service.NeueHandelspost(&briefkastenprobe{},
				standprobe{stand: derHeutigeStand()},
				kontaktnachschlagprobe{klagt: gestolpert})
		},

		"der Stand": func() *service.Handelspost {
			return service.NeueHandelspost(&briefkastenprobe{},
				standprobe{klagt: gestolpert},
				kontaktnachschlagprobe{kontakt: einHandelnder()})
		},

		"das Zustellen": func() *service.Handelspost {
			return service.NeueHandelspost(&nurZustellenKlagt{klagt: gestolpert},
				standprobe{stand: derHeutigeStand()},
				kontaktnachschlagprobe{kontakt: einHandelnder()})
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, fehler := stelle().Empfangen(context.Background(),
				dasGegenueber, einePost(einNachfolger(), nil)); fehler == nil {
				t.Error("es kam kein Fehler")
			}
		})
	}
}

// nurZustellenKlagt lässt das Nachschlagen durch und klagt beim Schreiben.
//
// **Getrennt, weil sonst schon das Nachschlagen scheiterte** — und der Weg
// dahinter bliebe unbetreten.
type nurZustellenKlagt struct{ klagt error }

func (n *nurZustellenKlagt) Nachschlagen(
	context.Context, string,
) (*domain.Inboxeintrag, error) {
	return nil, nil //nolint:nilnil // „nichts" ist hier kein Fehler
}

func (n *nurZustellenKlagt) Zustellen(
	context.Context, domain.Inboxeintrag, domain.Handelsangebot, domain.Handelsangebot,
	domain.Zusagenwirkungsart,
) error {
	return n.klagt
}

// **Geschrieben wird gegen den gelesenen Stand** — Befund Z1 des Zweitblicks
// über Phase 6.8 (#1669): die Post reicht ihn an den Briefkasten, und ein
// überholter Stand kommt als Fehler zurück, nicht als „bekannt".
func TestDiePostSchreibtGegenDenGelesenenStand(t *testing.T) {
	t.Parallel()

	briefkasten := &briefkastenprobe{}
	post := service.NeueHandelspost(briefkasten,
		standprobe{stand: derHeutigeStand()},
		kontaktnachschlagprobe{kontakt: einHandelnder()})

	if _, fehler := post.Empfangen(context.Background(), dasGegenueber,
		einePost(einNachfolger(), nil)); fehler != nil {
		t.Fatalf("Empfangen: %v", fehler)
	}

	if briefkasten.vorher.Revision != derHeutigeStand().Revision ||
		briefkasten.vorher.Abdruck() != derHeutigeStand().Abdruck() {
		t.Errorf("geschrieben gegen %+v", briefkasten.vorher)
	}

	ueberholt := service.NeueHandelspost(&nurZustellenKlagt{klagt: domain.ErrUeberholt},
		standprobe{stand: derHeutigeStand()},
		kontaktnachschlagprobe{kontakt: einHandelnder()})

	if _, fehler := ueberholt.Empfangen(context.Background(), dasGegenueber,
		einePost(einNachfolger(), nil)); !errors.Is(fehler, domain.ErrUeberholt) {
		t.Errorf("ein überholter Stand gab %v", fehler)
	}
}
