// **Nach der Herkunft filtern** — DOM-01.AK-01
// (#1696):
// das Artboard `Pakete` zeichnet die Herkunft als Knöpfe im Navigator.
//
// **Seit UXR-17 ein Auswahlfeld** (#1925): fünf Werte sind mehr, als R7 aus
// `docs/UIREVIEW.md` Segmenten zugesteht (bis vier), und fünf Knöpfe im
// Navigator zählten im Bild mit (R2). Daneben steht die Art schon so.
import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Paketfilter } from './Paketfilter'
import type { Sichtwahl } from './paketsicht'

afterEach(cleanup)

const t = kataloge.de.seiten.pakete.plan

describe('Paketfilter', () => {
  it('wählt eine Herkunft und wieder alle', async () => {
    const aendern = vi.fn()
    const wahl: Sichtwahl = {
      suche: '',
      art: '',
      herkunft: 'eigen',
      ordnung: 'gruppen',
    }
    render(<Paketfilter wahl={wahl} arten={['mse-game']} aendern={aendern} />)

    const feld = screen.getByRole('combobox', { name: t.herkunft })
    expect(feld).toHaveValue('eigen')
    expect(
      screen.queryByRole('button', { name: t.herkunftsarten.eigen }),
    ).toBeNull()

    await userEvent.selectOptions(feld, t.herkunftsarten.bezogen)
    expect(aendern).toHaveBeenLastCalledWith({ ...wahl, herkunft: 'bezogen' })

    await userEvent.selectOptions(feld, t.herkunftAlle)
    expect(aendern).toHaveBeenLastCalledWith({ ...wahl, herkunft: '' })
  })
})
