package app

// **Die Symbolbibliothek am Fenster** — DOM-02
// (#1697).
//
// Sechs Wege reichen durch; Ein- und Ausgeben fangen mit einem Dialog an und
// stehen deshalb auf `Bindungen`, wie `TurnierAusgabeSpeichern`.

import (
	"context"
	"strings"
)

// Symbolbibliothek gibt die Symbole, deren Name `suche` enthält.
func (b *Bindungen) Symbolbibliothek(ctx context.Context, suche string) SymbolbibliothekAntwort {
	return b.app(ctx).Symbolbibliothek(b.ruf(ctx), suche)
}

// BibliothekssymbolAufnehmen legt das Symbol eines Sets in die Bibliothek.
func (b *Bindungen) BibliothekssymbolAufnehmen(ctx context.Context, setID, name string) BibliothekssymbolAntwort {
	return b.app(ctx).BibliothekssymbolAufnehmen(b.ruf(ctx), setID, name)
}

// BibliothekssymbolKopieren legt ein Symbol unter anderem Namen noch einmal
// an.
func (b *Bindungen) BibliothekssymbolKopieren(ctx context.Context, id, name string) BibliothekssymbolAntwort {
	return b.app(ctx).BibliothekssymbolKopieren(b.ruf(ctx), id, name)
}

// BibliothekssymbolUmbenennen gibt einem Symbol einen anderen Namen.
func (b *Bindungen) BibliothekssymbolUmbenennen(ctx context.Context, id, name string) BibliothekssymbolAntwort {
	return b.app(ctx).BibliothekssymbolUmbenennen(b.ruf(ctx), id, name)
}

// BibliothekssymbolLoeschen nimmt ein Symbol aus der Bibliothek.
func (b *Bindungen) BibliothekssymbolLoeschen(ctx context.Context, id string, bestaetigt bool) SymbolfolgenAntwort {
	return b.app(ctx).BibliothekssymbolLoeschen(b.ruf(ctx), id, bestaetigt)
}

// BibliothekssymbolEinsetzen macht ein Symbol zu dem eines Sets.
func (b *Bindungen) BibliothekssymbolEinsetzen(ctx context.Context, id, setID string) SymbolEinsatzAntwort {
	return b.app(ctx).BibliothekssymbolEinsetzen(b.ruf(ctx), id, setID)
}

// BibliothekssymbolTeile gibt die Zeichnung eines Symbols für den Editor.
func (b *Bindungen) BibliothekssymbolTeile(ctx context.Context, id string) SymbolAntwort {
	return b.app(ctx).BibliothekssymbolTeile(b.ruf(ctx), id)
}

// BibliothekssymbolSichern legt eine neue Fassung ab.
func (b *Bindungen) BibliothekssymbolSichern(ctx context.Context, id string, teile []SymbolteilDTO) SymbolaenderungAntwort {
	return b.app(ctx).BibliothekssymbolSichern(b.ruf(ctx), id, teile)
}

// BibliothekssymbolZuruecknehmen holt die vorige Fassung zurück.
func (b *Bindungen) BibliothekssymbolZuruecknehmen(ctx context.Context, id string) SymbolaenderungAntwort {
	return b.app(ctx).BibliothekssymbolZuruecknehmen(b.ruf(ctx), id)
}

// BibliothekssymbolNachziehen legt die jetzige Fassung in die Sets, die die
// vorige tragen.
func (b *Bindungen) BibliothekssymbolNachziehen(ctx context.Context, id string) SymbolfolgenAntwort {
	return b.app(ctx).BibliothekssymbolNachziehen(b.ruf(ctx), id)
}

// BibliothekssymbolFassungen zeichnet ein Symbol in den Fassungen eines Sets.
func (b *Bindungen) BibliothekssymbolFassungen(ctx context.Context, id, setID string) SymbolfassungenAntwort {
	return b.app(ctx).BibliothekssymbolFassungen(b.ruf(ctx), id, setID)
}

// BibliothekssymbolEinlesen fragt nach einer `.mse-symbol` und legt sie als
// neues Symbol an — es heißt wie die Datei.
func (b *Bindungen) BibliothekssymbolEinlesen(ctx context.Context) BibliothekssymbolAntwort {
	ctx = b.ruf(ctx)
	texte := DialogtexteFuer(b.app(ctx).sprache)

	return nachSymboldialog(ctx, b.app(ctx), Dateiauswahl{
		Titel: texte.SymbolEinlesen, Filter: texte.SymbolFilter, Muster: "*.mse-symbol",
	}, func(pfad string) BibliothekssymbolAntwort {
		return b.app(ctx).BibliothekssymbolEinlesenAus(ctx, pfad, "")
	})
}

// BibliothekssymbolAusgeben fragt nach einem Ort und schreibt das Symbol
// dorthin. `name` schlägt den Dateinamen vor.
func (b *Bindungen) BibliothekssymbolAusgeben(ctx context.Context, id, name string) BibliothekssymbolAntwort {
	ctx = b.ruf(ctx)
	texte := DialogtexteFuer(b.app(ctx).sprache)

	return nachSymboldialog(ctx, b.app(ctx), Dateiauswahl{
		Titel: texte.SymbolAusgeben, Filter: texte.SymbolFilter, Muster: "*.mse-symbol",
		Vorschlag: symbolvorschlag(name, texte.SymbolVorschlag), Speichern: true,
	}, func(pfad string) BibliothekssymbolAntwort {
		return b.app(ctx).BibliothekssymbolAusgabeSchreiben(ctx, id, pfad)
	})
}

// symbolvorschlag macht aus dem Namen eines Symbols einen Dateinamen —
// **ohne Trenner**, die ihn in einen Ordner verwandelten.
func symbolvorschlag(name, vorgabe string) string {
	bereinigt := strings.TrimSpace(strings.NewReplacer("/", "-", ":", "-", "\\", "-").Replace(name))
	if bereinigt == "" {
		return vorgabe
	}

	return bereinigt + ".mse-symbol"
}

// nachSymboldialog fragt nach der Datei und ruft den Weg — **ein
// weggeklickter Dialog ist ein Abbruch und kein Fehler**, wie bei
// `nachDialog`.
func nachSymboldialog(
	ctx context.Context, anwendung *App, auswahl Dateiauswahl, weg func(pfad string) BibliothekssymbolAntwort,
) BibliothekssymbolAntwort {
	pfad, err := anwendung.waehleDatei(ctx, auswahl)
	if err != nil {
		return BibliothekssymbolAntwort{Fehler: Uebersetze(ctx, err)}
	}

	if pfad == "" {
		return BibliothekssymbolAntwort{Abgebrochen: true}
	}

	return weg(pfad)
}
