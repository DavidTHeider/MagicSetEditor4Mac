package storage

// **Die Fehlerwege der gerechneten Massenänderung** — #2516. Dieselben
// Stolperstellen wie bei `FelderSetzen` (`stolpermassen_test.go`): eine
// Massenänderung, die einen Fehler schluckte, hinterließe einen halben
// Bestand.

import (
	"database/sql"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func gerechneterSchritt() []domain.Kartenschritt {
	return []domain.Kartenschritt{{KarteID: 1, Werte: []domain.Value{{Field: "typ", Text: "Artefakt"}}}}
}

func TestEineGerechneteMassenaenderungMeldetDieFehlerpfade(t *testing.T) {
	for _, treiber := range []string{
		"stolper-fragen", "stolper-schreiben", "stolper-feld-schreiben", "stolper-commit", "stolper-feld-commit",
	} {
		t.Run(treiber, func(t *testing.T) {
			db, err := sql.Open(treiber, "")
			if err != nil {
				t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
			}

			t.Cleanup(func() { _ = db.Close() })

			zweiterSchreibzaehler = 0

			if _, fehler := NeuerKartenschreiber(db).FelderGerechnetSetzen(
				probekontext(t), gerechneterSchritt(), nil, "typ"); fehler == nil {
				t.Error("die Massenänderung kam ohne Fehler zurück")
			}
		})
	}
}

// **Auf einer geschlossenen Datenbank beginnt keine Transaktion.**
func TestEineGerechneteMassenaenderungAufEinerGeschlossenenDatenbank(t *testing.T) {
	if _, fehler := NeuerKartenschreiber(geschlosseneDatenbank(t)).FelderGerechnetSetzen(
		probekontext(t), gerechneterSchritt(), nil, "typ"); fehler == nil {
		t.Error("die Massenänderung begann auf einer geschlossenen Datenbank")
	}
}
