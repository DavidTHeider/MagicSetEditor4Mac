// **⌘↩ im Quelltextfeld** — Werkstück U13 von Phase 5.1 (#1424): wer einen
// Ausdruck tippt, soll ihn auswerten können, ohne die Hand vom Feld zu nehmen.
import { fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Quelltextfeld } from './Quelltextfeld'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'

function zeige(aufAusfuehren?: () => void) {
  belegeApi({ Skriptfarben: () => Promise.resolve({ merkmale: [] }) })

  render(
    <Sprachraum sprache="de">
      <Quelltextfeld
        bezeichnung="Ausdruck"
        wert="1 + 1"
        setzeWert={vi.fn()}
        aufAusfuehren={aufAusfuehren}
      />
    </Sprachraum>,
  )

  return screen.getByLabelText('Ausdruck')
}

afterEach(raeumeApi)

describe('⌘↩ im Quelltextfeld', () => {
  it('führt aus — mit ⌘ oder Strg, nicht mit Enter allein', () => {
    const aufAusfuehren = vi.fn()
    const feld = zeige(aufAusfuehren)

    fireEvent.keyDown(feld, { key: 'Enter', metaKey: true })
    expect(aufAusfuehren).toHaveBeenCalledTimes(1)

    fireEvent.keyDown(feld, { key: 'Enter', ctrlKey: true })
    expect(aufAusfuehren).toHaveBeenCalledTimes(2)

    // **Der Beleg, dass jemand zuhört, steht darüber** (§3): ein Enter ohne
    // Taste dazu ist ein Zeilenumbruch und kein Lauf.
    fireEvent.keyDown(feld, { key: 'Enter' })
    expect(aufAusfuehren).toHaveBeenCalledTimes(2)
  })

  it('tut ohne Rückruf nichts — und stürzt nicht', () => {
    const feld = zeige()

    fireEvent.keyDown(feld, { key: 'Enter', metaKey: true })

    expect(feld).toBeInTheDocument()
  })
})
