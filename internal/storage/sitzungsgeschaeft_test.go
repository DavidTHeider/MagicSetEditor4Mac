package storage_test

// **Ein Übergang der Sitzung ist ein Geschäft** — Nachlauf über ID-02,
// Befund 2 (#1701).
//
// Beim Anmelden standen die Zeiger vor dem festgehaltenen Zustand, bei der
// gesperrten Wiederherstellung Kennwort und Code davor. Scheiterte der
// zweite Schritt, war der erste geschrieben: ein Mensch im Zeiger, der nie
// angemeldet war, oder ein neues Kennwort hinter einer Sperre, die blieb.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// zweiMitProfil legt zwei Nutzer mit je einem Profil an und zeigt auf den
// ersten.
func zweiMitProfil(t *testing.T, ablage *storage.Nutzerablage) (erster, zweiter domain.Nutzer, zweitesProfil domain.Profil) {
	t.Helper()

	ctx := probekontext(t)

	var profile [2]domain.Profil

	for stelle, name := range []string{"Ada", "Bo"} {
		angelegt, fehler := ablage.NutzerAnlegen(ctx, domain.Nutzer{Name: name})
		if fehler != nil {
			t.Fatal(fehler)
		}

		profil, fehler := ablage.ProfilAnlegen(ctx, domain.Profil{NutzerID: angelegt.ID, Name: name})
		if fehler != nil {
			t.Fatal(fehler)
		}

		profile[stelle] = profil

		if stelle == 0 {
			erster = angelegt
		} else {
			zweiter = angelegt
		}
	}

	if fehler := ablage.ZeigerSetzen(ctx, erster.ID, profile[0].ID); fehler != nil {
		t.Fatal(fehler)
	}

	return erster, zweiter, profile[1]
}

func TestDieAnmeldungStelltZeigerUndSitzungInEinemZug(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))
	ctx := probekontext(t)

	erster, zweiter, zweitesProfil := zweiMitProfil(t, ablage)

	if fehler := ablage.SitzungFesthalten(ctx, domain.SitzungAbgemeldet); fehler != nil {
		t.Fatal(fehler)
	}

	// **Ein Profil, das es nicht gibt, lässt beides stehen.**
	if fehler := ablage.AnmeldungFesthalten(ctx, zweiter.ID, 4711); !errors.Is(fehler, domain.ErrNotFound) {
		t.Fatalf("ein fehlendes Profil: %v", fehler)
	}

	if wer, lesefehler := ablage.Nutzerzeiger(ctx); lesefehler != nil || wer != erster.ID {
		t.Errorf("der Zeiger ging trotzdem um: %d %v", wer, lesefehler)
	}

	if zustand, lesefehler := ablage.SitzungLesen(ctx); lesefehler != nil || zustand != domain.SitzungAbgemeldet {
		t.Errorf("die Sitzung ging trotzdem auf: %q %v", zustand, lesefehler)
	}

	if fehler := ablage.AnmeldungFesthalten(ctx, zweiter.ID, zweitesProfil.ID); fehler != nil {
		t.Fatal(fehler)
	}

	wer, fehler := ablage.Nutzerzeiger(ctx)
	if fehler != nil || wer != zweiter.ID {
		t.Errorf("Nutzerzeiger %d %v", wer, fehler)
	}

	profil, fehler := ablage.Profilzeiger(ctx)
	if fehler != nil || profil != zweitesProfil.ID {
		t.Errorf("Profilzeiger %d %v", profil, fehler)
	}

	if zustand, lesefehler := ablage.SitzungLesen(ctx); lesefehler != nil || zustand != domain.SitzungAngemeldet {
		t.Errorf("die Sitzung: %q %v", zustand, lesefehler)
	}
}

func TestDieGesperrteWiederherstellungOeffnetInEinemZug(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))
	ctx := probekontext(t)

	ada, fehler := ablage.NutzerAnlegen(ctx, domain.Nutzer{Name: "Ada", Kennwort: "alt"})
	if fehler != nil {
		t.Fatal(fehler)
	}

	if setzfehler := ablage.CodesSetzen(ctx, ada.ID, []string{"gehuetet-1"}); setzfehler != nil {
		t.Fatal(setzfehler)
	}

	frei, fehler := ablage.FreieCodes(ctx, ada.ID)
	if fehler != nil || len(frei) != 1 {
		t.Fatalf("%v %v", frei, fehler)
	}

	if fehler := ablage.SitzungFesthalten(ctx, domain.SitzungGesperrt); fehler != nil {
		t.Fatal(fehler)
	}

	// **Ein fremder Code öffnet nicht** — und die Sperre bleibt.
	if gesetzt := ablage.KennwortMitCodeSetzen(ctx, 4711, frei[0].ID, "neu", domain.SitzungAngemeldet); !errors.Is(gesetzt, domain.ErrNotFound) {
		t.Fatalf("fremder Nutzer: %v", gesetzt)
	}

	if zustand, lesefehler := ablage.SitzungLesen(ctx); lesefehler != nil || zustand != domain.SitzungGesperrt {
		t.Errorf("ein gescheitertes Geschäft hat entsperrt: %q %v", zustand, lesefehler)
	}

	if gesetzt := ablage.KennwortMitCodeSetzen(ctx, ada.ID, frei[0].ID, "neu", domain.SitzungAngemeldet); gesetzt != nil {
		t.Fatal(gesetzt)
	}

	if zustand, lesefehler := ablage.SitzungLesen(ctx); lesefehler != nil || zustand != domain.SitzungAngemeldet {
		t.Errorf("die Sitzung: %q %v", zustand, lesefehler)
	}
}
