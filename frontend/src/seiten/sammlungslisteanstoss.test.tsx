// **Rechtsklick und ⇧F10 in allen drei Ansichten der Sammlung** — AP-6,
// Teil 2 (#2363). Die Liste reicht sie an `anstossen` weiter; was daraus
// wird, prüft `sammlungrechtsklick.test.tsx` an der Seite.
import { fireEvent, render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'

import { platzhalter } from '../haken/sammlungszeilen'
import { Sprachraum } from '../sprache/kontext'
import { leereSicht, type Sicht } from './sammlungssicht'
import { Sammlungsliste } from './Sammlungsliste'

const bestand = [
  { ...platzhalter(0), id: 1, bezeichnung: 'Erste', lagerort: 'Kiste A' },
  { ...platzhalter(1), id: 2, bezeichnung: 'Zweite', lagerort: 'Kiste A' },
]

function zeige(teil: Partial<Sicht>) {
  const anstossen = vi.fn()
  render(
    <Sprachraum sprache="de">
      <Sammlungsliste
        exemplare={bestand}
        sicht={{ ...leereSicht, ...teil }}
        gewaehlt={new Set()}
        anhaken={vi.fn()}
        gehalten={undefined}
        waehlen={vi.fn()}
        anstossen={anstossen}
      />
    </Sprachraum>,
  )

  return anstossen
}

describe('der Anstoß in der Sammlungsliste', () => {
  it('kommt in der Tabelle von der Zeile', () => {
    const anstossen = zeige({})

    fireEvent.contextMenu(screen.getByRole('row', { name: /Zweite/ }))

    expect(anstossen).toHaveBeenCalledWith(expect.anything(), 2)
  })

  // **Ein Kopf ist kein Exemplar** — sein Rechtsklick öffnet nichts.
  it('kommt in den Gruppen von einer Zeile und nicht vom Kopf', () => {
    const anstossen = zeige({ ansicht: 'gruppen', gruppierung: 'lagerort' })

    const kopf = screen.getByRole('rowheader').closest('[role="row"]')
    expect(kopf).toBeInstanceOf(HTMLElement)
    fireEvent.contextMenu(kopf ?? document.body)
    expect(anstossen).not.toHaveBeenCalled()

    fireEvent.contextMenu(screen.getByRole('row', { name: /Erste/ }))
    expect(anstossen).toHaveBeenCalledWith(expect.anything(), 1)
  })

  it('kommt in der Galerie von der Kachel, auch mit ⇧F10', () => {
    const anstossen = zeige({ ansicht: 'galerie' })
    const kachel = screen.getByRole('button', { name: 'Zweite wählen' })

    fireEvent.contextMenu(kachel)
    fireEvent.keyDown(kachel, { key: 'F10', shiftKey: true })

    expect(anstossen).toHaveBeenCalledTimes(2)
    expect(anstossen).toHaveBeenLastCalledWith(expect.anything(), 2)
  })
})
