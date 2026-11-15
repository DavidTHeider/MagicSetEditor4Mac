// **Woher etwas kam, im Fenster** — Werkstück 12 von Phase 4.12
// (#1351).
import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { Herkunftszeile } from './Herkunftszeile'
import { Sprachraum } from '../sprache/kontext'
import { keineHerkunft } from '../apiherkunft'
import { zeichnetNichts } from '../test/leererbaum'

function zeige(herkunft = keineHerkunft) {
  render(
    <Sprachraum sprache="de">
      <Herkunftszeile herkunft={herkunft} />
    </Sprachraum>,
  )
}

describe('Die Herkunftszeile', () => {
  it('nennt das Profil, aus dem etwas übernommen wurde', () => {
    zeige({ ...keineHerkunft, art: 'profil', quelle: 'Ada' })

    expect(screen.getByText(/Ada/)).toBeInTheDocument()
  })

  it('nennt die Datei, aus der etwas eingelesen wurde', () => {
    zeige({ ...keineHerkunft, art: 'einfuhr', quelle: 'sammlung.csv' })

    expect(screen.getByText(/sammlung\.csv/)).toBeInTheDocument()
  })

  it('zeigt gar nichts, wo etwas hier entstanden ist', () => {
    // **Und nicht „Herkunft: —"** — das wäre eine Auskunft über nichts (§6).
    const { container } = render(
      <Sprachraum sprache="de">
        <Herkunftszeile herkunft={keineHerkunft} />
      </Sprachraum>,
    )

    expect(zeichnetNichts(container)).toBe(true)
  })
})
