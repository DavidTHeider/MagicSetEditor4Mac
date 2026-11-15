package service

// **Ein Set lesen, bevor es den Bestand erreicht** — die Strecke, die das
// Einfahren und seine Vorschau teilen (FMT-03.AK-01,
// #1692).
//
// # Warum eine Strecke und nicht zwei
//
// **Weil eine Vorschau, die anders liest als das Einfahren, lügt**: sie
// sagte „geht", und das Einfahren danach „geht nicht". Deshalb liest sie
// dieselbe Datei auf demselben Weg, mit derselben Prüfung und denselben
// Meldungen — und hört vor dem Ablegen auf.

import (
	"context"
	"errors"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// gelesenesSet ist ein Set aus einer Datei, bevor es den Bestand erreicht.
type gelesenesSet struct {
	set     *domain.Set
	roh     []domain.RawBlock
	befunde []domain.Befund
}

// setLesen öffnet ein Setpaket, deutet es und prüft, was ohne Spielpaket
// prüfbar ist.
func setLesen(pfad string) (gelesen *gelesenesSet, fehler error) {
	paket, err := msepaket.Oeffne(pfad)
	if err != nil {
		if eingeordnet := unlesbar(pfad, err); eingeordnet != nil {
			return nil, eingeordnet
		}

		if eingeordnet := unzugaenglich(pfad, err); eingeordnet != nil {
			return nil, eingeordnet
		}

		return nil, fmt.Errorf("Set öffnen: %w", err)
	}

	// Der Schließfehler wird angehängt, nicht verworfen: ein Archiv, das sich
	// nicht schließen lässt, ist die zweite Hälfte der Nachricht.
	defer func() { fehler = errors.Join(fehler, paket.Schliessen()) }()

	if paket.Art != msepaket.ArtSet {
		return nil, domain.NewFieldError("art", domain.SchluesselValidierung,
			"%q ist kein Set, sondern %q.", pfad, paket.Art)
	}

	set, roh, befunde := ausBaum(paket)

	// **Die Quelldatei reist mit** ([0011](../../docs/entscheidungen/0011-die-quelldatei-wird-mitgefuehrt.md)).
	//
	// Ohne sie ist die Byte-Zusage aus #48 nicht zu halten: aus `card_values`
	// kommen Einrückung, Abstände, Kommentare und Leerzeilen nicht zurück.
	//
	// **Ohne zweiten Griff ins Archiv**: `msepaket` hat die Hauptdatei beim
	// Öffnen gelesen und hält ihre Bytes. Sie noch einmal zu holen brächte
	// einen Fehlerweg für eine Datei, die gerade gelesen wurde — einen Zweig
	// ohne Fall (§3).
	set.Quelle = &domain.Quelle{
		Datei:  paket.Art.Hauptdatei(),
		Inhalt: paket.KopfRoh,
	}

	// **Die Domäne prüft, bevor der Bestand schreibt.**
	//
	// Bis zum Zweitblick über Phase 2 wurde `Validate` im ganzen Weg nie
	// gerufen: ein Set ohne Spiel landete still mit `spiel = ''` in der
	// Datenbank, weil das Schema nur `NOT NULL` verlangt, nicht „nicht leer".
	// Eine fachliche Regel, die nur in der Domäne steht und die niemand ruft,
	// ist keine.
	//
	// Geprüft wird, was **ohne** das Spielpaket prüfbar ist. Die Felddefinitionen
	// liegen dort, und ob es installiert ist, entscheidet sich erst in Phase 5.
	if befund := set.ValidateBasics(); befund != nil {
		return nil, befund
	}

	return &gelesenesSet{set: set, roh: roh, befunde: ausMsepaket(append(paket.Befunde, befunde...))}, nil
}

// Setvorschau sagt, was das Einfahren einer Datei täte — ohne zu schreiben.
type Setvorschau struct {
	Name, Spiel, Stil string
	Karten            int
	Befunde           []domain.Befund

	// SchonOffen ist die Kennung des Sets, das aus derselben Quelle schon im
	// Bestand liegt — `0`, wenn keines. Das Einfahren gäbe dann dieses zurück.
	SchonOffen int64
}

// Vorschau liest ein Setpaket wie [Importeur.Set] und hört vor dem Ablegen auf.
//
// **Ohne Platz in der Einfuhr**: sie schreibt nichts, und ein laufendes
// Einfahren soll niemandem die Frage verwehren, was eine Datei enthält.
func (i *Importeur) Vorschau(ctx context.Context, pfad string) (*Setvorschau, error) {
	schon, err := i.schonOffen(ctx, pfad)
	if err != nil {
		return nil, err
	}

	gelesen, err := setLesen(pfad)
	if err != nil {
		return nil, err
	}

	vorschau := &Setvorschau{
		Name: gelesen.set.Name, Spiel: gelesen.set.Game, Stil: gelesen.set.Style,
		Karten: len(gelesen.set.Cards), Befunde: gelesen.befunde,
	}

	if schon != nil {
		vorschau.SchonOffen = schon.SetID
	}

	return vorschau, nil
}
