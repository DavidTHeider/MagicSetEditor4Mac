// **Die Sprache an die native Menüleiste, der Reihe nach** — Befund 3 des
// Nachlaufs über Phase 6.2.
//
// `useLauf` verwirft im Fenster die verspätete Antwort; Go aber setzt die
// Leiste bei **jedem** Ruf, der ankommt. Zwei schnelle Wechsel konnten dort
// in umgekehrter Reihenfolge ankommen — und die Leiste stand zuletzt in der
// älteren Sprache, während das Fenster die neuere zeigte.
import { afterEach, describe, expect, it, vi } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'
import { spracheMelden } from './menue'

afterEach(() => {
  raeumeApi()
})

// eineRunde lässt alles laufen, was bis zur nächsten Makroaufgabe ansteht.
const eineRunde = () =>
  new Promise((fertig) => {
    setTimeout(fertig, 0)
  })

describe('die Sprache an die Menüleiste', () => {
  it('schickt den nächsten Wechsel erst, wenn der vorige angekommen ist', async () => {
    const angefangen: string[] = []
    let loslassen: (() => void) | undefined

    belegeApi({
      SpracheSetzen: vi.fn((sprache: string) => {
        angefangen.push(sprache)

        return sprache === 'en'
          ? new Promise<{ fehler: null }>((fertig) => {
              loslassen = () => {
                fertig({ fehler: null })
              }
            })
          : Promise.resolve({ fehler: null })
      }),
    })

    const erste = spracheMelden('en')
    const zweite = spracheMelden('de')
    await eineRunde()

    // **Der zweite wartet**, solange der erste in Go unterwegs ist.
    expect(angefangen).toEqual(['en'])

    loslassen?.()
    await erste
    await zweite

    expect(angefangen).toEqual(['en', 'de'])
  })

  it('lässt einen gescheiterten Wechsel den nächsten nicht aufhalten', async () => {
    const angefangen: string[] = []

    belegeApi({
      SpracheSetzen: vi.fn((sprache: string) => {
        angefangen.push(sprache)

        return sprache === 'en'
          ? Promise.reject(new Error('die Brücke ist weg'))
          : Promise.resolve({ fehler: null })
      }),
    })

    const erste = spracheMelden('en')
    const zweite = spracheMelden('de')

    await expect(erste).rejects.toThrow('die Brücke ist weg')
    await zweite

    expect(angefangen).toEqual(['en', 'de'])
  })
})
