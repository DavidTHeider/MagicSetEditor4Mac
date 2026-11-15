// **Art, Vereinbarung und offener Beschluss am Vorgang** — TRADE-03.AK-02
// (#1717):
// ein Teilen oder eine Deckleihe sagt, dass sie keinen Bestand bewegt; eine
// Einigung sagt, dass sie eine Vereinbarung ist; und „Zugestimmt haben"
// steht nur neben einem Beschluss, dem die Stimmen gelten.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'
import { Handeln } from './Handeln'
import { Sprachraum } from '../sprache/kontext'
import { einKontakt, eineListe } from './peerkontaktprobe'
import { eineUebersicht, eineZeile, einVorgang } from './handelslisteprobe'
import { kataloge } from '../sprache/kataloge'

import type { Handelsvorgang } from '../apihandelsleitung'

const eigene = kataloge.de.seiten.handeln

async function zeige(vorgang: Handelsvorgang) {
  belegeApi({
    Peerkontakte: () =>
      Promise.resolve(
        eineListe([einKontakt({ stand: 'bestaetigt', rechte: ['handel'] })]),
      ),
    HandelUebersicht: () => eineUebersicht([eineZeile({}, vorgang)]),
    HandelVorgaenge: () =>
      Promise.resolve({ vorgaenge: [vorgang], fehler: null }),
  })

  render(
    <Sprachraum sprache="de">
      <Handeln />
    </Sprachraum>,
  )

  await userEvent
    .setup()
    .click(await screen.findByRole('button', { name: /Turfer/ }))
}

describe('Der Stand eines Vorgangs', () => {
  afterEach(() => {
    raeumeApi()
    vi.restoreAllMocks()
  })

  it('nennt eine Deckleihe, die keinen Bestand bewegt, und die Vereinbarung', async () => {
    await zeige(
      einVorgang({ art: 'deckleihe', zustand: 'angenommen', beschluss: '' }),
    )

    expect(
      await screen.findByText(
        `${eigene.vorgangArt}: ${eigene.arten.deckleihe}`,
      ),
    ).toBeInTheDocument()
    expect(screen.getByText(eigene.artOhneBestand)).toBeInTheDocument()
    expect(screen.getByText(/Das ist eine/)).toBeInTheDocument()
    expect(
      screen.queryByText(new RegExp(eigene.vorgangZustimmungen)),
    ).not.toBeInTheDocument()
  })

  it('zeigt die Stimmen neben einem offenen Beschluss', async () => {
    await zeige(
      einVorgang({
        zustand: 'angeboten',
        beschluss: 'zustimmen',
        zustimmungen: ['bich'],
      }),
    )

    expect(
      await screen.findByText(`${eigene.vorgangZustimmungen}: bich`),
    ).toBeInTheDocument()
    expect(screen.queryByText(eigene.artOhneBestand)).not.toBeInTheDocument()
  })

  it('nennt eine unbekannte Art beim Namen', async () => {
    await zeige(einVorgang({ art: 'irgendwas', beschluss: 'zustimmen' }))

    expect(
      await screen.findByText(`${eigene.vorgangArt}: irgendwas`),
    ).toBeInTheDocument()
    expect(
      screen.getByText(`${eigene.vorgangZustimmungen}: ${eigene.vorgangKeine}`),
    ).toBeInTheDocument()
  })
})
