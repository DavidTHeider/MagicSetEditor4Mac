package service

// **Eine Einfuhr zeigt vorher, was sie zusammenführt** — DOM-03.AK-01
// (#1698).
//
// # Der Befund
//
// Eine Zeile, die es im Bestand schon gab, wurde beim Einlesen **still
// erhöht** — richtig für eine zweite Liste, falsch für die eigene Ausfuhr,
// die jemand ein zweites Mal einliest: danach stand jede Menge doppelt da,
// und der Bericht sagte es erst hinterher.
//
// # Was jetzt gilt
//
// **Die Vorschau schreibt nichts** und sagt je Zeile, was geschähe: neu,
// auf eine vorhandene Zeile (vorher und nachher) oder auf eine frühere Zeile
// derselben Einfuhr. **Wer einliest, wählt die Zusammenführung** — addieren,
// wie bisher und als Vorgabe, oder Vorhandene auslassen.
//
// **Marke `new`** (§15): MSE2 führt keine Sammlung — gesucht am 2026-09-27
// unter `/Users/david/mse2-quelle/src` (Fassung f9d9356 vom 2021-07-06), ein
// Treffer für `collection` meint dort Skriptsammlungen.

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Zusammenfuehrung sagt, was mit einer Zeile geschieht, die der Bestand schon
// führt — gleicher Druck, Zustand, Sprache, Ausführung und Absicht.
type Zusammenfuehrung string

const (
	// ZusammenAddieren erhöht die vorhandene Zeile — die Vorgabe, und das
	// Verhalten seit #1292.
	ZusammenAddieren Zusammenfuehrung = "addieren"

	// ZusammenAuslassen lässt sie, wie sie ist.
	//
	// **Nur, was vor dieser Einfuhr da war**: eine Zeile, die dieselbe
	// Einfuhr eben angelegt hat, wächst weiter — zweimal dieselbe Karte in
	// einer Liste sind zwei Stück, keine Dublette mit dem Bestand.
	ZusammenAuslassen Zusammenfuehrung = "auslassen"
)

// geprueft gibt die Vorgabe für eine leere Angabe und weist eine unbekannte ab.
func (z Zusammenfuehrung) geprueft() (Zusammenfuehrung, error) {
	switch z {
	case "":
		return ZusammenAddieren, nil
	case ZusammenAddieren, ZusammenAuslassen:
		return z, nil
	}

	return "", domain.NewFieldError("zusammenfuehrung", domain.SchluesselValidierung,
		"Die Zusammenführung %q gibt es nicht — addieren oder auslassen.", string(z))
}

// Einfuhrart sagt, was mit einer Zeile der Einfuhr geschähe.
type Einfuhrart string

const (
	// EinfuhrNeu legt eine Zeile an.
	EinfuhrNeu Einfuhrart = "neu"

	// EinfuhrZusammen trifft eine Zeile, die der Bestand schon führt.
	EinfuhrZusammen Einfuhrart = "zusammen"

	// EinfuhrWiederholt trifft eine frühere Zeile derselben Einfuhr.
	EinfuhrWiederholt Einfuhrart = "wiederholt"
)

// Vorschauposten ist eine Zeile der Vorschau.
type Vorschauposten struct {
	// Zeile ist die Stelle in der Eingabe, wie ein Befund sie nennt — beim
	// CSV zählt die Kopfzeile mit.
	Zeile int

	Exemplar domain.Exemplar
	Art      Einfuhrart

	// Vorher und Nachher sind die Mengen der getroffenen Zeile **beim
	// Addieren** — vor und nach diesem Posten. Eine neue beginnt bei null.
	Vorher  int
	Nachher int
}

// Sammelvorschau ist, was eine Einfuhr täte — und was sie nicht lesen konnte.
type Sammelvorschau struct {
	Posten  []Vorschauposten
	Befunde []domain.Befund
}

// Vorschauen sagt je Zeile einer Liste, was das Einlesen täte — **ohne zu
// schreiben**.
//
// **Eine unbrauchbare Zeile ist ein Befund mit ihrer Stelle**, wie beim
// Einlesen; die übrigen stehen trotzdem da.
func (l *Listeneinfuhr) Vorschauen(ctx context.Context, zeilen []Listenzeile) (
	Sammelvorschau, error,
) {
	vorschau := Sammelvorschau{Posten: []Vorschauposten{}, Befunde: []domain.Befund{}}

	for stelle, zeile := range zeilen {
		exemplar := zeile.alsExemplar()

		if fehler := vorbereiten(exemplar); fehler != nil {
			vorschau.Befunde = append(vorschau.Befunde, domain.Befund{
				Ort: fmt.Sprintf("Zeile %d", stelle+1), Meldung: fehler.Error(), Art: domain.ArtEingabe,
			})

			continue
		}

		if fehler := l.einordnen(ctx, &vorschau, stelle+1, exemplar); fehler != nil {
			return Sammelvorschau{}, fehler
		}
	}

	return vorschau, nil
}

// einordnen trägt einen Posten ein — auf eine frühere Zeile derselben
// Einfuhr, auf eine vorhandene des Bestands oder neu.
//
// **Die frühere Zeile zuerst**: sie steht beim Einlesen schon da, wenn diese
// kommt, und ihr Nachher ist dieses Vorher.
func (l *Listeneinfuhr) einordnen(
	ctx context.Context, vorschau *Sammelvorschau, zeile int, exemplar *domain.Exemplar,
) error {
	posten := Vorschauposten{Zeile: zeile, Exemplar: *exemplar, Art: EinfuhrNeu}

	for frueher := len(vorschau.Posten) - 1; frueher >= 0; frueher-- {
		if gleicheZeile(&vorschau.Posten[frueher].Exemplar, exemplar) {
			posten.Art, posten.Vorher = EinfuhrWiederholt, vorschau.Posten[frueher].Nachher
			posten.Nachher = posten.Vorher + exemplar.Menge
			vorschau.Posten = append(vorschau.Posten, posten)

			return nil
		}
	}

	vorhanden, err := l.vorhandene(ctx, exemplar)
	if err != nil {
		return err
	}

	for _, schon := range vorhanden {
		if gleicheZeile(schon, exemplar) {
			posten.Art, posten.Vorher = EinfuhrZusammen, schon.Menge

			break
		}
	}

	posten.Nachher = posten.Vorher + exemplar.Menge
	vorschau.Posten = append(vorschau.Posten, posten)

	return nil
}
