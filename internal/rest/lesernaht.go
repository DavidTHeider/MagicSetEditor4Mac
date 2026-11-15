// Die Naht zu den lesenden Endpunkten.
//
// **Eine eigene Datei**, weil `naehte.go` an der Zeilengrenze aus §2 stand.

package rest

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Leser ist, was die lesenden Endpunkte brauchen.
//
// Die Schnittstelle sagt, was `rest` **braucht** — nicht, was `storage`
// anzubieten hat (§2).
type Leser interface {
	Setkoepfe(ctx context.Context) ([]domain.SetHead, error)
	Karten(ctx context.Context, setID int64) ([]*domain.CardRow, error)

	// **Und der Volltext** — seit
	// #1107.
	//
	// Beide standen am MCP-Server (`lesen.felder`, `lesen.suche`) und hier
	// nicht. Der Leser dahinter ist derselbe; gefehlt hat, dass jemand sie
	// verlangt.
	Felder(ctx context.Context, setID int64) ([]string, error)
	Suche(ctx context.Context, frage domain.SearchQuery) (domain.SearchResult, error)
}
