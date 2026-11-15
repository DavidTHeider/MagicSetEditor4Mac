// **Der Trenner zwischen zwei Spalten** — AP-12
// (#2369),
// Befund B-25: „Ein Trenner zwischen den Regionen, ziehbar und mit
// Pfeiltasten, der in dieselben Einstellungen schreibt.“ Die Breiten standen
// bis dahin nur als Zahl in den Einstellungen.
//
// **adapted** aus MSE2: dort trennt ein `wxSplitterWindow` Liste und Tafel,
// mit einer Mindestbreite und nur mit der Maus (`src/gui/set/cards_panel.cpp`
// 52–55, `src/gui/set/keywords_panel.cpp` 88–91, Fassung f9d9356 vom
// 2021-07-06); gemerkt wird dort nichts. Hier kommen Tastatur und Gedächtnis
// dazu, die Grenzen aus `domain.BreiteMindestens/Hoechstens`.
import { fireEvent, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { Spaltentrenner } from './Spaltentrenner'

beforeEach(() => {
  Element.prototype.setPointerCapture = vi.fn()
})

function zeige(teile: Partial<Parameters<typeof Spaltentrenner>[0]> = {}): {
  aendern: ReturnType<typeof vi.fn>
  vorschau: ReturnType<typeof vi.fn>
} {
  const aendern = vi.fn()
  const vorschau = vi.fn()

  render(
    <Spaltentrenner
      bezeichnung="Breite von „Karten“"
      wertetext="240 Punkte"
      steuert="region-1"
      breite={240}
      mindestens={120}
      hoechstens={800}
      richtung={1}
      aendern={aendern}
      vorschau={vorschau}
      {...teile}
    />,
  )

  return { aendern, vorschau }
}

describe('der Spaltentrenner', () => {
  it('ist ein Trenner mit Wert, Grenzen und der Spalte, die er steuert', () => {
    zeige()

    const trenner = screen.getByRole('separator', {
      name: 'Breite von „Karten“',
    })
    expect(trenner).toHaveAttribute('aria-orientation', 'vertical')
    expect(trenner).toHaveAttribute('aria-valuenow', '240')
    expect(trenner).toHaveAttribute('aria-valuemin', '120')
    expect(trenner).toHaveAttribute('aria-valuemax', '800')
    expect(trenner).toHaveAttribute('aria-valuetext', '240 Punkte')
    expect(trenner).toHaveAttribute('aria-controls', 'region-1')
    expect(trenner).toHaveAttribute('tabindex', '0')
  })

  it('geht mit den Pfeiltasten in Schritten von 16, mit ⇧ von 64', async () => {
    const { aendern } = zeige()
    screen.getByRole('separator').focus()

    await userEvent.keyboard('{ArrowRight}')
    expect(aendern).toHaveBeenLastCalledWith(256)

    await userEvent.keyboard('{ArrowLeft}')
    expect(aendern).toHaveBeenLastCalledWith(224)

    await userEvent.keyboard('{Shift>}{ArrowRight}{/Shift}')
    expect(aendern).toHaveBeenLastCalledWith(304)
  })

  // **Rechts liegt die andere Spalte** — beim Inspector wächst die Breite
  // nach links, und ← macht ihn breiter.
  it('dreht die Richtung für die Spalte rechts', async () => {
    const { aendern } = zeige({ breite: 320, richtung: -1 })
    screen.getByRole('separator').focus()

    await userEvent.keyboard('{ArrowLeft}')

    expect(aendern).toHaveBeenLastCalledWith(336)
  })

  it('springt mit Pos1 und Ende an die Grenzen', async () => {
    const { aendern } = zeige({ hoechstens: 500 })
    screen.getByRole('separator').focus()

    await userEvent.keyboard('{Home}')
    expect(aendern).toHaveBeenLastCalledWith(120)

    await userEvent.keyboard('{End}')
    expect(aendern).toHaveBeenLastCalledWith(500)
  })

  // **Die Grenze hält** — und ein Schritt, der nichts ändert, schreibt
  // nichts.
  it('bleibt an der Grenze stehen und schreibt dort nichts', async () => {
    const { aendern } = zeige({ breite: 800 })
    screen.getByRole('separator').focus()

    await userEvent.keyboard('{ArrowRight}')
    expect(aendern).not.toHaveBeenCalled()

    // **Ein zweites Ereignis durch denselben Weg**, dessen Wirkung man sieht.
    await userEvent.keyboard('{ArrowLeft}')
    expect(aendern).toHaveBeenCalledExactlyOnceWith(784)
  })

  // **Beim Ziehen wächst die Spalte mit, geschrieben wird beim Loslassen** —
  // sonst ginge jede Bewegung als Schreibvorgang hinaus.
  it('zeigt beim Ziehen die Breite und schreibt sie einmal, beim Loslassen', () => {
    const { aendern, vorschau } = zeige()
    const trenner = screen.getByRole('separator')

    fireEvent.pointerDown(trenner, { clientX: 300, pointerId: 1, button: 0 })
    fireEvent.pointerMove(trenner, { clientX: 350, pointerId: 1 })
    expect(vorschau).toHaveBeenLastCalledWith(290)
    fireEvent.pointerMove(trenner, { clientX: 1300, pointerId: 1 })
    expect(vorschau).toHaveBeenLastCalledWith(800)
    expect(aendern).not.toHaveBeenCalled()

    fireEvent.pointerUp(trenner, { clientX: 1300, pointerId: 1 })

    expect(aendern).toHaveBeenCalledExactlyOnceWith(800)
    expect(vorschau).toHaveBeenLastCalledWith(undefined)
  })

  it('zieht beim Inspector nach links breiter und hält die untere Grenze', () => {
    const { aendern } = zeige({ breite: 320, richtung: -1 })
    const trenner = screen.getByRole('separator')

    fireEvent.pointerDown(trenner, { clientX: 900, pointerId: 1, button: 0 })
    fireEvent.pointerMove(trenner, { clientX: 1400, pointerId: 1 })
    fireEvent.pointerUp(trenner, { clientX: 1400, pointerId: 1 })

    expect(aendern).toHaveBeenCalledExactlyOnceWith(120)
  })

  it('schreibt nichts, wenn der Zug dort endet, wo er begann', () => {
    const { aendern, vorschau } = zeige()
    const trenner = screen.getByRole('separator')

    fireEvent.pointerDown(trenner, { clientX: 300, pointerId: 1, button: 0 })
    fireEvent.pointerUp(trenner, { clientX: 300, pointerId: 1 })

    expect(aendern).not.toHaveBeenCalled()
    expect(vorschau).toHaveBeenLastCalledWith(undefined)
  })

  it('beachtet weder die rechte Maustaste noch eine Bewegung ohne Zug', () => {
    const { aendern, vorschau } = zeige()
    const trenner = screen.getByRole('separator')

    fireEvent.pointerMove(trenner, { clientX: 350, pointerId: 1 })
    fireEvent.pointerDown(trenner, { clientX: 300, pointerId: 1, button: 2 })
    fireEvent.pointerMove(trenner, { clientX: 350, pointerId: 1 })
    fireEvent.pointerUp(trenner, { clientX: 350, pointerId: 1 })

    expect(vorschau).not.toHaveBeenCalled()
    expect(aendern).not.toHaveBeenCalled()
  })

  // **Ein abgebrochener Zug** (das System nimmt den Zeiger weg) schreibt
  // nichts und lässt die Spalte, wie sie war.
  it('verwirft einen abgebrochenen Zug', () => {
    const { aendern, vorschau } = zeige()
    const trenner = screen.getByRole('separator')

    fireEvent.pointerDown(trenner, { clientX: 300, pointerId: 1, button: 0 })
    fireEvent.pointerMove(trenner, { clientX: 400, pointerId: 1 })
    fireEvent.pointerCancel(trenner, { pointerId: 1 })

    expect(aendern).not.toHaveBeenCalled()
    expect(vorschau).toHaveBeenLastCalledWith(undefined)
  })

  it('lässt andere Tasten durch', async () => {
    const { aendern } = zeige()
    screen.getByRole('separator').focus()

    await userEvent.keyboard('{ArrowDown}a')

    expect(aendern).not.toHaveBeenCalled()
  })
})
