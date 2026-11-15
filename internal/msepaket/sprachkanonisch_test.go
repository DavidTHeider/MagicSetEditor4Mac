package msepaket_test

// **Ein Sprachpaket trifft Spiel und Feld in beiden Schreibweisen** —
// #1980.
//
// MSE liest jeden Schlüssel kanonisch (`src/util/io/reader.cpp`, Zeile 217)
// und schlägt kanonisch nach (`SubLocale::tr`, `src/data/locale.cpp`,
// Zeilen 82 und 90; Fassung f9d9356 vom 2021-07-06). Die Pakete im Pack
// schreiben Leerzeichen (`card name:`), ein Feld darf `card_name` heißen.

import "testing"

func TestEinSprachpaketTrifftBeideSchreibweisen(t *testing.T) {
	t.Parallel()

	paket := paketAusText("short name: Probe\n"+
		"game:\n"+
		"\tprobe_spiel:\n"+
		"\t\tcard name: Kartenname\n"+
		"\t\trule_text: Regeltext\n", "xx.mse-locale")
	if paket == nil {
		t.Fatal("das Sprachpaket ließ sich nicht zerteilen")
	}

	sprache := paket.Sprachmodell()

	for _, fall := range []struct {
		name, spiel, feld, erwartet string
	}{
		{"Leerzeichen in der Datei, Unterstrich im Feld", "probe_spiel", "card_name", "Kartenname"},
		{"Unterstrich in der Datei, Leerzeichen im Feld", "probe_spiel", "rule text", "Regeltext"},
		{"das Spiel mit Leerzeichen", "probe spiel", "card name", "Kartenname"},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			if gefunden := sprache.Feldbeschriftung(fall.spiel, fall.feld); gefunden != fall.erwartet {
				t.Errorf("Beschriftung %q, erwartet %q", gefunden, fall.erwartet)
			}
		})
	}
}
