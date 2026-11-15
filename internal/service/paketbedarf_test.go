package service_test

// **Was ein Set an Paketen braucht, und was davon fehlt** — DOM-01.AK-04
// (#1696):
// „Fehlende Abhängigkeiten lassen sich aus dem betroffenen Fachworkflow
// diagnostizieren und gezielt beheben."

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// bedarfsplatz ist ein Katalog mit Spiel und Stil — der Stil braucht eine
// Symbolschrift, die fehlt — und drei Sets.
func bedarfsplatz(t *testing.T) *service.Paketverwalter {
	t.Helper()

	verwaltet := t.TempDir()
	katalog := &ablage{pakete: []domain.Paket{
		{Art: "mse-game", Name: "probe", Version: "2026-08-15", Pfad: filepath.Join(verwaltet, "probe.mse-game")},
		{
			Art: "mse-style", Name: "probe-schlicht", Version: "2026-08-15",
			Pfad: filepath.Join(verwaltet, "probe-schlicht.mse-style"),
			Braucht: []domain.Bedarf{
				{Paket: "probe.mse-game", Version: "2026-09-01"},
				{Paket: "probe-symbole.mse-symbol-font"},
			},
		},
	}}

	return service.NeuerPaketverwalter(katalog, verwaltet, verwaltet).
		MitSetkoepfen(setkoepfe{koepfe: []domain.SetHead{
			{ID: 1, Name: "Erstes", Game: "probe", Style: "schlicht"},
			{ID: 2, Name: "Zweites", Game: "fremd", Style: "kunst"},
			{ID: 3, Name: "Drittes", Game: "probe"},
		}})
}

func TestDerBedarfEinesSetsGehtDurchDieAbhaengigkeiten(t *testing.T) {
	befunde, err := bedarfsplatz(t).Setbedarf(t.Context(), 1)
	if err != nil {
		t.Fatalf("Bedarf: %v", err)
	}

	// **Der Stil ist da, aber ihm fehlt eine Schrift, und sein Spiel ist zu
	// alt** — zwei Befunde, beide am Stil.
	if len(befunde) != 2 ||
		befunde[0].Paket != "probe-schlicht.mse-style" || befunde[0].Art != service.AbhaengigkeitZuAlt ||
		befunde[1].Braucht != "probe-symbole.mse-symbol-font" || befunde[1].Art != service.AbhaengigkeitFehlt {
		t.Errorf("die Befunde: %+v", befunde)
	}
}

func TestEinSetOhneSeinSpielUndSeinenStil(t *testing.T) {
	befunde, err := bedarfsplatz(t).Setbedarf(t.Context(), 2)
	if err != nil {
		t.Fatalf("Bedarf: %v", err)
	}

	// **Das Set selbst braucht sie** — und für den Stil steht der Name, den
	// die Stilsuche zuerst liest: die Kurzform mit dem Spiel davor.
	if len(befunde) != 2 ||
		befunde[0].Paket != "Zweites" || befunde[0].Braucht != "fremd.mse-game" ||
		befunde[1].Braucht != "fremd-kunst.mse-style" ||
		befunde[0].Art != service.AbhaengigkeitFehlt || befunde[1].Art != service.AbhaengigkeitFehlt {
		t.Errorf("die Befunde: %+v", befunde)
	}
}

func TestEinSetOhneStilUndMitAllemBrauchtNichts(t *testing.T) {
	befunde, err := bedarfsplatz(t).Setbedarf(t.Context(), 3)
	if err != nil || len(befunde) != 0 {
		t.Errorf("Befunde: %+v, %v", befunde, err)
	}
}

