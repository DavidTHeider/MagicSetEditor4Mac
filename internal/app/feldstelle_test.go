package app_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ein Feld sagt, zu welcher Reihe es gehört** — Werkstück 3 von Phase 3.6
// (#1167).
//
// # Warum das an der Fassade steht und nicht im Fenster
//
// Die Regel — „ein Feld `F N` gehört zur Reihe von `F`, wenn es `F` gibt" —
// steht in `domain`. Sie ein zweites Mal in TypeScript zu schreiben hieße,
// zwei Fassungen zu haben, die an dem Tag auseinanderlaufen, an dem jemand
// eine davon ändert (§2).
func TestEineFeldbeschreibungNenntIhreReiheUndStelle(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	anwendung.MitStilwahl(&stilwaehler{felder: []domain.Field{
		{Name: "name"},
		{Name: "name 2"},
		{Name: "rarity"},
		{Name: "loyalty cost 16"},
	}})

	antwort := anwendung.Kartenfelder(t.Context(), "1")
	if antwort.Fehler != nil {
		t.Fatalf("Fehler: %+v", antwort.Fehler)
	}

	nach := map[string]app.FeldbeschreibungDTO{}
	for _, feld := range antwort.Felder {
		nach[feld.Name] = feld
	}

	// **Das Grundfeld trägt seine eigene Reihe und die Stelle 1.**
	//
	// Nicht „keine Reihe": es **ist** die erste Stelle, und wer das Feld
	// ansieht, soll sehen, dass es Geschwister hat.
	if nach["name"].Reihe != "name" || nach["name"].Stelle != 1 {
		t.Errorf("name: Reihe %q, Stelle %d", nach["name"].Reihe, nach["name"].Stelle)
	}

	if nach["name 2"].Reihe != "name" || nach["name 2"].Stelle != 2 {
		t.Errorf("name 2: Reihe %q, Stelle %d", nach["name 2"].Reihe, nach["name 2"].Stelle)
	}

	// **Ein Feld ohne Geschwister trägt keine Reihe.** Eine Reihe aus einem
	// Feld wäre eine Auskunft, die nichts sagt — und das Fenster zeigte an
	// jedem Feld ein „1. Stelle", das niemand braucht.
	if nach["rarity"].Reihe != "" || nach["rarity"].Stelle != 0 {
		t.Errorf("rarity: Reihe %q, Stelle %d", nach["rarity"].Reihe, nach["rarity"].Stelle)
	}

	// **Und `loyalty cost 16` ist keine sechzehnte Stelle** — es gibt kein
	// Feld `loyalty cost`.
	if nach["loyalty cost 16"].Reihe != "" || nach["loyalty cost 16"].Stelle != 0 {
		t.Errorf("loyalty cost 16: Reihe %q, Stelle %d",
			nach["loyalty cost 16"].Reihe, nach["loyalty cost 16"].Stelle)
	}
}
