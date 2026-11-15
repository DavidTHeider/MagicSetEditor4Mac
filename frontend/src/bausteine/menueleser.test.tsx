// **Ein Menüeintrag, der schreibt, ist für einen Leser gesperrt** — #2288
// (#2288),
// aus dem Nachlauf über Phase G.
//
// G10 sperrte Knöpfe mit `schreibt`; die Einträge eines Menüs kannten die
// Sperre nicht, und die Handlungszeile der Karten bot einem Leser
// Duplizieren, Ausschneiden, Einfügen und Löschen an. Gesperrt heißt wie am
// Knopf: `aria-disabled`, **fokussierbar**, der Grund als Beschreibung — und
// die Handlung läuft nicht.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import { Schreibrechtraum } from '../haken/schreibrecht'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { Menueliste } from './Menueliste'

function zeige(nurLesend: boolean) {
  const einfuegen = vi.fn()
  const kopieren = vi.fn()
  const geschlossen = vi.fn()

  render(
    <Sprachraum sprache="de">
      <Schreibrechtraum nurLesend={nurLesend}>
        <Menueliste
          id="probe"
          bezeichnung="Probe"
          stelle={0}
          aufTaste={vi.fn()}
          aufWahl={geschlossen}
          eintraege={[
            {
              schluessel: 'einfuegen',
              name: 'Einfügen',
              aufWahl: einfuegen,
              schreibt: true,
            },
            { schluessel: 'kopieren', name: 'Kopieren', aufWahl: kopieren },
          ]}
        />
      </Schreibrechtraum>
    </Sprachraum>,
  )

  return { einfuegen, kopieren, geschlossen }
}

describe('Ein Menü für einen Leser', () => {
  it('sperrt den schreibenden Eintrag und erklärt ihn', async () => {
    const { einfuegen, kopieren, geschlossen } = zeige(true)

    const gesperrt = screen.getByRole('menuitem', { name: 'Einfügen' })
    expect(gesperrt).toHaveAttribute('aria-disabled', 'true')
    expect(gesperrt).toHaveAccessibleDescription(
      kataloge.de.bausteine.nurlesen.knopf,
    )

    await userEvent.click(gesperrt)
    expect(einfuegen).not.toHaveBeenCalled()
    expect(geschlossen).not.toHaveBeenCalled()

    // **Der lesende daneben geht** — Beleg, dass jemand zuhört (§3).
    await userEvent.click(screen.getByRole('menuitem', { name: 'Kopieren' }))
    expect(kopieren).toHaveBeenCalledTimes(1)
  })

  it('lässt ihn einem Mitarbeiter offen', async () => {
    const { einfuegen } = zeige(false)

    const offen = screen.getByRole('menuitem', { name: 'Einfügen' })
    expect(offen).not.toHaveAttribute('aria-disabled')

    await userEvent.click(offen)
    expect(einfuegen).toHaveBeenCalledTimes(1)
  })
})
