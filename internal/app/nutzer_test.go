// **Nutzer und Profile über die Fassade** — Werkstück 1 von Phase 4.12
// (#1351).
package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Die Übersicht kommt mit Profilen an.**
func TestDieNutzeruebersichtKommtMitProfilenAn(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitNutzern(&nutzerfuehrer{zeilen: []service.Nutzerzeile{{
		Nutzer:  domain.Nutzer{ID: 1, Name: "Eine"},
		Profile: []domain.Profil{{ID: 10, NutzerID: 1, Name: "Mein Profil"}},
	}}})

	antwort := anwendung.Nutzer(t.Context())
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if len(antwort.Nutzer) != 1 || len(antwort.Nutzer[0].Profile) != 1 {
		t.Fatalf("%+v", antwort.Nutzer)
	}

	// **Kennungen reisen als Text** — eine JavaScript-Zahl verlöre oberhalb
	// von 2⁵³ die letzten Stellen.
	if antwort.Nutzer[0].ID != "1" || antwort.Nutzer[0].Profile[0].ID != "10" {
		t.Errorf("Kennungen: %+v", antwort.Nutzer[0])
	}
}

// **Eine leere Übersicht ist eine leere Liste, nicht `nil`** (§3).
//
// `json:",omitempty"` ließe sie sonst aus der Antwort verschwinden, und die
// Seite bliebe leer statt „noch niemand".
func TestEineLeereNutzeruebersichtIstEineListe(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitNutzern(&nutzerfuehrer{})

	if antwort := anwendung.Nutzer(t.Context()); antwort.Nutzer == nil {
		t.Error("die Liste ist nil")
	}
}

// **Ein Fehler der Anwendung kommt als Umschlag an** (§8).
func TestEinNutzerfehlerKommtAlsUmschlagAn(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitNutzern(&nutzerfuehrer{fehler: domain.NewFieldError(
		"name", domain.SchluesselPflichtfeld, "leer")})

	if antwort := anwendung.Nutzer(t.Context()); antwort.Fehler == nil {
		t.Error("der Fehler ging verloren")
	}

	if antwort := anwendung.NutzerAnlegen(t.Context(), "", ""); antwort.Fehler == nil {
		t.Error("der Fehler beim Anlegen ging verloren")
	}
}

// **Ein neuer Nutzer kommt mit einer leeren Profilliste zurück**, nicht mit
// `nil` — dieselbe Zusage wie oben, eine Ebene tiefer.
func TestEinNeuerNutzerHatEineLeereProfilliste(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitNutzern(&nutzerfuehrer{})

	antwort := anwendung.NutzerAnlegen(t.Context(), "Zweite", "")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if antwort.Nutzer == nil || antwort.Nutzer.Profile == nil {
		t.Errorf("%+v", antwort.Nutzer)
	}
}

// **Eine unlesbare Nutzerkennung wird am Feld gemeldet.**
func TestEineUnlesbareNutzerkennungWirdGemeldet(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitNutzern(&nutzerfuehrer{})

	antwort := anwendung.ProfilAnlegen(t.Context(), "keine Zahl", "Arbeit")
	if antwort.Fehler == nil {
		t.Fatal("die krumme Kennung kam durch")
	}

	if antwort.Fehler.Feld != "nutzerID" {
		t.Errorf("Feld %q", antwort.Fehler.Feld)
	}
}

// **Ein Profil kommt mit seinen Kennungen an.**
func TestEinNeuesProfilKommtMitKennungenAn(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitNutzern(&nutzerfuehrer{})

	antwort := anwendung.ProfilAnlegen(t.Context(), "7", "Turnier")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if antwort.Profil.NutzerID != "7" || antwort.Profil.Name != "Turnier" {
		t.Errorf("%+v", antwort.Profil)
	}
}

// **Und ein Fehler beim Anlegen eines Profils ebenso.**
func TestEinProfilfehlerKommtAlsUmschlagAn(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitNutzern(&nutzerfuehrer{fehler: errors.New("die Platte schweigt")})

	if antwort := anwendung.ProfilAnlegen(t.Context(), "7", "Turnier"); antwort.Fehler == nil {
		t.Error("der Fehler ging verloren")
	}
}

// nutzerfuehrer ist ein Nutzerdienst im Arbeitsspeicher.
type nutzerfuehrer struct {
	zeilen []service.Nutzerzeile
	fehler error

	// ueber ist die Antwort auf „wer über wen schreiben darf" (ID-03).
	ueber error
}

func (f *nutzerfuehrer) UeberNutzer(context.Context, int64) error { return f.ueber }

func (f *nutzerfuehrer) UeberProfil(context.Context, int64) error { return f.ueber }

func (f *nutzerfuehrer) Uebersicht(context.Context) ([]service.Nutzerzeile, error) {
	return f.zeilen, f.fehler
}

func (f *nutzerfuehrer) NutzerAnlegen(_ context.Context, neu domain.Nutzer) (
	service.Nutzerzeile, error,
) {
	if f.fehler != nil {
		return service.Nutzerzeile{}, f.fehler
	}

	neu.ID = 2

	// **Mit einem ersten Profil**, wie der Speicher es anlegt. Ein Doppel ohne
	// eines wäre großzügiger als der Betrieb (§3).
	return service.Nutzerzeile{
		Nutzer:  neu,
		Profile: []domain.Profil{{ID: 200, NutzerID: 2, Name: "Standard"}},
	}, nil
}

func (f *nutzerfuehrer) ProfilAnlegen(_ context.Context, neu domain.Profil) (
	domain.Profil, error,
) {
	if f.fehler != nil {
		return domain.Profil{}, f.fehler
	}

	neu.ID = 20

	return neu, nil
}

// **Ändern und Löschen** — Werkstück 5 (#1351).
func (f *nutzerfuehrer) NutzerAendern(_ context.Context, geaendert domain.Nutzer) (
	domain.Nutzer, error,
) {
	if f.fehler != nil {
		return domain.Nutzer{}, f.fehler
	}

	return geaendert, nil
}

func (f *nutzerfuehrer) ProfilAendern(_ context.Context, geaendert domain.Profil) (
	domain.Profil, error,
) {
	if f.fehler != nil {
		return domain.Profil{}, f.fehler
	}

	return geaendert, nil
}

func (f *nutzerfuehrer) NutzerLoeschen(context.Context, int64, string,
	...func([]domain.Profil, [][]domain.Betroffenes) error,
) error {
	return f.fehler
}

func (f *nutzerfuehrer) ProfilLoeschen(context.Context, int64, string,
	...func([]domain.Profil, [][]domain.Betroffenes) error,
) error {
	return f.fehler
}
