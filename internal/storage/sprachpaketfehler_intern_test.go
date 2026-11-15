package storage

// **Eine Zeile, die sich nicht lesen lässt, meldet sich** — auf beiden
// Lesewegen, und ohne dass die Liste halb zurückkommt (LOC-02, #1688).

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEineUnlesbareSprachpaketzeileMeldetSich(t *testing.T) {
	db := migrierteDatenbank(t)

	for _, anweisung := range []string{
		`DROP TABLE sprachpakete`,
		`CREATE TABLE sprachpakete (kennung TEXT, name TEXT, basis TEXT, stand TEXT, lizenz TEXT,
		   herkunft TEXT, texte TEXT, angelegt TEXT, geaendert TEXT)`,
		`INSERT INTO sprachpakete (kennung, name) VALUES ('da', NULL)`,
	} {
		if _, fehler := db.ExecContext(probekontext(t), anweisung); fehler != nil {
			t.Fatal(fehler)
		}
	}

	ablage := NeueSprachpaketablage(db)

	if pakete, fehler := ablage.Sprachpakete(probekontext(t)); fehler == nil || pakete != nil {
		t.Errorf("die Liste: %v, %v", pakete, fehler)
	}

	if _, fehler := ablage.Sprachpaket(probekontext(t), "da"); fehler == nil ||
		errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("gelesen: %v", fehler)
	}

	if fehler := ablage.SprachpaketAendern(probekontext(t), "da",
		func(*domain.Sprachpaket) error { return nil }); fehler == nil || errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("geändert: %v", fehler)
	}
}

// **Scheitert das Schreiben im Zug, bleibt der Zug ohne Wirkung.**
func TestEinScheiterndesSchreibenImZugMeldetSich(t *testing.T) {
	db := migrierteDatenbank(t)
	ablage := NeueSprachpaketablage(db)

	if fehler := ablage.SprachpaketAnlegen(probekontext(t), domain.Sprachpaket{
		Kennung: "da", Name: "Dansk", Basis: "en", Stand: domain.StandEntwurf,
	}); fehler != nil {
		t.Fatal(fehler)
	}

	if _, fehler := db.ExecContext(probekontext(t), `CREATE TRIGGER nein BEFORE UPDATE ON sprachpakete
		BEGIN SELECT RAISE(ABORT, 'nein'); END`); fehler != nil {
		t.Fatal(fehler)
	}

	if fehler := ablage.SprachpaketAendern(probekontext(t), "da", func(paket *domain.Sprachpaket) error {
		paket.Name = "Fort"

		return nil
	}); fehler == nil {
		t.Error("das Schreiben ging durch")
	}
}
