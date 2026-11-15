// **Ein Stilspeicher, der eine ungespeicherte Datei sieht** — Werkstück 3 von
// Phase 3.3 (#1164).
package service_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// stilpaket legt ein Stilpaket samt einem Set an, das es benutzt.
func stilpaket(t *testing.T) (wurzel, pfad string) {
	t.Helper()

	wurzel = t.TempDir()
	pfad = filepath.Join(wurzel, "probe-eins.mse-style")

	if err := os.MkdirAll(pfad, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	inhalt := "mse version: 2.0.0\ngame: probe\ncard width: 375\ncard height: 100\n"

	if err := os.WriteFile(filepath.Join(pfad, "style"), []byte(inhalt), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	return wurzel, pfad
}

// **Der Ersatz kommt in der Vorlage an.**
//
// Das ist der ganze Zweck: der Stilspeicher öffnet dieses **eine** Paket mit
// der ersetzten Datei, und was daraus wird, ist die Vorlage der Vorschau.
func TestDerErsatzKommtInDerVorlageAn(t *testing.T) {
	t.Parallel()

	wurzel, pfad := stilpaket(t)

	speicher := service.NeuerStilspeicher(
		&koepfequelle{koepfe: []domain.SetHead{{ID: 1, Game: "probe", Style: "eins"}}}, wurzel)

	vorlage, _, da := speicher.StilFuerSet(t.Context(), 1)
	if !da {
		t.Fatal("das Stilpaket ging nicht auf")
	}

	if vorlage.Stil.Breite != 375 {
		t.Fatalf("die Breite ist %d", vorlage.Stil.Breite)
	}

	// **Ein eigener Speicher mit dem Ersatz** — der vorhandene bleibt, wie er ist.
	mitErsatz := speicher.MitErsatz(service.Paketersatz{
		Paket: pfad, Datei: "style",
		Inhalt: []byte("mse version: 2.0.0\ngame: probe\ncard width: 999\ncard height: 100\n"),
	})

	geaendert, _, auch := mitErsatz.StilFuerSet(t.Context(), 1)
	if !auch {
		t.Fatal("das ersetzte Stilpaket ging nicht auf")
	}

	if geaendert.Stil.Breite != 999 {
		t.Errorf("die Breite ist %d — der Ersatz kam nicht an", geaendert.Stil.Breite)
	}
}

// **Der vorhandene Speicher merkt sich den Ersatz nicht.**
//
// Die Gegenprobe: ohne sie sagte der Lauf darüber auch dann ja, wenn der
// Ersatz in den gemeinsamen Zwischenspeicher liefe — und dann zeichnete jede
// Karte nach einer Vorlage, die es auf der Platte nicht gibt.
func TestDerGemeinsameSpeicherBleibtUnberuehrt(t *testing.T) {
	t.Parallel()

	wurzel, pfad := stilpaket(t)

	speicher := service.NeuerStilspeicher(
		&koepfequelle{koepfe: []domain.SetHead{{ID: 1, Game: "probe", Style: "eins"}}}, wurzel)

	mitErsatz := speicher.MitErsatz(service.Paketersatz{
		Paket: pfad, Datei: "style",
		Inhalt: []byte("mse version: 2.0.0\ngame: probe\ncard width: 999\ncard height: 100\n"),
	})

	if _, _, da := mitErsatz.StilFuerSet(t.Context(), 1); !da {
		t.Fatal("das ersetzte Stilpaket ging nicht auf")
	}

	vorlage, _, da := speicher.StilFuerSet(t.Context(), 1)
	if !da {
		t.Fatal("das Stilpaket ging nicht auf")
	}

	if vorlage.Stil.Breite != 375 {
		t.Errorf("der Ersatz landete im gemeinsamen Speicher: %d", vorlage.Stil.Breite)
	}
}

// **Ein Ersatz auf ein anderes Paket lässt dieses unberührt.**
//
// Verglichen wird der **Pfad**: wer eine Vorlage bearbeitet, ändert nicht jede
// andere mit, die zufällig geöffnet wird.
func TestEinErsatzAufEinAnderesPaketGreiftNicht(t *testing.T) {
	t.Parallel()

	wurzel, _ := stilpaket(t)

	speicher := service.NeuerStilspeicher(
		&koepfequelle{koepfe: []domain.SetHead{{ID: 1, Game: "probe", Style: "eins"}}}, wurzel)

	mitErsatz := speicher.MitErsatz(service.Paketersatz{
		Paket: filepath.Join(wurzel, "probe-zwei.mse-style"), Datei: "style",
		Inhalt: []byte("card width: 999\n"),
	})

	vorlage, _, da := mitErsatz.StilFuerSet(t.Context(), 1)
	if !da {
		t.Fatal("das Stilpaket ging nicht auf")
	}

	if vorlage.Stil.Breite != 375 {
		t.Errorf("der fremde Ersatz griff: %d", vorlage.Stil.Breite)
	}
}

// **Ein unlesbarer Ersatz lässt das Paket nicht aufgehen — mit Befund.**
//
// Wer beim Tippen eine Zeile zerlegt, bekommt den Grund und nicht das Bild von
// vorhin (§8).
func TestEinUnlesbarerErsatzGibtEinenBefund(t *testing.T) {
	t.Parallel()

	wurzel, pfad := stilpaket(t)

	speicher := service.NeuerStilspeicher(
		&koepfequelle{koepfe: []domain.SetHead{{ID: 1, Game: "probe", Style: "eins"}}}, wurzel)

	mitErsatz := speicher.MitErsatz(service.Paketersatz{
		Paket: pfad, Datei: "style", Inhalt: []byte(": ohne Schlüssel\n"),
	})

	_, befunde, da := mitErsatz.StilFuerSet(t.Context(), 1)
	if da {
		t.Fatal("der unlesbare Ersatz kam durch")
	}

	if len(befunde) == 0 {
		t.Fatal("es gab keinen Befund")
	}

	gesehen := false

	for _, befund := range befunde {
		if strings.Contains(befund.Meldung, "Schlüssel") {
			gesehen = true
		}
	}

	if !gesehen {
		t.Errorf("kein Befund nennt den Grund: %+v", befunde)
	}
}
