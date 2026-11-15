package service_test

// **Der wirksame Akteur aus dem Kontext** — Werkstück O1 von Phase 5.2
// (#1425).
//
// # Was hier geprüft wird
//
// Aus einem Kontext — Profil, Zeiger, Token, Fläche — und dem Bestand —
// Eigner, Freigaben — entsteht **ein** Akteur, und jede Fachoperation fragt
// denselben. Die Fälle sind die der Abnahme: Person A im Profil von B bleibt
// A; ein Leser schreibt nicht, auch nicht auf die Platte; ein Token erweitert
// nie.

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// bestandFuerAkteure legt drei Nutzer und ein Profil an: das Profil 7 gehört
// Nutzer 1; Nutzer 2 hat darauf eine Freigabe, Nutzer 3 keine.
func bestandFuerAkteure(rolle domain.Rolle) (*nutzerdoppel, *freigabendoppel) {
	nutzer := &nutzerdoppel{
		nutzer: []domain.Nutzer{{ID: 1, Name: "Eigner"}, {ID: 2, Name: "Gast"}, {ID: 3, Name: "Fremder"}},
		profile: map[int64][]domain.Profil{
			1: {{ID: 7, NutzerID: 1, Name: "Turnier", Sichtbarkeit: "privat"}},
		},
		nutzerzeiger:  1,
		arbeitsprofil: 7,
	}

	freigaben := neuesFreigabendoppel()
	if rolle != "" {
		freigaben.abgelegt = []domain.Profilfreigabe{{ProfilID: 7, NutzerID: 2, Rolle: rolle}}
	}

	return nutzer, freigaben
}

func autorisierung(nutzer *nutzerdoppel, freigaben *freigabendoppel) *service.Autorisierung {
	return service.NeueAutorisierung(nutzer,
		service.NeueFreigabenverwaltung(freigaben, nutzer))
}

// Arbeitsprofil gibt das gesetzte Profil dieses Aufrufs — oder den gesetzten
// Fehler.
func (d *nutzerdoppel) Arbeitsprofil(context.Context) (int64, error) {
	if d.arbeitsprofilfehler != nil {
		return 0, d.arbeitsprofilfehler
	}

	return d.arbeitsprofil, nil
}

// **Der Eigner in seinem eigenen Profil ist Verwalter** — ohne Zeile im
// Bestand, und am Fenster ohne Token.
func TestDerEignerIstVerwalterSeinesProfils(t *testing.T) {
	t.Parallel()

	nutzer, freigaben := bestandFuerAkteure("")
	wer, fehler := autorisierung(nutzer, freigaben).Akteur(context.Background())
	if fehler != nil {
		t.Fatalf("Akteur: %v", fehler)
	}

	erwartet := domain.Akteur{
		NutzerID: 1, ProfilID: 7, EignerID: 1, Rolle: domain.RolleVerwalter,
		Modus: domain.ModusFenster,
	}
	if wer != erwartet {
		t.Fatalf("Akteur %+v, erwartet %+v", wer, erwartet)
	}

	if schreibfehler := autorisierung(nutzer, freigaben).Pruefe(context.Background(), domain.WirkungSchreiben); schreibfehler != nil {
		t.Errorf("der Eigner schreibt nicht: %v", schreibfehler)
	}

	// **Auch der Verwalter braucht für das Draußen einen Nachweis.**
	fehler = autorisierung(nutzer, freigaben).Pruefe(context.Background(), domain.WirkungExternSchreiben)
	if !errors.Is(fehler, domain.ErrErlaubnis) {
		t.Errorf("ohne Nachweis nach draußen: %v, erwartet ErrErlaubnis", fehler)
	}
}

// **Person A im Profil von B bleibt A** — und ihre Rolle dort ist die
// Freigabe, nicht die des Eigners.
func TestEinGastBleibtErSelbst(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		rolle     domain.Rolle
		nurLesend bool
		schreibt  error
	}{
		"Leser liest nur":         {domain.RolleLeser, true, domain.ErrNurLesend},
		"Mitarbeiter schreibt":    {domain.RolleMitarbeiter, false, nil},
		"Verwalter schreibt auch": {domain.RolleVerwalter, false, nil},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			nutzer, freigaben := bestandFuerAkteure(fall.rolle)
			nutzer.nutzerzeiger = 2

			wer, fehler := autorisierung(nutzer, freigaben).Akteur(context.Background())
			if fehler != nil {
				t.Fatalf("Akteur: %v", fehler)
			}

			if wer.NutzerID != 2 || wer.EignerID != 1 || wer.Rolle != fall.rolle || wer.NurLesend != fall.nurLesend {
				t.Fatalf("Akteur %+v", wer)
			}

			fehler = autorisierung(nutzer, freigaben).Pruefe(context.Background(), domain.WirkungSchreiben)
			if (fall.schreibt == nil) != (fehler == nil) || (fehler != nil && !errors.Is(fehler, fall.schreibt)) {
				t.Errorf("Schreiben: %v, erwartet %v", fehler, fall.schreibt)
			}
		})
	}
}

