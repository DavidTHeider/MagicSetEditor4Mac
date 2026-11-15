// **Eine Partie in drei Schritten** — P6
// (#2374),
// Befund C-04, Artboard `PartieStart09`: das Formular geht mit Namen und
// Decks auf, ⌘↩ fängt an, und während der Anfang läuft, sagt der Knopf es —
// ein zweiter legt keine zweite Partie an.
import { act, fireEvent, screen, waitFor } from '@testing-library/react'
import { afterEach, expect, it, vi } from 'vitest'

import { raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import { eineAntwort } from './partieprobe'
import { mitDecks, zeige } from './spielenprobe'

import type { PartieAntwort } from '../apipartie'

afterEach(raeumeApi)

const eigene = kataloge.de.seiten.spielen

const quellen = {
  GezeigtesProfil: () =>
    Promise.resolve({
      profil: { id: '7', nutzerID: '1', name: 'Ada' },
      fehler: null,
    }),
  Weiterarbeiten: () =>
    Promise.resolve({
      zuletzt: null,
      geoeffnet: [
        {
          objekt: {
            art: 'deck',
            kennung: '2',
            name: 'Kontrolle',
            angeheftet: false,
          },
          vorgang: 'geoeffnet',
          zeitpunkt: '',
        },
      ],
      eingefuehrt: [],
      ausgegeben: [],
      angeheftet: [],
      fehler: null,
    }),
  Partien: () =>
    Promise.resolve({
      partien: [
        {
          id: 'p0',
          fassung: '9',
          spiel: 'magic',
          stand: 'beendet',
          sitze: [
            { kennung: 'sitz1', name: 'Ada' },
            { kennung: 'sitz2', name: 'Bo' },
          ],
          objekte: 0,
          zug: 3,
          eigenerSitz: '',
          dran: false,
        },
      ],
      fehler: null,
    }),
}

it('geht mit Namen und Decks auf', async () => {
  mitDecks(quellen)
  zeige()

  expect(await screen.findByDisplayValue('Ada')).toBeInTheDocument()
  await waitFor(() => {
    expect(screen.getByLabelText(`${eigene.sitzName} 2`)).toHaveValue('Bo')
  })

  // **Das zuletzt geöffnete Deck am ersten Sitz** — und ohne ein zweites
  // zuletzt geöffnetes spiegelt der zweite Sitz den ersten.
  const decks = screen.getAllByLabelText(eigene.sitzDeck)
  expect(decks[0]).toHaveValue('2')
  expect(decks[1]).toHaveValue('2')
})

it('fängt mit ⌘↩ an, und nur einmal, solange es läuft', async () => {
  let fertig: (antwort: PartieAntwort) => void = () => undefined
  const angelegt = vi.fn(
    () =>
      new Promise<PartieAntwort>((los) => {
        fertig = los
      }),
  )
  mitDecks({ ...quellen, PartieAnlegen: angelegt })
  zeige()

  const name = await screen.findByDisplayValue('Ada')
  await waitFor(() => {
    expect(screen.getByRole('button', { name: eigene.anfangen })).toBeEnabled()
  })

  fireEvent.keyDown(name, { key: 'Enter', metaKey: true })
  fireEvent.keyDown(name, { key: 'Enter', metaKey: true })

  expect(
    await screen.findByRole('button', { name: eigene.anfangenLaeuft }),
  ).toBeDisabled()
  expect(angelegt).toHaveBeenCalledTimes(1)

  // **Ohne ⌘ ist ↩ nur ↩.**
  await act(async () => {
    fertig(eineAntwort())
    await Promise.resolve()
  })
  fireEvent.keyDown(name, { key: 'Enter' })
  expect(angelegt).toHaveBeenCalledTimes(1)
})

// **⌘↩ fängt nur an, was vollständig ist** — ohne Namen im Profil fehlt der
// am ersten Sitz, und das Kürzel tut dann so wenig wie der gesperrte Knopf.
it('fängt mit ⌘↩ nichts Unvollständiges an', async () => {
  const angelegt = vi.fn(() => Promise.resolve(eineAntwort()))
  mitDecks({
    ...quellen,
    GezeigtesProfil: () =>
      Promise.resolve({
        profil: { id: '7', nutzerID: '1', name: '' },
        fehler: null,
      }),
    PartieAnlegen: angelegt,
  })
  zeige()

  const zweiter = await screen.findByDisplayValue('Bo')
  fireEvent.keyDown(zweiter, { key: 'Enter', metaKey: true })

  expect(await screen.findByText(eigene.fehltName)).toBeInTheDocument()
  expect(angelegt).not.toHaveBeenCalled()
})

// **Wer die Decks wieder abwählt, erfährt, dass eines fehlt** — die
// Vorbelegung nimmt die Wahl nicht weg.
it('sagt, dass ein Deck fehlt, wenn keines mehr gewählt ist', async () => {
  mitDecks(quellen)
  zeige()

  await screen.findByDisplayValue('Ada')
  for (const deck of screen.getAllByLabelText(eigene.sitzDeck)) {
    fireEvent.change(deck, { target: { value: '' } })
  }

  expect(await screen.findByText(eigene.fehltDeck)).toBeInTheDocument()
  expect(screen.getByRole('button', { name: eigene.anfangen })).toBeDisabled()
})
