// **Ein Feld auf vielen Karten setzen — an der Kommandozeile**
// (#1390).
//
// # Warum ein eigener Befehl und keine Schleife über `mse feld`
//
// **Weil die Atomarität die Fähigkeit ist.** Ein Ablauf, der ein Feld über
// zweihundert Karten ändert, macht sonst zweihundert Transaktionen; bricht er
// in der Mitte ab, ist das Set halb geändert, und niemand weiß, welche Hälfte.
// Dieser Weg schreibt in **einer**.
//
// # Warum es ihn bis heute nicht gab
//
// `docs/FAEHIGKEITEN.md` führte die Fähigkeit als *Viele Felder setzen* und
// begründete die Fensterbindung damit, das Fenster schreibe „die Eingaben
// **einer** Karte in einem Zug, weil eine Maske viele Felder hat" — und die
// Bindung setzt **ein** Feld auf **vielen** Karten. Die Begründung beschrieb
// eine andere Operation; gefunden bei der Prüfung der 24 Flächenlücken
// (Werkstück E von #1324).

package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// massenfeldbericht ist, was `--json` zurückgibt.
type massenfeldbericht struct {
	Geaendert int `json:"geaendert"`

	// Vorgang benennt diesen Schritt im Verlauf
	// (#1393).
	// `mse zuruecknehmen <set> <vorgang>` trifft damit genau ihn.
	Vorgang string `json:"vorgang"`
}

// felderSetzenBefehl setzt ein Feld auf mehreren Karten.
//
// **Die Kartenkennungen stehen am Ende und nicht am Anfang**, weil es beliebig
// viele sind: `mse felder-setzen 1 rarity common 7 8 9`. Ein Set, ein Feld,
// ein Wert — und dann die Liste.
func felderSetzenBefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) < 4 {
		an.klage("felder-setzen braucht Setkennung, Feldname, Wert und mindestens eine "+
			"Kartenkennung.\n%s", hilfe)

		return CodeBedienung
	}

	setID, gut := kennung(an, "die Setkennung", rest[0])
	if !gut {
		return CodeBedienung
	}

	karteIDs := make([]int64, 0, len(rest)-3)

	for _, roh := range rest[3:] {
		karteID, richtig := kennung(an, "eine Kartenkennung", roh)
		if !richtig {
			return CodeBedienung
		}

		karteIDs = append(karteIDs, karteID)
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		return felderSetzen(ctx, db, an, setID, karteIDs, rest[1], rest[2])
	})
}

// felderVorschauBefehl zeigt, was `felder-setzen` je Karte änderte — und
// schreibt nichts (Werkstück U6, #1417).
func felderVorschauBefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) < 4 {
		an.klage("felder-vorschau braucht Setkennung, Feldname, Wert und mindestens eine "+
			"Kartenkennung.\n%s", hilfe)

		return CodeBedienung
	}

	setID, gut := kennung(an, "die Setkennung", rest[0])
	if !gut {
		return CodeBedienung
	}

	karteIDs := make([]int64, 0, len(rest)-3)

	for _, roh := range rest[3:] {
		karteID, richtig := kennung(an, "eine Kartenkennung", roh)
		if !richtig {
			return CodeBedienung
		}

		karteIDs = append(karteIDs, karteID)
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		naht, err := app.Schreibnaht(db, domain.UrheberKommandozeile)
		if err != nil {
			an.klageUeber(err)

			return CodeFachlich
		}

		vorschau, fehler := naht.FelderVorschau(ctx, setID, karteIDs, rest[1], rest[2])
		if fehler != nil {
			an.klageUeber(fehler)

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, vorschau)
		}

		for _, zeile := range vorschau.Zeilen {
			name := zeile.Name
			if name == "" {
				name = "Karte " + strconv.FormatInt(zeile.KarteID, 10)
			}

			alt := zeile.Alt
			if !zeile.GabAlt {
				alt = "(kein Wert)"
			}

			if zeile.Ausschluss != "" {
				an.sag("%d  %s: %q bleibt — %s\n", zeile.KarteID, name, alt, zeile.Ausschluss)

				continue
			}

			an.sag("%d  %s: %q → %q\n", zeile.KarteID, name, alt, zeile.Neu)
		}

		return CodeGut
	})
}

// felderSetzen schreibt über dieselbe Naht wie Fenster und MCP-Server.
//
// **Nicht unmittelbar über `storage`**: dort ginge es an der Setgrenze und an
// der Vorlage vorbei — ein nicht editierbares oder berechnetes Feld liefe
// durch, und eine fremde Kartenkennung änderte eine Karte in einem Set, das
// niemand geöffnet hat
// ([0041](../../docs/entscheidungen/0041-die-setgrenze-liegt-im-dienst.md)).
func felderSetzen(
	ctx context.Context, db *sql.DB, an *schreiber,
	setID int64, karteIDs []int64, feld, wert string,
) int {
	naht, err := app.Schreibnaht(db, domain.UrheberKommandozeile)
	if err != nil {
		an.klageUeber(err)

		return CodeFachlich
	}

	var (
		vorgang string
		fehler  error
	)

	// **Mit Erwartung geht es den vergleichenden Weg** (U6, #1417): die
	// Zeilen aus `felder-vorschau --json`, und der Bestand vergleicht, bevor
	// er schreibt.
	if an.erwartet == "" {
		vorgang, fehler = naht.FelderSetzen(ctx, setID, karteIDs, feld, wert)
	} else {
		var erwartet []domain.Feldvorwert
		if err := json.Unmarshal([]byte(an.erwartet), &erwartet); err != nil {
			an.klage("--erwartet ist kein JSON aus felder-vorschau: %v", err)

			return CodeBedienung
		}

		vorgang, fehler = naht.FelderSetzenMitErwartung(ctx, setID, erwartet, feld, wert)
	}

	if fehler != nil {
		an.klageUeber(fehler)

		return CodeFachlich
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, massenfeldbericht{
			Geaendert: len(karteIDs), Vorgang: vorgang,
		})
	}

	// **Der Vorgang steht auch im Klartext.** Ein Mensch, der eine
	// Massenänderung gezielt zurücknehmen will, braucht ihn genauso wie ein
	// Skript — und ohne ihn hier müsste er `--json` fahren.
	an.sag("Gesetzt auf %d Karten. Vorgang: %s\n", len(karteIDs), vorgang)

	return an.abschluss()
}
