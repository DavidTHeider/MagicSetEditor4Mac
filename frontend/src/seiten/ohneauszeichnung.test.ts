import { describe, expect, it } from 'vitest'

import { ohneAuszeichnung } from './ohneauszeichnung'

// **Ein Hinweis zeigt den Text, nicht seine Auszeichnungen**
// (#687).
describe('ohneAuszeichnung', () => {
  const faelle: [string, string, string][] = [
    [
      'der Fall aus der Anwendung',
      '<word-list-type-en><word-list-type-en><sep-soft> — </sep-soft>',
      ' — ',
    ],
    ['Aromatext', '<i-flavor>Ein Spruch.</i-flavor>', 'Ein Spruch.'],
    ['ohne Auszeichnung bleibt alles', 'Feuerball', 'Feuerball'],
    // **Ein Kleinerzeichen ist keine Auszeichnung.** Es kommt in einem
    // Kartentext vor, und es soll stehen bleiben.
    ['ein Vergleich bleibt stehen', 'Stärke 3 < 4', 'Stärke 3 < 4'],
    ['ein einzelnes Kleinerzeichen', '<', '<'],
    // **Ein geschütztes Kleinerzeichen ist eines.** MSE legt ein getipptes `<`
    // als `ESCAPED_LANGLE` ab, damit es beim nächsten Lesen keine Marke
    // beginnt; gezeigt wird es als das, was jemand getippt hat.
    //
    //	wxUniChar untag_char(wxUniChar c) {
    //	  if (c == ESCAPED_LANGLE) return _('<');
    //	  else if (c == CONNECTION_SPACE) return _(' ');
    //	  else return c;
    //	}
    //
    // `src/util/tagged_string.cpp`, Zeilen 14–18, Fassung f9d9356 vom
    // 2021-07-06.
    ['ein geschütztes Kleinerzeichen', 'a\u0001b', 'a<b'],
    [
      'ein Verbindungsleerzeichen liest sich als Leerzeichen',
      'a\uEB00b',
      'a b',
    ],
    ['leer bleibt leer', '', ''],
  ]

  for (const [name, roh, erwartet] of faelle) {
    it(name, () => {
      expect(ohneAuszeichnung(roh)).toBe(erwartet)
    })
  }
})
