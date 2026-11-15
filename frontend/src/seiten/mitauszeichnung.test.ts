// **Getippt wird der Klartext, geschrieben der Wert mit seinen Auszeichnungen**
// (#687, Stufe 2).
import { describe, expect, it } from 'vitest'

import { mitAuszeichnung } from './mitauszeichnung'
import { ohneAuszeichnung } from './ohneauszeichnung'

// GESCHUETZT ist MSEs `ESCAPED_LANGLE` — ein getipptes `<` steht so in der
// Datei (`src/util/tagged_string.hpp`, Zeile 20, Fassung f9d9356 vom
// 2021-07-06).
const GESCHUETZT = '\u0001'
const VERBINDUNG = '\uEB00'

describe('Der Rückweg in den ausgezeichneten Wert', () => {
  const faelle: [string, string, string, string][] = [
    [
      'ein Wert ohne Auszeichnung bleibt, was er ist',
      'Blitzschlag',
      'Blitzschlag!',
      'Blitzschlag!',
    ],
    [
      'getippt wird innerhalb der Auszeichnung',
      '<i-flavor>Es brennt.</i-flavor>',
      'Es brennt!',
      '<i-flavor>Es brennt!</i-flavor>',
    ],
    [
      'geleert bleibt die Auszeichnung stehen',
      '<i-flavor>Es brennt.</i-flavor>',
      '',
      '<i-flavor></i-flavor>',
    ],
    [
      'angehängt wird innen, nicht dahinter',
      '<i-flavor>Es</i-flavor>',
      'Es brennt',
      '<i-flavor>Es brennt</i-flavor>',
    ],
    [
      'davor geschrieben steht innen',
      '<i-flavor>Es</i-flavor>',
      'Und Es',
      '<i-flavor>Und Es</i-flavor>',
    ],
    [
      'eine Marke ohne Inhalt nimmt den Text auf',
      '<i-flavor></i-flavor>',
      'Neu',
      '<i-flavor>Neu</i-flavor>',
    ],
    [
      'über eine Marke hinweg gelöscht bleibt der Baum heil',
      '<w>Creature</w><s> — </s>Goblin',
      'Crelin',
      '<w>Cre</w>lin',
    ],
    [
      'über eine Marke hinweg ersetzt',
      '<w>Creature</w><s> — </s>Goblin',
      'CreXlin',
      '<w>Cre</w>Xlin',
    ],
    [
      'geschachtelt und ganz geleert',
      '<b><i>Text</i></b>',
      '',
      '<b><i></i></b>',
    ],
    [
      'ein getipptes Kleinerzeichen wird geschützt',
      '<i>x</i>',
      'x<y',
      '<i>x' + GESCHUETZT + 'y</i>',
    ],
    [
      'ein geschütztes Kleinerzeichen bleibt geschützt',
      '<i>a' + GESCHUETZT + 'b</i>',
      'a<b!',
      '<i>a' + GESCHUETZT + 'b!</i>',
    ],
    [
      'nichts geändert heißt nichts angefasst',
      '<w>Creature</w><s> — </s>Goblin',
      'Creature — Goblin',
      '<w>Creature</w><s> — </s>Goblin',
    ],
  ]

  for (const [name, roh, klartext, erwartet] of faelle) {
    it(name, () => {
      expect(mitAuszeichnung(roh, klartext)).toBe(erwartet)
    })

    // **Die Zusage, an der alles hängt**: was der Editor zeigt, ist genau das,
    // was jemand getippt hat. Bricht sie, ist ein Wert verloren.
    it(name + ' — und der Klartext kommt heil zurück', () => {
      expect(ohneAuszeichnung(mitAuszeichnung(roh, klartext))).toBe(klartext)
    })
  }

  // **Die beiden Zeichen, die MSE für sich behält** —
  // #718.
  //
  // # Der Befund
  //
  // Aus dem Zweitblick durch Codex vom 2026-08-21. Er hat die Zusage über 23
  // Auszeichnungsformen und 10 Sonderzeichen gefahren — **230 Rundreisen, 46
  // Fehlschläge**, und alle 46 an denselben zwei Zeichen: `U+0001` und
  // `U+EB00`.
  //
  // `ohneAuszeichnung` deutet sie um — das eine als `<`, das andere als
  // Leerzeichen —, und `mitAuszeichnung` schützte nur `<`. Ein Klartext, der
  // eines von beiden **wirklich** enthält, kam deshalb anders zurück, als er
  // hineinging.
  //
  // # Warum sie herausfallen und nicht kodiert werden
  //
  // **Weil MSE dieselbe Zweideutigkeit hat.** `escape()` bildet allein `<` auf
  // `ESCAPED_LANGLE` ab (`src/util/tagged_string.cpp`, Zeilen 19–22, Fassung
  // f9d9356 vom 2021-07-06); ein `\1` in der Eingabe ist danach von einem
  // geschützten `<` nicht mehr zu unterscheiden. Wer sie kodierte, erfände
  // eine Darstellung, die kein anderes Programm liest.
  //
  // Beide sind in einem Kartentext ohnehin nichts: `U+0001` ist ein
  // Steuerzeichen, `U+EB00` liegt im privaten Bereich.
  describe('die beiden Zeichen, die MSE für sich behält', () => {
    const reserviert: [string, string][] = [
      ['ein Steuerzeichen', GESCHUETZT],
      ['ein Zeichen aus dem privaten Bereich', VERBINDUNG],
    ]

    for (const [name, zeichen] of reserviert) {
      it(`${name} fällt aus dem getippten Text`, () => {
        const roh = mitAuszeichnung('<i>alt</i>', 'a' + zeichen + 'b')

        expect(roh).toBe('<i>ab</i>')
        expect(ohneAuszeichnung(roh)).toBe('ab')
      })
    }

    // **Und was schon im Wert steht, bleibt** — es ist dort kein Text,
    // sondern die Schreibweise eines Zeichens, das MSE selbst gesetzt hat.
    it('lässt ein geschütztes Zeichen im Wert unangetastet', () => {
      const roh = '<i>a' + GESCHUETZT + 'b</i>'

      expect(mitAuszeichnung(roh, ohneAuszeichnung(roh))).toBe(roh)
    })
  })
})
