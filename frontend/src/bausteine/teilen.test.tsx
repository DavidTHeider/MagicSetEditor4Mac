// **Teilen, wie ein Mensch es sieht** — NET-04
// (#1708).
//
// AK-02: vor dem Teilen stehen Datei, Größe, was mitgeht und was fehlt.
// AK-03: zu groß für Mail hat einen Hinweis, eine Einladung ihren Code zum
// Kopieren. AK-04: Teilen ist eine Kopie, und „Fertig" räumt sie weg —
// ein Schließen der Seite auch.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { Teilen } from './Teilen'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'

import type { TeilenAntwort, TeilgutDTO } from '../apiteilen'

const eigene = kataloge.de.bausteine.teilen

function gut(teile: Partial<TeilgutDTO>): TeilenAntwort {
  return {
    teilgut: {
      kennung: 'a'.repeat(32),
      art: 'sammlung',
      titel: 'Sammlung',
      dateiname: 'Sammlung.csv',
      groesse: 2048,
      text: '',
      mitgehendes: ['exemplare', 'erwerb'],
      verluste: [],
      zuGrossFuerMail: false,
      ...teile,
    },
    fehler: null,
  }
}

const leer: TeilenAntwort = { teilgut: null, fehler: null }

afterEach(() => {
  raeumeApi()
})

function zeige(art: TeilgutDTO['art'], bezug = '') {
  return render(
    <Sprachraum sprache="de">
      <Teilen art={art} bezug={bezug} />
    </Sprachraum>,
  )
}

it('zeigt vorher, was mitgeht, und bietet dann an', async () => {
  const anbieten = vi.fn(() => Promise.resolve(leer))
  const verwerfen = vi.fn(() => Promise.resolve(leer))
  belegeApi({
    TeilenVorbereiten: () => Promise.resolve(gut({})),
    TeilenAnbieten: anbieten,
    TeilenVerwerfen: verwerfen,
  })

  zeige('sammlung')
  const hand = userEvent.setup()

  await hand.click(screen.getByRole('button', { name: eigene.knopf }))

  expect(await screen.findByText(/Sammlung\.csv/)).toBeInTheDocument()
  expect(screen.getByText(eigene.teile.erwerb)).toBeInTheDocument()
  expect(screen.getByText(eigene.kopie)).toBeInTheDocument()
  expect(screen.queryByText(eigene.zuGross)).toBeNull()

  await hand.click(screen.getByRole('button', { name: eigene.anbieten }))
  expect(anbieten).toHaveBeenCalledWith('a'.repeat(32))

  await hand.click(screen.getByRole('button', { name: eigene.fertig }))
  expect(verwerfen).toHaveBeenCalledWith('a'.repeat(32))
  expect(
    await screen.findByRole('button', { name: eigene.knopf }),
  ).toBeInTheDocument()
})

it('nennt Verluste, zu Großes und bietet bei einer Einladung den Code', async () => {
  belegeApi({
    TeilenVorbereiten: () =>
      Promise.resolve(
        gut({
          art: 'einladung',
          text: 'mse4mac-kontakt:1:abc',
          verluste: ['zuordnungen'],
          zuGrossFuerMail: true,
        }),
      ),
    TeilenVerwerfen: () => Promise.resolve(leer),
  })

  const { unmount } = zeige('einladung', 'mse4mac-kontakt:1:abc')
  const hand = userEvent.setup()

  await hand.click(screen.getByRole('button', { name: eigene.knopf }))

  expect(await screen.findByText(eigene.teile.zuordnungen)).toBeInTheDocument()
  expect(screen.getByText(eigene.zuGross)).toBeInTheDocument()
  await hand.click(screen.getByRole('button', { name: eigene.kopieren }))

  // **Das Aushängen räumt die Kopie weg**, bevor die Fassade geleert wird.
  unmount()
})

it('meldet, was beim Vorbereiten scheitert', async () => {
  belegeApi({
    TeilenVorbereiten: () =>
      Promise.resolve({
        teilgut: null,
        fehler: {
          code: 'validierung',
          schluessel: 'validierungsgrund',
          meldung: 'nichts da',
          feld: '',
          werte: { grund: 'nichts da' },
          vorgang: '',
        },
      }),
  })

  zeige('deck', '5')
  await userEvent.click(screen.getByRole('button', { name: eigene.knopf }))

  expect(await screen.findByRole('alert')).toHaveTextContent('nichts da')
})

// **Wer die Seite verlässt, lässt keine Kopie liegen** (AK-04).
it('räumt beim Verlassen weg, was vorbereitet war', async () => {
  const verwerfen = vi.fn(() => Promise.resolve(leer))
  belegeApi({
    TeilenVorbereiten: () => Promise.resolve(gut({})),
    TeilenVerwerfen: verwerfen,
  })

  const { unmount } = zeige('sammlung')
  await userEvent.click(screen.getByRole('button', { name: eigene.knopf }))
  await screen.findByText(/Sammlung\.csv/)

  unmount()

  expect(verwerfen).toHaveBeenCalledWith('a'.repeat(32))
})

// **Ein Teil, den der Katalog nicht kennt, steht roh da** — er kann nur aus
// einer neueren Go-Seite kommen, und verschwiegen wäre er schlimmer.
it('nennt einen unbekannten Teil beim Schlüssel', async () => {
  belegeApi({
    TeilenVorbereiten: () =>
      Promise.resolve(gut({ mitgehendes: ['neuerteil'] })),
    TeilenVerwerfen: () => Promise.resolve(leer),
  })

  const { unmount } = zeige('sammlung')
  await userEvent.click(screen.getByRole('button', { name: eigene.knopf }))

  expect(await screen.findByText('neuerteil')).toBeInTheDocument()
  unmount()
})

// **Ein Deck lässt sich auch als Liste kopieren** — für einen Deckbaukasten,
// der sie aus der Zwischenablage liest.
it('kopiert bei einem Deck die Liste', async () => {
  const liste = vi.fn(() =>
    Promise.resolve({ liste: 'Deck\n4 Lightning Bolt\n', fehler: null }),
  )
  belegeApi({
    TeilenVorbereiten: () => Promise.resolve(gut({ art: 'deck' })),
    TeilenVerwerfen: () => Promise.resolve(leer),
    DecklisteAusgeben: liste,
  })

  const { unmount } = zeige('deck', '5')
  const hand = userEvent.setup()

  await hand.click(screen.getByRole('button', { name: eigene.knopf }))
  await hand.click(
    await screen.findByRole('button', { name: eigene.listeKopieren }),
  )

  expect(liste).toHaveBeenCalledWith('5')
  unmount()
})
