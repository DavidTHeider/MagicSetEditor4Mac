// **Die schreibenden Deckwerkzeuge** — Werkstück 7 von Phase 4.6
// (#1293).
//
// **Eine eigene Datei**, damit `decks.go` unter der Zeilengrenze aus §2
// bleibt. Der Schnitt folgt der Frage: dort steht, was ein Modell **erfährt**,
// hier, was es **ändert**.

package mcp

import (
	"context"
	"encoding/json"
)

// schreibenDeckAnlegen legt ein leeres Deck an.
func schreibenDeckAnlegen(naht Decks) Werkzeug {
	return Werkzeug{
		Name: "schreiben.deck.anlegen", Art: Schreibend,
		Erklaerung: "ÄNDERT den Bestand: legt ein leeres Deck an. Die Plätze " +
			"kommen danach mit »schreiben.deckplatz.legen« hinein. Ein Deck " +
			"gehört zu einem Spiel — ohne das gäbe es kein Format.",
		Schema: objekt(map[string]any{
			"name":  text("Der Name des Decks. Je Spiel nur einmal."),
			"spiel": text("Der Name des Spiels, etwa »magic«."),
			"notiz": text("Was jemand sich dazu aufschreibt. Darf leer sein."),
		}, "name", "spiel"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var wahl struct {
				Name  string `json:"name"`
				Spiel string `json:"spiel"`
				Notiz string `json:"notiz"`
			}

			if err := lies(roh, &wahl); err != nil {
				return "", err
			}

			deck, fehler := naht.DeckAnlegen(ctx, wahl.Name, wahl.Spiel, wahl.Notiz)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(deck)
		},
	}
}

// schreibenDeckUmbenennen schreibt Name und Notiz fort.
func schreibenDeckUmbenennen(naht Decks) Werkzeug {
	return Werkzeug{
		Name: "schreiben.deck.umbenennen", Art: Schreibend,
		Erklaerung: "ÄNDERT den Bestand: schreibt Name und Notiz eines Decks " +
			"fort. Die Fassung aus »lesen.deck« gehört dazu — hat inzwischen " +
			"jemand anderes geschrieben, wird der Ruf abgewiesen statt zu " +
			"überschreiben.",
		Schema: objekt(map[string]any{
			"deckID":  zahl("Die Kennung des Decks."),
			"fassung": zahl("Die Fassung aus »lesen.deck«."),
			"name":    text("Der neue Name."),
			"notiz":   text("Die neue Notiz. Darf leer sein."),
		}, "deckID", "fassung", "name"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var wahl struct {
				DeckID  int64  `json:"deckID"`
				Fassung int64  `json:"fassung"`
				Name    string `json:"name"`
				Notiz   string `json:"notiz"`
			}

			if err := lies(roh, &wahl); err != nil {
				return "", err
			}

			deck, fehler := naht.DeckUmbenennen(
				ctx, wahl.DeckID, wahl.Fassung, wahl.Name, wahl.Notiz)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(deck)
		},
	}
}

// schreibenDeckKopieren legt eine Kopie unter neuem Namen an — DOM-03.AK-02
// (#1698).
func schreibenDeckKopieren(naht Decks) Werkzeug {
	return Werkzeug{
		Name: "schreiben.deck.kopieren", Art: Schreibend,
		Erklaerung: "ÄNDERT den Bestand: legt eine Kopie eines Decks an — mit " +
			"Spiel, Notiz, Format und allen Plätzen, unter einem neuen Namen. " +
			"Zugeordnete Exemplare, Leihen und Marken bleiben beim Original. " +
			"Einen Namen, den es im Spiel schon gibt, weist der Ruf ab.",
		Schema: objekt(map[string]any{
			"deckID": zahl("Die Kennung des Decks, das kopiert wird."),
			"name":   text("Der Name der Kopie."),
		}, "deckID", "name"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var wahl struct {
				DeckID int64  `json:"deckID"`
				Name   string `json:"name"`
			}

			if err := lies(roh, &wahl); err != nil {
				return "", err
			}

			kopie, fehler := naht.DeckKopieren(ctx, wahl.DeckID, wahl.Name)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(kopie)
		},
	}
}

// schreibenDeckFehlkartenWuenschen setzt, was einem Deck fehlt, auf die
// Wunschliste — DOM-03.AK-02 (#1698).
func schreibenDeckFehlkartenWuenschen(naht Decks) Werkzeug {
	return Werkzeug{
		Name: "schreiben.deck.fehlkarten.wuenschen", Art: Schreibend,
		Erklaerung: "ÄNDERT den Bestand: setzt jeden Druck, der dem Deck fehlt " +
			"und noch nicht gewünscht ist, mit der fehlenden Menge auf die " +
			"Wunschliste. Ein schon gewünschter Druck bleibt, wie er ist, und " +
			"steht unter »vorhanden«; ein zweiter Ruf legt nichts doppelt an.",
		Schema: objekt(map[string]any{
			"deckID": zahl("Die Kennung des Decks."),
		}, "deckID"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var wahl struct {
				DeckID int64 `json:"deckID"`
			}

			if err := lies(roh, &wahl); err != nil {
				return "", err
			}

			ergebnis, fehler := naht.DeckFehlkartenWuenschen(ctx, wahl.DeckID)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(ergebnis)
		},
	}
}

