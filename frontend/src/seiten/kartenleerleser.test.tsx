// **Ein Leser fügt nichts in ein leeres Set ein** — neuer Befund des
// Nachlaufs über Phase G
// (#2276).
//
// Der Knopf im Leerzustand der Kartenlasche rief `AblageEinfuegen` ohne
// Schreibsperre; für einen Leser stand er bedienbar da und scheiterte erst am
// Tor der Bindung.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import { Schreibrechtraum } from '../haken/schreibrecht'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { Kartenleer } from './Kartenleer'

describe('Das leere Set für einen Leser', () => {
  it('sperrt das Einfügen und erklärt es', async () => {
    const einfuegen = vi.fn()

    render(
      <Sprachraum sprache="de">
        <Schreibrechtraum nurLesend>
          <Kartenleer
            ablage={{
              kopieren: vi.fn(),
              ausschneiden: vi.fn(),
              einfuegen,
              laeuft: false,
              anzahl: 0,
            }}
          />
        </Schreibrechtraum>
      </Sprachraum>,
    )

    const knopf = screen.getByRole('button', {
      name: kataloge.de.seiten.karten.ablage.einfuegen,
    })
    expect(knopf).toHaveAttribute('aria-disabled', 'true')
    expect(knopf).toHaveAccessibleDescription(
      kataloge.de.bausteine.nurlesen.knopf,
    )

    await userEvent.click(knopf)
    expect(einfuegen).not.toHaveBeenCalled()
  })
})
