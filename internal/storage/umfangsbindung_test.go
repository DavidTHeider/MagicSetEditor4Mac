package storage_test

// **Der Vergleich mit der Vorschau im Schreibgeschäft** — Befund 9 des
// Nachlaufs zum Zweitblick über Phase 5.2 (#1435).

import (
	"errors"
	"fmt"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// zweitesProfil legt ein zweites Profil an und gibt seine Kennung.
func zweitesProfil(t *testing.T, ablage *storage.Nutzerablage, name string) int64 {
	t.Helper()

	profil, fehler := ablage.ProfilAnlegen(t.Context(),
		domain.Profil{NutzerID: 1, Name: name})
	if fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	return profil.ID
}

// **Die Bindung sieht das Profil, das gleich verschwindet — mit Namen.**
func TestDieBindungSiehtDasProfilImGeschaeft(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))
	kennung := zweitesProfil(t, ablage, "Turnier")

	var gesehen []domain.Profil

	if fehler := ablage.ProfilLoeschen(t.Context(), kennung, "Turnier", 0,
		func(profile []domain.Profil, umfaenge [][]domain.Betroffenes) error {
			gesehen = profile

			if len(umfaenge) != len(profile) {
				return fmt.Errorf("%d Umfänge zu %d Profilen",
					len(umfaenge), len(profile))
			}

			return nil
		},
	); fehler != nil {
		t.Fatalf("löschen: %v", fehler)
	}

	if len(gesehen) != 1 || gesehen[0].ID != kennung || gesehen[0].Name != "Turnier" {
		t.Fatalf("die Bindung sah %+v", gesehen)
	}
}

// **Sagt die Bindung Nein, bleibt das Profil stehen.**
//
// **Das ist die Zusage**: der Vergleich liegt im Geschäft, und ein Nein dort
// nimmt das Löschen mit zurück — nicht nur den Vergleich.
func TestEinNeinDerBindungLaesstDasProfilStehen(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))
	kennung := zweitesProfil(t, ablage, "Turnier")

	eigener := errors.New("so nicht")

	if fehler := ablage.ProfilLoeschen(t.Context(), kennung, "Turnier", 0,
		func([]domain.Profil, [][]domain.Betroffenes) error { return eigener },
	); !errors.Is(fehler, eigener) {
		t.Fatalf("abgewiesen wurde mit %v", fehler)
	}

	profile, lesefehler := ablage.Profile(t.Context(), 1)
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if len(profile) != 2 {
		t.Errorf("%d Profile — gelöscht wurde trotzdem", len(profile))
	}
}

// **Beim Menschen zählt sie alle seine Profile.**
func TestDieBindungSiehtAlleProfileEinesMenschen(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueNutzerablage(db)

	zweitesProfil(t, ablage, "Turnier")
	zweitesProfil(t, ablage, "Drittes")

	zweiter, fehler := ablage.NutzerAnlegen(t.Context(),
		domain.Nutzer{Name: "Zweiter"})
	if fehler != nil {
		t.Fatalf("den zweiten Menschen anlegen: %v", fehler)
	}

	_ = zweiter

	var gesehen int

	if loeschfehler := ablage.NutzerLoeschen(t.Context(), 1, storage.Erstnutzer,
		func(profile []domain.Profil, umfaenge [][]domain.Betroffenes) error {
			gesehen = len(profile)

			if len(umfaenge) != len(profile) {
				return fmt.Errorf("%d Umfänge zu %d Profilen",
					len(umfaenge), len(profile))
			}

			return nil
		},
	); loeschfehler != nil {
		t.Fatalf("löschen: %v", loeschfehler)
	}

	if gesehen != 3 {
		t.Errorf("die Bindung sah %d Profile, erwartet 3", gesehen)
	}
}

// **Und ohne Bindung bleibt alles, wie es war** — die Gegenprobe dazu.
//
// **Ohne sie sagte der Lauf oben nur, dass ein Nein durchkommt**, nicht dass
// die Löschung ohne Bindung überhaupt geht.
func TestOhneBindungLoeschtDerWegWieVorher(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))
	kennung := zweitesProfil(t, ablage, "Turnier")

	if fehler := ablage.ProfilLoeschen(
		t.Context(), kennung, "Turnier", 0); fehler != nil {
		t.Fatalf("löschen: %v", fehler)
	}

	profile, lesefehler := ablage.Profile(t.Context(), 1)
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if len(profile) != 1 {
		t.Errorf("%d Profile", len(profile))
	}
}

// **Eine nil-Bindung ist keine** — der Weg, den `service` nimmt, wenn keine
// Revision genannt ist.
func TestEineLeereBindungHaeltNichtsAuf(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))
	kennung := zweitesProfil(t, ablage, "Turnier")

	if fehler := ablage.ProfilLoeschen(t.Context(), kennung, "Turnier", 0,
		nil); fehler != nil {
		t.Fatalf("löschen: %v", fehler)
	}
}

// **Ein Zähler, der klemmt, meldet sich und löscht nicht.**
//
// **Über eine entfernte Tafel**: der Umfang zählt Sets, Karten und mehr, und
// ohne die Tafel `sets` scheitert die erste Zählung **im Geschäft** — genau
// dort, wo die Bindung läuft.
func TestEinKlemmenderZaehlerHaeltDieLoeschungAuf(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueNutzerablage(db)
	kennung := zweitesProfil(t, ablage, "Turnier")

	if _, fehler := db.ExecContext(t.Context(), `DROP TABLE sets`); fehler != nil {
		t.Fatalf("die Tafel entfernen: %v", fehler)
	}

	gerufen := false

	if fehler := ablage.ProfilLoeschen(t.Context(), kennung, "Turnier", 0,
		func([]domain.Profil, [][]domain.Betroffenes) error {
			gerufen = true

			return nil
		},
	); fehler == nil {
		t.Fatal("es kam durch")
	}

	if gerufen {
		t.Error("die Bindung wurde trotz des Fehlers gerufen")
	}
}
