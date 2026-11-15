package mcp

// **Die Werkzeuge der Zuordnung** — Werkstück U9 von Phase 5.1
// (#1420).
//
// **Eine eigene Datei**, damit `decks.go` und `decksschreiben.go` unter der
// Zeilengrenze aus §2 bleiben. Der Schnitt folgt der Frage: dort steht, was
// ein Deck ist und was darin liegt; hier, **welche Exemplare** dafür
// vorgesehen sind.

import (
	"context"
	"encoding/json"
)

// lesenDeckzuordnungen gibt Bedarf und Zuordnungen eines Decks.
func lesenDeckzuordnungen(naht Decks) Werkzeug {
	return Werkzeug{
		Name: "lesen.deck.zuordnungen", Art: Lesend,
		Erklaerung: "Sagt je Druck, was das Deck braucht, was davon in der " +
			"Sammlung liegt, wie viel davon **frei** ist (Besitz minus " +
			"Reservierungen minus Zuordnungen anderer Decks) und welche " +
			"Exemplare diesem Deck ausdrücklich zugeordnet sind. Eine " +
			"Deckliste allein reserviert nichts. ÄNDERT NICHTS.",
		Schema: objekt(map[string]any{
			"deckID": zahl("Die Kennung des Decks."),
		}, "deckID"),
		Ruf: deckruf(naht.DeckZuordnungen),
	}
}

// lesenDeckfehlkarten gibt, was zu besorgen wäre.
func lesenDeckfehlkarten(naht Decks) Werkzeug {
	return Werkzeug{
		Name: "lesen.deck.fehlkarten", Art: Lesend,
		Erklaerung: "Gibt die fehlenden Karten eines Decks als Liste in der " +
			"Gestalt, die »schreiben.sammlung.einlesen« versteht — Menge, " +
			"Name, Set und Nummer je Zeile. ÄNDERT NICHTS.",
		Schema: objekt(map[string]any{
			"deckID": zahl("Die Kennung des Decks."),
		}, "deckID"),
		Ruf: deckruf(naht.DeckFehlkarten),
	}
}

// lesenDeckliste gibt ein Deck als Liste — NET-04 (#1708).
func lesenDeckliste(naht Decks) Werkzeug {
	return Werkzeug{
		Name: "lesen.deck.liste", Art: Lesend,
		Erklaerung: "Gibt ein Deck als Liste in der Gestalt, die »lesen.deckliste« " +
			"wieder liest — Hauptdeck und Sideboard unter ihren Überschriften, " +
			"Menge und Name je Zeile. ÄNDERT NICHTS.",
		Schema: objekt(map[string]any{
			"deckID": zahl("Die Kennung des Decks."),
		}, "deckID"),
		Ruf: deckruf(naht.Deckliste),
	}
}

// lesenExemplardecks gibt den Rückweg von einem Exemplar zu seinen Decks.
func lesenExemplardecks(naht Decks) Werkzeug {
	return Werkzeug{
		Name: "lesen.exemplar.decks", Art: Lesend,
		Erklaerung: "Sagt, welche Decks ein Exemplar der Sammlung " +
			"ausdrücklich benutzen — der Rückweg zu »lesen.deck.zuordnungen«. " +
			"ÄNDERT NICHTS.",
		Schema: objekt(map[string]any{
			"exemplarID": zahl("Die Kennung des Exemplars aus »lesen.sammlung«."),
		}, "exemplarID"),
		// **Kein `deckruf`**: der liest `deckID`, und hier heißt die Angabe
		// anders. Ein Ruf, der stillschweigend die falsche Zahl läse, gäbe
		// jedem Aufruf die Decks des Exemplars 0.
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var wahl struct {
				ExemplarID int64 `json:"exemplarID"`
			}

			if err := lies(roh, &wahl); err != nil {
				return "", err
			}

			ergebnis, fehler := naht.ExemplarDecks(ctx, wahl.ExemplarID)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(ergebnis)
		},
	}
}

// schreibenDeckZuordnen legt Exemplare in einen Platz.
func schreibenDeckZuordnen(naht Decks) Werkzeug {
	return Werkzeug{
		Name: "schreiben.deck.zuordnen", Art: Schreibend,
		Erklaerung: "ÄNDERT den Bestand: ordnet einem Deckplatz Exemplare der " +
			"Sammlung zu. Sie sind danach für andere Decks nicht mehr frei. " +
			"Ein Wunsch lässt sich nicht zuordnen; ein Exemplar eines anderen " +
			"Drucks geht, wird aber als Abweichung geführt. Dieselbe " +
			"Zuordnung noch einmal setzt die Menge neu.",
		Schema: objekt(map[string]any{
			"deckID":     zahl("Die Kennung des Decks."),
			"platzID":    zahl("Die Kennung des Platzes aus »lesen.deck«."),
			"exemplarID": zahl("Die Kennung des Exemplars aus »lesen.sammlung«."),
			"menge":      zahl("Wie viele Stücke — mindestens eines."),
		}, "deckID", "platzID", "exemplarID", "menge"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var wahl struct {
				DeckID     int64 `json:"deckID"`
				PlatzID    int64 `json:"platzID"`
				ExemplarID int64 `json:"exemplarID"`
				Menge      int   `json:"menge"`
			}

			if err := lies(roh, &wahl); err != nil {
				return "", err
			}

			ergebnis, fehler := naht.DeckZuordnen(ctx, wahl.DeckID, wahl.PlatzID, wahl.ExemplarID, wahl.Menge)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(ergebnis)
		},
	}
}

// schreibenDeckZuordnungLoesen nimmt eine Zuordnung heraus.
func schreibenDeckZuordnungLoesen(naht Decks) Werkzeug {
	return Werkzeug{
		Name: "schreiben.deck.zuordnung.loesen", Art: Schreibend,
		Erklaerung: "ÄNDERT den Bestand: nimmt eine Zuordnung aus dem Deck. " +
			"Die Exemplare sind danach wieder frei; das Deck behält seinen " +
			"Platz und seinen Bedarf.",
		Schema: objekt(map[string]any{
			"deckID":      zahl("Die Kennung des Decks."),
			"zuordnungID": zahl("Die Kennung der Zuordnung aus »lesen.deck.zuordnungen«."),
		}, "deckID", "zuordnungID"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var wahl struct {
				DeckID      int64 `json:"deckID"`
				ZuordnungID int64 `json:"zuordnungID"`
			}

			if err := lies(roh, &wahl); err != nil {
				return "", err
			}

			if fehler := naht.DeckZuordnungLoesen(ctx, wahl.DeckID, wahl.ZuordnungID); fehler != nil {
				return "", fehler
			}

			return alsJSON(map[string]any{"zuordnungID": wahl.ZuordnungID, "geloest": true})
		},
	}
}
