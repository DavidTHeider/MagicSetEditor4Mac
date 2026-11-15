// **Ein Schriftwechsel zur Laufzeit erreicht die Fensterliste** — Nachlauf
// zum Zweitblick über Phase 6 von 0.9.0, zweiter Teil, R5
// (#2369).
//
// Die Sammlungsliste las die Wurzelschrift beim Zeichnen; das Fenster setzt
// die neue Größe aber erst danach in einem Effekt. Erst ein weiteres
// Zeichnen zog die Höhen nach — und dann stand die Rollstelle in alten
// Pixeln, und eine gehaltene Zeile konnte aus dem Ausschnitt fallen.
import { act, fireEvent, render, screen, within } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { keineHerkunft } from '../apiherkunft'
import type { ExemplarDTO } from '../api'
import { setzeSchriftgroesse } from '../erscheinungsbild'
import { leereSicht } from '../seiten/sammlungssicht'
import { Sammlungsliste } from '../seiten/Sammlungsliste'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { Fensterliste } from './Fensterliste'

const texte = kataloge.de.seiten.sammlung

let hoehe = 400
beforeEach(() => {
  Object.defineProperty(HTMLElement.prototype, 'clientHeight', {
    configurable: true,
    get: () => hoehe,
  })
})
afterEach(() => {
  hoehe = 400
  document.documentElement.style.fontSize = ''
  document.documentElement.style.removeProperty('--schriftgroesse')
})

describe('ein Schriftwechsel zur Laufzeit', () => {
  it('rechnet die Zeilen der Sammlung sofort neu', () => {
    const exemplar: ExemplarDTO = {
      id: 1,
      fassung: 1,
      printing: 'probe-druck',
      druckID: '',
      vorlaeufig: false,
      menge: 1,
      zustand: 'nearMint',
      sprache: 'en',
      finish: 'nonfoil',
      erwerbsdatum: '',
      erwerbspreis: '',
      erwerbswaehrung: '',
      notiz: '',
      name: 'Erste',
      setcode: 'm15',
      nummer: '',
      lagerort: '',
      wunsch: false,
      bezeichnung: 'Erste',
      herkunft: keineHerkunft,
    }
    render(
      <Sprachraum sprache="de">
        <Sammlungsliste
          exemplare={[exemplar]}
          sicht={leereSicht}
          gewaehlt={new Set()}
          anhaken={vi.fn()}
          gehalten={undefined}
          waehlen={vi.fn()}
        />
      </Sprachraum>,
    )
    const zeile = () =>
      within(screen.getByRole('grid', { name: texte.listeBezeichnung }))
        .getByText('Erste')
        .closest<HTMLElement>('[role="row"]')
    expect(zeile()?.style.height).toBe('40px')

    // **jsdom rechnet die Variable nicht in eine Schrift um** — die Größe
    // steht deshalb zusätzlich am Element; was geprüft wird, ist, dass die
    // Meldung von `setzeSchriftgroesse` ankommt.
    act(() => {
      document.documentElement.style.fontSize = '20px'
      setzeSchriftgroesse(document.documentElement, 'sehrgross')
    })

    expect(zeile()?.style.height).toBe('50px')
  })

  it('hält die Rollstelle, wenn die Zeilen höher werden', () => {
    const eintraege = Array.from({ length: 1000 }, (_, i) => ({
      id: i + 1,
      name: `Karte ${String(i + 1)}`,
    }))
    const zeichne = (zeilenhoehe: number) => (
      <Fensterliste
        bezeichnung="Exemplare"
        eintraege={eintraege}
        zeilenhoehe={zeilenhoehe}
        schluessel={(e) => e.id}
        spalten={['Name']}
        zeichne={(e) => <div role="gridcell">{e.name}</div>}
      />
    )
    const { rerender } = render(zeichne(40))
    const fenster = screen.getByTestId('fenster')
    fenster.scrollTop = 4000
    fireEvent.scroll(fenster)
    expect(screen.getByText('Karte 101')).toBeInTheDocument()

    rerender(zeichne(50))

    // **Dieselbe erste Zeile oben** — 100 Zeilen zu 50 Punkten.
    expect(fenster.scrollTop).toBe(5000)
    expect(screen.getByText('Karte 101')).toBeInTheDocument()
  })
})
