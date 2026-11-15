// **Die Sicht im Navigator** — Suche und Ausführung, und wie viel sie
// wegnimmt (#1200, U8 von Phase 5.1).
import { render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'

import { Sammlungssicht } from './Sammlungssicht'
import { leereSicht, type Sicht } from '../seiten/sammlungssicht'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'

const texte = kataloge.de.seiten.sammlung

function zeige(teil: Partial<Sicht> = {}) {
  const setze = vi.fn()
  render(
    <Sprachraum sprache="de">
      <Sammlungssicht
        sicht={{ ...leereSicht, ...teil }}
        setze={setze}
        gezeigt={3}
        gesamt={3}
        zuruecksetzen={vi.fn()}
      />
    </Sprachraum>,
  )

  return setze
}

// **Ansicht und Gruppierung stehen seit UXR-10 (#1918) über der Liste** —
// ihre Läufe stehen in `sammlungsleiste.test.tsx`.

// **Die Zahl der gezeigten Zeilen wird angesagt** — UX-10.AK-02 (#1682):
// wer tippt oder filtert, hört „3 von 3", ohne die Liste durchzugehen.
// Dieselbe Gestalt wie die Paketliste und der Schnellfilter der Karten.
describe('Die Zahl der gezeigten Zeilen', () => {
  it('steht als Status da', () => {
    zeige()

    expect(screen.getByRole('status')).toHaveTextContent(
      texte.gezeigt.replace('{gezeigt}', '3').replace('{gesamt}', '3'),
    )
  })
})
