package storage_test

// **Die Handel eines Profils, je Vorgang die letzte Revision** — UX-08.AK-04
// (#1680):
// wer einen Handel empfängt, soll ihn wählen und nicht seine Kennung tippen —
// und dafür muss es die Liste geben.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Je Vorgang einer, und zwar der letzte — zuletzt geänderte zuerst.**
func TestLetzteGibtJedenVorgangMitSeinerLetztenRevision(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	ablage := storage.NeueHandelsablage(db)
	ctx := probekontext(t)

	erste := einVorgangsangebot()
	zweite := einVorgangsangebot(func(a *domain.Handelsangebot) {
		a.Revision = 2
		a.Eltern = erste.Abdruck()
		a.Zustand = domain.HandelAngeboten
	})
	anderer := einVorgangsangebot(func(a *domain.Handelsangebot) {
		a.Vorgang = "h2"
	})

	for _, angebot := range []domain.Handelsangebot{erste, zweite, anderer} {
		if fehler := ablage.Schreiben(ctx, angebot); fehler != nil {
			t.Fatalf("schreiben: %v", fehler)
		}
	}

	// **Der ältere Vorgang bekommt eine ältere Zeit** — sonst entschiede
	// dieselbe Sekunde, und die Probe maß nur den Rückfall.
	if _, fehler := db.ExecContext(ctx,
		`UPDATE handelsrevisionen SET geaendert = '2026-01-01 00:00:00'
		  WHERE vorgang = 'h1'`); fehler != nil {
		t.Fatalf("zurückdatieren: %v", fehler)
	}

	liste, fehler := ablage.Letzte(ctx)
	if fehler != nil {
		t.Fatalf("Letzte: %v", fehler)
	}

	if len(liste) != 2 {
		t.Fatalf("%d Vorgänge, erwartet 2: %+v", len(liste), liste)
	}

	if liste[0].Vorgang != "h2" || liste[1].Vorgang != "h1" {
		t.Errorf("Reihe %s, %s — der zuletzt geänderte gehört nach vorn",
			liste[0].Vorgang, liste[1].Vorgang)
	}

	if liste[1].Revision != 2 || liste[1].Zustand != domain.HandelAngeboten {
		t.Errorf("h1 kam als Revision %d im Zustand %s", liste[1].Revision,
			liste[1].Zustand)
	}
}

// **Ein fremdes Profil sieht nichts davon** — und ein leeres Profil bekommt
// eine leere Liste und kein `nil`: „noch kein Handel" ist eine Antwort.
func TestLetzteSiehtNurDasEigeneProfil(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	ablage := storage.NeueHandelsablage(db)

	if fehler := ablage.Schreiben(probekontext(t), einVorgangsangebot()); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	fremd := zweitesProfil(t, storage.NeueNutzerablage(db), "Fremd")

	liste, fehler := ablage.Letzte(storage.MitProfil(t.Context(), fremd))
	if fehler != nil {
		t.Fatalf("Letzte: %v", fehler)
	}

	if liste == nil || len(liste) != 0 {
		t.Errorf("das fremde Profil sah %+v", liste)
	}
}
