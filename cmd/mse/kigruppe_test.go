package main

// **Die atomare Gruppe an der Kommandozeile** — Werkstück A13.1b von Phase
// 5.3 (#1452).
//
// **Die fachliche Zusage prüft `internal/flaechen`** — dort läuft derselbe
// Ablauf über alle drei Flächen. Hier stehen die **Bedienfehler** und die
// Gestalt der Ausgabe.

import (
	"bytes"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Ohne Argument ist es ein Bedienfehler.**
func TestEineGruppeOhneArgumentIstEinBedienfehler(t *testing.T) {
	ort, _ := eingefahren(t)

	if code, _, _ := fahreAn(t, ort, "ki-gruppe"); code != CodeBedienung {
		t.Errorf("Rückgabewert %d, erwartet %d", code, CodeBedienung)
	}
}

// **Und ein Argument ohne Gleichheitszeichen ebenso.**
//
// **Je Fall eine eigene Zeile**: „kennung ohne Felder" und „Felder ohne
// Kennung" sind zwei Vertipper, und wer nur einen prüft, hat den anderen
// offen.
func TestEinKaputtesGruppenargumentIstEinBedienfehler(t *testing.T) {
	ort, _ := eingefahren(t)

	for name, argument := range map[string]string{
		"ohne Gleichheitszeichen": "a1",
		"ohne Felder":             "a1=",
		"ohne Kennung":            "=name",
	} {
		t.Run(name, func(t *testing.T) {
			code, _, fehler := fahreAn(t, ort, "ki-gruppe", argument)
			if code != CodeBedienung {
				t.Errorf("Rückgabewert %d, erwartet %d", code, CodeBedienung)
			}

			if !strings.Contains(fehler, "kennung=feld") {
				t.Errorf("die Klage nennt die Gestalt nicht: %q", fehler)
			}
		})
	}
}

// **Einen Auftrag, den es nicht gibt, meldet sie fachlich.**
//
// **Und nicht als Bedienfehler**: die Gestalt stimmt, nur der Inhalt nicht —
// das ist der Unterschied, den ein Skript am Rückgabewert abliest.
func TestEineGruppeMitUnbekanntemAuftragIstFachlich(t *testing.T) {
	ort, _ := eingefahren(t)

	if code, _, _ := fahreAn(t, ort, "ki-gruppe",
		"gibtesnicht=name"); code != CodeFachlich {
		t.Errorf("Rückgabewert %d, erwartet %d", code, CodeFachlich)
	}
}

// **Die Ausgabe nennt je Karte ihre Felder.**
//
// # Warum sie direkt gerufen wird
//
// **Weil der Weg dorthin ein antwortendes Modell braucht** — das prüft
// `internal/flaechen` mit einer nachgestellten Laufzeit. Hier steht die
// **Gestalt** der Ausgabe, und die ist eine Frage an diese Funktion allein.
//
// **Und sie ist eine Zusage**: eine Gruppe, die nur „fertig" sagt,
// verschweigt, welches Feld wo gelandet ist. Weil sie ganz oder gar nicht
// geht, ist jede Zeile hier eine, die wirklich steht.
func TestDieGruppenausgabeNenntJedeKarte(t *testing.T) {
	var raus, rfehler bytes.Buffer

	an := &schreiber{ziel: &raus, fehler: &rfehler}

	if code := gruppenzeilen(an, app.KIGruppenantwort{
		Karten: 2, Felder: 3,
		Namen: []app.KIGruppenkarteDTO{
			{KarteID: 4, Felder: []string{"name", "text"}},
			{KarteID: 7, Felder: []string{"name"}},
		},
	}); code != CodeGut {
		t.Errorf("Rückgabewert %d, erwartet %d", code, CodeGut)
	}

	for _, teil := range []string{
		"2 Karten", "3 Felder", "4\tname, text", "7\tname",
	} {
		if !strings.Contains(raus.String(), teil) {
			t.Errorf("die Ausgabe nennt %q nicht:\n%s", teil, raus.String())
		}
	}
}