// **Ohne Rolle geht nichts** — auch das Lesen nicht: wer keine Freigabe hat
// und das Profil privat ist, hat dort nichts zu suchen.
func TestOhneRolleNichts(t *testing.T) {
	t.Parallel()

	nutzer, freigaben := bestandFuerAkteure("")
	nutzer.nutzerzeiger = 3

	fehler := autorisierung(nutzer, freigaben).Pruefe(context.Background(), domain.WirkungLesen)
	if !errors.Is(fehler, domain.ErrErlaubnis) {
		t.Fatalf("Lesen ohne Rolle: %v, erwartet ErrErlaubnis", fehler)
	}
}

// **Token und Fläche kommen aus dem Kontext** — und das Token engt ein.
func TestTokenUndFlaecheKommenAusDemKontext(t *testing.T) {
	t.Parallel()

	nutzer, freigaben := bestandFuerAkteure("")
	ctx := service.MitTokenumfang(service.MitDienstmodus(context.Background(), domain.ModusDienst), domain.ScopeLesen)

	wer, fehler := autorisierung(nutzer, freigaben).Akteur(ctx)
	if fehler != nil {
		t.Fatalf("Akteur: %v", fehler)
	}

	if wer.Token != domain.ScopeLesen || wer.Modus != domain.ModusDienst {
		t.Fatalf("Akteur %+v", wer)
	}

	fehler = autorisierung(nutzer, freigaben).Pruefe(ctx, domain.WirkungSchreiben)
	if !errors.Is(fehler, domain.ErrErlaubnis) {
		t.Errorf("der Verwalter mit Lesetoken schreibt: %v", fehler)
	}
}

// **Die Revision kommt aus der Ablage** — sie zählt Zeiger und Freigaben
// mit, damit ein älterer Akteur als älter erkennbar ist; und kommt sie nicht,
// gibt es keinen Akteur.
func TestDieRevisionKommtAusDerAblage(t *testing.T) {
	t.Parallel()

	nutzer, freigaben := bestandFuerAkteure("")
	nutzer.revision = 42

	wer, fehler := autorisierung(nutzer, freigaben).Akteur(context.Background())
	if fehler != nil || wer.Revision != 42 {
		t.Fatalf("Akteur %+v, %v — Revision 42 erwartet", wer, fehler)
	}

	kaputt := errors.New("kaputt")
	nutzer.revisionsfehler = kaputt

	if _, fehler := autorisierung(nutzer, freigaben).Akteur(context.Background()); !errors.Is(fehler, kaputt) {
		t.Errorf("ohne Revision: %v, erwartet den Fehler der Ablage", fehler)
	}
}

// Revision gibt den gesetzten Stand — oder den gesetzten Fehler.
func (d *nutzerdoppel) Revision(context.Context) (int64, error) {
	if d.revisionsfehler != nil {
		return 0, d.revisionsfehler
	}

	return d.revision, nil
}

// **Jeder Fehlerweg wird betreten** (§3): kein Arbeitsprofil, kein Zeiger,
// unbekanntes Profil, kaputte Freigaben.
func TestDieFehlerwegeDesAkteurs(t *testing.T) {
	t.Parallel()

	kaputt := errors.New("kaputt")

	faelle := map[string]struct {
		ctx      context.Context
		richten  func(*nutzerdoppel, *freigabendoppel)
		erwartet error
	}{
		"kein Arbeitsprofil": {context.Background(), func(n *nutzerdoppel, _ *freigabendoppel) { n.arbeitsprofilfehler = kaputt }, kaputt},
		"kein Zeiger":        {context.Background(), func(n *nutzerdoppel, _ *freigabendoppel) { n.nutzerzeigerfehler = kaputt }, kaputt},
		"niemand arbeitet":   {context.Background(), func(n *nutzerdoppel, _ *freigabendoppel) { n.nutzerzeiger = 0 }, domain.ErrNotFound},
		"unbekanntes Profil": {context.Background(), func(n *nutzerdoppel, _ *freigabendoppel) { n.arbeitsprofil = 99 }, domain.ErrNotFound},
		"kaputte Freigaben": {context.Background(), func(n *nutzerdoppel, f *freigabendoppel) {
			n.nutzerzeiger = 2
			f.kaputt = kaputt
		}, kaputt},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			nutzer, freigaben := bestandFuerAkteure("")
			fall.richten(nutzer, freigaben)

			_, fehler := autorisierung(nutzer, freigaben).Akteur(fall.ctx)
			if !errors.Is(fehler, fall.erwartet) {
				t.Fatalf("%s: %v, erwartet %v", name, fehler, fall.erwartet)
			}

			if fehler := autorisierung(nutzer, freigaben).Pruefe(fall.ctx, domain.WirkungLesen); !errors.Is(fehler, fall.erwartet) {
				t.Fatalf("%s über Pruefe: %v, erwartet %v", name, fehler, fall.erwartet)
			}
		})
	}
}

// **Ohne Angabe ist die Fläche das Fenster** — und ein Kontext ohne Token
// trägt keines.
func TestOhneAngabeGiltDasFenster(t *testing.T) {
	t.Parallel()

	if service.DienstmodusAus(context.Background()) != domain.ModusFenster {
		t.Error("ohne Angabe sollte das Fenster gelten")
	}

	if service.TokenumfangAus(context.Background()) != "" {
		t.Error("ohne Angabe sollte kein Token gelten")
	}
}
