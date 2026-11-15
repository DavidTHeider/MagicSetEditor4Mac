package msepaket

// **Eine Feldart, die MSE2 nicht kennt** — FMT-02, S10.NAT-02 und Befund F12
// (#1691).
//
// MSE2 bricht das Lesen eines Pakets ab, wenn ein Feld keine oder eine
// unbekannte Art trägt (`read_new<Field>`, `src/data/field.cpp` Zeilen 74–94,
// Fassung f9d9356 vom 2021-07-06: „expected key type" und „unsupported field
// type"). Dieses Programm liest ein solches Feld als Text weiter
// ([feldarten]) — **und sagt es jetzt**: die Werkstatt speicherte die Datei bis
// hierher ohne ein Wort, und MSE2 öffnete sie danach nicht mehr.

import (
	"fmt"
	"slices"
	"strings"
)

// feldbloecke sind die Blöcke, die ein Feld nach MSE2s `Field` beschreiben —
// im Spiel, im Stil und in der Exportvorlage.
var feldbloecke = map[string]bool{
	"card_field":       true,
	"set_field":        true,
	"extra_card_field": true,
	"styling_field":    true,
	"option_field":     true,
}

// Feldartbefund ist ein Feld ohne `type:` oder mit einer Art, die MSE2 nicht
// kennt.
//
// **Die Art steht einzeln da**, nicht nur im Satz: das Fenster setzt sie in
// seinen eigenen Katalogtext ein (§9), und bei einem fehlenden `type:` ist sie
// leer.
type Feldartbefund struct {
	Stelle  Stelle
	Art     string
	Meldung string
}

// Feldartbefunde nennt jedes Feld ohne `type:` und jedes mit einer Art, die
// MSE2 nicht kennt — in der Reihenfolge der Datei.
//
// **Eine Datei, die sich nicht zerteilen lässt, hat hier keinen Befund**: was
// an ihr kaputt ist, sagt der Zerteiler, und zwei Meldungen über denselben
// Fehler hülfen niemandem.
func Feldartbefunde(quelle string) []Feldartbefund {
	befunde := []Feldartbefund{}

	baum, err := Zerteile("feldarten", []byte(quelle))
	if err != nil {
		return befunde
	}

	for _, knoten := range baum.Wurzel {
		if !feldbloecke[Kanonisch(knoten.Schluessel)] {
			continue
		}

		art := knoten.Ersten("type")

		switch {
		case art == nil:
			befunde = append(befunde, Feldartbefund{
				Stelle: knoten.Stelle,
				Meldung: "Diesem Feld fehlt »type:«. MSE2 bricht beim Lesen des Pakets ab; " +
					"hier gilt es als Textfeld.",
			})

		case feldarten[art.Wert] == "":
			befunde = append(befunde, Feldartbefund{
				Stelle: art.Stelle,
				Art:    art.Wert,
				Meldung: fmt.Sprintf("Die Feldart »%s« kennt MSE2 nicht; es bricht beim Lesen des "+
					"Pakets ab. Hier gilt das Feld als Text. Bekannt sind: %s.",
					art.Wert, strings.Join(BekannteFeldarten(), ", ")),
			})
		}
	}

	return befunde
}

// BekannteFeldarten nennt die Arten, sortiert — dieselben wie MSE2.
func BekannteFeldarten() []string {
	namen := make([]string, 0, len(feldarten))
	for name := range feldarten {
		namen = append(namen, name)
	}

	slices.Sort(namen)

	return namen
}
