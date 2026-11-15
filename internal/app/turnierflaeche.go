package app

// **Turniere für die drei Flächen ohne Fenster** — Werkstück 5 von Phase 4.7
// (#1294).
//
// # Warum das hier steht und nicht in jeder Fläche
//
// **Damit alle vier dieselbe Anwendung benutzen** (§16). Dieselbe Form wie
// `decknahtflaechen.go` und `deckwerkstatt.go`.
//
// **Turniere brauchen keine Gegenstelle.** Anders als die Formatprüfung
// hängen sie an nichts außerhalb des Bestands — es gibt hier deshalb keinen
// Zweig „ohne Netz geht es nicht".

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// Turnierflaeche ist die Anwendung selbst, für eine Fläche ohne Fenster.
type Turnierflaeche struct{ app *App }

// TurnierflaecheUeber baut sie über einem Bestand.
func TurnierflaecheUeber(db *sql.DB) *Turnierflaeche {
	anwendung := NeueApp("")
	anwendung.MitTurniernaht(service.NeueTurnierleitung(
		storage.NeueTurnierablage(db),
		storage.NeueDeckablage(db),
		storage.NeuerSetLeser(db),
	).MitFormaten(storage.NeueFormatablage(db)))

	// **Und die Brücke zu den Partien** (G7,
	// #1464).
	// Sie steht **neben** der Turnierleitung: ein Turnier, an dem niemand
	// hier spielt, berührt sie nie.
	anwendung.MitMatchnaht(service.NeueMatchleitung(
		storage.NeueTurnierablage(db),
		storage.NeuePartieablage(db),
		storage.NeueMatchablage(db),
	))

	// **Und die Datei eines Turniers** (DOM-04.AK-03, #1699) — dieselbe
	// Austauschdatei wie bei der Partie, mit der Grenze des Turniers.
	anwendung.MitTurnierdatei(service.NeueTurnieraustauschdatei())

	return &Turnierflaeche{app: anwendung}
}

func (t *Turnierflaeche) Turniere(ctx context.Context, spiel string) TurniereAntwort {
	return t.app.Turniere(ctx, spiel)
}

func (t *Turnierflaeche) Turnier(ctx context.Context, turnierID string) TurnierAntwort {
	return t.app.Turnier(ctx, turnierID)
}

func (t *Turnierflaeche) TurnierAnlegen(
	ctx context.Context, name, spiel, notiz string, einzelspiel bool, formatID string,
) TurnierAntwort {
	return t.app.TurnierAnlegen(ctx, name, spiel, notiz, einzelspiel, formatID)
}

func (t *Turnierflaeche) TurnierAbschliessen(ctx context.Context, turnierID, fassung string) SchreibAntwort {
	return t.app.TurnierAbschliessen(ctx, turnierID, fassung)
}

func (t *Turnierflaeche) TurnierWiedereroeffnen(ctx context.Context, turnierID, fassung string) SchreibAntwort {
	return t.app.TurnierWiedereroeffnen(ctx, turnierID, fassung)
}

func (t *Turnierflaeche) TurnierProtokoll(ctx context.Context, turnierID string) ProtokollAntwort {
	return t.app.TurnierProtokoll(ctx, turnierID)
}

func (t *Turnierflaeche) TurnierKorrekturfolgen(
	ctx context.Context, turnierID, paarungID string, stand *ErgebnisDTO,
) KorrekturfolgenAntwort {
	return t.app.TurnierKorrekturfolgen(ctx, turnierID, paarungID, stand)
}

func (t *Turnierflaeche) TurnierKopieren(
	ctx context.Context, turnierID, name string, mitTeilnehmenden bool,
) TurnierAntwort {
	return t.app.TurnierKopieren(ctx, turnierID, name, mitTeilnehmenden)
}

func (t *Turnierflaeche) TurnierLoeschen(
	ctx context.Context, turnierID string,
) SchreibAntwort {
	return t.app.TurnierLoeschen(ctx, turnierID)
}

func (t *Turnierflaeche) TurnierAnmelden(
	ctx context.Context, turnierID, name, deckID string,
) TeilnehmerAntwort {
	return t.app.TurnierAnmelden(ctx, turnierID, name, deckID)
}

func (t *Turnierflaeche) TurnierAussteigen(
	ctx context.Context, turnierID, teilnehmerID string, ausgestiegen bool,
) SchreibAntwort {
	return t.app.TurnierAussteigen(ctx, turnierID, teilnehmerID, ausgestiegen)
}

func (t *Turnierflaeche) TurnierPaarungsvorschlag(
	ctx context.Context, turnierID string,
) PaarungenAntwort {
	return t.app.TurnierPaarungsvorschlag(ctx, turnierID)
}

func (t *Turnierflaeche) TurnierRundeLegen(
	ctx context.Context, turnierID string,
) RundeAntwort {
	return t.app.TurnierRundeLegen(ctx, turnierID)
}

func (t *Turnierflaeche) TurnierErgebnisSetzen(
	ctx context.Context, turnierID, fassung, paarungID string, stand *ErgebnisDTO,
) SchreibAntwort {
	return t.app.TurnierErgebnisSetzen(ctx, turnierID, fassung, paarungID, stand)
}

func (t *Turnierflaeche) TurnierRangliste(
	ctx context.Context, turnierID string,
) RanglisteAntwort {
	return t.app.TurnierRangliste(ctx, turnierID)
}

// Ausgabe, Einlesen und Bericht — DOM-04.AK-03 (#1699): als Text für MCP und
// REST, an der Platte für die Kommandozeile.

func (t *Turnierflaeche) TurnierAusgabeText(ctx context.Context, turnierID string) TurnierdateiAntwort {
	return t.app.TurnierAusgabeText(ctx, turnierID)
}

func (t *Turnierflaeche) TurnierAusgabeSchreiben(ctx context.Context, turnierID, pfad string) TurnierdateiAntwort {
	return t.app.TurnierAusgabeSchreiben(ctx, turnierID, pfad)
}

func (t *Turnierflaeche) TurnierEinlesenText(ctx context.Context, inhalt string) TurnierdateiAntwort {
	return t.app.TurnierEinlesenText(ctx, inhalt)
}

func (t *Turnierflaeche) TurnierEinlesenLesen(ctx context.Context, pfad string) TurnierdateiAntwort {
	return t.app.TurnierEinlesenLesen(ctx, pfad)
}

func (t *Turnierflaeche) TurnierBerichtText(ctx context.Context, turnierID string) TurnierdateiAntwort {
	return t.app.TurnierBerichtText(ctx, turnierID)
}

func (t *Turnierflaeche) TurnierBerichtSchreiben(ctx context.Context, turnierID, pfad string) TurnierdateiAntwort {
	return t.app.TurnierBerichtSchreiben(ctx, turnierID, pfad)
}
