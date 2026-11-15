// **Eine offene Klammer in einem Feld übersteht Sichern und Wiederöffnen**
// (C02, #2227).
//
// # Der Befund
//
// Gefunden von der Fuzzkampagne an einer Symboldatei, und dieselbe Klasse
// trifft die Setdatei: dieser Zerteiler zählt Klammern, um Skriptausdrücke
// über mehrere Zeilen zu lesen, und eine Zeile `name: Feuer [` öffnete einen
// Ausdruck — jedes Feld dahinter wurde seine Fortsetzung. Eine Karte verlor
// beim Wiederöffnen ihre übrigen Felder, ein Set seinen Titel. MSE2 zählt
// keine Klammern (`src/util/io/reader.cpp`, Zeilen 207–223, Fassung f9d9356
// vom 2021-07-06).
package flaechen_test

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestEineOffeneKlammerUeberstehtSichernUndOeffnen(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	setID := eingefahren(t, umgebung)
	karteID := ersteKarte(t, mussKommandozeile(t, umgebung, "karten", setID, "--json"))

	mussKommandozeile(t, umgebung, "feld", setID, karteID, "name", "Feuer [")

	ziel := filepath.Join(t.TempDir(), "klammer.mse-set")
	mussKommandozeile(t, umgebung, "sichern", setID, ziel)

	wieder := eigenerBestand(t)
	neu := setkennungAus(t, mussKommandozeile(t, wieder, "oeffnen", ziel, "--json"))
	karten := mussKommandozeile(t, wieder, "karten", neu, "--json")

	for _, erwartet := range []string{`Feuer [`, `rot`} {
		if !strings.Contains(karten, erwartet) {
			t.Errorf("%q fehlt nach dem Wiederöffnen:\n%s", erwartet, karten)
		}
	}
}

func TestEinTitelMitOffenerKlammerUeberstehtSichernUndOeffnen(t *testing.T) {
	umgebung := mitEigenerSaat(t)

	const titel = "Probe [Klammer"

	setID := neuesSet(t, umgebung, titel)

	ziel := filepath.Join(t.TempDir(), "titel.mse-set")
	mussKommandozeile(t, umgebung, "sichern", setID, ziel)

	wieder := eigenerBestand(t)
	mussKommandozeile(t, wieder, "oeffnen", ziel, "--json")

	gefunden := false
	for _, zeile := range setzeilenAus(t, mussKommandozeile(t, wieder, "sets", "--json")) {
		gefunden = gefunden || zeile.Name == titel
	}

	if !gefunden {
		t.Errorf("das Set heißt nach dem Wiederöffnen nicht mehr %q:\n%s", titel,
			mussKommandozeile(t, wieder, "sets", "--json"))
	}
}
