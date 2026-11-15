// Der Klient für die Sammeldateien von Scryfall
// (#888).
//
// **Eine eigene Datei, weil `scryfall.go` mit ihr über 400 Zeilen ginge** (§2).
// Sie gehört trotzdem zum Composition Root: `internal/schichten` lässt die
// Kante auf `internal/scryfall` nur in benannten Dateien zu, und diese ist eine
// davon — die Aufzählung bleibt eine Aufzählung.

package app

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
)

// Sammelquellenaht baut den Klienten für die Sammeldateien.
//
// **Sie steht hier, weil `internal/scryfall` hier stehen darf** (§2) — dieselbe
// Aufzählung, die den Klienten begrenzt.
func Sammelquellenaht(wahl Dienstwahl) (Sammelquelle, error) {
	// **Über die eine Stelle, an der ein Klient entsteht** (Phase 4.8) — dort
	// wird auch der Dienstschalter geprüft, und eine eigene Konstruktion wäre
	// ein Loch in der Zusage „aus heißt null Anfragen".
	klient, fehler := Scryfallklient(wahl)
	if fehler != nil {
		return nil, fehler
	}

	return sammelquelle{klient: klient}, nil
}

// sammelquelle übersetzt zwischen der Naht und dem Klienten.
type sammelquelle struct{ klient *scryfall.Klient }

// Verzeichnis gibt die angebotenen Sammeldateien.
func (s sammelquelle) Verzeichnis(ctx context.Context) ([]Sammeleintrag, error) {
	eintraege, _, fehler := s.klient.Bulkverzeichnis(ctx)
	if fehler != nil {
		return nil, fehler
	}

	gestalt := make([]Sammeleintrag, 0, len(eintraege))
	for _, einer := range eintraege {
		gestalt = append(gestalt, Sammeleintrag{
			Art: einer.Art, Name: einer.Name, Beschreibung: einer.Beschreibung,
			Adresse: einer.Adresse, Bytes: einer.Groesse, Stand: einer.Stand,
		})
	}

	return gestalt, nil
}

// Holen lädt eine Sammeldatei.
func (s sammelquelle) Holen(ctx context.Context, eintrag Sammeleintrag, ziel string) error {
	return s.klient.BulkHolen(ctx, scryfall.Bulkeintrag{
		Art: eintrag.Art, Adresse: eintrag.Adresse, Groesse: eintrag.Bytes,
	}, ziel)
}
