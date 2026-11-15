// Die Befehle `karte`, `suche` und `felder` — der Lebenszyklus einer Karte und
// der Volltext auf der Kommandozeile
// (#1107).
//
// **Fünf Fähigkeiten, die der MCP-Server längst hatte.** Die fachliche Naht
// ist in beiden Fällen dieselbe — `app.Schreibnaht` für das Schreiben,
// `storage.NeuerSetLeser` für die Suche —; gefehlt hat allein, dass jemand sie
// von hier aus ruft. Genau das ist die leere Zelle, die #1107 zählt: keine
// Entscheidung gegen die Fähigkeit, sondern gar keine.

package main

import (
	"context"
	"database/sql"
	"strconv"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// karteAntwort nennt die Kennung einer angelegten oder kopierten Karte.
//
// **Als Zeichenkette wie überall sonst**: eine Kartenkennung wird in JSON
// sonst zu einer Fließkommazahl, sobald sie groß genug wird.
type karteAntwort struct {
	KarteID string `json:"karteID"`
}

// trefferDTO ist eine gefundene Stelle.
type trefferDTO struct {
	KarteID string `json:"karteID"`
	SetID   string `json:"setID"`
	Nummer  int    `json:"nummer"`
	Feld    string `json:"feld"`
	Wert    string `json:"wert"`
}

// suchAntwort trägt die Treffer — und ob gekürzt wurde.
//
// **Gekuerzt reist mit** (§8, und dieselbe Zusage wie im Fenster): eine stille
// Kürzung sieht aus wie ein vollständiges Ergebnis, und ein Skript, das daraus
// „mehr gibt es nicht" schließt, irrt.
type suchAntwort struct {
	Treffer  []trefferDTO `json:"treffer"`
	Gekuerzt bool         `json:"gekuerzt"`
}

// felderAntwort trägt die Feldnamen.
type felderAntwort struct {
	Felder []string `json:"felder"`
}

// setvorsatz ist der Vorsatz, an dem `suche` sein Set erkennt.
//
// **Im selben Wort wie `--guete=`** und aus demselben Grund: ein Flag mit
// nachfolgendem Wert bräuchte einen Vorgriff in `zerlege`.
const setvorsatz = "--set="

// kartenbefehl prüft die Argumente von `karte` und führt den Unterbefehl aus.
//
// **Mit Unterbefehl und nicht als vier eigene Befehle** — wie `schluesselwort`.
// Die vier gehören zusammen, und `mse karte 3 loeschen 7` liest sich als das,
// was es ist: eine Änderung an Set 3.
func kartenbefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) < 2 {
		an.klage("karte braucht Setkennung und einen Unterbefehl.\n%s", hilfe)

		return CodeBedienung
	}

	setID, gut := kennung(an, "die Setkennung", rest[0])
	if !gut {
		return CodeBedienung
	}

	unterbefehl, weitere := rest[1], rest[2:]

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		return karteAusfuehren(ctx, db, an, setID, unterbefehl, weitere)
	})
}

// karteAusfuehren wählt den Unterbefehl.
func karteAusfuehren(
	ctx context.Context, db *sql.DB, an *schreiber, setID int64,
	unterbefehl string, weitere []string,
) int {
	naht, err := app.Schreibnaht(db, domain.UrheberKommandozeile)
	if err != nil {
		an.klageUeber(err)

		return CodeFachlich
	}

	switch unterbefehl {
	case "anlegen":
		if len(weitere) != 0 {
			an.klage("karte anlegen nimmt keine weiteren Argumente.\n%s", hilfe)

			return CodeBedienung
		}

		return angelegt(an, func() (int64, error) { return naht.KarteAnlegen(ctx, setID) })

	case "kopieren":
		karteID, gut := eineKarte(an, weitere, "karte kopieren")
		if !gut {
			return CodeBedienung
		}

		return angelegt(an, func() (int64, error) {
			return naht.KarteKopieren(ctx, setID, karteID)
		})

	case "loeschen":
		return loeschbefehl(ctx, an, naht, setID, weitere)

	case "tauschen":
		return tauschbefehl(ctx, an, naht, setID, weitere)

	default:
		an.klage("unbekannter Unterbefehl %q von karte.\n%s", unterbefehl, hilfe)

		return CodeBedienung
	}
}

