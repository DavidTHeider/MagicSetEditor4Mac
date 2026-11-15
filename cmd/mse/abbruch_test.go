// **Ein Signal bricht ab** (#886).
//
// # Warum das eine eigene Zusage ist
//
// Ein Rastern über ein ganzes Set dauert. Wer ⌃C drückt, will, dass es
// aufhört — und nicht, dass die Kommandozeile den Lauf zu Ende bringt und
// **danach** merkt, dass jemand abgebrochen hat.
//
// **Der Weg dorthin ist ein Kontext**, den `main` an das Signal bindet und der
// bis in den Rasterer reicht (`internal/render/rasterer.go:192` fragt ihn je
// Ebene). Geprüft wird hier, dass `lauf` ihn wirklich durchreicht.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

// **Ein abgebrochener Kontext hält jeden Befehl an.**
//
// Nicht nur `sets`: ein Abbruch, der bei einem Befehl greift und beim nächsten
// nicht, ist keiner.
//
// # Warum der Bestand gefüllt ist
//
// **Damit jeder Fall etwas zu verlieren hat.** Der erste Anlauf fuhr gegen
// einen leeren Bestand, und `karten 1` scheiterte dort ohnehin — an einem Set,
// das es nicht gibt. Die Gegenprobe hat es gezeigt: mit neutralisiertem
// Kontext blieben zwei von sechs Fällen **grün**, weil sie aus einem anderen
// Grund scheiterten.
//
// **Und warum es eingefahren und nicht angelegt ist**: ein Set ohne Quelldatei
// lässt sich nicht sichern (die Byte-Zusage aus #48 gilt für ein solches
// nicht), und `sichern` scheiterte dann wieder aus dem falschen Grund. Mit dem
// eingefahrenen Probeset würde **jeder** der sechs Fälle ohne Abbruch
// gelingen — erst dann sagt sein Fehlschlag etwas über den Abbruch (§10).
func TestEinAbgebrochenerKontextHaeltJedenBefehlAn(t *testing.T) {
	ort, setID, karteID := eingefahrenesSet(t)

	faelle := map[string][]string{
		"sets":     {"sets"},
		"karten":   {"karten", setID},
		"feld":     {"feld", setID, karteID, "name", "x"},
		"oeffnen":  {"oeffnen", probeset(t)},
		"sichern":  {"sichern", setID, filepath.Join(t.TempDir(), "x.mse-set")},
		"zeichnen": {"zeichnen", karteID, filepath.Join(t.TempDir(), "x.png")},
	}

	for name, argumente := range faelle {
		t.Run(name, func(t *testing.T) {
			ctx, abbrechen := context.WithCancel(probekontext(t))
			abbrechen()

			var raus, rfehler bytes.Buffer

			code := lauf(ctx, argumente, nil, &raus, &rfehler, ort)
			if code != CodeAbgebrochen {
				t.Errorf("%q meldete %d statt %d (%s)",
					name, code, CodeAbgebrochen, rfehler.String())
			}
		})
	}
}

// **Ein Abbruch heißt 130 und nicht „die Antwort ist nein".**
//
// Die Übereinkunft der Shells ist 128 plus Signalnummer, und SIGINT ist 2. Ein
// Skript, das `until mse zeichnen …; do` schleift, unterscheidet daran, ob es
// weiterprobieren soll — mit `CodeFachlich` liefe es nach ⌃C weiter.
//
// # Was er nicht ist
//
// **Ein Abbruch nach getaner Arbeit bleibt Erfolg.** Wer während des
// Schreibens drückt und trifft den Augenblick nach dem Schreiben, bekommt 0 —
// geschrieben ist geschrieben (§7, Regel 3), und 130 hieße für ein Skript,
// dass nichts passiert sei.
func TestEinAbbruchNachGetanerArbeitBleibtErfolg(t *testing.T) {
	ctx, abbrechen := context.WithCancel(probekontext(t))
	abbrechen()

	if code := abgebrochenOder(ctx, CodeGut); code != CodeGut {
		t.Errorf("aus einem geglückten Lauf wurde %d", code)
	}
}

// **Ein Fehlschlag ohne Abbruch bleibt sein eigener.**
//
// Ohne diesen Fall stünde `abgebrochenOder` auch dann grün da, wenn es jeden
// Fehler in 130 übersetzte.
func TestOhneAbbruchBleibtDerFehlerWerErIst(t *testing.T) {
	if code := abgebrochenOder(probekontext(t), CodeFachlich); code != CodeFachlich {
		t.Errorf("aus %d wurde %d", CodeFachlich, code)
	}
}

// **Die Hilfe braucht keinen Bestand und hört deshalb auch nicht auf ihn.**
//
// Ein Abbruch, der eine Übersicht verhindert, wäre eine Zusage ohne Nutzen.
func TestDieHilfeLaeuftAuchNachEinemAbbruch(t *testing.T) {
	ctx, abbrechen := context.WithCancel(probekontext(t))
	abbrechen()

	var raus, rfehler bytes.Buffer

	if code := lauf(ctx, []string{"hilfe"}, nil, &raus, &rfehler,
		frischerBestand(t)); code != CodeGut {
		t.Errorf("die Hilfe wurde abgebrochen: %d", code)
	}

	if !strings.Contains(raus.String(), "mse") {
		t.Error("die Hilfe kam nicht")
	}
}

// eingefahrenesSet fährt das Probeset ein und gibt Ort, Set- und
// Kartenkennung.
//
// **Eingefahren und nicht angelegt**: nur so hat das Set eine Quelldatei, und
// nur dann kann `sichern` überhaupt gelingen.
func eingefahrenesSet(t *testing.T) (ort, setID, karteID string) {
	t.Helper()

	ort, karteID = eingefahren(t)

	_, aus, _ := fahreAn(t, ort, "sets", "--json")

	var gelesen setsAntwort
	if err := json.Unmarshal([]byte(aus), &gelesen); err != nil || len(gelesen.Sets) == 0 {
		t.Fatalf("die Setliste: %v (%s)", err, aus)
	}

	return ort, gelesen.Sets[0].ID, karteID
}
