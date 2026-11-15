// **Freigaben über die Fassade** — Werkstück 11 von Phase 4.12
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

// freigabenfuehrer ist eine Verwaltung im Arbeitsspeicher.
type freigabenfuehrer struct {
	abgelegt []domain.Profilfreigabe

	// rolle ist, was [freigabenfuehrer.Rolle] gibt — leer heißt: keine.
	rolle domain.Rolle

	// gesehen ist, was das letzte Setzen zu sehen bekam.
	gesehen domain.Profilfreigabe

	// geloescht sind die Kennungen des letzten Löschens.
	geloescht [2]int64

	fehler error

	// datenfehler klagt **nur** an den drei Datenwegen — nicht an der Rolle.
	//
	// **Zwei Felder und nicht eines**: mit einem gemeinsamen scheiterte schon
	// die Rollenfrage, und die drei Zweige dahinter blieben unbetreten. Das
	// ist genau die Gestalt aus §10 — ein Lauf, der scheitert, bevor er die
	// Stelle erreicht, die er meint.
	datenfehler error
}

func (f *freigabenfuehrer) Freigaben(context.Context, int64) (
	[]domain.Profilfreigabe, error,
) {
	if f.datenfehler != nil {
		return nil, f.datenfehler
	}

	return f.abgelegt, f.fehler
}

func (f *freigabenfuehrer) Rolle(
	context.Context, int64, domain.Profil,
) (domain.Rolle, error) {
	return f.rolle, f.fehler
}

func (f *freigabenfuehrer) Verlangt(
	_ context.Context, _ int64, _ domain.Profil, verlangt domain.Rolle,
) error {
	if f.fehler != nil {
		return f.fehler
	}

	if !f.rolle.Reicht(verlangt) {
		return domain.Freigabefehler(verlangt)
	}

	return nil
}

func (f *freigabenfuehrer) FreigabeSetzen(
	_ context.Context, freigabe domain.Profilfreigabe,
) error {
	f.gesehen = freigabe

	if f.datenfehler != nil {
		return f.datenfehler
	}

	return f.fehler
}

func (f *freigabenfuehrer) FreigabeLoeschen(
	_ context.Context, profilID, nutzerID int64,
) error {
	f.geloescht = [2]int64{profilID, nutzerID}

	if f.datenfehler != nil {
		return f.datenfehler
	}

	return f.fehler
}

// mitFreigaben baut eine Anwendung über einer Verwaltung.
//
// **Mit einer Nutzernaht und einem Profilwechsel dazu**: die Fassade fragt
// beide — wer arbeitet, und welches Profil gemeint ist.
func mitFreigaben(fuehrer *freigabenfuehrer) *app.App {
	anwendung := app.NeueApp("")

	nutzer := &nutzerfuehrer{zeilen: []service.Nutzerzeile{{
		Nutzer: domain.Nutzer{ID: 1, Name: "Ich"},
		Profile: []domain.Profil{
			{ID: 10, NutzerID: 1, Name: "Meins"},
		},
	}, {
		Nutzer: domain.Nutzer{ID: 2, Name: "Zweiter"},
		Profile: []domain.Profil{
			{ID: 20, NutzerID: 2, Name: "Fremd"},
		},
	}}}

	anwendung.MitNutzern(nutzer)
	anwendung.MitProfilwechsel(&wechseldoppel{
		gezeigt: domain.Profil{ID: 10, NutzerID: 1},
	})
	anwendung.MitFreigaben(fuehrer)

	return anwendung
}

