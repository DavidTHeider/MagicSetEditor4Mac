// **Deck und Turnier in der Menüleiste** — Zweitblick über Phase 6 von
// 0.9.0, zweiter Teil, Befund 6: Überlauf, Kontextmenü und ⇧F10 teilten die
// Liste, die Menüleiste fehlte — anders als beim Exemplar. Jetzt melden
// beide ihre Liste über `useHandlungsliste` an.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import type { Objekthandlungen } from '../objekthandlungen'
import { Objekthandlungsraum } from '../objekthandlungsraum'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { bauartprobe, deckprobe, formatprobe } from './deckprobe'
import { Decks } from './Decks'

const decktexte = kataloge.de.seiten.decks

afterEach(() => {
  raeumeApi()
})

it('meldet das offene Deck mit seiner Liste an der Menüleiste an', async () => {
  const melden = vi.fn<
    (meldung: Objekthandlungen) => Promise<{ fehler: null }>
  >(() => Promise.resolve({ fehler: null }))
  belegeApi({
    ObjekthandlungenMelden: melden,
    Decks: () => Promise.resolve({ decks: [deckprobe()], fehler: null }),
    Deck: () => Promise.resolve({ deck: deckprobe(), fehler: null }),
    Formate: () => Promise.resolve({ formate: [formatprobe()], fehler: null }),
    Bauarten: () =>
      Promise.resolve({ bauarten: [bauartprobe()], fehler: null }),
  })
  render(
    <Sprachraum sprache="de">
      <Objekthandlungsraum>
        <Decks wechseln={vi.fn()} />
      </Objekthandlungsraum>
    </Sprachraum>,
  )

  await userEvent.click(
    await screen.findByRole('button', { name: `${decktexte.waehlen}: Aggro` }),
  )

  await waitFor(() => {
    const deck = melden.mock.calls
      .map(([meldung]) => meldung)
      .filter((meldung) => meldung.art === 'deck')
      .at(-1)
    expect(deck?.titel).toBe(decktexte.deckmenue)
    expect(deck?.eintraege.map((eintrag) => eintrag.name)).toContain(
      decktexte.umbenennen,
    )
  })
})
