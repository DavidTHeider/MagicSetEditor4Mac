// **Die Fehlerwege beim Einfügen** — Werkstück 1 von Phase 3.7
// (#1168).
//
// **Eigene Datei**, weil `kartenablage_test.go` sonst über 400 Zeilen ginge
// (§2). Der Schnitt folgt der Frage: dort steht, **was** die Ablage tut, hier,
// **was zurückbleibt, wenn es schiefgeht.**
package service_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Eine Karte, die entstanden ist, verschwindet nicht.**
//
// Scheitert das Füllen, kommt die Kennung trotzdem zurück — sonst bliebe eine
// leere Karte im Bestand, von der niemand weiß.
func TestEineEntstandeneKarteKommtAuchBeiEinemFehlerZurueck(t *testing.T) {
	t.Parallel()

	doppel := zweiSets()
	ablagen := kartenablage(doppel)

	inhalt, err := ablagen.Kopieren(t.Context(), 1, []int64{11})
	if err != nil {
		t.Fatalf("kopieren: %v", err)
	}

	doppel.schreibfehler = domain.ErrBelegt

	ergebnis, fehler := ablagen.Einfuegen(t.Context(), 2, inhalt)
	if fehler == nil {
		t.Fatal("der Fehler kam als Erfolg zurück")
	}

	if ergebnis == nil {
		t.Fatal("es kam kein Ergebnis zurück")
	}
}

// **Ein unlesbarer Bestand meldet sich beim Kopieren.**
func TestEinUnlesbarerBestandMeldetSichBeimKopieren(t *testing.T) {
	t.Parallel()

	for name, doppel := range map[string]*ablagedoppel{
		"beim Lesen der Sets":   {fehler: domain.ErrBelegt},
		"beim Lesen der Karten": {koepfe: zweiSets().koepfe, kartenfehler: domain.ErrBelegt},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, fehler := kartenablage(doppel).Kopieren(
				t.Context(), 1, []int64{11}); fehler == nil {
				t.Fatal("der Fehler kam als Erfolg zurück")
			}
		})
	}
}

// **Und ein Set, das es nicht gibt, ebenso** — auf beiden Wegen.
func TestEinUnbekanntesSetWirdAufBeidenWegenGemeldet(t *testing.T) {
	t.Parallel()

	ablagen := kartenablage(zweiSets())

	if _, fehler := ablagen.Kopieren(t.Context(), 99, []int64{11}); !errors.Is(
		fehler, domain.ErrNotFound) {
		t.Errorf("kopieren: %v", fehler)
	}

	if _, fehler := ablagen.Einfuegen(
		t.Context(), 99, "game_name: magic\ncard:\n\tname: Erste\n"); !errors.Is(
		fehler, domain.ErrNotFound) {
		t.Errorf("einfügen: %v", fehler)
	}
}

// **Eine Karte ohne Felder ist eine leere Karte, kein Fehler.**
//
// MSE legt eine Karte ohne Werte an, wenn jemand eine anlegt; eine kopierte
// ohne Werte ist dasselbe.
func TestEineKarteOhneFelderWirdLeerAngelegt(t *testing.T) {
	t.Parallel()

	doppel := zweiSets()

	ergebnis, err := kartenablage(doppel).Einfuegen(
		t.Context(), 2, "game_name: magic\ncard:\n")
	if err != nil {
		t.Fatalf("einfügen: %v", err)
	}

	if len(ergebnis.Karten) != 1 {
		t.Fatalf("es entstanden %d Karten", len(ergebnis.Karten))
	}

	if len(doppel.werte) != 0 {
		t.Errorf("es wurden Werte gesetzt: %v", doppel.werte)
	}
}

// **Ein Anlegefehler bricht ab, und was schon entstand, kommt mit.**
func TestEinAnlegefehlerBrichtAb(t *testing.T) {
	t.Parallel()

	doppel := zweiSets()
	doppel.anlegefehler = domain.ErrBelegt

	ergebnis, fehler := kartenablage(doppel).Einfuegen(
		t.Context(), 2, "game_name: magic\ncard:\n\tname: Erste\n")
	if fehler == nil {
		t.Fatal("der Fehler kam als Erfolg zurück")
	}

	if ergebnis == nil || len(ergebnis.Karten) != 0 {
		t.Errorf("Ergebnis: %v", ergebnis)
	}
}

// **Ein unlesbarer Inhalt wird abgewiesen, nicht geraten.**
//
// `Zerteile` kennt genau einen Fehlerfall: eine Zeile, die tiefer einrückt,
// als ihr Vorgänger es erlaubt. Aus der Zwischenablage kommt so etwas leicht —
// jemand kopiert die Mitte eines Blocks.
func TestEinUnlesbarerInhaltWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	doppel := zweiSets()

	// **Ein leerer Schlüssel** — das ist der eine Fehlerfall, den `Zerteile`
	// kennt (`zerteiler.go`, Zeile 351). Aus der Zwischenablage kommt so etwas
	// leicht: jemand kopiert eine Zeile, die mit einem Doppelpunkt beginnt.
	_, fehler := kartenablage(doppel).Einfuegen(
		t.Context(), 2, "game_name: magic\n: ohne Schlüssel\n")
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("Fehler war %v — will ErrValidation", fehler)
	}

	if len(doppel.angelegt) != 0 {
		t.Errorf("es wurden %d Karten angelegt", len(doppel.angelegt))
	}
}

// **Eine Leerzeile im Kartenblock wird kein Feld.**
//
// # Warum das keine Wache im Code braucht
//
// `Zerteile` nimmt Leerzeilen und Kommentare **nicht als Kinder** auf —
// nachgemessen am 2026-08-27. Eine Wache gegen den leeren Schlüssel wäre ein
// Zweig, den kein Lauf betritt (§3). **Dieser Lauf hält die Aussage**: kommt
// eine Leerzeile mit, entsteht kein Feld daraus.
func TestEineLeerzeileImKartenblockWirdKeinFeld(t *testing.T) {
	t.Parallel()

	doppel := zweiSets()

	// **Eine Leerzeile im Block** — sie steht eingerückt und wird damit ein
	// Kind der Karte, aber ohne Schlüssel.
	ergebnis, err := kartenablage(doppel).Einfuegen(
		t.Context(), 2, "game_name: magic\ncard:\n\tname: Erste\n\t\n")
	if err != nil {
		t.Fatalf("einfügen: %v", err)
	}

	if len(doppel.werte[ergebnis.Karten[0]]) != 1 {
		t.Errorf("es kamen %v", doppel.werte[ergebnis.Karten[0]])
	}
}
