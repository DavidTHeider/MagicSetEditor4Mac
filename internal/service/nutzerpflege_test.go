package service_test

// **Nutzer und Profile ändern und löschen** — Werkstück 5 von Phase 4.12
// (#1351).

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein unbrauchbarer Name erreicht den Bestand nicht.**
//
// **Zweimal geprüft und nicht einmal**: die Domäne sagt, was ein Name ist; der
// `CHECK` in der Wanderung ist die zweite Reihe, und er nennt **kein Feld**.
// Wer nur ihn hätte, bekäme „constraint failed" statt „ein Name trägt keine
// Leerzeichen am Rand" (§8).
func TestEinUnbrauchbarerNameErreichtDenBestandNichtImDienst(t *testing.T) {
	t.Parallel()

	doppel := &nutzerdoppel{}
	verwaltung := service.NeueNutzerverwaltung(doppel)

	if _, fehler := verwaltung.NutzerAendern(t.Context(),
		domain.Nutzer{ID: 1, Name: " Wer "}); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Errorf("Nutzer: %v", fehler)
	}

	if _, fehler := verwaltung.ProfilAendern(t.Context(),
		domain.Profil{ID: 1, NutzerID: 1, Name: ""}); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Errorf("Profil: %v", fehler)
	}

	if doppel.geaendert.ID != 0 || doppel.profilgeandert.ID != 0 {
		t.Error("etwas erreichte den Bestand")
	}
}

// **Ein gültiger Name kommt an.**
func TestEinGueltigerNameKommtAn(t *testing.T) {
	t.Parallel()

	doppel := &nutzerdoppel{}
	verwaltung := service.NeueNutzerverwaltung(doppel)

	if _, fehler := verwaltung.NutzerAendern(t.Context(),
		domain.Nutzer{ID: 1, Name: "Wer"}); fehler != nil {
		t.Fatalf("ändern: %v", fehler)
	}

	if doppel.geaendert.Name != "Wer" {
		t.Errorf("angekommen ist %+v", doppel.geaendert)
	}

	if _, fehler := verwaltung.ProfilAendern(t.Context(),
		domain.Profil{ID: 1, NutzerID: 1, Name: "Turnier"}); fehler != nil {
		t.Fatalf("ändern: %v", fehler)
	}

	if doppel.profilgeandert.Name != "Turnier" {
		t.Errorf("angekommen ist %+v", doppel.profilgeandert)
	}
}

// **Ein Nutzer lässt sich löschen.**
func TestEinNutzerLaesstSichImDienstLoeschen(t *testing.T) {
	t.Parallel()

	doppel := &nutzerdoppel{}

	if fehler := service.NeueNutzerverwaltung(doppel).NutzerLoeschen(
		t.Context(), 7, "Wer"); fehler != nil {
		t.Fatalf("löschen: %v", fehler)
	}

	if doppel.geloescht != 7 {
		t.Errorf("gelöscht wurde %d", doppel.geloescht)
	}
}

// **Wer das gezeigte Profil löscht, steht danach in einem anderen.**
//
// # Warum der Zeiger vorher wandert
//
// **Sonst zeigte er auf ein Profil, das es nicht mehr gibt.**
// `storage.Nutzerablage.Profilzeiger` fiele dann auf das **erste** zurück —
// ohne dass es jemand veranlasst hätte, und ohne dass es jemand sähe.
func TestWerDasGezeigteProfilLoeschtStehtDanachInEinemAnderen(t *testing.T) {
	t.Parallel()

	doppel := &nutzerdoppel{
		nutzer: []domain.Nutzer{{ID: 1, Name: "Wer"}},
		profile: map[int64][]domain.Profil{
			1: {{ID: 1, NutzerID: 1, Name: "Erstes"}, {ID: 2, NutzerID: 1, Name: "Zweites"}},
		},
		zeiger: 2,
	}

	if fehler := service.NeueNutzerverwaltung(doppel).ProfilLoeschen(
		t.Context(), 2, "Zweites"); fehler != nil {
		t.Fatalf("löschen: %v", fehler)
	}

	if doppel.zeiger != 1 {
		t.Errorf("der Zeiger steht auf %d, erwartet war 1", doppel.zeiger)
	}

	if doppel.profilgeloescht != 2 {
		t.Errorf("gelöscht wurde %d", doppel.profilgeloescht)
	}
}

// **Ein anderes Profil zu löschen lässt den Zeiger stehen.**
//
// **Und schreibt ihn nicht**: ein Zeiger, der schon richtig steht, noch einmal
// gesetzt, wäre ein Bestandsschreiben ohne Wirkung.
func TestEinAnderesProfilZuLoeschenLaesstDenZeigerStehen(t *testing.T) {
	t.Parallel()

	doppel := &nutzerdoppel{
		nutzer: []domain.Nutzer{{ID: 1, Name: "Wer"}},
		profile: map[int64][]domain.Profil{
			1: {{ID: 1, NutzerID: 1, Name: "Erstes"}, {ID: 2, NutzerID: 1, Name: "Zweites"}},
		},
		zeiger: 1,
	}

	if fehler := service.NeueNutzerverwaltung(doppel).ProfilLoeschen(
		t.Context(), 2, "Zweites"); fehler != nil {
		t.Fatalf("löschen: %v", fehler)
	}

	if doppel.gesetzt != 0 {
		t.Errorf("%d Schreibvorgänge am Zeiger", doppel.gesetzt)
	}
}

