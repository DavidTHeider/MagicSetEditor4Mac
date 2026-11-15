package service_test

// **Was eine Zeile sagt, die nicht aufgeht** — G01
// (#2248).
//
// Gefunden am nativen Fenster (`g01einsteiger`): eine Zeile, die es bei
// Scryfall nicht gibt, stand als „nicht gefunden: Scryfall kennt eine Karte zu
// dieser Suche nicht — und es gibt keinen Stand im Speicher" da — der rohe
// Fehler samt seiner Klasse, und in jeder Sprache deutsch. Eine Zeile ohne
// Druck ist kein Fehler, sondern ein Ergebnis; ein Ausfall ist einer und
// trägt ihn, damit die Oberfläche ihn übersetzen kann.

import (
	"errors"
	"fmt"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestEinNameOhneDruckBeiScryfallIstKeinFehler(t *testing.T) {
	t.Parallel()

	quelle := &listenquelle{fehler: map[string]error{
		`!"Gibtesnichtxyz"`: fmt.Errorf("%w: Scryfall kennt eine Karte zu dieser Suche nicht", domain.ErrNotFound),
	}}

	deutung, fehler := service.NeueListeneinfuhr(quelle, &sammlungsdoppel{}).Deuten(t.Context(), "1 Gibtesnichtxyz")
	if fehler != nil {
		t.Fatalf("Deuten: %v", fehler)
	}

	zeile := deutung.Zeilen[0]
	if zeile.Befund != "kein Druck heißt so" || zeile.Fehler != nil || zeile.Drucke != 0 {
		t.Errorf("Zeile: Befund %q, Fehler %v, Drucke %d", zeile.Befund, zeile.Fehler, zeile.Drucke)
	}
}

func TestEinAusfallTraegtSeinenFehler(t *testing.T) {
	t.Parallel()

	ausfall := errors.New("kein Netz")
	quelle := &listenquelle{fehler: map[string]error{`!"Counterspell"`: ausfall}}

	deutung, fehler := service.NeueListeneinfuhr(quelle, &sammlungsdoppel{}).Deuten(t.Context(), "1 Counterspell")
	if fehler != nil {
		t.Fatalf("Deuten: %v", fehler)
	}

	if !errors.Is(deutung.Zeilen[0].Fehler, ausfall) {
		t.Errorf("der Ausfall ging verloren: %v", deutung.Zeilen[0].Fehler)
	}
}

func TestMehrAlsEineSeiteSagtEsAuchOhneText(t *testing.T) {
	t.Parallel()

	quelle := &listenquelle{antworten: map[string]scryfall.Treffer{
		`!"Island"`:       {Weitere: true, Karten: []*scryfall.Karte{druck("a", "Island", "lea", "1"), druck("b", "Island", "leb", "2")}},
		`!"Counterspell"`: {Karten: []*scryfall.Karte{druck("c", "Counterspell", "lea", "54"), druck("d", "Counterspell", "7ed", "67")}},
	}}

	deutung, fehler := service.NeueListeneinfuhr(quelle, &sammlungsdoppel{}).Deuten(t.Context(), "1 Island\n1 Counterspell")
	if fehler != nil {
		t.Fatalf("Deuten: %v", fehler)
	}

	if !deutung.Zeilen[0].Weitere || deutung.Zeilen[1].Weitere {
		t.Errorf("Weitere: %v, %v", deutung.Zeilen[0].Weitere, deutung.Zeilen[1].Weitere)
	}
}
