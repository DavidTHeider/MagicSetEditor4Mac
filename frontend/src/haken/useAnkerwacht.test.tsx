// **Ein offenes Menü folgt seinem Anker, auch ohne Rollen** — Rest R2 des
// Nachlaufs über Phase 6 von 0.9.0: die Deckliste sortiert „zuletzt
// benutzt“ und lädt nach dem Öffnen nach, und die Zeile rückte unter einem
// offenen Menü weg, ohne dass jemand rollte.
import { render, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { useAnkerwacht } from './useAnkerwacht'

function Wache({
  aktiv,
  anker,
  folgen,
}: {
  aktiv: boolean
  anker: Element | null
  folgen: () => void
}) {
  useAnkerwacht(aktiv, () => anker, folgen)

  return null
}

// bild wartet ein Bild ab — die Wache schaut je Bild einmal.
const bild = () =>
  new Promise<void>((fertig) => {
    requestAnimationFrame(() => {
      fertig()
    })
  })

afterEach(() => {
  document.body.replaceChildren()
})

describe('useAnkerwacht', () => {
  it('meldet, wenn der Anker wandert, und schweigt, solange er steht', async () => {
    const anker = document.createElement('div')
    document.body.append(anker)
    let oben = 100
    anker.getBoundingClientRect = () => new DOMRect(20, oben, 200, 40)
    const folgen = vi.fn()

    render(<Wache aktiv anker={anker} folgen={folgen} />)
    await bild()
    await bild()
    expect(folgen).not.toHaveBeenCalled()

    oben = 160

    await waitFor(() => {
      expect(folgen).toHaveBeenCalledTimes(1)
    })
    await bild()
    expect(folgen).toHaveBeenCalledTimes(1)
  })

  it('meldet, wenn das Fenster seine Größe ändert', async () => {
    const anker = document.createElement('div')
    document.body.append(anker)
    anker.getBoundingClientRect = () => new DOMRect(20, 100, 200, 40)
    const folgen = vi.fn()
    const breite = window.innerWidth

    render(<Wache aktiv anker={anker} folgen={folgen} />)
    await bild()
    vi.spyOn(window, 'innerWidth', 'get').mockReturnValue(breite - 100)

    await waitFor(() => {
      expect(folgen).toHaveBeenCalledTimes(1)
    })
    vi.restoreAllMocks()
  })

  it('meldet, wenn der Anker den Baum verlässt', async () => {
    const anker = document.createElement('div')
    document.body.append(anker)
    anker.getBoundingClientRect = () => new DOMRect(20, 100, 200, 40)
    const folgen = vi.fn()

    render(<Wache aktiv anker={anker} folgen={folgen} />)
    await bild()
    anker.remove()

    await waitFor(() => {
      expect(folgen).toHaveBeenCalledTimes(1)
    })
  })

  it('schaut nicht, solange nichts offen ist, und hört beim Schließen auf', async () => {
    const anker = document.createElement('div')
    document.body.append(anker)
    let oben = 100
    anker.getBoundingClientRect = () => new DOMRect(20, oben, 200, 40)
    const folgen = vi.fn()

    const { rerender } = render(
      <Wache aktiv={false} anker={anker} folgen={folgen} />,
    )
    oben = 160
    await bild()
    await bild()
    expect(folgen).not.toHaveBeenCalled()

    // **Ein Beleg, dass jemand zuhört** (§3): offen meldet derselbe Weg.
    rerender(<Wache aktiv anker={anker} folgen={folgen} />)
    await bild()
    oben = 200
    await waitFor(() => {
      expect(folgen).toHaveBeenCalledTimes(1)
    })

    rerender(<Wache aktiv={false} anker={anker} folgen={folgen} />)
    oben = 240
    await bild()
    await bild()
    expect(folgen).toHaveBeenCalledTimes(1)
  })
})
