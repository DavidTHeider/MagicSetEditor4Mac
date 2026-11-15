// **Eine Zeile ohne Platz zeigt keine Wahl** — seit UXR-11
// (#1919)
// steht die Wahl erst auf Wunsch, und die Seite fragt sie nur für Zeilen mit
// passendem Platz an. Der Baustein sagt es trotzdem selbst: ohne Platz gäbe
// es nichts, dem ein Exemplar zuzuordnen wäre.
import { render, screen } from '@testing-library/react'
import { expect, it, vi } from 'vitest'

import { deckprobe } from '../seiten/deckprobe'
import { Sprachraum } from '../sprache/kontext'
import { Zuordnungswahl } from './Zuordnungswahl'

it('zeichnet nichts, wenn kein Platz den Druck der Zeile nennt', () => {
  render(
    <Sprachraum sprache="de">
      <Zuordnungswahl
        deck={deckprobe({ plaetze: [] })}
        zeile={{
          printing: 'a1b2c3d4-0000-4000-8000-00000000000a',
          druckID: '4',
          vorlaeufig: false,
          bezeichnung: 'Blitz · M11 #149',
          gebraucht: 1,
          vorhanden: 0,
          fehlend: 1,
          grund: '',
          reserviert: 0,
          zugeordnet: 0,
          frei: 0,
          zuordnungen: [],
        }}
        exemplare={[]}
        angehalten={false}
        aufZuordnen={vi.fn()}
      />
    </Sprachraum>,
  )

  // **Der Sprachraum legt Mantines Stile ab** — gezählt wird, was der
  // Baustein zeichnet: kein Feld, kein Knopf.
  expect(screen.queryByRole('button')).toBeNull()
  expect(screen.queryByRole('combobox')).toBeNull()
  expect(screen.queryByRole('spinbutton')).toBeNull()
})