func TestDerBedarfSagtWasNichtGeht(t *testing.T) {
	if _, err := bedarfsplatz(t).Setbedarf(t.Context(), 99); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("ein fehlendes Set: %v", err)
	}

	klemmt := errors.New("klemmt")

	if _, err := service.NeuerPaketverwalter(&ablage{}, t.TempDir()).
		MitSetkoepfen(setkoepfe{fehler: klemmt}).Setbedarf(t.Context(), 1); !errors.Is(err, klemmt) {
		t.Errorf("die Setköpfe klemmen: %v", err)
	}

	if _, err := service.NeuerPaketverwalter(&ablage{lesefehler: klemmt}, t.TempDir()).
		MitSetkoepfen(setkoepfe{koepfe: []domain.SetHead{{ID: 1, Game: "probe"}}}).
		Setbedarf(t.Context(), 1); !errors.Is(err, klemmt) {
		t.Errorf("der Katalog klemmt: %v", err)
	}

	// **Ohne Setköpfe kein Set** — eine Fläche ohne Bestand kennt keines.
	if _, err := service.NeuerPaketverwalter(&ablage{}, t.TempDir()).Setbedarf(t.Context(), 1); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("ohne Setköpfe: %v", err)
	}
}

// **Mehr als eine Stufe tief, und ein Kreis endet** — das Spiel braucht ein
// Include, das Include eine Schrift, die fehlt; und das Include braucht das
// Spiel zurück.
func TestDerBedarfGehtDurchJedeStufe(t *testing.T) {
	katalog := &ablage{pakete: []domain.Paket{
		{Art: "mse-game", Name: "tief", Braucht: []domain.Bedarf{{Paket: "tief-teile.mse-include"}}},
		{Art: "mse-include", Name: "tief-teile", Braucht: []domain.Bedarf{
			{Paket: "tief-schrift.mse-symbol-font"},
			{Paket: "tief.mse-game"},
		}},
	}}

	befunde, err := service.NeuerPaketverwalter(katalog, t.TempDir()).
		MitSetkoepfen(setkoepfe{koepfe: []domain.SetHead{{ID: 7, Name: "Tief", Game: "tief"}}}).
		Setbedarf(t.Context(), 7)
	if err != nil || len(befunde) != 1 ||
		befunde[0].Paket != "tief-teile.mse-include" || befunde[0].Braucht != "tief-schrift.mse-symbol-font" {
		t.Errorf("die Befunde: %+v, %v", befunde, err)
	}
}

// **Ein Stilwert, der schon mit dem Spiel beginnt**, wird in der Meldung so
// genannt, wie er dasteht — `probe-probe-weg` sucht niemand. Gesucht wird
// trotzdem unter beiden Namen.
func TestDerFehlendeStilHeisstWieErImKopfSteht(t *testing.T) {
	befunde, err := service.NeuerPaketverwalter(&ablage{pakete: []domain.Paket{{Art: "mse-game", Name: "probe"}}}, t.TempDir()).
		MitSetkoepfen(setkoepfe{koepfe: []domain.SetHead{{ID: 4, Name: "Viertes", Game: "probe", Style: "probe-weg"}}}).
		Setbedarf(t.Context(), 4)
	if err != nil || len(befunde) != 1 || befunde[0].Braucht != "probe-weg.mse-style" {
		t.Errorf("die Befunde: %+v, %v", befunde, err)
	}
}

// **Der Bedarf eines Kopfs, der noch nicht im Bestand steht** — FMT-03
// (#1692): die Vorschau vor dem Einfahren fragt mit dem Kopf aus der Datei,
// und die Antwort ist dieselbe wie später für das eingefahrene Set.
func TestDerBedarfEinesKopfsOhneSet(t *testing.T) {
	befunde, err := bedarfsplatz(t).Kopfbedarf(t.Context(),
		domain.SetHead{Name: "Aus der Datei", Game: "fremd", Style: "kunst"})
	if err != nil {
		t.Fatalf("Bedarf: %v", err)
	}

	if len(befunde) != 2 || befunde[0].Paket != "Aus der Datei" || befunde[0].Braucht != "fremd.mse-game" ||
		befunde[1].Braucht != "fremd-kunst.mse-style" {
		t.Errorf("die Befunde: %+v", befunde)
	}
}