// schreibenDeckLoeschen nimmt ein Deck heraus.
func schreibenDeckLoeschen(naht Decks) Werkzeug {
	return Werkzeug{
		Name: "schreiben.deck.loeschen", Art: Schreibend,
		Erklaerung: "ÄNDERT den Bestand: nimmt ein Deck samt seinen Plätzen " +
			"heraus. Die Karten selbst bleiben — ein Platz verweist nur.",
		Schema: objekt(map[string]any{
			"deckID": zahl("Die Kennung des Decks."),
		}, "deckID"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var wahl struct {
				DeckID int64 `json:"deckID"`
			}

			if err := lies(roh, &wahl); err != nil {
				return "", err
			}

			if fehler := naht.DeckLoeschen(ctx, wahl.DeckID); fehler != nil {
				return "", fehler
			}

			return alsJSON(map[string]any{"deckID": wahl.DeckID, "geloescht": true})
		},
	}
}

// schreibenDeckplatzLegen legt einen Platz ins Deck.
func schreibenDeckplatzLegen(naht Decks) Werkzeug {
	return Werkzeug{
		Name: "schreiben.deckplatz.legen", Art: Schreibend,
		Erklaerung: "ÄNDERT den Bestand: legt eine Karte ins Deck oder " +
			"schreibt ihre Menge fort. Gib ENTWEDER karteID (eine eigene " +
			"Karte) ODER printing (eine Scryfall-Druckkennung) — nie beides " +
			"und nie keines. Derselbe Bezug im selben Bereich ist eine höhere " +
			"Menge, kein zweiter Platz.",
		Schema: objekt(map[string]any{
			"deckID":   zahl("Die Kennung des Decks."),
			"karteID":  zahl("Die Kennung einer eigenen Karte. Weglassen für einen Druck."),
			"printing": text("Die Scryfall-Druckkennung. Weglassen für eine eigene Karte."),
			"menge":    zahl("Die Stückzahl — mindestens eine."),
			"bereich":  text("»hauptdeck« oder »sideboard«."),
		}, "deckID", "menge", "bereich"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var wahl struct {
				DeckID   int64  `json:"deckID"`
				KarteID  int64  `json:"karteID"`
				Printing string `json:"printing"`
				Menge    int    `json:"menge"`
				Bereich  string `json:"bereich"`
			}

			if err := lies(roh, &wahl); err != nil {
				return "", err
			}

			platz, fehler := naht.DeckplatzLegen(ctx, wahl.DeckID, wahl.KarteID,
				wahl.Printing, wahl.Menge, wahl.Bereich)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(platz)
		},
	}
}

// schreibenDeckplatzNehmen nimmt einen Platz heraus.
func schreibenDeckplatzNehmen(naht Decks) Werkzeug {
	return Werkzeug{
		Name: "schreiben.deckplatz.nehmen", Art: Schreibend,
		Erklaerung: "ÄNDERT den Bestand: nimmt einen Platz aus dem Deck. Die " +
			"Karte selbst bleibt.",
		Schema: objekt(map[string]any{
			"deckID":  zahl("Die Kennung des Decks."),
			"platzID": zahl("Die Kennung des Platzes aus »lesen.deck«."),
		}, "deckID", "platzID"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var wahl struct {
				DeckID  int64 `json:"deckID"`
				PlatzID int64 `json:"platzID"`
			}

			if err := lies(roh, &wahl); err != nil {
				return "", err
			}

			if fehler := naht.DeckplatzNehmen(ctx, wahl.DeckID, wahl.PlatzID); fehler != nil {
				return "", fehler
			}

			return alsJSON(map[string]any{"platzID": wahl.PlatzID, "genommen": true})
		},
	}
}

// schreibenDeckFormat stellt ein Deck auf ein Format.
func schreibenDeckFormat(naht Decks) Werkzeug {
	return Werkzeug{
		Name: "schreiben.deck.format", Art: Schreibend,
		Erklaerung: "ÄNDERT den Bestand: stellt ein Deck auf ein Format. " +
			"formatID 0 nimmt es wieder herunter — dann ist das Deck " +
			"ungeprüft, nicht falsch. Das Format muss zum Spiel des Decks " +
			"gehören.",
		Schema: objekt(map[string]any{
			"deckID":   zahl("Die Kennung des Decks."),
			"formatID": zahl("Die Kennung des Formats aus »lesen.formate«; 0 nimmt es herunter."),
		}, "deckID"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var wahl struct {
				DeckID   int64 `json:"deckID"`
				FormatID int64 `json:"formatID"`
			}

			if err := lies(roh, &wahl); err != nil {
				return "", err
			}

			deck, fehler := naht.DeckFormatWaehlen(ctx, wahl.DeckID, wahl.FormatID)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(deck)
		},
	}
}
