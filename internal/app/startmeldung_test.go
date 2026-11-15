// **Ein Startfehler erreicht jemanden** (#797).
package app

import (
	"errors"
	"strings"
	"testing"
)

// **Die Meldung geht an das Fenster, das es noch nicht gibt.**
//
// Vor `wails.Run` gibt es keinen Wails-Kontext, und `runtime.MessageDialog`
// verlangt einen — nachgelesen in der Wails-Dokumentation, nicht aus dem
// Gedächtnis. Deshalb ein eigener Weg, und deshalb als **Naht**: sonst öffnete
// jeder Testlauf ein Fenster.
func TestEinStartfehlerGehtAnDenMelder(t *testing.T) {
	var gesehen []string

	vorher := startmelder
	startmelder = func(titel, text string) error {
		gesehen = append(gesehen, titel+"|"+text)

		return nil
	}

	t.Cleanup(func() { startmelder = vorher })

	Startfehler(errors.New("der Bestand ließ sich nicht öffnen: keine Rechte"))

	if len(gesehen) != 1 {
		t.Fatalf("%d Meldungen, erwartet war eine", len(gesehen))
	}

	// **Der Grund steht drin, nicht nur „Fehler beim Start".** Genau das
	// verlangt der Abnahmekasten: wer die Anwendung nicht starten kann, muss
	// erfahren, woran es liegt.
	if !strings.Contains(gesehen[0], "keine Rechte") {
		t.Errorf("der Grund fehlt: %s", gesehen[0])
	}
}

// **Scheitert auch das Melden, geht es nach stderr** — und nicht ins Leere.
//
// Der Fall ist nicht theoretisch: auf einer anderen Plattform gibt es den Weg
// gar nicht, und dort ist stderr das Einzige, was bleibt.
func TestEinFehlgeschlagenerMelderSchreibtWenigstensNachStderr(t *testing.T) {
	vorher := startmelder
	startmelder = func(string, string) error { return errors.New("kein Weg") }

	t.Cleanup(func() { startmelder = vorher })

	// Kein Absturz, kein Verschlucken: der Aufruf kehrt zurück.
	Startfehler(errors.New("etwas ging schief"))
}

// **Kein Fehler, keine Meldung.**
func TestOhneFehlerWirdNichtsGemeldet(t *testing.T) {
	var gerufen int

	vorher := startmelder
	startmelder = func(string, string) error {
		gerufen++

		return nil
	}

	t.Cleanup(func() { startmelder = vorher })

	Startfehler(nil)

	if gerufen != 0 {
		t.Errorf("%d Meldungen für keinen Fehler", gerufen)
	}
}

// **Und die Meldung sagt, was zu tun ist** — Werkstück 8 von Phase 4.13
// (#1352).
//
// # Der Anlass des Issues, an seiner eigenen Stelle
//
// **Das Issue beginnt mit einem Fehlerdialog**, der `mkdir …: file exists`
// zeigte. Der Fall war ein Entwicklungsartefakt; **die Klasse ist es nicht** —
// und genau diese Klasse steht hier: `Startfehler` zeigte `fehler.Error()`
// roh, in einem nativen Modal, das sich nur mit „OK" wegklicken lässt.
//
// **Am 2026-09-07 abgezählt**, was diesen Weg erreichen kann: `lauf()` gibt
// vor `wails.Run` genau zwei Fehler zurück — `app.Datenbankpfad()` und
// `wails.Run` selbst. Der erste lautet im Ernstfall „Benutzerverzeichnis
// ermitteln: $HOME is not defined".
func TestEinStartfehlerSagtWasZuTunIst(t *testing.T) {
	var gesehen string

	vorher := startmelder
	startmelder = func(_, text string) error {
		gesehen = text

		return nil
	}

	t.Cleanup(func() { startmelder = vorher })

	Startfehler(errors.New("Benutzerverzeichnis ermitteln: $HOME is not defined"))

	// **Der nächste Schritt steht da** — und nicht nur, was schiefging.
	//
	// **Geprüft wird jetzt der Verweis auf die Protokolldatei.** Hier stand
	// „Diagnose", und der Text verwies auf »Hilfe → Diagnose« in einer
	// Anwendung, die in diesem Augenblick beendet wird — ein Widerspruch in
	// acht Zeilen, angestrichen vom Nachlauf zum Zweitblick über Phase 4.13.
	if !strings.Contains(gesehen, Protokollname) {
		t.Errorf("die Meldung nennt keinen erreichbaren nächsten Schritt:\n%s",
			gesehen)
	}

	// **Und der Grund bleibt als Spur darunter.** Wer hier steht, hat kein
	// Fenster und keine Seite mehr — das Protokoll gibt es, aber der rohe
	// Satz ist das, was er ohne Suchen weitergeben kann.
	if !strings.Contains(gesehen, "$HOME is not defined") {
		t.Errorf("der Grund ging verloren:\n%s", gesehen)
	}
}
