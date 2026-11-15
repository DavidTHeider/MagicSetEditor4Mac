package app

// **Eine Panik beim Öffnen ist ein Fehler, über jeden Weg derselbe** — F02
// (#2341),
// Befund A-02.
//
// # Warum
//
// Über die Bindung fing Wails eine Panik und gab dem Fenster einen Satz mit
// Paketpfad; über den Finder (`DateiOeffnen`) fing sie niemand, und Wails'
// Standardweg beendete den Prozess. Zwei Wege zu derselben Handlung, zwei
// Ausgänge — und einer davon ohne Fenster.
//
// # Was die Prüfungen halten
//
// Beide Wege geben denselben Code mit Kennung, und die Sperre des Tauschs
// ist danach frei. **Die Panik ist echt**: ohne Importeur greift das Öffnen
// ins Leere, wie es ein Programmierfehler täte.

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

func TestEinePanikBeimOeffnenIstUeberBeideWegeDerselbeFehler(t *testing.T) {
	a, _ := finderprobe(t)
	fehler := make(chan *Fehler, 1)
	a.MitMeldung(func(_ context.Context, ereignis string, daten ...any) {
		if ereignis == EreignisDateiFehler && len(daten) == 2 {
			if gemeldet, ist := daten[1].(*Fehler); ist {
				fehler <- gemeldet
			}
		}
	})

	a.importeur = nil
	pfad := filepath.Join("..", "..", "testdaten", "probe.mse-set")

	ueberBindung := a.SetOeffnen(t.Context(), pfad)
	if ueberBindung.Fehler == nil || ueberBindung.Fehler.Code != CodeUnbekannt || ueberBindung.Fehler.Vorgang == "" {
		t.Fatalf("über die Bindung: %+v, erwartet war %q mit Kennung", ueberBindung.Fehler, CodeUnbekannt)
	}

	a.DateiOeffnen(pfad)

	select {
	case ueberFinder := <-fehler:
		if ueberFinder.Code != ueberBindung.Fehler.Code || ueberFinder.Vorgang == "" {
			t.Errorf("über den Finder %+v, über die Bindung %+v — zwei Klassen", ueberFinder, ueberBindung.Fehler)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("über den Finder kam keine Meldung")
	}

	if !a.lauf.wechsel.TryLock() {
		t.Fatal("die Sperre des Tauschs ist nach der Panik noch gehalten")
	}

	a.lauf.wechsel.Unlock()
}
