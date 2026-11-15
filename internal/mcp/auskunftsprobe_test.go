// **Ein Doppel für die lesenden Auskünfte**
// (#888).
//
// **Es steht hier und nicht in jeder Datei**, die die Werkzeugliste baut: sie
// bekommen alle dieselbe Antwort, und eine Kopie je Datei liefe auseinander,
// sobald jemand ein Feld ergänzt.
package mcp_test

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// stilleAuskunft antwortet auf jede Frage dasselbe.
type stilleAuskunft struct{ fehler error }

// Regelhinweise gibt einen offiziellen und einen inoffiziellen Hinweis.
//
// **Beide**, weil die Unterscheidung die Zusage des Werkzeugs ist: eine
// Anmerkung des Scryfall-Teams ist keine Regel, und ein Doppel mit nur einem
// offiziellen Hinweis ließe das ungeprüft.
func (s stilleAuskunft) Regelhinweise(context.Context, string) (mcp.Regelauskunft, error) {
	if s.fehler != nil {
		return mcp.Regelauskunft{}, s.fehler
	}

	return mcp.Regelauskunft{
		Kartenname: "Prüfstein",
		Hinweise: []mcp.Regelhinweis{
			{Offiziell: true, Quelle: "wotc", Datum: "2004-10-04", Text: "Offiziell."},
			{Offiziell: false, Quelle: "scryfall", Datum: "2021-07-23", Text: "Anmerkung."},
		},
		Adresse: "https://beispiel.invalid/karte/1",
		Befunde: []string{},
	}, nil
}

// Manakosten meldet einen Kosten ohne Beanstandung.
func (s stilleAuskunft) Manakosten(context.Context, string) ([]string, error) {
	return nil, s.fehler
}

// stilleSammeldaten meldet eine Datei, die bereitliegt.
type stilleSammeldaten struct{ fehler error }

// Stand sagt, was liegt.
func (s stilleSammeldaten) Stand(context.Context) (mcp.Sammelstand, error) {
	if s.fehler != nil {
		return mcp.Sammelstand{}, s.fehler
	}

	return mcp.Sammelstand{
		Art: "oracle_cards", Pfad: "/probe/oracle_cards.jsonl.gz",
		Vorhanden: true, Bytes: 24531600, Angeboten: 24531600,
		Stand: "2026-08-23T21:01:56.596+00:00",
	}, nil
}

// Holen tut dasselbe — das Doppel lädt nichts.
func (s stilleSammeldaten) Holen(ctx context.Context) (mcp.Sammelstand, error) {
	return s.Stand(ctx)
}
