// Die Naht zu den Auskünften für die Flächen ohne Fenster
// (#888).
//
// **Hier stand „ohne Bestand"**, weil eine Regelauskunft nichts aus der
// Datenbank liest. Das stimmt für die Auskunft und nicht mehr für die Frage
// davor: seit Phase 4.8 steht in den Einstellungen, **ob Scryfall überhaupt
// gefragt werden darf**
// ([0063](../../docs/entscheidungen/0063-was-ein-dienst-hier-ist.md)), und die
// Einstellungen liegen im Bestand.
//
// **Ohne ihn ginge ein abgeschalteter Dienst an dieser Fläche weiter hinaus.**

package app

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Fremdauskunft ist die Auskunft, wie eine Fläche ohne Fenster sie benutzt.
type Fremdauskunft struct{ quelle Auskunftsquelle }

// Fremdauskunftsnaht baut sie.
func Fremdauskunftsnaht(db *sql.DB) (*Fremdauskunft, error) {
	quelle, fehler := Auskunftsnaht(dienstwahlUeber(db))
	if fehler != nil {
		return nil, fehler
	}

	return &Fremdauskunft{quelle: quelle}, nil
}

// Regelauskunft ist, was eine Fläche ohne Fenster bekommt.
type Regelauskunft struct {
	Kartenname   string            `json:"kartenname"`
	Hinweise     []Regelhinweisang `json:"hinweise"`
	Adresse      string            `json:"adresse"`
	Befunde      []string          `json:"befunde"`
	AlterStunden int               `json:"alterStunden"`
}

// Regelhinweisang ist ein einzelner Hinweis.
//
// **`Offiziell` steht vor `Quelle`**, weil es die Frage ist, die zählt: eine
// Anmerkung des Scryfall-Teams erklärt ein inoffizielles Format, und sie als
// Regel weiterzugeben wäre falsch.
type Regelhinweisang struct {
	Offiziell bool   `json:"offiziell"`
	Quelle    string `json:"quelle"`
	Datum     string `json:"datum"`
	Text      string `json:"text"`
}

// Regelhinweise holt die Hinweise zu einer Karte.
func (f *Fremdauskunft) Regelhinweise(ctx context.Context, name string) (Regelauskunft, error) {
	geholt, fehler := f.quelle.Regelhinweise(ctx, name)
	if fehler != nil {
		return Regelauskunft{}, fehler
	}

	hinweise := make([]Regelhinweisang, 0, len(geholt.Hinweise))
	for _, einer := range geholt.Hinweise {
		hinweise = append(hinweise, Regelhinweisang{
			Offiziell: string(einer.Quelle) == scryfallOffiziell,
			Quelle:    string(einer.Quelle),
			Datum:     einer.Datum, Text: einer.Text,
		})
	}

	return Regelauskunft{
		Kartenname: geholt.Kartenname, Hinweise: hinweise, Adresse: geholt.Adresse,
		Befunde: alsMeldungen(geholt.Befunde), AlterStunden: int(geholt.Alter.Hours()),
	}, nil
}

// Manakosten prüft einen Kosten gegen die Symbologie.
func (f *Fremdauskunft) Manakosten(ctx context.Context, kosten string) ([]string, error) {
	return f.quelle.Manakostenpruefung(ctx, kosten)
}

// Satz holt die Angaben zu einem Satz.
func (f *Fremdauskunft) Satz(ctx context.Context, code string) (*service.Satzauskunft, error) {
	return f.quelle.Satz(ctx, code)
}
