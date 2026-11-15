package main

// **Zurückspielen auf einem frischen Rechner** — Befund 7 des Zweitblicks über
// Phase 2 von 0.9.0 (#2323): M02 ließ für Ansehen und Zurückspielen die
// Wanderung weg, und danach fragte die Kommandozeile trotzdem nach Profil und
// Rolle — in einem Bestand, den es noch nicht gab. Der Weg „neuer Mac, alte
// Sicherung" scheiterte damit an der ersten Abfrage. Bis M02 wanderte er den
// leeren Bestand erst und spielte dann zurück.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestZurueckAufEinemFrischenRechner(t *testing.T) {
	alt := frischerBestand(t)

	if code, _, fehler := fahreAn(t, alt, "sets"); code != CodeGut {
		t.Fatalf("anlegen: %d %s", code, fehler)
	}

	sicherung := filepath.Join(t.TempDir(), "vom-alten-mac.msesicherung")
	if code, _, fehler := fahreAn(t, alt, "bestand-sichern", sicherung); code != CodeGut {
		t.Fatalf("sichern: %d %s", code, fehler)
	}

	// **Kein Bestand am Ort** — der neue Rechner.
	neu := filepath.Join(t.TempDir(), "neuer-mac", "bestand.db")

	code, aus, fehler := fahreAn(t, neu, "--json", "bestand-schauen", sicherung)
	if code != CodeGut {
		t.Fatalf("auf dem frischen Rechner lässt sich die Sicherung nicht ansehen: %d %s", code, fehler)
	}

	var schau struct {
		AngelegtAm string `json:"angelegtAm"`
	}

	if err := json.Unmarshal([]byte(aus), &schau); err != nil {
		t.Fatalf("%v: %s", err, aus)
	}

	if code, _, fehler := fahreAn(t, neu, "bestand-zurueckspielen", sicherung, schau.AngelegtAm); code != CodeGut {
		t.Fatalf("auf dem frischen Rechner spielt sie nicht zurück: %d %s", code, fehler)
	}

	if _, err := os.Stat(neu); err != nil {
		t.Fatalf("danach liegt kein Bestand da: %v", err)
	}

	if code, _, fehler := fahreAn(t, neu, "sets"); code != CodeGut {
		t.Errorf("danach öffnet die heutige Fassung ihn nicht: %d %s", code, fehler)
	}
}
