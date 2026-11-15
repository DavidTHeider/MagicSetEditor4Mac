package app_test

// **Die Naht zwischen Dienst und Fläche** — Werkstück C von Phase 4.11
// (#1338).
//
// Geprüft wird die Übersetzung, die zwischen beiden steht: aus Angaben werden
// DTOs, aus Befunden Texte, und aus einem `nil`-Slice eine leere Liste.

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// zaehlauskunft gibt ein Spiel mit einer Dimension und ein paar Karten.
type zaehlauskunft struct {
	angaben []service.Auszaehlangabe
	karten  []*domain.CardRow
	fehler  error
}

func (z *zaehlauskunft) Auszaehldimensionen(_ context.Context, _ int64) (
	[]service.Auszaehlangabe, service.Vorlage, error,
) {
	return z.angaben, service.Vorlage{
		Kartenfelder: []domain.Field{{Name: "farbe", Kind: domain.KindText}},
	}, z.fehler
}

func (z *zaehlauskunft) VorlageFuer(
	_ context.Context, _ *domain.CardRow,
) service.Vorlage {
	return service.Vorlage{
		Kartenfelder: []domain.Field{{Name: "farbe", Kind: domain.KindText}},
	}
}

func (z *zaehlauskunft) Karten(_ context.Context, _ int64) ([]*domain.CardRow, error) {
	return z.karten, z.fehler
}

func zaehlnaht(auskunft *zaehlauskunft) *app.Auszaehlnaht {
	return app.NeueAuszaehlnaht(service.NeuerAuszaehldienst(auskunft))
}

func probezaehlung() *zaehlauskunft {
	return &zaehlauskunft{
		angaben: []service.Auszaehlangabe{{
			Name: "farbe", Beschreibung: "je Farbe", Skript: "card.farbe",
			Gruppen: []string{"rot", "blau"},
			Farben:  map[string]string{"rot": "rgb(156,15,15)"},
		}},
		karten: []*domain.CardRow{{
			ID: 1, Values: []domain.Value{{Field: "farbe", Text: "rot"}},
		}},
	}
}

// **Aus Angaben werden DTOs.**
func TestDieAuszaehlnahtGibtDasAngebotAlsDTO(t *testing.T) {
	t.Parallel()

	dimensionen, fehler := zaehlnaht(probezaehlung()).Angebot(probekontext(t), 1)
	if fehler != nil {
		t.Fatalf("Angebot: %v", fehler)
	}

	if len(dimensionen) != 1 || dimensionen[0].Name != "farbe" ||
		dimensionen[0].Beschreibung != "je Farbe" {
		t.Errorf("das Angebot lautet %+v", dimensionen)
	}
}

// **Ein leeres Angebot ist eine leere Liste, kein `nil`.**
func TestDieAuszaehlnahtGibtEineLeereListe(t *testing.T) {
	t.Parallel()

	dimensionen, fehler := zaehlnaht(&zaehlauskunft{}).Angebot(probekontext(t), 1)
	if fehler != nil || dimensionen == nil {
		t.Errorf("Angebot gab %+v / %v", dimensionen, fehler)
	}
}

// **Eine Auszählung kommt samt Farben und leeren Gruppen durch.**
func TestDieAuszaehlnahtGibtDieGruppenAlsDTO(t *testing.T) {
	t.Parallel()

	gezaehlt, fehler := zaehlnaht(probezaehlung()).Auszaehlen(probekontext(t), 1, "farbe")
	if fehler != nil {
		t.Fatalf("Auszaehlen: %v", fehler)
	}

	if gezaehlt.Dimension != "farbe" || len(gezaehlt.Gruppen) != 2 {
		t.Fatalf("die Auszählung lautet %+v", gezaehlt)
	}

	if gezaehlt.Gruppen[0].Name != "rot" || gezaehlt.Gruppen[0].Karten != 1 ||
		gezaehlt.Gruppen[0].Farbe != "rgb(156,15,15)" {
		t.Errorf("die erste Gruppe ist %+v", gezaehlt.Gruppen[0])
	}

	// **Die zweite Gruppe hat null Karten und steht trotzdem da** — die
	// erklärte Reihenfolge ist eine Aussage über das Set.
	if gezaehlt.Gruppen[1].Karten != 0 {
		t.Errorf("die zweite Gruppe ist %+v", gezaehlt.Gruppen[1])
	}

	if gezaehlt.Befunde == nil {
		t.Error("die Befunde sind nil")
	}
}

// **Jede Gruppe nennt ihre Karten — als Zeichenketten, wie jede Kennung an
// dieser Naht** (UXR-F6, #2013). Eine Gruppe ohne Karten nennt eine leere
// Liste und kein `null`: die Seite filtert damit, ohne zu prüfen, ob es sie
// gibt.
func TestDieAuszaehlnahtNenntDieKartenJeGruppe(t *testing.T) {
	t.Parallel()

	gezaehlt, fehler := zaehlnaht(probezaehlung()).Auszaehlen(probekontext(t), 1, "farbe")
	if fehler != nil {
		t.Fatalf("Auszaehlen: %v", fehler)
	}

	if !slices.Equal(gezaehlt.Gruppen[0].KartenIDs, []string{"1"}) {
		t.Errorf("rot trägt %v", gezaehlt.Gruppen[0].KartenIDs)
	}

	if gezaehlt.Gruppen[1].KartenIDs == nil || len(gezaehlt.Gruppen[1].KartenIDs) != 0 {
		t.Errorf("blau trägt %#v", gezaehlt.Gruppen[1].KartenIDs)
	}
}

// **Ein Fehler des Dienstes kommt durch — an beiden Wegen.**
func TestDieAuszaehlnahtReichtDenFehlerDurch(t *testing.T) {
	t.Parallel()

	eigener := errors.New("kein Spielpaket")

	naht := zaehlnaht(&zaehlauskunft{fehler: eigener})

	if dimensionen, fehler := naht.Angebot(probekontext(t), 1); !errors.Is(fehler, eigener) ||
		dimensionen != nil {
		t.Errorf("Angebot gab %+v / %v", dimensionen, fehler)
	}

	if gezaehlt, fehler := naht.Auszaehlen(probekontext(t), 1, "farbe"); !errors.Is(
		fehler, eigener) || gezaehlt != nil {
		t.Errorf("Auszaehlen gab %+v / %v", gezaehlt, fehler)
	}
}
