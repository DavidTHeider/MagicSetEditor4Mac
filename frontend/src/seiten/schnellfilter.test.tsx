// **Das Filterfeld** — Werkstück U5 von Phase 5.1
// (#1416).
import { fireEvent, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { Schnellfilter } from './Schnellfilter'
import type { Schnellfilterstand } from '../haken/useSchnellfilter'

function stand(teile: Partial<Schnellfilterstand> = {}): Schnellfilterstand {
  return {
    anfrage: '',
    setzeAnfrage: vi.fn(),
    treffer: undefined,
    laedt: false,
    fehler: undefined,
    ...teile,
  }
}

function zeige(filter: Schnellfilterstand, gezeigt = 2, gesamt = 3) {
  render(
    <Sprachraum sprache="de">
      <Schnellfilter filter={filter} gesamt={gesamt} gezeigt={gezeigt} />
    </Sprachraum>,
  )
}

describe('Schnellfilter', () => {
  it('reicht jedes Zeichen weiter und beschreibt seine Sprache am Feld', async () => {
    const filter = stand()
    zeige(filter)

    const feld = screen.getByLabelText('Karten filtern')
    expect(feld).toHaveAccessibleDescription(/Leerraum trennt Teile/)

    await userEvent.type(feld, 'g')

    expect(filter.setzeAnfrage).toHaveBeenCalledWith('g')
    expect(
      screen.queryByRole('button', { name: 'Filter leeren' }),
    ).not.toBeInTheDocument()
    expect(screen.queryByRole('status')).not.toBeInTheDocument()
  })

  // **Die Hilfe hinter „?“** — AP-11 (#2368), B-27: sie stand dauerhaft
  // vierzeilig unter dem Feld. Das Vorleseprogramm hört sie weiter als
  // Beschreibung des Feldes.
  it('stellt die Hilfe hinter „?“ und behält sie als Beschreibung', () => {
    zeige(stand())

    const feld = screen.getByLabelText('Karten filtern')
    expect(feld).toHaveAccessibleDescription(/Leerraum trennt Teile/)
    expect(
      document.getElementById(feld.getAttribute('aria-describedby') ?? ''),
    ).toHaveClass('sr-only')
    expect(
      screen.getByRole('button', { name: 'Erklärung zu Karten filtern' }),
    ).toBeInTheDocument()
  })

  it('zeigt bei einer Anfrage die Trefferzahl und einen Knopf zum Leeren', async () => {
    const filter = stand({ anfrage: 'goblin', treffer: new Set(['1']) })
    zeige(filter, 1, 3)

    expect(screen.getByRole('status')).toHaveTextContent('1 von 3 Karten')

    await userEvent.click(screen.getByRole('button', { name: 'Filter leeren' }))

    expect(filter.setzeAnfrage).toHaveBeenCalledWith('')
  })

  // **Escape leert** — `FilterCtrl::onChar`, und nur, wenn etwas drinsteht.
  it('leert mit Escape, aber nur wenn etwas drinsteht', () => {
    const voll = stand({ anfrage: 'goblin', treffer: new Set() })
    const { unmount } = render(
      <Sprachraum sprache="de">
        <Schnellfilter filter={voll} gesamt={3} gezeigt={0} />
      </Sprachraum>,
    )

    fireEvent.keyDown(screen.getByLabelText('Karten filtern'), {
      key: 'Escape',
    })
    expect(voll.setzeAnfrage).toHaveBeenCalledWith('')

    unmount()

    const leer = stand()
    zeige(leer)
    fireEvent.keyDown(screen.getByLabelText('Karten filtern'), {
      key: 'Escape',
    })
    expect(leer.setzeAnfrage).not.toHaveBeenCalled()
  })

  it('zeigt einen Fehler statt der Trefferzahl', () => {
    zeige(stand({ anfrage: 'x', fehler: 'Das Set gibt es nicht.' }))

    expect(screen.getByRole('alert')).toHaveTextContent(
      'Das Set gibt es nicht.',
    )
    expect(screen.queryByRole('status')).not.toBeInTheDocument()
  })
})
