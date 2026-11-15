package msescript_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **Derselbe Quelltext wird einmal zerteilt, nicht bei jeder Karte**
// (#672).
//
// # Warum das zählt
//
// `magic.mse-game` bringt ein Initskript von über 10.000 Zeilen mit, und es
// wird für **jede** Karte neu zerteilt. Gemessen über 120 Vorschauen:
// `Zerteile` allein war **8,8 %** der gesamten Rechenzeit.
//
// **Der Baum hängt nicht an der Karte.** Er ist dasselbe Ergebnis desselben
// Textes, und der Auswerter liest ihn nur — nachgesehen: keine Zuweisung an
// ein Knotenfeld außerhalb des Zerteilers.
func TestZerteileMerktSichDenBaum(t *testing.T) {
	t.Parallel()

	quelle := `a := 1
b := a + 2
c := { b() }`

	erster, err := msescript.Zerteile(quelle)
	if err != nil {
		t.Fatalf("unerwarteter Fehler: %v", err)
	}

	zweiter, err := msescript.Zerteile(quelle)
	if err != nil {
		t.Fatalf("unerwarteter Fehler: %v", err)
	}

	// **Derselbe Baum, nicht ein gleicher.** Das ist die Zusage: der zweite
	// Aufruf hat nicht gerechnet.
	if erster != zweiter {
		t.Error("der zweite Aufruf hat neu zerteilt")
	}
}

// **Ein anderer Quelltext gibt einen anderen Baum.**
func TestZerteileUnterscheidetQuellen(t *testing.T) {
	t.Parallel()

	eins, err := msescript.Zerteile(`x := 1`)
	if err != nil {
		t.Fatalf("unerwarteter Fehler: %v", err)
	}

	zwei, err := msescript.Zerteile(`x := 2`)
	if err != nil {
		t.Fatalf("unerwarteter Fehler: %v", err)
	}

	if eins == zwei {
		t.Fatal("zwei Quelltexte teilen sich einen Baum")
	}

	// **Und er rechnet auch das Richtige** — ein Speicher, der den falschen
	// Baum herausgibt, gäbe zwei verschiedene Zeiger auf dasselbe Ergebnis.
	wert, err := msescript.AuswertenQuelle(
		t.Context(), `x := 2`, msescript.Bibliothek().Darueber())
	if err != nil {
		t.Fatalf("unerwarteter Fehler: %v", err)
	}

	if wert.AlsText() != "2" {
		t.Errorf("Ergebnis = %q, erwartet \"2\"", wert.AlsText())
	}
}

// **Ein Fehler wird nicht gemerkt.**
//
// Dieselbe Regel wie bei [msepaket.Bildquelle]: ein gemerkter Fehlschlag gälte
// bis zum Neustart. Hier kommt dazu, dass ein unlesbarer Quelltext gar keinen
// Baum hat, den man aufheben könnte.
func TestZerteileMerktKeinenFehler(t *testing.T) {
	t.Parallel()

	for range 2 {
		_, err := msescript.Zerteile(`a := (`)
		if err == nil {
			t.Fatal("erwartet wurde ein Fehler, es kam keiner")
		}
	}
}

// **Die Grenze räumt ganz**, und danach wird wieder gerechnet.
//
// Dieselbe Wahl wie bei [msepaket.Bildquelle]: eine Verdrängung nach Alter
// verlangt Buchführung über jeden Zugriff, und der Nutzen wäre einer für einen
// Fall, den es nicht gibt — ein Set benutzt ein gutes Dutzend Skripte immer
// wieder.
func TestZerteileRaeumtAnDerGrenze(t *testing.T) {
	t.Parallel()

	quelle := `grenzprobe := 1`

	erster, err := msescript.Zerteile(quelle)
	if err != nil {
		t.Fatalf("unerwarteter Fehler: %v", err)
	}

	for i := range msescript.GrenzeBaumspeicher + 1 {
		if _, fehler := msescript.Zerteile(
			`fueller := ` + strings.Repeat("1", i+1),
		); fehler != nil {
			t.Fatalf("unerwarteter Fehler: %v", fehler)
		}
	}

	zweiter, err := msescript.Zerteile(quelle)
	if err != nil {
		t.Fatalf("unerwarteter Fehler: %v", err)
	}

	if erster == zweiter {
		t.Error("die Grenze hat nicht geräumt")
	}
}

// **Die Grenze ist aus einer Messung abgeleitet, nicht geraten.**
//
// # Warum das ein Lauf ist und keine Zeile im Kommentar
//
// Weil hier bis zum 2026-08-29 eine geratene Zahl stand — 256, mit der
// Begründung, ein Stil zeichne „ein gutes Dutzend" Skripte. Gemessen sind es
// **561** verschiedene Quelltexte für **ein** Set, und der Speicher räumte
// deshalb 316-mal in einem einzigen `mse packarten`.
//
// **Was dieser Lauf hält**: dass die Grenze Raum für mehrere Sets lässt. Was
// er **nicht** hält: dass 561 heute noch stimmt — eine Messung veraltet, und
// wer ein größeres Spielpaket öffnet, misst neu. Die Zahl trägt deshalb ihr
// Datum im Kommentar.
func TestDieGrenzeTraegtMehrereSets(t *testing.T) {
	t.Parallel()

	const sets = 4

	if msescript.GrenzeBaumspeicher < sets*msescript.Arbeitsmenge {
		t.Errorf("die Grenze ist %d — die gemessene Arbeitsmenge eines Sets ist "+
			"%d, und %d Sets zugleich passen damit nicht hinein",
			msescript.GrenzeBaumspeicher, msescript.Arbeitsmenge, sets)
	}
}
