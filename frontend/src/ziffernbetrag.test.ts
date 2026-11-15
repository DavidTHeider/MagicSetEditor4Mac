// **Ein Betrag in der kleinsten Einheit ist eine Folge von Ziffern** —
// Nachlauf über Phase 6.3, N5.
import { describe, expect, it } from 'vitest'

import { ziffernbetrag } from './ziffernbetrag'

describe('ziffernbetrag', () => {
  it('liest Ziffern, auch mit Leerraum außen', () => {
    expect(ziffernbetrag('1200')).toBe(1200)
    expect(ziffernbetrag(' 0 ')).toBe(0)
  })

  // **»1.000« ist für `Number` eins** und im Deutschen tausend; »1,000« ist
  // für `Number` gar keine Zahl und im Englischen tausend. Geraten wird
  // keines von beiden.
  it('weist jede andere Schreibweise ab, statt sie zu deuten', () => {
    for (const eingabe of [
      '1.000',
      '1,000',
      '12,50',
      '10.00',
      '1 000',
      '-5',
      '1e3',
      '',
      '0x10',
    ]) {
      expect(ziffernbetrag(eingabe), eingabe).toBeUndefined()
    }
  })

  // **Mehr, als eine Zahl in JavaScript genau fasst, geht nicht hinüber** —
  // es käme auf der Go-Seite als eine andere Zahl an.
  it('weist eine Zahl jenseits der genauen Ganzzahlen ab', () => {
    expect(ziffernbetrag('9007199254740993')).toBeUndefined()
  })
})
