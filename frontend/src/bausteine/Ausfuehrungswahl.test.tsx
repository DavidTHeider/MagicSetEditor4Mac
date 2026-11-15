// **Die Wahl der Ausführung** — Werkstück T2 von Phase 5.6
// (#1467).
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import { Ausfuehrungswahl } from './Ausfuehrungswahl'

describe('Ausfuehrungswahl', () => {
  it('zeigt alle drei Ausführungen', () => {
    render(<Ausfuehrungswahl wert="nonfoil" setzeWert={vi.fn()} />)

    for (const wort of ['Nonfoil', 'Foil', 'Etched']) {
      expect(screen.getByRole('option', { name: wort })).toBeInTheDocument()
    }
  })

  // **Der Nullwert ist `nonfoil` und keine leere Zeile.** Ein Auswahlfeld ohne
  // passenden Eintrag zeigt sonst den ersten — und der Entwurf sagte etwas
  // anderes als das Feld.
  it('steht ohne Angabe auf nonfoil', () => {
    render(<Ausfuehrungswahl wert={'' as 'nonfoil'} setzeWert={vi.fn()} />)

    expect(screen.getByRole('combobox')).toHaveValue('nonfoil')
  })

  it('meldet die gewählte Ausführung', async () => {
    const setzeWert = vi.fn()

    render(<Ausfuehrungswahl wert="nonfoil" setzeWert={setzeWert} />)

    await userEvent.selectOptions(screen.getByRole('combobox'), 'etched')

    expect(setzeWert).toHaveBeenCalledWith('etched')
  })
})
