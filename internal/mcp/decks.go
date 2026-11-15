// **Decks und Formate über MCP** — Werkstück 7 von Phase 4.6
// (#1293).
//
// Einordnung nach §16: **erweiternd** — fünfzehn Werkzeuge kommen hinzu, fünf
// lesend und zehn schreibend. Keines ändert eine bestehende Fläche.

package mcp

import (
	"context"
	"encoding/json"
)

// Decks ist, was MCP für Decks und Formate braucht.
//
// **Eine Schnittstelle für beides**, weil beides zusammen gebraucht wird: ein
// Deck ohne Format lässt sich nicht prüfen, ein Format ohne Deck nicht
// benutzen.
type Decks interface {
	Decks(ctx context.Context, spiel string) (any, error)
	Deck(ctx context.Context, deckID int64) (any, error)
	DeckAnlegen(ctx context.Context, name, spiel, notiz string) (any, error)
	DeckUmbenennen(ctx context.Context, deckID, fassung int64, name, notiz string) (
		any, error)
	DeckLoeschen(ctx context.Context, deckID int64) error

	// DeckKopieren legt eine Kopie unter neuem Namen an (DOM-03.AK-02, #1698).
	DeckKopieren(ctx context.Context, deckID int64, name string) (any, error)

	// DeckAusSet legt aus den Karten eines Sets ein Deck an (FMT-04, #1693).
	DeckAusSet(ctx context.Context, setID int64, name string, menge int) (any, error)

	// DeckFehlkartenWuenschen setzt, was fehlt, auf die Wunschliste (DOM-03).
	DeckFehlkartenWuenschen(ctx context.Context, deckID int64) (any, error)
	DeckplatzLegen(ctx context.Context, deckID, karteID int64,
		printing string, menge int, bereich string) (any, error)
	DeckplatzNehmen(ctx context.Context, deckID, platzID int64) error
	DeckplatzVerschieben(ctx context.Context, deckID, platzID int64, bereich string) (
		any, error)
	DeckFormatWaehlen(ctx context.Context, deckID, formatID int64) (any, error)
	DeckPruefen(ctx context.Context, deckID int64) (any, error)
	DeckAbgleichen(ctx context.Context, deckID int64) (any, error)

	// **Die Zuordnung** — Werkstück U9 von Phase 5.1 (#1420).
	DeckZuordnungen(ctx context.Context, deckID int64) (any, error)
	DeckZuordnen(ctx context.Context, deckID, platzID, exemplarID int64, menge int) (any, error)
	DeckZuordnungLoesen(ctx context.Context, deckID, zuordnungID int64) error
	DeckFehlkarten(ctx context.Context, deckID int64) (any, error)

	// Deckliste gibt ein Deck als Liste (NET-04, #1708).
	Deckliste(ctx context.Context, deckID int64) (any, error)
	ExemplarDecks(ctx context.Context, exemplarID int64) (any, error)

	Bauarten(ctx context.Context) (any, error)
	Formate(ctx context.Context, spiel string) (any, error)
	FormatAusBauart(ctx context.Context, bauart, name, spiel, scryfallname string) (
		any, error)
	FormatSetzen(ctx context.Context, format map[string]any) (any, error)
	FormatLoeschen(ctx context.Context, formatID int64) error
}

// decks gibt die zwanzig Werkzeuge.
func decks(naht Decks) (lesend, schreibend []Werkzeug) {
	if naht == nil {
		return nil, nil
	}

	return []Werkzeug{
		lesenDecks(naht), lesenDeck(naht), lesenDeckpruefung(naht),
		lesenDeckabgleich(naht), lesenBauarten(naht),
		lesenDeckzuordnungen(naht), lesenDeckfehlkarten(naht), lesenExemplardecks(naht),
		lesenDeckliste(naht),
	}, []Werkzeug{
		schreibenDeckAnlegen(naht), schreibenDeckUmbenennen(naht),
		schreibenDeckLoeschen(naht), schreibenDeckKopieren(naht), schreibenDeckAusSet(naht),
		schreibenDeckFehlkartenWuenschen(naht),
		schreibenDeckplatzLegen(naht),
		schreibenDeckplatzNehmen(naht), schreibenDeckplatzVerschieben(naht),
		schreibenDeckFormat(naht),
		schreibenFormatBauart(naht), schreibenFormatSetzen(naht),
		schreibenFormatLoeschen(naht), lesenFormate(naht),
		schreibenDeckZuordnen(naht), schreibenDeckZuordnungLoesen(naht),
	}
}

// lesenDecks nennt die Decks eines Spiels.
func lesenDecks(naht Decks) Werkzeug {
	return Werkzeug{
		Name: "lesen.decks", Art: Lesend,
		Erklaerung: "Nennt die Decks — ohne Angabe alle, sonst die eines Spiels. " +
			"Ein Deck ist eine Zusammenstellung und kein Set: es verweist auf " +
			"Karten, es kopiert sie nicht. ÄNDERT NICHTS.",
		Schema: objekt(map[string]any{
			"spiel": text("Der Name des Spiels — leer für alle."),
		}),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var wahl struct {
				Spiel string `json:"spiel"`
			}

			if err := lies(roh, &wahl); err != nil {
				return "", err
			}

			liste, fehler := naht.Decks(ctx, wahl.Spiel)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(liste)
		},
	}
}

