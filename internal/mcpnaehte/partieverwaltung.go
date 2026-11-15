package mcpnaehte

// **Die vier Verwaltungswege einer Partie am MCP-Server** — PLAY-03.AK-05
// (#1711).
//
// **Über dieselbe kopflose Anwendung wie die fünf daneben** und nicht über
// einen eigenen Weg: was ein Modell tut, soll dasselbe sein, was ein Mensch
// im Fenster tut (§16).
//
// **Ausfuhr und Einfuhr gehen über den Text**: ein MCP-Server spricht über
// stdio, und ein Pfad wäre dort einer auf einem Rechner, den der Rufende
// womöglich nicht sieht.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

func (p partien) PartieLoeschen(ctx context.Context, kennung string) (
	mcp.Partieverwaltungsangaben, error,
) {
	return verwaltungAus(
		app.PartieflaecheUeber(p.db).PartieLoeschen(ctx, kennung))
}

func (p partien) PartieKopieren(ctx context.Context, kennung string) (
	mcp.Partieverwaltungsangaben, error,
) {
	return verwaltungAus(
		app.PartieflaecheUeber(p.db).PartieKopieren(ctx, kennung))
}

func (p partien) PartieAusfuhr(ctx context.Context, kennung string) (
	mcp.Partieverwaltungsangaben, error,
) {
	return verwaltungAus(
		app.PartieflaecheUeber(p.db).PartieAusfuhrText(ctx, kennung))
}

func (p partien) PartieEinfuhr(ctx context.Context, inhalt string) (
	mcp.Partieverwaltungsangaben, error,
) {
	return verwaltungAus(
		app.PartieflaecheUeber(p.db).PartieEinfuhrText(ctx, inhalt))
}

// verwaltungAus bildet die Antwort der Anwendung auf die des Servers ab.
//
// **Der Pfad geht nicht mit**: hier wird keine Datei angefasst, und ein
// leeres Feld wäre eine Auskunft über etwas, das gar nicht geschehen ist.
func verwaltungAus(antwort app.PartieverwaltungAntwort) (
	mcp.Partieverwaltungsangaben, error,
) {
	if antwort.Fehler != nil {
		return mcp.Partieverwaltungsangaben{}, antwort.Fehler.AlsFehler()
	}

	befunde := antwort.Befunde
	if befunde == nil {
		befunde = []string{}
	}

	return mcp.Partieverwaltungsangaben{
		Kennung: antwort.Kennung, Inhalt: antwort.Inhalt, Befunde: befunde,
	}, nil
}
