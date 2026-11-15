// **Jedes Zonenfeld trägt eine unsichtbare Marke über der Leinwand** — A04
// (#2218).
//
// # Wofür
//
// Ein leerer Friedhof steht nur auf der Leinwand; wer eine Karte dorthin
// zieht — eine Probe mit der Maus der Abnahmeleitung —, braucht seinen Ort
// in Punkten der Seite. Die Karten haben ihn schon (`Kartenknopf`), die
// Zonen bis A04 nicht.
//
// # Was sie nicht dürfen
//
// **Den Zeiger nehmen** und **vorgelesen werden**: sie liegen über der
// Fläche, die selbst zieht, und sie sind keine Bedienelemente. Dieselbe
// Gestalt wie `data-einfuehrung` und `data-palettenzeile`.
import { render, screen, waitFor } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { einObjekt, einePartie } from '../seiten/partieprobe'
import { Sprachraum } from '../sprache/kontext'
import { Tischflaeche } from './Tischflaeche'
import { tischplan } from '../tisch/tischplan'
import { einMaler } from '../test/malerprobe'

import type { Tischmalerei } from '../tisch/maler'

const plan = tischplan(
  einePartie({
    objekte: [
      einObjekt({
        kennung: 'a',
        zone: 'hand',
        name: 'Wald',
        eigner: 'sitz1',
        beherrscher: 'sitz1',
        geheim: false,
      }),
    ],
  }),
)

describe('die Marken der Zonenfelder', () => {
  it('stehen je Feld über der Leinwand, ohne Zeiger und ohne Stimme', async () => {
    const maler = einMaler({ breite: 800, hoehe: 600 })

    render(
      <Sprachraum sprache="de">
        <Tischflaeche
          plan={plan}
          bild={() => undefined}
          malerei={() => Promise.resolve<Tischmalerei>({ art: 'da', maler })}
        />
      </Sprachraum>,
    )

    const flaeche = await screen.findByRole('application')
    await waitFor(() => {
      expect(maler.zeichne).toHaveBeenCalled()
    })

    const marken = flaeche.querySelectorAll('[data-zonenfeld]')
    expect(marken).toHaveLength(plan.felder.length)

    const friedhof = flaeche.querySelector(
      '[data-zonenfeld="friedhof"][data-sitz="sitz1"]',
    )
    expect(friedhof).not.toBeNull()
    expect(friedhof).toHaveAttribute('aria-hidden', 'true')
    expect(friedhof).toHaveClass('pointer-events-none')
    expect(friedhof?.getAttribute('style')).toMatch(/width: \d/)
  })
})
