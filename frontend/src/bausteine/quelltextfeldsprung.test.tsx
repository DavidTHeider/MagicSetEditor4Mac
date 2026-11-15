import { fireEvent, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Quelltextfeld } from './Quelltextfeld'

import type { Sprungziel } from './quelltextsprung'

function zeige(sprungziel: Sprungziel | null, aufZeiger = vi.fn()) {
  belegeApi({ Skriptfarben: () => Promise.resolve({ merkmale: [] }) })

  const ansicht = render(
    <Sprachraum sprache="de">
      <Quelltextfeld
        bezeichnung="Inhalt"
        wert={'erste\nzweite Zeile\ndritte'}
        setzeWert={vi.fn()}
        sprungziel={sprungziel}
        aufZeiger={aufZeiger}
      />
    </Sprachraum>,
  )

  return {
    ...ansicht,
    aufZeiger,
    feld: screen.getByLabelText<HTMLTextAreaElement>('Inhalt'),
  }
}

describe('Der Sprung im Quelltextfeld', () => {
  afterEach(raeumeApi)

  it('setzt den Zeiger an Zeile und Spalte und holt das Feld in den Fokus', () => {
    const { feld } = zeige({ zeile: 2, spalte: 3, stand: 1 })

    expect(feld).toHaveFocus()
    expect(feld.selectionStart).toBe('erste\n'.length + 2)
    expect(feld.selectionEnd).toBe(feld.selectionStart)
  })

  it('springt bei einem neuen Stand erneut, bei demselben nicht', () => {
    const { feld, rerender } = zeige({ zeile: 1, spalte: 1, stand: 1 })

    feld.setSelectionRange(4, 4)

    rerender(
      <Sprachraum sprache="de">
        <Quelltextfeld
          bezeichnung="Inhalt"
          wert={'erste\nzweite Zeile\ndritte'}
          setzeWert={vi.fn()}
          sprungziel={{ zeile: 1, spalte: 1, stand: 1 }}
          aufZeiger={vi.fn()}
        />
      </Sprachraum>,
    )

    expect(feld.selectionStart).toBe(4)

    rerender(
      <Sprachraum sprache="de">
        <Quelltextfeld
          bezeichnung="Inhalt"
          wert={'erste\nzweite Zeile\ndritte'}
          setzeWert={vi.fn()}
          sprungziel={{ zeile: 3, spalte: 1, stand: 2 }}
          aufZeiger={vi.fn()}
        />
      </Sprachraum>,
    )

    expect(feld.selectionStart).toBe('erste\nzweite Zeile\n'.length)
  })

  it('hält eine Zeile hinter dem Ende am Ende fest', () => {
    const { feld } = zeige({ zeile: 9, spalte: 40, stand: 1 })

    expect(feld.selectionStart).toBe('erste\nzweite Zeile\ndritte'.length)
  })

  it('meldet den Zeiger beim Bewegen und hält ihn beim Verlassen', async () => {
    const { feld, aufZeiger } = zeige(null)

    await userEvent.click(feld)
    feld.setSelectionRange(2, 2)
    fireEvent.select(feld)

    expect(aufZeiger).toHaveBeenLastCalledWith(2)

    // **Beim Verlassen bleibt der letzte Stand gemeldet** — der Inspector
    // braucht ihn noch für den Klick, der das Feld gerade verlassen hat.
    await userEvent.tab()

    expect(aufZeiger).toHaveBeenLastCalledWith(2)
  })
})