// tauschbefehl liest die beiden Kennungen und tauscht.
func tauschbefehl(
	ctx context.Context, an *schreiber, naht app.Kartenschreibnaht,
	setID int64, weitere []string,
) int {
	if len(weitere) != 2 {
		an.klage("karte tauschen braucht zwei Kartenkennungen.\n%s", hilfe)

		return CodeBedienung
	}

	eine, guteEine := kennung(an, "die erste Kartenkennung", weitere[0])
	andere, guteAndere := kennung(an, "die zweite Kartenkennung", weitere[1])

	if !guteEine || !guteAndere {
		return CodeBedienung
	}

	return getan(an, naht.KartenTauschen(ctx, setID, eine, andere), "Getauscht.\n")
}

// loeschbefehl nimmt eine Karte aus dem Set.
//
// **`--trotzdem` bestätigt ein Löschen, das einen Deckplatz mitnimmt**
// (#1369).
// Ohne die Angabe weist der Dienst eine Karte ab, die in einem Deck liegt, und
// nennt die Decks — dieselbe Auskunft wie an jeder anderen Fläche (§16).
func loeschbefehl(
	ctx context.Context, an *schreiber, naht app.Kartenschreibnaht,
	setID int64, weitere []string,
) int {
	trotzdem := false
	blank := make([]string, 0, len(weitere))

	for _, wort := range weitere {
		if wort == "--trotzdem" {
			trotzdem = true

			continue
		}

		blank = append(blank, wort)
	}

	karteID, gut := eineKarte(an, blank, "karte loeschen")
	if !gut {
		return CodeBedienung
	}

	return getan(an, naht.KarteLoeschen(ctx, setID, karteID, trotzdem), "Gelöscht.\n")
}

// eineKarte liest die eine Kartenkennung, die ein Unterbefehl braucht.
func eineKarte(an *schreiber, weitere []string, was string) (int64, bool) {
	if len(weitere) != 1 {
		an.klage("%s braucht genau eine Kartenkennung.\n%s", was, hilfe)

		return 0, false
	}

	return kennung(an, "die Kartenkennung", weitere[0])
}

// angelegt führt eine Anlage aus und nennt die neue Kennung.
func angelegt(an *schreiber, arbeit func() (int64, error)) int {
	neue, fehler := arbeit()
	if fehler != nil {
		an.klageUeber(fehler)

		return CodeFachlich
	}

	kennung := strconv.FormatInt(neue, 10)

	if an.alsJSON {
		return alsJSONSchreiben(an, karteAntwort{KarteID: kennung})
	}

	an.sag("Angelegt als Karte %s.\n", kennung)

	return an.abschluss()
}

// getan meldet einen Vorgang ohne Rückgabewert.
func getan(an *schreiber, fehler error, satz string) int {
	if fehler != nil {
		an.klageUeber(fehler)

		return CodeFachlich
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, struct{}{})
	}

	an.sag("%s", satz)

	return an.abschluss()
}

// suchbefehl sucht im Volltext.
//
// **Ohne Set ist die Suche bestandsweit**, wie im Fenster: `--set=` schränkt
// ein. Ein Pflichtargument machte den häufigen Fall — „wo steht das
// überhaupt?" — zum umständlichen.
func suchbefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	text, setroh := "", ""

	for _, argument := range rest {
		if strings.HasPrefix(argument, setvorsatz) {
			setroh = strings.TrimPrefix(argument, setvorsatz)

			continue
		}

		if text != "" {
			an.klage("suche nimmt genau einen Suchtext.\n%s", hilfe)

			return CodeBedienung
		}

		text = argument
	}

	if text == "" {
		an.klage("suche braucht einen Suchtext.\n%s", hilfe)

		return CodeBedienung
	}

	var setID int64

	if setroh != "" {
		gelesen, gut := kennung(an, "die Setkennung", setroh)
		if !gut {
			return CodeBedienung
		}

		setID = gelesen
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		return suchen(ctx, db, an, domain.SearchQuery{Text: text, SetID: setID})
	})
}