// **Gibt es kein anderes Profil, sagt es das — und löscht nicht.**
func TestOhneAnderesProfilWirdNichtGeloescht(t *testing.T) {
	t.Parallel()

	doppel := &nutzerdoppel{
		nutzer:  []domain.Nutzer{{ID: 1, Name: "Wer"}},
		profile: map[int64][]domain.Profil{1: {{ID: 1, NutzerID: 1, Name: "Erstes"}}},
		zeiger:  1,
	}

	fehler := service.NeueNutzerverwaltung(doppel).ProfilLoeschen(
		t.Context(), 1, "Erstes")

	if !errors.Is(fehler, domain.ErrConflict) {
		t.Errorf("Fehler %v, erwartet war ErrConflict", fehler)
	}

	if doppel.profilgeloescht != 0 {
		t.Error("es wurde trotzdem gelöscht")
	}
}

// **Jeder Fehlerweg der Pflege reicht durch.**
func TestJederFehlerwegDerPflegeReichtDurch(t *testing.T) {
	t.Parallel()

	welt := func() *nutzerdoppel {
		return &nutzerdoppel{
			nutzer: []domain.Nutzer{{ID: 1, Name: "Wer"}},
			profile: map[int64][]domain.Profil{
				1: {{ID: 1, NutzerID: 1, Name: "Erstes"}, {ID: 2, NutzerID: 1, Name: "Zweites"}},
			},
			zeiger: 2,
		}
	}

	faelle := map[string]func(*service.Nutzerverwaltung) error{
		"Nutzer ändern": func(v *service.Nutzerverwaltung) error {
			_, err := v.NutzerAendern(t.Context(), domain.Nutzer{ID: 1, Name: "X"})

			return err
		},
		"Profil ändern": func(v *service.Nutzerverwaltung) error {
			_, err := v.ProfilAendern(t.Context(), domain.Profil{ID: 1, NutzerID: 1, Name: "X"})

			return err
		},
		"Nutzer löschen": func(v *service.Nutzerverwaltung) error {
			return v.NutzerLoeschen(t.Context(), 1, "Wer")
		},
		"Profil löschen": func(v *service.Nutzerverwaltung) error {
			return v.ProfilLoeschen(t.Context(), 2, "Zweites")
		},
	}

	for name, ruf := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			doppel := welt()
			doppel.pflegefehler = errors.New("die Platte schweigt")

			if fehler := ruf(service.NeueNutzerverwaltung(doppel)); fehler == nil {
				t.Error("der Fehler ging verloren")
			}
		})
	}
}

// **Und die beiden Wege über den Zeiger ebenso.**
func TestDieZeigerwegeDerPflegeReichenIhreFehlerDurch(t *testing.T) {
	t.Parallel()

	t.Run("der Zeiger lässt sich nicht lesen", func(t *testing.T) {
		t.Parallel()

		doppel := &nutzerdoppel{zeigerfehler: errors.New("kein Zeiger")}

		if fehler := service.NeueNutzerverwaltung(doppel).ProfilLoeschen(
			t.Context(), 1, "Erstes"); fehler == nil {
			t.Error("der Fehler ging verloren")
		}
	})

	// **Der Zeigerfehler ist hierher gewandert** — Befund 1 des Zweitblicks
	// über die Phasen 4.0 bis 4.13 (2026-09-07).
	//
	// Der Dienst setzte den Zeiger **vor** dem Löschen und hatte dafür einen
	// eigenen Fehlerweg. Genau das war der Befund: wies die Ablage danach ab,
	// blieb der Zeiger bewegt. Jetzt setzt sie ihn in derselben Transaktion,
	// in der sie löscht — und ihr Fehler ist **der** Fehler.
	t.Run("das Löschen samt Zeiger scheitert", func(t *testing.T) {
		t.Parallel()

		doppel := &nutzerdoppel{
			nutzer: []domain.Nutzer{{ID: 1, Name: "Wer"}},
			profile: map[int64][]domain.Profil{
				1: {{ID: 1, NutzerID: 1, Name: "Erstes"}, {ID: 2, NutzerID: 1, Name: "Zweites"}},
			},
			zeiger:       2,
			pflegefehler: errors.New("kein Schreiben"),
		}

		if fehler := service.NeueNutzerverwaltung(doppel).ProfilLoeschen(
			t.Context(), 2, "Zweites"); fehler == nil {
			t.Error("der Fehler ging verloren")
		}

		// **Und der Zeiger steht, wo er stand.**
		if doppel.zeiger != 2 {
			t.Errorf("der Zeiger steht auf %d — ein abgewiesenes Löschen hat "+
				"ihn bewegt", doppel.zeiger)
		}
	})

	t.Run("die Übersicht klagt", func(t *testing.T) {
		t.Parallel()

		doppel := &nutzerdoppel{
			zeiger: 2,
			fehler: errors.New("die Platte schweigt"),
		}

		if fehler := service.NeueNutzerverwaltung(doppel).ProfilLoeschen(
			t.Context(), 2, "Zweites"); fehler == nil {
			t.Error("der Fehler ging verloren")
		}
	})
}
