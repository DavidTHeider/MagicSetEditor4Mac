// **Eine Einladung genügt** — NET-01b
// (#1954).
//
// # Wogegen hier geprüft wird
//
// **Der Name, der hinausgeht, wird gefragt und nicht vorausgefüllt** — und
// was die Meldung beim Einladenden ergab, steht nach der Annahme da: in
// jedem der drei Stände, und beim Scheitern mit seinem Grund.
import { screen, waitFor } from '@testing-library/react'
import { afterEach, expect, it, vi } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'
import { einsetzen } from '../sprache/einsetzen'

import {
  alsAntwort,
  derCode,
  eigene,
  eineVorschau,
  pruefe,
  stelleAuf,
} from '../test/einladungsprobe'

import type { Einloesemeldung, Peerkontakteantwort } from '../apipeerkontakt'

afterEach(() => {
  raeumeApi()
})

function gemeldetMit(
  gemeldet: Einloesemeldung,
): () => Promise<Peerkontakteantwort> {
  return () =>
    Promise.resolve({ kontakte: [], zwecke: [], gemeldet, fehler: null })
}

it('fragt nach dem Namen, der hinausgeht, und meldet sich damit', async () => {
  const annehmen = vi.fn(gemeldetMit({ stand: 'gemeldet' }))
  belegeApi({
    KontakteinladungPruefen: () =>
      alsAntwort(eineVorschau({ einloesbar: true })),
    KontakteinladungAnnehmen: annehmen,
  })

  const { hand } = stelleAuf()
  await pruefe(hand)

  // **Leer und nicht vorausgefüllt** — ein Name, der hinausgeht, ist eine
  // Entscheidung.
  const feld = await screen.findByLabelText(eigene.meinName)
  expect(feld).toHaveValue('')

  await hand.type(feld, 'Bert')
  await hand.click(screen.getByRole('button', { name: eigene.spaeter }))

  await waitFor(() => {
    expect(annehmen).toHaveBeenCalledWith(derCode, 'Anna', '', false, 'Bert')
  })

  expect(
    await screen.findByText(
      einsetzen(eigene.aufgenommen, { name: 'Anna' }) +
        ' ' +
        einsetzen(eigene.gemeldet, { name: 'Anna' }),
    ),
  ).toBeInTheDocument()
})

it('fragt nicht, wenn die Einladung nicht einlösbar ist', async () => {
  belegeApi({
    KontakteinladungPruefen: () => alsAntwort(eineVorschau()),
    KontakteinladungAnnehmen: gemeldetMit({ stand: 'ohne' }),
  })

  const { hand } = stelleAuf()
  await pruefe(hand)

  await hand.click(await screen.findByRole('button', { name: eigene.spaeter }))

  expect(screen.queryByLabelText(eigene.meinName)).toBeNull()
  expect(
    await screen.findByText(
      einsetzen(eigene.aufgenommen, { name: 'Anna' }) +
        ' ' +
        einsetzen(eigene.nichtGemeldet, { name: 'Anna' }),
    ),
  ).toBeInTheDocument()
})

it('nennt beim Scheitern den Grund und den Rückfall', async () => {
  belegeApi({
    KontakteinladungPruefen: () =>
      alsAntwort(eineVorschau({ einloesbar: true })),
    KontakteinladungAnnehmen: gemeldetMit({
      stand: 'gescheitert',
      grund: 'unbeantwortet',
      meldung: 'dein Gegenüber hat die Einlösung nicht beantwortet',
    }),
  })

  const { hand } = stelleAuf()
  await pruefe(hand)

  await hand.click(await screen.findByRole('button', { name: eigene.spaeter }))

  expect(
    await screen.findByText(
      einsetzen(eigene.aufgenommen, { name: 'Anna' }) +
        ' ' +
        einsetzen(eigene.meldungGescheitert, {
          name: 'Anna',
          grund: eigene.meldegruende.unbeantwortet,
        }),
    ),
  ).toBeInTheDocument()
})

it('nimmt für einen Grund, den es nicht kennt, den allgemeinen', async () => {
  belegeApi({
    KontakteinladungPruefen: () =>
      alsAntwort(eineVorschau({ einloesbar: true })),
    KontakteinladungAnnehmen: gemeldetMit({
      stand: 'gescheitert',
      grund: 'neu',
    }),
  })

  const { hand } = stelleAuf()
  await pruefe(hand)

  await hand.click(await screen.findByRole('button', { name: eigene.spaeter }))

  expect(
    await screen.findByText(
      einsetzen(eigene.aufgenommen, { name: 'Anna' }) +
        ' ' +
        einsetzen(eigene.meldungGescheitert, {
          name: 'Anna',
          grund: eigene.meldegruende.gescheitert,
        }),
    ),
  ).toBeInTheDocument()
})

// **Ein Nachladefehler nimmt die Annahme nicht zurück** (§7, Regel 3;
// #1960):
// die Annahme ist geschrieben, die Meldung steht da, und die Liste nebenan
// lädt selbst nach — kein Fehler am Schritt, der zum Wiederholen einlüde.
it('nimmt eine Annahme mit Nachladefehler als geschehen', async () => {
  belegeApi({
    KontakteinladungPruefen: () => alsAntwort(eineVorschau()),
    KontakteinladungAnnehmen: () =>
      Promise.resolve({
        kontakte: [],
        zwecke: [],
        gemeldet: { stand: 'ohne' },
        nachladefehler: {
          code: 'unbekannt',
          schluessel: '',
          meldung: 'Die Liste klemmt',
        },
        fehler: null,
      }),
  })

  const { hand, nachAufnahme } = stelleAuf()
  await pruefe(hand)

  await hand.click(await screen.findByRole('button', { name: eigene.spaeter }))

  expect(
    await screen.findByText(
      new RegExp(einsetzen(eigene.aufgenommen, { name: 'Anna' })),
    ),
  ).toBeInTheDocument()
  expect(nachAufnahme).toHaveBeenCalledTimes(1)
  expect(screen.queryByRole('alert')).toBeNull()
})
