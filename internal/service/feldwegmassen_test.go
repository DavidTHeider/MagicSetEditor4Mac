package service_test

// **Eine Massenänderung speichert, was gerechnet ist** — #2516 (Rest aus
// #2396). Der Feldweg rechnet den Wert auf **jeder** Karte, zieht nach, wer
// ihn liest, und gibt der Ablage je Karte einen Schritt mit dem Stand, aus dem
// gerechnet wurde; sie schreibt alle in einem.

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// FelderGerechnetSetzen hält die Schritte fest, die der Feldweg schickt.
func (w *wegablage) FelderGerechnetSetzen(
	_ context.Context, schritte []domain.Kartenschritt, erwartet []domain.Feldvorwert, feld string,
) (string, error) {
	if w.fehler != nil {
		return "", w.fehler
	}

	w.reihen++
	w.massenfeld, w.massenerwartung = feld, erwartet

	for _, schritt := range schritte {
		w.erwartet = append(w.erwartet, schritt.Erwartet)

		for _, wert := range schritt.Werte {
			w.geschrieben = append(w.geschrieben, wert.Field+"="+wert.Text)
		}
	}

	return "vorgang9", nil
}

func TestFeldwegRechnetDieMassenaenderungAufJederKarte(t *testing.T) {
	t.Parallel()

	ablage := &wegablage{}
	stile := &wegstile{vorlage: service.Vorlage{Kartenfelder: nachzugfelder()}, gefunden: true}
	weg := service.NeuerFeldweg(wegkarten{karte: nachzugkarte()}, wegfelder{felder: nachzugfelder()}, ablage).
		MitNachzug(stile, nil, nil)
	erwartet := []domain.Feldvorwert{{KarteID: 7, Wert: "[alt]", GabEs: true}, {KarteID: 8, Wert: "[alt]", GabEs: true}}

	vorgang, fehler := weg.MassenSetzen(t.Context(), []int64{7, 8}, "eins", "Funke", erwartet)
	if fehler != nil || vorgang != "vorgang9" {
		t.Fatalf("setzen: %q, %v", vorgang, fehler)
	}

	if want := []string{"eins=[Funke]", "leser=[Funke]!", "eins=[Funke]", "leser=[Funke]!"}; !slices.Equal(ablage.geschrieben, want) {
		t.Errorf("geschrieben %q, erwartet %q", ablage.geschrieben, want)
	}

	// **Ein Aufruf, je Karte der Stand ihrer Rechnung**, und die Erwartung
	// aus der Vorschau reicht durch.
	if want := []string{rechenstand(nil), rechenstand(nil)}; ablage.reihen != 1 || !slices.Equal(ablage.erwartet, want) {
		t.Errorf("%d Aufrufe, Stände %q", ablage.reihen, ablage.erwartet)
	}

	if ablage.massenfeld != "eins" || len(ablage.massenerwartung) != 2 {
		t.Errorf("Feld %q, Erwartung %v", ablage.massenfeld, ablage.massenerwartung)
	}
}

// **Ohne lesbare Vorlage bleibt der Rohwert**, wie beim einzelnen Feld (#720)
// — und ohne Stand, denn gerechnet wurde nichts.
func TestEineMassenaenderungOhneVorlageSchreibtRoh(t *testing.T) {
	t.Parallel()

	ablage := &wegablage{}
	weg := service.NeuerFeldweg(wegkarten{karte: nachzugkarte()},
		wegfelder{fehler: errors.New("die Vorlage ist unlesbar")}, ablage).
		MitNachzug(&wegstile{vorlage: service.Vorlage{Kartenfelder: nachzugfelder()}, gefunden: true}, nil, nil)

	if _, fehler := weg.MassenSetzen(t.Context(), []int64{7}, "eins", "Funke", nil); fehler != nil {
		t.Fatalf("setzen: %v", fehler)
	}

	if want := []string{"eins=Funke"}; !slices.Equal(ablage.geschrieben, want) || !slices.Equal(ablage.erwartet, []string{""}) {
		t.Errorf("geschrieben %q, Stände %q", ablage.geschrieben, ablage.erwartet)
	}
}
