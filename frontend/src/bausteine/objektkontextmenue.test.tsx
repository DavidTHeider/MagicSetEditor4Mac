// **Dieselbe Liste an einer anderen Stelle** — GESTE (#2358): die Seite
// meldet die Handlungen eines Objekts einmal, und der Rechtsklick an seiner
// Vorschau zeigt genau sie.
import { act, cleanup, fireEvent, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Schreibrechtraum } from '../haken/schreibrecht'
import { useHandlungsliste } from '../haken/useHandlungsliste'
import { Objekthandlungsraum } from '../objekthandlungsraum'
import { belegeApi, raeumeApi } from '../test/api'
import type { Menueeintrag } from './Menue'
import { Objektkontextmenue } from './Objektkontextmenue'

afterEach(() => {
  cleanup()
  raeumeApi()
})

function Liste({
  kennung,
  eintraege,
}: {
  kennung: string
  eintraege: readonly Menueeintrag[]
}) {
  useHandlungsliste({ titel: 'Karte', art: 'karte', kennung, eintraege })

  return null
}

function Probe({
  gemeldet,
  angezeigt,
  eintraege,
  nurLesend = false,
}: {
  gemeldet: string
  angezeigt: string
  eintraege: readonly Menueeintrag[]
  nurLesend?: boolean
}) {
  return (
    <Objekthandlungsraum>
      <Schreibrechtraum nurLesend={nurLesend}>
        <Liste kennung={gemeldet} eintraege={eintraege} />
        <Objektkontextmenue art="karte" kennung={angezeigt}>
          <button type="button">Vorschau</button>
        </Objektkontextmenue>
      </Schreibrechtraum>
    </Objekthandlungsraum>
  )
}

describe('Objektkontextmenue', () => {
  it('legt die gemeldete Liste um die Vorschau und löst über sie aus', async () => {
    belegeApi({
      ObjekthandlungenMelden: () => Promise.resolve({ fehler: null }),
    })
    const duplizieren = vi.fn()
    render(
      <Probe
        gemeldet="7"
        angezeigt="7"
        eintraege={[
          {
            schluessel: 'duplizieren',
            name: 'Duplizieren',
            aufWahl: duplizieren,
          },
          { schluessel: 'hoch', name: 'Nach oben', angehalten: true },
        ]}
      />,
    )
    await act(async () => {
      await Promise.resolve()
    })

    fireEvent.contextMenu(screen.getByRole('button', { name: 'Vorschau' }), {
      clientX: 4,
      clientY: 4,
    })
    expect(screen.getByRole('menu', { name: 'Karte' })).toBeInTheDocument()
    expect(screen.getByRole('menuitem', { name: 'Nach oben' })).toHaveAttribute(
      'aria-disabled',
      'true',
    )
    await userEvent.keyboard('{Enter}')

    expect(duplizieren).toHaveBeenCalledTimes(1)
  })

  it('zeigt Schreibendes für einen Leser gesperrt und löst es nicht aus', async () => {
    belegeApi({
      ObjekthandlungenMelden: () => Promise.resolve({ fehler: null }),
    })
    const loeschen = vi.fn()
    render(
      <Probe
        gemeldet="7"
        angezeigt="7"
        nurLesend
        eintraege={[
          {
            schluessel: 'loeschen',
            name: 'Löschen',
            schreibt: true,
            aufWahl: loeschen,
          },
        ]}
      />,
    )
    await act(async () => {
      await Promise.resolve()
    })

    fireEvent.contextMenu(screen.getByRole('button', { name: 'Vorschau' }), {
      clientX: 4,
      clientY: 4,
    })
    const eintrag = screen.getByRole('menuitem', { name: /Löschen/ })
    expect(eintrag).toHaveAttribute('aria-disabled', 'true')
    await userEvent.keyboard('{Enter}')

    expect(loeschen).not.toHaveBeenCalled()
  })

  it('legt nichts um ein anderes Objekt als das gewählte', async () => {
    belegeApi({
      ObjekthandlungenMelden: () => Promise.resolve({ fehler: null }),
    })
    render(
      <Probe
        gemeldet="7"
        angezeigt="8"
        eintraege={[
          { schluessel: 'duplizieren', name: 'Duplizieren', aufWahl: vi.fn() },
        ]}
      />,
    )
    await act(async () => {
      await Promise.resolve()
    })

    const ereignis = fireEvent.contextMenu(
      screen.getByRole('button', { name: 'Vorschau' }),
    )

    expect(ereignis).toBe(true)
    expect(screen.queryByRole('menu')).toBeNull()
  })

  it('steht außerhalb eines Raums für sich', () => {
    render(
      <Objektkontextmenue art="karte" kennung="7">
        <button type="button">Vorschau</button>
      </Objektkontextmenue>,
    )

    expect(screen.getByRole('button', { name: 'Vorschau' })).toBeInTheDocument()
  })
})
