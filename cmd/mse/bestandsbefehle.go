// Die Befehle `statistik`, `zuruecknehmen`, `wiederherstellen` und
// `einstellungen`
// (#1107).
//
// **Vier Fähigkeiten, die nur den Bestand brauchen** — kein Paket, kein
// Zuhause, kein Fenster.

package main

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// statistikAntwort trägt die Kennzahlen eines Sets.
type statistikAntwort struct {
	Statistik app.StatistikDTO `json:"statistik"`
}

// verteilungAntwort trägt die Verteilung eines Felds.
type verteilungAntwort struct {
	Verteilung []app.WertzahlDTO `json:"verteilung"`
}

// einstellungenAntwort trägt die Einstellungen samt erlaubten Werten.
type einstellungenAntwort struct {
	Einstellungen []app.EinstellungDTO `json:"einstellungen"`
}

// statistikbefehl zählt ein Set aus.
//
// **Mit wahlfreiem Feld**: ohne eines kommen die Kennzahlen, mit einem die
// Verteilung seiner Werte. Zwei Befehle wären zwei Namen für dieselbe Frage —
// „was steht in diesem Set?".
func statistikbefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) == 0 || len(rest) > 2 {
		an.klage("statistik braucht die Setkennung, wahlweise ein Feld.\n%s", hilfe)

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

		weg := app.ZahlennahtUeber(db)

		if len(rest) == 2 {
			return verteilungZeigen(ctx, an, weg, setID, rest[1])
		}

		return kennzahlenZeigen(ctx, an, weg, setID)
	})
}

// kennzahlenZeigen schreibt Kartenzahl, Feldzahl, leere Karten und Belegung.
func kennzahlenZeigen(ctx context.Context, an *schreiber, weg *app.Zahlenweg, setID int64) int {
	statistik, fehler := weg.Statistik(ctx, setID)
	if fehler != nil {
		an.klageUeber(fehler)

		return CodeFachlich
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, statistikAntwort{Statistik: statistik})
	}

	an.sag("Karten\t%d\nFelder\t%d\nLeere\t%d\n",
		statistik.Karten, statistik.Felder, statistik.Leere)

	for _, feld := range statistik.Belegung {
		an.sag("%s\t%d\n", feld.Name, feld.Karten)
	}

	return an.abschluss()
}

// verteilungZeigen schreibt, wie viele Karten welchen Wert tragen.
func verteilungZeigen(
	ctx context.Context, an *schreiber, weg *app.Zahlenweg, setID int64, feld string,
) int {
	verteilung, fehler := weg.Verteilung(ctx, setID, feld)
	if fehler != nil {
		an.klageUeber(fehler)

		return CodeFachlich
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, verteilungAntwort{Verteilung: verteilung})
	}

	if len(verteilung) == 0 {
		an.sag("Keine Karte trägt dieses Feld.\n")

		return an.abschluss()
	}

	for _, zeile := range verteilung {
		an.sag("%s\t%d\n", zeile.Name, zeile.Karten)
	}

	return an.abschluss()
}

// bewegungsbefehl nimmt eine Änderung zurück oder stellt sie wieder her.
//
// **Beide Richtungen über einen Weg**: sie unterscheiden sich nur im Aufruf,
// und zwei Kopien derselben Wache liefen auseinander.
func bewegungsbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string, zurueck bool,
) int {
	was := "wiederherstellen"
	if zurueck {
		was = "zuruecknehmen"
	}

	// **Ein Vorgang darf dahinter stehen** (#1393): ohne ihn der jüngste
	// Schritt des Sets, mit ihm genau dieser. Ein Skript, das eine
	// Massenänderung zurücknimmt, träfe sonst die Änderung, die ein anderer
	// Prozess inzwischen geschrieben hat.
	if len(rest) < 1 || len(rest) > 2 {
		an.klage("%s braucht eine Setkennung und darf einen Vorgang tragen.\n%s",
			was, hilfe)

		return CodeBedienung
	}

	setID, gut := kennung(an, "die Setkennung", rest[0])
	if !gut {
		return CodeBedienung
	}

	vorgang := ""
	if len(rest) == 2 {
		vorgang = rest[1]
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		return bewegen(ctx, an, app.VerlaufsnahtUeber(db), setID, vorgang, zurueck)
	})
}

