// **Die Zuordnung an der Kommandozeile** — Werkstück U9 von Phase 5.1
// (#1420).

package main

import (
	"context"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// deckZuordnungen zeigt Bedarf und Zuordnungen eines Decks.
//
// **Der Rückgabewert sagt, ob etwas fehlt** — wie beim Abgleich.
func deckZuordnungen(
	ctx context.Context, flaeche *app.Deckflaeche, an *schreiber, rest []string,
) int {
	if len(rest) != 1 {
		an.klage("deck zuordnungen braucht genau eine Kennung.\n%s", hilfe)

		return CodeBedienung
	}

	antwort := flaeche.DeckZuordnungen(ctx, rest[0])
	if antwort.Fehler != nil {
		return klageUeberNaht(an, antwort.Fehler)
	}

	if an.alsJSON {
		if code := alsJSONSchreiben(an, antwort); code != CodeGut {
			return code
		}

		return urteilscode(antwort.Fehlend == 0)
	}

	// **Zugesagt steht dabei, seit es die Zahl wirklich gibt** (T2, #1467,
	// 0120). Ohne sie stünde „3/4, frei 1" da, und niemand könnte die beiden
	// Zahlen zusammenbringen — das Fenster nennt sie längst.
	for _, zeile := range antwort.Zeilen {
		an.sag("%d/%d\tfrei %d\tzugeordnet %d\tzugesagt %d\t%s\n",
			zeile.Vorhanden, zeile.Gebraucht, zeile.Frei, zeile.Zugeordnet,
			zeile.Reserviert, bezeichnungOder(zeile))

		for _, zuordnung := range zeile.Zuordnungen {
			an.sag("  #%s\t%d×\t%s%s\n", zuordnung.ID, zuordnung.Menge,
				zuordnung.Bezeichnung, abweichungstext(zuordnung.Abweichend))
		}
	}

	an.sag("fehlend %d, unbekannt %d\n", antwort.Fehlend, antwort.Unbekannt)

	if code := an.abschluss(); code != CodeGut {
		return code
	}

	return urteilscode(antwort.Fehlend == 0)
}

// bezeichnungOder gibt den Namen — sonst den Grund oder die Kennung.
func bezeichnungOder(zeile app.ZuordnungszeileDTO) string {
	switch {
	case zeile.Grund != "":
		return zeile.Grund
	case zeile.Bezeichnung != "":
		return zeile.Bezeichnung
	case zeile.Printing != "":
		return zeile.Printing
	default:
		return "Druck " + zeile.DruckID
	}
}

// abweichungstext markiert eine Zuordnung, die einen anderen Druck nennt.
func abweichungstext(abweichend bool) string {
	if abweichend {
		return "\t(anderer Druck)"
	}

	return ""
}

// deckZuordnen legt Exemplare in einen Platz.
func deckZuordnen(
	ctx context.Context, flaeche *app.Deckflaeche, an *schreiber, rest []string,
) int {
	if len(rest) != 4 {
		an.klage("deck zuordnen braucht Deck, Platz, Exemplar und Menge.\n%s", hilfe)

		return CodeBedienung
	}

	menge, fehler := strconv.Atoi(rest[3])
	if fehler != nil {
		an.klage("die Menge %q ist keine Zahl.\n%s", rest[3], hilfe)

		return CodeBedienung
	}

	antwort := flaeche.DeckZuordnen(ctx, rest[0], rest[1], rest[2], menge)
	if antwort.Fehler != nil {
		return klageUeberNaht(an, antwort.Fehler)
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, antwort.Zuordnung)
	}

	an.sag("zugeordnet: %d× Exemplar %s an Platz %s (#%s)\n",
		antwort.Zuordnung.Menge, antwort.Zuordnung.ExemplarID,
		antwort.Zuordnung.PlatzID, antwort.Zuordnung.ID)

	return an.abschluss()
}

// deckZuordnungLoesen nimmt eine Zuordnung heraus.
func deckZuordnungLoesen(
	ctx context.Context, flaeche *app.Deckflaeche, an *schreiber, rest []string,
) int {
	if len(rest) != 2 {
		an.klage("deck zuordnung-loesen braucht Deck und Zuordnung.\n%s", hilfe)

		return CodeBedienung
	}

	antwort := flaeche.DeckZuordnungLoesen(ctx, rest[0], rest[1])
	if antwort.Fehler != nil {
		return klageUeberNaht(an, antwort.Fehler)
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, map[string]any{"zuordnungID": rest[1], "geloest": true})
	}

	an.sag("gelöst: Zuordnung %s\n", rest[1])

	return an.abschluss()
}

// deckFehlkarten gibt, was zu besorgen wäre — als Liste.
func deckFehlkarten(
	ctx context.Context, flaeche *app.Deckflaeche, an *schreiber, rest []string,
) int {
	if len(rest) != 1 {
		an.klage("deck fehlkarten braucht genau eine Kennung.\n%s", hilfe)

		return CodeBedienung
	}

	antwort := flaeche.DeckFehlkarten(ctx, rest[0])
	if antwort.Fehler != nil {
		return klageUeberNaht(an, antwort.Fehler)
	}

	if an.alsJSON {
		if code := alsJSONSchreiben(an, antwort); code != CodeGut {
			return code
		}

		return urteilscode(antwort.Liste == "")
	}

	an.sag("%s", antwort.Liste)

	if code := an.abschluss(); code != CodeGut {
		return code
	}

	// **Fehlende Karten sind ein fachliches Nein** — wie beim Abgleich.
	return urteilscode(antwort.Liste == "")
}

// deckListe gibt ein Deck als Liste — in der Gestalt, die `deck-liste` liest.
func deckListe(
	ctx context.Context, flaeche *app.Deckflaeche, an *schreiber, rest []string,
) int {
	if len(rest) != 1 {
		an.klage("deck liste braucht genau eine Kennung.\n%s", hilfe)

		return CodeBedienung
	}

	antwort := flaeche.DecklisteAusgeben(ctx, rest[0])
	if antwort.Fehler != nil {
		return klageUeberNaht(an, antwort.Fehler)
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, antwort)
	}

	an.sag("%s", antwort.Liste)

	return an.abschluss()
}

// exemplarDecks sagt, welche Decks ein Exemplar benutzen.
//
// **Ein eigener Befehl und kein Unterbefehl von `sammlung`**: dort ist das
// erste Wort eine **Druckkennung** (`mse sammlung <printing>`), und ein
// Unterbefehl »decks« wäre ein Druck, der so heißt.
func exemplarDecks(
	ctx context.Context, flaeche *app.Deckflaeche, an *schreiber, rest []string,
) int {
	if len(rest) != 1 {
		an.klage("exemplar decks braucht genau eine Kennung.\n%s", hilfe)

		return CodeBedienung
	}

	antwort := flaeche.ExemplarDecks(ctx, rest[0])
	if antwort.Fehler != nil {
		return klageUeberNaht(an, antwort.Fehler)
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, antwort.Decks)
	}

	for _, deck := range antwort.Decks {
		an.sag("%s\t%d×\t%s\n", deck.DeckID, deck.Menge, deck.Name)
	}

	return an.abschluss()
}
