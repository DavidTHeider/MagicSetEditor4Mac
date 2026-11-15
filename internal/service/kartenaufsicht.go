// Diese Datei hält die Setgrenze auf dem Lebenszyklus einer Karte.
//
// Sie steht für sich, weil sie eine Frage beantwortet, die keine andere Datei
// stellt: **gehört diese Karte in das Set, das der Aufrufer meint?**

package service

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Kartensicht sagt, zu welchem Set eine Karte gehört.
//
// **Die Schnittstelle steht hier, nicht in `storage`** (§2): sie ist das, was
// dieser Anwendungsfall braucht, und nicht das, was jener Speicher anbietet.
type Kartensicht interface {
	KarteLesen(ctx context.Context, karteID int64) (*domain.CardRow, error)
}

// Kartenlebenszyklus sind die vier Vorgänge, die eine Karte entstehen,
// verschwinden oder ihren Platz wechseln lassen.
type Kartenlebenszyklus interface {
	KarteAnlegen(ctx context.Context, setID int64) (int64, error)
	KarteKopieren(ctx context.Context, karteID int64) (int64, error)
	KarteLoeschen(ctx context.Context, karteID int64) error

	// KartenLoeschen entfernt mehrere in **einer** Transaktion und in einem
	// Verlaufsschritt — der Weg des Ausschneidens (Befund 4 des Zweitblicks
	// über Phase 3.7).
	KartenLoeschen(ctx context.Context, karteIDs []int64) error

	KartenTauschen(ctx context.Context, eineID, andereID int64) error
}

// Kartenaufsicht führt den Lebenszyklus einer Karte mit dem Set, in dem sie
// liegen muss.
//
// # Warum es sie gibt
//
// Die Vorgänge darunter nehmen eine blanke Kartenkennung. Im Fenster reicht
// das: die Kennung kommt aus der Liste des offenen Sets, und ein Mensch mit
// einer Sicht erfindet keine. **Über MCP nicht** — dort kommt sie von einem
// Sprachmodell, und eine verwechselte, erfundene oder aus einem früheren
// Gespräch übernommene Kennung ist kein Angriff, sondern ein gewöhnlicher
// Fehlerfall. Sie löschte bis zum 2026-08-22 eine Karte in einem Set, das
// niemand geöffnet hatte.
//
// Der Vermerk dazu steht in
// [0041](../../docs/entscheidungen/0041-die-setgrenze-liegt-im-dienst.md).
//
// # Was sie nicht ist
//
// Keine Rechteprüfung. Es gibt hier keine Nutzenden, keine Rollen und nichts
// zu verwehren — es gibt **eine** Frage, und sie lautet, ob zwei Angaben
// zusammenpassen. Wer hier mehr hineinliest, baut später eine Rechteverwaltung
// auf einen Vergleich zweier Zahlen.
type Kartenaufsicht struct {
	sicht  Kartensicht
	zyklus Kartenlebenszyklus

	// feld und massen kommen über [Kartenaufsicht.MitFeldweg] dazu.
	feld   Feldschreiber
	massen Massenfeldschreiber

	// formen kommt über [Kartenaufsicht.MitFeldformen] dazu — die Vorlage,
	// gegen die [Kartenaufsicht.FelderSetzen] ein Feld prüft.
	formen Kartenfeldquelle

	// decks kommt über [Kartenaufsicht.MitDeckwache] dazu — sie hält ein
	// Löschen an, das einen Deckplatz mitnähme (#1369).
	decks Deckwache
}

// NeueKartenaufsicht bindet Sicht und Lebenszyklus zusammen.
func NeueKartenaufsicht(sicht Kartensicht, zyklus Kartenlebenszyklus) *Kartenaufsicht {
	return &Kartenaufsicht{sicht: sicht, zyklus: zyklus}
}

// gehoertZu prüft, ob die Karte im genannten Set liegt.
//
// **`ErrNotFound` und nicht `ErrValidation`.** Aus der Sicht dessen, der ein
// Set genannt hat, gibt es diese Karte dort nicht — und das ist keine
// fehlerhafte Eingabe, sondern eine, die ins Leere zeigt. Dieselbe Antwort
// bekommt, wer eine Kennung nennt, die es überhaupt nicht gibt: **wer nach
// einer fremden Karte fragt, erfährt nicht, dass es sie gibt.**
func (a *Kartenaufsicht) gehoertZu(ctx context.Context, setID, karteID int64) error {
	karte, fehler := a.sicht.KarteLesen(ctx, karteID)
	if fehler != nil {
		return fehler
	}

	if karte.SetID != setID {
		return fremdeKarte(setID, karteID)
	}

	return nil
}

