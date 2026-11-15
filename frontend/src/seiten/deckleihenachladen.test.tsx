// **Ein Nachladefehler ist kein gescheiterter Leihschritt** — §7, Regel 3,
// dieselbe Klasse wie bei den Kontakten
// (#1960).
//
// # Wogegen hier geprüft wird
//
// Go meldet nach einem geglückten Schritt, dass die Liste danach nicht zu
// lesen war — `nachladefehler` neben einer leeren Liste. Die Seite nimmt den
// Schritt als geschehen und sagt **an der Liste**, was fehlt. Bis dahin trug
// sie die leere Liste ein und behauptete „keine Leihe".
import { screen, waitFor, within } from '@testing-library/react'
import { expect, it, vi } from 'vitest'

import {
  eigene,
  eineLeihe,
  gewaehlt,
  leiheApi,
  oeffneAnnahme,
  umschlagGewaehlt,
  waehleBertUndDeck,
  waehleUmschlag,
  waehleZiel,
  zeige,
} from '../test/leiheprobe'

// geschrieben ist die Antwort eines Schritts, nach dem nichts zu lesen war.
function geschrieben(pfad = '') {
  return {
    leihen: [],
    pfad,
    nachladefehler: {
      code: 'unbekannt' as const,
      schluessel: '',
      meldung: 'Die Liste klemmt',
    },
    fehler: null,
  }
}

it('nimmt eine Annahme mit Nachladefehler als geschehen', async () => {
  const genommen = vi.fn().mockResolvedValue(geschrieben())

  leiheApi({
    Deckleihen: vi.fn().mockResolvedValue({
      leihen: [eineLeihe()],
      pfad: '',
      fehler: null,
    }),
    DeckleiheAnnehmen: genommen,
    UmschlagWaehlen: vi
      .fn()
      .mockResolvedValue(umschlagGewaehlt('/tmp/fremd.msepackage')),
  })

  const hand = zeige()

  await hand.click(screen.getByRole('button', { name: eigene.leiheAnsehen }))
  expect(
    await screen.findByRole('button', { name: eigene.leiheWiderrufen }),
  ).toBeInTheDocument()

  await oeffneAnnahme(hand)
  await waehleUmschlag(hand, '/tmp/fremd.msepackage')
  await hand.click(screen.getByRole('button', { name: eigene.leiheAnnehmen }))

  // **Der Grund steht an der Liste, und nur dort** — kein Fehler am Schritt,
  // keine Behauptung „keine Leihe", und die alte Zeile ist fort: sie zeigte
  // einen Stand von vorher.
  expect(await screen.findByRole('alert')).toHaveTextContent('Die Liste klemmt')
  expect(screen.getAllByRole('alert')).toHaveLength(1)
  expect(screen.queryByText(eigene.leiheKeine)).not.toBeInTheDocument()
  expect(
    screen.queryByRole('button', { name: eigene.leiheWiderrufen }),
  ).not.toBeInTheDocument()
  expect(genommen).toHaveBeenCalledWith('/tmp/fremd.msepackage')
})

it('nennt den abgelegten Umschlag auch mit Nachladefehler', async () => {
  leiheApi({
    DeckleiheAnbieten: vi
      .fn()
      .mockResolvedValue(geschrieben('/tmp/leihe.msepackage')),
    DeckleiheZielWaehlen: vi
      .fn()
      .mockResolvedValue(gewaehlt('/tmp/leihe.msepackage')),
  })

  const hand = zeige()

  await waehleBertUndDeck(hand)
  await hand.click(screen.getByRole('button', { name: eigene.leiheFachleute }))
  await hand.type(screen.getByLabelText(eigene.leiheWoran), 'p1')
  await waehleZiel(hand, '/tmp/leihe.msepackage')
  await hand.click(screen.getByRole('button', { name: eigene.leiheAnbieten }))

  // **Abgelegt ist abgelegt** — der Pfad steht da, und der Grund daneben.
  await waitFor(() => {
    expect(screen.getByRole('status')).toHaveTextContent(
      '/tmp/leihe.msepackage',
    )
  })
  expect(
    within(screen.getByRole('alert')).getByText('Die Liste klemmt'),
  ).toBeInTheDocument()
})
