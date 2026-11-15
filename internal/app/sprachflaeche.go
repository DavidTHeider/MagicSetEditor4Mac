package app

// **Die eigenen Übersetzungen für die drei Flächen ohne Fenster** — LOC-02
// (#1688).
//
// Dieselbe Form wie `symbolbibliotheksflaeche.go`: die Fläche **ist** die
// Anwendung, damit Kommandozeile, MCP und REST dieselben Methoden rufen wie
// das Fenster (§16). Die Dialoge fehlen — wer ohne Fenster aus- oder
// eingibt, nennt den Pfad oder reicht den Text.

import (
	"context"
	"database/sql"
)

// Sprachflaeche ist die Anwendung selbst, für eine Fläche ohne Fenster.
type Sprachflaeche struct{ app *App }

// SprachflaecheUeber baut sie über einem Bestand — mit dem Riegel vor dem
// Schreiben einer Datei, wie im Fenster.
func SprachflaecheUeber(db *sql.DB) *Sprachflaeche {
	anwendung := NeueApp("")
	anwendung.MitSprachpaketen(NeueSprachpaketeUeber(db).MitPruefer(AutorisierungAus(db)))

	return &Sprachflaeche{app: anwendung}
}

func (f *Sprachflaeche) Sprachpakete(ctx context.Context) SprachpaketlisteAntwort {
	return f.app.Sprachpakete(ctx)
}

func (f *Sprachflaeche) Sprachpaket(ctx context.Context, kennung string) SprachpaketAntwort {
	return f.app.Sprachpaket(ctx, kennung)
}

func (f *Sprachflaeche) SprachpaketAnlegen(ctx context.Context, kennung, name, basis string) SprachpaketAntwort {
	return f.app.SprachpaketAnlegen(ctx, kennung, name, basis)
}

func (f *Sprachflaeche) SprachpaketAendern(
	ctx context.Context, kennung string, angaben SprachpaketangabenDTO, texte map[string]string,
) SprachpaketAntwort {
	return f.app.SprachpaketAendern(ctx, kennung, angaben, texte)
}

func (f *Sprachflaeche) SprachpaketLoeschen(ctx context.Context, kennung string) SchreibAntwort {
	return f.app.SprachpaketLoeschen(ctx, kennung)
}

func (f *Sprachflaeche) SprachpaketAusgabeSchreiben(ctx context.Context, kennung, pfad string) SprachpaketAntwort {
	return f.app.SprachpaketAusgabeSchreiben(ctx, kennung, pfad)
}

func (f *Sprachflaeche) SprachpaketAusgabeText(ctx context.Context, kennung string) SprachpaketdateiAntwort {
	return f.app.SprachpaketAusgabeText(ctx, kennung)
}

func (f *Sprachflaeche) SprachpaketEinlesenLesen(ctx context.Context, pfad string, ersetzen bool) SprachpaketAntwort {
	return f.app.SprachpaketEinlesenLesen(ctx, pfad, ersetzen)
}

func (f *Sprachflaeche) SprachpaketEinlesenText(ctx context.Context, inhalt string, ersetzen bool) SprachpaketAntwort {
	return f.app.SprachpaketEinlesenText(ctx, inhalt, ersetzen)
}

// Sprachwerkstatt ist dieselbe Fläche in der Gestalt, die MCP und REST
// erwarten: eine Antwort oder ein Fehler.
type Sprachwerkstatt struct{ flaeche *Sprachflaeche }

// NeueSprachwerkstatt baut sie über einem Bestand.
func NeueSprachwerkstatt(db *sql.DB) *Sprachwerkstatt {
	return &Sprachwerkstatt{flaeche: SprachflaecheUeber(db)}
}

func (w *Sprachwerkstatt) Sprachpakete(ctx context.Context) (any, error) {
	antwort := w.flaeche.Sprachpakete(ctx)

	return ausMCP(antwort.Pakete, antwort.Fehler)
}

func (w *Sprachwerkstatt) Sprachpaket(ctx context.Context, kennung string) (any, error) {
	return sprachErgebnis(w.flaeche.Sprachpaket(ctx, kennung))
}

func (w *Sprachwerkstatt) SprachpaketAnlegen(ctx context.Context, kennung, name, basis string) (any, error) {
	return sprachErgebnis(w.flaeche.SprachpaketAnlegen(ctx, kennung, name, basis))
}

// SprachpaketAendern nimmt die Angaben als Karte — **was fehlt, bleibt**.
func (w *Sprachwerkstatt) SprachpaketAendern(
	ctx context.Context, kennung string, angaben, texte map[string]string,
) (any, error) {
	gesetzt := func(schluessel string) *string {
		if wert, da := angaben[schluessel]; da {
			return &wert
		}

		return nil
	}

	return sprachErgebnis(w.flaeche.SprachpaketAendern(ctx, kennung, SprachpaketangabenDTO{
		Name: gesetzt("name"), Stand: gesetzt("stand"), Lizenz: gesetzt("lizenz"),
		Herkunft: gesetzt("herkunft"),
	}, texte))
}

func (w *Sprachwerkstatt) SprachpaketLoeschen(ctx context.Context, kennung string) (any, error) {
	antwort := w.flaeche.SprachpaketLoeschen(ctx, kennung)

	return ausMCP(map[string]string{"geloescht": kennung}, antwort.Fehler)
}

func (w *Sprachwerkstatt) SprachpaketAusgabe(ctx context.Context, kennung string) (any, error) {
	antwort := w.flaeche.SprachpaketAusgabeText(ctx, kennung)

	return ausMCP(antwort, antwort.Fehler)
}

func (w *Sprachwerkstatt) SprachpaketEinlesen(ctx context.Context, inhalt string, ersetzen bool) (any, error) {
	return sprachErgebnis(w.flaeche.SprachpaketEinlesenText(ctx, inhalt, ersetzen))
}

// sprachErgebnis reicht eine Antwort als Ergebnis oder Fehler weiter.
func sprachErgebnis(antwort SprachpaketAntwort) (any, error) {
	return ausMCP(antwort, antwort.Fehler)
}
