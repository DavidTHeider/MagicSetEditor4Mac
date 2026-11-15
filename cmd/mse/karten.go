// Die Befehle `karten` und `feld` — lesen und schreiben
// (#886).

package main

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// karteDTO ist eine Karte, wie die Kommandozeile sie ausgibt.
//
// **Der Name eigens**, obwohl er auch in `Werte` steht: er ist die eine
// Angabe, nach der jemand eine Karte sucht, und ein Skript soll ihn nicht aus
// einer Liste klauben müssen.
type karteDTO struct {
	ID     string            `json:"id"`
	Nummer int               `json:"nummer"`
	Name   string            `json:"name"`
	Werte  map[string]string `json:"werte"`
}

// kartenAntwort ist die Hülle — mit einer Liste, die nie `null` ist.
type kartenAntwort struct {
	Karten []karteDTO `json:"karten"`
}

// feldAntwort sagt, was ein Schreiben ersetzt hat.
//
// **Der vorige Wert reist mit**, damit ein Skript zurückkann — dieselbe
// Zusage wie beim MCP-Server
// ([0010](../../docs/entscheidungen/0010-mcp-als-eigener-prozess.md)).
// `Hatte` unterscheidet „war leer" von „gab es nicht": eine leere Zeile ist
// nicht dasselbe wie keine Zeile.
type feldAntwort struct {
	Vorher string `json:"vorher"`
	Hatte  bool   `json:"hatte"`
}

// kartenAuflisten gibt die Karten eines Sets.
//
// # Warum zuerst gefragt wird, ob es das Set gibt
//
// `storage` gibt für ein unbekanntes Set eine **leere Liste** und keinen
// Fehler — im Fenster ist das richtig, denn dort kann niemand ein Set wählen,
// das es nicht gibt. Auf der Kommandozeile kann er es: `mse karten 4711` ist
// ein Tippfehler, und „keine Karte" wäre darauf eine irreführende Antwort. Ein
// Skript schlösse daraus, das Set sei leer.
func kartenAuflisten(ctx context.Context, db *sql.DB, an *schreiber, setID int64) int {
	leser := storage.NeuerSetLeser(db)

	koepfe, err := leser.Setkoepfe(ctx)
	if err != nil {
		an.klageUeber(err)

		return CodeFachlich
	}

	if !kenntSet(koepfe, setID) {
		// **Über den Domänenfehler und nicht als eigener Satz** — Befund 3
		// des Nachlaufs über Phase 3.8
		// (#1253).
		//
		// Ein eigener Satz über `an.klage` trüge seine **Klasse** nicht: der
		// Dienst antwortet `"code":"nichtGefunden"`, und ein Lauf, der „überall
		// gleich" zusagt, könnte das an der Kommandozeile nicht prüfen.
		an.klageUeber(fmt.Errorf("%w: das Set %d gibt es nicht",
			domain.ErrNotFound, setID))

		return CodeFachlich
	}

	zeilen, err := leser.Karten(ctx, setID)
	if err != nil {
		an.klageUeber(err)

		return CodeFachlich
	}

	antwort := kartenAntwort{Karten: make([]karteDTO, 0, len(zeilen))}

	for _, zeile := range zeilen {
		werte := make(map[string]string, len(zeile.Values))
		for _, wert := range zeile.Values {
			werte[wert.Field] = wert.Text
		}

		antwort.Karten = append(antwort.Karten, karteDTO{
			ID:     strconv.FormatInt(zeile.ID, 10),
			Nummer: zeile.Number,
			Name:   werte["name"],
			Werte:  werte,
		})
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, antwort)
	}

	if len(antwort.Karten) == 0 {
		an.sag("Keine Karte in diesem Set.\n")

		return an.abschluss()
	}

	for _, karte := range antwort.Karten {
		an.sag("%s\t%d\t%s\n", karte.ID, karte.Nummer, karte.Name)
	}

	return an.abschluss()
}

// feldSetzen schreibt einen Feldwert und nennt den vorigen.
func feldSetzen(
	ctx context.Context, db *sql.DB, an *schreiber, setID, karteID int64, feld, wert string,
) int {
	// **Dieselbe fachliche Naht wie Fenster und MCP-Server.** Wer hier
	// `storage.NeuerKartenschreiber` unmittelbar nähme, bekäme einen rohen
	// `text`-Wert, den kein Paket liest — genau der Befund aus #710.
	naht, err := app.Schreibnaht(db, domain.UrheberKommandozeile)
	if err != nil {
		an.klageUeber(err)

		return CodeFachlich
	}

	vorher, hatte, fehler := naht.FeldSetzenMitVorwert(ctx, setID, karteID, feld, wert)
	if fehler != nil {
		an.klageUeber(fehler)

		return CodeFachlich
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, feldAntwort{Vorher: vorher, Hatte: hatte})
	}

	if !hatte {
		an.sag("Gesetzt. Das Feld gab es vorher nicht.\n")

		return an.abschluss()
	}

	an.sag("Gesetzt. Vorher: %s\n", vorher)

	return an.abschluss()
}

// kennung liest eine Kennung aus einem Argument.
//
// **Ein Bedienfehler und kein fachlicher**: `mse karten keinezahl` ist ein
// falscher Aufruf, kein Set, das es nicht gibt.
func kennung(an *schreiber, was, roh string) (int64, bool) {
	zahl, err := strconv.ParseInt(roh, 10, 64)
	if err != nil {
		an.klage("%s muss eine Zahl sein, nicht %q.\n", was, roh)

		return 0, false
	}

	return zahl, true
}

// kenntSet sagt, ob eine Kennung zu einem Set im Bestand gehört.
func kenntSet(koepfe []domain.SetHead, setID int64) bool {
	for _, kopf := range koepfe {
		if kopf.ID == setID {
			return true
		}
	}

	return false
}
