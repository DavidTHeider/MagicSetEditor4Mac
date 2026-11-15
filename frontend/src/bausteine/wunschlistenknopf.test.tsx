// **Was einem Deck fehlt, auf die Wunschliste** — DOM-03.AK-02
// (#1698).
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Wunschlistenknopf } from './Wunschlistenknopf'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { deckprobe } from '../seiten/deckprobe'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'

import type { Api } from '../api'

const texte = kataloge.de.seiten.decks

function zeige(
  DeckFehlkartenWuenschen: Api['DeckFehlkartenWuenschen'],
  fehlend = 2,
) {
  const fuehre = vi.fn(
    async (_schluessel: string, arbeit: () => Promise<void>) => {
      await arbeit()
    },
  )

  belegeApi({ DeckFehlkartenWuenschen })

  render(
    <Sprachraum sprache="de">
      <Wunschlistenknopf
        stand={{ fuehre, vorgang: { laeuft: () => false } } as never}
        deck={deckprobe({ id: '7' })}
        fehlend={fehlend}
      />
    </Sprachraum>,
  )

  return fuehre
}

afterEach(() => {
  raeumeApi()
})

describe('Fehlkarten auf die Wunschliste', () => {
  it('setzt sie darauf und sagt, was neu und was schon da war', async () => {
    const wuenschen = vi.fn(() =>
      Promise.resolve({
        neu: [],
        vorhanden: [],
        unbekannt: 0,
        fehler: null,
      }),
    )
    wuenschen.mockResolvedValueOnce({
      neu: [{} as never, {} as never],
      vorhanden: [{} as never],
      unbekannt: 0,
      fehler: null,
    })

    const fuehre = zeige(wuenschen)

    expect(screen.queryByRole('status')).toBeNull()

    await userEvent.click(
      screen.getByRole('button', { name: texte.aufWunschliste }),
    )

    expect(await screen.findByRole('status')).toHaveTextContent(
      einsetzen(texte.wunschlisteErgebnis, { neu: '2', schon: '1' }),
    )
    expect(wuenschen).toHaveBeenCalledWith('7')
    expect(fuehre).toHaveBeenCalledWith('zuordnung:7', expect.any(Function))
  })

  // **Wenn nichts fehlt, gibt es nichts zu wünschen** — der Knopf stünde da
  // und täte nichts.
  it('steht nicht da, wenn nichts fehlt', () => {
    zeige(vi.fn(), 0)

    expect(
      screen.queryByRole('button', { name: texte.aufWunschliste }),
    ).toBeNull()
  })
})
