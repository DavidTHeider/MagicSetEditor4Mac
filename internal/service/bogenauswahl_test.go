package service_test

import (
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Eine Teilmenge auf den Druckbogen** — Werkstück 4 von Phase 3.6
// (#1167).
//
// # Warum das im Dienst geprüft wird und nicht am Fenster
//
// **Weil „welche Karten gehen hinaus?" eine fachliche Frage ist** (§2). Stünde
// der Filter in der Fassade, ginge er jeder anderen Fläche verloren — und
// MSE2 stellt die Frage vor **jeder** Ausgabe, nicht vor einer
// (`SetWindow::selectionChoices`, Fassung f9d9356 vom 2021-07-06).

// **Ohne Auswahl kommt das ganze Set.**
func TestOhneAuswahlKommtDasGanzeSet(t *testing.T) {
	t.Parallel()

	ziel := filepath.Join(t.TempDir(), "bogen.pdf")

	lauf, err := bogenleger(t, 3).Lege(
		t.Context(), 7, ziel, service.StandardBogen(), domain.Kartenauswahl{})
	if err != nil {
		t.Fatalf("Lege: %v", err)
	}

	if lauf.Karten != 3 {
		t.Errorf("es kamen %d Karten auf den Bogen", lauf.Karten)
	}
}

// **Mit Auswahl kommen nur die gewählten.**
func TestMitAuswahlKommenNurDieGewaehlten(t *testing.T) {
	t.Parallel()

	ziel := filepath.Join(t.TempDir(), "bogen.pdf")

	lauf, err := bogenleger(t, 3).Lege(
		t.Context(), 7, ziel, service.StandardBogen(),
		domain.Kartenauswahl{Karten: []int64{2}})
	if err != nil {
		t.Fatalf("Lege: %v", err)
	}

	if lauf.Karten != 1 {
		t.Errorf("es kamen %d Karten auf den Bogen", lauf.Karten)
	}
}

// **Eine Auswahl, die keine Karte trifft, wird gemeldet** — und schreibt
// keine Datei.
//
// Ein Bogen über das ganze Set wäre die teuerste Art, einen Tippfehler zu
// beantworten; ein leerer Bogen die verwirrendste.
func TestEineAuswahlOhneTrefferSchreibtKeinenBogen(t *testing.T) {
	t.Parallel()

	ziel := filepath.Join(t.TempDir(), "bogen.pdf")

	_, err := bogenleger(t, 3).Lege(
		t.Context(), 7, ziel, service.StandardBogen(),
		domain.Kartenauswahl{Karten: []int64{99}})
	if err == nil {
		t.Fatal("die leere Auswahl kam als Erfolg zurück")
	}

	if _, statfehler := filepath.Glob(ziel); statfehler != nil {
		t.Fatalf("glob: %v", statfehler)
	}
}

// **Und die Bildausgabe nimmt dieselbe Auswahl** — Befund 2 des Zweitblicks
// über Phase 3.6.
//
// MSE2 stellt die Wahl vor `onFileExportImages` genauso
// (`src/gui/set/window.cpp`, Zeilen 668–673, Fassung f9d9356 vom 2021-07-06);
// hier fehlte sie, während Druckbogen und Exportvorlage sie schon hatten.
func TestDieBildausgabeNimmtEineAuswahl(t *testing.T) {
	t.Parallel()

	ausfuhr := service.NeueBildausfuhr(
		kartenDouble{leserDouble: karten("Erste", "Zweite", "Dritte")}, nil, nil)

	gewaehlt, err := ausfuhr.Namen(t.Context(), 7, "{name}", service.FormatPNG,
		domain.BildkonfliktNummerierenNeu, domain.Kartenauswahl{Karten: []int64{2}})
	if err != nil {
		t.Fatalf("Namen: %v", err)
	}

	if len(gewaehlt) != 1 || gewaehlt[0].KarteID != 2 {
		t.Errorf("es kamen %v", gewaehlt)
	}

	// **Und eine Auswahl ohne Treffer wird gemeldet** — vor der Namensvergabe,
	// damit keine Lücken in der Nummerierung entstehen.
	if _, fehler := ausfuhr.Namen(t.Context(), 7, "{name}", service.FormatPNG,
		domain.BildkonfliktNummerierenNeu,
		domain.Kartenauswahl{Karten: []int64{99}}); fehler == nil {
		t.Fatal("die leere Auswahl kam als Erfolg zurück")
	}
}
