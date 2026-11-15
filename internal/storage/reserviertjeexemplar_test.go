package storage_test

// **Was insgesamt festgehalten ist, je Exemplar** — Werkstück `T2` von Phase
// 5.6 (#1467).
//
// # Warum eine eigene Frage neben `Reservierungen`
//
// **Weil die Deckzuordnung sie in **einem** Ruf braucht.** Sie rechnet über
// die ganze Sammlung; eine Frage je Zeile wäre ein Ruf je Exemplar, und wer
// ihn vergisst, zeigt eine Sammlung ohne Zusagen, ohne dass jemand es merkt —
// dieselbe Überlegung wie bei `ZugeordnetJeExemplar` (U9, #1420).

import (
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Zwei Vorgänge auf derselben Zeile zählen zusammen.**
//
// **Und eine abgelaufene Frist zählt nicht mit**: sie hält nichts mehr fest,
// und wer sie mitrechnete, zeigte eine Sammlung als gebunden, die frei ist.
func TestReserviertJeExemplarZaehltDieOffenenZusagen(t *testing.T) {
	db := frischeDatenbank(t)
	ctx := probekontext(t)

	id, err := storage.NeuerSammlungsspeicher(db).Anlegen(ctx, probeexemplar())
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	ablage := storage.NeueReservierungsablage(db)

	for _, zusage := range []domain.Reservierung{
		{Vorgang: "h1", Revision: 1, ExemplarID: id, Menge: 1, Fassung: 1},
		{Vorgang: "h2", Revision: 1, ExemplarID: id, Menge: 1, Fassung: 1},
		{
			Vorgang: "h3", Revision: 1, ExemplarID: id, Menge: 1, Fassung: 1,
			Frist: time.Now().Add(-time.Hour),
		},
	} {
		if fehler := ablage.Festhalten(ctx, zusage, ueber(probeexemplar(), 99)); fehler != nil {
			t.Fatalf("%s festhalten: %v", zusage.Vorgang, fehler)
		}
	}

	gezaehlt, fehler := ablage.ReserviertJeExemplar(ctx)
	if fehler != nil {
		t.Fatalf("zählen: %v", fehler)
	}

	if gezaehlt[id] != 2 {
		t.Errorf("festgehalten sind %d Stück, erwartet waren 2 — %+v",
			gezaehlt[id], gezaehlt)
	}
}

// **Ohne Zusage steht nichts drin — und keine Null.**
//
// **Eine leere Tabelle und kein `nil`**: die Deckzuordnung liest daraus mit
// `map[…]`, und ein `nil` gäbe dort zwar null, aber ein Aufrufer, der die
// Länge prüft, bekäme eine Aussage über etwas anderes.
func TestOhneZusageIstDieTabelleLeer(t *testing.T) {
	gezaehlt, fehler := storage.NeueReservierungsablage(frischeDatenbank(t)).
		ReserviertJeExemplar(probekontext(t))
	if fehler != nil {
		t.Fatalf("zählen: %v", fehler)
	}

	if gezaehlt == nil || len(gezaehlt) != 0 {
		t.Errorf("die Tabelle ist %+v", gezaehlt)
	}
}
