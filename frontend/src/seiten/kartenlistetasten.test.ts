// **Was eine Taste in einer Zeile der Kartenliste bedeutet** —
// #1290.
//
// Die Gestalt der Liste prüft `kartenlistetabstelle.test.tsx` am gezeichneten
// Bauteil; hier stehen die Ränder, die dort nur mit Mühe zu treffen wären.

import { afterEach, describe, expect, it, vi } from 'vitest'
import type { KeyboardEvent } from 'react'

import { blaettertaste, zurZeile } from './kartenlistetasten'
import type { KartenzeileDTO } from '../api'

const drei: KartenzeileDTO[] = [1, 2, 3].map((nummer) => ({
  id: String(nummer),
  nummer,
  name: `K${String(nummer)}`,
  felder: 1,
}))

// **Jeder Fall beginnt an einem leeren Körper.** Ohne das trüge der Fokus des
// vorigen Falls in den nächsten hinein — und eine Zusicherung über
// `document.activeElement` prüfte dann den Fall davor.
afterEach(() => {
  document.body.innerHTML = ''
})

// ereignis baut ein Tastenereignis über einer Zeile mit zwei Bedienelementen.
function ereignis(
  taste: string,
  zusatz: Partial<KeyboardEvent<HTMLElement>> = {},
) {
  const zeile = document.createElement('div')
  zeile.tabIndex = 0
  zeile.innerHTML = '<input type="checkbox" /><button type="button">x</button>'
  document.body.append(zeile)

  // **Der Spion liegt daneben und nicht am Ereignis.** `expect(fall.preventDefault)`
  // löste die Methode von ihrem Objekt, und `unbound-method` weist das zu Recht
  // ab — ein `this`, das an einer anderen Stelle steht, ist ein Fehler, den
  // niemand sieht.
  const gestoppt = vi.fn()

  return {
    zeile,
    gestoppt,
    fall: {
      key: taste,
      currentTarget: zeile,
      preventDefault: gestoppt,
      metaKey: false,
      altKey: false,
      ctrlKey: false,
      shiftKey: false,
      ...zusatz,
    } as unknown as KeyboardEvent<HTMLElement>,
  }
}

describe('blaettertaste', () => {
  it('lässt eine Zusatztaste in Ruhe', () => {
    const waehlen = vi.fn()

    for (const zusatz of [
      { metaKey: true },
      { altKey: true },
      { ctrlKey: true },
      { shiftKey: true },
    ]) {
      const { fall, gestoppt } = ereignis('ArrowDown', zusatz)
      blaettertaste(fall, drei, 0, waehlen)
      expect(gestoppt).not.toHaveBeenCalled()
    }

    // **Der Beleg, dass der Weg überhaupt trägt** (§3): dieselbe Taste ohne
    // Zusatz wirkt.
    const { fall } = ereignis('ArrowDown')
    blaettertaste(fall, drei, 0, waehlen)
    expect(waehlen).toHaveBeenCalledExactlyOnceWith('2')
  })

  it('tut bei einer Taste ohne Bedeutung nichts', () => {
    const waehlen = vi.fn()
    const { fall, gestoppt } = ereignis('a')

    blaettertaste(fall, drei, 0, waehlen)

    expect(gestoppt).not.toHaveBeenCalled()
    expect(waehlen).not.toHaveBeenCalled()
  })

  it('verbraucht die Taste am Rand, wählt aber nichts', () => {
    const waehlen = vi.fn()
    const { fall, gestoppt } = ereignis('ArrowUp')

    blaettertaste(fall, drei, 0, waehlen)

    expect(gestoppt).toHaveBeenCalled()
    expect(waehlen).not.toHaveBeenCalled()
  })

  it('führt mit → in die Zeile und mit ← zurück', () => {
    const waehlen = vi.fn()
    const { zeile, fall } = ereignis('ArrowRight')

    blaettertaste(fall, drei, 1, waehlen)
    expect(document.activeElement).toBe(zeile.querySelector('input'))

    const zurueck = {
      ...fall,
      key: 'ArrowLeft',
      currentTarget: zeile,
      preventDefault: vi.fn<() => void>(),
    } as unknown as KeyboardEvent<HTMLElement>

    blaettertaste(zurueck, drei, 1, waehlen)
    expect(document.activeElement).toBe(zeile)
  })

  it('bleibt am rechten Rand der Zeile stehen', () => {
    const waehlen = vi.fn()
    const { zeile, fall } = ereignis('ArrowRight')

    zeile.querySelector('button')?.focus()
    blaettertaste(fall, drei, 1, waehlen)

    expect(document.activeElement).toBe(zeile.querySelector('button'))
  })

  it('tut in einer Zeile ohne Bedienelemente nichts Schlimmes', () => {
    const waehlen = vi.fn()
    const { zeile, fall, gestoppt } = ereignis('ArrowRight')
    zeile.innerHTML = ''

    blaettertaste(fall, drei, 1, waehlen)

    expect(gestoppt).toHaveBeenCalled()
    expect(document.activeElement).toBe(document.body)
  })
})

// **Rollen zu einer Stelle** — der Teil, ohne den Ende ins Leere ginge.
describe('zurZeile', () => {
  function gitter(scrollTop: number, hoehe = 400) {
    const knoten = document.createElement('div')
    Object.defineProperty(knoten, 'clientHeight', { value: hoehe })
    knoten.scrollTop = scrollTop

    return knoten
  }

  it('tut ohne Knoten nichts', () => {
    expect(() => {
      zurZeile(null, 5, 40)
    }).not.toThrow()
  })

  it('tut ohne gemessene Zeilenhöhe nichts', () => {
    const knoten = gitter(100)

    zurZeile(knoten, 5, 0)

    expect(knoten.scrollTop).toBe(100)
  })

  it('rollt nach oben, wenn die Zeile darüber liegt', () => {
    const knoten = gitter(400)

    zurZeile(knoten, 2, 40)

    expect(knoten.scrollTop).toBe(80)
  })

  it('rollt nach unten, wenn die Zeile darunter liegt', () => {
    const knoten = gitter(0)

    zurZeile(knoten, 20, 40)

    // 21 * 40 = 840, minus die sichtbaren 400
    expect(knoten.scrollTop).toBe(440)
  })

  it('lässt eine sichtbare Zeile in Ruhe', () => {
    const knoten = gitter(200)

    zurZeile(knoten, 6, 40)

    expect(knoten.scrollTop).toBe(200)
  })
})
