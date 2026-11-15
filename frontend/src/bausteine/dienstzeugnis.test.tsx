// **Welches Zeugnis ein Dienst vorzeigt** — NET-03
// (#1707).
//
// Mit einer eigenen PKI nennt der Posten die Stelle, die das Zeugnis
// ausgestellt hat, und wie lange es gilt. Beim Zeugnis auf der Peeridentität
// steht beides nicht da — es gibt dort keine Stelle und kein Ende.
import { render, screen } from '@testing-library/react'
import { afterEach, expect, it } from 'vitest'

import { Dienstposten } from './Dienstposten'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import { einsetzen } from '../sprache/einsetzen'
import { alsZeitpunkt } from '../zeitpunkt'

import type { DienstAntwort } from '../apidienste'

const eigene = kataloge.de.seiten.technik

function laufend(aussteller: string, bis: string): DienstAntwort {
  return {
    stand: {
      art: 'rest',
      laeuft: true,
      adresse: 'https://[::]:8443',
      tls: true,
      marke: 'die-marke',
      namen: ['dienst.haus.test'],
      adressen: [],
      fingerabdruck: 'AB:CD',
      zeugnisAussteller: aussteller,
      zeugnisBis: bis,
      lauscher: ['https://[::]:8443'],
      fehlende: [],
    },
    fehler: null,
  }
}

afterEach(() => {
  raeumeApi()
})

function zeige(antwort: DienstAntwort) {
  belegeApi({ RESTStand: () => Promise.resolve(antwort) })

  render(
    <Sprachraum sprache="de">
      <Dienstposten art="rest" />
    </Sprachraum>,
  )
}

it('nennt Stelle und Ende eines eigenen Zeugnisses', async () => {
  zeige(laufend('Haus-Stelle', '2027-01-02T03:04:05Z'))

  expect(
    await screen.findByText(
      einsetzen(eigene.dienstzeugnis.eigenes, {
        aussteller: 'Haus-Stelle',
        bis: alsZeitpunkt('2027-01-02T03:04:05Z', 'de'),
      }),
    ),
  ).toBeInTheDocument()
})

it('nennt beim Zeugnis der Identität keine Stelle', async () => {
  zeige(laufend('', ''))

  // **Die Gegenprobe, dass der Posten steht** (§3): der Abdruck ist da.
  expect(await screen.findByText('AB:CD')).toBeInTheDocument()
  expect(screen.queryByText(/Haus-Stelle/)).toBeNull()
  expect(screen.getByText(eigene.dienstzeugnis.identitaet)).toBeInTheDocument()
})
