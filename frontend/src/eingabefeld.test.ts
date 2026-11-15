// **Was als Eingabefeld gilt** — Werkstück 1 von Phase 3.7
// (#1168).
import { describe, expect, it } from 'vitest'

import { inEingabe } from './eingabefeld'

// bau stellt ein Element her und gibt es als Ereignisziel zurück.
function bau(html: string): EventTarget {
  const huelle = document.createElement('div')
  huelle.innerHTML = html

  const kind = huelle.firstElementChild
  if (kind === null) {
    throw new Error('kein Element')
  }

  return kind
}

describe('inEingabe', () => {
  it.each([
    ['<input />', true],
    ['<input type="text" />', true],
    ['<input type="search" />', true],
    ['<input type="number" />', true],
    ['<textarea></textarea>', true],
    ['<div contenteditable="true"></div>', true],

    // **Ein Ankreuzfeld tippt nicht.** Es steht in jeder Zeile der
    // Kartenliste, und wer eine Karte anhakt, hat den Fokus darauf — ⌘C
    // meinte dort die Karten und traf nichts.
    ['<input type="checkbox" />', false],
    ['<input type="radio" />', false],
    ['<input type="button" />', false],
    ['<input type="file" />', false],
    ['<input type="range" />', false],
    ['<input type="color" />', false],

    ['<button></button>', false],
    ['<div></div>', false],
    ['<div contenteditable="false"></div>', false],
  ])('%s → %s', (html, erwartet) => {
    expect(inEingabe(bau(html))).toBe(erwartet)
  })

  it('sagt nein zu etwas, das kein Element ist', () => {
    expect(inEingabe(null)).toBe(false)
    expect(inEingabe(new EventTarget())).toBe(false)
  })

  // **Die Vererbung trägt mit** — ein `<b>` in einem bearbeitbaren Bereich.
  it('findet den bearbeitbaren Bereich über einen Nachfahren', () => {
    const bereich = document.createElement('div')
    bereich.setAttribute('contenteditable', '')

    const fett = document.createElement('b')
    bereich.append(fett)

    expect(inEingabe(fett)).toBe(true)
  })
})
