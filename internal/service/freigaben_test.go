package service_test

// **Wer in welches Profil darf** — Werkstück 11 von Phase 4.12
// (#1351).

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// einProfil ist die Vorlage, aus der die Fälle ihre Abweichung bauen.
func einProfil(teile ...func(*domain.Profil)) domain.Profil {
	profil := domain.Profil{ID: 7, NutzerID: 1, Name: "Fremd"}
	for _, teil := range teile {
		teil(&profil)
	}

	return profil
}

// **Der Eigner ist Verwalter — ohne Zeile im Bestand.**
//
// **Eine Freigabe an sich selbst wäre eine Zeile, die niemand löschen dürfte**
// — und ein Bestand, in dem sie fehlt, wäre kaputt.
func TestDerEignerIstVerwalter(t *testing.T) {
	t.Parallel()

	ablage := neuesFreigabendoppel()
	verwaltung := service.NeueFreigabenverwaltung(ablage, nil)

	rolle, fehler := verwaltung.Rolle(context.Background(), 1, einProfil())
	if fehler != nil {
		t.Fatalf("%v", fehler)
	}

	if rolle != domain.RolleVerwalter {
		t.Errorf("Rolle = %q", rolle)
	}

	// **Und der Bestand wurde gar nicht gefragt** — die Antwort steht schon
	// am Profil.
	if ablage.gefragt != 0 {
		t.Errorf("der Bestand wurde %dmal gefragt", ablage.gefragt)
	}
}

// **Sonst zählt die Freigabe.**
func TestDieFreigabeEntscheidet(t *testing.T) {
	t.Parallel()

	ablage := neuesFreigabendoppel()
	ablage.abgelegt = []domain.Profilfreigabe{
		{ProfilID: 7, NutzerID: 2, Rolle: domain.RolleMitarbeiter},
	}

	rolle, fehler := service.NeueFreigabenverwaltung(ablage, nil).Rolle(
		context.Background(), 2, einProfil())
	if fehler != nil {
		t.Fatalf("%v", fehler)
	}

	if rolle != domain.RolleMitarbeiter {
		t.Errorf("Rolle = %q", rolle)
	}
}

// **Sonst die Sichtbarkeit — und die gibt höchstens `leser`.**
func TestDieSichtbarkeitGibtHoechstensLeser(t *testing.T) {
	t.Parallel()

	imBestand := einProfil(func(p *domain.Profil) {
		p.Sichtbarkeit = string(domain.SichtbarImBestand)
	})

	rolle, fehler := service.NeueFreigabenverwaltung(
		neuesFreigabendoppel(), nil).Rolle(context.Background(), 2, imBestand)
	if fehler != nil {
		t.Fatalf("%v", fehler)
	}

	if rolle != domain.RolleLeser {
		t.Errorf("Rolle = %q", rolle)
	}

	t.Run("und ein privates gibt gar keine", func(t *testing.T) {
		// **Der Beleg, dass die Sichtbarkeit den Unterschied macht** — ohne
		// ihn stünde der Fall darüber auch dann grün da, wenn jeder Fremde
		// `leser` bekäme (§10).
		rolle, fehler := service.NeueFreigabenverwaltung(
			neuesFreigabendoppel(), nil).Rolle(
			context.Background(), 2, einProfil())
		if fehler != nil {
			t.Fatalf("%v", fehler)
		}

		if rolle != "" {
			t.Errorf("Rolle = %q, erwartet keine", rolle)
		}
	})
}

// **`Verlangt` weist ab, wenn die Rolle nicht reicht — mit dem Satz, der sagt,
// was verlangt war.**
func TestVerlangtWeistAb(t *testing.T) {
	t.Parallel()

	ablage := neuesFreigabendoppel()
	ablage.abgelegt = []domain.Profilfreigabe{
		{ProfilID: 7, NutzerID: 2, Rolle: domain.RolleLeser},
	}

	verwaltung := service.NeueFreigabenverwaltung(ablage, nil)

	// **Ein Leser kommt nicht hinein** — das ist die Aussage von Vermerk 0070.
	fehler := verwaltung.Verlangt(
		context.Background(), 2, einProfil(), domain.RolleMitarbeiter)
	if !errors.Is(fehler, domain.ErrErlaubnis) {
		t.Fatalf("Fehler = %v, erwartet ErrErlaubnis", fehler)
	}

	if fehler.Error() == "" ||
		!errors.Is(fehler, domain.ErrErlaubnis) {
		t.Errorf("die Absage sagt nichts: %v", fehler)
	}

	// **Und für das, was er darf, kommt er durch** (§10).
	if fehler := verwaltung.Verlangt(
		context.Background(), 2, einProfil(), domain.RolleLeser); fehler != nil {
		t.Errorf("der Leser durfte nicht lesen: %v", fehler)
	}
}

// **Ein stolpernder Bestand wird weitergereicht.**
func TestEinStolpernderFreigabenbestand(t *testing.T) {
	t.Parallel()

	ablage := neuesFreigabendoppel()
	ablage.kaputt = errors.New("die Platte ist voll")

	verwaltung := service.NeueFreigabenverwaltung(ablage, nil)

	if _, fehler := verwaltung.Rolle(
		context.Background(), 2, einProfil()); fehler == nil {
		t.Error("Rolle: der Fehler kam nicht durch")
	}

	if fehler := verwaltung.Verlangt(
		context.Background(), 2, einProfil(), domain.RolleLeser); fehler == nil {
		t.Error("Verlangt: der Fehler kam nicht durch")
	}
}

