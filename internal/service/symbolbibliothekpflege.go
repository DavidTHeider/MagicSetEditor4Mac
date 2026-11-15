package service

// **Ein Symbol der Bibliothek bearbeiten** — DOM-02.AK-02 und AK-03
// (#1697).
//
// # Zwei Fassungen, und die Sets bleiben, wie sie sind
//
// **Wer speichert, legt eine neue Fassung ab, und die bisherige wird die
// vorige** — die Spalte `vorher` trägt sie, und *Zurücknehmen* tauscht beide.
// **Die Sets behalten ihre Kopie**: ein Set, das die bisherige Fassung trägt,
// zeichnet weiter mit ihr, bis jemand es **nachzieht**. Die Antwort nennt
// diese Sets, damit das eine Entscheidung ist und kein Zufall.
//
// # Nachziehen ist ein gezieltes Neuzeichnen
//
// **Die Datei des Sets wird unter ihrem Namen neu geschrieben**; der Wert im
// Set bleibt. Ein Kartenplan trägt die Teile des Symbols selbst
// (`render.Bildbaum.Symbol`), und der Schlüssel eines gerasterten Bildes geht
// über den Plan (`Texturschluessel`) — neu gezeichnet werden damit genau die
// Karten der nachgezogenen Sets, und jede andere behält ihr Bild (AK-03).
//
// Portierungsmarke **new** — die Bibliothek gibt es in MSE2 nicht (Vermerk
// 0142).

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// Symbolaenderung ist eine neue Fassung samt den Sets, die die vorige tragen.
type Symbolaenderung struct {
	Symbol   domain.Bibliothekssymbol
	Veraltet []string
}

// Zeichnung gibt die Zeichnung eines Symbols — für den Editor.
func (b *Symbolbibliothek) Zeichnung(ctx context.Context, id int64) (msepaket.Symbolzeichnung, error) {
	symbol, err := b.speicher.Symbol(ctx, id)
	if err != nil {
		return msepaket.Symbolzeichnung{}, err
	}

	zeichnung, _, err := msepaket.SymbolzeichnungAus(symbol.Name, []byte(symbol.Inhalt))
	if err != nil {
		return msepaket.Symbolzeichnung{}, symbollesefehler(symbol.Name, err)
	}

	return zeichnung, nil
}

// Speichern legt eine neue Fassung ab — mit denselben Grenzen wie eine
// eingelesene Datei.
func (b *Symbolbibliothek) Speichern(
	ctx context.Context, id int64, zeichnung msepaket.Symbolzeichnung,
) (Symbolaenderung, error) {
	if err := msepaket.PruefeSymbolumfang(zeichnung); err != nil {
		return Symbolaenderung{}, err
	}

	symbol, err := b.speicher.Symbol(ctx, id)
	if err != nil {
		return Symbolaenderung{}, err
	}

	return b.fassungWechseln(ctx, symbol, msepaket.SchreibeSymbolzeichnung(zeichnung))
}

// Zuruecknehmen holt die vorige Fassung zurück; die jetzige wird die vorige.
func (b *Symbolbibliothek) Zuruecknehmen(ctx context.Context, id int64) (Symbolaenderung, error) {
	symbol, err := b.speicher.Symbol(ctx, id)
	if err != nil {
		return Symbolaenderung{}, err
	}

	if symbol.Vorher == "" {
		return Symbolaenderung{}, domain.NewFieldError("id", domain.SchluesselSymbolOhneVorige,
			"%s hat keine vorige Fassung.", symbol.Name)
	}

	return b.fassungWechseln(ctx, symbol, []byte(symbol.Vorher))
}

// fassungWechseln schreibt einen neuen Inhalt und nennt die Sets, die den
// bisherigen tragen.
//
// **Derselbe Inhalt noch einmal ändert nichts**: sonst würde die vorige
// Fassung mit der jetzigen überschrieben, und der Weg zurück wäre fort.
func (b *Symbolbibliothek) fassungWechseln(
	ctx context.Context, symbol domain.Bibliothekssymbol, inhalt []byte,
) (Symbolaenderung, error) {
	abdruck := symbolabdruck(inhalt)
	if abdruck == symbol.Abdruck {
		return Symbolaenderung{Symbol: symbol, Veraltet: []string{}}, nil
	}

	traeger, err := b.traegerJeAbdruck(ctx)
	if err != nil {
		return Symbolaenderung{}, err
	}

	if err := b.speicher.SymbolInhaltSetzen(ctx, symbol.ID, string(inhalt), abdruck); err != nil {
		return Symbolaenderung{}, err
	}

	symbol.Vorher, symbol.Inhalt, symbol.Abdruck = symbol.Inhalt, string(inhalt), abdruck

	return Symbolaenderung{Symbol: symbol, Veraltet: traegernamen(traeger[symbolabdruck([]byte(symbol.Vorher))])}, nil
}

// Nachziehen legt die jetzige Fassung in die Sets, die die vorige tragen, und
// nennt sie.
//
// **Jedes Set für sich**: scheitert eines, sind die davor nachgezogen, und die
// Antwort nennt sie neben dem Fehler — ein halber Stand, der gesagt wird,
// statt einer, den niemand sieht.
func (b *Symbolbibliothek) Nachziehen(ctx context.Context, id int64) ([]string, error) {
	if err := b.darfSchreiben(ctx); err != nil {
		return nil, err
	}

	symbol, err := b.speicher.Symbol(ctx, id)
	if err != nil {
		return nil, err
	}

	if symbol.Vorher == "" {
		return []string{}, nil
	}

	zeichnung, _, err := msepaket.SymbolzeichnungAus(symbol.Name, []byte(symbol.Inhalt))
	if err != nil {
		return nil, symbollesefehler(symbol.Name, err)
	}

	traeger, err := b.traegerJeAbdruck(ctx)
	if err != nil {
		return nil, err
	}

	nachgezogen := []symboltraeger{}

	for _, einer := range traeger[symbolabdruck([]byte(symbol.Vorher))] {
		if _, err := b.ablage.Ablegen(einer.setID, zeichnung, einer.datei); err != nil {
			return traegernamen(nachgezogen), err
		}

		nachgezogen = append(nachgezogen, einer)
	}

	return traegernamen(nachgezogen), nil
}
