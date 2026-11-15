package rest_test

// **Die Doppel der Schnittstellen** — OPS-09 (#1907): willig, zählend,
// unwillig, wie bei jeder anderen Naht.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

func (williger) Schnittstellen(context.Context) ([]rest.Schnittstelle, error) {
	return []rest.Schnittstelle{{Name: "codex", Stand: "offen", Datei: "/x/.codex/config.toml", Zeilen: []string{"a"}}}, nil
}

func (z *zaehlender) Schnittstellen(context.Context) ([]rest.Schnittstelle, error) {
	z.zaehle()

	return []rest.Schnittstelle{}, nil
}

func (u unwilliger) Schnittstellen(context.Context) ([]rest.Schnittstelle, error) {
	return nil, u.fehler
}
