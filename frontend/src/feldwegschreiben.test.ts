// **Ein offenes Feld wird verlassen, bevor das Bezugsobjekt wechselt** —
// Befund 1 und 5 des Nachlaufs über Phase 4.3
// (#1290).

import { afterEach, describe, expect, it, vi } from 'vitest'

import { offenesFeld, wegschreiben } from './feldwegschreiben'

afterEach(() => {
  document.body.innerHTML = ''
})

function mitFokus(gestalt: string, beschriftung = 'name') {
  document.body.innerHTML = gestalt
  const knoten = document.querySelector<HTMLElement>(
    `[aria-label="${beschriftung}"]`,
  )
  knoten?.focus()

  return knoten
}

describe('wegschreiben', () => {
  for (const [was, gestalt] of [
    ['ein Textfeld', '<input aria-label="name" />'],
    ['ein mehrzeiliges Feld', '<textarea aria-label="name"></textarea>'],
    // **Ein Auswahlfeld gehört dazu** (Befund 5): es schreibt zwar sofort, aber
    // der Fokus soll auch von dort mitwandern.
    [
      'ein Auswahlfeld',
      '<select aria-label="name"><option>a</option></select>',
    ],
  ] as const) {
    it(`verlässt ${was}`, () => {
      const knoten = mitFokus(gestalt)
      const verlassen = vi.fn()
      knoten?.addEventListener('blur', verlassen)

      wegschreiben()

      expect(verlassen).toHaveBeenCalled()
      expect(document.activeElement).toBe(document.body)
    })

    it(`nennt ${was} beim Namen`, () => {
      mitFokus(gestalt)

      expect(offenesFeld()).toBe('name')
    })
  }

  it('lässt einen Knopf in Ruhe', () => {
    // **Ein Knopf trägt keinen ungeschriebenen Wert.** Ihn zu verlassen nähme
    // dem Nutzer den Fokus, ohne etwas zu retten.
    const knoten = mitFokus(
      '<button type="button" aria-label="name">x</button>',
    )

    wegschreiben()

    expect(document.activeElement).toBe(knoten)
    expect(offenesFeld()).toBeUndefined()
  })

  it('tut ohne Fokus nichts', () => {
    document.body.innerHTML = '<p>nichts</p>'

    expect(() => {
      wegschreiben()
    }).not.toThrow()
    expect(offenesFeld()).toBeUndefined()
  })
})
