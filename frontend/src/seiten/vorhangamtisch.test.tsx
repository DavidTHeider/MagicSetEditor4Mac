// **Der Vorhang am Tisch** — P7
// (#2375),
// Artboard `Vorhang09`: der Tisch ruht, **beide Hände verdeckt**, bis der
// neue Sitz bestätigt; ↩ bestätigt, Esc lässt den bisherigen weiterhandeln.
//
// # Wogegen
//
// **Gegen P6**: nach Annas „Behalten“ wechselte die Sicht ohne Nachfrage zu
// Bert — und Berts Hand lag offen vor Anna. Davor suchte man ihn in der
// Sichtwahl, und Annas Hand lag offen vor Bert.
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import nutzerereignis from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { Partietisch } from './Partietisch'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { einLauf, einObjekt, einePartie, einSitz } from './partieprobe'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'

import type { Partie, Sitz } from '../apipartie'
import type { Partielauf } from '../haken/usePartie'

afterEach(raeumeApi)

const eigene = kataloge.de.seiten.spielen

function beginn(
  anna: Sitz['eroeffnung'],
  bert: Sitz['eroeffnung'],
  fuer = 'sitz1',
): Partie {
  return einePartie({
    fuer,
    sitze: [
      einSitz({ kennung: 'sitz1', name: 'Anna', eroeffnung: anna }),
      einSitz({ kennung: 'sitz2', name: 'Bert', eroeffnung: bert }),
    ],
    objekte: [
      einObjekt({
        kennung: 'sitz1/h0',
        name: 'Glutläufer',
        zone: 'hand',
        beherrscher: 'sitz1',
        geheim: false,
      }),
    ],
  })
}

// zeige stellt den Tisch an ein Gerät ohne Gäste — die Freigabe ist zu.
function zeige(lauf: Partielauf, partie: Partie) {
  belegeApi({
    PartieFreigabeStand: () =>
      Promise.resolve({
        laeuft: false,
        partie: 'p1',
        wirt: 'sitz1',
        vergeben: [],
        fehler: null,
      }),
  })
  const bild = (stand: Partie) => (
    <Sprachraum sprache="de">
      <Partietisch lauf={{ ...lauf, partie: stand }} partie={stand} />
    </Sprachraum>
  )
  const ansicht = render(bild(partie))
  fireEvent.click(screen.getByRole('button', { name: eigene.tischListe }))

  return {
    hand: nutzerereignis.setup(),
    neu: (stand: Partie) => {
      ansicht.rerender(bild(stand))
    },
  }
}

const vorhangFuer = (wer: string) =>
  screen.findByRole('region', {
    name: einsetzen(eigene.vorhangTitel, { wer }),
  })

it('verdeckt beide Hände, bis der neue Sitz bestätigt', async () => {
  const ansehen = vi.fn()
  const { hand, neu } = zeige(einLauf({ ansehen }), beginn('offen', 'offen'))

  expect(screen.getByText('Glutläufer')).toBeInTheDocument()

  neu(beginn('behalten', 'offen'))
  await vorhangFuer('Bert')

  // **Nichts vom vorigen Sitz** — weder seine Hand noch die Tafel des
  // Beginns, und die Führung des Wirts mit der Rückschau ruht mit.
  expect(screen.queryByText('Glutläufer')).toBeNull()
  expect(screen.queryByRole('region', { name: eigene.beginnTitel })).toBeNull()
  expect(screen.queryByRole('button', { name: eigene.pausieren })).toBeNull()
  expect(ansehen).not.toHaveBeenCalled()

  const ich = screen.getByRole('button', {
    name: einsetzen(eigene.vorhangBestaetigen, { wer: 'Bert' }),
  })
  await waitFor(() => {
    expect(ich).toHaveFocus()
  })

  await hand.keyboard('{Enter}')

  expect(ansehen).toHaveBeenCalledExactlyOnceWith('p1', 'sitz2')
})

it('lässt mit Esc den bisherigen Sitz weiterhandeln', async () => {
  const ansehen = vi.fn()
  const { hand, neu } = zeige(einLauf({ ansehen }), beginn('offen', 'offen'))

  neu(beginn('behalten', 'offen'))
  await vorhangFuer('Bert')
  await waitFor(() => {
    expect(
      screen.getByRole('button', {
        name: einsetzen(eigene.vorhangBestaetigen, { wer: 'Bert' }),
      }),
    ).toHaveFocus()
  })

  await hand.keyboard('{Escape}')

  expect(await screen.findByText('Glutläufer')).toBeInTheDocument()
  expect(screen.queryByRole('region', { name: /Bert/ })).toBeNull()
  expect(ansehen).not.toHaveBeenCalled()
})

// **Offene Hände sind eine ausdrückliche Wahl** — dann folgt die Sicht.
it('folgt mit „Offenen Händen“ ohne Vorhang', async () => {
  const ansehen = vi.fn()
  const { hand, neu } = zeige(einLauf({ ansehen }), beginn('offen', 'offen'))

  const schalter = screen.getByRole('checkbox', { name: eigene.offeneHaende })
  expect(schalter).not.toBeChecked()
  await hand.click(schalter)

  neu(beginn('behalten', 'offen'))

  await waitFor(() => {
    expect(ansehen).toHaveBeenCalledExactlyOnceWith('p1', 'sitz2')
  })
  expect(
    screen.queryByRole('region', {
      name: einsetzen(eigene.vorhangTitel, { wer: 'Bert' }),
    }),
  ).toBeNull()
})

// **Hinter dem Vorhang ist nichts gewählt** — eine gewählte Handkarte nennte
// ihren Namen in der Leiste, und nach dem Vorhang stünde sie noch da.
it('nimmt die Wahl einer Karte mit', async () => {
  const { hand, neu } = zeige(einLauf(), beginn('offen', 'offen'))

  await hand.click(screen.getByRole('button', { name: /Glutläufer/ }))
  expect(screen.getByRole('button', { name: /Glutläufer/ })).toHaveAttribute(
    'aria-pressed',
    'true',
  )

  neu(beginn('behalten', 'offen'))
  await vorhangFuer('Bert')
  await hand.keyboard('{Escape}')

  expect(
    await screen.findByRole('button', { name: /Glutläufer/ }),
  ).toHaveAttribute('aria-pressed', 'false')
})

// **Und sie lässt sich zurücknehmen** — dann geht der Vorhang wieder zu.
it('zieht den Vorhang wieder, wenn „Offene Hände“ abgewählt ist', async () => {
  const ansehen = vi.fn()
  const { hand, neu } = zeige(einLauf({ ansehen }), beginn('offen', 'offen'))

  const schalter = screen.getByRole('checkbox', { name: eigene.offeneHaende })
  await hand.click(schalter)
  await hand.click(schalter)
  expect(schalter).not.toBeChecked()

  neu(beginn('behalten', 'offen'))

  await vorhangFuer('Bert')
  expect(ansehen).not.toHaveBeenCalled()
})