// **Die Antwort trägt Freigaben, Sichtbarkeit und die eigene Rolle.**
func TestDieFreigabenantwortTraegtDieEigeneRolle(t *testing.T) {
	t.Parallel()

	fuehrer := &freigabenfuehrer{
		rolle: domain.RolleVerwalter,
		abgelegt: []domain.Profilfreigabe{
			{
				ProfilID: 10, NutzerID: 2, Rolle: domain.RolleLeser,
				Angelegt: "2026-09-06T00:00:00Z",
			},
		},
	}

	antwort := mitFreigaben(fuehrer).Freigaben(t.Context(), "10")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if antwort.EigeneRolle != string(domain.RolleVerwalter) {
		t.Errorf("EigeneRolle = %q", antwort.EigeneRolle)
	}

	if len(antwort.Freigaben) != 1 {
		t.Fatalf("%+v", antwort.Freigaben)
	}

	// **Der Name kommt mit** — eine Kennung allein sagt nichts.
	if antwort.Freigaben[0].Name != "Zweiter" {
		t.Errorf("Name = %q", antwort.Freigaben[0].Name)
	}

	if antwort.Freigaben[0].NutzerID != "2" ||
		antwort.Freigaben[0].Rolle != string(domain.RolleLeser) {
		t.Errorf("%+v", antwort.Freigaben[0])
	}
}

// **Nur ein Verwalter setzt und löscht.**
func TestNurEinVerwalterSetztUndLoescht(t *testing.T) {
	t.Parallel()

	angaben := app.FreigabeangabenDTO{
		ProfilID: "20", NutzerID: "1", Rolle: string(domain.RolleLeser),
	}

	t.Run("ein Mitarbeiter wird abgewiesen", func(t *testing.T) {
		fuehrer := &freigabenfuehrer{rolle: domain.RolleMitarbeiter}

		if antwort := mitFreigaben(fuehrer).FreigabeSetzen(
			t.Context(), angaben); antwort.Fehler == nil {
			t.Error("der Mitarbeiter durfte setzen")
		}

		if antwort := mitFreigaben(fuehrer).FreigabeLoeschen(
			t.Context(), "20", "1"); antwort.Fehler == nil {
			t.Error("der Mitarbeiter durfte löschen")
		}
	})

	t.Run("ein Verwalter darf", func(t *testing.T) {
		// **Ohne diesen Fall stünde der darüber auch dann grün da, wenn jeder
		// abgewiesen würde** (§10).
		fuehrer := &freigabenfuehrer{rolle: domain.RolleVerwalter}

		if antwort := mitFreigaben(fuehrer).FreigabeSetzen(
			t.Context(), angaben); antwort.Fehler != nil {
			t.Fatalf("%+v", antwort.Fehler)
		}

		if fuehrer.gesehen.ProfilID != 20 || fuehrer.gesehen.NutzerID != 1 ||
			fuehrer.gesehen.Rolle != domain.RolleLeser {
			t.Errorf("gesehen = %+v", fuehrer.gesehen)
		}

		if antwort := mitFreigaben(fuehrer).FreigabeLoeschen(
			t.Context(), "20", "1"); antwort.Fehler != nil {
			t.Fatalf("%+v", antwort.Fehler)
		}
	})
}

