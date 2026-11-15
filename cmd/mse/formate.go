// **Formate an der Kommandozeile** — Werkstück 7 von Phase 4.6
// (#1293).
//
// Ein Format ist **Daten** (Vermerk 0061): wer ein eigenes spielt, legt es an.

package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// formatbefehl bedient `mse formate` und `mse format …`.
func formatbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	return mitDecks(ctx, bestandsort, an, func(ctx context.Context, flaeche *app.Deckflaeche) int {
		if len(rest) == 0 {
			return formatlisteZeigen(ctx, flaeche, an, "")
		}

		switch rest[0] {
		case "bauarten":
			return bauartenZeigen(flaeche, an)

		case "neu":
			return formatAnlegen(ctx, flaeche, an, rest[1:])

		case "setzen":
			return formatSetzen(ctx, flaeche, an, rest[1:])

		case "loeschen":
			return formatLoeschen(ctx, flaeche, an, rest[1:])
		}

		return formatlisteZeigen(ctx, flaeche, an, rest[0])
	})
}

// formatlisteZeigen nennt die Formate.
func formatlisteZeigen(
	ctx context.Context, flaeche *app.Deckflaeche, an *schreiber, spiel string,
) int {
	antwort := flaeche.Formate(ctx, spiel)
	if antwort.Fehler != nil {
		return klageUeberNaht(an, antwort.Fehler)
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, antwort)
	}

	for _, format := range antwort.Formate {
		an.sag("%s\t%s\t%s\t%d–%d\t%s\t%d gleiche\t%s\n",
			format.ID, format.Spiel, format.Name,
			format.Mindestgroesse, format.Hoechstgroesse,
			sideboardwort(format), format.HoechstzahlGleich, format.Scryfallname)
	}

	return an.abschluss()
}

// bauartenZeigen nennt die drei Bauarten samt ihren nachgelesenen Zahlen.
func bauartenZeigen(flaeche *app.Deckflaeche, an *schreiber) int {
	antwort := flaeche.Bauarten()

	if an.alsJSON {
		return alsJSONSchreiben(an, antwort)
	}

	for _, bauart := range antwort.Bauarten {
		an.sag("%s\t%d–%d\t%s\t%d gleiche\n", bauart.Bauart,
			bauart.Format.Mindestgroesse, bauart.Format.Hoechstgroesse,
			sideboardwort(bauart.Format), bauart.Format.HoechstzahlGleich)
	}

	return an.abschluss()
}

// formatAnlegen legt ein Format aus einer Bauart an.
func formatAnlegen(
	ctx context.Context, flaeche *app.Deckflaeche, an *schreiber, rest []string,
) int {
	if len(rest) < 3 {
		an.klage("format neu braucht Bauart, Namen und Spiel.\n%s", hilfe)

		return CodeBedienung
	}

	scryfallname := ""
	if len(rest) > 3 {
		scryfallname = rest[3]
	}

	antwort := flaeche.FormatAusBauart(ctx, rest[0], rest[1], rest[2], scryfallname)
	if antwort.Fehler != nil {
		return klageUeberNaht(an, antwort.Fehler)
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, antwort)
	}

	an.sag("%s\n", antwort.Format.ID)

	return an.abschluss()
}

// formatSetzen schreibt ein Format fort.
//
// **Alle Zahlen auf einmal**, und das ist Absicht: `0` heißt bei dreien
// „keine Grenze" und beim Sideboard „keines". Ein Befehl, der einzelne Felder
// setzte, müsste „nicht genannt" von „null" unterscheiden — und an der
// Kommandozeile kann er das nicht.
//
// **Die siebte Stelle nimmt zusätzlich das Wort `offen`.** Ein Sideboard ohne
// Obergrenze ist etwas anderes als keines (CR 100.4b), und es hat deshalb
// seine eigene Marke. Sie **an dieser Stelle** zu nehmen statt als eigenes
// Kennzeichen ist der Punkt: eine Stelle, die genannt werden **muss**, kann
// nicht vergessen werden — und wer den Namen eines limitierten Formats
// ändert, verlöre es sonst still (Befund 1 des Nachlaufs über Phase 4.6).
func formatSetzen(
	ctx context.Context, flaeche *app.Deckflaeche, an *schreiber, rest []string,
) int {
	if len(rest) < 8 {
		an.klage("format setzen braucht Kennung, Fassung, Namen, Spiel und "+
			"vier Zahlen.\n%s", hilfe)

		return CodeBedienung
	}

	zahlen := make([]int, 0, 4)
	offen := false

	for stelle, roh := range rest[4:8] {
		if stelle == sideboardstelle && strings.EqualFold(roh, "offen") {
			offen = true
			zahlen = append(zahlen, 0)

			continue
		}

		wert, err := strconv.Atoi(roh)
		if err != nil {
			an.klage("%q ist keine Zahl.\n%s", roh, hilfe)

			return CodeBedienung
		}

		zahlen = append(zahlen, wert)
	}

	scryfallname := ""
	if len(rest) > 8 {
		scryfallname = rest[8]
	}

	antwort := flaeche.FormatSetzen(ctx, app.FormatDTO{
		ID: rest[0], Fassung: rest[1], Name: rest[2], Spiel: rest[3],
		Mindestgroesse: zahlen[0], Hoechstgroesse: zahlen[1],
		Sideboardgroesse: zahlen[2], SideboardOffen: offen,
		HoechstzahlGleich: zahlen[3],
		Scryfallname:      scryfallname,
	})
	if antwort.Fehler != nil {
		return klageUeberNaht(an, antwort.Fehler)
	}

	return getan(an, nil, "Gesetzt.\n")
}

// formatLoeschen nimmt ein Format heraus — die Decks bleiben.
func formatLoeschen(
	ctx context.Context, flaeche *app.Deckflaeche, an *schreiber, rest []string,
) int {
	if len(rest) != 1 {
		an.klage("format loeschen braucht genau eine Kennung.\n%s", hilfe)

		return CodeBedienung
	}

	antwort := flaeche.FormatLoeschen(ctx, rest[0])
	if antwort.Fehler != nil {
		return klageUeberNaht(an, antwort.Fehler)
	}

	return getan(an, nil,
		"Gelöscht. Decks, die darauf standen, sind jetzt ungeprüft.\n")
}

// sideboardstelle ist, an welcher der vier Zahlen das Sideboard steht.
//
// **Als Konstante und nicht als `2` im Code**: die Stelle steht in der Hilfe,
// im Test und hier, und drei Zweien laufen auseinander.
const sideboardstelle = 2

// sideboardwort sagt in einem Wort, was mit dem Sideboard ist.
//
// **Drei Antworten für zwei Felder.** Wer nur die Zahl zeigte, machte „kein
// Sideboard" und „ohne Grenze" wieder ununterscheidbar — genau der Fehler,
// gegen den es die Marke gibt.
func sideboardwort(format app.FormatDTO) string {
	if format.SideboardOffen {
		return "Sideboard offen"
	}

	return fmt.Sprintf("%d im Sideboard", format.Sideboardgroesse)
}
