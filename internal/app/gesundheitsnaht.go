package app

// **Die Gesundheitsnaht über einem Bestand** — Werkstück O9 von Phase 5.2
// (#1433).
//
// # Warum sie hier steht und nicht bei jedem Aufrufer
//
// **Weil `cmd/` den Dienst nicht kennen darf** (§2): erlaubt sind `app`,
// `domain`, `storage` — und die Schichtenregel hat es angestrichen, als
// `cmd/mse/gesundheit.go` `service` importierte. Dieselbe Gestalt wie
// [SchemastandUeber]: eine Zeile, die die zwei Schichten verbindet, an genau
// einer Stelle.

import (
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// GesundheitsnahtUeber baut die Schau über einer offenen Datenbank.
//
// **Ohne Fremdschau**, und das ist an jeder Fläche dieselbe Entscheidung: was
// über fremde Dienste bekannt ist, führt der Dienstkatalog (A1). Ihn hier zu
// fragen hieße, eine Zustandsabfrage an eine Netzabfrage zu hängen — und dann
// dauerte „lebst du" so lange wie der langsamste fremde Server (§16, Regel 3).
func GesundheitsnahtUeber(db *sql.DB) Gesundheitsnaht {
	return service.NeueBetriebsschau(storage.NeueBetriebslage(db), nil)
}
