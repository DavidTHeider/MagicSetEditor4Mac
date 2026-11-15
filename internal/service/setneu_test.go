// **Ein neues, leeres Set** —
// #891.
package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// neuanlage merkt sich, was angelegt werden sollte.
type neuanlage struct {
	set    *domain.Set
	roh    []domain.RawBlock
	fehler error
}

func (a *neuanlage) SetAnlegen(
	_ context.Context, set *domain.Set, roh []domain.RawBlock,
) (int64, error) {
	a.set, a.roh = set, roh

	return 7, a.fehler
}

// **Ein neues Set trägt Name, Spiel, Stil — und den Namen als Setwert.**
//
// `set info: title:` ist das Feld, das ein geschriebenes `.mse-set` trägt;
// ohne es hieße das Set im Bestand so und in der Datei gar nicht.
func TestEinNeuesSetTraegtSeinenNamenZweimal(t *testing.T) {
	t.Parallel()

	probe := &neuanlage{}

	kennung, err := service.NeuerImporteur(probe).Neu(
		t.Context(), "Mein Set", "probe", "probe-schlicht")
	if err != nil {
		t.Fatalf("Neu: %v", err)
	}

	if kennung != 7 {
		t.Errorf("die Kennung ist %d", kennung)
	}

	if probe.set.Name != "Mein Set" || probe.set.Game != "probe" ||
		probe.set.Style != "probe-schlicht" {
		t.Errorf("angelegt wurde %+v", probe.set)
	}

	if len(probe.set.Values) != 1 || probe.set.Values[0].Field != "title" ||
		probe.set.Values[0].Text != "Mein Set" {
		t.Errorf("die Setwerte sind %+v", probe.set.Values)
	}

	// **Keine Rohblöcke.** Ein neues Set hat keine Herkunft, also nichts, was
	// ein Schreiber unverändert zurückgeben müsste.
	if probe.roh != nil {
		t.Errorf("es wurden Rohblöcke mitgegeben: %+v", probe.roh)
	}

	// **Und keine Karten** — aber eine leere Liste, kein `nil`: der Schreiber
	// zählt sie, und `nil` zu zählen wäre dasselbe, sähe aber wie ein
	// vergessener Wert aus.
	if probe.set.Cards == nil || len(probe.set.Cards) != 0 {
		t.Errorf("die Karten sind %+v", probe.set.Cards)
	}
}

// **Ohne Spiel, Namen und Stil geht es nicht.**
//
// **Der Stil steht seit dem 2026-08-25 dabei**, und zwar aus derselben Quelle
// wie das Spiel — `Set::validate` verlangt beide:
//
//	if (!stylesheet) {
//	  throw Error(_ERROR_("no stylesheet specified for the set"));
//	}
//	if (stylesheet->game != game) {
//	  throw Error(_ERROR_("stylesheet and set refer to different game"));
//	}
//
// `src/data/set.cpp`, Zeilen 143–149, Fassung f9d9356 vom 2021-07-06. Gefunden
// hat es der Nachlauf des Zweitblicks über Phase Y: hier stand die Behauptung,
// MSEs Bedingung sitze am Knopf des Dialogs, und das war eine Ablehnung ohne
// Deckung (§10).
func TestEinNeuesSetBrauchtSpielNamenUndStil(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		setname, spiel, stil, feld string
	}{
		"ohne Spiel":       {"Mein Set", "  ", "probe-schlicht", "game"},
		"ohne Namen":       {"   ", "probe", "probe-schlicht", "name"},
		"ohne Stil":        {"Mein Set", "probe", "  ", "stil"},
		"Stil eines ande.": {"Mein Set", "probe", "magic-m15", "stil"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			probe := &neuanlage{}

			_, err := service.NeuerImporteur(probe).Neu(
				t.Context(), fall.setname, fall.spiel, fall.stil)

			var befund *domain.FieldError
			if !errors.As(err, &befund) {
				t.Fatalf("Fehler %v, erwartet war ein Feldbefund", err)
			}

			if befund.Field != fall.feld {
				t.Errorf("der Befund nennt das Feld %q, erwartet %q", befund.Field, fall.feld)
			}

			if probe.set != nil {
				t.Error("es wurde trotzdem angelegt")
			}
		})
	}
}

// **Ein Fehler des Speichers kommt durch.**
func TestEinGescheitertesAnlegenMeldetSich(t *testing.T) {
	t.Parallel()

	kaputt := errors.New("die Platte ist voll")
	probe := &neuanlage{fehler: kaputt}

	if _, err := service.NeuerImporteur(probe).Neu(
		t.Context(), "Mein Set", "probe", "probe-schlicht"); !errors.Is(err, kaputt) {
		t.Errorf("Fehler %v, erwartet war die volle Platte", err)
	}
}
