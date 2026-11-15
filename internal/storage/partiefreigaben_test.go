package storage_test

// **Die Einwilligung im Bestand** — Werkstück G5 von Phase 5.5
// (#1462).

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// einePartieablage gibt eine Ablage, in der `p1` schon steht.
//
// **Mit Ereignis**, denn die Hälfte dieser Läufe fragt, was beim Vergessen
// mitgeht.
func einePartieablage(t *testing.T) (*storage.Partieablage, context.Context) {
	t.Helper()

	ablage, ctx := partieablage(t), probekontext(t)

	partie, fehler := ablage.Anlegen(ctx, einePartie())
	if fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	if fehler := ablage.Anwenden(ctx, partie.Fassung, "b1", partie,
		[]domain.Ereignis{{
			PartieID: partie.ID, Folge: 1, Art: "ziehen",
			Oeffentlich: map[string]any{"sitz": "sitz-1"},
			Kette:       "abc",
		}}); fehler != nil {
		t.Fatalf("anwenden: %v", fehler)
	}

	return ablage, ctx
}

func TestOhneEinwilligungIstDieListeLeer(t *testing.T) {
	t.Parallel()

	ablage, ctx := einePartieablage(t)

	gaben, fehler := ablage.Freigaben(ctx, "p1")
	if fehler != nil {
		t.Fatalf("Fehler: %v", fehler)
	}

	if len(gaben) != 0 {
		t.Fatalf("erwartet nichts, bekam %v", gaben)
	}
}

func TestEineEinwilligungKommtZurueck(t *testing.T) {
	t.Parallel()

	ablage, ctx := einePartieablage(t)

	if fehler := ablage.Freigeben(ctx, "p1", domain.Offenlegung{
		Von: "anna", An: []string{"ben", "cem"},
	}); fehler != nil {
		t.Fatalf("Fehler: %v", fehler)
	}

	gaben, fehler := ablage.Freigaben(ctx, "p1")
	if fehler != nil {
		t.Fatalf("Fehler: %v", fehler)
	}

	if len(gaben) != 1 || gaben[0].Von != "anna" || len(gaben[0].An) != 2 {
		t.Fatalf("erwartet eine Einwilligung an zwei, bekam %v", gaben)
	}
}

// **Die zweite ersetzt die erste** — und das ist die Zusage, für die es das
// Geschäft gibt.
func TestDieZweiteEinwilligungErsetztDieErste(t *testing.T) {
	t.Parallel()

	ablage, ctx := einePartieablage(t)

	for _, an := range [][]string{{"ben", "cem"}, {"dora"}} {
		if fehler := ablage.Freigeben(ctx, "p1", domain.Offenlegung{
			Von: "anna", An: an,
		}); fehler != nil {
			t.Fatalf("Fehler: %v", fehler)
		}
	}

	gaben, fehler := ablage.Freigaben(ctx, "p1")
	if fehler != nil {
		t.Fatalf("Fehler: %v", fehler)
	}

	if len(gaben) != 1 || len(gaben[0].An) != 1 || gaben[0].An[0] != "dora" {
		t.Fatalf("erwartet nur die zweite Liste, bekam %v", gaben)
	}
}

// **Und sie lässt die eines anderen Sitzes stehen.**
func TestEineEinwilligungLaesstDieFremdeStehen(t *testing.T) {
	t.Parallel()

	ablage, ctx := einePartieablage(t)

	for _, von := range []string{"anna", "ben"} {
		if fehler := ablage.Freigeben(ctx, "p1", domain.Offenlegung{
			Von: von, An: []string{"cem"},
		}); fehler != nil {
			t.Fatalf("Fehler: %v", fehler)
		}
	}

	gaben, fehler := ablage.Freigaben(ctx, "p1")
	if fehler != nil {
		t.Fatalf("Fehler: %v", fehler)
	}

	if len(gaben) != 2 {
		t.Fatalf("erwartet zwei Einwilligungen, bekam %v", gaben)
	}
}

func TestEineZuruecknahmeRaeumtDieZeilenWeg(t *testing.T) {
	t.Parallel()

	ablage, ctx := einePartieablage(t)

	if fehler := ablage.Freigeben(ctx, "p1", domain.Offenlegung{
		Von: "anna", An: []string{"ben"},
	}); fehler != nil {
		t.Fatalf("Fehler: %v", fehler)
	}

	if fehler := ablage.Freigabeloeschen(ctx, "p1", "anna"); fehler != nil {
		t.Fatalf("Fehler: %v", fehler)
	}

	gaben, fehler := ablage.Freigaben(ctx, "p1")
	if fehler != nil {
		t.Fatalf("Fehler: %v", fehler)
	}

	if len(gaben) != 0 {
		t.Fatalf("erwartet nichts, bekam %v", gaben)
	}
}

// **Vergessen nimmt den Strom und die Einwilligungen mit** — über die
// Fremdschlüssel, nicht über ein zweites `DELETE`.
//
// **Geprüft an einer neuen Partie mit derselben Kennung**: seit #1815 gibt es
// für eine vergessene Partie keine Freigaben mehr zu lesen, sondern ein
// „nicht gefunden". Was die Kaskade schützt, ist genau das hier — dass eine
// neue Partie nichts von der alten erbt.
func TestVergessenRaeumtAllesMit(t *testing.T) {
	t.Parallel()

	ablage, ctx := einePartieablage(t)

	if fehler := ablage.Freigeben(ctx, "p1", domain.Offenlegung{
		Von: "anna", An: []string{"ben"},
	}); fehler != nil {
		t.Fatalf("Fehler: %v", fehler)
	}

	if fehler := ablage.Vergessen(ctx, "p1"); fehler != nil {
		t.Fatalf("Fehler: %v", fehler)
	}

	if _, fehler := ablage.Partie(ctx, "p1"); !errors.Is(
		fehler, domain.ErrNotFound) {
		t.Fatalf("erwartet ErrNotFound, bekam %v", fehler)
	}

	if _, fehler := ablage.Freigaben(ctx, "p1"); !errors.Is(fehler, domain.ErrNotFound) {
		t.Fatalf("die Freigaben einer vergessenen Partie: %v", fehler)
	}

	if _, fehler := ablage.Anlegen(ctx, einePartie()); fehler != nil {
		t.Fatalf("neu anlegen: %v", fehler)
	}

	gaben, fehler := ablage.Freigaben(ctx, "p1")
	if fehler != nil {
		t.Fatalf("Fehler: %v", fehler)
	}

	if len(gaben) != 0 {
		t.Fatalf("die Einwilligungen blieben liegen: %v", gaben)
	}

	strom, fehler := ablage.Ereignisse(ctx, "p1", 0, 0)
	if fehler != nil {
		t.Fatalf("Fehler: %v", fehler)
	}

	if len(strom) != 0 {
		t.Fatalf("der Strom blieb liegen: %v", strom)
	}
}

// **Was es nicht gibt, wird auch nicht vergessen** — sonst antwortete ein
// Tippfehler mit „erledigt".
func TestEineFremdePartieWirdNichtVergessen(t *testing.T) {
	t.Parallel()

	ablage, ctx := einePartieablage(t)

	fehler := ablage.Vergessen(ctx, "gibtesnicht")

	if !errors.Is(fehler, domain.ErrNotFound) {
		t.Fatalf("erwartet ErrNotFound, bekam %v", fehler)
	}
}
