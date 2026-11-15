package service

import (
	"slices"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// mitErklaertenNamen gibt jedem Feldstil den Namen, unter dem sein Feld
// erklärt ist — #1980.
//
// MSE macht beide Seiten kanonisch: den Feldnamen in `Field::after_reading`
// (`src/data/field.cpp`, Zeile 68) und jeden Schlüssel beim Lesen
// (`src/util/io/reader.cpp`, Zeile 217; Fassung f9d9356 vom 2021-07-06).
// `card style: rule_text:` trifft dort das Feld `rule text`.
//
// **Hier und nicht beim Lesen**, und das ist eine Entscheidung: der
// Feldname geht als Schlüssel der Kartenwerte in den Bestand. Wer ihn beim
// Lesen umschriebe, verwaiste jeden Wert, den diese Anwendung unter einem
// Namen mit Unterstrich angelegt hat. An der Naht bekommt der **Stil** den
// Namen des **Spiels**, und jeder rohe Vergleich danach trägt — Feldart,
// Feldfläche fürs Fenster, Sichtseite einer Stiloption.
//
// Ein Schlüssel ohne erklärtes Feld bleibt, wie er ist.
func mitErklaertenNamen(stil msepaket.Stil, kartenfelder []domain.Field) msepaket.Stil {
	felder := erklaerteNamen(kartenfelder)
	for name, erklaert := range erklaerteNamen(stil.Zusatzfelder) {
		if _, schon := felder[name]; !schon {
			felder[name] = erklaert
		}
	}

	optionen := map[string]string{}
	for _, option := range stil.Stiloptionen {
		optionen[msepaket.Kanonisch(option.Name)] = option.Name
	}

	stil.Felder = umbenannt(stil.Felder, felder)
	stil.Zusatzstile = umbenannt(stil.Zusatzstile, felder)
	stil.Optionsstile = umbenannt(stil.Optionsstile, optionen)

	return stil
}

// erklaerteNamen ordnet der kanonischen Form den erklärten Namen zu — bei
// zwei Feldern derselben Form das erste.
func erklaerteNamen(felder []domain.Field) map[string]string {
	namen := make(map[string]string, len(felder))

	for _, feld := range felder {
		if _, schon := namen[msepaket.Kanonisch(feld.Name)]; !schon {
			namen[msepaket.Kanonisch(feld.Name)] = feld.Name
		}
	}

	return namen
}

// umbenannt gibt eine Kopie, in der jeder Feldstil den erklärten Namen trägt.
//
// **Eine Kopie**: das Stilmodell kann noch an anderer Stelle hängen, und
// ein Umbenennen am Ort änderte es dort mit.
func umbenannt(stile []msepaket.Feldstil, namen map[string]string) []msepaket.Feldstil {
	neu := slices.Clone(stile)

	for stelle := range neu {
		if erklaert, da := namen[msepaket.Kanonisch(neu[stelle].Feld)]; da {
			neu[stelle].Feld = erklaert
		}
	}

	return neu
}
