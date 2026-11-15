package app_test

import (
	"context"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// einordnungskatalog gibt einen Katalog mit Gruppe, Rang und Art.
type einordnungskatalog struct {
	planlos

	pakete []domain.Paket
}

// planlos gibt Plan, Ersetzen und Zurück ohne Wirkung — für Doppel, deren
// Läufe sie nicht brauchen (DOM-01.AK-02, #1696).
type planlos struct{}

// Herkunftsart ohne Wirkung: mitgeliefert, wie ein Paket ohne Ort.
func (planlos) Herkunftsart(domain.Paket) domain.Paketherkunftsart { return domain.PaketMitgeliefert }

func (planlos) PlanEntfernen(context.Context, string, string) (service.Paketplan, error) {
	return service.Paketplan{}, nil
}

// Kopieren gibt eine leere Kopie — FMT-02 (#1691).
func (planlos) Kopieren(context.Context, string, string, string) (domain.Paket, error) {
	return domain.Paket{}, nil
}

func (planlos) Weitergeben(context.Context, []string, string) (string, error) { return "", nil }

func (planlos) Einlesen(context.Context, string) (*service.Uebernahme, error) { return nil, nil }

func (planlos) PlanEinbinden(context.Context, string) (service.Paketplan, error) {
	return service.Paketplan{}, nil
}

func (planlos) PlanZurueck(context.Context, string, string) (service.Paketplan, error) {
	return service.Paketplan{}, nil
}

func (planlos) Ersetzen(context.Context, string, bool) (domain.Paket, error) {
	return domain.Paket{}, nil
}

func (planlos) Zurueck(context.Context, string, string) (domain.Paket, error) {
	return domain.Paket{}, nil
}

func (planlos) Setbedarf(context.Context, int64) ([]service.Abhaengigkeitsbefund, error) {
	return nil, nil
}

func (planlos) Kopfbedarf(context.Context, domain.SetHead) ([]service.Abhaengigkeitsbefund, error) {
	return nil, nil
}

func (e *einordnungskatalog) Auffrischen(context.Context) ([]domain.Befund, error) {
	return nil, nil
}

func (e *einordnungskatalog) Anlegen(context.Context, service.Neuanlage) (domain.Paket, error) {
	return domain.Paket{}, nil
}

func (e *einordnungskatalog) Vermerken(context.Context, ...service.Vermerk) ([]domain.Befund, error) {
	return nil, nil
}

func (e *einordnungskatalog) Uebernehmen(context.Context, string) (*service.Uebernahme, error) {
	return nil, nil
}

func (e *einordnungskatalog) Abhaengigkeiten(context.Context) ([]service.Abhaengigkeitsbefund, error) {
	return nil, nil
}

func (e *einordnungskatalog) Liste(context.Context) ([]domain.Paket, error) {
	return e.pakete, nil
}

// Uebersicht gibt die Pakete ohne Zusatz — DOM-01.AK-01 (#1696): das
// Fenster liest den Katalog seitdem über sie.
func (e *einordnungskatalog) Uebersicht(context.Context) ([]service.Paketzeile, error) {
	zeilen := make([]service.Paketzeile, 0, len(e.pakete))
	for _, paket := range e.pakete {
		zeilen = append(zeilen, service.Paketzeile{Paket: paket})
	}

	return zeilen, nil
}

func (e *einordnungskatalog) Installieren(context.Context, string) (domain.Paket, error) {
	return domain.Paket{}, nil
}

func (e *einordnungskatalog) Entfernen(context.Context, string, string, string) error {
	return nil
}

// **Die Oberfläche bekommt, was sie zum Ordnen braucht** (#244).
//
// Drei Angaben, und keine davon rechnet das Fenster sich selbst aus:
// `gruppe` und `rang` stehen in der Paketdatei, `artrang` ist MSEs
// `PackageType` (`package_update_list.hpp`, Zeilen 57–67, Fassung f9d9356 vom
// 2021-07-06). Die dritte könnte das Fenster aus `art` ableiten — dann stünde
// die portierte Reihenfolge in TypeScript, wo kein Go-Test sie hält.
func TestDerKatalogNenntGruppeRangUndArtrang(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitPaketen(&einordnungskatalog{pakete: []domain.Paket{
		{Art: "mse-style", Name: "magic-m15", Kurzname: "m15 normal", Gruppe: "magic/m15 style", Rang: 95},
		{Art: "mse-game", Name: "magic"},
	}})

	antwort := anwendung.Pakete(t.Context())
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if len(antwort.Pakete) != 2 {
		t.Fatalf("%d Pakete", len(antwort.Pakete))
	}

	stil, spiel := antwort.Pakete[0], antwort.Pakete[1]

	if stil.Gruppe != "magic/m15 style" || stil.Rang != 95 {
		t.Errorf("Stil: %q / %d", stil.Gruppe, stil.Rang)
	}

	// **Der Anzeigename kommt fertig** (UX-05, S07.NAT-02, #1677) — mit
	// Kurznamen dieser, ohne ihn die Kennung, beide großgeschrieben wie MSE2.
	if stil.Anzeigename != "M15 Normal" || spiel.Anzeigename != "Magic" {
		t.Errorf("Anzeigenamen: %q, %q", stil.Anzeigename, spiel.Anzeigename)
	}

	// **Das Spiel steht vor dem Stil** — MSEs Reihenfolge, nicht die
	// alphabetische. Geprüft wird die Ordnung und nicht die Zahl: sie überlebt
	// es, wenn jemand eine achte Art dazwischenschiebt.
	if spiel.Artrang >= stil.Artrang {
		t.Errorf("Spiel (%d) steht nicht vor Stil (%d)", spiel.Artrang, stil.Artrang)
	}
}
