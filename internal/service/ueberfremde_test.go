package service_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Über einen fremden Menschen schreibt nur, wer ihn werden kann** — ID-03
// (#1702).
//
// Die Matrix sagt es für jede Fähigkeit am Träger Nutzer: „wer arbeitet,
// darf über sich schreiben; fremde Nutzer verlangen deren zweiten Faktor"
// (0074). **Werden kann ihn, wen der Wechsel zu ihm nicht nach einem Faktor
// fragte** — ohne Faktor, oder mit einem, der beim Wechsel abgeschaltet ist.
func TestUeberEinenMenschenSchreibtNurWerIhnWerdenKann(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name      string
		arbeitet  int64
		ziel      int64
		verlangt  bool
		abgelehnt bool
	}{
		{"über sich selbst, auch mit Faktor", 1, 1, true, false},
		{"über einen Fremden ohne Faktor", 1, 2, false, false},
		{"über einen Fremden mit Faktor", 1, 2, true, true},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			pruefer := &pruefdoppel{verlangt: map[int64]bool{fall.ziel: fall.verlangt}}
			verwaltung := service.NeueNutzerverwaltung(zweiMenschen(fall.arbeitet)).MitFaktor(pruefer)

			fehler := verwaltung.UeberNutzer(t.Context(), fall.ziel)

			if fall.abgelehnt != errors.Is(fehler, domain.ErrFremderNutzer) {
				t.Fatalf("Antwort %v, abgelehnt erwartet: %v", fehler, fall.abgelehnt)
			}

			// **Und die Absage ist eine Erlaubnisfrage**, keine Validierung:
			// am Fenster steht sonst das Formular noch einmal da.
			if fall.abgelehnt && !errors.Is(fehler, domain.ErrErlaubnis) {
				t.Errorf("die Absage ist keine Erlaubnisfrage: %v", fehler)
			}
		})
	}
}

// **Ohne Prüfer fragt nichts** — dieselbe Lage wie beim Wechsel
// (`TestOhnePrueferWirdNichtGefragt`): eine Verwaltung ohne Faktor kennt
// keinen, also kann niemand geschützt sein.
func TestOhnePrueferIstNiemandGeschuetzt(t *testing.T) {
	t.Parallel()

	if fehler := service.NeueNutzerverwaltung(zweiMenschen(1)).
		UeberNutzer(t.Context(), 2); fehler != nil {
		t.Errorf("abgewiesen: %v", fehler)
	}
}

// **Wer nicht sagen kann, wer arbeitet oder was der Wechsel verlangte, lässt
// nicht durch** — ein Lesefehler ist kein Ja.
func TestEinLesefehlerIstKeinJa(t *testing.T) {
	t.Parallel()

	kaputt := errors.New("der Bestand antwortet nicht")

	ablage := zweiMenschen(1)
	ablage.nutzerzeigerfehler = kaputt

	if fehler := service.NeueNutzerverwaltung(ablage).MitFaktor(&pruefdoppel{}).
		UeberNutzer(t.Context(), 2); !errors.Is(fehler, kaputt) {
		t.Errorf("wer arbeitet, unlesbar: %v", fehler)
	}

	if fehler := service.NeueNutzerverwaltung(zweiMenschen(1)).
		MitFaktor(&pruefdoppel{verlangtfehler: kaputt}).
		UeberNutzer(t.Context(), 2); !errors.Is(fehler, kaputt) {
		t.Errorf("was der Wechsel verlangte, unlesbar: %v", fehler)
	}
}

// zweiMenschen ist ein Bestand mit zwei Nutzern und je einem Profil; es
// arbeitet `arbeitet`.
func zweiMenschen(arbeitet int64) *nutzerdoppel {
	return &nutzerdoppel{
		nutzer: []domain.Nutzer{{ID: 1, Name: "Wer"}, {ID: 2, Name: "Bob"}},
		profile: map[int64][]domain.Profil{
			1: {{ID: 1, NutzerID: 1, Name: "Eigenes"}},
			2: {{ID: 2, NutzerID: 2, Name: "Bobs"}},
		},
		zeiger:       arbeitet,
		nutzerzeiger: arbeitet,
	}
}

// **Über ein Profil schreibt, wer über seinen Eigner schreiben darf — und im
// gezeigten Profil nur, wessen Rolle bis zum Schreiben reicht** (ID-03, der
// native Befund F01). Das zweite ist die Zusage des Banners: wer dort „nur
// lesen" sieht, ändert auch die Angaben des Profils nicht.
func TestUeberEinProfilSchreibtWerUeberSeinenEignerDarf(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name       string
		profil     int64
		gezeigt    int64
		rolle      domain.Rolle
		geschuetzt bool
		erwartet   error
	}{
		{"das eigene, gezeigt", 1, 1, "", false, nil},
		{"ein fremdes, ungeschützt, von außen", 2, 1, "", false, nil},
		{"ein fremdes, geschützt", 2, 1, "", true, domain.ErrFremderNutzer},
		{"ein fremdes, gezeigt, ohne Rolle", 2, 2, "", false, domain.ErrNurLesend},
		{"ein fremdes, gezeigt, als Leser", 2, 2, domain.RolleLeser, false, domain.ErrNurLesend},
		{"ein fremdes, gezeigt, als Mitarbeiter", 2, 2, domain.RolleMitarbeiter, false, nil},
		{"eines, das es nicht gibt", 9, 1, "", false, domain.ErrNotFound},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			ablage := zweiMenschen(1)
			ablage.zeiger = fall.gezeigt

			freigaben := neuesFreigabendoppel()
			if fall.rolle != "" {
				freigaben.abgelegt = []domain.Profilfreigabe{{ProfilID: 2, NutzerID: 1, Rolle: fall.rolle}}
			}

			verwaltung := service.NeueNutzerverwaltung(ablage).
				MitFaktor(&pruefdoppel{verlangt: map[int64]bool{2: fall.geschuetzt}}).
				MitFreigaben(service.NeueFreigabenverwaltung(freigaben, ablage))

			fehler := verwaltung.UeberProfil(t.Context(), fall.profil)

			if fall.erwartet == nil && fehler != nil || fall.erwartet != nil && !errors.Is(fehler, fall.erwartet) {
				t.Errorf("Antwort %v, erwartet %v", fehler, fall.erwartet)
			}
		})
	}
}

// **Ein unlesbarer Zeiger oder unlesbare Freigaben sind kein Ja** — auch
// nicht am gezeigten Profil.
func TestUeberEinProfilIstEinLesefehlerKeinJa(t *testing.T) {
	t.Parallel()

	kaputt := errors.New("der Bestand antwortet nicht")

	ohneZeiger := zweiMenschen(1)
	ohneZeiger.zeigerfehler = kaputt

	if fehler := service.NeueNutzerverwaltung(ohneZeiger).
		UeberProfil(t.Context(), 1); !errors.Is(fehler, kaputt) {
		t.Errorf("der Zeiger unlesbar: %v", fehler)
	}

	gezeigt := zweiMenschen(1)
	gezeigt.zeiger = 2

	freigaben := neuesFreigabendoppel()
	freigaben.kaputt = kaputt

	if fehler := service.NeueNutzerverwaltung(gezeigt).
		MitFreigaben(service.NeueFreigabenverwaltung(freigaben, gezeigt)).
		UeberProfil(t.Context(), 2); !errors.Is(fehler, kaputt) {
		t.Errorf("die Freigaben unlesbar: %v", fehler)
	}
}
