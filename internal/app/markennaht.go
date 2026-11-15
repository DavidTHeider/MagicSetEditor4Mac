package app

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Die Marken für die Flächen ohne Fenster** —
// #1208.
//
// **Ein Composition Root und kein zweiter Anwendungsfall** (§2): was eine
// Marke ist und wer sie setzen darf, steht in `service.Markenpflege`. Hier
// steht nur, woraus sie über einem Bestand gebaut wird — und dass es
// **dieselbe** ist, die das Fenster benutzt.

// Fremdmarken sind die Marken, wie eine Fläche ohne Fenster sie benutzt.
//
// **Eine eigene Gestalt und nicht `*service.Markenpflege` unmittelbar**:
// `cmd/` darf `internal/service` nicht kennen (§2, Schichtenwächter), und
// diese Hülle ist die Naht, die es davon fernhält.
//
// **Die Namen tragen ihren Gegenstand** — `MarkeSetzen` und nicht `Setzen`,
// wie `VerweisSetzen` daneben. Eine Vorrichtung, die mehrere solcher Nähte
// belegt, hat sonst ein `Setzen`, das drei Dinge meinen könnte.
type Fremdmarken struct{ pflege *service.Markenpflege }

// Markennaht baut die Marken über einem Bestand.
func Markennaht(db *sql.DB) *Fremdmarken {
	return &Fremdmarken{
		pflege: service.NeueMarkenpflege(
			storage.NeueMarken(db), Kartenlesernaht(db), storage.NeuerSetLeser(db)),
	}
}

// MarkeSetzen hängt eine Marke an eine Karte.
func (f *Fremdmarken) MarkeSetzen(ctx context.Context, setID, karteID int64, marke string) error {
	return f.pflege.Setzen(ctx, setID, karteID, marke)
}

// MarkeEntfernen nimmt eine Marke wieder ab.
func (f *Fremdmarken) MarkeEntfernen(ctx context.Context, setID, karteID int64, marke string) error {
	return f.pflege.Entfernen(ctx, setID, karteID, marke)
}

// KartenMarken nennt die Marken einer Karte.
func (f *Fremdmarken) KartenMarken(ctx context.Context, setID, karteID int64) ([]string, error) {
	return f.pflege.AnKarte(ctx, setID, karteID)
}

// Setmarken nennt die Marken eines Sets.
func (f *Fremdmarken) Setmarken(ctx context.Context, setID int64) ([]string, error) {
	return f.pflege.ImSet(ctx, setID)
}
