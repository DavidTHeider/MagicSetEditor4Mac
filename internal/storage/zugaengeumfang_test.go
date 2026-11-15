package storage_test

// **Was mit einem Menschen oder Profil an Zugängen verschwindet** — #2287
// (#2287),
// aus dem Nachlauf über Phase G.
//
// Profilfreigaben hängen mit `ON DELETE CASCADE` an Profil **und** Mensch
// (Wanderung 0044). Die Vorschau einer Löschung zählte sie in keiner der
// beiden Richtungen: weder die Zugänge anderer an einem Profil, das geht,
// noch die Zugänge eines Menschen, der geht, an fremden Profilen.

import (
	"context"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// gruppeIn sucht eine Gruppe in einem Umfang.
func gruppeIn(umfang []domain.Betroffenes, was string) (domain.Betroffenes, bool) {
	for _, gruppe := range umfang {
		if gruppe.Was == was {
			return gruppe, true
		}
	}

	return domain.Betroffenes{}, false
}

func TestDerProfilumfangZaehltDieZugaengeAnderer(t *testing.T) {
	db := frischeDatenbank(t)
	zweiter := zweiterNutzer(t, db)

	if fehler := storage.NeueFreigabenablage(db).FreigabeSetzen(t.Context(), domain.Profilfreigabe{
		ProfilID: 1, NutzerID: zweiter, Rolle: domain.RolleLeser,
	}); fehler != nil {
		t.Fatalf("freigeben: %v", fehler)
	}

	umfang, fehler := storage.Profilumfang(probekontext(t), db, 1)
	if fehler != nil {
		t.Fatalf("der Umfang: %v", fehler)
	}

	zugaenge, gibt := gruppeIn(umfang, domain.BetroffenZugaenge)
	if !gibt || zugaenge.Anzahl != 1 || len(zugaenge.Namen) != 1 || zugaenge.Namen[0] != "Zweiter" {
		t.Errorf("die Zugänge sind %+v (%v) — erwartet einer, „Zweiter“", zugaenge, gibt)
	}
}

func TestFremdeZugaengeZaehlenOhneNamen(t *testing.T) {
	db := frischeDatenbank(t)
	zweiter := zweiterNutzer(t, db)

	if fehler := storage.NeueFreigabenablage(db).FreigabeSetzen(t.Context(), domain.Profilfreigabe{
		ProfilID: 1, NutzerID: zweiter, Rolle: domain.RolleMitarbeiter,
	}); fehler != nil {
		t.Fatalf("freigeben: %v", fehler)
	}

	fremd, fehler := storage.FremdeZugaenge(probekontext(t), db, zweiter)
	if fehler != nil {
		t.Fatalf("die fremden Zugänge: %v", fehler)
	}

	if fremd.Was != domain.BetroffenZugaenge || fremd.Anzahl != 1 || len(fremd.Namen) != 0 {
		t.Errorf("die fremden Zugänge sind %+v — erwartet einer, ohne Namen", fremd)
	}

	// **Der erste Mensch hat keinen** — sein Profil ist sein eigenes.
	eigen, eigenfehler := storage.FremdeZugaenge(probekontext(t), db, 1)
	if eigenfehler != nil || eigen.Anzahl != 0 {
		t.Errorf("der erste Mensch: %+v, %v", eigen, eigenfehler)
	}
}

// **Und das Schreibgeschäft sieht dieselbe Zahl** — sonst wiese die Bindung
// jede Löschung eines Menschen ab, der einen fremden Zugang hat.
func TestDieLoeschungEinesMenschenBindetSeineFremdenZugaenge(t *testing.T) {
	db := frischeDatenbank(t)
	zweiter := zweiterNutzer(t, db)

	if fehler := storage.NeueFreigabenablage(db).FreigabeSetzen(t.Context(), domain.Profilfreigabe{
		ProfilID: 1, NutzerID: zweiter, Rolle: domain.RolleLeser,
	}); fehler != nil {
		t.Fatalf("freigeben: %v", fehler)
	}

	var gesehen []domain.Betroffenes

	bindung := func(_ []domain.Profil, umfaenge [][]domain.Betroffenes) error {
		for _, umfang := range umfaenge {
			gesehen = append(gesehen, umfang...)
		}

		return nil
	}

	if fehler := storage.NeueNutzerablage(db).NutzerLoeschen(context.Background(), zweiter,
		"Zweiter", bindung); fehler != nil {
		t.Fatalf("löschen: %v", fehler)
	}

	if zugaenge, gibt := gruppeIn(gesehen, domain.BetroffenZugaenge); !gibt || zugaenge.Anzahl != 1 {
		t.Errorf("das Geschäft sah %+v — erwartet den einen fremden Zugang", gesehen)
	}
}
