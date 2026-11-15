// **Kartentext ohne seine Auszeichnungen** —
// #687.
//
// # Wofür
//
// Ein Feldwert trägt Auszeichnungen: `<i-flavor>` für den Aromatext,
// `<word-list-type-en>` für einen Wortlisteneintrag, `<sep-soft>` für einen
// weichen Trenner. Sie stehen so in der `.mse-set`-Datei und bleiben dort —
// sie zu entfernen hieße, den Wert zu ändern.
//
// **Für einen Hinweis gilt das nicht.** Die hergeleitete Angabe unter einem
// Feld sagt, was dort stünde, wenn niemand etwas einträgt; sie ist ein Satz an
// jemanden und kein Wert. `Hergeleitet: <word-list-type-en><sep-soft> —
// </sep-soft>` beantwortet keine Frage.
//
// # Was sie nicht tut
//
// **Sie fasst nichts an, was geschrieben wird.** Den Weg zurück in den rohen
// Wert geht [mitAuszeichnung]; diese Datei nimmt nur weg. Stufe 3 von #687 —
// ein Feld, das die Auszeichnung *darstellt*, kursiv als kursiv — ist eine
// eigene Entscheidung.

import { GESCHUETZT, markeAn, VERBINDUNGSLEERZEICHEN } from './marken'

// ohneAuszeichnung nimmt die spitzen Klammern samt Inhalt heraus — und macht
// die geschützten Zeichen wieder lesbar.
//
//	String untag(const String& str) { … }
//	wxUniChar untag_char(wxUniChar c) {
//	  if (c == ESCAPED_LANGLE) return _('<');
//	  else if (c == CONNECTION_SPACE) return _(' ');
//	  else return c;
//	}
//
// `src/util/tagged_string.cpp`, Zeilen 14–18 und 24–37, Fassung f9d9356 vom
// 2021-07-06.
//
// **Was wie eine Marke aussieht, und nur das.** Ein `<` in einem Kartentext
// ist selten, aber möglich — `3 < 4` soll stehen bleiben. Deshalb muss auf die
// öffnende Klammer ein Buchstabe oder ein Schrägstrich folgen ([markeAn]).
//
// **Und die beiden geschützten Zeichen kommen zurück**: ein getipptes `<` legt
// MSE als `ESCAPED_LANGLE` ab, ein aufgelöstes Schlüsselwort setzt
// `CONNECTION_SPACE`. Beide sind für die Datei gedacht und nicht für die
// Anzeige; wer sie stehen ließe, zeigte ein unsichtbares Zeichen an der Stelle
// eines gewöhnlichen.
export function ohneAuszeichnung(text: string): string {
  let klar = ''

  let stelle = 0
  while (stelle < text.length) {
    const laenge = markeAn(text, stelle)
    if (laenge > 0) {
      stelle += laenge

      continue
    }

    klar += lesbar(text.charAt(stelle))
    stelle++
  }

  return klar
}

// lesbar macht aus einem geschützten Zeichen das, was es meint.
//
// **`charAt` und nicht `[]`**: der Indexzugriff gibt `string | undefined`, und
// der Rückfall darauf wäre ein Zweig, den kein Aufruf betreten kann (§10) —
// die Schleife läuft nur über vorhandene Stellen.
function lesbar(zeichen: string): string {
  if (zeichen === GESCHUETZT) {
    return '<'
  }

  if (zeichen === VERBINDUNGSLEERZEICHEN) {
    return ' '
  }

  return zeichen
}