// verlaufbefehl nennt die Schritte eines Sets — der jüngste zuerst, mit dem
// Vorgang, den `zuruecknehmen` nimmt (Werkstück U6, #1417).
func verlaufbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 1 {
		an.klage("verlauf braucht genau eine Setkennung.\n%s", hilfe)

		return CodeBedienung
	}

	setID, gut := kennung(an, "die Setkennung", rest[0])
	if !gut {
		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		schritte, fehler := app.VerlaufsnahtUeber(db).Verlauf(ctx, setID)
		if fehler != nil {
			an.klageUeber(fehler)

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, map[string]any{"schritte": schritte})
		}

		if len(schritte) == 0 {
			an.sag("Kein Verlauf.\n")

			return an.abschluss()
		}

		for _, schritt := range schritte {
			zurueck := ""
			if schritt.Zurueckgenommen {
				zurueck = " (zurückgenommen)"
			}

			an.sag("%s  %s  Karte %s  %s: %q → %q  %d Karte(n)  %s  Vorgang %s%s\n",
				schritt.Zeitpunkt, schritt.Handlung, schritt.KarteID, schritt.Feld,
				schritt.Vorher, schritt.Nachher, schritt.Umfang, schritt.Urheber,
				schritt.Vorgang, zurueck)
		}

		return an.abschluss()
	})
}

// bewegen führt die Bewegung aus und schreibt, was sie bewegt hat.
func bewegen(
	ctx context.Context, an *schreiber, weg *app.Verlaufsweg,
	setID int64, vorgang string, zurueck bool,
) int {
	richtung := weg.Wiederherstellen
	if zurueck {
		richtung = weg.Zuruecknehmen
	}

	bewegung, fehler := richtung(ctx, setID, vorgang)
	if fehler != nil {
		an.klageUeber(fehler)

		return CodeFachlich
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, bewegung)
	}

	// **„Nichts zu bewegen" ist kein Fehler.** Ein Skript, das bis zum Ende
	// zurücknimmt, liefe sonst in einen Rückgabewert, den es nicht von einem
	// echten Fehlschlag unterscheiden kann.
	if !bewegung.Bewegt {
		an.sag("Nichts zu bewegen.\n")

		return an.abschluss()
	}

	an.sag("Bewegt: Karte %s, Feld %s, %d Karte(n), Urheber %s.\n",
		bewegung.KarteID, bewegung.Feld, bewegung.Umfang, bewegung.Urheber)

	return an.abschluss()
}

// einstellungsbefehl liest oder schreibt eine Einstellung.
func einstellungsbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	// **Die drei Verben des Austauschs** — Werkstück O3 von Phase 5.2
	// (#1427).
	// Sie stehen in `einstellungsaustausch.go`; hier steht nur die Weiche.
	if len(rest) > 0 && rest[0] != "setzen" {
		return austauschbefehl(ctx, bestandsort, an, rest[0], rest[1:])
	}

	if len(rest) != 0 && len(rest) != 3 {
		an.klage("einstellungen nimmt nichts, »setzen <schluessel> <wert>«, "+
			"»ausfuhr [abschnitt…]«, »vorschau <datei>« oder "+
			"»einfuhr <datei> <revision> [ebene:schluessel…]«.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		weg := app.EinstellungsnahtUeber(db)

		if len(rest) == 3 {
			return getan(an, weg.Setzen(ctx, rest[1], rest[2]), "Gesetzt.\n")
		}

		return einstellungenZeigen(ctx, an, weg)
	})
}

// einstellungenZeigen schreibt jede Einstellung samt erlaubten Werten.
func einstellungenZeigen(ctx context.Context, an *schreiber, weg *app.Einstellungsweg) int {
	alle, fehler := weg.Alle(ctx)
	if fehler != nil {
		an.klageUeber(fehler)

		return CodeFachlich
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, einstellungenAntwort{Einstellungen: alle})
	}

	for _, einstellung := range alle {
		an.sag("%s\t%s\t%s\n", einstellung.Schluessel, einstellung.Wert, einstellung.Bereich)
	}

	return an.abschluss()
}
