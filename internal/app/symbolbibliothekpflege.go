package app

// **Ein Symbol der Bibliothek bearbeiten** — DOM-02.AK-02 und AK-03
// (#1697).
//
// Die Zeichnung für den Editor, eine neue Fassung, der Weg zurück und das
// Nachziehen der Sets, die die vorige tragen. **Was davon die Sets berührt,
// entscheidet der Dienst**; hier wird übersetzt.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// SymbolaenderungAntwort trägt eine neue Fassung samt den Sets, die die
// vorige tragen — sie behalten ihre Kopie, bis jemand sie nachzieht.
type SymbolaenderungAntwort struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Veraltet []string `json:"veraltet"`
	Fehler   *Fehler  `json:"fehler"`
}

// BibliothekssymbolTeile gibt die Zeichnung eines Symbols für den Editor.
func (a *App) BibliothekssymbolTeile(ctx context.Context, id string) SymbolAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("id", id)
	if befund != nil {
		return SymbolAntwort{Teile: []SymbolteilDTO{}, Fehler: befund}
	}

	if a.symbolbibliothek == nil {
		return SymbolAntwort{Teile: []SymbolteilDTO{}, Fehler: keineSymbolbibliothek()}
	}

	symbol, err := a.symbolbibliothek.Symbol(ctx, kennung)
	if err != nil {
		return SymbolAntwort{Teile: []SymbolteilDTO{}, Fehler: Uebersetze(ctx, err)}
	}

	zeichnung, err := a.symbolbibliothek.Zeichnung(ctx, kennung)
	if err != nil {
		return SymbolAntwort{Teile: []SymbolteilDTO{}, Fehler: Uebersetze(ctx, err)}
	}

	return SymbolAntwort{Teile: teileAlsDTO(zeichnung.Teile), Name: symbol.Name}
}

// BibliothekssymbolSichern legt eine neue Fassung ab.
func (a *App) BibliothekssymbolSichern(ctx context.Context, id string, teile []SymbolteilDTO) SymbolaenderungAntwort {
	ctx = a.imProfil(ctx)

	return a.symbolaenderung(ctx, id, func(ctx context.Context, f Symbolbibliotheksfuehrer, kennung int64) (service.Symbolaenderung, error) {
		return f.Speichern(ctx, kennung, msepaket.Symbolzeichnung{Teile: teileAusDTO(teile)})
	})
}

// BibliothekssymbolZuruecknehmen holt die vorige Fassung zurück.
func (a *App) BibliothekssymbolZuruecknehmen(ctx context.Context, id string) SymbolaenderungAntwort {
	ctx = a.imProfil(ctx)

	return a.symbolaenderung(ctx, id, func(ctx context.Context, f Symbolbibliotheksfuehrer, kennung int64) (service.Symbolaenderung, error) {
		return f.Zuruecknehmen(ctx, kennung)
	})
}

// symbolaenderung ruft einen Weg, der eine Fassung wechselt.
func (a *App) symbolaenderung(
	ctx context.Context, id string,
	weg func(context.Context, Symbolbibliotheksfuehrer, int64) (service.Symbolaenderung, error),
) SymbolaenderungAntwort {
	kennung, befund := benannteKennungAus("id", id)
	if befund != nil {
		return SymbolaenderungAntwort{Veraltet: []string{}, Fehler: befund}
	}

	if a.symbolbibliothek == nil {
		return SymbolaenderungAntwort{Veraltet: []string{}, Fehler: keineSymbolbibliothek()}
	}

	aenderung, err := weg(ctx, a.symbolbibliothek, kennung)
	if err != nil {
		return SymbolaenderungAntwort{Veraltet: []string{}, Fehler: Uebersetze(ctx, err)}
	}

	return SymbolaenderungAntwort{
		ID: id, Name: aenderung.Symbol.Name, Veraltet: nichtNilText(aenderung.Veraltet),
	}
}

// BibliothekssymbolNachziehen legt die jetzige Fassung in die Sets, die die
// vorige tragen — **und nennt sie auch, wenn es unterwegs scheitert**.
func (a *App) BibliothekssymbolNachziehen(ctx context.Context, id string) SymbolfolgenAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("id", id)
	if befund != nil {
		return SymbolfolgenAntwort{Sets: []string{}, Fehler: befund}
	}

	if a.symbolbibliothek == nil {
		return SymbolfolgenAntwort{Sets: []string{}, Fehler: keineSymbolbibliothek()}
	}

	nachgezogen, err := a.symbolbibliothek.Nachziehen(ctx, kennung)
	if err != nil {
		return SymbolfolgenAntwort{Sets: nichtNilText(nachgezogen), Fehler: Uebersetze(ctx, err)}
	}

	return SymbolfolgenAntwort{Sets: nichtNilText(nachgezogen)}
}
