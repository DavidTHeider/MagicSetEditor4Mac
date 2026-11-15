// **Ein neuer Tausch mit einem Klick** — P12
// (#2380),
// Befund C-18 und H1 ≤ 5: unter „Neuer Tausch …" stehen die Gegenüber, mit
// denen gehandelt werden darf, als eigene Knöpfe.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Tauschpartnerleiste } from './Tauschpartnerleiste'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'
import { belegeApi, raeumeApi } from '../test/api'

import type { Peerkontakt } from '../apipeerkontakt'

const texte = kataloge.de.seiten.handeln.markttisch

afterEach(() => {
  raeumeApi()
})

const kontakt = (
  kennung: string,
  name: string,
  teil: Partial<Peerkontakt> = {},
) =>
  ({
    kennung,
    fingerabdruck: '',
    selbstgewaehlterName: name,
    eigenerName: '',
    stand: 'bestaetigt',
    gesehen: '',
    rechte: ['handel'],
    adresse: '',
    ...teil,
  }) as Peerkontakt

describe('die Leiste der Tauschpartner', () => {
  it('nennt jedes Gegenüber, mit dem gehandelt werden darf, und schlägt es auf', async () => {
    belegeApi({
      Peerkontakte: () =>
        Promise.resolve({
          kontakte: [
            kontakt('k-bert', 'Bert', { eigenerName: 'Bert B.' }),
            kontakt('k-cleo', 'Cleo'),
            kontakt('k-dora', 'Dora', { stand: 'unbestaetigt' }),
            kontakt('k-emil', 'Emil', { rechte: ['partie'] }),
          ],
          fehler: null,
        }),
    } as never)
    const waehlen = vi.fn()

    render(
      <Sprachraum sprache="de">
        <Tauschpartnerleiste waehlen={waehlen} />
      </Sprachraum>,
    )

    const bert = await screen.findByRole('button', {
      name: texte.neuMit.replace('{name}', 'Bert B.'),
    })
    expect(
      screen.getByRole('button', {
        name: texte.neuMit.replace('{name}', 'Cleo'),
      }),
    ).toBeVisible()

    // **Nur wer bestätigt ist und handeln darf.**
    expect(screen.queryByText('Dora')).toBeNull()
    expect(screen.queryByText('Emil')).toBeNull()

    await userEvent.click(bert)
    expect(waehlen).toHaveBeenCalledWith('k-bert', 'Bert B.')
  })

  it('steht ab vier solchen Gegenübern nicht da — dann hilft die Schublade', async () => {
    const gefragt = vi.fn(() =>
      Promise.resolve({
        kontakte: [
          kontakt('k-bert', 'Bert'),
          kontakt('k-cleo', 'Cleo'),
          kontakt('k-dora', 'Dora'),
          kontakt('k-emil', 'Emil'),
        ],
        fehler: null,
      }),
    )
    belegeApi({ Peerkontakte: gefragt } as never)

    const { container } = render(
      <Sprachraum sprache="de">
        <Tauschpartnerleiste waehlen={vi.fn()} />
      </Sprachraum>,
    )

    // **Das Maß aus UXR-15** (`handelnmass.test.tsx`): mit zehn Gegenübern
    // stünden sonst fünf Knöpfe mehr links, und welche fünf, wäre Zufall.
    await vi.waitFor(() => {
      expect(gefragt).toHaveBeenCalled()
    })
    await new Promise((weiter) => setTimeout(weiter, 0))
    expect(container.querySelector('button')).toBeNull()
  })

  it('steht ohne ein solches Gegenüber gar nicht da', async () => {
    const gefragt = vi.fn(() => Promise.resolve({ kontakte: [], fehler: null }))
    belegeApi({ Peerkontakte: gefragt } as never)

    const { container } = render(
      <Sprachraum sprache="de">
        <Tauschpartnerleiste waehlen={vi.fn()} />
      </Sprachraum>,
    )

    await vi.waitFor(() => {
      expect(gefragt).toHaveBeenCalled()
    })
    expect(container.querySelector('button')).toBeNull()
  })
})
