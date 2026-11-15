package msepaket

import (
	"bytes"
	"errors"
	"unicode/utf8"
)

// ErrKodierung meldet eine Paketdatei, die kein gültiges UTF-8 ist —
// FMT-06.AK-01 (#1695).
//
// **MSE2 liest nur UTF-8** und wirft `Invalid UTF-8 sequence on line N`
// (`read_utf8_line`, `src/util/io/reader.cpp`, Zeilen 179–183, Fassung f9d9356
// vom 2021-07-06). Dieses Programm tut dasselbe mit Stelle — `adapted` nur
// darin, dass die Meldung eine Spalte trägt. Geraten wird keine Kodierung:
// ein Latin-1-Paket als UTF-8 zu lesen gäbe falsche Zeichen, still.
var ErrKodierung = errors.New("die Datei ist kein gültiges UTF-8")

// utf16Marken sind die Byte-Reihenfolge-Marken von UTF-16.
var utf16Marken = [][]byte{{0xFF, 0xFE}, {0xFE, 0xFF}}

// kodierungPruefen weist eine Datei in UTF-16 ab — **benannt**, statt ihre
// Nullbytes als Inhalt zu lesen. MSE2 kann sie ebenso wenig.
func kodierungPruefen(datei string, inhalt []byte) error {
	for _, marke := range utf16Marken {
		if bytes.HasPrefix(inhalt, marke) {
			return &Zerteilfehler{
				Stelle:  Stelle{Datei: datei, Zeile: 1, Spalte: 1},
				Meldung: "die Datei ist in UTF-16 geschrieben — Pakete sind UTF-8",
				Grund:   ErrKodierung,
			}
		}
	}

	return nil
}

// ungueltigAb gibt die Spalte des ersten Bytes, das kein gültiges UTF-8 ist —
// null, wenn die Zeile gültig ist. Gezählt wird in Bytes ab eins.
func ungueltigAb(zeile string) int {
	for stelle, zeichen := range zeile {
		if zeichen == utf8.RuneError {
			if _, breite := utf8.DecodeRuneInString(zeile[stelle:]); breite == 1 {
				return stelle + 1
			}
		}
	}

	return 0
}
