// **Ein Menü hinter einem Knopf steht fest am Fenster** — AP-6
// (#2363),
// gemessen am nativen Fenster mit `ap6decks09`: unter dem Knopf „…“ einer
// Deckzeile schnitt die Spalte des Navigators (`overflow`) das Menü ab —
// derselbe Fehler, den das Kontextmenü mit AP-5 hatte. Es steht deshalb
// fest unter dem Knopf, und am unteren Rand klappt es über ihn.
import { act, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Menue } from './Menue'

afterEach(() => {
  vi.restoreAllMocks()
})

// lege setzt den Knopf an eine Stelle und gibt dem Menü 200 × 120 Punkte.
function lege(knopf: {
  links: number
  oben: number
  unten: number
  rechts: number
}) {
  vi.spyOn(HTMLElement.prototype, 'getBoundingClientRect').mockImplementation(
    function (this: HTMLElement) {
      if (this.getAttribute('role') === 'menu') {
        const oben = Number.parseFloat(this.style.top)
        const links = Number.parseFloat(this.style.left)

        return new DOMRect(links, oben, 200, 120)
      }

      return new DOMRect(
        knopf.links,
        knopf.oben,
        knopf.rechts - knopf.links,
        knopf.unten - knopf.oben,
      )
    },
  )
}

function zeige(ausrichtung: 'links' | 'rechts' = 'links') {
  render(
    <Menue
      bezeichnung="Handlungen"
      ausrichtung={ausrichtung}
      eintraege={[{ schluessel: 'a', name: 'Öffnen', aufWahl: vi.fn() }]}
    >
      …
    </Menue>,
  )
  fireEvent.click(screen.getByRole('button', { name: 'Handlungen' }))

  return screen.getByRole('menu')
}

describe('die Lage eines Knopfmenüs', () => {
  it('steht fest unter dem Knopf, links bündig', () => {
    lege({ links: 40, oben: 100, unten: 130, rechts: 80 })

    const menue = zeige()

    expect(menue.style.position).toBe('fixed')
    expect(menue.style.left).toBe('40px')
    expect(menue.style.top).toBe('134px')
  })

  it('richtet sich rechts am Knopf aus, wenn es so verlangt ist', () => {
    lege({ links: 400, oben: 100, unten: 130, rechts: 440 })

    const menue = zeige('rechts')

    expect(menue.style.left).toBe('240px')
  })

  it('klappt am unteren Fensterrand über den Knopf', () => {
    const hoehe = window.innerHeight
    lege({ links: 40, oben: hoehe - 60, unten: hoehe - 30, rechts: 80 })

    const menue = zeige()

    expect(Number.parseFloat(menue.style.top)).toBe(hoehe - 60 - 4 - 120)
  })
})

// **Beim Rollen und Verkleinern folgt es seinem Knopf** — Befund 5 des
// Zweitblicks über Phase 6 von 0.9.0: sonst stünde es fest, während sein
// Knopf wandert. Und ein langes Menü rollt in sich.
describe('ein offenes Knopfmenü, wenn sich etwas bewegt', () => {
  it('folgt seinem Knopf und rollt in sich', () => {
    const knopf = { links: 40, oben: 100, unten: 130, rechts: 80 }
    lege(knopf)
    const menue = zeige()
    expect(menue).toHaveClass('overflow-y-auto')
    expect(menue.style.top).toBe('134px')

    // **Rollen im Menü misst nicht neu.**
    knopf.oben = 10
    knopf.unten = 40
    act(() => {
      menue.dispatchEvent(new Event('scroll'))
    })
    expect(screen.getByRole('menu').style.top).toBe('134px')

    knopf.oben = 50
    knopf.unten = 80
    act(() => {
      document.dispatchEvent(new Event('scroll'))
    })
    expect(screen.getByRole('menu').style.top).toBe('84px')

    knopf.oben = 70
    knopf.unten = 100
    act(() => {
      window.dispatchEvent(new Event('resize'))
    })
    expect(screen.getByRole('menu').style.top).toBe('104px')
  })

  // **Und ohne Rollen** — Rest R2 des Nachlaufs: eine Liste, die sich
  // umordnet, rückt den Knopf, ohne etwas zu melden.
  it('folgt seinem Knopf, wenn die Seite sich ohne Rollen verschiebt', async () => {
    const knopf = { links: 40, oben: 100, unten: 130, rechts: 80 }
    lege(knopf)
    zeige()

    knopf.oben = 200
    knopf.unten = 230

    await waitFor(() => {
      expect(screen.getByRole('menu').style.top).toBe('234px')
    })
  })
})
