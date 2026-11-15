// **Die Meldung eines Pflichtfelds schiebt nichts** — #2024, gefunden am
// nativen Fenster (Hüllenprobe `uxr04`).
//
// Ein leeres Pflichtfeld meldet sich beim Verlassen. Stand darunter ein Knopf,
// schob die Meldung ihn hinab, **bevor** die Maustaste losgelassen war — der
// erste Klick auf *Zurück* im Assistenten *Ein neues Set* landete daneben.
// Der Platz für die Meldung steht deshalb schon da, bevor sie kommt.
//
// **Was dieser Lauf nicht misst**: die Höhe. jsdom zeichnet nichts; ob der
// Platz so hoch ist wie die Meldung, belegt das Drehbuch `uxr04` am Fenster,
// das *Zurück* aus dem leeren Feld heraus drückt.
import { fireEvent, render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { Pflichtfeld } from './Pflichtfeld'

function zeige(hinweis?: string) {
  render(
    <Sprachraum sprache="de">
      <Pflichtfeld
        beschriftung="Name des Sets"
        wert=""
        geaendert={vi.fn()}
        {...(hinweis === undefined ? {} : { hinweis })}
      />
    </Sprachraum>,
  )
}

describe('Der Platz für die Meldung eines Pflichtfelds', () => {
  it('steht vor der Meldung da und weicht ihr', () => {
    zeige()

    const platz = document.querySelector('[data-befundplatz]')
    expect(platz).not.toBeNull()
    expect(platz).toHaveAttribute('aria-hidden', 'true')

    fireEvent.blur(screen.getByLabelText('Name des Sets'))

    expect(screen.getByRole('alert')).toBeInTheDocument()
    expect(document.querySelector('[data-befundplatz]')).toBeNull()
  })

  // **Auch mit Hinweis**: der Hinweis bleibt stehen, wenn die Meldung kommt
  // (`Feld` zeigt beide), die Meldung käme also darunter dazu.
  it('steht auch unter einem Hinweis', () => {
    zeige('Ein Name zum Wiederfinden.')

    expect(document.querySelector('[data-befundplatz]')).not.toBeNull()
  })
})
