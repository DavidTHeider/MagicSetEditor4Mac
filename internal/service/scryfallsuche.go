package service

// **Die Kartensuche bei Scryfall** — Werkstück 1 von Phase 3.6
// (#1167).
//
// # Der Befund
//
// `scryfall.Klient.Suche` gibt es seit Phase V, mit Tests und voller
// Abdeckung — und **keinem Aufrufer**. Zehn Bindungen sprechen mit Scryfall,
// und keine davon sucht: man tippt einen Kartennamen und bekommt **einen**
// Vorschlag, keine Trefferliste. Wer nicht weiß, wie die Karte genau heißt,
// kommt nicht weiter.
//
// **Ein Paket ohne Aufrufer ist keine Anwendungsfunktion** (§2, und dort steht
// es, weil es schon einmal gefehlt hat).
//
// # Was diese Auskunft zurückgibt und was nicht
//
// **Was man braucht, um zu wählen**: Name, Set, Nummer, Seltenheit,
// Manakosten, Typzeile und ein kleines Bild. **Nicht** die ganze Karte — wer
// eine übernimmt, holt sie über den Abgleich, und dort steht auch, was sich
// ändern würde (§16, Regel 6).
//
// # Nachgelesen, nicht erinnert (§16)
//
// Die Suchsyntax ist Scryfalls eigene: `/cards/search?q=…` nimmt dieselbe
// Sprache wie das Suchfeld der Website — `t:creature c:r cmc<=3`. Sie hier
// nachzubauen wäre eine zweite Sprache, die niemand kennt und die bei jeder
// Erweiterung von Scryfall auseinanderliefe. **Die Frage geht deshalb
// unverändert hinaus**, und die Oberfläche erklärt sie.

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
)

// Fremdsuche ist, was die Suche von einem fremden Dienst braucht.
//
// **Eine eigene Schnittstelle neben [Fremdauskuenfte]**, obwohl derselbe
// Klient beide erfüllt: sie sagt, was **gebraucht** wird (§2). Wer sucht,
// braucht keine Symbologie.
type Fremdsuche interface {
	Suche(ctx context.Context, frage string) (
		treffer scryfall.Treffer, alter time.Duration,
		befunde []domain.Befund, fehler error)
}

// Suchtreffer ist eine gefundene Karte, so weit man sie zum Wählen braucht.
type Suchtreffer struct {
	// Kennung ist die Scryfall-Kennung des **Drucks**.
	//
	// Sie ist der Bezug, an dem alles Weitere hängt — die Preise, die
	// Sammlung, die Übernahme.
	Kennung string

	Name       string
	Setcode    string
	Setname    string
	Nummer     string
	Seltenheit string
	Manakosten string
	Typzeile   string

	// Bild ist die Adresse eines kleinen Kartenbilds — leer, wenn es keines
	// gibt.
	//
	// **Eine Adresse und keine Bytes.** Ein Suchergebnis mit zwanzig
	// eingebetteten Bildern wäre ein Megabyte JSON; wer eines sehen will,
	// holt es über den vorhandenen Bildweg.
	Bild string
}

// Suchauskunft ist das Ergebnis einer Suche samt ihrer Herkunft.
type Suchauskunft struct {
	Treffer []Suchtreffer

	// Weitere sagt, ob Scryfall mehr hat, als hier steht.
	//
	// **Es wird nicht nachgeladen** — der Grund steht bei [scryfall.Treffer]:
	// eine Suche, die von selbst durch alle Seiten liefe, stellte ohne Zutun
	// beliebig viele Anfragen.
	Weitere bool

	Befunde []domain.Befund
	Alter   time.Duration
}

// NeueSuche baut die Suche über einer Fremdquelle.
func NeueSuche(fremd Fremdsuche) *Kartensuche { return &Kartensuche{fremd: fremd} }

// Kartensuche fragt Scryfall nach Karten.
type Kartensuche struct{ fremd Fremdsuche }

// Suche gibt die Treffer zu einer Frage.
//
// **Eine leere Frage ist keine Suche.** Scryfall antwortete darauf mit einem
// Fehler; ihn erst hinausgehen zu lassen wäre eine Anfrage, deren Antwort
// feststeht.
func (k *Kartensuche) Suche(ctx context.Context, frage string) (*Suchauskunft, error) {
	if strings.TrimSpace(frage) == "" {
		return nil, fmt.Errorf(
			"%w: ohne Frage gibt es nichts zu suchen", domain.ErrValidation)
	}

	treffer, alter, befunde, fehler := k.fremd.Suche(ctx, frage)
	if fehler != nil {
		return nil, fehler
	}

	gefunden := make([]Suchtreffer, 0, len(treffer.Karten))
	for _, karte := range treffer.Karten {
		gefunden = append(gefunden, alsSuchtreffer(karte))
	}

	return &Suchauskunft{
		Treffer: gefunden,
		Weitere: treffer.Weitere,
		Befunde: befunde,
		Alter:   alter,
	}, nil
}

// alsSuchtreffer nimmt aus einer Karte, was zum Wählen reicht.
//
// **Die Zuordnung steht an einer Stelle** (§16): der fremde Verbund bleibt in
// `internal/scryfall`, und was davon in die Anwendung geht, entscheidet diese
// Funktion. Wer den fremden Typ weiterreichte, hätte die Gestalt eines fremden
// Anbieters in seiner Fachlogik.
func alsSuchtreffer(karte *scryfall.Karte) Suchtreffer {
	return Suchtreffer{
		Kennung:    karte.Kennung,
		Name:       karte.Name,
		Setcode:    karte.Set,
		Setname:    karte.Setname,
		Nummer:     karte.Nummer,
		Seltenheit: karte.Seltenheit,
		Manakosten: karte.Manakosten,
		Typzeile:   karte.Typzeile,
		Bild:       kleinesBild(karte),
	}
}

// kleinesBild gibt die Adresse eines kleinen Kartenbilds.
//
// **Und die Vorderseite einer zweiseitigen Karte**: `image_uris` steht bei
// einer DFC nicht auf der obersten Ebene, sondern an jeder Seite. Wer nur oben
// nachsähe, ließe jede zweiseitige Karte ohne Bild — und das sind bei
// modernen Sätzen viele.
func kleinesBild(karte *scryfall.Karte) string {
	if karte.Bilder.Klein != "" {
		return karte.Bilder.Klein
	}

	if len(karte.Seiten) > 0 {
		return karte.Seiten[0].Bilder.Klein
	}

	return ""
}
