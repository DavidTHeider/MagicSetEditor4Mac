package mtgjson

import (
	"bytes"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Hinweis ist ein Befund ohne seine Stelle — die kennt erst der Leser.
//
// Getrennt, weil die Zuordnung eine einzelne Karte sieht und nicht weiß, aus
// welcher Datei und welchem Set sie stammt. Wer ihr das mitgäbe, reichte zwei
// Angaben durch die ganze Abbildung, nur damit am Ende eine Meldung vollständig
// ist.
type Hinweis struct {
	Feld    string
	Meldung string

	// JeKarte sagt, dass dieser Hinweis **nicht** über die Karten eines Sets
	// gebündelt werden darf.
	//
	// Der Regelfall ist das Gegenteil: ein unbekanntes Feld führen fast alle
	// Karten, und je Karte gemeldet ergäbe das für eine echte MTGJSON-Datei
	// Tausende gleichlautender Zeilen (siehe `buendeln_test.go`). Eine
	// Namenskollision trägt dagegen den **Wert** in ihrer Meldung — sie zu
	// bündeln hieße, alle Werte bis auf einen wegzuwerfen.
	JeKarte bool
}

// Karte bildet eine rohe MTG-JSON-Karte auf Kartenfelder ab.
//
// Die Reihenfolge der Werte ist die der Zuordnung und damit stabil; die des
// Rests ist alphabetisch. Beides ist Absicht: eine Abbildung in Go hat keine
// Reihenfolge, und ein Rundlauf, dessen Reihenfolge bei jedem Lauf anders ist,
// ließe sich nicht vergleichen.
//
// # Sie scheitert nicht
//
// **Seit dem 2026-08-24 gibt sie keinen Fehler mehr zurück** (#888). Was sich
// nicht lesen lässt, wird ein Hinweis und kostet sein Feld — §8. Der frühere
// Rückgabewert ließ **zwei** Fehlerzweige bei den Aufrufern stehen, die kein
// Aufruf mehr betreten konnte; sie sind mit entfernt.
func (z Zuordnung) Karte(roh map[string]json.RawMessage) (*Karte, []Hinweis) {
	karte := &Karte{}

	belegt := make(map[string]bool, len(z.Felder))

	var hinweise []Hinweis

	for _, paar := range z.Felder {
		if paar.Feld == "" {
			continue
		}

		werte, vorhanden := paar.werteAus(roh)
		if !vorhanden {
			continue
		}

		text, err := paar.wandeln(werte)
		if err != nil {
			// **Ein unlesbarer Wert kostet sein Feld, nicht die Datei**
			// (#888).
			//
			// Ein Abbruch hier bräche an **einer** Zahl in **einem** Feld
			// **einer** Karte den Import eines ganzen Sets ab — genau der Fall,
			// den AGENTS.md §8 ausschließt: „Ein Set, von dem 300 Karten lesbar
			// sind und zwei nicht, wird mit 300 Karten und zwei benannten
			// Befunden geöffnet, nicht abgewiesen."
			//
			// **JeKarte, nicht gebündelt**: die Meldung trägt den Grund des
			// Einzelfalls, und den bei mehreren Karten wegzuwerfen ließe
			// niemanden nachsehen.
			hinweise = append(hinweise, Hinweis{
				Feld: paar.Feld, JeKarte: true,
				Meldung: fmt.Sprintf(
					"Feld %q ließ sich nicht lesen und bleibt leer: %v", paar.Feld, err),
			})

			continue
		}

		// **Ein leerer Wert wird nicht angelegt.** MTG JSON führt leere Listen
		// (`"subtypes": []`) für Karten, die keine Untertypen haben; daraus ein
		// leeres Kartenfeld zu machen hieße, eine Angabe zu erfinden, die die
		// Datei nicht macht.
		if text == "" {
			continue
		}

		belegt[paar.Feld] = true

		karte.Werte = append(karte.Werte, domain.Value{Field: paar.Feld, Text: text})
	}

	return karte, append(hinweise, z.rest(karte, roh, belegt)...)
}

// rest sammelt die Schlüssel, die die Zuordnung nicht kennt.
func (z Zuordnung) rest(karte *Karte, roh map[string]json.RawMessage, belegt map[string]bool) []Hinweis {
	unbekannt := make([]string, 0, len(roh))

	for schluessel := range roh {
		if _, kennt := z.Feld(schluessel); !kennt {
			unbekannt = append(unbekannt, schluessel)
		}
	}

	slices.Sort(unbekannt)

	hinweise := make([]Hinweis, 0, len(unbekannt))

	for _, schluessel := range unbekannt {
		// **Ein Schlüssel, der heißt wie ein belegtes Kartenfeld, wird nicht
		// mitgeführt.** Zwei Werte für dasselbe Feld sind an einer Karte nicht
		// darstellbar — `card_values` hat den Primärschlüssel `(card_id, feld)`,
		// und der Import scheiterte mit einem Konflikt, den niemand einer
		// einzelnen Zeile der Datei zuordnen könnte. Der Wert geht deshalb in
		// die Meldung: benannt ist er nicht verloren.
		if belegt[schluessel] {
			hinweise = append(hinweise, Hinweis{
				Feld:    schluessel,
				JeKarte: true,
				Meldung: fmt.Sprintf(
					"die Zuordnung für »%s« kennt dieses Feld nicht, und mitführen geht nicht: das Kartenfeld heißt genauso. Der Wert war %s",
					z.Spiel, string(roh[schluessel])),
			})

			continue
		}

		karte.Rest = append(karte.Rest, domain.Value{Field: schluessel, Text: eng(roh[schluessel])})

		hinweise = append(hinweise, Hinweis{
			Feld:    schluessel,
			Meldung: fmt.Sprintf("die Zuordnung für »%s« kennt dieses Feld nicht", z.Spiel),
		})
	}

	return hinweise
}

// eng legt einen rohen JSON-Wert in seiner knappen Form ab.
//
// **Normalisiert und nicht Byte für Byte** — anders als bei `.mse-set`, und der
// Unterschied hat einen Grund. Dort ist die Datei die Quelle und muss unverändert
// wiederkommen; hier ist sie es nicht, und ein Wert, dessen Text von der
// Einrückung der Quelldatei abhängt, ließe den Rundlauf an einer Formatierung
// scheitern statt an einem Inhalt.
//
// Gefunden hat das der Lauf gegen bezogene Daten: die Quelle ist knapp
// geschrieben, unsere Ausgabe eingerückt — und `"colors": ["W"]` kam mit
// Zeilenumbrüchen und Einrückung zurück.
func eng(roh json.RawMessage) string {
	var knapp bytes.Buffer

	// **Kann nicht scheitern**: `roh` stammt aus einem geglückten `Unmarshal`
	// und ist damit gültiges JSON. Ein Fehlerzweig, den kein Aufruf betreten
	// kann, wäre eine Behauptung darüber, was im Ernstfall geschieht (§3).
	_ = json.Compact(&knapp, roh) //nolint:errcheck // siehe darüber: kann nicht scheitern

	return knapp.String()
}

// werteAus holt die rohen Werte eines Paars; der zweite Rückgabewert sagt, ob
// die Karte wenigstens einen der Schlüssel führt.
func (p Feldpaar) werteAus(roh map[string]json.RawMessage) ([]json.RawMessage, bool) {
	werte := make([]json.RawMessage, len(p.JSON))
	vorhanden := false

	for i, schluessel := range p.JSON {
		if wert, da := roh[schluessel]; da {
			werte[i] = wert
			vorhanden = true
		}
	}

	return werte, vorhanden
}

// wandeln wendet die Wandlung eines Paars an; ohne eigene die wörtliche.
func (p Feldpaar) wandeln(werte []json.RawMessage) (string, error) {
	if p.Wandeln == nil {
		return ersteZeichenkette(werte)
	}

	return p.Wandeln(werte)
}
