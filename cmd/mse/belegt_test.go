package main

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ein belegter Bestand zeigt einen Rat, keine Treibermeldung**
// (#886).
//
// Gemessen am 2026-08-23 gegen ein Fenster, das länger als `busy_timeout`
// sperrt: die Kommandozeile gab
//
//	Migration beginnen: der Bestand ist gerade belegt: database is locked (5) (SQLITE_BUSY)
//
// Drei Lagen, die vorderste falsch — es ist keine Migration kaputt — und die
// hinterste eine rohe SQLite-Meldung, die §8 an keiner Oberfläche zulässt.
//
// # Was er nicht prüft
//
// **Ob dieser Weg den Fehler wirklich hervorbringt.** Das steht in
// `internal/storage/belegt_test.go` an einer echten Sperre. Hier steht die
// Darstellung, und der eingesetzte Fehler ist derselbe, den die Abbildung
// dort erzeugt.
func TestEinBelegterBestandZeigtEinenRat(t *testing.T) {
	t.Parallel()

	var raus, rfehler bytes.Buffer

	an := &schreiber{ziel: &raus, fehler: &rfehler}
	an.klageUeber(fmt.Errorf("Migration beginnen: %w: database is locked (5) (SQLITE_BUSY)",
		domain.ErrBelegt))

	gesagt := rfehler.String()

	if strings.Contains(gesagt, "SQLITE_BUSY") ||
		strings.Contains(gesagt, "database is locked") {
		t.Errorf("die Treibermeldung steht in der Ausgabe: %q", gesagt)
	}

	if strings.Contains(gesagt, "Migration") {
		t.Errorf("die Ausgabe zeigt auf eine Migration: %q", gesagt)
	}

	if !strings.Contains(gesagt, "belegt") {
		t.Errorf("die Ausgabe nennt den Grund nicht: %q", gesagt)
	}

	// **Der Rat gehört dazu.** Ein Grund ohne nächsten Schritt ist an einer
	// Kommandozeile so wenig wert wie „Keine Daten" auf einer leeren Seite
	// (§6).
	if !strings.Contains(gesagt, "noch einmal") {
		t.Errorf("die Ausgabe sagt nicht, was zu tun ist: %q", gesagt)
	}
}

// **Jeder andere Fehler bleibt, wie er ist.**
//
// Ohne diesen Fall stünde `klageUeber` auch dann grün da, wenn es jede Meldung
// durch denselben Satz ersetzte.
func TestEinAndererFehlerWirdUnveraendertGezeigt(t *testing.T) {
	t.Parallel()

	var raus, rfehler bytes.Buffer

	an := &schreiber{ziel: &raus, fehler: &rfehler}
	an.klageUeber(fmt.Errorf("die Karte lesen: %w", domain.ErrNotFound))

	if !strings.Contains(rfehler.String(), "die Karte lesen") {
		t.Errorf("der Zusammenhang ging verloren: %q", rfehler.String())
	}
}
