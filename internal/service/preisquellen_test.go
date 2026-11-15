package service_test

// **Eine Quelle widerruft nur ihre eigenen Reihen** — Werkstück A3 von Phase
// 5.3 (#1442).
//
// # Warum das ein eigener Lauf ist
//
// Bis hierher gab es **eine** Quelle, und „was die Antwort nicht nennt, ist
// weggefallen" war richtig. Mit einer zweiten wird daraus ein Fehler, den
// niemand sieht: ein Abruf bei Scryfall widerriefe die Reihen von MTGJSON, und
// im Portfolio verschwänden sie — ohne Meldung, mit einem Wert von null, der
// nichts bedeutet.

import (
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// probedruckkennung ist der Druck, um den es in diesen Läufen geht.
const probedruckkennung = "00000000-0000-4000-8000-000000000042"

// fremdeQuelle ist eine zweite Quelle neben Scryfall.
const fremdeQuelle domain.Preisquelle = "mtgjson"

// standVon baut einen genannten Stand einer Quelle.
func standVon(
	quelle domain.Preisquelle, waehrung string, wert int64, erhoben time.Time,
) domain.Preisstand {
	return domain.Preisstand{
		Printing: probedruckkennung,
		Quelle:   quelle,
		Markt:    domain.MarktPapier,
		Art:      domain.PreisartVerkauf,
		Waehrung: waehrung,
		Finish:   domain.FinishNonfoil,
		Lage:     domain.PreisGenannt,
		Wert:     wert,
		Erhoben:  erhoben,
	}
}

// **Ein Abruf bei Scryfall lässt die Reihen von MTGJSON stehen.**
func TestEinAbrufWiderruftNurDieEigenenReihen(t *testing.T) {
	t.Parallel()

	vorher := time.Date(2026, 9, 10, 8, 0, 0, 0, time.UTC)
	jetzt := time.Date(2026, 9, 11, 8, 0, 0, 0, time.UTC)

	ablage := &preisdoppel{
		staende: []domain.Preisstand{
			standVon(domain.PreisquelleScryfall, "eur", 417, vorher),
			standVon(fremdeQuelle, "eur", 390, vorher),
		},
	}

	// **Scryfall antwortet und nennt nichts mehr.**
	pflege := service.NeuePreispflege(ablage, nil,
		quellendoppel{staende: []domain.Preisstand{}},
		func() time.Time { return jetzt })

	if _, _, _, fehler := pflege.Holen(t.Context(), probedruckkennung); fehler != nil {
		t.Fatalf("holen: %v", fehler)
	}

	nachQuelle := map[domain.Preisquelle][]domain.Preislage{}

	for _, stand := range ablage.staende {
		if stand.Erhoben.Equal(jetzt) {
			nachQuelle[stand.Quelle] = append(nachQuelle[stand.Quelle], stand.Lage)
		}
	}

	if len(nachQuelle[domain.PreisquelleScryfall]) != 1 ||
		nachQuelle[domain.PreisquelleScryfall][0] != domain.PreisWeggefallen {
		t.Errorf("die eigene Reihe wurde nicht widerrufen: %v",
			nachQuelle[domain.PreisquelleScryfall])
	}

	if len(nachQuelle[fremdeQuelle]) != 0 {
		t.Errorf("die fremde Reihe wurde angefasst: %v", nachQuelle[fremdeQuelle])
	}
}

// **Ein Sammler widerruft alle seine Anbieter.**
//
// MTGJSON trägt die Preise von fünf Händlern, und jeder ist eine eigene Reihe
// (`mtgjson:cardkingdom`). Nennt MTGJSON für einen Druck nichts mehr, ist
// **jede** davon weggefallen — ein Vergleich auf Gleichheit fände `mtgjson`
// und `mtgjson:cardkingdom` ungleich und widerriefe nichts.
func TestEinSammlerWiderruftAlleSeineAnbieter(t *testing.T) {
	t.Parallel()

	vorher := time.Date(2026, 9, 10, 8, 0, 0, 0, time.UTC)
	jetzt := time.Date(2026, 9, 11, 8, 0, 0, 0, time.UTC)

	ablage := &preisdoppel{
		staende: []domain.Preisstand{
			standVon("mtgjson:cardkingdom", "usd", 417, vorher),
			standVon("mtgjson:cardmarket", "eur", 390, vorher),
			standVon(domain.PreisquelleScryfall, "eur", 400, vorher),
		},
	}

	pflege := service.NeuePreispflege(ablage, nil,
		quellendoppel{quelle: "mtgjson", staende: []domain.Preisstand{}},
		func() time.Time { return jetzt })

	if _, _, _, fehler := pflege.Holen(t.Context(), probedruckkennung); fehler != nil {
		t.Fatalf("holen: %v", fehler)
	}

	widerrufen := map[domain.Preisquelle]bool{}

	for _, stand := range ablage.staende {
		if stand.Erhoben.Equal(jetzt) && stand.Lage == domain.PreisWeggefallen {
			widerrufen[stand.Quelle] = true
		}
	}

	for _, quelle := range []domain.Preisquelle{
		"mtgjson:cardkingdom", "mtgjson:cardmarket",
	} {
		if !widerrufen[quelle] {
			t.Errorf("%q wurde nicht widerrufen", string(quelle))
		}
	}

	// **Und die fremde Quelle bleibt** — die Gegenprobe.
	if widerrufen[domain.PreisquelleScryfall] {
		t.Error("Scryfall wurde mitwiderrufen")
	}
}
