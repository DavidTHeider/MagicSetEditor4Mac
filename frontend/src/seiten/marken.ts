// **Was in einem Kartenwert eine Marke ist** — die eine Stelle, an der es
// steht (#687).
//
// Eigene Datei, weil zwei Wege dieselbe Antwort brauchen: [ohneAuszeichnung]
// nimmt die Marken heraus, [mitAuszeichnung] setzt einen Klartext zwischen
// sie zurück. Zwei Regeln liefen auseinander, und dann zeigte der Editor
// etwas anderes, als geschrieben würde.

// MARKE trifft eine öffnende oder schließende Marke — und nur die.
//
// **Ein `<` in einem Kartentext ist selten, aber möglich.** `3 < 4` soll
// stehen bleiben; deshalb muss auf die Klammer ein Buchstabe oder ein
// Schrägstrich folgen, und dazwischen darf keine weitere Klammer stehen.
//
// **`y` und nicht `g`**: dieser Ausdruck wird an einer **Stelle** angesetzt,
// nicht über die ganze Zeichenkette. `lastIndex` wird vor jedem Gebrauch
// gesetzt.
const MARKE = /<\/?[a-zA-Z][^<>]*>/y

// GESCHUETZT ist MSEs `ESCAPED_LANGLE`.
//
//	/// Escaped '<'
//	const Char ESCAPED_LANGLE = _('\1');
//
// `src/util/tagged_string.hpp`, Zeile 20, Fassung f9d9356 vom 2021-07-06. Ein
// **getipptes** `<` steht so in der Datei — sonst wäre es beim nächsten Lesen
// der Anfang einer Marke.
export const GESCHUETZT = '\u0001'

// VERBINDUNGSLEERZEICHEN ist MSEs `CONNECTION_SPACE`.
//
//	#define CONNECTION_SPACE _('\xEB00') // in private use area, untags to ' '
//
// `src/util/string.hpp`, Zeile 85, Fassung f9d9356 vom 2021-07-06. Es entsteht
// beim Auflösen von Schlüsselwörtern und liest sich als gewöhnliches
// Leerzeichen.
export const VERBINDUNGSLEERZEICHEN = '\uEB00'

// markeAn gibt die Länge der Marke an dieser Stelle — oder 0.
export function markeAn(text: string, stelle: number): number {
  if (text[stelle] !== '<') {
    return 0
  }

  MARKE.lastIndex = stelle
  const treffer = MARKE.exec(text)

  return treffer === null ? 0 : treffer[0].length
}

// istSchliessend sagt, ob die Marke an dieser Stelle eine schließende ist.
export function istSchliessend(text: string, stelle: number): boolean {
  return text.startsWith('</', stelle)
}

// rohStelle rechnet eine Stelle im Klartext in eine im rohen Wert um.
//
//	size_t untagged_to_index(const String& str, size_t pos, bool inside, …) {
//	  size_t i = start_index, p = 0;
//	  while (i < str.size()) {
//	    Char c = str.GetChar(i);
//	    if (c == _('<')) {
//	      bool is_close = is_substr(str, i, _("</"));
//	      if (p == pos && is_close == inside) break;
//	      i = skip_tag(str, i);
//	    } else {
//	      if (p == pos) break;
//	      i++; p++;
//	    }
//	  }
//	  return i;
//	}
//
// `src/util/tagged_string.cpp`, Zeilen 526–541, Fassung f9d9356 vom
// 2021-07-06.
//
// **`inside` ist hier fest `true`**, und das ist die ganze Entscheidung: an
// einer öffnenden Marke geht der Weg weiter, an einer schließenden hält er an.
// Damit landet jede Änderung **innerhalb** der Auszeichnung, in der sie
// steht — wer den Aromatext leert, behält `<i-flavor></i-flavor>`, und wer
// hinten anfügt, schreibt kursiv weiter. Genau so tippt es sich bei MSE, weil
// dort der Cursor zwischen den Marken steht und nicht davor.
export function rohStelle(roh: string, ziel: number): number {
  let stelle = 0
  let klar = 0

  while (stelle < roh.length) {
    const laenge = markeAn(roh, stelle)

    if (laenge > 0) {
      if (klar === ziel && istSchliessend(roh, stelle)) {
        return stelle
      }

      stelle += laenge

      continue
    }

    if (klar === ziel) {
      return stelle
    }

    stelle++
    klar++
  }

  return roh.length
}
