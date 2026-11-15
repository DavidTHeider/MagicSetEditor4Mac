package schriften_test

import (
	"os"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schriften"
)

// **Ein Paket, das während der Sitzung dazukommt, wird gefunden** — auch unter
// seinem **eigenen** Namen.
//
// Der Zweitblick über Phase G hat den Fall gefunden, und er ist seit
// #263 der
// vorgesehene Weg: Anwendung starten, Karte öffnen, das M15-Pack über die
// Bezugsquelle holen, dieselbe Karte wieder öffnen.
//
// Der Kommentar an `paketnamensindex` sagte, ein neues Paket werde „über den
// Dateinamensweg trotzdem sofort gefunden". **Das gilt nur, wenn der
// Schriftname der Dateiname ist** — `magic-m15.mse-style` verlangt
// `Beleren Bold`, und die Datei heißt `beleren-bold_P1.01.ttf`. Genau darum
// ging es in
// #502.
//
// Damit stand ein **leerer** Index für die ganze Sitzung fest, und die Karte
// blieb bis zum Neustart unbeschriftet.
func TestEinePaketschriftWirdAuchNachDemBezugGefunden(t *testing.T) {
	wurzel := t.TempDir()
	quelle, name := ersteSystemschrift(t)

	roh, err := os.ReadFile(quelle) //nolint:gosec // Pfad aus der Verzeichnisliste
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	suche := schriften.Neue(nil).MitPaketen(wurzel)

	// **Der erste Blick geht ins Leere** — genau wie vor dem Bezug. Er füllt
	// den Index, und bis hierher blieb der leer gemerkt.
	if _, err := suche.Suche(name, false); err == nil {
		t.Fatalf("die Schrift %q war schon da", name)
	}

	// **Jetzt kommt das Paket dazu**, unter einem Dateinamen, der **nicht**
	// der Schriftname ist.
	if err := schreibe(t, wurzel, "irgendwas_v2.ttf", roh); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	if _, err := suche.Suche(name, false); err != nil {
		t.Errorf("die Schrift %q wurde nach dem Bezug nicht gefunden: %v", name, err)
	}
}