// fremdeKarte ist der Befund, wenn eine Karte nicht im genannten Set liegt —
// **nicht gefunden**, nicht „verboten": aus der Sicht dieses Sets gibt es sie
// nicht (0041).
func fremdeKarte(setID, karteID int64) error {
	return fmt.Errorf("%w: die Karte %d liegt nicht im Set %d",
		domain.ErrNotFound, karteID, setID)
}

// KarteAnlegen legt eine Karte im Set an.
//
// **Ohne Prüfung, und das ist kein Versehen**: es gibt keine Karte, deren
// Zugehörigkeit man halten müsste. Ob es das Set gibt, entscheidet der
// Fremdschlüssel darunter — eine zweite Abfrage davor wäre ein Zeitfenster,
// kein Schutz.
func (a *Kartenaufsicht) KarteAnlegen(ctx context.Context, setID int64) (int64, error) {
	// **Eine neue Karte rechnet ihre Werte** (#2516), wo der Feldschreiber
	// rechnet — wie MSE beim Anlegen (`AddCardAction`).
	if rechner, rechnet := a.feld.(Kartenrechner); rechnet {
		return rechner.KarteGerechnetAnlegen(ctx, setID)
	}

	return a.zyklus.KarteAnlegen(ctx, setID)
}

// Kartenrechner legt eine Karte mit ihren gerechneten Werten an (#2516).
type Kartenrechner interface {
	KarteGerechnetAnlegen(ctx context.Context, setID int64) (int64, error)
}

// KarteKopieren legt eine Kopie an, wenn die Karte in dieses Set gehört.
func (a *Kartenaufsicht) KarteKopieren(ctx context.Context, setID, karteID int64) (int64, error) {
	if fehler := a.gehoertZu(ctx, setID, karteID); fehler != nil {
		return 0, fehler
	}

	return a.zyklus.KarteKopieren(ctx, karteID)
}

// KarteLoeschen nimmt die Karte aus dem Set, wenn sie darin liegt.
//
// **`trotzdem` bestätigt ein Löschen, das einen Deckplatz mitnimmt** (#1369).
// Ohne die Bestätigung wird eine Karte, die in einem Deck liegt, abgewiesen —
// mit den Namen der Decks in der Meldung.
func (a *Kartenaufsicht) KarteLoeschen(
	ctx context.Context, setID, karteID int64, trotzdem bool,
) error {
	if fehler := a.gehoertZu(ctx, setID, karteID); fehler != nil {
		return fehler
	}

	if fehler := a.deckhindernis(ctx, []int64{karteID}, trotzdem); fehler != nil {
		return fehler
	}

	return a.zyklus.KarteLoeschen(ctx, karteID)
}

// KartenLoeschen nimmt mehrere Karten in einem Zug aus dem Set.
//
// **Jede wird geprüft, bevor die erste fällt.** Sonst stünde die Zusage „ganz
// oder gar nicht" auf dem Kopf: eine fremde Kennung mitten in der Liste ließe
// die Karten davor schon gelöscht zurück.
func (a *Kartenaufsicht) KartenLoeschen(
	ctx context.Context, setID int64, karteIDs []int64, trotzdem bool,
) error {
	for _, karteID := range karteIDs {
		if fehler := a.gehoertZu(ctx, setID, karteID); fehler != nil {
			return fehler
		}
	}

	// **Ebenfalls vor der ersten** (#1369): eine Rückfrage nach dem halben
	// Löschen wäre keine.
	if fehler := a.deckhindernis(ctx, karteIDs, trotzdem); fehler != nil {
		return fehler
	}

	return a.zyklus.KartenLoeschen(ctx, karteIDs)
}

// KartenTauschen vertauscht zwei Karten desselben Sets.
//
// **Beide werden geprüft, und zwar vor dem ersten Schreiben.** Der Tausch
// selbst weist einen über die Setgrenze hinweg ohnehin ab (`storage`); geprüft
// wird hier trotzdem, weil jene Prüfung eine **relative** ist — sie fragt, ob
// beide im selben Set liegen, nicht, ob es das genannte ist.
func (a *Kartenaufsicht) KartenTauschen(ctx context.Context, setID, eineID, andereID int64) error {
	if fehler := a.gehoertZu(ctx, setID, eineID); fehler != nil {
		return fehler
	}

	if fehler := a.gehoertZu(ctx, setID, andereID); fehler != nil {
		return fehler
	}

	return a.zyklus.KartenTauschen(ctx, eineID, andereID)
}
