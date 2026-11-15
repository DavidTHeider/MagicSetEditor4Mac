package domain_test

// **Die Herkunft eines Modells ist eine Kette** — Werkstück A6 von Phase 5.3
// (#1445).

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// eineKette baut eine gültige Kette: Basis → Adapter → Umwandlung → Artefakt.
func eineKette() domain.KIHerkunft {
	return domain.KIHerkunft{Glieder: []domain.KIGlied{
		{
			Art: domain.KIBasis, Kennung: "basis",
			Quelle: "hf:beispiel/grundmodell", Revision: "a1b2c3d4",
			Dateien: []domain.KIDateihash{{Name: "modell.safetensors", Hash: "sha256:aaaa"}},
			Lizenz:  "Apache-2.0",
		},
		{
			Art: domain.KIAdapter, Kennung: "adapter", Aus: "basis",
			Quelle: "hf:beispiel/adapter", Revision: "e5f6a7b8",
			Dateien: []domain.KIDateihash{{Name: "adapter.safetensors", Hash: "sha256:bbbb"}},
			Lizenz:  "Apache-2.0",
		},
		{
			Art: domain.KIUmwandlung, Kennung: "gguf", Aus: "adapter",
			Quelle: "eigene Umwandlung", Revision: "2026-09-11",
			Dateien: []domain.KIDateihash{{Name: "modell.gguf", Hash: "sha256:cccc"}},
			Lizenz:  "Apache-2.0",
		},
		{
			Art: domain.KIArtefakt, Kennung: "ausgefuehrt", Aus: "gguf",
			Quelle: "ollama:beispiel", Revision: "sha256:dddd",
			Dateien: []domain.KIDateihash{{Name: "modell.gguf", Hash: "sha256:cccc"}},
			Lizenz:  "Apache-2.0",
		},
	}}
}

// **Eine vollständige Kette geht durch.**
func TestEineVollstaendigeModellherkunftGehtDurch(t *testing.T) {
	t.Parallel()

	if fehler := eineKette().Pruefe(); fehler != nil {
		t.Fatalf("die Kette wurde abgewiesen: %v", fehler)
	}
}

// **Eine Kette ohne Artefakt sagt nicht, was wirklich gelaufen ist.**
//
// Das Issue verlangt „Basis, Adapter/LoRA, Konvertierung/Quantisierung und
// **ausgeführtes Artefakt**". Ohne das letzte Glied beschreibt die Kette, was
// es gibt, und nicht, was benutzt wurde.
func TestEineHerkunftOhneArtefaktWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	kette := eineKette()
	kette.Glieder = kette.Glieder[:3]

	if fehler := kette.Pruefe(); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("der Fehler ist %v", fehler)
	}
}

// **Ein Kreis wird abgewiesen.**
//
// Das Issue verlangt eine **azyklische** Kette. Ein Kreis liefe beim Ablaufen
// ewig — und wäre als Herkunft ohnehin unsinnig: nichts stammt von sich selbst.
func TestEinKreisInDerHerkunftWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	kette := eineKette()
	kette.Glieder[0].Aus = "ausgefuehrt"

	if fehler := kette.Pruefe(); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("der Fehler ist %v", fehler)
	}
}

// **Ein Glied, dessen Vorgänger es nicht gibt, ebenso.**
func TestEinFehlenderVorgaengerWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	kette := eineKette()
	kette.Glieder[1].Aus = "gibtesnicht"

	if fehler := kette.Pruefe(); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("der Fehler ist %v", fehler)
	}
}

// **Und ein unvollständiges Glied auch — jedes Feld für sich.**
func TestJedesFeldEinesGliedsIstPflicht(t *testing.T) {
	t.Parallel()

	for name, kaputt := range map[string]func(*domain.KIGlied){
		"ohne Kennung":  func(g *domain.KIGlied) { g.Kennung = "" },
		"ohne Quelle":   func(g *domain.KIGlied) { g.Quelle = "" },
		"ohne Revision": func(g *domain.KIGlied) { g.Revision = "" },
		"ohne Lizenz":   func(g *domain.KIGlied) { g.Lizenz = "" },
		"ohne Dateien":  func(g *domain.KIGlied) { g.Dateien = nil },
		"mit einer Datei ohne Hash": func(g *domain.KIGlied) {
			g.Dateien = []domain.KIDateihash{{Name: "modell.gguf"}}
		},
		"mit einer Datei ohne Namen": func(g *domain.KIGlied) {
			g.Dateien = []domain.KIDateihash{{Hash: "sha256:cccc"}}
		},
		"mit einer Art, die es nicht gibt": func(g *domain.KIGlied) {
			g.Art = "geraten"
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			kette := eineKette()
			kaputt(&kette.Glieder[3])

			if fehler := kette.Pruefe(); !errors.Is(fehler, domain.ErrValidation) {
				t.Errorf("%s kam durch: %v", name, fehler)
			}
		})
	}
}

// **Eine leere Kette ist »unbekannt« und kein Fehler.**
//
// # Der Unterschied, um den es geht
//
// Das Issue sagt: „Fehlende Herkunft explizit unbekannt; **kein erratener
// HF-Commit aus einem Ollama-Alias**." Eine leere Kette sagt „wir wissen es
// nicht" — und das ist eine Auskunft. Eine **halbe** Kette sagt „wir wissen
// es", und das wäre die Lüge.
func TestEineLeereHerkunftIstUnbekanntUndKeinFehler(t *testing.T) {
	t.Parallel()

	leer := domain.KIHerkunft{}

	if fehler := leer.Pruefe(); fehler != nil {
		t.Fatalf("die leere Kette wurde abgewiesen: %v", fehler)
	}

	if !leer.Unbekannt() {
		t.Error("die leere Kette gilt nicht als unbekannt")
	}

	if eineKette().Unbekannt() {
		t.Error("eine volle Kette gilt als unbekannt")
	}
}

// **Und eine zu lange Kette wird abgewiesen.**
//
// Das Issue verlangt eine **begrenzte** Kette. Ohne Grenze wäre eine Herkunft
// mit tausend Gliedern zulässig — und was sie belegte, könnte niemand lesen.
func TestEineZuLangeHerkunftWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	kette := domain.KIHerkunft{}

	for i := range domain.KIHoechstensGlieder + 1 {
		glied := domain.KIGlied{
			Art: domain.KIUmwandlung, Kennung: string(rune('a' + i)),
			Quelle: "x", Revision: "y", Lizenz: "z",
			Dateien: []domain.KIDateihash{{Name: "d", Hash: "h"}},
		}

		if i > 0 {
			glied.Aus = string(rune('a' + i - 1))
		} else {
			glied.Art = domain.KIBasis
		}

		kette.Glieder = append(kette.Glieder, glied)
	}

	kette.Glieder[len(kette.Glieder)-1].Art = domain.KIArtefakt

	if fehler := kette.Pruefe(); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("der Fehler ist %v", fehler)
	}
}

// **Dieselbe Kennung zweimal wird abgewiesen.**
//
// **Sie ist der Bezug zwischen zwei Gliedern**, und zwei Glieder mit
// demselben Namen machen aus „stammt aus X" eine Frage, die niemand
// beantworten kann — sie gäbe zwei verschiedene Ketten und keinen Grund, die
// eine zu nehmen.
func TestDieselbeKennungZweimalWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	kette := eineKette()
	kette.Glieder[1].Kennung = "basis"

	if fehler := kette.Pruefe(); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("der Fehler ist %v", fehler)
	}
}
