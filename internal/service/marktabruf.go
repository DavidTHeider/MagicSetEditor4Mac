package service

// **Angebote zu einem Druck** — Werkstück A4 von Phase 5.3
// (#1443).
//
// # Was hier fachlich entschieden wird
//
// **Welcher Blueprint zu welchem Druck gehört.** Ein Marktplatz führt seine
// eigene Nummer je Kartenfassung; dieser Bestand führt Drucke. Dazwischen
// liegt eine Auflösung, und sie hat Regeln, die aus A2 stammen
// (#1441):
// **mehrdeutig ist ein Konflikt, kein Automerge.**
//
// # Was hier **nicht** geschieht
//
// **Es wird nichts gekauft.** Der Marktplatz wird gelesen; was ein Konto tut,
// ist Sache von T11
// (#1476).

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Marktquelle ist, was dieser Anwendungsfall von einem Marktplatz braucht.
//
// **Als Schnittstelle und in `domain`-Begriffen** (§2, §16): die Gestalt eines
// Anbieters gehört nicht in die Fachlogik, und ein Lauf hier soll ohne Netz
// auskommen.
type Marktquelle interface {
	// Quelle ist die Familie, unter der alles von diesem Marktplatz läuft.
	Quelle() domain.Preisquelle

	// System ist das Kennungssystem, unter dem dieser Marktplatz seine
	// Fassungsnummern vergibt.
	//
	// **Er sagt es selbst, statt dass es jemand aus [Marktquelle.Quelle]
	// ableitet.** Preisquelle und Kennungssystem sind zwei Namensräume;
	// eine Umwandlung über den Text wäre die stille Behauptung, dass sie
	// einer sind — und ergäbe für einen unbekannten Anbieter ein System,
	// das [domain.Druckkennung.Validate] abweist.
	System() domain.Kennungssystem

	// Fassungen gibt die Kartenfassungen einer Erweiterung, je Fassung ihre
	// Fremdkennungen.
	Fassungen(ctx context.Context, setcode string) (
		[]domain.Fremdfassung, []domain.Befund, error)

	// Angebote gibt die Posten zu einer Fassung.
	Angebote(ctx context.Context, fassung string, wunsch domain.Angebotswunsch) (
		[]domain.Marktangebot, []domain.Befund, error)
}

// Marktregister ist, was dieser Anwendungsfall von der Druckablage braucht.
//
// **Zwei Methoden und nicht vier**: [Druckregister] kann mehr, und ein Abruf,
// der nach Kandidaten suchen dürfte, könnte einen Druck anlegen. Er soll
// nicht.
//
// **Gesucht wird über den Scryfall-Alias und nicht über die Zeilennummer**,
// und das ist zugleich eine fachliche Aussage: ein Druck ohne Alias ist eine
// Eigenkarte oder eine vorläufige Zeile
// ([0086](../../docs/entscheidungen/0086-sieben-begriffe-fuer-eine-karte.md)),
// und kein Marktplatz der Welt kennt sie. Wer hier die Zeilennummer nähme,
// müsste denselben Fall als Fehlerweg nachbauen — und hätte einen Zweig, den
// keine Fläche erreicht.
type Marktregister interface {
	NachScryfall(ctx context.Context, printing string) (*domain.Druck, error)
	Sichern(ctx context.Context, druck *domain.Druck) error
}

// Marktergebnis ist, was ein Abruf ergeben hat.
type Marktergebnis struct {
	// Fassung ist die Nummer der Kartenfassung beim Anbieter.
	Fassung string

	// Zugeordnet sagt, ob die Fassung gerade erst gefunden wurde.
	Zugeordnet bool

	// Wunsch ist, wonach gefragt wurde — so wie er hinausging.
	Wunsch domain.Angebotswunsch

	Angebote []domain.Marktangebot
	Befunde  []domain.Befund
}

// Marktabruf liest die Angebote eines Marktplatzes zu einem Druck.
type Marktabruf struct {
	drucke Marktregister
	quelle Marktquelle
}

// NeuerMarktabruf baut ihn.
func NeuerMarktabruf(drucke Marktregister, quelle Marktquelle) *Marktabruf {
	return &Marktabruf{drucke: drucke, quelle: quelle}
}

// Angebote gibt, was zu einem Druck angeboten wird.
func (m *Marktabruf) Angebote(
	ctx context.Context, printing string, wunsch domain.Angebotswunsch,
) (Marktergebnis, error) {
	druck, fehler := m.drucke.NachScryfall(ctx, printing)
	if fehler != nil {
		return Marktergebnis{}, fehler
	}

	if druck == nil {
		return Marktergebnis{}, fmt.Errorf(
			"%w: den Druck %q gibt es in diesem Bestand nicht",
			domain.ErrNotFound, printing)
	}

	ergebnis, fehler := m.fassungFuer(ctx, druck)
	if fehler != nil {
		return Marktergebnis{}, fehler
	}

	ergebnis.Wunsch = wunsch

	angebote, befunde, holfehler := m.quelle.Angebote(ctx, ergebnis.Fassung, wunsch)
	if holfehler != nil {
		return Marktergebnis{}, holfehler
	}

	ergebnis.Befunde = append(ergebnis.Befunde, befunde...)
	ergebnis.Angebote = m.durchgelassen(angebote, wunsch, druck.Scryfall())

	return ergebnis, nil
}

// durchgelassen setzt Quelle und Druck und wendet den Wunsch an.
//
// **Der Wunsch wirkt zweimal**, und das ist kein doppelter Weg: er steht in
// der Abfrage an den Anbieter **und** hier. Ein Anbieter, der ihn übergeht,
// schickt sonst Posten, nach denen niemand gefragt hat — und niemand sähe es
// (§16, Drift wird erkannt, nicht erlitten).
func (m *Marktabruf) durchgelassen(
	angebote []domain.Marktangebot, wunsch domain.Angebotswunsch, alias string,
) []domain.Marktangebot {
	durch := make([]domain.Marktangebot, 0, len(angebote))

	for _, angebot := range angebote {
		angebot.Quelle = m.quelle.Quelle()
		angebot.Druck = alias

		if wunsch.Passt(angebot) {
			durch = append(durch, angebot)
		}
	}

	return durch
}
