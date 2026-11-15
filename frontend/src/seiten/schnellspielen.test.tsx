// **Schnell spielen** — P6
// (#2374),
// Artboard `PartieStart09`: aus dem Menü eines Decks oder der Palette mit
// der Vorbelegung gleich an den Tisch — S1 aus dem Deck in drei Handlungen.
// **Einmal je Wunsch, und nur, was vollständig ist**: fehlt etwas, steht das
// Formular da und sagt es.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { afterEach, expect, it, vi } from 'vitest'

import { Anschlussraum, useAnschluss } from '../anschlussraum'
import { Spielen } from './Spielen'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'
import { eineAntwort } from './partieprobe'
import { mitDecks } from './spielenprobe'
import { raeumeApi } from '../test/api'

import type { Seitenschluessel } from '../navigation'

afterEach(raeumeApi)

const eigene = kataloge.de.seiten.spielen

function Buehne() {
  const [aktuell, setzeAktuell] = useState<Seitenschluessel>('decks')

  return (
    <Sprachraum sprache="de">
      <Anschlussraum aktuell={aktuell} wechseln={setzeAktuell}>
        <Abgeber />
        <Spielen wechseln={() => undefined} />
      </Anschlussraum>
    </Sprachraum>
  )
}

function Abgeber() {
  const anschluss = useAnschluss()

  return (
    <button
      type="button"
      onClick={() => {
        anschluss.uebergib({
          art: 'partie',
          name: 'Kontrolle',
          deckID: '2',
          sofort: true,
        })
      }}
    >
      schnell
    </button>
  )
}

function mitProfil(name: string) {
  const angelegt = vi.fn(() => Promise.resolve(eineAntwort()))
  mitDecks({
    GezeigtesProfil: () =>
      Promise.resolve({
        profil: { id: '7', nutzerID: '1', name },
        fehler: null,
      }),
    PartieAnlegen: angelegt,
  })

  return angelegt
}

it('fängt mit der Vorbelegung gleich an, einmal', async () => {
  const angelegt = mitProfil('Ada')
  render(<Buehne />)

  await userEvent.click(await screen.findByRole('button', { name: 'schnell' }))

  await waitFor(() => {
    expect(angelegt).toHaveBeenCalledTimes(1)
  })
  expect(angelegt).toHaveBeenCalledWith(
    expect.objectContaining({
      sitze: [
        expect.objectContaining({ name: 'Ada', deckID: '2' }),
        // **Ein eigenes Deck** — welches, hängt daran, ob die Vorbelegung
        // vor oder nach der Übergabe stand.
        expect.objectContaining({
          name: eigene.sitzZweiVorgabe,
          deckID: expect.stringMatching(/^[12]$/) as string,
        }),
      ],
    }),
  )
})

it('fängt nichts Unvollständiges an und sagt, was fehlt', async () => {
  const angelegt = mitProfil('')
  render(<Buehne />)

  await userEvent.click(await screen.findByRole('button', { name: 'schnell' }))

  expect(await screen.findByText(eigene.fehltName)).toBeInTheDocument()
  expect(angelegt).not.toHaveBeenCalled()

  // **Und fängt auch nicht nachträglich an**, wenn der Name kommt — der
  // Wunsch war einer, und er ist vorbei.
  await userEvent.type(screen.getByLabelText(`${eigene.sitzName} 1`), 'Ada')
  expect(screen.getByRole('button', { name: eigene.anfangen })).toBeEnabled()
  expect(angelegt).not.toHaveBeenCalled()
})

// **Ein Anfang verbraucht, was mitgereist ist** — gefunden am nativen
// Fenster (`partiestart09`): nach dem Schließen des Tisches baute sich das
// Formular neu auf, nahm „Schnell spielen“ ein zweites Mal auf und legte
// gleich die nächste Partie an.
it('fängt nach dem Schließen des Tisches nicht noch einmal an', async () => {
  const angelegt = mitProfil('Ada')
  render(<Buehne />)

  await userEvent.click(await screen.findByRole('button', { name: 'schnell' }))
  await waitFor(() => {
    expect(angelegt).toHaveBeenCalledTimes(1)
  })

  await userEvent.click(
    await screen.findByRole('button', { name: eigene.verlassen }),
  )
  await screen.findByRole('heading', { level: 1, name: eigene.anfangTitel })
  await screen.findByDisplayValue('Ada')

  expect(angelegt).toHaveBeenCalledTimes(1)
})
