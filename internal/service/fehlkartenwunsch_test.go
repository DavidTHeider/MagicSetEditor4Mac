package service_test

// **Was einem Deck fehlt, auf die Wunschliste** — DOM-03.AK-02
// (#1698).
//
// # Die Regel
//
// **Ein Druck, den es noch nicht als Wunsch gibt, kommt mit der fehlenden
// Menge darauf. Ein Druck, der schon gewünscht ist, bleibt, wie er ist** —
// und die Antwort nennt ihn. Eine Menge, die jemand selbst gesetzt hat, wird
// nicht still verschoben (AK-03: kontrollierte Mengenänderungen); ein
// zweiter Ruf legt nichts doppelt an.

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// wunschanlagedoppel nimmt die neuen Wünsche entgegen — alle oder keinen.
type wunschanlagedoppel struct {
	angelegt []*domain.Exemplar
	rufe     int
	fehler   error
}

func (w *wunschanlagedoppel) AlleAnlegen(_ context.Context, exemplare []*domain.Exemplar) (
	[]int64, error,
) {
	w.rufe++

	if w.fehler != nil {
		return nil, w.fehler
	}

	kennungen := make([]int64, 0, len(exemplare))

	for stelle, exemplar := range exemplare {
		w.angelegt = append(w.angelegt, exemplar)
		kennungen = append(kennungen, int64(100+stelle))
	}

	return kennungen, nil
}

func wunschwerk(t *testing.T) (*pruefwerk, *wunschanlagedoppel, *service.Fehlkartenwunsch) {
	t.Helper()

	werk := neuesPruefwerk(t)
	anlage := &wunschanlagedoppel{}

	return werk, anlage, service.NeuerFehlkartenwunsch(werk.abgleich, werk.bau, anlage)
}

func TestFehlkartenKommenAufDieWunschliste(t *testing.T) {
	werk, anlage, wunsch := wunschwerk(t)

	// **Der Name kommt aus dem Deck**, und das Doppel nennt einen Druck ohne
	// eigenen Namen „Druck …" — wie der Deckbau, wenn Scryfall ihn kennt.
	werk.drucke.namen[druckB] = "Glut"
	werk.besitze(druckA, 1)
	werk.legeDruck(t, druckA, 4)
	werk.legeDruck(t, druckB, 2)

	ergebnis, fehler := wunsch.Wuenschen(t.Context(), werk.deck.ID)
	if fehler != nil {
		t.Fatalf("wünschen: %v", fehler)
	}

	if len(ergebnis.Neu) != 2 || anlage.rufe != 1 {
		t.Fatalf("neu %+v in %d Rufen, erwartet zwei in einem", ergebnis.Neu, anlage.rufe)
	}

	for stelle, erwartet := range []struct {
		printing string
		menge    int
		name     string
	}{{druckA, 3, "Druck " + druckA}, {druckB, 2, "Glut"}} {
		neu := ergebnis.Neu[stelle]
		if neu.Printing != erwartet.printing || neu.Menge != erwartet.menge ||
			neu.Name != erwartet.name || !neu.Wunsch || neu.ID != int64(100+stelle) {
			t.Errorf("Wunsch %d ist %+v, erwartet %+v", stelle, neu, erwartet)
		}

		// **Die Vorgaben eines neuen Exemplars** — dieselben wie im Formular.
		if neu.Zustand != domain.ZustandNearMint || neu.Sprache != domain.Vorgabesprache {
			t.Errorf("Wunsch %d steht in %s/%s", stelle, neu.Zustand, neu.Sprache)
		}
	}
}

// **Ein Wunsch, den es schon gibt, bleibt, wie er ist** — auch wenn seine
// Menge kleiner ist als das, was fehlt. Die Antwort nennt ihn samt beider
// Zahlen; das Ändern ist eine Entscheidung dessen, der die Liste führt.
func TestEinVorhandenerWunschBleibtUnberuehrt(t *testing.T) {
	werk, anlage, wunsch := wunschwerk(t)

	werk.wuensche(druckA, 1)
	werk.legeDruck(t, druckA, 4)

	ergebnis, fehler := wunsch.Wuenschen(t.Context(), werk.deck.ID)
	if fehler != nil {
		t.Fatalf("wünschen: %v", fehler)
	}

	if len(ergebnis.Neu) != 0 || anlage.rufe != 0 {
		t.Errorf("neu %+v, %d Rufe — ein vorhandener Wunsch wurde verdoppelt", ergebnis.Neu, anlage.rufe)
	}

	if len(ergebnis.Vorhanden) != 1 || ergebnis.Vorhanden[0].Printing != druckA ||
		ergebnis.Vorhanden[0].Fehlend != 4 || ergebnis.Vorhanden[0].Gewuenscht != 1 {
		t.Errorf("vorhanden %+v", ergebnis.Vorhanden)
	}
}

