// **Die Namen hinter den Kennungen eines Decks** —
// #1376.
//
// # Der Befund
//
// Die Deckliste zeigte, was ein Platz trägt: eine Kennung. Bei einem Druck eine
// UUID — `c58096e5-3cd4-4edd-b71a-3bdb47ab2536`. **Direkt darunter** stand nach
// einem Klick auf „Gegen das Format prüfen" dieselbe Karte mit Namen; nur die
// untere Liste war lesbar.
//
// # Warum im Dienst und nicht im Fenster
//
// **Damit jede Fläche sie hat** (§16). Ein Fenster, das die Namen selbst holt,
// lässt MCP und die Kommandozeile mit UUIDs zurück — und baut denselben Weg
// ein zweites Mal.
//
// # Was ohne Netz geschieht
//
// **Die Kennung bleibt stehen** (§16, Regel 5). Der Name eines Drucks kommt
// von Scryfall; ohne Netz bleibt das Feld leer, und das Fenster zeigt, was es
// hat. Eine Anwendung, die ohne Netz keine Deckliste mehr zeigt, ist falsch
// gebaut — und die Namen **eigener** Karten stehen im Bestand und kommen
// immer.

package service

import (
	"context"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
)

// Kartennamensnaht gibt den Namen einer eigenen Karte.
//
// **Dieselbe, die das Turnierwerk benutzt.** Eine zweite Naht für dieselbe
// Frage liefe mit jener auseinander.
type Kartennamensnaht interface {
	NameZurKarte(ctx context.Context, karteID int64) (string, error)

	// FeldZurKarte gibt einen beliebigen Feldwert — für die Manakosten.
	FeldZurKarte(ctx context.Context, karteID int64, feld string) (string, error)
}

// Drucknamensnaht gibt die Karten zu Druckkennungen — in **einem** Ruf.
//
// **Dieselbe Gestalt wie bei der Formatprüfung**: `POST /cards/collection` in
// Schüben zu 75 (#1373). Hundert Drucke sind zwei Anfragen statt hundert.
type Drucknamensnaht interface {
	Sammlung(ctx context.Context, kennungen []scryfall.Kennung) (
		scryfall.Sammelantwort, time.Duration, []domain.Befund, error)
}

// MitKartennamen setzt die Naht zu den Namen eigener Karten.
//
// **Als Nachrüstung und nicht als Pflichtfeld im Erzeuger.** Die
// Kommandozeile und der MCP-Server verdrahten nicht jede Naht; ein
// `nil`-Zugriff dort wäre ein Absturz für eine Bequemlichkeit.
func (d *Deckbau) MitKartennamen(naht Kartennamensnaht) *Deckbau {
	d.kartennamen = naht

	return d
}

// MitDrucknamen setzt die Naht zu den Namen echter Drucke.
func (d *Deckbau) MitDrucknamen(naht Drucknamensnaht) *Deckbau {
	d.drucknamen = naht

	return d
}

// namenEintragen schreibt die Namen in die Plätze.
//
// **Ohne Fehler nach außen.** Ein Name ist eine Zugabe; wer ihn nicht bekommt,
// bekommt die Kennung. Ein Deck, das wegen eines fehlenden Namens gar nicht
// aufgeht, wäre der schlechtere Tausch (§8).
func (d *Deckbau) namenEintragen(ctx context.Context, plaetze []domain.Deckplatz) {
	d.eigeneNamen(ctx, plaetze)
	d.drucknamenEintragen(ctx, plaetze)
}

// eigeneNamen trägt die Namen eigener Karten ein.
func (d *Deckbau) eigeneNamen(ctx context.Context, plaetze []domain.Deckplatz) {
	if d.kartennamen == nil {
		return
	}

	// **Je Karte einmal**, auch wenn sie im Hauptdeck und im Sideboard steht.
	gelesen := map[int64]string{}
	gelesenekosten := map[int64]string{}
	gelesenetypen := map[int64]string{}

	for stelle := range plaetze {
		kennung := plaetze[stelle].KarteID
		if kennung == 0 {
			continue
		}

		name, schon := gelesen[kennung]
		if !schon {
			gefunden, err := d.kartennamen.NameZurKarte(ctx, kennung)
			if err != nil {
				continue
			}

			name = gefunden
			gelesen[kennung] = name
		}

		plaetze[stelle].Bezeichnung = name

		// **Ein Fehler beim Feld lässt die Kosten leer, und das ist die
		// Antwort.** Nicht jedes Spiel führt „casting cost"; eine Deckliste
		// ohne Manasymbole ist besser als keine.
		kosten, schonKosten := gelesenekosten[kennung]
		if !schonKosten {
			//nolint:errcheck // Der leere Wert **ist** die Antwort, siehe oben.
			kosten, _ = d.kartennamen.FeldZurKarte(ctx, kennung, "casting cost")
			gelesenekosten[kennung] = kosten
		}

		plaetze[stelle].Manakosten = kosten

		// **Die Typzeile auf demselben Weg** (U10, #1421). Sie ist bei MSE
		// ein gerechnetes Feld mit `save value: false` (`docs/FORMATE.md`) —
		// steht sie nicht im Bestand, bleibt sie leer, und das ist die
		// Antwort.
		typzeile, schonTyp := gelesenetypen[kennung]
		if !schonTyp {
			//nolint:errcheck // Der leere Wert **ist** die Antwort, siehe oben.
			typzeile, _ = d.kartennamen.FeldZurKarte(ctx, kennung, "type")
			gelesenetypen[kennung] = typzeile
		}

		plaetze[stelle].Typzeile = typzeile
	}
}

// drucknamenEintragen trägt die Namen echter Drucke ein — in einem Ruf.
func (d *Deckbau) drucknamenEintragen(
	ctx context.Context, plaetze []domain.Deckplatz,
) {
	if d.drucknamen == nil {
		return
	}

	kennungen := make([]scryfall.Kennung, 0, len(plaetze))
	gesehen := map[string]bool{}

	for _, platz := range plaetze {
		if platz.Printing == "" || gesehen[platz.Printing] {
			continue
		}

		gesehen[platz.Printing] = true
		kennungen = append(kennungen, scryfall.Kennung{ID: platz.Printing})
	}

	if len(kennungen) == 0 {
		return
	}

	antwort, _, _, err := d.drucknamen.Sammlung(ctx, kennungen)
	if err != nil {
		return
	}

	// **Was nicht gefunden wurde, bleibt ohne Namen** — und das ist kein
	// Fehler: eine Kennung, die Scryfall nicht kennt, behält ihre Zeile (§8).
	namen := make(map[string]string, len(antwort.Karten))

	for _, karte := range antwort.Karten {
		namen[karte.Kennung] = karte.Name
	}

	// **Alles aus derselben Antwort** — Werkstück U10 von Phase 5.1 (#1421):
	// Kosten, Typzeile und Manawert kosten keinen zweiten Ruf.
	angaben := make(map[string]scryfall.Karte, len(antwort.Karten))

	for _, karte := range antwort.Karten {
		angaben[karte.Kennung] = *karte
	}

	for stelle := range plaetze {
		name, da := namen[plaetze[stelle].Printing]
		if !da {
			continue
		}

		karte := angaben[plaetze[stelle].Printing]

		plaetze[stelle].Bezeichnung = name
		plaetze[stelle].Manakosten = karte.Manakosten
		plaetze[stelle].Typzeile = karte.Typzeile
		plaetze[stelle].Manawert = karte.Manawert
	}
}
