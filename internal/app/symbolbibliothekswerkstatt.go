package app

// **Die Symbolbibliothek für MCP und REST** — DOM-02
// (#1697).
//
// **Wie die Turnierwerkstatt und aus demselben Grund**: `mcp.Symbolbibliothek`
// und der REST-Durchreicher sind gleichlautend, und zwei Abschriften desselben
// Durchreichers liefen bei der ersten Änderung auseinander.
//
// # Was hier anders ist als am Fenster
//
// **Eine Datei ist ein Text** — beide Flächen stehen an keiner Platte, die
// ihr Gegenüber sieht (Vermerk 0140). **Und die Liste trägt kein
// Vorschaubild**: ein Modell liest kein PNG, und wer die Zeichnung braucht,
// holt sie als Text.

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Symbolbibliothekswerkstatt reicht die Antworten der Anwendung durch.
type Symbolbibliothekswerkstatt struct {
	db      *sql.DB
	urheber domain.Urheber
}

// NeueSymbolbibliothekswerkstatt baut sie über einem Bestand; geschrieben
// wird als `urheber`.
func NeueSymbolbibliothekswerkstatt(db *sql.DB, urheber domain.Urheber) *Symbolbibliothekswerkstatt {
	return &Symbolbibliothekswerkstatt{db: db, urheber: urheber}
}

// mit baut die Fläche — je Ruf einmal — und reicht ihre Antwort weiter.
//
// **Mit Fehlerergebnis**, anders als bei der Turnierwerkstatt: ohne Zuhause
// gibt es kein Bilderverzeichnis, und eine Kopie ohne Ort legte sich ins
// Arbeitsverzeichnis.
func (w *Symbolbibliothekswerkstatt) mit(weg func(*Symbolbibliotheksflaeche) (any, *Fehler)) (any, error) {
	flaeche, err := SymbolbibliotheksflaecheUeber(w.db, w.urheber)
	if err != nil {
		return nil, err
	}

	return ausMCP(weg(flaeche))
}

func (w *Symbolbibliothekswerkstatt) Symbolbibliothek(ctx context.Context, suche string) (any, error) {
	return w.mit(func(f *Symbolbibliotheksflaeche) (any, *Fehler) {
		antwort := f.Symbolbibliothek(ctx, suche)
		for stelle := range antwort.Symbole {
			antwort.Symbole[stelle].Vorschau = ""
		}

		return antwort.Symbole, antwort.Fehler
	})
}

func (w *Symbolbibliothekswerkstatt) BibliothekssymbolText(ctx context.Context, id int64) (any, error) {
	return w.mit(func(f *Symbolbibliotheksflaeche) (any, *Fehler) {
		antwort := f.BibliothekssymbolText(ctx, zahl64(id))

		return antwort, antwort.Fehler
	})
}

func (w *Symbolbibliothekswerkstatt) BibliothekssymbolAufnehmen(
	ctx context.Context, setID int64, name string,
) (any, error) {
	return w.mit(func(f *Symbolbibliotheksflaeche) (any, *Fehler) {
		antwort := f.BibliothekssymbolAufnehmen(ctx, zahl64(setID), name)

		return antwort, antwort.Fehler
	})
}

func (w *Symbolbibliothekswerkstatt) BibliothekssymbolKopieren(
	ctx context.Context, id int64, name string,
) (any, error) {
	return w.mit(func(f *Symbolbibliotheksflaeche) (any, *Fehler) {
		antwort := f.BibliothekssymbolKopieren(ctx, zahl64(id), name)

		return antwort, antwort.Fehler
	})
}

func (w *Symbolbibliothekswerkstatt) BibliothekssymbolUmbenennen(
	ctx context.Context, id int64, name string,
) (any, error) {
	return w.mit(func(f *Symbolbibliotheksflaeche) (any, *Fehler) {
		antwort := f.BibliothekssymbolUmbenennen(ctx, zahl64(id), name)

		return antwort, antwort.Fehler
	})
}

func (w *Symbolbibliothekswerkstatt) BibliothekssymbolLoeschen(
	ctx context.Context, id int64, bestaetigt bool,
) (any, error) {
	return w.mit(func(f *Symbolbibliotheksflaeche) (any, *Fehler) {
		antwort := f.BibliothekssymbolLoeschen(ctx, zahl64(id), bestaetigt)

		return antwort, antwort.Fehler
	})
}

func (w *Symbolbibliothekswerkstatt) BibliothekssymbolEinsetzen(
	ctx context.Context, id, setID int64,
) (any, error) {
	return w.mit(func(f *Symbolbibliotheksflaeche) (any, *Fehler) {
		antwort := f.BibliothekssymbolEinsetzen(ctx, zahl64(id), zahl64(setID))

		return antwort, antwort.Fehler
	})
}

func (w *Symbolbibliothekswerkstatt) BibliothekssymbolAusText(
	ctx context.Context, inhalt, name string,
) (any, error) {
	return w.mit(func(f *Symbolbibliotheksflaeche) (any, *Fehler) {
		antwort := f.BibliothekssymbolAusText(ctx, inhalt, name)

		return antwort, antwort.Fehler
	})
}

// BibliothekssymbolTeile gibt die Zeichnung für einen Editor — dieselbe
// Gestalt, die die Rechnungen des Symboleditors nehmen.
func (w *Symbolbibliothekswerkstatt) BibliothekssymbolTeile(ctx context.Context, id int64) (any, error) {
	return w.mit(func(f *Symbolbibliotheksflaeche) (any, *Fehler) {
		antwort := f.BibliothekssymbolTeile(ctx, zahl64(id))

		return antwort, antwort.Fehler
	})
}

// BibliothekssymbolSichern legt eine neue Fassung ab. **Die Teile kommen als
// rohes JSON**: MCP und der Dienst kennen `SymbolteilDTO` nicht, gelesen wird
// hier — wie bei der Symbolwerkstatt.
func (w *Symbolbibliothekswerkstatt) BibliothekssymbolSichern(
	ctx context.Context, id int64, teile json.RawMessage,
) (any, error) {
	var gelesen []SymbolteilDTO
	if err := json.Unmarshal(teile, &gelesen); err != nil {
		return nil, SymbolteileUnlesbar(err)
	}

	return w.mit(func(f *Symbolbibliotheksflaeche) (any, *Fehler) {
		antwort := f.BibliothekssymbolSichern(ctx, zahl64(id), gelesen)

		return antwort, antwort.Fehler
	})
}

func (w *Symbolbibliothekswerkstatt) BibliothekssymbolZuruecknehmen(ctx context.Context, id int64) (any, error) {
	return w.mit(func(f *Symbolbibliotheksflaeche) (any, *Fehler) {
		antwort := f.BibliothekssymbolZuruecknehmen(ctx, zahl64(id))

		return antwort, antwort.Fehler
	})
}

func (w *Symbolbibliothekswerkstatt) BibliothekssymbolNachziehen(ctx context.Context, id int64) (any, error) {
	return w.mit(func(f *Symbolbibliotheksflaeche) (any, *Fehler) {
		antwort := f.BibliothekssymbolNachziehen(ctx, zahl64(id))

		return antwort, antwort.Fehler
	})
}
