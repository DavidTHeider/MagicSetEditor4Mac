// **Die Trenner im Arbeitsplatz** — AP-12
// (#2369),
// Befund B-25. Zwischen Navigator und Zentrum und zwischen Zentrum und
// Inspector steht je ein Trenner; was er wünscht, geht ins Gedächtnis, und
// die obere Grenze lässt dem Zentrum die 480 Punkte aus Vermerk 0076.
import { fireEvent, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { Arbeitsplatz } from './Arbeitsplatz'
import { effekteNachholen } from '../test/effekte'
import { gibJsdomEineBreite } from '../test/sichtbreite'
import { Sprachraum } from '../sprache/kontext'

import type { Arbeitsplatzgedaechtnis } from '../haken/useArbeitsplatzgedaechtnis'

let raeumeBreite: () => void = () => undefined

beforeEach(() => {
  raeumeBreite = gibJsdomEineBreite(1400)
  Element.prototype.setPointerCapture = vi.fn()
})

afterEach(() => {
  raeumeBreite()
})

function gedaechtnis(
  teile: Partial<Arbeitsplatzgedaechtnis> = {},
): Arbeitsplatzgedaechtnis {
  return {
    geladen: true,
    bereit: true,
    navigator: true,
    inspector: true,
    navigatorbreite: 300,
    inspectorbreite: 400,
    fehler: undefined,
    laeuft: () => false,
    merke: vi.fn(),
    breiteWuenschen: vi.fn(),
    ...teile,
  }
}

function zeige(stand: Arbeitsplatzgedaechtnis | undefined) {
  render(
    <Sprachraum sprache="de">
      <Arbeitsplatz
        bezeichnung="Kartenarbeitsplatz"
        navigator={{ bezeichnung: 'Karten', inhalt: <p>Die Liste</p> }}
        zentrum={<p>Die Vorschau</p>}
        inspector={{ bezeichnung: 'Felder', inhalt: <p>Das Formular</p> }}
        {...(stand === undefined ? {} : { gedaechtnis: stand })}
      />
    </Sprachraum>,
  )
}

describe('die Trenner im Arbeitsplatz', () => {
  it('stehen zwischen den Spalten, je mit ihrer Spalte und Breite', async () => {
    zeige(gedaechtnis())
    await effekteNachholen()

    const links = screen.getByRole('separator', {
      name: 'Breite von „Karten“',
    })
    const rechts = screen.getByRole('separator', {
      name: 'Breite von „Felder“',
    })
    expect(links).toHaveAttribute('aria-valuenow', '300')
    expect(links).toHaveAttribute('aria-valuetext', '300 Punkte')
    expect(rechts).toHaveAttribute('aria-valuenow', '400')
    expect(links).toHaveAttribute(
      'aria-controls',
      screen.getByRole('complementary', { name: 'Karten' }).id,
    )
    expect(rechts).toHaveAttribute(
      'aria-controls',
      screen.getByRole('complementary', { name: 'Felder' }).id,
    )
  })

  // **Dem Zentrum bleiben 480 Punkte** (Vermerk 0076): 1.400 Breite, der
  // Inspector 400, zwei Trenner zu 16 — für den Navigator bleiben 488.
  it('lässt dem Zentrum seine 480 Punkte', async () => {
    zeige(gedaechtnis())
    await effekteNachholen()

    expect(
      screen.getByRole('separator', { name: 'Breite von „Karten“' }),
    ).toHaveAttribute('aria-valuemax', '488')
    expect(
      screen.getByRole('separator', { name: 'Breite von „Felder“' }),
    ).toHaveAttribute('aria-valuemax', '588')
  })

  it('wünscht mit den Pfeiltasten die Breite im Gedächtnis', async () => {
    const stand = gedaechtnis()
    zeige(stand)
    await effekteNachholen()

    screen.getByRole('separator', { name: 'Breite von „Felder“' }).focus()
    await userEvent.keyboard('{ArrowLeft}')

    expect(stand.breiteWuenschen).toHaveBeenCalledExactlyOnceWith(
      'inspector',
      416,
    )
  })

  it('lässt die Spalte beim Ziehen mitwachsen und schreibt beim Loslassen', async () => {
    const stand = gedaechtnis()
    zeige(stand)
    await effekteNachholen()
    const trenner = screen.getByRole('separator', {
      name: 'Breite von „Karten“',
    })
    const spalte = screen.getByRole('complementary', { name: 'Karten' })

    fireEvent.pointerDown(trenner, { clientX: 300, pointerId: 1, button: 0 })
    fireEvent.pointerMove(trenner, { clientX: 340, pointerId: 1 })

    expect(spalte).toHaveStyle({ width: '340px' })
    expect(stand.breiteWuenschen).not.toHaveBeenCalled()

    fireEvent.pointerUp(trenner, { clientX: 340, pointerId: 1 })

    expect(stand.breiteWuenschen).toHaveBeenCalledExactlyOnceWith(
      'navigator',
      340,
    )
    expect(spalte).toHaveStyle({ width: '300px' })
  })

  it('steht nur an einer offenen Spalte', async () => {
    zeige(gedaechtnis({ inspector: false }))
    await effekteNachholen()

    expect(screen.getAllByRole('separator')).toHaveLength(1)
    expect(
      screen.getByRole('separator', { name: 'Breite von „Karten“' }),
    ).toBeInTheDocument()
  })

  // **Ohne Gedächtnis kein Trenner** — eine Breite, die niemand behält,
  // spränge beim nächsten Öffnen zurück.
  it('fehlt ohne Gedächtnis', async () => {
    zeige(undefined)
    await effekteNachholen()

    expect(screen.getByText('Die Vorschau')).toBeInTheDocument()
    expect(screen.queryByRole('separator')).toBeNull()
  })
})
