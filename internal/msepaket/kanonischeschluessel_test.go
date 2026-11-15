package msepaket_test

// **Ein Schlüssel mit Unterstrich ist derselbe wie mit Leerzeichen** —
// #1972.
//
// MSE2 macht jeden Schlüssel kanonisch, bevor es ihn vergleicht
// (`key = canonical_name_form(trim(key))`, `src/util/io/reader.cpp` Zeile
// 217, Fassung f9d9356 vom 2021-07-06); `z_index:` und `z index:` sind dort
// dasselbe. Die Feldstil- und Schriftleser verglichen den Schlüssel, wie er
// dastand — `include_file:` war derselbe Fall (QA-06b, #1967).

import (
	"testing"
)

// **Der Feldstil** — `feldstilAus`.
func TestEinFeldstilLiestSeineSchluesselKanonisch(t *testing.T) {
	t.Parallel()

	paket := paketAusText("card style:\n\tname:\n\t\tz_index: 3\n\t\trender_style: text\n"+
		"\t\talways_symbol: true\n\t\tpadding_left: 4\n\t\tsymbol_font:\n\t\t\tname: magic-mana\n",
		"probe.mse-style")
	if paket == nil {
		t.Fatal("der Stil ließ sich nicht zerteilen")
	}

	stil, _ := paket.Stilmodell()

	feld, da := stil.Feld("name")
	if !da {
		t.Fatalf("kein Feldstil für name: %+v", stil.Felder)
	}

	if feld.ZIndex != 3 || feld.Darstellung != "text" || !feld.ImmerSymbole ||
		feld.Symbolschrift.Name != "magic-mana" {
		t.Errorf("mit Unterstrich gelesen: z index %d, render style %q, always symbol %v, symbol font %q",
			feld.ZIndex, feld.Darstellung, feld.ImmerSymbole, feld.Symbolschrift.Name)
	}
}

// **Die Schrift** — `schriftAus`.
func TestEineSchriftLiestIhreSchluesselKanonisch(t *testing.T) {
	t.Parallel()

	paket := paketAusText("card style:\n\tname:\n\t\tfont:\n\t\t\tname: Beleren\n"+
		"\t\t\titalic_name: Beleren Kursiv\n\t\t\tscale_down_to: 6\n", "probe.mse-style")
	if paket == nil {
		t.Fatal("der Stil ließ sich nicht zerteilen")
	}

	stil, _ := paket.Stilmodell()
	feld, _ := stil.Feld("name")

	if feld.Schrift.KursivName != "Beleren Kursiv" || feld.Schrift.Kleinstgroesse != "6" {
		t.Errorf("mit Unterstrich gelesen: italic name %q, scale down to %q",
			feld.Schrift.KursivName, feld.Schrift.Kleinstgroesse)
	}
}
