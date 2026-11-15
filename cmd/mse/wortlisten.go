// **Die Auswahllisten eines Kartenfeldes** —
// #1074.
//
// Zwei Befehle, wie an jeder anderen Fläche: einer zeigt, einer wählt. Die
// Trennung ist §16 Regel 6 — nichts wird still übernommen.
package main

import (
	"context"
	"database/sql"
	"strconv"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// wortlistenbefehl zeigt, was ein Feld zur Auswahl anbietet.
func wortlistenbefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) < 2 {
		an.klage("wortlisten braucht die Kartenkennung und den Feldnamen.\n%s", hilfe)

		return CodeBedienung
	}

	karteID, gut := kennung(an, "die Kartenkennung", rest[0])
	if !gut {
		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		naht, err := app.WortlistennahtUeber(db, domain.UrheberKommandozeile)
		if err != nil {
			an.klageUeber(err)

			return CodeFachlich
		}

		bereiche, lesefehler := naht.Wortlisten(ctx, karteID, rest[1])
		if lesefehler != nil {
			an.klageUeber(lesefehler)

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, bereiche)
		}

		zeigeBereiche(an, bereiche)

		return an.abschluss()
	})
}

// zeigeBereiche schreibt die Auswahl in lesbarer Form.
//
// **Der Weg steht vor jedem Eintrag**, denn er ist das, was der zweite Befehl
// braucht. Eine Liste, aus der man die Nummer abzählen muss, wäre eine Liste
// zum Abzählen.
func zeigeBereiche(an *schreiber, bereiche []app.WortlistenbereichDTO) {
	if len(bereiche) == 0 {
		an.sag("Dieses Feld hat keine Auswahlliste.\n")

		return
	}

	for nummer, bereich := range bereiche {
		an.sag("Bereich %d — %s: %q\n", nummer, bereich.Liste, bereich.Text)

		zeigeEintraege(an, bereich.Eintraege, nil)
	}
}

// zeigeEintraege schreibt ein Menü samt seiner Untermenüs.
func zeigeEintraege(an *schreiber, eintraege []app.WortmenueeintragDTO, weg []int) {
	for nummer, eintrag := range eintraege {
		hierhin := append(append([]int(nil), weg...), nummer)

		// **Kästchen und Punkte sehen verschieden aus.** Ein Präfix wird
		// umgeschaltet, ein Punkt ersetzt — wer das nicht sieht, klickt
		// zweimal auf dasselbe und wundert sich.
		marke := "( )"
		if eintrag.IstPraefix {
			marke = "[ ]"
		}

		if eintrag.Aktiv {
			marke = strings.Replace(marke, " ", "x", 1)
		}

		an.sag("  %-10s %s %s\n", wegtext(hierhin), marke, eintrag.Name)

		zeigeEintraege(an, eintrag.Kinder, hierhin)
	}
}

// wegtext macht aus einem Weg die Form, die der Wahlbefehl entgegennimmt.
func wegtext(weg []int) string {
	teile := make([]string, 0, len(weg))
	for _, schritt := range weg {
		teile = append(teile, strconv.Itoa(schritt))
	}

	return strings.Join(teile, ".")
}

// wortwahlbefehl wählt einen Eintrag und schreibt das Feld.
func wortwahlbefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) < 5 {
		an.klage("wortwahl braucht Set, Karte, Feld, Bereich und Weg.\n%s", hilfe)

		return CodeBedienung
	}

	setID, gutesSet := kennung(an, "die Setkennung", rest[0])
	karteID, guteKarte := kennung(an, "die Kartenkennung", rest[1])

	bereich, guterBereich := zahlAus(an, "der Bereich", rest[3])

	weg, guterWeg := wegAus(an, rest[4])

	if !gutesSet || !guteKarte || !guterBereich || !guterWeg {
		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		naht, err := app.WortlistennahtUeber(db, domain.UrheberKommandozeile)
		if err != nil {
			an.klageUeber(err)

			return CodeFachlich
		}

		if wahlfehler := naht.Waehlen(
			ctx, setID, karteID, rest[2], bereich, weg); wahlfehler != nil {
			an.klageUeber(wahlfehler)

			return CodeFachlich
		}

		an.sag("Der Eintrag steht im Feld %s.\n", rest[2])

		return an.abschluss()
	})
}

// wegAus liest einen Weg der Form `2.0`.
func wegAus(an *schreiber, text string) ([]int, bool) {
	var weg []int

	for _, teil := range strings.Split(text, ".") {
		schritt, err := strconv.Atoi(teil)
		if err != nil || schritt < 0 {
			an.klage("Der Weg %q ist keine Folge von Nummern, etwa »3« oder »2.0«.", text)

			return nil, false
		}

		weg = append(weg, schritt)
	}

	return weg, true
}

// zahlAus liest eine nichtnegative Zahl.
func zahlAus(an *schreiber, was, text string) (int, bool) {
	zahl, err := strconv.Atoi(text)
	if err != nil || zahl < 0 {
		an.klage("%s ist keine Nummer: %q", was, text)

		return 0, false
	}

	return zahl, true
}