// **Ein Profil lässt sich nicht an seinen eigenen Eigner freigeben.**
//
// **Er ist Verwalter ohne Zeile** — eine Zeile daneben wäre eine, die ihm
// etwas nehmen könnte, das er ohnehin hat.
func TestKeineFreigabeAnDenEigenenEigner(t *testing.T) {
	t.Parallel()

	ablage := &nutzerdoppel{
		nutzer: []domain.Nutzer{{ID: 1, Name: "Wer"}},
		profile: map[int64][]domain.Profil{
			1: {{ID: 7, NutzerID: 1, Name: "Meins"}},
		},
	}

	verwaltung := service.NeueFreigabenverwaltung(neuesFreigabendoppel(), ablage)

	fehler := verwaltung.FreigabeSetzen(context.Background(),
		domain.Profilfreigabe{
			ProfilID: 7, NutzerID: 1, Rolle: domain.RolleLeser,
		})
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("Fehler = %v, erwartet ErrValidation", fehler)
	}

	// **Und an einen anderen geht es** (§10).
	if fehler := verwaltung.FreigabeSetzen(context.Background(),
		domain.Profilfreigabe{
			ProfilID: 7, NutzerID: 2, Rolle: domain.RolleLeser,
		}); fehler != nil {
		t.Errorf("die Freigabe an einen anderen wurde abgewiesen: %v", fehler)
	}
}

// **Ein Profil, das es nicht gibt, ist ein Nein.**
func TestKeineFreigabeAnEinFehlendesProfil(t *testing.T) {
	t.Parallel()

	ablage := &nutzerdoppel{nutzer: []domain.Nutzer{{ID: 1, Name: "Wer"}}}

	verwaltung := service.NeueFreigabenverwaltung(neuesFreigabendoppel(), ablage)

	if fehler := verwaltung.FreigabeSetzen(context.Background(),
		domain.Profilfreigabe{
			ProfilID: 4711, NutzerID: 2, Rolle: domain.RolleLeser,
		}); !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("Fehler = %v, erwartet ErrNotFound", fehler)
	}
}

// **Und die drei durchreichenden Wege reichen ihr Nein durch.**
func TestDieFreigabenwegeReichenIhrNeinDurch(t *testing.T) {
	t.Parallel()

	kaputt := errors.New("die Platte ist voll")

	t.Run("Lesen", func(t *testing.T) {
		doppel := neuesFreigabendoppel()
		doppel.kaputt = kaputt

		if _, fehler := service.NeueFreigabenverwaltung(doppel, nil).Freigaben(
			context.Background(), 7); fehler == nil {
			t.Error("der Fehler kam nicht durch")
		}
	})

	t.Run("Löschen", func(t *testing.T) {
		doppel := neuesFreigabendoppel()
		doppel.kaputt = kaputt

		if fehler := service.NeueFreigabenverwaltung(doppel, nil).FreigabeLoeschen(
			context.Background(), 7, 2); fehler == nil {
			t.Error("der Fehler kam nicht durch")
		}
	})

	t.Run("Setzen, krumme Freigabe", func(t *testing.T) {
		if fehler := service.NeueFreigabenverwaltung(
			neuesFreigabendoppel(), nil).FreigabeSetzen(
			context.Background(), domain.Profilfreigabe{ProfilID: 7},
		); fehler == nil {
			t.Error("die krumme Freigabe kam durch")
		}
	})

	t.Run("Setzen, die Profile lassen sich nicht lesen", func(t *testing.T) {
		// **Der zweite Lesefehler ist ein eigener Weg**: `Nutzer` glückt,
		// `Profile` nicht — und wer ihn verschluckte, meldete „das Profil gibt
		// es nicht", wo der Bestand nur nicht antwortet (§8).
		ablage := &nutzerdoppel{
			nutzer:       []domain.Nutzer{{ID: 1, Name: "Wer"}},
			profilfehler: kaputt,
		}

		if fehler := service.NeueFreigabenverwaltung(
			neuesFreigabendoppel(), ablage).FreigabeSetzen(
			context.Background(), domain.Profilfreigabe{
				ProfilID: 7, NutzerID: 2, Rolle: domain.RolleLeser,
			}); !errors.Is(fehler, kaputt) {
			t.Errorf("Fehler = %v, erwartet %v", fehler, kaputt)
		}
	})

	t.Run("Setzen, stolpernde Nutzerablage", func(t *testing.T) {
		ablage := &nutzerdoppel{fehler: kaputt}

		if fehler := service.NeueFreigabenverwaltung(
			neuesFreigabendoppel(), ablage).FreigabeSetzen(
			context.Background(), domain.Profilfreigabe{
				ProfilID: 7, NutzerID: 2, Rolle: domain.RolleLeser,
			}); fehler == nil {
			t.Error("der Fehler kam nicht durch")
		}
	})
}
