package app

// **Die Vorlage der eigenen Übersetzungen** — LOC-02
// (#1688).
//
// `sprachvorlage.json` ist der deutsche Katalog des Fensters, flach — und
// `frontend/src/sprache/sprachvorlage.test.ts` hält ihn gegen die Kataloge.
// **Eingebettet**, damit Kommandozeile, REST und MCP ohne Fenster gegen
// dieselbe Vorlage prüfen wie das Fenster.

import (
	_ "embed"
	"encoding/json"
	"sync"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

//go:embed sprachvorlage.json
var sprachvorlageRoh []byte

// Sprachvorlage gibt den deutschen Katalog flach — einmal gelesen.
//
// **Ohne Fehlerweg**: die Datei ist eingebettet, und der Lauf im Fenster
// schreibt sie selbst; eine, die sich nicht lesen ließe, käme nicht durch
// die Pipeline (`TestDieSprachvorlageIstDerDeutscheKatalog`).
var Sprachvorlage = sync.OnceValue(func() domain.Katalogvorlage {
	vorlage := domain.Katalogvorlage{}
	_ = json.Unmarshal(sprachvorlageRoh, &vorlage) //nolint:errcheck // siehe oben

	return vorlage
})
