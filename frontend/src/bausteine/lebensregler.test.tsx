// **Leben am Sitz** — P4
// (#2372),
// Befund C-20: Lebenspunkte gingen nur über ein Formular. Jetzt −/+ (⇧ für
// fünf) und „Wert setzen“, alles über `spieler.leben` mit einer Änderung —
// der Kern rechnet, und die Absenderregel bleibt dort.
import { fireEvent, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { einsetzen } from '../sprache/einsetzen'
import { Lebensregler } from './Lebensregler'

const eigene = kataloge.de.seiten.spielen

function zeige(wert = 20) {
  const aendere = vi.fn()
  render(
    <Sprachraum sprache="de">
      <Lebensregler wert={wert} aendere={aendere} />
    </Sprachraum>,
  )

  return aendere
}

describe('der Lebensregler', () => {
  it('ändert um eins, mit ⇧ um fünf', () => {
    const aendere = zeige()

    fireEvent.click(screen.getByRole('button', { name: eigene.lebenWeniger }))
    fireEvent.click(screen.getByRole('button', { name: eigene.lebenMehr }))
    fireEvent.click(screen.getByRole('button', { name: eigene.lebenWeniger }), {
      shiftKey: true,
    })
    fireEvent.click(screen.getByRole('button', { name: eigene.lebenMehr }), {
      shiftKey: true,
    })

    expect(aendere.mock.calls).toEqual([[-1], [1], [-5], [5]])
  })

  // **Gesetzt wird als Änderung** — „siebzehn“ überschriebe eine Antwort, die
  // sich mit einer anderen kreuzt (`spieler.leben`, §7 Regel 2).
  it('setzt einen Wert als Änderung', async () => {
    const aendere = zeige(20)
    const hand = userEvent.setup()

    await hand.click(
      screen.getByRole('button', {
        name: einsetzen(eigene.lebenSetzen, { stand: '20' }),
      }),
    )
    const feld = screen.getByLabelText(eigene.lebenNeu)
    await hand.clear(feld)
    await hand.type(feld, '17')
    await hand.click(
      screen.getByRole('button', { name: eigene.lebenUebernehmen }),
    )

    expect(aendere).toHaveBeenCalledWith(-3)
    expect(screen.queryByLabelText(eigene.lebenNeu)).toBeNull()
  })

  it('setzt nichts, wenn der Wert derselbe bleibt', async () => {
    const aendere = zeige(20)
    const hand = userEvent.setup()

    await hand.click(
      screen.getByRole('button', {
        name: einsetzen(eigene.lebenSetzen, { stand: '20' }),
      }),
    )
    await hand.click(
      screen.getByRole('button', { name: eigene.lebenUebernehmen }),
    )

    expect(aendere).not.toHaveBeenCalled()
    expect(screen.queryByLabelText(eigene.lebenNeu)).toBeNull()
  })
})
