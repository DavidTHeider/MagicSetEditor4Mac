// **Der Austausch nimmt sich zurück** — der Befund des Zweitblicks durch
// Codex über Phase 3.3.
//
// # Der Befund
//
// Der Tausch besteht aus zwei Umbenennungen: erst geht das Ziel in die
// Sicherung, dann die Nachbardatei an seinen Platz. Scheiterte nur die
// **zweite**, war die erste schon getan — das Ziel gab es nicht mehr, die
// Nachbardatei räumte der Aufrufer weg, und übrig blieb ein Paket **ohne
// diese Datei**, während ihr voriger Stand unbemerkt unter `.bak` lag.
//
// # Warum das ein weißer Lauf ist
//
// Einen Fehlschlag genau **zwischen** zwei Umbenennungen liefert kein
// Dateisystem auf Bestellung. Deshalb nimmt [austauschen] die Umbenennung
// entgegen; hier kommt eine, die beim zweiten Zug scheitert und sonst die
// echte tut. Was danach geprüft wird, ist die **Wirkung** auf der Platte und
// nicht die Buchführung darüber.
package msepaket

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// dreiDateien legt Ziel, Sicherungsort und Nachbardatei an.
func dreiDateien(t *testing.T) (ziel, sicherung, neben string) {
	t.Helper()

	ort := t.TempDir()
	ziel = filepath.Join(ort, "style")
	sicherung = filepath.Join(ort, "style.bak")
	neben = filepath.Join(ort, ".mse-schreibt-1")

	for pfad, inhalt := range map[string]string{ziel: "alt\n", neben: "neu\n"} {
		if err := os.WriteFile(pfad, []byte(inhalt), 0o600); err != nil {
			t.Fatalf("anlegen: %v", err)
		}
	}

	return ziel, sicherung, neben
}

// **Scheitert der zweite Zug, steht der vorige Stand wieder am Platz.**
func TestEinGescheiterterAustauschHoltDenVorigenStandZurueck(t *testing.T) {
	t.Parallel()

	ziel, sicherung, neben := dreiDateien(t)

	verweigert := errors.New("der zweite Zug scheitert")
	zuege := 0

	fehler := austauschen(ziel, sicherung, neben, func(alt, neu string) error {
		zuege++
		if zuege == 2 {
			return verweigert
		}

		return os.Rename(alt, neu)
	})

	if !errors.Is(fehler, verweigert) {
		t.Fatalf("der Fehler ist %v — er soll den zweiten Zug nennen", fehler)
	}

	da, lesefehler := os.ReadFile(ziel) // #nosec G304 -- aus t.TempDir()
	if lesefehler != nil {
		t.Fatalf("das Ziel ist weg: %v", lesefehler)
	}

	if string(da) != "alt\n" {
		t.Errorf("am Platz steht %q", da)
	}

	// **Und die Sicherung ist wieder leer.** Ein zurückgeholter Stand, der
	// zusätzlich unter `.bak` läge, sähe wie ein halber Tausch aus.
	if _, err := os.Stat(sicherung); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("die Sicherung liegt noch da: %v", err)
	}
}

// **Geht auch das Zurückholen schief, sagt die Meldung, wo der Stand liegt.**
//
// Das ist der Fall, in dem nichts mehr zu retten ist — und dann ist die
// einzige Hilfe der Ort, an dem der vorige Stand zu finden ist.
func TestEinGescheitertesZurueckholenNenntDenOrt(t *testing.T) {
	t.Parallel()

	ziel, sicherung, neben := dreiDateien(t)

	zuege := 0

	fehler := austauschen(ziel, sicherung, neben, func(alt, neu string) error {
		zuege++
		if zuege >= 2 {
			return errors.New("geht nicht")
		}

		return os.Rename(alt, neu)
	})

	if fehler == nil {
		t.Fatal("der Austausch meldete Erfolg")
	}

	if !strings.Contains(fehler.Error(), sicherung) {
		t.Errorf("die Meldung nennt %q nicht: %v", sicherung, fehler)
	}
}

// **Scheitert schon der erste Zug, bleibt alles, wie es war.**
func TestEinGescheiterterErsterZugLaesstAllesStehen(t *testing.T) {
	t.Parallel()

	ziel, sicherung, neben := dreiDateien(t)

	verweigert := errors.New("der erste Zug scheitert")

	fehler := austauschen(ziel, sicherung, neben,
		func(string, string) error { return verweigert })

	if !errors.Is(fehler, verweigert) {
		t.Fatalf("der Fehler ist %v", fehler)
	}

	da, lesefehler := os.ReadFile(ziel) // #nosec G304 -- aus t.TempDir()
	if lesefehler != nil || string(da) != "alt\n" {
		t.Errorf("am Platz steht %q (%v)", da, lesefehler)
	}
}
