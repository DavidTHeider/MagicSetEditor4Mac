// **Die kurzen Wege der Netzpartie** — P9c
// (#2377),
// Artboard `Netzpartie09`: S12 ≤ 3 und S13 ≤ 2. Von der Deckseite reist eine
// Einladung oder ein Beitritt zur Spielen-Seite — **einmal**, und mit dem,
// was mitgeht.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { afterEach, expect, it, vi } from 'vitest'

import { Anschlussraum, useAnschluss } from '../anschlussraum'
import { Spielen } from './Spielen'
import { Sprachraum } from '../sprache/kontext'
import { eineAntwort } from './partieprobe'
import { mitDecks } from './spielenprobe'
import { raeumeApi } from '../test/api'

import type { Anschluss } from '../anschluss'
import type { Seitenschluessel } from '../navigation'

afterEach(raeumeApi)

function Buehne({ schicke }: { schicke: Anschluss }) {
  const [aktuell, setzeAktuell] = useState<Seitenschluessel>('decks')

  return (
    <Sprachraum sprache="de">
      <Anschlussraum aktuell={aktuell} wechseln={setzeAktuell}>
        <Abgeber schicke={schicke} />
        <Spielen wechseln={() => undefined} />
      </Anschlussraum>
    </Sprachraum>
  )
}

function Abgeber({ schicke }: { schicke: Anschluss }) {
  const anschluss = useAnschluss()

  return (
    <button
      type="button"
      onClick={() => {
        anschluss.uebergib(schicke)
      }}
    >
      los
    </button>
  )
}

it('fängt mit einer Einladung am zweiten Sitz an', async () => {
  const angelegt = vi.fn(() => Promise.resolve(eineAntwort()))
  mitDecks({
    GezeigtesProfil: () =>
      Promise.resolve({
        profil: { id: '7', nutzerID: '1', name: 'Ada' },
        fehler: null,
      }),
    PartieAnlegen: angelegt,
  })
  render(
    <Buehne
      schicke={{
        art: 'partie',
        name: 'Bo',
        deckID: '2',
        sofort: true,
        eingeladen: { kennung: 'p-bo', name: 'Bo' },
      }}
    />,
  )

  await userEvent.click(await screen.findByRole('button', { name: 'los' }))

  await waitFor(() => {
    expect(angelegt).toHaveBeenCalledTimes(1)
  })
  const [entwurf] = angelegt.mock.calls[0] as unknown as [
    { sitze: { name: string; deckID: string; eingeladen: string }[] },
  ]
  expect(entwurf.sitze[0]?.deckID).toBe('2')
  expect(entwurf.sitze[1]).toMatchObject({
    name: 'Bo',
    deckID: '',
    eingeladen: 'p-bo',
  })
})

it('setzt sich mit dem mitgegebenen Deck an den fremden Tisch, einmal', async () => {
  const beigetreten = vi.fn(() =>
    Promise.resolve({
      laeuft: true,
      pausiert: false,
      gegenueber: 'p-ada',
      partie: '',
      fehler: null,
    }),
  )
  mitDecks({ PartieBeitreten: beigetreten, PartieBeitrittStand: beigetreten })
  render(
    <Buehne
      schicke={{
        art: 'mitspielen',
        name: 'Ada',
        gegenueber: 'p-ada',
        deckID: '2',
      }}
    />,
  )

  await userEvent.click(await screen.findByRole('button', { name: 'los' }))

  await waitFor(() => {
    expect(beigetreten).toHaveBeenCalledWith('p-ada', '', {
      deckID: '2',
      leiheID: '',
    })
  })
  expect(
    beigetreten.mock.calls.filter((ruf) => (ruf as unknown[]).length === 3),
  ).toHaveLength(1)
})
