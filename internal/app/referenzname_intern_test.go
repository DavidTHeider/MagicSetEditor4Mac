//go:build fremdpakete

// Der Dateiname, unter dem MSE2 seine Ausgabe ablegt
// (#535).
//
// # Der Befund, der diese Datei nötig gemacht hat
//
// Der Messlauf suchte das Referenzbild unter dem **rohen** Kartennamen. MSE2
// legt es unter dem **bereinigten** ab — und damit fiel jede Karte mit Komma
// oder Apostroph still aus der Stichprobe.
//
// Über die vier gemessenen Sätze waren das **64 von 574** Karten, und keine
// einzige davon fehlte wirklich. Bei den Planeswalkern wog es am schwersten:
// **68 % von ihnen** heißen `Ajani, Caller of the Pride` — mit Komma. Die
// neun, die gemessen wurden, waren die atypischen 32 %.
//
// Gefunden vom Zweitblick durch Codex über Phase I.
package app

import (
	"strings"
	"unicode"
)

// sauberDateiname bildet MSEs `clean_filename` nach.
//
//	bool is_filename_char(Char c) {
//	  return isAlnum(c) || c == _(' ') || c == _('_') || c == _('-') || c == _('.');
//	}
//	bool is_filename_char_trim(Char c) {
//	  return c == _(' ') || c == _('.');
//	}
//
// `src/util/file_utils.cpp`, Zeilen 47–71, Fassung f9d9356 vom 2021-07-06.
// Gerufen wird es beim Export aus `export_images`
// (`src/data/format/image.cpp`, Zeile 82).
//
// **Entfernt, nicht ersetzt.** Ein Komma verschwindet spurlos; aus
// `Ajani, Caller of the Pride` wird `Ajani Caller of the Pride` mit **zwei**
// Zeichen weniger und nicht mit einem Ersatzzeichen.
//
// **Führende Leerzeichen und Punkte fallen weg, abschließende auch.** Und ein
// Name, von dem nichts übrig bleibt, heißt `no-name`.
func sauberDateiname(name string) string {
	var sauber strings.Builder

	anfang := true

	for _, zeichen := range name {
		if !istDateizeichen(zeichen) || (anfang && istRandzeichen(zeichen)) {
			continue
		}

		anfang = false

		sauber.WriteRune(zeichen)
	}

	ergebnis := strings.TrimRightFunc(sauber.String(), istRandzeichen)
	if ergebnis == "" {
		return "no-name"
	}

	return ergebnis
}

// istDateizeichen ist MSEs `is_filename_char`.
//
// **`isAlnum` ist MSEs eigenes**, nicht das der C-Bibliothek: es gilt für den
// ganzen Unicode-Bereich, und deshalb steht hier `unicode` und nicht ASCII.
func istDateizeichen(zeichen rune) bool {
	switch {
	case zeichen >= 'a' && zeichen <= 'z',
		zeichen >= 'A' && zeichen <= 'Z',
		zeichen >= '0' && zeichen <= '9':
		return true

	case zeichen == ' ', zeichen == '_', zeichen == '-', zeichen == '.':
		return true
	}

	return zeichen > 127 && (unicode.IsLetter(zeichen) || unicode.IsDigit(zeichen))
}

// istRandzeichen ist MSEs `is_filename_char_trim`.
func istRandzeichen(zeichen rune) bool {
	return zeichen == ' ' || zeichen == '.'
}

// referenzdatei gibt den Dateinamen, unter dem MSE2 das Bild einer Karte ablegt
// — `clean_filename` über `{card.name}.png` als Ganzes, wie `export_images`
// es ruft (`src/data/format/image.cpp`, Zeile 82, Fassung f9d9356 vom
// 2021-07-06). Der Schlusspunkt eines Namens steht dann vor `.png` und bleibt.
func referenzdatei(name string) string {
	return sauberDateiname(name + ".png")
}
