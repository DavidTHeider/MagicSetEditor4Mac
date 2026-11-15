package app

// **Umbenennen, Löschen, Einsetzen und Ausgeben** — die Wege der
// Symbolbibliothek, die kein neues Symbol anlegen (DOM-02,
// #1697).
// Eine eigene Datei neben `symbolbibliothek.go`, damit keine von beiden über
// die Zeilengrenze aus §2 geht.

import (
	"context"
	"strings"
)

// BibliothekssymbolUmbenennen gibt einem Symbol einen anderen Namen.
func (a *App) BibliothekssymbolUmbenennen(ctx context.Context, id, name string) BibliothekssymbolAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("id", id)
	if befund != nil {
		return BibliothekssymbolAntwort{Fehler: befund}
	}

	if a.symbolbibliothek == nil {
		return BibliothekssymbolAntwort{Fehler: keineSymbolbibliothek()}
	}

	if err := a.symbolbibliothek.Umbenennen(ctx, kennung, name); err != nil {
		return BibliothekssymbolAntwort{Fehler: Uebersetze(ctx, err)}
	}

	// **Der Name, wie er gespeichert ist** — `PruefeSymbolname` schneidet
	// den Rand ab, und die Antwort soll nicht mehr zusagen als der Bestand.
	return BibliothekssymbolAntwort{ID: id, Name: strings.TrimSpace(name)}
}

// BibliothekssymbolLoeschen nimmt ein Symbol aus der Bibliothek und nennt die
// Sets, die es tragen. **Ohne `bestaetigt` geht ein verwendetes nicht** — die
// Absage nennt dieselben Sets (AK-02).
func (a *App) BibliothekssymbolLoeschen(ctx context.Context, id string, bestaetigt bool) SymbolfolgenAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("id", id)
	if befund != nil {
		return SymbolfolgenAntwort{Sets: []string{}, Fehler: befund}
	}

	if a.symbolbibliothek == nil {
		return SymbolfolgenAntwort{Sets: []string{}, Fehler: keineSymbolbibliothek()}
	}

	folgen, err := a.symbolbibliothek.Loeschen(ctx, kennung, bestaetigt)
	if err != nil {
		return SymbolfolgenAntwort{Sets: nichtNilText(folgen), Fehler: Uebersetze(ctx, err)}
	}

	return SymbolfolgenAntwort{Sets: nichtNilText(folgen)}
}

// BibliothekssymbolEinsetzen legt eine Kopie neben das Set und macht sie zu
// dessen Symbol.
func (a *App) BibliothekssymbolEinsetzen(ctx context.Context, id, setID string) SymbolEinsatzAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("id", id)
	if befund != nil {
		return SymbolEinsatzAntwort{Fehler: befund}
	}

	setKennung, befund := benannteKennungAus("setID", setID)
	if befund != nil {
		return SymbolEinsatzAntwort{Fehler: befund}
	}

	if a.symbolbibliothek == nil {
		return SymbolEinsatzAntwort{Fehler: keineSymbolbibliothek()}
	}

	datei, err := a.symbolbibliothek.Einsetzen(ctx, kennung, setKennung)
	if err != nil {
		return SymbolEinsatzAntwort{Fehler: Uebersetze(ctx, err)}
	}

	return SymbolEinsatzAntwort{Datei: datei}
}

// BibliothekssymbolAusgabeSchreiben schreibt ein Symbol als `.mse-symbol` an
// einen genannten Ort.
func (a *App) BibliothekssymbolAusgabeSchreiben(ctx context.Context, id, pfad string) BibliothekssymbolAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("id", id)
	if befund != nil {
		return BibliothekssymbolAntwort{Fehler: befund}
	}

	if a.symbolbibliothek == nil {
		return BibliothekssymbolAntwort{Fehler: keineSymbolbibliothek()}
	}

	if err := a.symbolbibliothek.Ausgeben(ctx, kennung, pfad); err != nil {
		return BibliothekssymbolAntwort{Fehler: Uebersetze(ctx, err)}
	}

	return BibliothekssymbolAntwort{ID: id, Pfad: pfad}
}
