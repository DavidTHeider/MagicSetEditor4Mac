package domain_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Eine Teilmenge ausgeben** — Werkstück 4 von Phase 3.6
// (#1167).

func kartenreihe(kennungen ...int64) []*domain.CardRow {
	karten := make([]*domain.CardRow, 0, len(kennungen))
	for _, kennung := range kennungen {
		karten = append(karten, &domain.CardRow{ID: kennung})
	}

	return karten
}

func kennungen(karten []*domain.CardRow) []int64 {
	gelesen := make([]int64, 0, len(karten))
	for _, karte := range karten {
		gelesen = append(gelesen, karte.ID)
	}

	return gelesen
}

// **Eine leere Auswahl ist das ganze Set.**
//
// # Warum das der Nullwert ist und nicht ein Sonderfall
//
// Weil MSE2 es so hält: `EXPORT_SEL_ENTIRE_SET` steht als **erste** Wahl in
// jedem Ausgabefenster (`SetWindow::selectionChoices`,
// `src/gui/set/window.cpp`, Zeilen 383–389, Fassung f9d9356 vom 2021-07-06).
// Wer nichts wählt, meint alles — und ein Aufrufer, der die Auswahl nicht
// kennt, bekommt damit das gewohnte Verhalten.
func TestEineLeereAuswahlIstDasGanzeSet(t *testing.T) {
	t.Parallel()

	var auswahl domain.Kartenauswahl

	if auswahl.Filtert() {
		t.Error("die leere Auswahl filtert")
	}

	gewaehlt := auswahl.Anwenden(kartenreihe(1, 2, 3))
	if len(gewaehlt) != 3 {
		t.Errorf("es kamen %v", kennungen(gewaehlt))
	}
}

// **Eine Auswahl nimmt, was genannt ist — in der Reihenfolge des Sets.**
//
// # Warum die Reihenfolge des Sets und nicht die der Nennung
//
// Ein Druckbogen legt die Karten in Setreihenfolge; wer die Auswahl in seiner
// eigenen Reihenfolge anwendete, bekäme einen Bogen, dessen Karten anders
// liegen, je nachdem, in welcher Reihenfolge jemand angehakt hat.
func TestEineAuswahlNimmtInSetreihenfolge(t *testing.T) {
	t.Parallel()

	auswahl := domain.Kartenauswahl{Karten: []int64{3, 1}}

	gewaehlt := auswahl.Anwenden(kartenreihe(1, 2, 3))

	if len(gewaehlt) != 2 || gewaehlt[0].ID != 1 || gewaehlt[1].ID != 3 {
		t.Errorf("es kamen %v", kennungen(gewaehlt))
	}
}

// **Eine Kennung, die es im Set nicht gibt, fällt weg.**
//
// Sie zu melden wäre die andere Möglichkeit — und die schlechtere: die Auswahl
// kommt aus einer Liste, die jemand vor dem Löschen einer Karte angehakt hat,
// und ein Druckbogen soll daran nicht scheitern (§8).
func TestEineFremdeKennungFaelltWeg(t *testing.T) {
	t.Parallel()

	auswahl := domain.Kartenauswahl{Karten: []int64{1, 99}}

	gewaehlt := auswahl.Anwenden(kartenreihe(1, 2))
	if len(gewaehlt) != 1 || gewaehlt[0].ID != 1 {
		t.Errorf("es kamen %v", kennungen(gewaehlt))
	}
}

// **Eine Auswahl, die keine Karte des Sets trifft, ist ein Fehler.**
//
// # Warum nicht „dann eben das ganze Set"
//
// Weil das die falsche Richtung wäre: wer eine Auswahl mitgibt, hat sich
// entschieden. Ein Bogen über alle 300 Karten, weil die Auswahl nicht passte,
// wäre die teuerste Art, einen Tippfehler zu beantworten.
func TestEineAuswahlOhneTrefferWirdGemeldet(t *testing.T) {
	t.Parallel()

	auswahl := domain.Kartenauswahl{Karten: []int64{99}}

	if fehler := auswahl.Pruefe(kartenreihe(1, 2)); fehler == nil {
		t.Fatal("die leere Treffermenge kam als Erfolg zurück")
	}

	// **Und ohne Auswahl gibt es nichts zu melden** — auch nicht bei einem
	// leeren Set: ein Set ohne Karten ergibt einen Bogen ohne Seiten, und das
	// ist eine andere Auskunft.
	var ohne domain.Kartenauswahl
	if fehler := ohne.Pruefe(nil); fehler != nil {
		t.Errorf("die leere Auswahl meldete %v", fehler)
	}
}

// **Doppelte Kennungen zählen einmal.** Sie entstehen an einer Fläche, die
// zweimal anhakt; eine Karte zweimal zu drucken wäre keine Auswahl, sondern
// eine Vervielfältigung.
func TestDoppelteKennungenZaehlenEinmal(t *testing.T) {
	t.Parallel()

	auswahl := domain.Kartenauswahl{Karten: []int64{2, 2}}

	if gewaehlt := auswahl.Anwenden(kartenreihe(1, 2)); len(gewaehlt) != 1 {
		t.Errorf("es kamen %v", kennungen(gewaehlt))
	}
}

// **Eine Auswahl, die trifft, geht durch.**
//
// Die Gegenprobe zur Zusage darüber: ohne sie wäre auch eine `Pruefe` grün,
// die **jede** Auswahl abwiese.
func TestEineTreffendeAuswahlGehtDurch(t *testing.T) {
	t.Parallel()

	auswahl := domain.Kartenauswahl{Karten: []int64{2}}

	if fehler := auswahl.Pruefe(kartenreihe(1, 2)); fehler != nil {
		t.Errorf("die treffende Auswahl meldete %v", fehler)
	}
}
