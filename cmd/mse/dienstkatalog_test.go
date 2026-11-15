package main

import (
	"strings"
	"testing"
)

func TestDerDienstkatalogNenntJedeFaehigkeit(t *testing.T) {
	ort, _, _ := mitKarte(t)

	code, aus, fehler := fahreAn(t, ort, "dienstkatalog")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	for _, erwartet := range []string{
		"scryfall\tsuche\tbereit", "scrydex\tsuche\taus", "gelaufen am",
		"internal/mtgapi/netzlauf_test.go",
	} {
		if !strings.Contains(aus, erwartet) {
			t.Errorf("die Ausgabe nennt %q nicht:\n%s", erwartet, aus)
		}
	}
}

func TestDerDienstkatalogAlsJSONTraegtVertragUndNachweis(t *testing.T) {
	ort, _, _ := mitKarte(t)

	code, aus, fehler := fahreAn(t, ort, "dienstkatalog", "--json")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	for _, erwartet := range []string{
		`"eintraege"`, `"basisadresse"`, `"giltFuerDieseFassung"`, `"abgerufen"`,
	} {
		if !strings.Contains(aus, erwartet) {
			t.Errorf("die Antwort nennt %q nicht: %s", erwartet, aus)
		}
	}

	if strings.Contains(aus, `"eintraege": null`) {
		t.Errorf("die Liste ist null: %s", aus)
	}
}

func TestDerDienstkatalogNimmtKeineArgumente(t *testing.T) {
	ort, _, _ := mitKarte(t)

	if code, _, _ := fahreAn(t, ort, "dienstkatalog", "scryfall"); code != CodeBedienung {
		t.Errorf("Rückgabewert %d, erwartet %d", code, CodeBedienung)
	}
}
