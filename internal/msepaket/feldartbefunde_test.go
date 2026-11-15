// **Eine unbekannte Feldart wird benannt** — FMT-02, S10.NAT-02 und Befund
// F12 (#1691).
package msepaket_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

func TestEineUnbekannteFeldartWirdBenannt(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		text   string
		zeilen []int
		wort   string
	}{
		"alle bekannt": {
			"card field:\n\ttype: text\n\tname: a\nset field:\n\ttype: package choice\n\tname: b\n", nil, "",
		},
		"eine unbekannte": {"card field:\n\ttype: audit-ungueltig\n\tname: a\n", []int{2}, "audit-ungueltig"},
		"eine fehlende":   {"mse version: 2.0.0\ncard field:\n\tname: a\n", []int{2}, "type:"},
		"im Stil und in der Exportvorlage": {
			"extra card field:\n\ttype: bild\n\tname: a\nstyling field:\n\ttype: text\n\tname: b\noption field:\n\ttype: zahl\n\tname: c\n",
			[]int{2, 8},
			"bild",
		},
		"ein type unter einem anderen Block zählt nicht": {"card style:\n\tname:\n\t\ttype: egal\n", nil, ""},
	} {
		befunde := msepaket.Feldartbefunde(fall.text)

		zeilen := make([]int, 0, len(befunde))
		for _, befund := range befunde {
			zeilen = append(zeilen, befund.Stelle.Zeile)
		}

		if len(zeilen) != len(fall.zeilen) {
			t.Errorf("%s: Befunde in %v, erwartet %v", name, zeilen, fall.zeilen)

			continue
		}

		for i := range zeilen {
			if zeilen[i] != fall.zeilen[i] {
				t.Errorf("%s: Befunde in %v, erwartet %v", name, zeilen, fall.zeilen)
			}
		}

		if fall.wort != "" && !strings.Contains(befunde[0].Meldung, fall.wort) {
			t.Errorf("%s: %q nennt %q nicht", name, befunde[0].Meldung, fall.wort)
		}
	}

	if befunde := msepaket.Feldartbefunde("card field:\n\t: kaputt\n\t\tzu tief\n"); befunde == nil {
		t.Error("eine unlesbare Datei gibt eine leere Liste und nicht nil")
	}
}

// **Die Art steht einzeln im Befund** — das Fenster setzt sie in seinen
// eigenen Satz ein, und ein fehlendes `type:` hat keine.
func TestEinFeldartbefundNenntSeineArt(t *testing.T) {
	t.Parallel()

	befunde := msepaket.Feldartbefunde("card field:\n\tname: a\ncard field:\n\ttype: bild\n")
	if len(befunde) != 2 || befunde[0].Art != "" || befunde[1].Art != "bild" {
		t.Fatalf("Befunde %+v, erwartet eine fehlende und die Art »bild«", befunde)
	}

	// **Die neun aus `read_new<Field>`** (`src/data/field.cpp` Zeilen 74–94,
	// Fassung f9d9356 vom 2021-07-06), sortiert.
	bekannt := strings.Join(msepaket.BekannteFeldarten(), ", ")
	if bekannt != "boolean, choice, color, image, info, multiple choice, package choice, symbol, text" ||
		!strings.Contains(befunde[1].Meldung, bekannt) {
		t.Errorf("bekannte Arten %q, Meldung %q", bekannt, befunde[1].Meldung)
	}
}
