// **Die Kartensuche für die Flächen ohne Fenster** — Werkstück 1 von
// Phase 3.6
// (#1167).
//
// **Dieselbe Fähigkeit, überall dieselbe Antwort** (§16): was das Fenster
// kann, können Kommandozeile, MCP und Dienst. Die Naht baut den Zugang je
// Aufruf, weil es dort keinen Start gibt, an dem sie einmal entstünde.

package app

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schluesselbund"
)

// Fremdsuche ist die Kartensuche für eine Fläche ohne Fenster.
type Fremdsuche struct{ quelle Suchquelle }

// Fremdsuchnaht baut sie samt Klient und Zwischenspeicher.
//
// **Sie braucht den Bestand, seit es Dienstschalter gibt** (Phase 4.8): dort
// steht, ob Scryfall gefragt werden darf. Ohne ihn ginge ein abgeschalteter
// Dienst an drei Flächen weiter hinaus, und die Zusage aus
// [0063](../../docs/entscheidungen/0063-was-ein-dienst-hier-ist.md) gälte nur
// im Fenster.
func Fremdsuchnaht(_ context.Context, db *sql.DB) (*Fremdsuche, error) {
	// **Über den Helfer in `scryfallverdrahtung.go`** — die Aufzählung der
	// Dateien, die `internal/storage` kennen dürfen, soll nicht mit jeder
	// Naht wachsen (§2).
	wahl := dienstwahlUeber(db)

	// **Der Kontext bleibt in der Aufschrift** — die drei Flächen rufen sie
	// mit einem, und die Naht liest nichts mehr beim Bauen (Befund 1 des
	// Zweitblicks über Phase 4.8). Ein Fehlerweg bleibt ebenso: er ist die
	// Aufschrift, die diese Flächen erwarten, und ein Umbau von drei
	// Aufrufern für einen Rückgabewert, der heute immer `nil` ist, wäre
	// Arbeit ohne Ertrag.
	return &Fremdsuche{
		quelle: Suchnaht(wahl, schluesselbund.Anmeldebund(), wahl),
	}, nil
}

// MitSuchquelle setzt die Naht — für Tests.
func (f *Fremdsuche) MitSuchquelle(quelle Suchquelle) *Fremdsuche {
	f.quelle = quelle

	return f
}

// Suchergebnis ist, was eine Fläche ohne Fenster bekommt.
//
// **Dieselben Felder wie [ScryfallsuchAntwort]**, ohne den Fehler: die Flächen
// ohne Fenster tragen ihn als `error` (§8).
type Suchergebnis struct {
	Treffer      []SuchtrefferDTO `json:"treffer"`
	Weitere      bool             `json:"weitere"`
	AlterStunden int              `json:"alterStunden"`
	Befunde      []string         `json:"befunde"`
}

// Suche fragt Scryfall nach Karten.
func (f *Fremdsuche) Suche(ctx context.Context, frage string) (Suchergebnis, error) {
	auskunft, err := f.quelle.Suche(ctx, frage)
	if err != nil {
		return Suchergebnis{}, err
	}

	treffer := make([]SuchtrefferDTO, 0, len(auskunft.Treffer))
	for _, eines := range auskunft.Treffer {
		treffer = append(treffer, SuchtrefferDTO{
			Kennung: eines.Kennung, Name: eines.Name,
			Setcode: eines.Setcode, Setname: eines.Setname,
			Nummer: eines.Nummer, Seltenheit: eines.Seltenheit,
			Manakosten: eines.Manakosten, Typzeile: eines.Typzeile,
			Bild: eines.Bild,
		})
	}

	return Suchergebnis{
		Treffer:      treffer,
		Weitere:      auskunft.Weitere,
		AlterStunden: int(auskunft.Alter.Hours()),
		Befunde:      befundtexte(auskunft.Befunde),
	}, nil
}
