// **Ohne gemerktes Feld zählt die Stelle** — ENG-02
// (#1727).
//
// Unter Wails 3 nimmt eine eigene Fläche über der Webansicht die Dateien aus
// dem Finder entgegen, und die Webansicht sieht kein `dragover` mehr — das
// gemerkte Feld bleibt leer. Welches Feld gemeint ist, ergibt sich dann aus
// der Stelle, die Go mit dem Ablegen meldet, und aus der Marke, die die
// Vorschau jedem Bildfeld gibt (`data-bildfeld`).
//
// **Was hier nicht geprüft wird**: dass Wails die Stelle in denselben
// Einheiten meldet, in denen `elementFromPoint` sucht — CSS-Pixel relativ
// zur Webansicht (`HandlePlatformFileDrop`, `window.ts`, v3.0.0-beta.23).
// Das zeigt nur ein Ablegen am Fenster (`docs/WAILS3.md`).
import { cleanup, renderHook } from '@testing-library/react'
import { type ReactNode } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Ablageraum, useAblage } from '../ablage'
import { bildfeldAn, useAbgelegteDateien } from './useAbgelegteDateien'

interface MitLaufzeit {
  runtime?: {
    OnFileDrop?: (
      rueckruf: (x: number, y: number, pfade: string[]) => void,
      ueberZiel: boolean,
    ) => void
    OnFileDropOff?: () => void
  }
}

afterEach(() => {
  cleanup()
  delete (globalThis as MitLaufzeit).runtime
  document.body.replaceChildren()
  vi.restoreAllMocks()
})

// bildflaeche legt eine Fläche mit der Marke ins Dokument — und ein Kind
// darin, denn getroffen wird das innerste Element.
function bildflaeche(feld: string): HTMLElement {
  const flaeche = document.createElement('button')
  flaeche.setAttribute('data-bildfeld', feld)
  const inneres = document.createElement('span')
  flaeche.append(inneres)
  document.body.append(flaeche)

  return inneres
}

describe('bildfeldAn', () => {
  it('findet das Feld über dem getroffenen Element', () => {
    const getroffen = bildflaeche('kunst')

    expect(bildfeldAn(10, 20, () => getroffen)).toBe('kunst')
  })

  it('findet keines, wo kein Bildfeld liegt — und keines, wo nichts liegt', () => {
    const daneben = document.createElement('div')
    document.body.append(daneben)

    expect(bildfeldAn(1, 2, () => daneben)).toBeUndefined()
    expect(bildfeldAn(1, 2, () => null)).toBeUndefined()
  })

  it('fragt im Betrieb das Dokument — und jsdom kennt die Suche nicht', () => {
    expect(bildfeldAn(1, 2)).toBeUndefined()

    const getroffen = bildflaeche('kunst')
    const suche = vi.fn(() => getroffen)
    Object.defineProperty(document, 'elementFromPoint', {
      configurable: true,
      value: suche,
    })

    try {
      expect(bildfeldAn(30, 40)).toBe('kunst')
      expect(suche).toHaveBeenCalledWith(30, 40)
    } finally {
      Reflect.deleteProperty(document, 'elementFromPoint')
    }
  })
})

describe('das Ablegen ohne gemerktes Feld', () => {
  it('setzt das Bild in das Feld an der Stelle', () => {
    let ablegen: ((x: number, y: number, pfade: string[]) => void) | undefined
    ;(globalThis as MitLaufzeit).runtime = {
      OnFileDrop: (rueckruf) => {
        ablegen = rueckruf
      },
      OnFileDropOff: vi.fn(),
    }

    const getroffen = bildflaeche('kunst')
    Object.defineProperty(document, 'elementFromPoint', {
      configurable: true,
      value: (x: number, y: number) =>
        x === 30 && y === 40 ? getroffen : null,
    })

    const einsetzen = vi.fn()

    try {
      renderHook(
        () => {
          const ablage = useAblage()

          useAbgelegteDateien(vi.fn())
          ablage.setzeBildweg(einsetzen)
        },
        {
          wrapper: ({ children }: { children: ReactNode }) => (
            <Ablageraum>{children}</Ablageraum>
          ),
        },
      )

      ablegen?.(5, 5, ['/Users/david/daneben.png'])
      expect(einsetzen).not.toHaveBeenCalled()

      ablegen?.(30, 40, ['/Users/david/urlaub.jpg'])
      expect(einsetzen).toHaveBeenCalledWith('kunst', '/Users/david/urlaub.jpg')
    } finally {
      Reflect.deleteProperty(document, 'elementFromPoint')
    }
  })
})
