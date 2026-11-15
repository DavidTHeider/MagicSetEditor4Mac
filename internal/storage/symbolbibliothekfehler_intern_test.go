package storage

// **Die Fehlerwege der Symbolbibliothek** — DOM-02 (#1697): ohne Profil,
// ohne Tafel und mit einer Zeile, die sich nicht lesen lässt. Ein innerer
// Lauf, weil er die Tafel gegen eine lockere tauscht.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// jederSymbolweg ruft jeden Weg und nennt die, die ohne Fehler durchkamen.
func jederSymbolweg(t *testing.T, ablage *Symbolbibliothek, ohneProfil bool) []string {
	t.Helper()

	ctx := probekontext(t)
	if ohneProfil {
		ctx = t.Context()
	}

	durch := []string{}
	pruefe := func(name string, fehler error) {
		if fehler == nil {
			durch = append(durch, name)
		}
	}

	_, fehler := ablage.Symbole(ctx)
	pruefe("Symbole", fehler)
	_, fehler = ablage.Symbol(ctx, 1)
	pruefe("Symbol", fehler)
	_, fehler = ablage.SymbolAnlegen(ctx, domain.Bibliothekssymbol{Name: "x", Inhalt: "x", Abdruck: "x"})
	pruefe("SymbolAnlegen", fehler)
	pruefe("SymbolUmbenennen", ablage.SymbolUmbenennen(ctx, 1, "y"))
	pruefe("SymbolInhaltSetzen", ablage.SymbolInhaltSetzen(ctx, 1, "y", "y"))
	pruefe("SymbolLoeschen", ablage.SymbolLoeschen(ctx, 1))

	return durch
}

func TestOhneProfilGehtKeinSymbolweg(t *testing.T) {
	if durch := jederSymbolweg(t, NeueSymbolbibliothek(migrierteDatenbank(t)), true); len(durch) != 0 {
		t.Errorf("ohne Profil kamen durch: %v", durch)
	}
}

func TestOhneTafelGehtKeinSymbolweg(t *testing.T) {
	db := migrierteDatenbank(t)
	if _, fehler := db.ExecContext(probekontext(t), `DROP TABLE symbolbibliothek`); fehler != nil {
		t.Fatal(fehler)
	}

	if durch := jederSymbolweg(t, NeueSymbolbibliothek(db), false); len(durch) != 0 {
		t.Errorf("ohne Tafel kamen durch: %v", durch)
	}
}

// **Eine Zeile, die sich nicht lesen lässt**, meldet sich — beim Aufzählen
// und beim Lesen eines einzelnen.
func TestEineUnlesbareSymbolzeileMeldetSich(t *testing.T) {
	db := migrierteDatenbank(t)

	for _, anweisung := range []string{
		`DROP TABLE symbolbibliothek`,
		`CREATE TABLE symbolbibliothek (id INTEGER, profil_id INTEGER, name TEXT, inhalt TEXT,
		   abdruck TEXT, vorher TEXT, angelegt TEXT, geaendert TEXT)`,
		`INSERT INTO symbolbibliothek (id, profil_id, name) VALUES (1, 1, NULL)`,
	} {
		if _, fehler := db.ExecContext(probekontext(t), anweisung); fehler != nil {
			t.Fatal(fehler)
		}
	}

	ablage := NeueSymbolbibliothek(db)

	if _, fehler := ablage.Symbole(probekontext(t)); fehler == nil {
		t.Error("die Liste las eine leere Zeile")
	}

	if _, fehler := ablage.Symbol(probekontext(t), 1); fehler == nil {
		t.Error("das Symbol las eine leere Zeile")
	}
}
