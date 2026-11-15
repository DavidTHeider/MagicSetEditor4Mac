// **Die ändernden Deckbefehle** — Werkstück 7 von Phase 4.6
// (#1293).
//
// **Eine eigene Datei**, damit `decks.go` unter der Zeilengrenze aus §2
// bleibt. Der Schnitt folgt der Frage: dort steht, was man **erfährt**, hier,
// was man **ändert**.

package main

import (
	"context"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// deckUmbenennen schreibt Name und Notiz fort.
func deckUmbenennen(
	ctx context.Context, flaeche *app.Deckflaeche, an *schreiber, rest []string,
) int {
	if len(rest) < 3 {
		an.klage("deck umbenennen braucht Kennung, Fassung und Namen.\n%s", hilfe)

		return CodeBedienung
	}

	notiz := ""
	if len(rest) > 3 {
		notiz = rest[3]
	}

	antwort := flaeche.DeckUmbenennen(ctx, rest[0], rest[1], rest[2], notiz)
	if antwort.Fehler != nil {
		return klageUeberNaht(an, antwort.Fehler)
	}

	return getan(an, nil, "Umbenannt.\n")
}

// deckKopieren legt eine Kopie unter neuem Namen an und nennt ihre Kennung.
func deckKopieren(
	ctx context.Context, flaeche *app.Deckflaeche, an *schreiber, rest []string,
) int {
	if len(rest) != 2 {
		an.klage("deck kopieren braucht die Kennung und den Namen der Kopie.\n%s", hilfe)

		return CodeBedienung
	}

	antwort := flaeche.DeckKopieren(ctx, rest[0], rest[1])
	if antwort.Fehler != nil {
		return klageUeberNaht(an, antwort.Fehler)
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, antwort)
	}

	an.sag("%s\n", antwort.Deck.ID)

	return an.abschluss()
}

// deckWuenschen setzt, was einem Deck fehlt, auf die Wunschliste — DOM-03.AK-02
// (#1698).
//
// **Der Rückgabewert ist null, auch wenn nichts neu dazukam**: ein Druck, der
// schon gewünscht ist, ist kein Fehlschlag, und die Ausgabe nennt ihn.
func deckWuenschen(
	ctx context.Context, flaeche *app.Deckflaeche, an *schreiber, rest []string,
) int {
	if len(rest) != 1 {
		an.klage("deck wuenschen braucht genau eine Kennung.\n%s", hilfe)

		return CodeBedienung
	}

	antwort := flaeche.DeckFehlkartenWuenschen(ctx, rest[0])
	if antwort.Fehler != nil {
		return klageUeberNaht(an, antwort.Fehler)
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, antwort)
	}

	for _, neu := range antwort.Neu {
		an.sag("neu\t%d×\t%s\n", neu.Menge, wunschname(neu.Name, neu.Printing))
	}

	for _, schon := range antwort.Vorhanden {
		an.sag("schon gewünscht\t%d von %d\t%s\n", schon.Gewuenscht, schon.Fehlend,
			wunschname(schon.Bezeichnung, schon.Printing))
	}

	if antwort.Unbekannt > 0 {
		an.sag("ohne Druck\t%d\n", antwort.Unbekannt)
	}

	return an.abschluss()
}

// wunschname gibt den Namen — oder die Kennung, wo keiner bekannt ist.
func wunschname(name, printing string) string {
	if name == "" {
		return printing
	}

	return name
}

// deckLoeschen nimmt ein Deck heraus.
func deckLoeschen(
	ctx context.Context, flaeche *app.Deckflaeche, an *schreiber, rest []string,
) int {
	if len(rest) != 1 {
		an.klage("deck loeschen braucht genau eine Kennung.\n%s", hilfe)

		return CodeBedienung
	}

	antwort := flaeche.DeckLoeschen(ctx, rest[0])
	if antwort.Fehler != nil {
		return klageUeberNaht(an, antwort.Fehler)
	}

	return getan(an, nil, "Gelöscht.\n")
}

// deckplatzLegen legt eine Karte ins Deck.
//
// **`--karte=<id>` oder ein Druck** — genau eines von beiden, wie die Domäne
// es verlangt.
func deckplatzLegen(
	ctx context.Context, flaeche *app.Deckflaeche, an *schreiber, rest []string,
) int {
	if len(rest) < 3 {
		an.klage("deck legen braucht Deck, Bezug und Menge.\n%s", hilfe)

		return CodeBedienung
	}

	menge, err := strconv.Atoi(rest[2])
	if err != nil {
		an.klage("die Menge %q ist keine Zahl.\n%s", rest[2], hilfe)

		return CodeBedienung
	}

	bereich := "hauptdeck"
	if len(rest) > 3 {
		bereich = rest[3]
	}

	karte, printing := "", rest[1]
	if nach, ist := abgeschnitten(rest[1], "--karte="); ist {
		karte, printing = nach, ""
	}

	antwort := flaeche.DeckplatzLegen(ctx, rest[0], karte, printing, menge, bereich)
	if antwort.Fehler != nil {
		return klageUeberNaht(an, antwort.Fehler)
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, antwort)
	}

	an.sag("%s\n", antwort.Platz.ID)

	return an.abschluss()
}

// deckplatzNehmen nimmt einen Platz heraus.
func deckplatzNehmen(
	ctx context.Context, flaeche *app.Deckflaeche, an *schreiber, rest []string,
) int {
	if len(rest) != 2 {
		an.klage("deck nehmen braucht Deck und Platz.\n%s", hilfe)

		return CodeBedienung
	}

	antwort := flaeche.DeckplatzNehmen(ctx, rest[0], rest[1])
	if antwort.Fehler != nil {
		return klageUeberNaht(an, antwort.Fehler)
	}

	return getan(an, nil, "Genommen.\n")
}

// deckplatzVerschieben legt einen Platz in den anderen Bereich.
//
// **Ein Ruf und nicht zwei** — Befund 7 des Zweitblicks über Phase 5.1.
func deckplatzVerschieben(
	ctx context.Context, flaeche *app.Deckflaeche, an *schreiber, rest []string,
) int {
	if len(rest) != 3 {
		an.klage("deck verschieben braucht Deck, Platz und Bereich.\n%s", hilfe)

		return CodeBedienung
	}

	antwort := flaeche.DeckplatzVerschieben(ctx, rest[0], rest[1], rest[2])
	if antwort.Fehler != nil {
		return klageUeberNaht(an, antwort.Fehler)
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, antwort)
	}

	an.sag("%s\n", antwort.Platz.Bereich)

	return an.abschluss()
}

// deckFormatWaehlen stellt ein Deck auf ein Format — **ohne Kennung nimmt es
// herunter**.
func deckFormatWaehlen(
	ctx context.Context, flaeche *app.Deckflaeche, an *schreiber, rest []string,
) int {
	if len(rest) == 0 {
		an.klage("deck format braucht ein Deck.\n%s", hilfe)

		return CodeBedienung
	}

	format := ""
	if len(rest) > 1 {
		format = rest[1]
	}

	antwort := flaeche.DeckFormatWaehlen(ctx, rest[0], format)
	if antwort.Fehler != nil {
		return klageUeberNaht(an, antwort.Fehler)
	}

	return getan(an, nil, "Gestellt.\n")
}

// abgeschnitten gibt den Rest hinter einem Präfix — und ob er dastand.
func abgeschnitten(wort, praefix string) (string, bool) {
	if len(wort) <= len(praefix) || wort[:len(praefix)] != praefix {
		return "", false
	}

	return wort[len(praefix):], true
}
