package storage_test

// **Die Revision der Rechte** — Werkstück O1 von Phase 5.2
// (#1425).
//
// **Sie steigt in der Ablage, nicht im Aufrufer**: jeder Weg, der Zeiger oder
// Freigaben schreibt, erhöht sie — auch die Kaskade beim Löschen eines
// Profils, die kein Aufrufer sieht.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestDieRevisionSteigtMitJedemWechselUndJederFreigabe(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	nutzer := storage.NeueNutzerablage(db)
	freigaben := storage.NeueFreigabenablage(db)

	revision := func(schritt string, erwartet int64) {
		t.Helper()

		stand, fehler := nutzer.Revision(t.Context())
		if fehler != nil {
			t.Fatalf("%s: Revision lesen: %v", schritt, fehler)
		}

		if stand != erwartet {
			t.Errorf("%s: Revision %d, erwartet %d", schritt, stand, erwartet)
		}
	}

	// **Die Wanderung selbst setzt schon Zeiger** — der Stand danach ist der
	// Nullpunkt dieses Laufs, nicht die Null.
	anfang, fehler := nutzer.Revision(t.Context())
	if fehler != nil {
		t.Fatalf("Anfang: %v", fehler)
	}

	zweiter, fehler := nutzer.NutzerAnlegen(t.Context(), domain.Nutzer{Name: "Zweiter"})
	if fehler != nil {
		t.Fatalf("Nutzer anlegen: %v", fehler)
	}

	erster, fehler := nutzer.Nutzerzeiger(t.Context())
	if fehler != nil {
		t.Fatalf("wer arbeitet: %v", fehler)
	}

	profile, fehler := nutzer.Profile(t.Context(), zweiter.ID)
	if fehler != nil || len(profile) == 0 {
		t.Fatalf("Profile des Zweiten: %v %v", profile, fehler)
	}

	if fehler := nutzer.ZeigerSetzen(t.Context(), zweiter.ID, profile[0].ID); fehler != nil {
		t.Fatalf("Zeiger setzen: %v", fehler)
	}

	// **Beide Zeiger in einem Zug — zwei Schreiber, zwei Schritte.** Was
	// zählt, ist, dass sie steigt; ob um eins oder zwei, sagt die Zahl der
	// Zeilen, die sich ändern.
	nachZeiger := revisionVon(t, nutzer)
	if nachZeiger <= anfang {
		t.Fatalf("ZeigerSetzen ließ die Revision bei %d", nachZeiger)
	}

	if fehler := nutzer.NutzerzeigerSetzen(t.Context(), erster); fehler != nil {
		t.Fatalf("Nutzerzeiger: %v", fehler)
	}

	revision("Nutzerzeiger", nachZeiger+1)

	if fehler := nutzer.ProfilzeigerSetzen(t.Context(), profile[0].ID); fehler != nil {
		t.Fatalf("Profilzeiger: %v", fehler)
	}

	revision("Profilzeiger", nachZeiger+2)

	freigabe := domain.Profilfreigabe{
		ProfilID: profile[0].ID, NutzerID: erster, Rolle: domain.RolleLeser,
	}
	if fehler := freigaben.FreigabeSetzen(t.Context(), freigabe); fehler != nil {
		t.Fatalf("Freigabe setzen: %v", fehler)
	}

	revision("Freigabe neu", nachZeiger+3)

	freigabe.Rolle = domain.RolleMitarbeiter
	if fehler := freigaben.FreigabeSetzen(t.Context(), freigabe); fehler != nil {
		t.Fatalf("Freigabe fortschreiben: %v", fehler)
	}

	revision("Freigabe fortgeschrieben", nachZeiger+4)

	if fehler := freigaben.FreigabeLoeschen(t.Context(), profile[0].ID, erster); fehler != nil {
		t.Fatalf("Freigabe löschen: %v", fehler)
	}

	revision("Freigabe weg", nachZeiger+5)

	// **Und die Kaskade zählt mit**: eine Freigabe, die mit ihrem Profil
	// verschwindet, ist ein Entzug, den niemand ausgesprochen hat.
	if fehler := freigaben.FreigabeSetzen(t.Context(), freigabe); fehler != nil {
		t.Fatalf("Freigabe erneut: %v", fehler)
	}

	if fehler := nutzer.ProfilzeigerSetzen(t.Context(), 1); fehler != nil {
		t.Fatalf("zurück ins erste Profil: %v", fehler)
	}

	vorKaskade := revisionVon(t, nutzer)

	if fehler := nutzer.NutzerLoeschen(t.Context(), zweiter.ID, "Zweiter"); fehler != nil {
		t.Fatalf("Nutzer löschen: %v", fehler)
	}

	nachKaskade := revisionVon(t, nutzer)
	if nachKaskade <= vorKaskade {
		t.Errorf("die Kaskade ließ die Revision bei %d", nachKaskade)
	}
}

// **Und eine Revision aus einem kaputten Bestand ist ein Fehler**, kein
// stilles Null — eine Null sähe aus wie ein Bestand, in dem nie jemand war.
func TestDieRevisionMeldetEinenKaputtenBestand(t *testing.T) {
	t.Parallel()

	if _, fehler := storage.NeueNutzerablage(kaputterBestand(t)).Revision(t.Context()); fehler == nil {
		t.Error("die Revision kam aus dem kaputten Bestand")
	}
}

// revisionVon liest die Revision — und bricht ab, wenn das nicht geht.
func revisionVon(t *testing.T, nutzer *storage.Nutzerablage) int64 {
	t.Helper()

	stand, fehler := nutzer.Revision(t.Context())
	if fehler != nil {
		t.Fatalf("Revision lesen: %v", fehler)
	}

	return stand
}
