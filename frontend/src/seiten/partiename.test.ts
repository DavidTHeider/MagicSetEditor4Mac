// **Wie eine Partie heißt** — UX-03.AK-02
// (#1675):
// die Brotkrume am Tisch sagt *Spielen › Partie gegen Bo › Am Tisch*, wie im
// Artboard *Tisch*.
//
// # Wogegen hier geprüft wird
//
// **Dass der Name aus der Sicht dessen kommt, der hinsieht.** Wer an einem
// Sitz spielt, spielt **gegen** die anderen; wer zusieht, sieht eine Partie
// **zwischen** allen; und wer allein spielt, hat kein Gegenüber.
import { describe, expect, test } from 'vitest'

import { einSitz, einePartie } from './partieprobe'
import { kataloge } from '../sprache/kataloge'
import { partiename } from './partiewoerter'

const vier = [
  einSitz({ kennung: 'sitz1', name: 'Anna' }),
  einSitz({ kennung: 'sitz2', name: 'Bert' }),
  einSitz({ kennung: 'sitz3', name: 'Cleo' }),
  einSitz({ kennung: 'sitz4', name: 'Dora' }),
]

describe('partiename', () => {
  test('nennt vom eigenen Sitz aus das Gegenüber', () => {
    expect(partiename(kataloge.de.seiten.spielen, einePartie(), 'de')).toBe(
      'Partie gegen Bert',
    )
  })

  // **Die Aufzählung in der Sprache des Fensters** — „und" hier, „, and"
  // dort.
  test('zählt mehrere in der Sprache des Fensters auf', () => {
    const partie = einePartie({ sitze: vier })

    expect(partiename(kataloge.de.seiten.spielen, partie, 'de')).toBe(
      'Partie gegen Bert, Cleo und Dora',
    )
    expect(partiename(kataloge.en.seiten.spielen, partie, 'en')).toBe(
      'Game against Bert, Cleo, and Dora',
    )
  })

  test('nennt einem Zuschauer alle', () => {
    expect(
      partiename(
        kataloge.de.seiten.spielen,
        einePartie({ fuer: 'zuschauer' }),
        'de',
      ),
    ).toBe('Partie zwischen Anna und Bert')
  })

  test('nennt eine Partie allein nach dem, der sie spielt', () => {
    expect(
      partiename(
        kataloge.de.seiten.spielen,
        einePartie({ sitze: [einSitz({ kennung: 'sitz1', name: 'Anna' })] }),
        'de',
      ),
    ).toBe('Partie von Anna')
  })
})
