package service_test

// **Was ein Profil vom Menschen erbt** — Werkstück 9 von Phase 4.12
// (#1351),
// Punkt 9.

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// menschenwerte ist ein Leser der Einstellungen eines Menschen.
type menschenwerte struct {
	werte  map[string]string
	fehler error
}

func (m menschenwerte) NutzerwerteZumProfil(context.Context) (map[string]string, error) {
	return m.werte, m.fehler
}

// **Der Wert des Menschen gilt, wo das Profil erbt.**
func TestEinErbendesProfilNimmtDenWertDesMenschen(t *testing.T) {
	t.Parallel()

	verwalter := service.NeuerEinstellungsverwalter(&einstellungsablage{}).
		MitNutzerwerten(menschenwerte{werte: map[string]string{
			domain.EinstellungErscheinungsbild: "dunkel",
			domain.EinstellungSprache:          "en",
		}})

	werte, err := verwalter.Alle(t.Context())
	if err != nil {
		t.Fatalf("%v", err)
	}

	if werte[domain.EinstellungErscheinungsbild] != "dunkel" {
		t.Errorf("%v", werte)
	}

	if werte[domain.EinstellungSprache] != "en" {
		t.Errorf("%v", werte)
	}
}

// **Und die Wahl des Profils sticht die des Menschen.**
//
// **Das ist die Richtung, und sie steht hier als Gegenprobe**: ohne diesen
// Lauf hielte der darüber auch, wenn die Verschmelzung immer den Menschen
// nähme.
func TestEineProfilwahlStichtDenMenschen(t *testing.T) {
	t.Parallel()

	verwalter := service.NeuerEinstellungsverwalter(&einstellungsablage{
		werte: map[string]string{domain.EinstellungErscheinungsbild: "hell"},
	}).MitNutzerwerten(menschenwerte{werte: map[string]string{
		domain.EinstellungErscheinungsbild: "dunkel",
	}})

	werte, err := verwalter.Alle(t.Context())
	if err != nil {
		t.Fatalf("%v", err)
	}

	if werte[domain.EinstellungErscheinungsbild] != "hell" {
		t.Errorf("%v", werte)
	}
}

// **Der Stand nennt beides: die Wahl und das, was gilt.**
//
// **Beide, und der Unterschied ist die ganze Zusage**: die Auswahlliste zeigt
// die Wahl, die Anwendung richtet sich nach dem Wirksamen.
func TestDerStandNenntWahlUndWirkung(t *testing.T) {
	t.Parallel()

	verwalter := service.NeuerEinstellungsverwalter(&einstellungsablage{}).
		MitNutzerwerten(menschenwerte{werte: map[string]string{
			domain.EinstellungErscheinungsbild: "dunkel",
		}})

	stand, err := verwalter.Stand(t.Context())
	if err != nil {
		t.Fatalf("%v", err)
	}

	eine := stand[domain.EinstellungErscheinungsbild]
	if eine.Gewaehlt != domain.Erben || eine.Wirksam != "dunkel" {
		t.Errorf("%+v", eine)
	}

	// **Und was niemand erbt, steht zweimal gleich da** — ohne Sonderfall
	// beim Aufrufer.
	andere := stand[domain.EinstellungAktualisierung]
	if andere.Gewaehlt != andere.Wirksam {
		t.Errorf("%+v", andere)
	}
}

// **Ein Lesefehler beim Menschen reißt ab.**
//
// **Und wird nicht verschluckt**: eine Antwort, die dann die Vorgaben zeigte,
// sähe aus wie eine Wahl — und niemand wüsste, dass seine eigene nicht
// gelesen wurde (§8).
func TestEinLesefehlerBeimMenschenReisstAb(t *testing.T) {
	t.Parallel()

	verwalter := service.NeuerEinstellungsverwalter(&einstellungsablage{}).
		MitNutzerwerten(menschenwerte{fehler: errors.New("die Platte schweigt")})

	if _, err := verwalter.Alle(t.Context()); err == nil {
		t.Error("der Fehler ging verloren")
	}

	if _, err := verwalter.Stand(t.Context()); err == nil {
		t.Error("der Fehler ging beim Stand verloren")
	}
}
