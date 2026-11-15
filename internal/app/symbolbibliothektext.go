package app

// **Ein Symbol der Bibliothek als Text** — DOM-02
// (#1697).
//
// Für MCP und REST: **sie stehen an keiner Platte, die ihr Gegenüber sieht**,
// und reichen die `.mse-symbol` deshalb als Text hinaus und herein —
// dieselbe Abwägung wie bei der Turnierdatei (Vermerk 0140). Fenster und
// Kommandozeile nehmen Dateien.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// BibliothekssymbolTextAntwort trägt ein Symbol samt seiner Datei im Text.
type BibliothekssymbolTextAntwort struct {
	ID   string `json:"id"`
	Name string `json:"name"`

	// Inhalt ist die `.mse-symbol`, wie sie geschrieben würde.
	Inhalt string `json:"inhalt"`

	Fehler *Fehler `json:"fehler"`
}

// BibliothekssymbolText gibt ein Symbol als Text.
func (a *App) BibliothekssymbolText(ctx context.Context, id string) BibliothekssymbolTextAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("id", id)
	if befund != nil {
		return BibliothekssymbolTextAntwort{Fehler: befund}
	}

	if a.symbolbibliothek == nil {
		return BibliothekssymbolTextAntwort{Fehler: keineSymbolbibliothek()}
	}

	symbol, err := a.symbolbibliothek.Symbol(ctx, kennung)
	if err != nil {
		return BibliothekssymbolTextAntwort{Fehler: Uebersetze(ctx, err)}
	}

	return BibliothekssymbolTextAntwort{ID: id, Name: symbol.Name, Inhalt: symbol.Inhalt}
}

// BibliothekssymbolAusText legt ein Symbol aus dem Text einer `.mse-symbol`
// an — der Name ist Pflicht, denn ein Text hat keinen Dateinamen.
func (a *App) BibliothekssymbolAusText(ctx context.Context, inhalt, name string) BibliothekssymbolAntwort {
	ctx = a.imProfil(ctx)

	return a.bibliothekssymbolAngelegt(ctx, func(ctx context.Context, f Symbolbibliotheksfuehrer) (domain.Bibliothekssymbol, error) {
		return f.EinlesenText(ctx, name, []byte(inhalt))
	})
}
