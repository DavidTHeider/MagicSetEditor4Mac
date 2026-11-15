package app

// **Die Validierung, die keinen eigenen Schlüssel hat** —
// #1601.
//
// # Eine eigene Datei, und nicht nur wegen der Zeilengrenze
//
// **Weil sie eine andere Frage beantwortet** (§2): nebenan steht, *welche
// Fehlerarten es gibt* und wie sie eingeordnet werden. Hier steht, wie eine
// Ablehnung **aussieht**, für die niemand einen Schlüssel vergeben hat — und
// das ist eine Frage an die Oberfläche, nicht an die Einordnung.

import (
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// validierungsgrund gibt die Validierung, die keinen eigenen Schlüssel hat.
//
// **Der Satz aus dem Fachkern reist als Wert mit**, damit der Katalog ihn in
// seinen eigenen Rahmen setzen kann — der Rahmen bleibt übersetzt, die
// Auskunft bleibt erhalten. Warum es dafür einen eigenen Schlüssel braucht,
// steht bei [domain.SchluesselValidierungsgrund].
func validierungsgrund(fehler error) *Fehler {
	return &Fehler{
		Code:       CodeValidierung,
		Schluessel: domain.SchluesselValidierungsgrund,
		Meldung:    fehler.Error(),
		Werte:      map[string]string{"grund": ohneMarke(fehler)},
	}
}

// ohneMarke schneidet den Wortlaut von [domain.ErrValidation] heraus.
//
// # Warum das kein Vergleich über Zeichenketten ist
//
// **Weil hier nichts entschieden wird** (§8). Die Art steht schon fest —
// `errors.Is` hat sie erkannt, und nur deshalb ist dieser Weg erreicht. Was
// hier geschieht, ist **Darstellung**: „Validierung fehlgeschlagen: " ist die
// Marke, an der die Maschine erkennt, und für einen Menschen ein Wort zuviel
// vor dem Satz, der ihm sagt, was zu tun ist.
//
// # Und sie steht selten vorn
//
// **Hier stand, wo sie nicht vorn stehe, bleibe alles stehen** — als
// Entscheidung, und ohne die Messung dazu. Gemessen am 2026-09-17: **56**
// Stellen setzen die Marke nach vorn, **190** umschließen einen Fehler mit
// ihrem eigenen Vorsatz. Am Fenster stand deshalb „der Sitz »sitz1«:
// Validierung fehlgeschlagen: ein Deckabzug ohne Karten ist kein Deck" —
// die Marke mitten im Satz, an den Nutzende ausgeliefert.
//
// **Eine Marke ist an jeder Stelle eine Marke.** Sie steht da, damit
// `errors.Is` greift; für einen Menschen ist sie überall ein Wort zuviel.
//
// **`ReplaceAll` und nicht `Replace` mit Anzahl eins**: zwei umschlossene
// Validierungen tragen sie zweimal, und eine halb entfernte Marke wäre die
// schlechteste der drei Möglichkeiten.
func ohneMarke(fehler error) string {
	marke := domain.ErrValidation.Error() + ": "

	return strings.ReplaceAll(fehler.Error(), marke, "")
}
