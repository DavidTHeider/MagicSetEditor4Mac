// **Die Vorgabe einer Textoption ist ein Skript** — FMT-02 (#1691).
//
// MSE2 liest `default:` eines Textfelds als Skript (`OptionalScript`,
// `REFLECT_N("default", default_script)` in `src/data/field/text.cpp` Zeile 34,
// Fassung f9d9356 vom 2021-07-06): eine Zeichenkette steht dort in
// Anführungszeichen, und ihr Wert ist ihr Inhalt. Ein ungequotetes Wort wie
// `default: Custom` im Pack wäre dort ein Name; es bleibt hier wörtlich stehen,
// denn ein Leser ohne Auswerter hat keinen Namen, den er nachschlagen könnte.
package msepaket_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

func TestDieVorgabeEinerTextoptionIstIhreZeichenkette(t *testing.T) {
	for wert, erwartet := range map[string]string{
		`"Meine Karten"`:        "Meine Karten",
		`"mit \"Zitat\""`:       `mit "Zitat"`,
		`Custom`:                "Custom",
		`"nur vorn`:             `"nur vorn`,
		`"eine \\ Rückschräge"`: `eine \ Rückschräge`,
	} {
		ort := filepath.Join(t.TempDir(), "vorgabe.mse-export-template")
		if err := os.MkdirAll(ort, 0o750); err != nil {
			t.Fatalf("anlegen: %v", err)
		}

		text := "mse version: 2.0.0\ngame: probe\nshort name: Vorgabe\n\noption field:\n\ttype: text\n\tname: kopf\n\tdefault: " + wert + "\n"
		if err := os.WriteFile(filepath.Join(ort, "export-template"), []byte(text), 0o600); err != nil {
			t.Fatalf("schreiben: %v", err)
		}

		paket, err := msepaket.Oeffne(ort)
		if err != nil {
			t.Fatalf("öffnen: %v", err)
		}

		optionen := paket.Exportmodell().Optionen
		_ = paket.Schliessen() //nolint:errcheck // nur gelesen

		if len(optionen) != 1 || optionen[0].Vorgabe != erwartet {
			t.Errorf("%s: %+v, erwartet %q", wert, optionen, erwartet)
		}
	}
}
