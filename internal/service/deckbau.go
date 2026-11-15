// **Der Deckbau** — Werkstück 3 von Phase 4.6
// (#1293),
// nach Vermerk [0061](../../docs/entscheidungen/0061-was-ein-deck-hier-ist.md).
//
// # Was diese Schicht dazutut
//
// Der Speicher schreibt Zeilen; hier steht, **was gilt**. Zwei Regeln kann nur
// diese Schicht halten, weil beide **mehrere** Entitäten betreffen (§2):
//
//   - Ein Platz kommt in ein Deck, **das es gibt**. Ohne diese Frage entstünde
//     aus einer falschen Kennung ein Fremdschlüsselfehler, und der sagt dem
//     Aufrufer nichts.
//   - Eine **eigene** Karte gehört zum Spiel des Decks. Ein Deck aus Karten
//     zweier Spiele ließe sich gegen kein Format prüfen — und ein Format ist
//     immer das eines Spiels.

package service

import (
	"context"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Deckspeicher ist, was der Deckbau zum Lesen und Schreiben braucht.
//
// **Die Schnittstelle steht hier und wird in `storage` erfüllt** (§2).
type Deckspeicher interface {
	Anlegen(ctx context.Context, deck domain.Deck) (domain.Deck, error)
	Aendern(ctx context.Context, deck domain.Deck) (domain.Deck, error)
	Deck(ctx context.Context, kennung int64) (domain.Deck, error)
	Decks(ctx context.Context, spiel string) ([]domain.Deck, error)
	PlatzSetzen(ctx context.Context, deck int64, platz domain.Deckplatz) (
		domain.Deckplatz, error)
	PlatzLoeschen(ctx context.Context, deck, platz int64) error

	// PlatzVerschieben legt einen Platz in den anderen Bereich — **in einem
	// Zug**, samt dem Zusammenlegen mit einem Platz, der dort schon liegt.
	//
	// **Nicht zwei Rufe**, und das ist der ganze Grund: Befund 7 des
	// Zweitblicks über Phase 5.1 hat gezeigt, dass „legen, dann nehmen" über
	// die Naht eine Karte in beiden Bereichen zurücklassen kann.
	PlatzVerschieben(ctx context.Context, deck, platz int64, bereich domain.Deckbereich) (
		domain.Deckplatz, error)
	Loeschen(ctx context.Context, kennung int64) error

	// Kopieren legt eine Kopie unter neuem Namen an — **Zeile und Plätze in
	// einem Zug** (DOM-03.AK-02, #1698).
	Kopieren(ctx context.Context, quelle int64, name string) (domain.Deck, error)
}

// Kartenspiel sagt, zu welchem Spiel eine eigene Karte gehört.
//
// **Eine Schnittstelle mit einem Verfahren**, und das ist Absicht: der
// Deckbau braucht von den Karten genau diese eine Auskunft, und eine breitere
// Schnittstelle machte jedes Doppel zu einer Kartenverwaltung.
type Kartenspiel interface {
	SpielZurKarte(ctx context.Context, karteID int64) (string, error)
}

// Formatnachschlag sagt, welches Format eine Kennung meint.
//
// **Eine eigene Schnittstelle und nicht [Formatspeicher]**: der Deckbau liest
// Formate und schreibt sie nicht. Ihm den ganzen Speicher zu geben hieße, ihm
// das Löschen mitzugeben, das er nicht braucht.
type Formatnachschlag interface {
	Format(ctx context.Context, kennung int64) (domain.Format, error)
}

// Deckbau ist der Anwendungsfall „ein Deck zusammenstellen".
type Deckbau struct {
	// kartennamen und drucknamen sind die Nähte zu den Namen (#1376) —
	// nachgerüstet, weil nicht jede Fläche sie verdrahtet.
	kartennamen Kartennamensnaht
	drucknamen  Drucknamensnaht

	speicher Deckspeicher
	karten   Kartenspiel
	formate  Formatnachschlag
}

// NeuerDeckbau baut ihn über einem Speicher.
func NeuerDeckbau(
	speicher Deckspeicher, karten Kartenspiel, formate Formatnachschlag,
) *Deckbau {
	return &Deckbau{speicher: speicher, karten: karten, formate: formate}
}

// FormatWaehlen stellt ein Deck auf ein Format — **null nimmt es herunter**.
//
// # Die Regel, die nur hier steht
//
// **Das Format gehört zum Spiel des Decks.** Ohne sie ließe sich ein
// Magic-Deck gegen ein VS-Format stellen, und die Antwort der Prüfung wäre
// Unsinn, nicht falsch.
//
// **Ein Deck ohne Format ist ungeprüft, nicht falsch** — dieselbe Haltung wie
// bei einer eigenen Karte, die Scryfall nicht kennt.
func (d *Deckbau) FormatWaehlen(ctx context.Context, deckID, formatID int64) (
	domain.Deck, error,
) {
	deck, err := d.speicher.Deck(ctx, deckID)
	if err != nil {
		return domain.Deck{}, err
	}

	if formatID != 0 {
		format, lesefehler := d.formate.Format(ctx, formatID)
		if lesefehler != nil {
			return domain.Deck{}, lesefehler
		}

		if format.Spiel != deck.Spiel {
			return domain.Deck{}, domain.NewFieldError("formatID",
				domain.SchluesselValidierung,
				"Das Format %q gehört zum Spiel %q, das Deck zu %q.",
				format.Name, format.Spiel, deck.Spiel)
		}
	}

	deck.FormatID = formatID

	return d.speicher.Aendern(ctx, deck)
}

// Anlegen legt ein Deck an.
//
// **Ein Deck, das Plätze mitbringt, wird abgewiesen.** Der Speicher legt die
// Deckzeile an und sonst nichts; ein Deck mit drei Plätzen käme mit null
// zurück, und der Aufrufer sähe einen Erfolg. Lieber ein lautes Nein als ein
// stiller Verlust (§8).
func (d *Deckbau) Anlegen(ctx context.Context, deck domain.Deck) (domain.Deck, error) {
	if err := deckVorbereiten(&deck); err != nil {
		return domain.Deck{}, err
	}

	if len(deck.Plaetze) > 0 {
		return domain.Deck{}, domain.NewFieldError("plaetze", domain.SchluesselValidierung,
			"Ein Deck entsteht leer; die Plätze kommen danach hinein.")
	}

	return d.speicher.Anlegen(ctx, deck)
}

// Umbenennen schreibt Name und Notiz fort.
//
// **Die Fassung ist die Bedingung** (§7, Regel 2). Sie steckt im
// mitgegebenen Deck, und das Ergebnis trägt die neue.
func (d *Deckbau) Umbenennen(ctx context.Context, deck domain.Deck) (domain.Deck, error) {
	if err := deckVorbereiten(&deck); err != nil {
		return domain.Deck{}, err
	}

	return d.speicher.Aendern(ctx, deck)
}

// Deck gibt ein Deck samt seinen Plätzen — **mit ihren Namen**
// (#1376).
func (d *Deckbau) Deck(ctx context.Context, kennung int64) (domain.Deck, error) {
	deck, err := d.speicher.Deck(ctx, kennung)
	if err != nil {
		return domain.Deck{}, err
	}

	d.namenEintragen(ctx, deck.Plaetze)

	return deck, nil
}

// Decks gibt die Decks eines Spiels — ein leeres Spiel heißt: alle.
func (d *Deckbau) Decks(ctx context.Context, spiel string) ([]domain.Deck, error) {
	return d.speicher.Decks(ctx, spiel)
}

// PlatzLegen legt einen Platz ins Deck oder schreibt seine Menge fort.
//
// **Ein Weg für beide Herkünfte.** Ob der Platz eine eigene Karte oder einen
// echten Druck nennt, entscheidet der Platz selbst; zwei Verfahren dafür wären
// zwei Fassungen derselben Prüfung.
func (d *Deckbau) PlatzLegen(
	ctx context.Context, deckID int64, platz domain.Deckplatz,
) (domain.Deckplatz, error) {
	platz.Printing = strings.TrimSpace(platz.Printing)

	if befunde := platz.Validate(); len(befunde) > 0 {
		return domain.Deckplatz{}, befunde[0]
	}

	deck, err := d.speicher.Deck(ctx, deckID)
	if err != nil {
		return domain.Deckplatz{}, err
	}

	if err := d.karteGehoertZumSpiel(ctx, deck.Spiel, platz.KarteID); err != nil {
		return domain.Deckplatz{}, err
	}

	return d.speicher.PlatzSetzen(ctx, deckID, platz)
}

// karteGehoertZumSpiel prüft, ob eine **eigene** Karte zum Spiel des Decks
// gehört. Für einen echten Druck ist nichts zu prüfen — er gehört keinem Set.
func (d *Deckbau) karteGehoertZumSpiel(
	ctx context.Context, spiel string, karteID int64,
) error {
	if karteID == 0 {
		return nil
	}

	kartenspiel, err := d.karten.SpielZurKarte(ctx, karteID)
	if err != nil {
		return err
	}

	if kartenspiel != spiel {
		return domain.NewFieldError("karteID", domain.SchluesselValidierung,
			"Diese Karte gehört zum Spiel %q, das Deck zu %q.", kartenspiel, spiel)
	}

	return nil
}

// PlatzVerschieben legt einen Platz in den anderen Bereich.
//
// **Ohne Spielprüfung**, und das ist kein Versäumnis: der Platz liegt bereits
// im Deck, und ein Bereichswechsel ändert seinen Bezug nicht. Geprüft wurde,
// als er hineinkam.
func (d *Deckbau) PlatzVerschieben(
	ctx context.Context, deckID, platzID int64, bereich domain.Deckbereich,
) (domain.Deckplatz, error) {
	return d.speicher.PlatzVerschieben(ctx, deckID, platzID, bereich)
}

// PlatzNehmen nimmt einen Platz aus dem Deck.
func (d *Deckbau) PlatzNehmen(ctx context.Context, deckID, platzID int64) error {
	return d.speicher.PlatzLoeschen(ctx, deckID, platzID)
}

// Kopieren legt eine Kopie eines Decks unter neuem Namen an — DOM-03.AK-02
// (#1698).
//
// **Der Name kommt vom Aufrufer und wird nicht erfunden.** Im Spiel ist er
// einmalig, und ein „(Kopie)", das hier entstünde, stünde in der Sprache
// dieses Programms im Bestand — das Fenster schlägt ihn in seiner vor.
//
// **Geprüft wird nur der Name**: Spiel und Plätze sind geprüft, seit sie im
// Original stehen, und die Kopie ändert an ihnen nichts.
func (d *Deckbau) Kopieren(ctx context.Context, quelle int64, name string) (
	domain.Deck, error,
) {
	name = strings.TrimSpace(name)
	if name == "" {
		return domain.Deck{}, domain.NewFieldError("name", domain.SchluesselPflichtfeld,
			"Eine Kopie braucht einen eigenen Namen.")
	}

	kopie, err := d.speicher.Kopieren(ctx, quelle, name)
	if err != nil {
		return domain.Deck{}, err
	}

	d.namenEintragen(ctx, kopie.Plaetze)

	return kopie, nil
}

// Loeschen nimmt ein Deck samt seinen Plätzen heraus.
func (d *Deckbau) Loeschen(ctx context.Context, kennung int64) error {
	return d.speicher.Loeschen(ctx, kennung)
}

// deckVorbereiten glättet ein Deck und prüft es.
//
// **In dieser Reihenfolge**, wie beim Exemplar: wer erst prüft, weist ein
// `"  Aggro"` ab, das nach dem Glätten in Ordnung wäre.
func deckVorbereiten(deck *domain.Deck) error {
	deck.Normalisieren()

	if befunde := deck.Validate(); len(befunde) > 0 {
		// **Der erste Befund geht hinaus, nicht alle** — dieselbe Naht wie
		// überall, und derselbe ausgesprochene Preis: wer zwei Felder falsch
		// ausgefüllt hat, korrigiert sie nacheinander.
		return befunde[0]
	}

	return nil
}
