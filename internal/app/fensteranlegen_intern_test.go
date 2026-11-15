package app

// **Auch im Fenster rechnet eine neue Karte und eine Massenänderung** — #2516.
// Das Fenster hat seinen eigenen Composition Root (`verdrahtung.go`); bis
// dahin legte es am Feldweg vorbei an (`a.anleger` war der Bestand), und eine
// Fläche, die anders anlegt als die übrigen, sähe niemand.

import (
	"strconv"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestDasFensterRechnetEineNeueKarteUndEineMassenaenderung(t *testing.T) {
	anwendung, setID, karteID, werte := nachzugsfenster(t)

	angelegt := anwendung.KarteAnlegen(probekontext(t), setID)
	if angelegt.Fehler != nil {
		t.Fatalf("anlegen: %+v", angelegt.Fehler)
	}

	neue, _ := strconv.ParseInt(angelegt.KarteID, 10, 64) //nolint:errcheck // vom Anlegen

	karte, fehler := storage.NeuerSetLeser(anwendung.db).KarteLesen(probekontext(t), neue)
	if fehler != nil || karte == nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if zeile, _ := karte.Value("zeile"); zeile != "[]" {
		t.Errorf("die neue Karte trägt zeile=%q, erwartet []", zeile)
	}

	if echo, _ := karte.Value("echo"); echo != "[]!" {
		t.Errorf("die neue Karte trägt echo=%q, erwartet []!", echo)
	}

	// **Wie das Fenster: erst die Vorschau, dann mit ihrer Erwartung** — und
	// die Vorschau zeigt schon, was gespeichert wird.
	vorschau := anwendung.FelderVorschau(probekontext(t), setID, []string{strconv.FormatInt(karteID, 10)}, "zeile", "Asche")
	if vorschau.Fehler != nil || len(vorschau.Zeilen) != 1 || vorschau.Zeilen[0].Neu != "[Asche]" {
		t.Fatalf("Vorschau: %+v", vorschau)
	}

	zeile := vorschau.Zeilen[0]
	if masse := anwendung.FelderSetzenMitErwartung(probekontext(t), setID,
		[]FeldvorwertDTO{{KarteID: zeile.KarteID, Wert: zeile.Alt, GabEs: zeile.GabAlt}}, "zeile", "Asche"); masse.Fehler != nil {
		t.Fatalf("Massenänderung: %+v", masse.Fehler)
	}

	if gelesen := werte(); gelesen["zeile"] != "[Asche]" || gelesen["echo"] != "[Asche]!" {
		t.Errorf("nach der Massenänderung zeile=%q, echo=%q — erwartet [Asche] und [Asche]!", gelesen["zeile"], gelesen["echo"])
	}
}
