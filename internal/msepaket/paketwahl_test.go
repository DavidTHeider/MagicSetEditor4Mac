package msepaket_test

import (
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Eine `package choice` ist eine eigene Art, keine Textzeile**
// (#632).
//
// Sie war bis hierher `KindUnknown` und fiel als Text durch — 627 Vorkommen in
// 349 Stilpaketen des Packs (#630).
//
// **Drei Angaben gehören dazu**, und ohne sie ließe sich die Liste nicht
// bauen: `match:` sagt, welche Pakete zur Wahl stehen, `required:`, ob der
// leere Wert erlaubt ist, und `empty name:`, wie er heißt
// (`src/data/field/package_choice.hpp:27–35`, Fassung f9d9356 vom
// 2021-07-06).
func TestEinePaketwahlTraegtMusterUndPflicht(t *testing.T) {
	paket, err := msepaket.Oeffne(filepath.Join("..", "..", "testdaten", "probe-stil.mse-style"))
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() { _ = paket.Schliessen() }) //nolint:errcheck // gelesen ist gelesen

	stil, befunde := paket.Stilmodell()
	if len(befunde) != 0 {
		t.Fatalf("Befunde: %+v", befunde)
	}

	gefunden := map[string]domain.StyleField{}
	for _, option := range stil.Stiloptionen {
		gefunden[option.Name] = option
	}

	pflicht, da := gefunden["mana symbols"]
	if !da {
		t.Fatalf("die Stiloption `mana symbols` fehlt: %+v", gefunden)
	}

	if pflicht.Kind != domain.KindPackageChoice {
		t.Errorf("Art %q, erwartet %q", pflicht.Kind, domain.KindPackageChoice)
	}

	if pflicht.Match != "probe-mana-*.mse-symbol-font" {
		t.Errorf("Match %q", pflicht.Match)
	}

	// **Ohne `required:` gilt `true`** — der Vorgabewert des Feldes in MSE2.
	if !pflicht.Required {
		t.Error("Required ist falsch, erwartet wahr — die Vorgabe ist Pflicht")
	}

	frei, da := gefunden["overlay"]
	if !da {
		t.Fatalf("die Stiloption `overlay` fehlt: %+v", gefunden)
	}

	if frei.Required {
		t.Error("Required ist wahr, erwartet falsch — `required: false` steht da")
	}

	if frei.EmptyName != "keines" {
		t.Errorf("EmptyName %q, erwartet `keines`", frei.EmptyName)
	}
}

// **Ohne `empty name:` heißt der leere Eintrag `none`** — so steht es im
// Kopf von `PackageChoiceField` (`package_choice.hpp:27`, dieselbe Fassung).
func TestOhneEigenenNamenHeisstDerLeereEintragNone(t *testing.T) {
	paket, err := msepaket.Oeffne(filepath.Join("..", "..", "testdaten", "probe-stil.mse-style"))
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() { _ = paket.Schliessen() }) //nolint:errcheck // gelesen ist gelesen

	stil, _ := paket.Stilmodell()

	for _, option := range stil.Stiloptionen {
		if option.Name == "pt symbols" && option.EmptyName != "none" {
			t.Errorf("EmptyName %q, erwartet `none`", option.EmptyName)
		}
	}
}
