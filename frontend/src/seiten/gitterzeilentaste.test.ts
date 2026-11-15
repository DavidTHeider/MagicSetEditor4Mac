// **Der Tastenvertrag einer Gitterzeile** —
// #1290.
//
// Die Ränder stehen hier; dass sie an der Paketliste ankommen, prüft
// `paketverwaltung.test.tsx` am gezeichneten Bauteil.

import { afterEach, describe, expect, it, vi } from 'vitest'
import type { KeyboardEvent } from 'react'

import { gitterzeilentaste } from './gitterzeilentaste'

afterEach(() => {
  document.body.innerHTML = ''
})

// dreiZeilen baut eine Tabelle mit drei Zeilen und je einem Knopf.
//
// **Sie gibt ein Tripel und keine Liste** — so braucht kein Fall unten ein
// `!`, und ein `?.` an einer Stelle, an der die Zeile sicher dasteht, verbürge
// das Gegenteil.
function dreiZeilen(): [HTMLElement, HTMLElement, HTMLElement] {
  const rumpf = document.createElement('tbody')
  rumpf.innerHTML =
    '<tr tabindex="0"><td><button type="button">a</button></td></tr>' +
    '<tr tabindex="-1"><td><button type="button">b</button></td></tr>' +
    '<tr tabindex="-1"><td><button type="button">c</button></td></tr>'

  const tabelle = document.createElement('table')
  tabelle.append(rumpf)
  document.body.append(tabelle)

  const [erste, zweite, dritte] = [...rumpf.children] as HTMLElement[]
  if (erste === undefined || zweite === undefined || dritte === undefined) {
    throw new Error('die Probe hat nicht drei Zeilen gebaut')
  }

  return [erste, zweite, dritte]
}

function fall(taste: string, zeile: HTMLElement, zusatz = {}) {
  const gestoppt = vi.fn()

  return {
    gestoppt,
    ereignis: {
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

describe('gitterzeilentaste', () => {
  it('lässt eine Zusatztaste in Ruhe', () => {
    const [erste] = dreiZeilen()
    const rueckenAuf = vi.fn()

    for (const zusatz of [
      { metaKey: true },
      { altKey: true },
      { ctrlKey: true },
      { shiftKey: true },
    ]) {
      const { ereignis, gestoppt } = fall('ArrowDown', erste, zusatz)
      gitterzeilentaste(ereignis, 0, 3, rueckenAuf)
      expect(gestoppt).not.toHaveBeenCalled()
    }

    // **Der Beleg, dass der Weg trägt** (§3).
    const { ereignis } = fall('ArrowDown', erste)
    gitterzeilentaste(ereignis, 0, 3, rueckenAuf)
    expect(rueckenAuf).toHaveBeenCalledExactlyOnceWith(1)
  })

  it('tut bei einer Taste ohne Bedeutung nichts', () => {
    const [erste] = dreiZeilen()
    const rueckenAuf = vi.fn()
    const { ereignis, gestoppt } = fall('x', erste)

    gitterzeilentaste(ereignis, 0, 3, rueckenAuf)

    expect(gestoppt).not.toHaveBeenCalled()
    expect(rueckenAuf).not.toHaveBeenCalled()
  })

  it('führt mit ↓ zur nächsten Zeile und fokussiert sie', () => {
    const [erste, zweite] = dreiZeilen()
    const rueckenAuf = vi.fn()

    gitterzeilentaste(fall('ArrowDown', erste).ereignis, 0, 3, rueckenAuf)

    expect(rueckenAuf).toHaveBeenCalledWith(1)
    expect(document.activeElement).toBe(zweite)
  })

  it('führt mit Pos1 und Ende an die Enden', () => {
    const [erste, , dritte] = dreiZeilen()
    const rueckenAuf = vi.fn()

    gitterzeilentaste(fall('End', erste).ereignis, 0, 3, rueckenAuf)
    expect(document.activeElement).toBe(dritte)

    gitterzeilentaste(fall('Home', dritte).ereignis, 2, 3, rueckenAuf)
    expect(document.activeElement).toBe(erste)
  })

  it('bleibt an den Rändern stehen', () => {
    const [erste] = dreiZeilen()
    const rueckenAuf = vi.fn()

    gitterzeilentaste(fall('ArrowUp', erste).ereignis, 0, 3, rueckenAuf)

    expect(rueckenAuf).toHaveBeenCalledWith(0)
    expect(document.activeElement).toBe(erste)
  })

  it('führt mit → in die Zeile', () => {
    const [, zweite] = dreiZeilen()
    const rueckenAuf = vi.fn()

    gitterzeilentaste(fall('ArrowRight', zweite).ereignis, 1, 3, rueckenAuf)

    expect(document.activeElement).toBe(zweite.querySelector('button'))
    expect(rueckenAuf).not.toHaveBeenCalled()
  })

  it('tut ohne Elternknoten nichts Schlimmes', () => {
    // **Eine Zeile ohne Tabelle gibt es im Betrieb nicht** — der Zweig steht
    // trotzdem da, weil `parentElement` `null` sein **kann** und ein Zugriff
    // darauf sonst der einzige Weg zu einer Panik wäre.
    const einzeln = document.createElement('tr')
    const rueckenAuf = vi.fn()

    gitterzeilentaste(fall('ArrowDown', einzeln).ereignis, 0, 3, rueckenAuf)

    expect(rueckenAuf).toHaveBeenCalledWith(1)
  })
})
