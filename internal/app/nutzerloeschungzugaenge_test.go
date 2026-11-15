package app_test

// **Die Löschung eines Menschen zählt seine Zugänge an fremden Profilen** —
// #2287 (#2287),
// durch den echten Speicher: Vorschau, Ausführung mit ihrer Revision, und ein
// neuer Zugang dazwischen verwirft sie.

import (
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestDieNutzerloeschungZaehltUndBindetFremdeZugaenge(t *testing.T) {
	ort := filepath.Join(t.TempDir(), "bestand.db")
	db := bestandBei(t, ort)
	ctx := probekontext(t)
	nutzer := storage.NeueNutzerablage(db)
	freigaben := storage.NeueFreigabenablage(db)

	// **Der erste Mensch darf an einem fremden Profil lesen.**
	erstes := fremdesProfil(t, nutzer)
	if fehler := freigaben.FreigabeSetzen(ctx, domain.Profilfreigabe{
		ProfilID: erstes.ID, NutzerID: 1, Rolle: domain.RolleLeser,
	}); fehler != nil {
		t.Fatalf("freigeben: %v", fehler)
	}

	anwendung := app.RuecksetzungsAnwendungUeber(db, ort)

	vorher := anwendung.Ruecksetzungsvorschau(ctx, "nutzer", "", "")
	if vorher.Fehler != nil {
		t.Fatalf("die Vorschau: %+v", vorher.Fehler)
	}

	gezaehlt := map[string]app.BetroffenesDTO{}
	for _, gruppe := range vorher.Betroffen {
		gezaehlt[gruppe.Was] = gruppe
	}

	if zugaenge := gezaehlt[domain.BetroffenZugaenge]; zugaenge.Anzahl != 1 || len(zugaenge.Namen) != 0 {
		t.Fatalf("die Zugänge sind %+v — erwartet einer, ohne Namen", zugaenge)
	}

	// **Ein zweiter fremder Zugang kommt dazu** — die Revision von vorher
	// gilt nicht mehr.
	zweites, profilfehler := nutzer.ProfilAnlegen(ctx, domain.Profil{NutzerID: erstes.NutzerID, Name: "Zweites"})
	if profilfehler != nil {
		t.Fatalf("das zweite Profil: %v", profilfehler)
	}

	if fehler := freigaben.FreigabeSetzen(ctx, domain.Profilfreigabe{
		ProfilID: zweites.ID, NutzerID: 1, Rolle: domain.RolleLeser,
	}); fehler != nil {
		t.Fatalf("freigeben: %v", fehler)
	}

	if antwort := anwendung.Zuruecksetzen(ctx, "nutzer", "", "", "Ich", "", vorher.Revision); antwort.Fehler == nil {
		t.Fatal("eine überholte Vorschau löschte den Menschen")
	}

	// **Mit der heutigen Revision geht es** — Vorschau und Schreibgeschäft
	// zählen dasselbe.
	heute := anwendung.Ruecksetzungsvorschau(ctx, "nutzer", "", "")
	if antwort := anwendung.Zuruecksetzen(ctx, "nutzer", "", "", "Ich", "", heute.Revision); antwort.Fehler != nil {
		t.Errorf("mit der heutigen Revision: %+v", antwort.Fehler)
	}
}
