// **Die Sitze stehen schon da** — P6
// (#2374),
// Befund C-04, Artboard `PartieStart09`: bis zum Anfangen tippte man zwei
// Namen und wählte zwei Decks (S1: neun Handlungen). Jetzt kommt der erste
// Sitz aus dem Profil und dem Deck, aus dem man kommt, der zweite aus dem
// letzten Gegenüber — **und ein fremdes Deck wird nie von selbst gewählt**.
import { describe, expect, it } from 'vitest'

import { vorbelegung } from './vorbelegung'

const deck = (id: string, name: string) => ({ id, name })
const decks = [
  deck('d1', 'Waldschritt'),
  deck('d2', 'Glutpfad'),
  deck('d3', 'Aschenrat'),
]

describe('die Vorbelegung der Sitze', () => {
  it('setzt das mitgebrachte Deck und den Profilnamen an den ersten Sitz', () => {
    expect(
      vorbelegung({
        decks,
        mitgebracht: 'd3',
        zuletzt: ['d2', 'd1'],
        profilname: 'Ada',
        gegenueber: 'Bo',
        ersatzname: 'Gegenüber',
      }),
    ).toEqual([
      { name: 'Ada', deckID: 'd3', profilID: '', leiheID: '' },
      { name: 'Bo', deckID: 'd2', profilID: '', leiheID: '' },
    ])
  })

  it('nimmt ohne mitgebrachtes Deck das zuletzt geöffnete', () => {
    const [erster, zweiter] = vorbelegung({
      decks,
      mitgebracht: undefined,
      zuletzt: ['d2', 'd1'],
      profilname: 'Ada',
      gegenueber: undefined,
      ersatzname: 'Gegenüber',
    })

    expect(erster?.deckID).toBe('d2')
    expect(zweiter).toEqual({
      name: 'Gegenüber',
      deckID: 'd1',
      profilID: '',
      leiheID: '',
    })
  })

  // **Ein fremdes Deck nie**: was nicht in der eigenen Liste steht — auch
  // ein zuletzt geöffnetes, das es nicht mehr gibt —, wird übergangen.
  it('wählt nur eigene Decks und spiegelt sonst das erste', () => {
    const [erster, zweiter] = vorbelegung({
      decks: [deck('d1', 'Waldschritt')],
      mitgebracht: 'fremd',
      zuletzt: ['weg', 'd1'],
      profilname: '',
      gegenueber: undefined,
      ersatzname: 'Gegenüber',
    })

    expect(erster).toEqual({
      name: '',
      deckID: 'd1',
      profilID: '',
      leiheID: '',
    })
    expect(zweiter?.deckID).toBe('d1')
  })

  it('setzt ohne Deck keines', () => {
    expect(
      vorbelegung({
        decks: [],
        mitgebracht: undefined,
        zuletzt: [],
        profilname: 'Ada',
        gegenueber: undefined,
        ersatzname: 'Gegenüber',
      }).map((sitz) => sitz.deckID),
    ).toEqual(['', ''])
  })

  // **Mit einer Einladung ist der zweite Sitz der Gast** — P9c (#2377): er
  // bringt sein Deck mit, und an seinem Sitz liegt keines.
  it('setzt einen Eingeladenen an den zweiten Sitz', () => {
    expect(
      vorbelegung({
        decks,
        mitgebracht: 'd3',
        zuletzt: ['d2'],
        profilname: 'Ada',
        gegenueber: 'Cy',
        ersatzname: 'Gegenüber',
        eingeladen: { kennung: 'p-bo', name: 'Bo' },
      }),
    ).toEqual([
      { name: 'Ada', deckID: 'd3', profilID: '', leiheID: '' },
      { name: 'Bo', deckID: '', profilID: '', leiheID: '', eingeladen: 'p-bo' },
    ])
  })
})
