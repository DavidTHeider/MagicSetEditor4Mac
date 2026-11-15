// **Eine Leiste für die Kennzahlen eines Bereichs** — AP-9
// (#2366),
// Entwurf `Bereiche09`: über Set, Sammlung, Deck und Paketen derselbe
// Baustein, **jede Zahl ein Ziel**.
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import { Kennzahlleiste } from './Kennzahlleiste'

describe('die Kennzahlleiste', () => {
  it('führt mit einem Klick zum Ziel einer Zahl', async () => {
    const hin = vi.fn()
    render(
      <Kennzahlleiste
        bezeichnung="Kennzahlen des Decks"
        kennzahlen={[
          {
            schluessel: 'hauptdeck',
            was: 'Hauptdeck',
            wert: '60',
            unter: 'Sideboard 15',
            hin,
          },
          { schluessel: 'pruefung', was: 'Prüfung', wert: 'gültig' },
        ]}
      />,
    )

    const leiste = screen.getByRole('toolbar', { name: 'Kennzahlen des Decks' })
    await userEvent.click(
      within(leiste).getByRole('button', { name: 'Hauptdeck 60 Sideboard 15' }),
    )
    expect(hin).toHaveBeenCalledTimes(1)

    // **Ohne Ziel kein Knopf** — eine Fläche, die wie einer aussieht und
    // nichts tut, wäre eine lose Bindung.
    expect(within(leiste).getByText('gültig')).toBeInTheDocument()
    expect(within(leiste).getAllByRole('button')).toHaveLength(1)
  })

  // **Laden in Endhöhe** (AP-4): die Zeile steht, unsichtbar gefüllt, und
  // sagt, dass sie lädt — sonst rückte, was darunter steht, beim Eintreffen.
  it('steht beim Laden in ihrer Endhöhe und verspricht kein Ziel', () => {
    const hin = vi.fn()
    render(
      <Kennzahlleiste
        bezeichnung="Kennzahlen der Sammlung"
        kennzahlen={[{ schluessel: 'wert', was: 'Wert', wert: undefined, hin }]}
      />,
    )

    const leiste = screen.getByRole('toolbar', {
      name: 'Kennzahlen der Sammlung',
    })
    expect(leiste).toHaveAttribute('aria-busy', 'true')
    expect(within(leiste).queryByRole('button')).toBeNull()
    expect(within(leiste).getByText('Wert')).toBeInTheDocument()
  })

  // **Eine Tabstelle für die ganze Leiste** — sonst rückten drei Zahlen
  // zwischen Kopf und Suchfeld der Sammlung, und der Kernweg `sammlung`
  // verlöre sein Ziel „≤ 13. Tabstelle“ (`docs/UX-ARCHITEKTUR.md`). Die
  // Pfeile gehen in ihr weiter, wie in jeder Symbolleiste.
  it('ist eine Tabstelle, und die Pfeile gehen in ihr weiter', async () => {
    const ziele = { a: vi.fn(), b: vi.fn(), c: vi.fn() }
    render(
      <>
        <button type="button">davor</button>
        <Kennzahlleiste
          bezeichnung="Kennzahlen"
          kennzahlen={[
            { schluessel: 'a', was: 'A', wert: '1', hin: ziele.a },
            { schluessel: 'x', was: 'X', wert: 'ohne Ziel' },
            { schluessel: 'b', was: 'B', wert: '2', hin: ziele.b },
            { schluessel: 'c', was: 'C', wert: '3', hin: ziele.c },
          ]}
        />
        <button type="button">danach</button>
      </>,
    )

    await userEvent.click(screen.getByRole('button', { name: 'davor' }))
    await userEvent.tab()
    expect(screen.getByRole('button', { name: 'A 1' })).toHaveFocus()
    await userEvent.tab()
    expect(screen.getByRole('button', { name: 'danach' })).toHaveFocus()

    await userEvent.tab({ shift: true })
    expect(screen.getByRole('button', { name: 'A 1' })).toHaveFocus()
    await userEvent.keyboard('{ArrowRight}')
    expect(screen.getByRole('button', { name: 'B 2' })).toHaveFocus()
    await userEvent.keyboard('{End}')
    expect(screen.getByRole('button', { name: 'C 3' })).toHaveFocus()
    await userEvent.keyboard('{ArrowRight}')
    expect(screen.getByRole('button', { name: 'A 1' })).toHaveFocus()
    await userEvent.keyboard('{ArrowLeft}')
    expect(screen.getByRole('button', { name: 'C 3' })).toHaveFocus()
    await userEvent.keyboard('{Home}')
    expect(screen.getByRole('button', { name: 'A 1' })).toHaveFocus()

    // **Wer wiederkommt, landet, wo er war** — die wandernde Tabstelle.
    await userEvent.keyboard('{ArrowRight}')
    await userEvent.tab()
    await userEvent.tab({ shift: true })
    expect(screen.getByRole('button', { name: 'B 2' })).toHaveFocus()
    await userEvent.keyboard('{Enter}')
    expect(ziele.b).toHaveBeenCalledTimes(1)
  })
})
