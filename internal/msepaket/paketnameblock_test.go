package msepaket_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Die Namen eines Pakets lesen sich auch als Block** (C02,
// #2227).
//
// [msepaket.Schreibe] setzt einen Wert, der eine Klammer offen lässt, als
// Block — und `full name` ist freier Text aus der Paketanlage. MSE2 liest
// jeden Wert in beiden Formen (`src/util/io/reader.cpp`, Zeilen 258–272,
// Fassung f9d9356 vom 2021-07-06).
//
// **`game:` liest sich nicht so**: in einer `.mse-locale` ist es ein
// Abschnitt und kein Text — als Block gelesen bekäme ein Sprachpaket seine
// Übersetzungen als Spielnamen (gemessen über den vollen Pack, 2026-10-11).
func TestDieNamenEinesPaketsLesenSichAuchAlsBlock(t *testing.T) {
	t.Parallel()

	wurzel := filepath.Join(t.TempDir(), "probe.mse-game")
	if err := os.MkdirAll(wurzel, 0o750); err != nil {
		t.Fatal(err)
	}

	inhalt := "mse version: 2.0.0\nshort name:\n\tKurz [\nfull name:\n\tVoll [\n"
	if err := os.WriteFile(filepath.Join(wurzel, "game"), []byte(inhalt), 0o600); err != nil {
		t.Fatal(err)
	}

	paket, err := msepaket.Oeffne(wurzel)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	if paket.KurzerName != "Kurz [" || paket.VollerName != "Voll [" {
		t.Errorf("Namen %q / %q", paket.KurzerName, paket.VollerName)
	}
}
