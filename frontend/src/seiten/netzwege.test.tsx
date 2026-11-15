// **Die zwei kurzen Wege auf der Deckseite** — P9c
// (#2377),
// Artboard `Netzpartie09`: „Bo einladen“ (S12) und „Bei Bo mitspielen“
// (S13) für das erste Gegenüber mit dem Recht am Tisch — mit dem eigenen
// Deck, mit dem zuletzt gearbeitet wurde.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it } from 'vitest'

import { Anschlussraum, useAnschluss } from '../anschlussraum'
import { Netzwege } from './Netzwege'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'
import { einKontakt, eineListe } from './peerkontaktprobe'
import { belegeApi, raeumeApi } from '../test/api'

afterEach(raeumeApi)

const eigene = kataloge.de.seiten.spielen

function Zeiger() {
  const anschluss = useAnschluss()

  return <output>{JSON.stringify(anschluss.offen ?? null)}</output>
}

function zeige(decks: { id: string }[]) {
  belegeApi({
    Peerkontakte: () =>
      Promise.resolve(
        eineListe([
          einKontakt({
            kennung: 'p-bo',
            eigenerName: 'Bo',
            stand: 'bestaetigt',
            rechte: ['partie'],
          }),
          einKontakt({
            kennung: 'p-cy',
            eigenerName: 'Cy',
            stand: 'bestaetigt',
            rechte: ['partie'],
          }),
        ]),
      ),
    GezeigtesProfil: () => Promise.resolve({ profil: null, fehler: null }),
    Weiterarbeiten: () =>
      Promise.resolve({
        geoeffnet: [{ objekt: { art: 'deck', kennung: '2' } }],
        fehler: null,
      }),
  } as never)

  render(
    <Sprachraum sprache="de">
      <Anschlussraum aktuell="decks" wechseln={() => undefined}>
        <Netzwege decks={decks} />
        <Zeiger />
      </Anschlussraum>
    </Sprachraum>,
  )
}

it('lädt das erste Gegenüber mit dem eigenen Deck ein', async () => {
  zeige([{ id: '1' }, { id: '2' }])

  await userEvent.click(
    await screen.findByRole('button', {
      name: eigene.netzEinladen.replace('{name}', 'Bo'),
    }),
  )

  expect(JSON.parse(screen.getByRole('status').textContent)).toEqual({
    art: 'partie',
    name: 'Bo',
    deckID: '2',
    sofort: true,
    eingeladen: { kennung: 'p-bo', name: 'Bo' },
  })
  expect(
    screen.queryByRole('button', {
      name: eigene.netzEinladen.replace('{name}', 'Cy'),
    }),
  ).toBeNull()
})

it('setzt sich mit dem eigenen Deck an den Tisch des ersten Gegenübers', async () => {
  zeige([{ id: '1' }, { id: '2' }])

  await userEvent.click(
    await screen.findByRole('button', {
      name: eigene.netzMitspielen.replace('{name}', 'Bo'),
    }),
  )

  expect(JSON.parse(screen.getByRole('status').textContent)).toEqual({
    art: 'mitspielen',
    name: 'Bo',
    gegenueber: 'p-bo',
    deckID: '2',
  })
})

// **Ohne eigenes Deck lädt niemand ein** — am ersten Sitz läge keines.
it('lädt ohne eigenes Deck nicht ein, spielt aber mit', async () => {
  zeige([])

  expect(
    await screen.findByRole('button', {
      name: eigene.netzMitspielen.replace('{name}', 'Bo'),
    }),
  ).toBeInTheDocument()
  expect(
    screen.queryByRole('button', {
      name: eigene.netzEinladen.replace('{name}', 'Bo'),
    }),
  ).toBeNull()
})
