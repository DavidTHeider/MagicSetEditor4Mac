// **Ein Set ohne Karte sagt es, bevor ausgegeben wird** — G02
// (#2249).
//
// Bilder und Druckbogen eines leeren Sets zeigten „0 Karten" an einer
// Auswahl, die sich nicht bedienen ließ, und sonst nichts — kein Satz, warum,
// und kein Weg zu den Karten.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Ausgeben } from './Ausgeben'

const texte = kataloge.de

afterEach(() => {
  raeumeApi()
})

function satz(karten: number) {
  return {
    id: '4',
    name: 'Leer',
    spiel: 'magic',
    stil: 'm15',
    herkunft: '',
    karten,
    herkunftFehlt: false,
  }
}

describe('Ausgeben ohne Karte', () => {
  it('nennt das leere Set und führt zu den Karten', async () => {
    belegeApi({ Sets: () => Promise.resolve({ sets: [satz(0)] }) })

    const wechseln = vi.fn()

    render(
      <Sprachraum sprache="de">
        <Ausgeben setID="4" seite="ausgeben" wechseln={wechseln} />
      </Sprachraum>,
    )

    expect(
      await screen.findByText(texte.allgemein.arbeitsplatz.ohneKartenTitel),
    ).toBeInTheDocument()

    await userEvent.click(
      screen.getByRole('button', {
        name: texte.allgemein.arbeitsplatz.zuDenKarten,
      }),
    )

    expect(wechseln).toHaveBeenCalledWith('karten')
  })
})