// lesenDeck gibt ein Deck samt seinen Plätzen.
func lesenDeck(naht Decks) Werkzeug {
	return Werkzeug{
		Name: "lesen.deck", Art: Lesend,
		Erklaerung: "Gibt ein Deck samt seinen Plätzen. Ein Platz nennt " +
			"ENTWEDER eine eigene Karte (karteID) ODER einen echten Druck " +
			"(printing) — nie beides. ÄNDERT NICHTS.",
		Schema: objekt(map[string]any{
			"deckID": zahl("Die Kennung des Decks aus »lesen.decks«."),
		}, "deckID"),
		Ruf: deckruf(naht.Deck),
	}
}

// lesenDeckpruefung beurteilt ein Deck gegen sein Format.
func lesenDeckpruefung(naht Decks) Werkzeug {
	return Werkzeug{
		Name: "lesen.deck.pruefung", Art: Lesend,
		Erklaerung: "Prüft ein Deck gegen sein gewähltes Format. Je Karte " +
			"kommt eines von Scryfalls Wörtern — legal, not_legal, restricted, " +
			"banned — oder »ungeprüft«. ACHTUNG: »ungeprüft« heißt NICHT " +
			"»unzulässig«: eine selbst entworfene Karte ist in keinem Format " +
			"gelistet, und das macht das Deck nicht kaputt. ÄNDERT NICHTS.",
		Schema: objekt(map[string]any{
			"deckID": zahl("Die Kennung des Decks. Es muss ein Format gewählt haben."),
		}, "deckID"),
		Ruf: deckruf(naht.DeckPruefen),
	}
}

// lesenDeckabgleich sagt, was von einem Deck in der Sammlung liegt.
func lesenDeckabgleich(naht Decks) Werkzeug {
	return Werkzeug{
		Name: "lesen.deck.bestand", Art: Lesend,
		Erklaerung: "Sagt, was von einem Deck in der Sammlung liegt und was " +
			"fehlt — je Druck, nicht je Platz: vier im Hauptdeck und eine im " +
			"Sideboard verlangen fünf Karten. »fehlend« und »unbekannt« stehen " +
			"getrennt; ein Wunsch zählt nicht als Besitz. ÄNDERT NICHTS.",
		Schema: objekt(map[string]any{
			"deckID": zahl("Die Kennung des Decks."),
		}, "deckID"),
		Ruf: deckruf(naht.DeckAbgleichen),
	}
}

// lesenBauarten nennt die drei Bauarten samt ihren Zahlen.
func lesenBauarten(naht Decks) Werkzeug {
	return Werkzeug{
		Name: "lesen.formatbauarten", Art: Lesend,
		Erklaerung: "Nennt die drei Deckbauarten samt den Zahlen, die in den " +
			"Comprehensive Rules stehen: konstruiert (60 Karten, 15 im " +
			"Sideboard, höchstens 4 gleiche), commander (genau 100, kein " +
			"Sideboard, jede Karte einmal) und limitiert (mindestens 40). " +
			"Sie sind ein Vorschlag für ein neues Format, keine Vorschrift. " +
			"ÄNDERT NICHTS.",
		Schema: objekt(map[string]any{}),
		Ruf: func(ctx context.Context, _ json.RawMessage) (string, error) {
			liste, fehler := naht.Bauarten(ctx)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(liste)
		},
	}
}

// lesenFormate nennt die Formate eines Spiels.
//
// **Es steht bei den schreibenden und ist lesend** — nein: es steht in der
// **Liste** der schreibenden Rückgabe, weil `decks` genau zwei Listen gibt.
// Seine Art ist [Lesend], und danach entscheidet der Server.
func lesenFormate(naht Decks) Werkzeug {
	return Werkzeug{
		Name: "lesen.formate", Art: Lesend,
		Erklaerung: "Nennt die Formate — ohne Angabe alle, sonst die eines " +
			"Spiels. Ein Format ist hier ein Datensatz und keine eingebaute " +
			"Liste: es wird angelegt, nicht mitgeliefert. ÄNDERT NICHTS.",
		Schema: objekt(map[string]any{
			"spiel": text("Der Name des Spiels — leer für alle."),
		}),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var wahl struct {
				Spiel string `json:"spiel"`
			}

			if err := lies(roh, &wahl); err != nil {
				return "", err
			}

			liste, fehler := naht.Formate(ctx, wahl.Spiel)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(liste)
		},
	}
}

// deckruf baut den Ruf eines Werkzeugs, das nur eine Deckkennung nimmt.
//
// **Vier Werkzeuge, ein Ruf.** Ohne ihn stünde derselbe Block viermal da, und
// eine Änderung daran träfe drei davon.
func deckruf(arbeit func(context.Context, int64) (any, error)) func(
	context.Context, json.RawMessage) (string, error) {
	return func(ctx context.Context, roh json.RawMessage) (string, error) {
		var wahl struct {
			DeckID int64 `json:"deckID"`
		}

		if err := lies(roh, &wahl); err != nil {
			return "", err
		}

		ergebnis, fehler := arbeit(ctx, wahl.DeckID)
		if fehler != nil {
			return "", fehler
		}

		return alsJSON(ergebnis)
	}
}
