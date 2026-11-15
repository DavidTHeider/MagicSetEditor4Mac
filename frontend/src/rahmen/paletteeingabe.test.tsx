// **⏎ im Feld führt die oberste Zeile aus** — G07
// (#2254).
//
// Der Fuß der Palette sagt „↵ ausführen", und wer „Sets" tippt, steht im
// Feld. Dort tat ⏎ nichts: ausgeführt wurde erst, wenn ↓ eine Zeile gewählt
// hatte. **Die oberste ist sichtbar vorgewählt** (`data-vorwahl`) — ein
// unsichtbarer Standard wäre eine Überraschung. Ohne Eingabe gibt es keine
// Absicht, und ⏎ tut nichts.
import { fireEvent, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Palette } from './Palette'

afterEach(raeumeApi)

function zeige() {
  belegeApi({
    Befehlskatalog: () => Promise.resolve({ befehle: [], fehler: null }),
  })
  const waehlen = vi.fn()
  const schliessen = vi.fn()

  render(
    <Sprachraum sprache="de">
      <Palette
        schliessen={schliessen}
        waehlen={waehlen}
        waehleBefehl={vi.fn()}
      />
    </Sprachraum>,
  )

  return { waehlen, feld: screen.getByRole('textbox') }
}

describe('⏎ im Feld der Palette', () => {
  it('führt die oberste Zeile aus und zeigt sie vorher als gewählt', async () => {
    const { waehlen, feld } = zeige()

    await userEvent.type(feld, 'Sets')

    const vorgewaehlt = document.querySelectorAll('[data-vorwahl]')
    expect(vorgewaehlt).toHaveLength(1)
    expect(vorgewaehlt[0]).toHaveTextContent(kataloge.de.seiten.sets.titel)

    await userEvent.keyboard('{Enter}')

    expect(waehlen).toHaveBeenCalledWith('sets')
  })

  it('tut ohne Eingabe nichts', async () => {
    const { waehlen, feld } = zeige()

    feld.focus()
    await userEvent.keyboard('{Enter}')

    expect(document.querySelector('[data-vorwahl]')).toBeNull()
    expect(waehlen).not.toHaveBeenCalled()

    // **Der Zuhörer** (§3): dieselbe Taste nach einer Eingabe wirkt.
    await userEvent.type(feld, 'Sets{Enter}')
    expect(waehlen).toHaveBeenCalledWith('sets')
  })

  // **Eine Eingabemethode hat ⏎ für sich** — bei Japanisch oder Chinesisch
  // bestätigt die Taste die Zeichenwahl, und ausgeführt wird dann nichts.
  it('führt während einer Eingabemethode nichts aus', async () => {
    const { waehlen, feld } = zeige()

    await userEvent.type(feld, 'Sets')
    fireEvent.keyDown(feld, { key: 'Enter', isComposing: true })
    expect(waehlen).not.toHaveBeenCalled()

    fireEvent.keyDown(feld, { key: 'Enter' })
    expect(waehlen).toHaveBeenCalledWith('sets')
  })
})
