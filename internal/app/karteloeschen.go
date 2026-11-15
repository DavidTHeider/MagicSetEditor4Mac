package app

// Eine Karte löschen — #621.
//
// # Warum das ein eigenes Werkstück war
//
// #562 hat das
// Anlegen gebaut und das Löschen ausdrücklich vertagt:
//
//	Es gehört in denselben Zug, geht aber nicht ohne Rücknahme: eine gelöschte
//	Karte ohne ⌘Z ist Datenverlust mit einem Klick.
//
// Der Verlauf trug nur Feldänderungen. Seit er eine eigene Art für ganze
// Karten hat und die Rücknahme dieselbe Kennung zurückbringt, ist der Weg
// zurück da — und damit ist das Löschen vertretbar.

import (
	"context"
)

// Kartenloescher entfernt eine Karte — die Naht zum Speicher.
//
// **Eine eigene Naht neben [Kartenanleger]**, obwohl derselbe Schreiber beide
// erfüllt. Ein Name, der die Hälfte seiner Methoden nicht mehr beschreibt, ist
// eine kleine Unwahrheit, die jeder spätere Leser erst nachprüfen muss.
type Kartenloescher interface {
	KarteLoeschen(ctx context.Context, setID, karteID int64, trotzdem bool) error
}

// KarteLoeschen entfernt eine Karte samt ihren Feldwerten.
//
// **Die Rücknahme liegt im Speicher, nicht hier.** Diese Schicht übersetzt
// zwischen Wails und dem Dienst (§2); dass ein Löschen zurückzunehmen ist,
// entscheidet der Verlauf.
// **`trotzdem` bestätigt ein Löschen, das einen Deckplatz mitnimmt** (#1369).
// Ohne die Bestätigung wird eine Karte, die in einem Deck liegt, abgewiesen —
// mit den Namen der Decks in der Meldung.
func (a *App) KarteLoeschen(
	ctx context.Context, setID, karteID string, trotzdem bool,
) SchreibAntwort {
	ctx = a.imProfil(ctx)

	setkennung, befund := benannteKennungAus("setID", setID)
	if befund != nil {
		return SchreibAntwort{Fehler: befund}
	}

	kennung, befund := kennungAus(karteID)
	if befund != nil {
		return SchreibAntwort{Fehler: befund}
	}

	return SchreibAntwort{
		Fehler: Uebersetze(ctx, a.loescher.KarteLoeschen(ctx, setkennung, kennung, trotzdem)),
	}
}

// MitKartenloescher setzt die Naht — für Tests und den Composition Root.
func (a *App) MitKartenloescher(loescher Kartenloescher) { a.loescher = loescher }
