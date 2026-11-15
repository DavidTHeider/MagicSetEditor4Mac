package storage_test

// **Kontakte im Bestand** — ID-04
// (#1703).

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestEinKontaktGehtHineinUndKommtWieder(t *testing.T) {
	ablage := storage.NeueAdressbuchablage(frischeDatenbank(t))
	ctx := probekontext(t)

	ada := domain.Kontakt{
		Name: "Ada Lovelace", Ursprung: "0A1B", Bild: []byte{1, 2}, Bildart: "image/png",
		Wege: []domain.Kontaktweg{{Art: domain.KontaktwegEmail, Wert: "ada@example.org"}},
	}

	// **Ohne Wege und ohne Bild geht es auch.**
	if fehler := ablage.KontakteUebernehmen(ctx, []*domain.Kontakt{&ada, {Name: "Bo"}}); fehler != nil {
		t.Fatal(fehler)
	}

	kontakte, fehler := ablage.Kontakte(ctx)
	if fehler != nil || len(kontakte) != 2 {
		t.Fatalf("%v %v", kontakte, fehler)
	}

	erster := kontakte[0]
	if erster.ID == 0 || erster.Fassung != 1 || erster.Name != "Ada Lovelace" || len(erster.Wege) != 1 ||
		erster.Wege[0].Wert != "ada@example.org" || string(erster.Bild) != "\x01\x02" || erster.Uebernommen.IsZero() {
		t.Errorf("%+v", erster)
	}

	if len(kontakte[1].Bild) != 0 || kontakte[1].Wege == nil {
		t.Errorf("ohne Bild und Wege: %+v", kontakte[1])
	}

	// **Dieselbe Karte zweimal ist eine Dublette** — der Ursprung ist je
	// Profil einmalig —, **und mit ihr geht nichts hinein**: auch der neue
	// Kontakt davor nicht (ein Geschäft).
	cy := domain.Kontakt{Name: "Cy"}
	if fehler := ablage.KontakteUebernehmen(ctx, []*domain.Kontakt{&cy, &ada}); !errors.Is(fehler, domain.ErrDoppelt) {
		t.Errorf("dieselbe Karte zweimal: %v", fehler)
	}

	if nachher, lesefehler := ablage.Kontakte(ctx); lesefehler != nil || len(nachher) != 2 {
		t.Errorf("eine halbe Übernahme blieb: %d Kontakte", len(nachher))
	}

	// **Ein anderes Profil sieht ihn nicht.**
	if andere, lesefehler := ablage.Kontakte(storage.MitProfil(t.Context(), 99)); lesefehler != nil || len(andere) != 0 {
		t.Errorf("ein fremdes Profil sieht %v", andere)
	}
}

func TestEinKontaktWirdGegenSeineFassungErsetztUndGeloescht(t *testing.T) {
	ablage := storage.NeueAdressbuchablage(frischeDatenbank(t))
	ctx := probekontext(t)

	if fehler := ablage.KontakteUebernehmen(ctx, []*domain.Kontakt{{Name: "Ada", Ursprung: "0A1B"}}); fehler != nil {
		t.Fatal(fehler)
	}

	gelesen, lesefehler := ablage.Kontakte(ctx)
	if lesefehler != nil || len(gelesen) != 1 {
		t.Fatalf("%v %v", gelesen, lesefehler)
	}

	id := gelesen[0].ID

	neu := domain.Kontakt{
		ID: id, Fassung: 1, Name: "Ada Lovelace", Ursprung: "0A1B",
		Wege: []domain.Kontaktweg{{Art: domain.KontaktwegTelefon, Wert: "1"}},
	}

	if fehler := ablage.KontakteUebernehmen(ctx, []*domain.Kontakt{&neu}); fehler != nil {
		t.Fatal(fehler)
	}

	// **Gegen eine überholte Fassung ändert sich nichts.**
	if fehler := ablage.KontakteUebernehmen(ctx, []*domain.Kontakt{&neu}); !errors.Is(fehler, domain.ErrUeberholt) {
		t.Errorf("überholt: %v", fehler)
	}

	if fehler := ablage.KontaktLoeschen(ctx, id, 1); !errors.Is(fehler, domain.ErrUeberholt) {
		t.Errorf("überholt löschen: %v", fehler)
	}

	for name, fehler := range map[string]error{
		"ersetzen": ablage.KontakteUebernehmen(ctx, []*domain.Kontakt{{ID: 4711, Fassung: 1, Name: "X"}}),
		"löschen":  ablage.KontaktLoeschen(ctx, 4711, 1),
	} {
		if !errors.Is(fehler, domain.ErrNotFound) {
			t.Errorf("%s ohne Kontakt: %v", name, fehler)
		}
	}

	kontakte, lesefehler := ablage.Kontakte(ctx)
	if lesefehler != nil {
		t.Fatal(lesefehler)
	}

	if len(kontakte) != 1 || kontakte[0].Name != "Ada Lovelace" || kontakte[0].Fassung != 2 {
		t.Fatalf("%+v", kontakte)
	}

	if fehler := ablage.KontaktLoeschen(ctx, id, 2); fehler != nil {
		t.Fatal(fehler)
	}

	if rest, lesefehler := ablage.Kontakte(ctx); lesefehler != nil || len(rest) != 0 {
		t.Errorf("nach dem Löschen: %v", rest)
	}
}
