// **Die Schublade geht für das Symbol weit auf** — AP-12
// (#2369),
// Befund B-26. Die Schublade hält den Wunsch, die Seite darin bietet ihn an
// (`haken/useSchubladenweite.ts`); weit wird sie nur, solange das Symbol
// offen ist — Stil und Setangaben brauchen die Breite nicht.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import { useSchubladenweite } from '../haken/useSchubladenweite'
import { Sprachraum } from '../sprache/kontext'
import { Setschublade } from './Setschublade'

import type { Setschluessel } from '../navigation'
import type { Schublade } from '../schubladen'

const eigenschaften: Schublade = {
  schluessel: 'eigenschaften',
  laschen: ['stil', 'setangaben', 'symbol', 'schluesselwoerter'],
}

// Probe spielt die Seite in der Schublade: sie zeigt den Stand und schaltet.
function Probe() {
  const weite = useSchubladenweite()

  return (
    <button type="button" onClick={weite?.umschalten}>
      {weite?.weit === true ? 'weit' : 'schmal'}
    </button>
  )
}

function zeige(aktuell: Setschluessel) {
  const ansicht = render(
    <Sprachraum sprache="de">
      <Setschublade
        schublade={eigenschaften}
        aktuell={aktuell}
        wechseln={vi.fn()}
        schliessen={vi.fn()}
      >
        <Probe />
      </Setschublade>
    </Sprachraum>,
  )

  return ansicht
}

function breite(): string {
  return (
    document
      .querySelector<HTMLElement>('.mantine-Drawer-root')
      ?.style.getPropertyValue('--drawer-size') ?? ''
  )
}

describe('die Weite der Setschublade', () => {
  it('geht beim Symbol auf Wunsch weit auf', async () => {
    zeige('symbol')
    const schmal = breite()
    expect(schmal).toContain('28.75rem')

    await userEvent.click(screen.getByRole('button', { name: 'schmal' }))

    expect(screen.getByRole('button', { name: 'weit' })).toBeInTheDocument()
    expect(breite()).toContain('56rem')
  })

  // **Der Wunsch gilt dem Symbol** — eine andere Lasche bleibt schmal, auch
  // wenn er gesetzt ist; zurück beim Symbol ist er wieder da.
  it('bleibt bei einer anderen Lasche schmal und merkt sich den Wunsch', async () => {
    const { rerender } = zeige('symbol')
    await userEvent.click(screen.getByRole('button', { name: 'schmal' }))

    rerender(
      <Sprachraum sprache="de">
        <Setschublade
          schublade={eigenschaften}
          aktuell="stil"
          wechseln={vi.fn()}
          schliessen={vi.fn()}
        >
          <Probe />
        </Setschublade>
      </Sprachraum>,
    )
    expect(breite()).toContain('28.75rem')

    rerender(
      <Sprachraum sprache="de">
        <Setschublade
          schublade={eigenschaften}
          aktuell="symbol"
          wechseln={vi.fn()}
          schliessen={vi.fn()}
        >
          <Probe />
        </Setschublade>
      </Sprachraum>,
    )
    expect(breite()).toContain('56rem')
  })
})
