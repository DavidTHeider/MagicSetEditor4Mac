package app

// **Die Symbolbibliothek an der Fassade** — DOM-02
// (#1697).
//
// Die Bibliothek selbst steht in `internal/service`; hier wird **übersetzt**:
// Kennungen aus Text, Zeitpunkte als Text, die Zeichnung als Vorschaubild.
// Was ein verwendetes Symbol ist und wann es nicht gelöscht wird, weiß diese
// Schicht nicht (§2).

import (
	"context"
	"fmt"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Symbolbibliotheksfuehrer ist die Naht zur Bibliothek.
type Symbolbibliotheksfuehrer interface {
	Liste(ctx context.Context, suche string) ([]service.Symboleintrag, error)
	Aufnehmen(ctx context.Context, setID int64, name string) (domain.Bibliothekssymbol, error)
	Kopieren(ctx context.Context, id int64, name string) (domain.Bibliothekssymbol, error)
	Umbenennen(ctx context.Context, id int64, name string) error
	Loeschen(ctx context.Context, id int64, bestaetigt bool) ([]string, error)
	Einsetzen(ctx context.Context, id, setID int64) (string, error)
	Ausgeben(ctx context.Context, id int64, ziel string) error
	Einlesen(ctx context.Context, pfad, name string) (domain.Bibliothekssymbol, error)
	Symbol(ctx context.Context, id int64) (domain.Bibliothekssymbol, error)
	EinlesenText(ctx context.Context, name string, roh []byte) (domain.Bibliothekssymbol, error)

	// Die Bearbeitung — AK-02 und AK-03.
	Zeichnung(ctx context.Context, id int64) (msepaket.Symbolzeichnung, error)
	Speichern(ctx context.Context, id int64, zeichnung msepaket.Symbolzeichnung) (service.Symbolaenderung, error)
	Zuruecknehmen(ctx context.Context, id int64) (service.Symbolaenderung, error)
	Nachziehen(ctx context.Context, id int64) ([]string, error)
	Fassungsplaene(ctx context.Context, id, setID int64, groesse float64) ([]service.Fassungsplan, error)
}

// BibliothekssymbolDTO ist ein Symbol der Bibliothek, wie die Oberfläche es
// zeigt.
type BibliothekssymbolDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`

	// Abdruck ist die Prüfsumme der Zeichnung — zwei gleiche Abdrücke sind
	// dasselbe Symbol.
	Abdruck string `json:"abdruck"`

	// Angelegt und Geaendert stehen nach RFC 3339 — leer, wo der Bestand
	// keinen lesbaren Zeitpunkt hält.
	Angelegt  string `json:"angelegt"`
	Geaendert string `json:"geaendert"`

	// Vorschau ist ein PNG als Data-URI — leer, wenn es sich nicht packen
	// ließ.
	Vorschau string `json:"vorschau"`

	// Verwendet nennt die Sets, die dieses Symbol tragen.
	Verwendet []string `json:"verwendet"`

	// Vorige sagt, dass es eine vorige Fassung gibt, die sich zurückholen
	// lässt (AK-02).
	Vorige bool `json:"vorige"`
}

// SymbolbibliothekAntwort trägt die Symbole oder einen Fehler.
type SymbolbibliothekAntwort struct {
	Symbole []BibliothekssymbolDTO `json:"symbole"`
	Fehler  *Fehler                `json:"fehler"`
}

// BibliothekssymbolAntwort trägt ein angelegtes, umbenanntes, ein- oder
// ausgegebenes Symbol.
type BibliothekssymbolAntwort struct {
	ID   string `json:"id"`
	Name string `json:"name"`

	// Pfad ist die Datei, aus der gelesen oder in die geschrieben wurde —
	// leer, wo keine im Spiel ist.
	Pfad string `json:"pfad"`

	// Abgebrochen sagt, dass niemand eine Datei gewählt hat — kein Fehler.
	Abgebrochen bool `json:"abgebrochen"`

	Fehler *Fehler `json:"fehler"`
}

// SymbolfolgenAntwort nennt die Sets, die ein gelöschtes — oder wegen ihnen
// nicht gelöschtes — Symbol tragen.
type SymbolfolgenAntwort struct {
	Sets   []string `json:"sets"`
	Fehler *Fehler  `json:"fehler"`
}

// SymbolEinsatzAntwort nennt die Datei, die neben dem Set liegt.
type SymbolEinsatzAntwort struct {
	Datei  string  `json:"datei"`
	Fehler *Fehler `json:"fehler"`
}

// Bibliotheksvorschau ist die Kantenlänge einer Vorschau in Punkten —
// **doppelt so groß wie gezeigt**, für Bildschirme mit doppelter Dichte.
const Bibliotheksvorschau = 96

// MitSymbolbibliothek setzt die Naht.
func (a *App) MitSymbolbibliothek(fuehrer Symbolbibliotheksfuehrer) { a.symbolbibliothek = fuehrer }

// keineSymbolbibliothek meldet die fehlende Naht.
func keineSymbolbibliothek() *Fehler {
	return Nutzerfehler(fmt.Errorf(
		"%w: die Symbolbibliothek ist nicht eingerichtet", domain.ErrValidation))
}

// Symbolbibliothek gibt die Symbole, deren Name `suche` enthält.
func (a *App) Symbolbibliothek(ctx context.Context, suche string) SymbolbibliothekAntwort {
	ctx = a.imProfil(ctx)

	if a.symbolbibliothek == nil {
		return SymbolbibliothekAntwort{Symbole: []BibliothekssymbolDTO{}, Fehler: keineSymbolbibliothek()}
	}

	eintraege, err := a.symbolbibliothek.Liste(ctx, suche)
	if err != nil {
		return SymbolbibliothekAntwort{Symbole: []BibliothekssymbolDTO{}, Fehler: Uebersetze(ctx, err)}
	}

	dtos := make([]BibliothekssymbolDTO, 0, len(eintraege))
	for _, eintrag := range eintraege {
		dtos = append(dtos, a.bibliothekssymbolAlsDTO(eintrag))
	}

	return SymbolbibliothekAntwort{Symbole: dtos}
}

// bibliothekssymbolAlsDTO übersetzt einen Eintrag.
//
// **Eine Vorschau, die sich nicht packen lässt, ist leer** und kein Fehler
// der Liste — dieselbe Abwägung wie bei der Symbolpalette (§8).
func (a *App) bibliothekssymbolAlsDTO(eintrag service.Symboleintrag) BibliothekssymbolDTO {
	vorschau, _ := a.alsDatenURI(service.Symbolvorschau(eintrag.Zeichnung, Bibliotheksvorschau)) //nolint:errcheck // siehe oben

	return BibliothekssymbolDTO{
		ID:        kennungText(eintrag.Symbol.ID),
		Name:      eintrag.Symbol.Name,
		Abdruck:   eintrag.Symbol.Abdruck,
		Angelegt:  zeitpunktText(eintrag.Symbol.Angelegt),
		Geaendert: zeitpunktText(eintrag.Symbol.Geaendert),
		Vorschau:  vorschau,
		Verwendet: nichtNilText(eintrag.Verwendet),
		Vorige:    eintrag.Symbol.Vorher != "",
	}
}

// zeitpunktText schreibt einen Zeitpunkt nach RFC 3339 — leer für den
// Nullwert, der „unbekannt" heißt.
func zeitpunktText(zeitpunkt time.Time) string {
	if zeitpunkt.IsZero() {
		return ""
	}

	return zeitpunkt.UTC().Format(time.RFC3339)
}

// nichtNilText gibt eine leere Liste statt `nil` — über die Brücke wird
// daraus sonst `null` (`leerelisten_test.go`).
func nichtNilText(texte []string) []string {
	if texte == nil {
		return []string{}
	}

	return texte
}

// BibliothekssymbolAufnehmen legt das Symbol eines Sets in die Bibliothek.
func (a *App) BibliothekssymbolAufnehmen(ctx context.Context, setID, name string) BibliothekssymbolAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("setID", setID)
	if befund != nil {
		return BibliothekssymbolAntwort{Fehler: befund}
	}

	return a.bibliothekssymbolAngelegt(ctx, func(ctx context.Context, f Symbolbibliotheksfuehrer) (domain.Bibliothekssymbol, error) {
		return f.Aufnehmen(ctx, kennung, name)
	})
}

// BibliothekssymbolKopieren legt ein Symbol unter anderem Namen noch einmal
// an.
func (a *App) BibliothekssymbolKopieren(ctx context.Context, id, name string) BibliothekssymbolAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("id", id)
	if befund != nil {
		return BibliothekssymbolAntwort{Fehler: befund}
	}

	return a.bibliothekssymbolAngelegt(ctx, func(ctx context.Context, f Symbolbibliotheksfuehrer) (domain.Bibliothekssymbol, error) {
		return f.Kopieren(ctx, kennung, name)
	})
}

// BibliothekssymbolEinlesenAus liest eine `.mse-symbol` als neues Symbol —
// ohne Namen heißt es wie die Datei.
func (a *App) BibliothekssymbolEinlesenAus(ctx context.Context, pfad, name string) BibliothekssymbolAntwort {
	ctx = a.imProfil(ctx)

	antwort := a.bibliothekssymbolAngelegt(ctx, func(ctx context.Context, f Symbolbibliotheksfuehrer) (domain.Bibliothekssymbol, error) {
		return f.Einlesen(ctx, pfad, name)
	})

	if antwort.Fehler == nil {
		antwort.Pfad = pfad
	}

	return antwort
}

// bibliothekssymbolAngelegt ruft einen Weg, der ein Symbol anlegt.
func (a *App) bibliothekssymbolAngelegt(
	ctx context.Context, weg func(context.Context, Symbolbibliotheksfuehrer) (domain.Bibliothekssymbol, error),
) BibliothekssymbolAntwort {
	if a.symbolbibliothek == nil {
		return BibliothekssymbolAntwort{Fehler: keineSymbolbibliothek()}
	}

	symbol, err := weg(ctx, a.symbolbibliothek)
	if err != nil {
		return BibliothekssymbolAntwort{Fehler: Uebersetze(ctx, err)}
	}

	return BibliothekssymbolAntwort{ID: kennungText(symbol.ID), Name: symbol.Name}
}
