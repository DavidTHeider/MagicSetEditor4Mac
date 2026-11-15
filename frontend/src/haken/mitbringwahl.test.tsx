// **Die Deckwahl beim Beitritt** — P9c
// (#2377):
// ein Deck des Wirts, die eigenen Decks und die angenommenen Leihen, die sich
// spielen lassen; vorgewählt das zuletzt geöffnete eigene Deck.
import { renderHook, waitFor } from '@testing-library/react'
import type { ReactNode } from 'react'
import { afterEach, expect, it } from 'vitest'

import { einDeck } from '../seiten/partieprobe'
import { eineLeihe } from '../test/leiheprobe'
import { belegeApi, raeumeApi } from '../test/api'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'
import { leihwert } from '../spiel/sitzwahl'
import { useMitbringwahl } from './useMitbringwahl'

afterEach(raeumeApi)

const eigene = kataloge.de.seiten.spielen

function raum({ children }: { children: ReactNode }) {
  return <Sprachraum sprache="de">{children}</Sprachraum>
}

it('bietet Wirt, eigene Decks und spielbare Leihen und wählt vor', async () => {
  belegeApi({
    Decks: () =>
      Promise.resolve({
        decks: [einDeck('1', 'Waldschritt'), einDeck('2', 'Glutpfad')],
        fehler: null,
      }),
    Deckleihen: () =>
      Promise.resolve({
        leihen: [
          eineLeihe({ kennung: 'l1', verliehen: false }),
          eineLeihe({ kennung: 'l2', verliehen: true }),
        ],
        pfad: '',
        fehler: null,
      }),
    GezeigtesProfil: () => Promise.resolve({ profil: null, fehler: null }),
    Weiterarbeiten: () =>
      Promise.resolve({
        geoeffnet: [{ objekt: { art: 'deck', kennung: '2' } }],
        fehler: null,
      }),
  } as never)

  const { result } = renderHook(() => useMitbringwahl(), { wrapper: raum })

  expect(result.current.vorgabe).toBeUndefined()
  await waitFor(() => {
    expect(result.current.vorgabe).toBe('2')
  })
  await waitFor(() => {
    expect(result.current.eintraege.map((eintrag) => eintrag.wert)).toEqual([
      '',
      '1',
      '2',
      leihwert('l1'),
    ])
  })
  expect(result.current.eintraege[0]?.name).toBe(eigene.gastMitbringenWirt)
})
