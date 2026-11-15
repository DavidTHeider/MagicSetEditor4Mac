package service_test

// **Auch eine Gruppe speichert, was gerechnet ist** — #2396
// (Issue (#2396)).
//
// Die Übernahme eines KI-Vorschlags über mehrere Karten (A13.1b, #1452) schrieb
// bis dahin an der Auflösung vorbei: ein Verbundfeld landete als Zeile im
// Bestand, die die Vorlage nie liest (#696), und kein Wertskript rechnete.

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// GruppeSetzenMitHerkunft schreibt jeden Schritt — **in einem Aufruf**, denn
// die Ablage sagt eine Transaktion für alle zu.
func (w *wegablage) GruppeSetzenMitHerkunft(_ context.Context, schritte []domain.Kartenschritt) error {
	if w.fehler != nil {
		return w.fehler
	}

	w.reihen++

	for _, schritt := range schritte {
		w.herkuenfte = append(w.herkuenfte, schritt.Herkunft)
		w.feldherkuenfte = append(w.feldherkuenfte, schritt.Herkuenfte...)
		w.erwartet = append(w.erwartet, schritt.Erwartet)

		for _, wert := range schritt.Werte {
			w.geschrieben = append(w.geschrieben, wert.Field+"="+wert.Text)
		}
	}

	return nil
}

func gruppenschritte(werte ...string) []domain.Kartenschritt {
	schritte := make([]domain.Kartenschritt, 0, len(werte))

	for i, wert := range werte {
		schritte = append(schritte, domain.Kartenschritt{
			KarteID: int64(7 + i), SetID: 1, Herkunft: "ki",
			Werte:      []domain.Value{{Field: "eins", Text: wert}},
			Herkuenfte: []domain.KIFeldherkunft{{Feld: "eins"}},
		})
	}

	return schritte
}

func TestFeldwegRechnetJedeKarteEinerGruppe(t *testing.T) {
	t.Parallel()

	ablage := &wegablage{}
	weg := service.NeuerFeldweg(wegkarten{karte: nachzugkarte()}, wegfelder{felder: nachzugfelder()}, ablage).
		MitNachzug(&wegstile{vorlage: service.Vorlage{Kartenfelder: nachzugfelder()}, gefunden: true}, nil, nil)

	if fehler := weg.GruppeSetzenMitHerkunft(t.Context(), gruppenschritte("Funke", "Glut")); fehler != nil {
		t.Fatalf("setzen: %v", fehler)
	}

	want := []string{"eins=[Funke]", "leser=[Funke]!", "eins=[Glut]", "leser=[Glut]!"}
	if !slices.Equal(ablage.geschrieben, want) {
		t.Errorf("geschrieben %v, erwartet %v", ablage.geschrieben, want)
	}

	// **Eine Transaktion für alle** — die Zusage von A13.1b bleibt.
	if ablage.reihen != 1 {
		t.Errorf("%d Aufrufe der Ablage, erwartet einer", ablage.reihen)
	}

	// **Die Herkunft geht unverändert mit.**
	if want := []string{"ki", "ki"}; !slices.Equal(ablage.herkuenfte, want) {
		t.Errorf("Herkunft %v, erwartet %v", ablage.herkuenfte, want)
	}

	if len(ablage.feldherkuenfte) != 2 {
		t.Errorf("Feldherkunft %v, erwartet zwei Zeilen", ablage.feldherkuenfte)
	}
}

// **Ein Feld, das nicht gespeichert wird, hält die ganze Gruppe auf** — vor
// dem ersten Schreiben, wie beim einzelnen Weg ([service.Feldweg.FelderVonAussen]).
func TestFeldwegWeistEineGruppeMitBerechnetemFeldAb(t *testing.T) {
	t.Parallel()

	ablage := &wegablage{}
	weg := service.NeuerFeldweg(wegkarten{karte: nachzugkarte()}, wegfelder{felder: nachzugfelder()}, ablage)

	schritte := gruppenschritte("Funke")
	schritte = append(schritte, domain.Kartenschritt{
		KarteID: 9, SetID: 1, Werte: []domain.Value{{Field: "anzeige", Text: "?"}},
	})

	fehler := weg.GruppeSetzenMitHerkunft(t.Context(), schritte)
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("Fehler %v, erwartet ErrValidation", fehler)
	}

	if ablage.reihen != 0 || len(ablage.geschrieben) != 0 {
		t.Errorf("geschrieben %v in %d Aufrufen, erwartet nichts", ablage.geschrieben, ablage.reihen)
	}
}