// **Was da ist, wird nicht gewünscht**, und ohne Neues wird nicht
// geschrieben.
func TestEinVollstaendigesDeckWuenschtNichts(t *testing.T) {
	werk, anlage, wunsch := wunschwerk(t)

	werk.besitze(druckA, 4)
	werk.legeDruck(t, druckA, 4)

	ergebnis, fehler := wunsch.Wuenschen(t.Context(), werk.deck.ID)
	if fehler != nil || len(ergebnis.Neu) != 0 || len(ergebnis.Vorhanden) != 0 || anlage.rufe != 0 {
		t.Errorf("Ergebnis %+v, %d Rufe, Fehler %v", ergebnis, anlage.rufe, fehler)
	}
}

// **Ein eigener Entwurf ohne Verweis hat keinen Druck** — er kann nicht auf
// die Wunschliste und wird gezählt statt verschwiegen.
func TestEinEigenerEntwurfWirdGezaehltStattGewuenscht(t *testing.T) {
	werk, anlage, wunsch := wunschwerk(t)

	werk.legeKarte(t, 2)

	ergebnis, fehler := wunsch.Wuenschen(t.Context(), werk.deck.ID)
	if fehler != nil || ergebnis.Unbekannt != 2 || len(ergebnis.Neu) != 0 || anlage.rufe != 0 {
		t.Errorf("Ergebnis %+v, %d Rufe, Fehler %v", ergebnis, anlage.rufe, fehler)
	}
}

// **Ein vorläufiger Druck wird über seine Zeile gewünscht** (A2, #1441) —
// er hat keinen Alias, und ein Wunsch ohne Bezug gäbe es nicht.
func TestEinVorlaeufigerDruckWirdUeberSeineZeileGewuenscht(t *testing.T) {
	werk, _, wunsch := wunschwerk(t)

	if _, fehler := werk.bau.PlatzLegen(t.Context(), werk.deck.ID, domain.Deckplatz{
		DruckID: 9, Menge: 2, Bereich: domain.BereichHauptdeck,
	}); fehler != nil {
		t.Fatalf("legen: %v", fehler)
	}

	ergebnis, fehler := wunsch.Wuenschen(t.Context(), werk.deck.ID)
	if fehler != nil || len(ergebnis.Neu) != 1 ||
		ergebnis.Neu[0].DruckID != 9 || ergebnis.Neu[0].Printing != "" || ergebnis.Neu[0].Menge != 2 {
		t.Errorf("Ergebnis %+v, Fehler %v", ergebnis, fehler)
	}
}

// **Jeder Fehler hält an** — ein halber Abgleich gäbe eine halbe Liste.
func TestDerWunschwegMeldetSeineFehler(t *testing.T) {
	t.Run("der Abgleich", func(t *testing.T) {
		werk, anlage, wunsch := wunschwerk(t)
		werk.legeDruck(t, druckA, 4)
		werk.sammlung.fehler = errKaputt

		if _, fehler := wunsch.Wuenschen(t.Context(), werk.deck.ID); !errors.Is(fehler, errKaputt) {
			t.Errorf("Fehler = %v", fehler)
		}

		if anlage.rufe != 0 {
			t.Error("trotzdem geschrieben")
		}
	})

	t.Run("das Deck für die Namen", func(t *testing.T) {
		werk, anlage, wunsch := wunschwerk(t)
		werk.legeDruck(t, druckA, 4)
		werk.decks.deckfehlerAb = werk.decks.deckrufe + 2

		if _, fehler := wunsch.Wuenschen(t.Context(), werk.deck.ID); !errors.Is(fehler, errKaputt) {
			t.Errorf("Fehler = %v", fehler)
		}

		if anlage.rufe != 0 {
			t.Error("trotzdem geschrieben")
		}
	})

	t.Run("das Schreiben", func(t *testing.T) {
		werk, anlage, wunsch := wunschwerk(t)
		werk.legeDruck(t, druckA, 4)
		anlage.fehler = domain.ErrConflict

		if _, fehler := wunsch.Wuenschen(t.Context(), werk.deck.ID); !errors.Is(fehler, domain.ErrConflict) {
			t.Errorf("Fehler = %v", fehler)
		}
	})
}

// **Der Abgleich zählt die Wünsche mit — getrennt vom Besitz** (0057): sie
// fehlen weiter, und die Wunschliste weiß, dass sie schon darauf stehen.
func TestDerAbgleichZaehltDieWuenscheGetrennt(t *testing.T) {
	werk := neuesPruefwerk(t)

	werk.wuensche(druckA, 2)
	werk.wuensche(druckA, 1)
	werk.besitze(druckA, 1)
	werk.legeDruck(t, druckA, 4)

	abgleich, fehler := werk.abgleich.Fuer(t.Context(), werk.deck.ID)
	if fehler != nil {
		t.Fatalf("abgleichen: %v", fehler)
	}

	if zeile := abgleich.Zeilen[0]; zeile.Vorhanden != 1 || zeile.Fehlend != 3 || zeile.Gewuenscht != 3 {
		t.Errorf("%+v", zeile)
	}
}
