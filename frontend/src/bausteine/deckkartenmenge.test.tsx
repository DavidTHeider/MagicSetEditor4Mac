// **Die Menge einer Karte im Deck ist eine ganze Zahl ab eins** — F03
// (#2245).
//
// Bis dahin las `Number(menge)` das Feld beim Legen: ein leeres wurde null,
// „2.5" ging als Kommazahl an ein `int` in Go, und die Absage kam als Satz
// über JSON statt am Feld.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import { Anschlussraum, useAnschluss } from '../anschlussraum'
import { Deckkartenwahl } from './Deckkartenwahl'
import { Sprachraum } from '../sprache/kontext'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'
import { belegeApi } from '../test/api'

const texte = kataloge.de.seiten.decks

// Abgeber gibt einen Druck herüber, damit die Wahl bereit ist.
function Abgeber() {
  const anschluss = useAnschluss()

  return (
    <button
      type="button"
      onClick={() => {
        anschluss.uebergib({
          art: 'deck',
          name: 'Aschenwächter',
          printing: 'p-4',
        })
      }}
    >
      abgeben
    </button>
  )
}

async function zeige() {
  const legen = vi.fn()

  belegeApi({
    Sets: () => Promise.resolve({ sets: [], fehler: null }),
    Kartenzeilen: () => Promise.resolve({ zeilen: [], fehler: null }),
  })

  render(
    <Sprachraum sprache="de">
      <Anschlussraum aktuell="decks" wechseln={() => undefined}>
        <Abgeber />
        <Deckkartenwahl spiel="magic" angehalten={false} legen={legen} />
      </Anschlussraum>
    </Sprachraum>,
  )

  await userEvent.click(screen.getByRole('button', { name: 'abgeben' }))

  return legen
}

describe('die Menge beim Legen', () => {
  it.each(['0', '2.5', '-1'])(
    'hält das Legen bei %j am Feld an',
    async (roh) => {
      const legen = await zeige()
      const menge = screen.getByLabelText(texte.menge)

      await userEvent.clear(menge)
      await userEvent.type(menge, roh)

      expect(menge).toHaveAttribute('aria-invalid', 'true')
      expect(menge).toHaveAccessibleDescription(
        einsetzen(kataloge.de.allgemein.ganzeZahlAb, { kleinstens: '1' }),
      )
      expect(screen.getByRole('button', { name: texte.legen })).toBeDisabled()
      expect(legen).not.toHaveBeenCalled()
    },
  )

  it('hält das Legen an, solange das Feld leer ist', async () => {
    await zeige()

    await userEvent.clear(screen.getByLabelText(texte.menge))

    expect(screen.getByRole('button', { name: texte.legen })).toBeDisabled()
  })

  it('legt mit der getippten Menge', async () => {
    const legen = await zeige()
    const menge = screen.getByLabelText(texte.menge)

    await userEvent.clear(menge)
    await userEvent.type(menge, '3')
    await userEvent.click(screen.getByRole('button', { name: texte.legen }))

    expect(legen).toHaveBeenCalledWith('', 'p-4', 3, expect.any(String))
  })
})
