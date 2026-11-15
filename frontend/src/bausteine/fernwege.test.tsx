// **Wie die Gegenüber angebunden sind** — NET-02.AK-03
// (#1706):
// „Direkter Weg, Relay und begründeter Ausfall sind nachvollziehbar".
//
// Der Kasten sagte nur, dass eine Einladung aussteht. Ob ein Gegenüber
// direkt oder über das Relais eines Fremden sprach, sah niemand.
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { Fernhorchposten } from './Fernhorchposten'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import { einsetzen } from '../sprache/einsetzen'

import type { Fernantwort, Fernweg } from '../apipeerkontakt'

const texte = kataloge.de.seiten.technik

function laufend(wege: Fernweg[], angebunden = 0): Fernantwort {
  return {
    stand: { laeuft: true, einladung: 'mse4mac:1:tc1abc', wege, angebunden },
    fehler: null,
  }
}

function zeige() {
  render(
    <Sprachraum sprache="de">
      <Fernhorchposten />
    </Sprachraum>,
  )

  return userEvent.setup()
}

afterEach(() => {
  raeumeApi()
})

it('nennt je Gegenüber den Weg', async () => {
  belegeApi({
    FernHorchStand: () =>
      Promise.resolve(
        laufend([
          { art: 'direkt', adresse: '203.0.113.7:41641' },
          { art: 'relais', adresse: 'fra' },
          { art: 'aufbau', adresse: '' },
        ]),
      ),
  })

  zeige()

  const liste = await screen.findByRole('list', { name: texte.fernwege.titel })
  const zeilen = within(liste).getAllByRole('listitem')

  expect(zeilen.map((zeile) => zeile.textContent)).toEqual([
    einsetzen(texte.fernwege.direkt, { adresse: '203.0.113.7:41641' }),
    einsetzen(texte.fernwege.relais, { adresse: 'fra' }),
    texte.fernwege.aufbau,
  ])
})

// **Angebunden, aber ohne Weg** — SEC-21 (#2162). tailcat v0.6.0 nennt in
// seinem Stand nie eine Gegenstelle; der Kasten sagte deshalb „noch hat
// niemand diese Einladung benutzt", während eine Verbindung offen war.
it('sagt, dass jemand angebunden ist, auch wenn der Weg nicht zu ermitteln ist', async () => {
  belegeApi({ FernHorchStand: () => Promise.resolve(laufend([], 2)) })

  zeige()

  expect(
    await screen.findByText(
      einsetzen(texte.fernwege.unbestimmt, { anzahl: '2' }),
    ),
  ).toBeInTheDocument()
  expect(screen.queryByText(texte.fernwege.keine)).toBeNull()
})

it('sagt, wenn noch niemand die Einladung benutzt hat', async () => {
  belegeApi({ FernHorchStand: () => Promise.resolve(laufend([])) })

  zeige()

  expect(await screen.findByText(texte.fernwege.keine)).toBeInTheDocument()
  expect(screen.queryByRole('list', { name: texte.fernwege.titel })).toBeNull()
})

// **Der Weg ändert sich, während der Kasten offen ist** — aus dem Relais
// wird ein direkter Weg. Gelesen wird auf Knopfdruck, nicht im Hintergrund:
// der Stand fragt tailcat, und ein Takt, der das still täte, liefe auch dann,
// wenn niemand hinsieht.
it('liest die Wege auf Knopfdruck neu', async () => {
  const gerufen = vi
    .fn<() => Promise<Fernantwort>>()
    .mockResolvedValueOnce(laufend([{ art: 'relais', adresse: 'fra' }]))
    .mockResolvedValueOnce(
      laufend([{ art: 'direkt', adresse: '203.0.113.7:41641' }]),
    )
  belegeApi({ FernHorchStand: gerufen })

  const hand = zeige()

  expect(
    await screen.findByText(
      einsetzen(texte.fernwege.relais, { adresse: 'fra' }),
    ),
  ).toBeInTheDocument()

  await hand.click(
    screen.getByRole('button', {
      name: kataloge.de.bausteine.zustaende.neuLaden,
    }),
  )

  expect(
    await screen.findByText(
      einsetzen(texte.fernwege.direkt, { adresse: '203.0.113.7:41641' }),
    ),
  ).toBeInTheDocument()
  expect(gerufen).toHaveBeenCalledTimes(2)
})
