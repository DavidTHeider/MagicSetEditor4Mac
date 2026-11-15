package service

// **Ein Feldstil trägt den Namen, unter dem sein Feld erklärt ist** —
// #1980.
//
// MSE macht beide Seiten kanonisch: den Feldnamen in `Field::after_reading`
// (`src/data/field.cpp`, Zeile 68) und jeden Schlüssel beim Lesen
// (`src/util/io/reader.cpp`, Zeile 217; Fassung f9d9356 vom 2021-07-06).
// Deshalb trifft `card style: rule_text:` das Feld `rule text`. Hier blieb
// der Schlüssel roh, und jeder Vergleich danach — Feldart, Feldfläche fürs
// Fenster, Stiloption — ging an ihm vorbei.
//
// Gemessen am 2026-10-01 über `full-magic-pack/data`: sechs Stilschlüssel
// mit Unterstrich, darunter `rule_text` (`magic-baseball-1980-topps`) und
// `flavor_text`, `level_11_text` (`magic-m15-secret-lair-gatewatch`).

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEinFeldstilTraegtDenErklaertenNamen(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		name, feld, schluessel, erwartet string
	}{
		{"Unterstrich im Stil, Leerzeichen im Spiel", "rule text", "rule_text", "rule text"},
		{"Leerzeichen im Stil, Unterstrich im Spiel", "rule_text", "rule text", "rule_text"},
		{"gleich geschrieben", "rule_text", "rule_text", "rule_text"},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			wurzel, speicher := saatplatz(t)
			zeileErsetzen(t, filepath.Join(wurzel, "probe.mse-game", "game"),
				"\n\tname: text\n", "\n\tname: "+fall.feld+"\n")
			zeileErsetzen(t, filepath.Join(wurzel, "probe-schlicht.mse-style", "style"),
				"\n\ttext:\n", "\n\t"+fall.schluessel+":\n")

			vorlage, befunde, da := speicher.StilFuerSet(t.Context(), 7)
			if !da {
				t.Fatalf("kein Stil, Befunde %v", befunde)
			}

			if !hatFeldstil(vorlage, fall.erwartet) {
				t.Errorf("kein Feldstil %q unter %v", fall.erwartet, feldstilnamen(vorlage))
			}
		})
	}
}

// **Dasselbe für die Sichtseite einer Stiloption** — `styling style:` zu
// `styling field:`, dieselbe Gestalt wie `card style:` zu `card field:`.
func TestEinOptionsstilTraegtDenErklaertenNamen(t *testing.T) {
	t.Parallel()

	wurzel, speicher := saatplatz(t)
	zeileErsetzen(t, filepath.Join(wurzel, "probe-schlicht.mse-style", "style"),
		"\n\trahmen zeigen:\n", "\n\trahmen_zeigen:\n")

	vorlage, befunde, da := speicher.StilFuerSet(t.Context(), 7)
	if !da {
		t.Fatalf("kein Stil, Befunde %v", befunde)
	}

	if _, gefunden := vorlage.Stil.Optionsstil("rahmen zeigen"); !gefunden {
		t.Error("die Stiloption `rahmen zeigen` findet ihre Sichtseite nicht")
	}
}

// **Und für die Zusatzfelder** — `extra card style:` zu `extra card field:`.
func TestEinZusatzstilTraegtDenErklaertenNamen(t *testing.T) {
	t.Parallel()

	wurzel, speicher := saatplatz(t)
	stil := filepath.Join(wurzel, "probe-schlicht.mse-style", "style")
	zeileErsetzen(t, stil, "\n\tname: rahmen\n", "\n\tname: rahmen farbe\n")
	zeileErsetzen(t, stil, "\n\trahmen:\n", "\n\trahmen_farbe:\n")

	vorlage, befunde, da := speicher.StilFuerSet(t.Context(), 7)
	if !da {
		t.Fatalf("kein Stil, Befunde %v", befunde)
	}

	for _, zusatz := range vorlage.Stil.Zusatzstile {
		if zusatz.Feld == "rahmen farbe" {
			return
		}
	}

	t.Errorf("kein Zusatzstil `rahmen farbe` unter %+v", vorlage.Stil.Zusatzstile)
}

func hatFeldstil(vorlage Vorlage, name string) bool {
	for _, feld := range vorlage.Stil.Felder {
		if feld.Feld == name {
			return true
		}
	}

	return false
}

func feldstilnamen(vorlage Vorlage) []string {
	namen := make([]string, 0, len(vorlage.Stil.Felder))
	for _, feld := range vorlage.Stil.Felder {
		namen = append(namen, feld.Feld)
	}

	return namen
}

// saatplatz legt Spiel und Stil der Saat in eine eigene Wurzel — wie
// `stempelplatz`, aber ohne dessen Pause umzustellen: die ist ein Paketwert,
// und parallele Läufe teilten ihn.
func saatplatz(t *testing.T) (wurzel string, speicher *Stilspeicher) {
	t.Helper()

	wurzel = t.TempDir()
	for _, name := range []string{"probe.mse-game", "probe-schlicht.mse-style"} {
		if err := os.CopyFS(filepath.Join(wurzel, name), os.DirFS(filepath.Join("..", "..", "saat", name))); err != nil {
			t.Fatal(err)
		}
	}

	return wurzel, NeuerStilspeicher(stempelkoepfe{}, wurzel)
}
