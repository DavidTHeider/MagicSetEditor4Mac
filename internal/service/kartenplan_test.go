package service_test

import (
	"context"
	"slices"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// festerStil gibt immer denselben Stil.
type festerStil struct {
	vorlage service.Vorlage
	befunde []domain.Befund
	da      bool
}

func (f festerStil) StilFuerSet(context.Context, int64) (
	service.Vorlage, []domain.Befund, bool,
) {
	return f.vorlage, f.befunde, f.da
}

// **Derselbe Stil für jede Karte.** Dass die Karte einen eigenen tragen darf,
// prüft `kartenstil_test.go`; hier geht es um den Weg zum Plan (#395).
func (f festerStil) StilFuerKarte(context.Context, *domain.CardRow) (
	service.Vorlage, []domain.Befund, bool,
) {
	return f.vorlage, f.befunde, f.da
}

// planprobekarte gibt eine Karte, wie sie aus dem Bestand kommt.
func planprobekarte() *domain.CardRow {
	return &domain.CardRow{ID: 1, SetID: 7, Number: 1, Values: []domain.Value{
		{Field: "name", Text: "Feuerball"},
	}}
}

// **Der Weg, den #206 verlangt: die Anwendung bekommt den Stilplan.**
//
// # Woran der Test die beiden Pläne unterscheidet
//
// **Nicht an den Maßen.** Der erste Anlauf verglich 375 × 523 gegen die Größe
// des Stils — und der Stil der eigenen Saat misst genau das. Der Test war rot,
// obwohl der Weg stimmte; ein anderer Stil hätte ihn grün gemacht, ohne dass
// etwas geprüft gewesen wäre.
//
// **Sondern am Text.** Der Behelfsplan schreibt `feld: wert`, weil er nicht
// weiß, wo etwas hingehört; der Stilplan schreibt den Wert an seine Stelle.
func TestDerKartenplanNimmtDenStilWennEsIhnGibt(t *testing.T) {
	bauer := service.NeuerPlanbauer().MitStilen(festerStil{vorlage: service.Vorlage{Stil: saatstil(t)}, da: true})

	plan, _, err := bauer.Kartenplan(t.Context(), planprobekarte())
	if err != nil {
		t.Fatalf("bauen: %v", err)
	}

	for _, ebene := range plan.Ebenen {
		if strings.HasPrefix(ebene.Klartext(), "name: ") {
			t.Errorf("die Ebene trägt %q — das ist der Behelfsplan", ebene.Klartext())
		}
	}

	if !slices.ContainsFunc(plan.Ebenen, func(e render.Ebene) bool {
		return e.Klartext() == "Feuerball"
	}) {
		t.Errorf("keine Ebene trägt den Kartennamen: %+v", plan.Ebenen)
	}
}

// **Ohne Stilquelle bleibt es beim Behelfsplan — mit einem Befund.**
//
// Der Befund ist der Punkt: eine Anwendung, die stillschweigend behilft, sieht
// aus wie eine, die den Stil nicht kann (§8).
func TestOhneStilquelleKommtDerBehelfsplanMitBefund(t *testing.T) {
	plan, befunde, err := service.NeuerPlanbauer().Kartenplan(t.Context(), planprobekarte())
	if err != nil {
		t.Fatalf("bauen: %v", err)
	}

	if plan.Breite != service.Kartenbreite {
		t.Errorf("Breite %g, erwartet war der Behelfsplan", plan.Breite)
	}

	if len(befunde) == 0 {
		t.Error("kein Befund — niemand sieht, dass hier behelfen wird")
	}
}

// **Die Befunde der Stilquelle kommen durch.**
//
// Sie sagen, *warum* kein Stil da ist — „kein Paketverzeichnis eingerichtet"
// ist eine andere Auskunft als „das Paket fehlt".
func TestDieBefundeDerStilquelleErreichenDenPlan(t *testing.T) {
	quelle := festerStil{da: false, befunde: []domain.Befund{{Meldung: "das Paket fehlt"}}}

	_, befunde, err := service.NeuerPlanbauer().MitStilen(quelle).
		Kartenplan(t.Context(), planprobekarte())
	if err != nil {
		t.Fatalf("bauen: %v", err)
	}

	if len(befunde) != 1 || befunde[0].Meldung != "das Paket fehlt" {
		t.Errorf("%v", befunde)
	}
}

// **Ein Stil, der keine Kartengröße nennt, kostet die Vorschau nicht.**
//
// `Stilplan` weist ihn mit einem Fehler ab, und das ist dort richtig. Hier wird
// daraus der Behelfsplan samt Grund: eine leere Seite wäre die schlechtere
// Antwort, und dass ein fremdes Paket unvollständig ist, ist der Normalfall
// (§8).
func TestEinStilOhneKartengroesseFuehrtZumBehelfsplan(t *testing.T) {
	quelle := festerStil{da: true}

	plan, befunde, err := service.NeuerPlanbauer().MitStilen(quelle).
		Kartenplan(t.Context(), planprobekarte())
	if err != nil {
		t.Fatalf("bauen: %v", err)
	}

	if plan.Breite != service.Kartenbreite {
		t.Errorf("Breite %g, erwartet war der Behelfsplan", plan.Breite)
	}

	if len(befunde) == 0 {
		t.Error("kein Befund — der Grund für den Behelf fehlt")
	}
}

// **Ein abgebrochener Lauf bricht ab.**
//
// Eine Vorschau wird abgebrochen, nicht abgewartet (§7).
func TestEinAbgebrochenerKartenplanBrichtAb(t *testing.T) {
	ctx, abbrechen := context.WithCancel(t.Context())
	abbrechen()

	if _, _, err := service.NeuerPlanbauer().Kartenplan(ctx, planprobekarte()); err == nil {
		t.Error("kein Fehler bei abgebrochenem Kontext")
	}
}

// abbrechenderStil bricht den Lauf ab und gibt einen untauglichen Stil.
type abbrechenderStil struct{ abbrechen context.CancelFunc }

func (a abbrechenderStil) StilFuerSet(context.Context, int64) (
	service.Vorlage, []domain.Befund, bool,
) {
	a.abbrechen()

	return service.Vorlage{}, nil, true
}

func (a abbrechenderStil) StilFuerKarte(context.Context, *domain.CardRow) (
	service.Vorlage, []domain.Befund, bool,
) {
	a.abbrechen()

	return service.Vorlage{}, nil, true
}

// **Ein Abbruch mitten im Stilplan bleibt ein Abbruch.**
//
// `Stilplan` scheitert hier an der fehlenden Kartengröße — dieselbe Stelle, an
// der ein untauglicher Stil zum Behelfsplan führt. Wer nicht unterscheidet,
// zeichnet für eine Vorschau, die niemand mehr sehen will (§7).
func TestEinAbbruchImStilplanFuehrtNichtZumBehelf(t *testing.T) {
	ctx, abbrechen := context.WithCancel(t.Context())

	bauer := service.NeuerPlanbauer().MitStilen(abbrechenderStil{abbrechen: abbrechen})

	plan, _, err := bauer.Kartenplan(ctx, planprobekarte())
	if err == nil {
		t.Errorf("kein Fehler, Plan %+v — ein Abbruch wurde zum Behelfsplan", plan)
	}
}
