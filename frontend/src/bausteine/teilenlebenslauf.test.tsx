// **Eine Teilung, die nach ihrem Bauteil ankommt, wird verworfen** —
// Befund Z7 des Zweitblicks über Phase 6.7 (NET-Teil).
//
// Wer die Seite verlässt, während die Kopie noch entsteht, hatte bis dahin
// nichts zum Aufräumen in der Hand: die Kennung kam erst danach — und blieb
// liegen. **Und ein Verwerfen, das scheitert, behält die Kennung** für den
// nächsten Versuch.
import { StrictMode } from 'react'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { Teilen } from './Teilen'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'

import type { TeilenAntwort } from '../apiteilen'

const eigene = kataloge.de.bausteine.teilen
const kennung = 'b'.repeat(32)

const vorbereitet: TeilenAntwort = {
  teilgut: {
    kennung,
    art: 'set',
    titel: 'Set',
    dateiname: 'Set.mse-set',
    groesse: 10,
    text: '',
    mitgehendes: [],
    verluste: [],
    zuGrossFuerMail: false,
  },
  fehler: null,
}

afterEach(() => {
  raeumeApi()
})

function zeige() {
  return render(
    <Sprachraum sprache="de">
      <Teilen art="set" bezug="3" />
    </Sprachraum>,
  )
}

it('verwirft eine Teilung, die nach dem Verlassen ankommt', async () => {
  let fertigstellen: (antwort: TeilenAntwort) => void = () => undefined
  const verwerfen = vi.fn(() =>
    Promise.resolve({ teilgut: null, fehler: null }),
  )
  belegeApi({
    TeilenVorbereiten: () =>
      new Promise<TeilenAntwort>((erfuellen) => {
        fertigstellen = erfuellen
      }),
    TeilenVerwerfen: verwerfen,
  })

  const { unmount } = zeige()
  await userEvent.click(screen.getByRole('button', { name: eigene.knopf }))

  unmount()
  fertigstellen(vorbereitet)

  await waitFor(() => {
    expect(verwerfen).toHaveBeenCalledWith(kennung)
  })
})

it('räumt nach einem gescheiterten Verwerfen noch einmal', async () => {
  const verwerfen = vi
    .fn()
    .mockResolvedValueOnce({
      teilgut: null,
      fehler: { code: 'unbekannt', schluessel: '', meldung: 'klemmt' },
    })
    .mockResolvedValue({ teilgut: null, fehler: null })
  belegeApi({
    TeilenVorbereiten: () => Promise.resolve(vorbereitet),
    TeilenVerwerfen: verwerfen,
  })

  zeige()
  await userEvent.click(screen.getByRole('button', { name: eigene.knopf }))
  await userEvent.click(
    await screen.findByRole('button', { name: eigene.fertig }),
  )
  await screen.findByRole('alert')

  // **Die Teilung steht noch da** — und „Fertig" räumt ein zweites Mal.
  await userEvent.click(screen.getByRole('button', { name: eigene.fertig }))
  expect(
    await screen.findByRole('button', { name: eigene.knopf }),
  ).toBeInTheDocument()
  expect(verwerfen).toHaveBeenNthCalledWith(2, kennung)
})

// **Unter StrictMode bleibt eine Teilung stehen** (Nachlauf über Phase 6.7,
// N5): React baut den Effekt auf, ab und wieder auf.
it('zeigt eine Teilung auch unter StrictMode', async () => {
  const verwerfen = vi.fn()
  belegeApi({
    TeilenVorbereiten: () => Promise.resolve(vorbereitet),
    TeilenVerwerfen: verwerfen,
  })

  const { unmount } = render(
    <StrictMode>
      <Sprachraum sprache="de">
        <Teilen art="set" bezug="3" />
      </Sprachraum>
    </StrictMode>,
  )
  await userEvent.click(screen.getByRole('button', { name: eigene.knopf }))

  expect(await screen.findByText(/Set\.mse-set/)).toBeInTheDocument()
  expect(verwerfen).not.toHaveBeenCalled()

  // **Erst das Verlassen räumt** — vor dem Aufräumen der Fassade.
  unmount()
  expect(verwerfen).toHaveBeenCalledWith(kennung)
})

// **Eine Antwort ohne Teilung zeigt nichts** — und räumt nichts.
it('bleibt beim Knopf, wenn nichts vorbereitet wurde', async () => {
  const verwerfen = vi.fn()
  belegeApi({
    TeilenVorbereiten: () => Promise.resolve({ teilgut: null, fehler: null }),
    TeilenVerwerfen: verwerfen,
  })

  const { unmount } = zeige()
  await userEvent.click(screen.getByRole('button', { name: eigene.knopf }))
  expect(
    await screen.findByRole('button', { name: eigene.knopf }),
  ).toBeEnabled()

  unmount()

  expect(verwerfen).not.toHaveBeenCalled()
})