// suchen führt die Suche aus und schreibt die Treffer.
func suchen(ctx context.Context, db *sql.DB, an *schreiber, frage domain.SearchQuery) int {
	// **Eine Setkennung, die auf nichts zeigt, ist ein Fehler und kein leeres
	// Ergebnis** (#1336).
	//
	// Das Issue stellt die Frage ausdrücklich: ist `--set=` ein Filter, der
	// nichts trifft, oder eine Kennung, die keine ist? Sie ist eine Kennung —
	// wer eine Zahl angibt, meint ein bestimmtes Set, und ein Tippfehler darin
	// ist kein leeres Ergebnis.
	if frage.SetID != 0 {
		if fehler := app.SetDaSein(ctx, db, frage.SetID); fehler != nil {
			an.klageUeber(fehler)

			return CodeFachlich
		}
	}

	ergebnis, fehler := storage.NeuerSetLeser(db).Suche(ctx, frage)
	if fehler != nil {
		an.klageUeber(fehler)

		return CodeFachlich
	}

	antwort := suchAntwort{
		Treffer:  make([]trefferDTO, 0, len(ergebnis.Hits)),
		Gekuerzt: ergebnis.Truncated,
	}

	for _, treffer := range ergebnis.Hits {
		antwort.Treffer = append(antwort.Treffer, trefferDTO{
			KarteID: strconv.FormatInt(treffer.CardID, 10),
			SetID:   strconv.FormatInt(treffer.SetID, 10),
			Nummer:  treffer.Number,
			Feld:    treffer.Field,
			Wert:    treffer.Value,
		})
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, antwort)
	}

	if len(antwort.Treffer) == 0 {
		an.sag("Kein Treffer.\n")

		return an.abschluss()
	}

	for _, treffer := range antwort.Treffer {
		an.sag("%s\t%s\t%d\t%s\t%s\n",
			treffer.SetID, treffer.KarteID, treffer.Nummer, treffer.Feld, treffer.Wert)
	}

	if antwort.Gekuerzt {
		an.sag("Gekürzt — es gibt mehr.\n")
	}

	return an.abschluss()
}

// felderbefehl nennt die Feldnamen eines Sets.
func felderbefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) != 1 {
		an.klage("felder braucht genau eine Setkennung.\n%s", hilfe)

		return CodeBedienung
	}

	setID, gut := kennung(an, "die Setkennung", rest[0])
	if !gut {
		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		if !kenntSetIm(ctx, db, an, setID) {
			return CodeFachlich
		}

		return felderZeigen(ctx, db, an, setID)
	})
}

// felderZeigen schreibt die Feldnamen.
func felderZeigen(ctx context.Context, db *sql.DB, an *schreiber, setID int64) int {
	felder, fehler := storage.NeuerSetLeser(db).Felder(ctx, setID)
	if fehler != nil {
		an.klageUeber(fehler)

		return CodeFachlich
	}

	antwort := felderAntwort{Felder: make([]string, 0, len(felder))}
	antwort.Felder = append(antwort.Felder, felder...)

	if an.alsJSON {
		return alsJSONSchreiben(an, antwort)
	}

	if len(antwort.Felder) == 0 {
		an.sag("Keine Felder in diesem Set.\n")

		return an.abschluss()
	}

	for _, feld := range antwort.Felder {
		an.sag("%s\n", feld)
	}

	return an.abschluss()
}
