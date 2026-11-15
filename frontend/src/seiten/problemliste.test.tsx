// **Die Problemliste** — Werkstück U5 von Phase 5.1
// (#1416).
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { Problemliste, feldVon, problemeAus } from './Problemliste'

describe('problemeAus', () => {
  it('stellt den Feldfehler voran und liest das Feld aus jedem Rasterbefund', () => {
    const liste = problemeAus({ feld: 'name', meldung: 'Der Name fehlt.' }, [
      'Ebene 3 (rule text): die Schrift MPlantin ist nicht zu finden',
      'Ebene 7: der Hintergrund fehlt',
    ])

    expect(liste).toEqual([
      { feld: 'name', meldung: 'Der Name fehlt.' },
      {
        feld: 'rule text',
        meldung:
          'Ebene 3 (rule text): die Schrift MPlantin ist nicht zu finden',
      },
      { feld: undefined, meldung: 'Ebene 7: der Hintergrund fehlt' },
    ])
  })

  it('ist ohne Fehler und Befunde leer', () => {
    expect(problemeAus(undefined, [])).toEqual([])
    expect(feldVon('irgendwas')).toBeUndefined()
  })
})

describe('Problemliste', () => {
  it('sagt, dass nichts anliegt', () => {
    render(
      <Sprachraum sprache="de">
        <Problemliste probleme={[]} zumFeld={vi.fn()} />
      </Sprachraum>,
    )

    expect(screen.getByRole('region', { name: 'Probleme' })).toHaveTextContent(
      'Keine Probleme an dieser Karte.',
    )
  })

  it('führt zum Feld — und hat für einen Befund ohne Feld keinen Knopf', async () => {
    const zumFeld = vi.fn()
    render(
      <Sprachraum sprache="de">
        <Problemliste
          probleme={[
            { feld: 'name', meldung: 'Der Name fehlt.' },
            { feld: undefined, meldung: 'Ebene 7: der Hintergrund fehlt' },
          ]}
          zumFeld={zumFeld}
        />
      </Sprachraum>,
    )

    await userEvent.click(screen.getByRole('button', { name: 'Zum Feld name' }))

    expect(zumFeld).toHaveBeenCalledWith('name')
    expect(screen.getAllByRole('button')).toHaveLength(1)
    expect(screen.getByText('ohne Feldbezug')).toBeInTheDocument()
  })
})
