package storage_test

// **Die Ablage des Arbeitsverlaufs** — Werkstück U11 von Phase 5.1 (#1422).

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func arbeitsschritt(kennung string, vorgang domain.Arbeitsvorgang, stunde int) domain.Arbeitsschritt {
	return domain.Arbeitsschritt{
		Objekt:    domain.Arbeitsobjekt{Art: domain.ArbeitsartSet, Kennung: kennung, Name: "Set " + kennung},
		Vorgang:   vorgang,
		Zeitpunkt: time.Date(2026, time.September, 9, stunde, 0, 0, 0, time.UTC),
	}
}

func TestSchritteWerdenVermerktUndJuengsterGewinnt(t *testing.T) {
	t.Parallel()

	ablage := storage.NeuerArbeitsverlauf(frischeDatenbank(t))
	ctx := probekontext(t)

	for _, schritt := range []domain.Arbeitsschritt{
		arbeitsschritt("1", domain.VorgangGeoeffnet, 8),
		arbeitsschritt("2", domain.VorgangGeoeffnet, 9),
		arbeitsschritt("1", domain.VorgangGeoeffnet, 10),
		arbeitsschritt("1", domain.VorgangAusgegeben, 11),
	} {
		if err := ablage.Vermerken(ctx, schritt); err != nil {
			t.Fatalf("vermerken: %v", err)
		}
	}

	schritte, err := ablage.Schritte(ctx)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	// **Drei Zeilen, nicht vier**: Set 1 geöffnet steht einmal, mit zehn Uhr.
	if len(schritte) != 3 {
		t.Fatalf("%d Schritte, erwartet 3: %+v", len(schritte), schritte)
	}

	if schritte[0].Objekt.Kennung != "1" || schritte[0].Vorgang != domain.VorgangAusgegeben ||
		schritte[1].Objekt.Kennung != "1" || schritte[1].Zeitpunkt.Hour() != 10 {
		t.Errorf("Reihenfolge: %+v", schritte)
	}

	if schritte[0].Objekt.Name != "Set 1" || schritte[0].Zeitpunkt.Location() != time.UTC {
		t.Errorf("der Schnappschuss ist %+v", schritte[0])
	}
}

// **Ein fremdes Profil sieht nichts** — nicht die Zeilen, nicht die Namen.
func TestDerVerlaufIstJeProfil(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	ablage := storage.NeuerArbeitsverlauf(db)

	if err := ablage.Vermerken(storage.MitProfil(context.Background(), 1), arbeitsschritt("1", domain.VorgangGeoeffnet, 8)); err != nil {
		t.Fatalf("vermerken: %v", err)
	}

	if err := ablage.Anheften(storage.MitProfil(context.Background(), 1), domain.Arbeitsobjekt{Art: domain.ArbeitsartSet, Kennung: "1", Name: "Geheim"}); err != nil {
		t.Fatalf("anheften: %v", err)
	}

	fremd := storage.MitProfil(context.Background(), 2)

	if schritte, err := ablage.Schritte(fremd); err != nil || len(schritte) != 0 {
		t.Errorf("Profil 2 sieht %v, %v", schritte, err)
	}

	if angeheftet, err := ablage.Angeheftete(fremd); err != nil || len(angeheftet) != 0 {
		t.Errorf("Profil 2 sieht angeheftet %v, %v", angeheftet, err)
	}

	// **Ohne Profil geht nichts** — weder lesen noch schreiben.
	if _, err := ablage.Schritte(context.Background()); err == nil {
		t.Error("ohne Profil kamen Schritte")
	}

	if err := ablage.Vermerken(context.Background(), arbeitsschritt("1", domain.VorgangGeoeffnet, 8)); err == nil {
		t.Error("ohne Profil wurde vermerkt")
	}
}

func TestAnheftenUndLoesen(t *testing.T) {
	t.Parallel()

	ablage := storage.NeuerArbeitsverlauf(frischeDatenbank(t))
	ctx := probekontext(t)
	deck := domain.Arbeitsobjekt{Art: domain.ArbeitsartDeck, Kennung: "3", Name: "Mono Rot"}

	if err := ablage.Anheften(ctx, deck); err != nil {
		t.Fatalf("anheften: %v", err)
	}

	// **Zweimal anheften ist einmal** — und der neue Name gewinnt.
	deck.Name = "Mono Rot, Fassung zwei"
	if err := ablage.Anheften(ctx, deck); err != nil {
		t.Fatalf("noch einmal anheften: %v", err)
	}

	angeheftet, err := ablage.Angeheftete(ctx)
	if err != nil || len(angeheftet) != 1 || angeheftet[0].Name != "Mono Rot, Fassung zwei" {
		t.Fatalf("angeheftet %v, %v", angeheftet, err)
	}

	geloest, err := ablage.Loesen(ctx, domain.ArbeitsartDeck, "3")
	if err != nil || !geloest {
		t.Fatalf("lösen: %v, %v", geloest, err)
	}

	if nochmal, err := ablage.Loesen(ctx, domain.ArbeitsartDeck, "3"); err != nil || nochmal {
		t.Errorf("ein zweites Lösen sagte %v, %v", nochmal, err)
	}

	if angeheftet, err := ablage.Angeheftete(ctx); err != nil || len(angeheftet) != 0 {
		t.Errorf("nach dem Lösen: %v, %v", angeheftet, err)
	}
}

// **Was es nicht mehr gibt, wird vergessen** — aus beiden Tabellen.
func TestVergessenRaeumtSchritteUndAngeheftetes(t *testing.T) {
	t.Parallel()

	ablage := storage.NeuerArbeitsverlauf(frischeDatenbank(t))
	ctx := probekontext(t)

	if err := ablage.Vermerken(ctx, arbeitsschritt("1", domain.VorgangGeoeffnet, 8)); err != nil {
		t.Fatalf("vermerken: %v", err)
	}

	if err := ablage.Anheften(ctx, domain.Arbeitsobjekt{Art: domain.ArbeitsartSet, Kennung: "1"}); err != nil {
		t.Fatalf("anheften: %v", err)
	}

	if err := ablage.Vergessen(ctx, domain.ArbeitsartSet, "1"); err != nil {
		t.Fatalf("vergessen: %v", err)
	}

	schritte, err := ablage.Schritte(ctx)
	if err != nil {
		t.Fatalf("schritte: %v", err)
	}

	angeheftet, err := ablage.Angeheftete(ctx)
	if err != nil {
		t.Fatalf("angeheftet: %v", err)
	}

	if len(schritte) != 0 || len(angeheftet) != 0 {
		t.Errorf("nach dem Vergessen: %v / %v", schritte, angeheftet)
	}
}

func TestUngueltigesWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	ablage := storage.NeuerArbeitsverlauf(frischeDatenbank(t))
	ctx := probekontext(t)

	if err := ablage.Vermerken(ctx, domain.Arbeitsschritt{}); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("ein leerer Schritt gab %v", err)
	}

	if err := ablage.Anheften(ctx, domain.Arbeitsobjekt{Art: "karte", Kennung: "1"}); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("eine erfundene Art gab %v", err)
	}
}
