// **Ein Balken in der Sprache des Fensters** — LOC-03.AK-01
// (#1689).
//
// **Das Prozentzeichen gehört zur Zahl.** Hier stand im Quelltext
// »{karten} ({prozent} %)«, und das ist im Englischen falsch: dort steht
// kein Abstand. Im Deutschen steht einer — ein **geschützter**, damit das
// Zeichen nicht allein in die nächste Zeile rutscht (gemessen am
// 2026-09-22).
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { Balken } from './Balken'

it.each([
  ['de', 'de', '121 (39\u00a0%)'],
  ['en', 'en', '121 (39%)'],

  // **Dieselbe Sprache, eine andere Gegend** — LOC-03.AK-03 (#1689): die
  // Schweiz schreibt Prozent ohne Abstand.
  ['de', 'de-CH', '121 (39%)'],
] as const)(
  '%s in %s schreibt den Anteil als %j',
  (sprache, darstellung, erwartet) => {
    render(
      <Sprachraum sprache={sprache} darstellung={darstellung}>
        <ul>
          <Balken name="common" karten={121} ganzes={312} anteil="121/312" />
        </ul>
      </Sprachraum>,
    )

    expect(screen.getByRole('listitem').textContent).toContain(erwartet)
  },
)

// **Ein wählbarer Balken ist ein Knopf** — UXR-F6
// (#2013).
//
// Er sagt über `aria-pressed`, ob er gewählt ist, antwortet auf Klick und
// Tastatur (ein Knopf tut beides) und trägt weiter Zahl und Anteil als Text.
// Ein Balken ohne Karten wählt nichts, und der Titel sagt warum.
describe('Ein wählbarer Balken', () => {
  function waehlbarBalken(
    karten: number,
    teile: { gewaehlt?: boolean; waehlen?: () => void } = {},
  ) {
    render(
      <Sprachraum sprache="de">
        <ul>
          <Balken
            name="rare"
            karten={karten}
            ganzes={12}
            anteil="Anteil"
            waehlbar={{
              gewaehlt: teile.gewaehlt ?? false,
              waehlen: teile.waehlen ?? vi.fn(),
              leerTitel: 'Keine Karte trägt diesen Wert.',
            }}
          />
        </ul>
      </Sprachraum>,
    )
  }

  it('ist ein Knopf mit Zahl und Anteil im Namen', () => {
    waehlbarBalken(3)

    expect(
      screen.getByRole('button', { name: /rare.*3 \(25\s%\)/ }),
    ).toHaveAttribute('aria-pressed', 'false')
  })

  it('sagt, dass er gewählt ist', () => {
    waehlbarBalken(3, { gewaehlt: true })

    expect(screen.getByRole('button')).toHaveAttribute('aria-pressed', 'true')
  })

  it('wählt mit der Maus und mit der Tastatur', async () => {
    const waehlen = vi.fn()

    waehlbarBalken(3, { waehlen })

    // Tab, dann Enter und Leertaste: das tut ein Knopf, ohne dass jemand es
    // verdrahtet — hier steht, dass es dabei bleibt.
    await userEvent.tab()
    expect(screen.getByRole('button')).toHaveFocus()

    await userEvent.keyboard('{Enter}')
    await userEvent.keyboard(' ')
    expect(waehlen).toHaveBeenCalledTimes(2)

    await userEvent.click(screen.getByRole('button'))
    expect(waehlen).toHaveBeenCalledTimes(3)
  })

  it('wählt nichts, wenn keine Karte den Wert trägt, und sagt warum', async () => {
    const waehlen = vi.fn()

    waehlbarBalken(0, { waehlen })

    const knopf = screen.getByRole('button')
    expect(knopf).toBeDisabled()
    expect(knopf).toHaveAttribute('title', 'Keine Karte trägt diesen Wert.')

    await userEvent.click(knopf)
    expect(waehlen).not.toHaveBeenCalled()
  })

  // **Ohne `waehlbar` bleibt es ein Zeilenpunkt zum Lesen** — die Belegung
  // zählt Felder, und die sind keine Gruppen.
  it('ist ohne Wahl kein Knopf', () => {
    render(
      <Sprachraum sprache="de">
        <ul>
          <Balken name="name" karten={3} ganzes={12} anteil="Anteil" />
        </ul>
      </Sprachraum>,
    )

    expect(screen.queryByRole('button')).not.toBeInTheDocument()
  })
})
