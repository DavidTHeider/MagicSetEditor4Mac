package service_test

// **Die Kontoschau** — Werkstück `T11` von Phase 5.6
// (#1476).
//
// # Was hier zugesagt wird
//
//   - **Ausgeschaltet heißt ausgeschaltet**, und zwar bevor irgendetwas
//     hinausgeht: die Quelle wird gar nicht gefragt.
//   - **Der Riegel greift auch bei eingeschaltetem Schalter** — ein Leser
//     schickt nichts hinaus.
//   - Der geglückte Weg gibt Konto, Angebote und Bestellungen in einem Zug.
//   - **Und jeder Fehler der Quelle kommt an** — eine halbe Übersicht sähe
//     aus wie ein leeres Konto.

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// kontoquelle antwortet auf die drei Rufe und zählt sie.
type kontoquelle struct {
	rufe int

	kontofehler, angebotsfehler, bestellfehler error
}

func (k *kontoquelle) Konto(context.Context) (domain.Marktkonto, error) {
	k.rufe++

	return domain.Marktkonto{AppID: 3, Name: "Probe", NutzerID: 7},
		k.kontofehler
}

func (k *kontoquelle) EigeneAngebote(context.Context, int64) (
	[]domain.Eigenangebot, error,
) {
	k.rufe++

	return []domain.Eigenangebot{{ID: 1, Name: "Tranquil Cove", Menge: 2}},
		k.angebotsfehler
}

func (k *kontoquelle) Bestellungen(context.Context, int, int) (
	[]domain.Marktbestellung, []domain.Befund, error,
) {
	k.rufe++

	return []domain.Marktbestellung{{
			ID: 9, Zustand: domain.BestellungHubOffen, UeberNullkonto: true,
		}}, []domain.Befund{{Ort: "Bestellung 9", Meldung: "unbekannt"}},
		k.bestellfehler
}

// schalterprobe gibt eine feste Stellung.
type schalterprobe struct {
	stellung string
	fehler   error
}

func (s schalterprobe) Wert(context.Context, string) (string, error) {
	return s.stellung, s.fehler
}

// **Ausgeschaltet fragt niemanden.**
func TestEineAusgeschalteteKontoschauFragtNiemanden(t *testing.T) {
	t.Parallel()

	quelle := &kontoquelle{}

	_, fehler := service.NeueKontoschau(quelle,
		schalterprobe{stellung: domain.KontoschauAus}, nil).
		Lesen(t.Context(), 0, 0)
	if !errors.Is(fehler, domain.ErrErlaubnis) {
		t.Fatalf("der Fehler ist keine Erlaubnisfrage: %v", fehler)
	}

	if quelle.rufe != 0 {
		t.Errorf("die Quelle wurde %dmal gefragt", quelle.rufe)
	}
}

// **Und ein Schalter, der klagt, ebenso.**
func TestEinKlagenderSchalterFragtNiemanden(t *testing.T) {
	t.Parallel()

	quelle := &kontoquelle{}

	if _, fehler := service.NeueKontoschau(quelle,
		schalterprobe{fehler: errKaputt}, nil).
		Lesen(t.Context(), 0, 0); fehler == nil {
		t.Fatal("die Klage wurde verschluckt")
	}

	if quelle.rufe != 0 {
		t.Errorf("die Quelle wurde %dmal gefragt", quelle.rufe)
	}
}

// **Der Riegel greift auch bei eingeschaltetem Schalter.**
//
// **Die Gegenprobe zum Lauf darüber**: ein Schalter allein ist keine
// Autorisierung — ein Leser darf nichts hinausschicken, und das entscheidet
// nicht die Einstellung, sondern der Akteur (0078).
func TestDerRiegelGreiftAuchBeiEingeschalteterKontoschau(t *testing.T) {
	t.Parallel()

	quelle := &kontoquelle{}

	_, fehler := service.NeueKontoschau(quelle,
		schalterprobe{stellung: domain.KontoschauAn},
		verweigerndeWache{}).Lesen(t.Context(), 0, 0)
	if !errors.Is(fehler, domain.ErrErlaubnis) {
		t.Fatalf("der Fehler ist keine Erlaubnisfrage: %v", fehler)
	}

	if quelle.rufe != 0 {
		t.Errorf("die Quelle wurde %dmal gefragt", quelle.rufe)
	}
}

// verweigerndeWache sagt zu jeder Wirkung nein.
type verweigerndeWache struct{}

func (verweigerndeWache) Pruefe(context.Context, domain.Wirkung) error {
	return domain.MitGrund(domain.ErrErlaubnis, nil, "ein Leser schickt nichts")
}

// **Der geglückte Weg gibt alles drei — samt Befunden.**
func TestDieKontoschauGibtKontoAngeboteUndBestellungen(t *testing.T) {
	t.Parallel()

	stand, fehler := service.NeueKontoschau(&kontoquelle{},
		schalterprobe{stellung: domain.KontoschauAn}, nil).
		Lesen(t.Context(), 1, 20)
	if fehler != nil {
		t.Fatalf("Lesen: %v", fehler)
	}

	if stand.Konto.NutzerID != 7 || len(stand.Angebote) != 1 ||
		len(stand.Bestellungen) != 1 {
		t.Errorf("der Stand kam als %+v an", stand)
	}

	// **Die Befunde kommen mit** — eine Bestellung, die dieses Programm nicht
	// deutet, soll nicht still durchgehen (§8).
	if len(stand.Befunde) != 1 {
		t.Errorf("die Befunde kamen als %+v an", stand.Befunde)
	}
}

// **Jeder Fehler der Quelle kommt an.**
//
// **Und keine halbe Übersicht**: ein Konto ohne seine Bestellungen sähe aus
// wie ein Konto ohne Bestellungen.
func TestJederFehlerDerKontoquelleKommtAn(t *testing.T) {
	t.Parallel()

	for name, quelle := range map[string]*kontoquelle{
		"das Konto":        {kontofehler: errKaputt},
		"die Angebote":     {angebotsfehler: errKaputt},
		"die Bestellungen": {bestellfehler: errKaputt},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			stand, fehler := service.NeueKontoschau(quelle,
				schalterprobe{stellung: domain.KontoschauAn}, nil).
				Lesen(t.Context(), 0, 0)
			if fehler == nil {
				t.Fatal("der Fehler wurde verschluckt")
			}

			if stand.Konto.NutzerID != 0 || len(stand.Angebote) != 0 {
				t.Errorf("es kam ein halber Stand: %+v", stand)
			}
		})
	}
}
