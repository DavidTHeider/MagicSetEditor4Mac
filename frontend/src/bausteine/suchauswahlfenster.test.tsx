// **Eine Auswahl mit Suche, Teil 6: was erst das Fenster zeigte** — UXR-F1
// (#2007),
// gemessen mit `uxrf1` am nativen Fenster.
//
// 1. **Nach dem Wählen war der Fokus fort**: das Feld ist gesperrt, solange
//    geschrieben wird, und WebKit nimmt einem gesperrten Feld den Fokus —
//    danach ging ein Pfeil ins Leere (#1990, dieselbe Klasse wie bei Schalter
//    und Segmenten). jsdom tut das nicht; der Test stellt es mit `blur()` nach.
// 2. **Beim Öffnen per Klick war eine fremde Zeile hervorgehoben**: das Öffnen
//    rollt die Liste, WebKit schickt danach ein `mousemove` an derselben
//    Stelle, und die Zeile unter dem ruhenden Zeiger galt als gewählt.
import { fireEvent, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import { Probe, feld } from '../test/suchauswahlprobe'

// zeile ist die Zeile an einer Stelle der Liste — ohne sie bricht der Lauf ab.
function zeile(
  passt: (zeile: HTMLElement, stelle: number) => boolean,
): HTMLElement {
  const gefunden = screen.getAllByRole('option').find(passt)
  if (gefunden === undefined) {
    throw new Error('keine passende Zeile in der Liste')
  }

  return gefunden
}

describe('Suchauswahl — am Fenster gemessen', () => {
  it('gibt dem Feld den Fokus zurück, wenn das Schreiben endet', async () => {
    const { rerender } = render(<Probe aendern={vi.fn()} />)

    await userEvent.click(feld())
    await userEvent.keyboard('{ArrowDown}{Enter}')

    const fokus = vi.spyOn(feld(), 'focus')
    feld().blur()
    rerender(<Probe angehalten aendern={vi.fn()} />)
    expect(document.body).toHaveFocus()

    rerender(<Probe aendern={vi.fn()} />)

    expect(feld()).toHaveFocus()
    expect(fokus).toHaveBeenCalledWith({ focusVisible: true })
  })

  it('nimmt keinen Fokus, den es beim Wählen nicht hatte', async () => {
    const { rerender } = render(
      <>
        <Probe aendern={vi.fn()} />
        <button type="button">Daneben</button>
      </>,
    )

    await userEvent.click(feld())
    await userEvent.click(zeile((_, stelle) => stelle === 2))
    screen.getByRole('button', { name: 'Daneben' }).focus()
    feld().blur()

    rerender(
      <>
        <Probe angehalten aendern={vi.fn()} />
        <button type="button">Daneben</button>
      </>,
    )
    rerender(
      <>
        <Probe aendern={vi.fn()} />
        <button type="button">Daneben</button>
      </>,
    )

    expect(feld()).not.toHaveFocus()
  })

  it('hebt bei einem Zeiger, der stillsteht, keine fremde Zeile hervor', () => {
    render(<Probe />)

    fireEvent.click(feld(), { clientX: 40, clientY: 12 })
    const gewaehlt = feld().getAttribute('aria-activedescendant')
    const fremde = zeile((kandidat) => kandidat.id !== gewaehlt)

    // **Die Probe hat eine Wahl, und die fremde Zeile ist eine andere** —
    // sonst prüfte der Lauf nichts.
    expect(gewaehlt).not.toBeNull()

    fireEvent.mouseMove(fremde, { clientX: 40, clientY: 12 })
    expect(feld()).toHaveAttribute('aria-activedescendant', gewaehlt)

    // **Bewegt er sich, folgt die Hervorhebung** wie bisher.
    fireEvent.mouseMove(fremde, { clientX: 41, clientY: 80 })
    expect(feld()).toHaveAttribute('aria-activedescendant', fremde.id)
  })
})