// **Eine leere Liste ist eine Liste** (§3) — auf jedem Fehlerweg.
func TestEineLeereFreigabenlisteIstEineListe(t *testing.T) {
	t.Parallel()

	kaputt := &freigabenfuehrer{
		rolle: domain.RolleVerwalter, fehler: errors.New("die Platte ist voll"),
	}

	faelle := map[string]func(*app.App) app.FreigabenAntwort{
		"Lesen, krumme Kennung": func(a *app.App) app.FreigabenAntwort {
			return a.Freigaben(context.Background(), "keine-zahl")
		},
		"Lesen, Profil unbekannt": func(a *app.App) app.FreigabenAntwort {
			return a.Freigaben(context.Background(), "4711")
		},
		"Lesen, stolpernder Dienst": func(a *app.App) app.FreigabenAntwort {
			return a.Freigaben(context.Background(), "10")
		},
		"Setzen, krumme Profilkennung": func(a *app.App) app.FreigabenAntwort {
			return a.FreigabeSetzen(context.Background(),
				app.FreigabeangabenDTO{ProfilID: "keine-zahl", NutzerID: "1"})
		},
		"Setzen, krumme Nutzerkennung": func(a *app.App) app.FreigabenAntwort {
			return a.FreigabeSetzen(context.Background(),
				app.FreigabeangabenDTO{ProfilID: "10", NutzerID: "keine-zahl"})
		},
		"Setzen, Profil unbekannt": func(a *app.App) app.FreigabenAntwort {
			return a.FreigabeSetzen(context.Background(),
				app.FreigabeangabenDTO{ProfilID: "4711", NutzerID: "1"})
		},
		"Setzen, stolpernder Dienst": func(a *app.App) app.FreigabenAntwort {
			return a.FreigabeSetzen(context.Background(),
				app.FreigabeangabenDTO{ProfilID: "10", NutzerID: "2"})
		},
		"Löschen, krumme Kennung": func(a *app.App) app.FreigabenAntwort {
			return a.FreigabeLoeschen(context.Background(), "keine-zahl", "1")
		},
		"Löschen, stolpernder Dienst": func(a *app.App) app.FreigabenAntwort {
			return a.FreigabeLoeschen(context.Background(), "10", "2")
		},
	}

	for name, ruf := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			antwort := ruf(mitFreigaben(kaputt))

			if antwort.Fehler == nil {
				t.Error("der Fehler kam nicht durch")
			}

			if antwort.Freigaben == nil {
				t.Error("die Liste ist nil")
			}
		})
	}
}

// **Und wenn nicht zu sagen ist, wer arbeitet, wird abgewiesen.**
//
// **Und nicht durchgelassen**: wer das nicht sagen kann, kann auch nicht
// sagen, ob jemand darf — eine Erlaubnis aus Unwissen wäre keine (§8).
func TestOhneArbeitendenNutzerWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	anwendung.MitNutzern(&nutzerfuehrer{zeilen: []service.Nutzerzeile{{
		Nutzer:  domain.Nutzer{ID: 1, Name: "Ich"},
		Profile: []domain.Profil{{ID: 10, NutzerID: 1, Name: "Meins"}},
	}}})
	anwendung.MitProfilwechsel(&wechseldoppel{
		fehler: errors.New("kein Zeiger"),
	})
	anwendung.MitFreigaben(&freigabenfuehrer{rolle: domain.RolleVerwalter})

	if antwort := anwendung.Freigaben(t.Context(), "10"); antwort.Fehler == nil {
		t.Error("das Lesen kam durch")
	}

	if antwort := anwendung.FreigabeSetzen(t.Context(),
		app.FreigabeangabenDTO{ProfilID: "10", NutzerID: "2"},
	); antwort.Fehler == nil {
		t.Error("das Setzen kam durch")
	}
}

// **Und die Bindungen des Fensters reichen durch.**
func TestDieFreigabenbindungenReichenDurch(t *testing.T) {
	t.Parallel()

	fuehrer := &freigabenfuehrer{rolle: domain.RolleVerwalter}
	bindungen := app.NeueBindungen(mitFreigaben(fuehrer))

	if antwort := bindungen.Freigaben(t.Context(), "10"); antwort.Fehler != nil {
		t.Errorf("Freigaben: %+v", antwort.Fehler)
	}

	if antwort := bindungen.FreigabeSetzen(t.Context(), app.FreigabeangabenDTO{
		ProfilID: "20", NutzerID: "1", Rolle: string(domain.RolleLeser),
	}); antwort.Fehler != nil {
		t.Errorf("FreigabeSetzen: %+v", antwort.Fehler)
	}

	if antwort := bindungen.FreigabeLoeschen(t.Context(), "20", "1"); antwort.Fehler != nil {
		t.Errorf("FreigabeLoeschen: %+v", antwort.Fehler)
	}

	if fuehrer.geloescht != [2]int64{20, 1} {
		t.Errorf("gelöscht wurde %v", fuehrer.geloescht)
	}
}
