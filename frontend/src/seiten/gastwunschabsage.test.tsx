// **Ein abgewiesener Wunsch steht am Tisch** —
// #2478.
//
// Bis dahin endete die Sitzung an jedem Nein des Kerns, und der Gast saß an
// keinem Tisch mehr. Jetzt schickt der Wirt sein Bild mit dem Grund, und das
// Fenster sagt ihn — der Wunsch verschwindet nicht still (Artboard `Tisch09`,
// „Ablehnung“).
import { screen, waitFor } from '@testing-library/react'
import { expect, it, vi } from 'vitest'

import { einePartie } from './partieprobe'
import { einsetzen } from '../sprache/einsetzen'
import { ereignisTischbild } from '../haken/usePartiegast'
import { kataloge } from '../sprache/kataloge'
import {
  gastApi,
  laufzeitbelegen,
  waehleWirt,
  zeigeGast,
} from './partiegastprobe'

const eigene = kataloge.de.seiten.spielen

it('sagt den Grund, mit dem der Wirt einen Wunsch abwies', async () => {
  const horcher = laufzeitbelegen()
  gastApi({
    PartieBeitreten: vi.fn().mockResolvedValue({
      laeuft: true,
      partie: 'p1',
      gegenueber: 'abcd',
      pausiert: false,
      absage: '',
      fehler: null,
    }),
  })

  const hand = zeigeGast()
  await waehleWirt(hand)
  await hand.click(screen.getByRole('button', { name: eigene.gastBeitreten }))
  await screen.findByText(eigene.gastSitzt, { exact: false })

  const satz = einsetzen(eigene.gastWunschAbgewiesen, {
    grund: 'Konflikt: die Fassung gilt nicht mehr',
  })

  await waitFor(() => {
    horcher.get(ereignisTischbild)?.('p1', {
      partie: einePartie(),
      erlaubt: [],
      handlungen: [],
      absage: {
        befehl: 'b1',
        art: 'objekt.tappen',
        grund: 'Konflikt: die Fassung gilt nicht mehr',
      },
    })
  })
  expect((await screen.findAllByText(satz)).length).toBeGreaterThan(0)

  // **Das nächste Bild ohne Absage nimmt den Satz wieder weg.**
  horcher.get(ereignisTischbild)?.('p1', {
    partie: einePartie(),
    erlaubt: [],
    handlungen: [],
  })
  await waitFor(() => {
    expect(screen.queryByText(satz)).not.toBeInTheDocument()
  })
})
