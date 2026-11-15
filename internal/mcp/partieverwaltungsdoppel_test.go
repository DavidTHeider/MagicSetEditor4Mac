package mcp_test

// **Die vier Verwaltungswege an der Partienprobe** — PLAY-03.AK-05
// (#1711).
//
// **Eine eigene Datei**, damit `partiewerkzeug_test.go` unter der
// Zeilengrenze aus §2 bleibt. Der Typ bleibt einer.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

func (p *partienprobe) PartieLoeschen(_ context.Context, kennung string) (
	mcp.Partieverwaltungsangaben, error,
) {
	p.gesehen.kennung = kennung

	return mcp.Partieverwaltungsangaben{Befunde: []string{}}, p.klagt
}

func (p *partienprobe) PartieKopieren(_ context.Context, kennung string) (
	mcp.Partieverwaltungsangaben, error,
) {
	p.gesehen.kennung = kennung

	return mcp.Partieverwaltungsangaben{
		Kennung: kennung + "-kopie", Befunde: []string{},
	}, p.klagt
}

func (p *partienprobe) PartieAusfuhr(_ context.Context, kennung string) (
	mcp.Partieverwaltungsangaben, error,
) {
	p.gesehen.kennung = kennung

	return mcp.Partieverwaltungsangaben{
		Kennung: kennung, Inhalt: `{"fassung":1}`, Befunde: []string{},
	}, p.klagt
}

func (p *partienprobe) PartieEinfuhr(_ context.Context, inhalt string) (
	mcp.Partieverwaltungsangaben, error,
) {
	p.gesehen.inhalt = inhalt

	return mcp.Partieverwaltungsangaben{
		Kennung: "eingelesen", Befunde: []string{"ein fremdes Feld"},
	}, p.klagt
}
