// **Eine Auswahl mit Suche, Teil 3: die Maus** — UXR-F1
// (#2007).
//
// Klick in die Liste, Zeiger über einer Zeile, Klick daneben und das
// Verlassen des Feldes. Tasten in `suchauswahltasten.test.tsx`, Gestalt in
// `suchauswahl.test.tsx`.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import { Probe, feld } from '../test/suchauswahlprobe'

describe('Suchauswahl — Maus', () => {
  it('wählt auf einen Klick in die Liste', async () => {
    const aendern = vi.fn()
    render(<Probe aendern={aendern} />)

    await userEvent.click(feld())
    await userEvent.click(screen.getByRole('option', { name: 'Africa/Cairo' }))

    expect(aendern).toHaveBeenCalledExactlyOnceWith('Africa/Cairo')
    expect(feld()).toHaveValue('Africa/Cairo')
    expect(screen.queryByRole('listbox')).not.toBeInTheDocument()
    // **Der Fokus bleibt**, wo er war: die Liste nimmt ihn nie.
    expect(feld()).toHaveFocus()
  })

  it('wählt auch die Zeile, die „wie das System“ heißt, mit leerem Wert', async () => {
    const aendern = vi.fn()
    render(<Probe aendern={aendern} />)

    await userEvent.click(feld())
    await userEvent.click(
      screen.getByRole('option', { name: 'Wie das System' }),
    )

    expect(aendern).toHaveBeenCalledExactlyOnceWith('')
    expect(feld()).toHaveValue('Wie das System')
  })

  it('führt die Hervorhebung dem Zeiger nach', async () => {
    render(<Probe />)
    await userEvent.click(feld())

    // **Der Zeiger bewegt sich wirklich** — einer, der stillsteht, hebt nichts
    // hervor (`suchauswahlfenster.test.tsx`).
    await userEvent.pointer({
      target: screen.getByRole('option', { name: 'Africa/Cairo' }),
      coords: { clientX: 30, clientY: 90 },
    })

    expect(feld()).toHaveAttribute(
      'aria-activedescendant',
      screen.getByRole('option', { name: 'Africa/Cairo' }).id,
    )
  })

  it('schließt auf einen Klick daneben und stellt die Eingabe her', async () => {
    render(
      <>
        <Probe />
        <button type="button">Daneben</button>
      </>,
    )

    await userEvent.click(feld())
    await userEvent.keyboard('paris')
    await userEvent.click(screen.getByRole('button', { name: 'Daneben' }))

    expect(screen.queryByRole('listbox')).not.toBeInTheDocument()
    expect(feld()).toHaveValue('Europe/Berlin')
  })

  it('schließt, wenn der Fokus das Feld verlässt', async () => {
    const aendern = vi.fn()
    render(
      <>
        <Probe aendern={aendern} />
        <button type="button">Danach</button>
      </>,
    )

    await userEvent.click(feld())
    await userEvent.keyboard('paris')
    await userEvent.tab()

    expect(screen.getByRole('button', { name: 'Danach' })).toHaveFocus()
    expect(screen.queryByRole('listbox')).not.toBeInTheDocument()
    expect(feld()).toHaveValue('Europe/Berlin')
    expect(aendern).not.toHaveBeenCalled()
  })
})
