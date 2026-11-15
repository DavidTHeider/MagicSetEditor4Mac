// **Eine Marke, wenn die Kennzahlen stehen** — PERF02
// (#2352).
//
// Die Zahlen erscheinen im Überblick; `kennzahlen09` misst an
// dieser Marke, wie lange Start und Startseite bis zur sichtbaren Kennzahl
// brauchen. **Je Erscheinen eine**, damit ein zweiter Aufruf der Startseite
// seine eigene bekommt.
import { render } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { kennzahlmarke, Ueberblick } from './Ueberblick'

const zahlen = { sets: 2, karten: 7, exemplare: 1, decks: 0 }

function zeige() {
  return render(
    <Sprachraum sprache="de">
      <Ueberblick
        zahlen={zahlen}
        wechseln={() => undefined}
        waehleSet={() => undefined}
        probleme={undefined}
        nachladen={() => undefined}
      />
    </Sprachraum>,
  )
}

afterEach(() => {
  performance.clearMarks(kennzahlmarke)
})

describe('die Marke der Kennzahlen', () => {
  it('setzt eine Marke, sobald der Überblick steht', () => {
    zeige()

    expect(performance.getEntriesByName(kennzahlmarke, 'mark')).toHaveLength(1)
  })

  it('setzt je Erscheinen eine und nicht je Zeichnen', () => {
    const erster = zeige()
    erster.rerender(
      <Sprachraum sprache="de">
        <Ueberblick
          zahlen={{ ...zahlen, karten: 8 }}
          wechseln={() => undefined}
          waehleSet={() => undefined}
          probleme={undefined}
          nachladen={() => undefined}
        />
      </Sprachraum>,
    )
    erster.unmount()
    zeige()

    expect(performance.getEntriesByName(kennzahlmarke, 'mark')).toHaveLength(2)
  })

  // **Nicht, wenn der Platzhalter erscheint** — AP-4 (#2361): seit dem steht
  // der Überblick schon beim Laden da, und die Marke käme sonst vor jeder
  // Zahl. `kennzahlen09` mäße dann den Platzhalter.
  it('setzt sie erst mit den Zahlen und nicht mit dem Platzhalter', () => {
    const ansicht = render(
      <Sprachraum sprache="de">
        <Ueberblick
          zahlen={undefined}
          wechseln={() => undefined}
          waehleSet={() => undefined}
          probleme={undefined}
          nachladen={() => undefined}
        />
      </Sprachraum>,
    )
    expect(performance.getEntriesByName(kennzahlmarke, 'mark')).toHaveLength(0)

    ansicht.rerender(
      <Sprachraum sprache="de">
        <Ueberblick
          zahlen={zahlen}
          wechseln={() => undefined}
          waehleSet={() => undefined}
          probleme={undefined}
          nachladen={() => undefined}
        />
      </Sprachraum>,
    )
    expect(performance.getEntriesByName(kennzahlmarke, 'mark')).toHaveLength(1)
  })
})
