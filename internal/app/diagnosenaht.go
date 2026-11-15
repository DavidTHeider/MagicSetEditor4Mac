package app

// **Der Schemastand für die Diagnose** — Befund 2 des Zweitblicks über
// Phase 3.5
// (#1211).
//
// **Eine eigene Datei, weil sie `internal/storage` kennt** (§2). Die Aufzählung
// der Composition Roots ist eine Aufzählung; was nicht darin steht, greift
// nicht auf den Speicher.

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// schemanaht liest die verbuchten Migrationen über einer offenen Datenbank.
type schemanaht struct{ db *sql.DB }

func (s schemanaht) AngewendeteMigrationen(ctx context.Context) ([]string, error) {
	return storage.AngewendeteMigrationen(ctx, s.db)
}

// Waisen zählt die Zeilen, die zu nichts mehr gehören (#1220).
func (s schemanaht) Waisen(ctx context.Context) (string, error) {
	zahlen, err := storage.Waisen(ctx, s.db)
	if err != nil {
		return "", err
	}

	return storage.WaisenText(zahlen), nil
}

// SchemastandUeber baut die Naht über einer offenen Datenbank.
func SchemastandUeber(db *sql.DB) Schemastand { return schemanaht{db: db} }
