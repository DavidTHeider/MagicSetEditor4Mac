// **Der Arbeitsplatz reicht bis an den Fensterrand, und jede Spalte rollt für
// sich** — UX-15 (#1771).
//
// Gemessen am 2026-09-24 in WebKit mit dem echten Bestand: der Vorlagenbaum
// mit 536 Paketen machte die Seite **60.317 Punkte** hoch. Wer unten im Baum
// eine Datei wählte, sah sie oben im Zentrum aufgehen — vierzig Bildschirme
// weiter oben. Die Artboards *Karten*, *Werkstatt*, *Sammlung* und *Decks*
// geben jeder Spalte die Höhe des Fensters.
import { render, screen } from '@testing-library/react'
import { afterEach, beforeEach, expect, it, vi } from 'vitest'

import { Arbeitsplatz } from './Arbeitsplatz'
import { gibJsdomEineBreite } from '../test/sichtbreite'
import { Sprachraum } from '../sprache/kontext'

let raeumeBreite: () => void = () => undefined

beforeEach(() => {
  raeumeBreite = gibJsdomEineBreite(1400)
})

afterEach(() => {
  raeumeBreite()
  vi.restoreAllMocks()
  vi.unstubAllGlobals()
})

function Probe() {
  return (
    <Sprachraum sprache="de">
      <Arbeitsplatz
        bezeichnung="Vorlagenwerkstatt"
        navigator={{ bezeichnung: 'Baum', inhalt: <p>Der Baum</p> }}
        zentrum={<p>Die Datei</p>}
        inspector={{ bezeichnung: 'Auskunft', inhalt: <p>Die Auskunft</p> }}
      />
    </Sprachraum>
  )
}

it('gibt den drei Spalten die Höhe bis an den Fensterrand und jeder ihr eigenes Rollen', async () => {
  vi.stubGlobal('innerHeight', 900)
  vi.spyOn(HTMLElement.prototype, 'getBoundingClientRect').mockReturnValue({
    top: 200,
  } as DOMRect)

  render(<Probe />)

  const navigator = screen.getByRole('complementary', { name: 'Baum' })
  const zentrum = screen.getByText('Die Datei').parentElement
  const inspector = screen.getByRole('complementary', { name: 'Auskunft' })

  // 900 Fenster, 200 darüber, 32 Innenrand unter `main`.
  await vi.waitFor(() => {
    expect(navigator.parentElement).toHaveStyle({ height: '668px' })
  })

  for (const spalte of [navigator, zentrum, inspector]) {
    expect(spalte).toHaveClass('overflow-y-auto', 'min-h-0')
  }
})

// **Ein flaches Fenster zerdrückt die Spalten nicht**: unter der Mindesthöhe
// rollt die Seite, statt dass eine Spalte zum Schlitz wird.
it('lässt die Spalten nicht unter die Mindesthöhe fallen', async () => {
  vi.stubGlobal('innerHeight', 500)
  vi.spyOn(HTMLElement.prototype, 'getBoundingClientRect').mockReturnValue({
    top: 300,
  } as DOMRect)

  render(<Probe />)

  const navigator = screen.getByRole('complementary', { name: 'Baum' })

  await vi.waitFor(() => {
    expect(navigator.parentElement).toHaveStyle({ height: '320px' })
  })
})
