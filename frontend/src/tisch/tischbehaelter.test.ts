// **Die Fläche folgt ihrem Behälter und nicht dem Fenster** — gefunden am
// 2026-09-24 mit der WebKit-Probe
// (#1828),
// beim Tisch nach dem Artboard (Befund F09).
//
// # Wogegen hier geprüft wird
//
// **Gegen einen Tisch, der nach einer Größenänderung abgeschnitten stand.**
// PixiJS passt die Leinwand erst im nächsten Bild an (`queueResize` in
// pixi.js 8.20.1, `lib/app/ResizePlugin.mjs`); `useTisch` las die Fläche
// im selben Augenblick nach — und bekam die alte. Und wenn die Seitenleiste
// auf- oder zuging, änderte sich der Behälter, ohne dass das Fenster
// `resize` meldete: dann wurde gar nicht gemessen.
//
// **Die Probe verhält sich wie PixiJS**: ihr Maler kennt die neue Größe erst,
// wenn er angepasst wird. Eine Vorrichtung, die sofort Bescheid wüsste,
// prüfte sich selbst (AGENTS.md §3).
import { act, renderHook, waitFor } from '@testing-library/react'
import { afterEach, beforeEach, expect, it, vi } from 'vitest'

import { einePartie } from '../seiten/partieprobe'
import { einMaler } from '../test/malerprobe'
import { tischplan } from './tischplan'
import { useTisch } from './useTisch'

import type { Tischmalerei } from './maler'

const plan = tischplan(einePartie())

// beobachter sind die Rückrufe jedes `ResizeObserver`, den der Lauf anlegt,
// samt dem Knoten, den er beobachtet.
let beobachter: { melde: () => void; knoten: Element[] }[] = []

beforeEach(() => {
  beobachter = []
  vi.stubGlobal(
    'ResizeObserver',
    class {
      private readonly eintrag: { melde: () => void; knoten: Element[] }

      constructor(melde: () => void) {
        this.eintrag = { melde, knoten: [] }
        beobachter.push(this.eintrag)
      }

      observe(knoten: Element) {
        this.eintrag.knoten.push(knoten)
      }

      disconnect() {
        beobachter = beobachter.filter((einer) => einer !== this.eintrag)
      }
    },
  )
})

afterEach(() => {
  vi.unstubAllGlobals()
})

// leinwandImBehaelter gibt eine Leinwand, die in einem Behälter steht —
// dem Knoten, dem PixiJS über `resizeTo` folgt.
function leinwandImBehaelter() {
  const behaelter = document.createElement('div')
  const leinwand = document.createElement('canvas')
  behaelter.append(leinwand)

  return { behaelter, leinwand }
}

function haenge(
  leinwand: HTMLCanvasElement,
  malerei: () => Promise<Tischmalerei>,
) {
  return renderHook(() =>
    useTisch({
      plan,
      bild: () => undefined,
      malerei,
      leinwand: { current: leinwand },
    }),
  )
}

it('passt die Fläche an, wenn ihr Behälter sich ändert', async () => {
  const maler = einMaler({ breite: 800, hoehe: 600 })
  const { behaelter, leinwand } = leinwandImBehaelter()
  const lauf = haenge(leinwand, () => Promise.resolve({ art: 'da', maler }))

  await waitFor(() => {
    expect(lauf.result.current.masse.breite).toBe(800)
  })

  // **Beobachtet wird der Behälter** — die Leinwand selbst trägt die Maße,
  // die PixiJS ihr gibt, und änderte sich erst nach dem Anpassen.
  expect(beobachter.flatMap((einer) => einer.knoten)).toContain(behaelter)

  maler.behaelterWird({ breite: 1000, hoehe: 700 })
  act(() => {
    for (const einer of beobachter) einer.melde()
  })

  expect(lauf.result.current.masse).toEqual({ breite: 1000, hoehe: 700 })
  expect(maler.passeAn).toHaveBeenCalled()

  lauf.unmount()
  expect(beobachter).toEqual([])
})

// **Ohne Maler gibt es nichts anzupassen** — die Leinwand steht, aber
// gezeichnet wird nicht, und die Fläche bleibt, was sie war.
it('passt nichts an, solange kein Maler da ist', async () => {
  const { leinwand } = leinwandImBehaelter()
  const lauf = haenge(leinwand, () =>
    Promise.resolve({ art: 'ohne', grund: 'kein-webgl' }),
  )

  await waitFor(() => {
    expect(lauf.result.current.hindernis).toBe('kein-webgl')
  })

  act(() => {
    for (const einer of beobachter) einer.melde()
  })

  expect(lauf.result.current.masse).toEqual({ breite: 0, hoehe: 0 })
  lauf.unmount()
})
