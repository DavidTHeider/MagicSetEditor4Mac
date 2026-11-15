package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Eine Übernahme schreibt in einem Zug — oder gar nicht** —
// #1026.
//
// # Der Zustand, den es zu verhindern gilt
//
// Bis zum 2026-08-24 schrieb `Uebernehmen` Feld für Feld, jedes in seiner
// eigenen Transaktion. Scheiterte das dritte von fünf, standen zwei neu da,
// drei alt — und der Aufrufer bekam `nil, fehler`. **Welche Hälfte durch war,
// stand nirgends.**
//
// §7, Regel 3 sagt „geschrieben ist geschrieben", und das bleibt richtig. Der
// Fehlervertrag sagte es nur nicht: er gab die Liste der schon geschriebenen
// Felder nicht heraus, und damit fehlte genau die Angabe, die eine Rücknahme
// bräuchte.
//
// # Warum der Weg „alles in einer Transaktion" hier keine Zusage verschiebt
//
// `storage.FelderSetzenMitVorwert` sagt seit
// #714
// **beides** zu: alles oder nichts, **und** seit #2396 ein Verlaufsschritt
// für alle — die Übernahme wird weder halb geschrieben noch halb
// zurückgenommen.
func TestEineUebernahmeSchreibtInEinemZug(t *testing.T) {
	t.Parallel()

	schreiber := &halbschreiber{stolpert: "rule text"}
	abgleich := service.NeuerScryfallabgleich(fremdkarte(), nil, neuesVerweisdoppel(), schreiber)

	vorschlag := mussVorschlag(t, abgleich, map[string]string{}, "Prüfstein")

	_, fehler := abgleich.Uebernehmen(t.Context(), 1, 2, vorschlag,
		[]string{"name", "rule text", "super type"})
	if fehler == nil {
		t.Fatal("der Schreibfehler kam nicht durch")
	}

	if !errors.Is(fehler, domain.ErrConflict) {
		t.Errorf("Fehler %v — der des Schreibers war es nicht", fehler)
	}

	// **Nichts steht halb da.** Der Schreiber hat die Änderung als Ganzes
	// abgelehnt; geschrieben wurde deshalb kein einziges Feld.
	if len(schreiber.geschrieben) != 0 {
		t.Errorf("halb geschrieben: %v", schreiber.geschrieben)
	}

	// **Und es war ein Zug**, nicht drei. Ohne diese Zusicherung bliebe die
	// Zeile darüber auch dann grün, wenn der erste Aufruf zufällig der
	// scheiternde wäre.
	if schreiber.aufrufe != 1 {
		t.Errorf("%d Schreibaufrufe, erwartet war einer", schreiber.aufrufe)
	}
}

// **Und was durchgeht, geht vollständig durch.**
func TestEineGeglueckteUebernahmeSchreibtAlleFelder(t *testing.T) {
	t.Parallel()

	schreiber := &halbschreiber{}
	abgleich := service.NeuerScryfallabgleich(fremdkarte(), nil, neuesVerweisdoppel(), schreiber)

	vorschlag := mussVorschlag(t, abgleich, map[string]string{}, "Prüfstein")

	uebernommen, fehler := abgleich.Uebernehmen(t.Context(), 1, 2, vorschlag,
		[]string{"name", "rule text"})
	if fehler != nil {
		t.Fatalf("Übernehmen: %v", fehler)
	}

	if len(uebernommen) != 2 || len(schreiber.geschrieben) != 2 {
		t.Errorf("übernommen %v, geschrieben %v", uebernommen, schreiber.geschrieben)
	}

	if schreiber.aufrufe != 1 {
		t.Errorf("%d Schreibaufrufe, erwartet war einer", schreiber.aufrufe)
	}
}

// halbschreiber stellt einen Schreiber nach, der bei **einem** genannten Feld
// scheitert.
//
// **Beide Wege**, damit dieser Test den Unterschied zeigt und nicht die
// Übersetzung: einzeln geschrieben bleibt das erste Feld stehen, in einem Zug
// bleibt nichts stehen.
type halbschreiber struct {
	stolpert    string
	geschrieben []string
	aufrufe     int
}

func (h *halbschreiber) FeldSetzenVonAussen(
	_ context.Context, _, _ int64, feld, _, _ string,
) (vorher string, hatte bool, fehler error) {
	h.aufrufe++

	if feld == h.stolpert {
		return "", false, domain.ErrConflict
	}

	h.geschrieben = append(h.geschrieben, feld)

	return "", false, nil
}

// FelderVonAussen ist alles oder nichts — wie die Transaktion darunter.
func (h *halbschreiber) FelderVonAussen(
	_ context.Context, _, _ int64, werte []domain.Value, _ string,
) (vorher string, hatte bool, fehler error) {
	h.aufrufe++

	for _, wert := range werte {
		if wert.Field == h.stolpert {
			return "", false, domain.ErrConflict
		}
	}

	for _, wert := range werte {
		h.geschrieben = append(h.geschrieben, wert.Field)
	}

	return "", false, nil
}
