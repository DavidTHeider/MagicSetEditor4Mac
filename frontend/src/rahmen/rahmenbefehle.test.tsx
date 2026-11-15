// **Der Weg von der Palette in die Kommandozentrale** — Werkstück U12 von
// Phase 5.1 (#1423). Er steht in einer eigenen Datei, weil `Rahmen.test.tsx`
// an der Zeilengrenze steht (§2).
import { fireEvent, render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { Rahmen } from './Rahmen'
import { belegeApi, raeumeApi } from '../test/api'
import type { BefehlDTO } from '../apibefehle'

const sets: BefehlDTO = {
  name: 'lesen.sets',
  faehigkeit: 'Sets auflisten',
  cli: 'mse sets',
  rest: 'GET /sets',
  erklaerung: 'Zählt die geöffneten Sets auf.',
  schreibend: false,
  idempotent: true,
  gefaehrlich: false,
  schema: { type: 'object' },
}

afterEach(() => {
  raeumeApi()
})

describe('Palette und Kommandozentrale', () => {
  it('öffnet die Zentrale über einem Befehl, die Palette geht vorher zu', async () => {
    belegeApi({
      Befehlskatalog: () => Promise.resolve({ befehle: [sets], fehler: null }),
    })

    render(
      <Sprachraum sprache="de">
        <Rahmen aktuell="start" wechseln={vi.fn()} suchen={vi.fn()}>
          <p>{kataloge.de.seiten.start.was}</p>
        </Rahmen>
      </Sprachraum>,
    )

    fireEvent.keyDown(window, { key: 'k', metaKey: true })

    await userEvent.type(
      screen.getByRole('textbox', {
        name: kataloge.de.allgemein.palette.eingabe,
      }),
      'lesen.sets',
    )

    await userEvent.click(
      await within(screen.getByRole('dialog')).findByRole('button', {
        name: /lesen\.sets/,
      }),
    )

    // **Kein Dialog über einem Dialog** (§6): genau einer steht da, und es
    // ist die Zentrale.
    expect(screen.getAllByRole('dialog')).toHaveLength(1)
    const zentrale = screen.getByRole('dialog')
    expect(zentrale).toHaveAttribute(
      'aria-label',
      kataloge.de.bausteine.kommandozentrale.titel,
    )

    await userEvent.click(
      within(zentrale).getByRole('button', {
        name: kataloge.de.bausteine.kommandozentrale.schliessen,
      }),
    )

    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })
})
