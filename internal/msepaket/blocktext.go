package msepaket

import "strings"

// Blocktext gibt den Wert eines Eintrags als mehrzeiligen Text — so, wie MSE
// ihn liest:
//
//	previous_value += line.substr(expected_indent); // strip expected indent
//
// `Reader::getValue`, `src/util/io/reader.cpp`, Zeilen 255–298, Fassung
// f9d9356 vom 2021-07-06. **Weg fällt genau die Einrückung des Blocks**, jede
// tiefere bleibt. Unter `extra_data:` stehen die Stilvorlagen und unter jeder
// ihre Felder (#2298) — [Knoten.Skripttext] nimmt **jede** Einrückung weg,
// richtig für ein Skript, falsch für einen Wert: beim Sichern stand dann
// `keyword_icons` neben `magic-nouveau` statt darunter.
//
// **Die Einrückung des Blocks ist die des Schlüssels plus eins** —
// `Reader::handle` ruft erst `enterBlock`, und mit diesem `expected_indent`
// liest `getValue`. Eine erste Zeile, die tiefer steht, behält den
// Überschuss (Befund 5 des Zweitblicks über E/F). Nur wo der Schlüssel keine
// Rohzeile trägt, gilt die erste Zeile. Eine Zeile, die der Schreiber neu
// zusammengesetzt hat, trägt keine Rohzeile; sie bekommt die Tabs ihrer
// Tiefe im Baum.
func (k *Knoten) Blocktext() string {
	if strings.TrimSpace(k.Wert) != "" {
		return k.Wert
	}

	basis := -1
	if k.Roh != "" {
		basis = einzugVon(k.Roh) + 1
	}

	var zeilen []string

	var sammle func(knoten []*Knoten, tiefe int)
	sammle = func(knoten []*Knoten, tiefe int) {
		for _, kind := range knoten {
			zeilen = append(zeilen, blockzeile(kind, tiefe, &basis))
			sammle(kind.Kinder, tiefe+1)
		}
	}

	sammle(k.Kinder, 0)

	return strings.Join(zeilen, "\n")
}

// blockzeile gibt eine Zeile des Blocks ohne seine Grundeinrückung.
func blockzeile(knoten *Knoten, tiefe int, basis *int) string {
	if knoten.Roh == "" {
		return strings.Repeat("\t", tiefe) + knoten.Zeile()
	}

	tabs := einzugVon(knoten.Roh)
	if *basis < 0 {
		*basis = tabs - tiefe
	}

	return knoten.Roh[min(*basis, tabs):]
}

// einzugVon zählt die Tabs, mit denen eine Rohzeile beginnt.
func einzugVon(roh string) int {
	return len(roh) - len(strings.TrimLeft(roh, "\t"))
}
