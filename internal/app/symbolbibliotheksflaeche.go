package app

// **Die Symbolbibliothek für die drei Flächen ohne Fenster** — DOM-02
// (#1697).
//
// Dieselbe Form wie `turnierflaeche.go`: die Fläche **ist** die Anwendung,
// damit MCP, Kommandozeile und REST dieselben Methoden rufen wie das Fenster
// (§16). Die Dialoge fehlen hier — wer ohne Fenster ein- oder ausgibt,
// nennt den Pfad.

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// Symbolbibliotheksflaeche ist die Anwendung selbst, für eine Fläche ohne
// Fenster.
type Symbolbibliotheksflaeche struct{ app *App }

// SymbolbibliotheksflaecheUeber baut sie über einem Bestand und dem
// Bilderverzeichnis — **dem Ort, an dem die eingesetzten Kopien liegen**, wie
// bei der Symbolwerkstatt. Geschrieben wird als `urheber`.
func SymbolbibliotheksflaecheUeber(db *sql.DB, urheber domain.Urheber) (*Symbolbibliotheksflaeche, error) {
	bilder, fehler := Bilderverzeichnis()
	if fehler != nil {
		return nil, fehler
	}

	leser := storage.NeuerSetLeser(db)
	anwendung := NeueApp("")
	anwendung.MitSymbolbibliothek(service.NeueSymbolbibliothek(storage.NeueSymbolbibliothek(db),
		service.NeueSymbolablage(service.NeueBildablage(bilder)),
		setwertnaht{leser: leser, schreiber: storage.NeuerKartenschreiber(db).Als(urheber)}, leser).
		MitPruefer(AutorisierungAus(db)))

	return &Symbolbibliotheksflaeche{app: anwendung}, nil
}

func (f *Symbolbibliotheksflaeche) Symbolbibliothek(ctx context.Context, suche string) SymbolbibliothekAntwort {
	return f.app.Symbolbibliothek(ctx, suche)
}

func (f *Symbolbibliotheksflaeche) BibliothekssymbolAufnehmen(
	ctx context.Context, setID, name string,
) BibliothekssymbolAntwort {
	return f.app.BibliothekssymbolAufnehmen(ctx, setID, name)
}

func (f *Symbolbibliotheksflaeche) BibliothekssymbolKopieren(
	ctx context.Context, id, name string,
) BibliothekssymbolAntwort {
	return f.app.BibliothekssymbolKopieren(ctx, id, name)
}

func (f *Symbolbibliotheksflaeche) BibliothekssymbolUmbenennen(
	ctx context.Context, id, name string,
) BibliothekssymbolAntwort {
	return f.app.BibliothekssymbolUmbenennen(ctx, id, name)
}

func (f *Symbolbibliotheksflaeche) BibliothekssymbolLoeschen(
	ctx context.Context, id string, bestaetigt bool,
) SymbolfolgenAntwort {
	return f.app.BibliothekssymbolLoeschen(ctx, id, bestaetigt)
}

func (f *Symbolbibliotheksflaeche) BibliothekssymbolEinsetzen(
	ctx context.Context, id, setID string,
) SymbolEinsatzAntwort {
	return f.app.BibliothekssymbolEinsetzen(ctx, id, setID)
}

func (f *Symbolbibliotheksflaeche) BibliothekssymbolEinlesenAus(
	ctx context.Context, pfad, name string,
) BibliothekssymbolAntwort {
	return f.app.BibliothekssymbolEinlesenAus(ctx, pfad, name)
}

func (f *Symbolbibliotheksflaeche) BibliothekssymbolAusgabeSchreiben(
	ctx context.Context, id, pfad string,
) BibliothekssymbolAntwort {
	return f.app.BibliothekssymbolAusgabeSchreiben(ctx, id, pfad)
}

func (f *Symbolbibliotheksflaeche) BibliothekssymbolText(ctx context.Context, id string) BibliothekssymbolTextAntwort {
	return f.app.BibliothekssymbolText(ctx, id)
}

func (f *Symbolbibliotheksflaeche) BibliothekssymbolAusText(
	ctx context.Context, inhalt, name string,
) BibliothekssymbolAntwort {
	return f.app.BibliothekssymbolAusText(ctx, inhalt, name)
}

func (f *Symbolbibliotheksflaeche) BibliothekssymbolTeile(ctx context.Context, id string) SymbolAntwort {
	return f.app.BibliothekssymbolTeile(ctx, id)
}

func (f *Symbolbibliotheksflaeche) BibliothekssymbolSichern(
	ctx context.Context, id string, teile []SymbolteilDTO,
) SymbolaenderungAntwort {
	return f.app.BibliothekssymbolSichern(ctx, id, teile)
}

func (f *Symbolbibliotheksflaeche) BibliothekssymbolZuruecknehmen(ctx context.Context, id string) SymbolaenderungAntwort {
	return f.app.BibliothekssymbolZuruecknehmen(ctx, id)
}

func (f *Symbolbibliotheksflaeche) BibliothekssymbolNachziehen(ctx context.Context, id string) SymbolfolgenAntwort {
	return f.app.BibliothekssymbolNachziehen(ctx, id)
}
