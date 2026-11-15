// **Die Naht zwischen Plan und Grafikkarte** — Werkstück G3 von Phase 5.5
// (#1460).
//
// **Hier läuft kein PixiJS.** Geprüft wird die Entscheidung davor — ob
// gezeichnet werden kann und was geschieht, wenn nicht. Die Naht selbst ist
// ein Feld und im Test ein Doppel.
import { afterEach, describe, expect, it, vi } from 'vitest'

import { grafikprobe, tischmalerei } from './maler'

import type { Tischmaler } from './maler'
import { einMaler } from '../test/malerprobe'

// einKontext ist ein WebGL-Kontext, der übersetzt — oder nicht.
const einKontext = ({ uebersetzt = true, verloren = false } = {}) => {
  const freigeben = vi.fn()
  const uebersetzen = vi.fn()

  return {
    freigeben,
    uebersetzen,
    gl: {
      FRAGMENT_SHADER: 0x8b30,
      COMPILE_STATUS: 0x8b81,
      isContextLost: () => verloren,
      createShader: () => ({}),
      shaderSource: () => undefined,
      compileShader: uebersetzen,
      // **Ein verlorener Kontext antwortet `null`**, nicht `false` — gemessen
      // am nativen Fenster am 2026-10-06 (UI-N01, #2089).
      getShaderParameter: () => (uebersetzt ? true : null),
      deleteShader: () => undefined,
      getExtension: (name: string) =>
        name === 'WEBGL_lose_context' ? { loseContext: freigeben } : null,
    },
  }
}

const dokument = (getContext: (art: string) => unknown) =>
  ({ createElement: () => ({ getContext }) }) as unknown as Document

// einOrt ist der Behälter, den React hält — im Dokument.
const einOrt = () => {
  const ort = document.createElement('div')
  document.body.append(ort)

  return ort
}

// mitWebGL lässt jede Leinwand des Dokuments diesen Kontext geben.
const mitWebGL = (kontext = einKontext()) => {
  vi.spyOn(HTMLCanvasElement.prototype, 'getContext').mockReturnValue(
    kontext.gl as never,
  )

  return kontext
}

const maler: Tischmaler = einMaler()

describe('grafikprobe', () => {
  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('sagt »kein-webgl«, wo es keines gibt', () => {
    // jsdom bringt kein WebGL mit — und genau das ist der Fall, für den es
    // den Rückfall gibt.
    expect(grafikprobe(document)).toBe('kein-webgl')
  })

  it('nimmt webgl2, wo es da ist, und gibt den Probekontext wieder frei', () => {
    const kontext = einKontext()

    expect(
      grafikprobe(dokument((art) => (art === 'webgl2' ? kontext.gl : null))),
    ).toBeUndefined()
    // **Sonst sammeln sich Kontexte** — WebKit hält nur eine begrenzte Zahl
    // offen und nimmt dann dem ältesten seinen, und das ist der des Tischs.
    expect(kontext.freigeben).toHaveBeenCalledTimes(1)
  })

  it('fällt auf webgl zurück, wo nur das da ist', () => {
    const kontext = einKontext()

    expect(
      grafikprobe(dokument((art) => (art === 'webgl' ? kontext.gl : null))),
    ).toBeUndefined()
  })

  it('sagt »kein-webgl«, wenn das Anlegen des Kontexts wirft', () => {
    expect(
      grafikprobe(
        dokument(() => {
          throw new Error('kein Kontext')
        }),
      ),
    ).toBe('kein-webgl')
  })

  it('ein Kontext, der nicht übersetzt, endet nach einem Versuch als »kein-renderer«', () => {
    // **Die begrenzte Arbeit** — UI-N01 (#2089): PixiJS 8.20.1 halbiert bei
    // jedem Fehlschlag bis null und versucht es dann ohne Ende weiter
    // (`checkMaxIfStatementsInShader`). Übersetzt schon der Shader ohne ein
    // einziges `if` nicht, kommt PixiJS gar nicht erst dran.
    const kontext = einKontext({ uebersetzt: false })

    expect(grafikprobe(dokument(() => kontext.gl))).toBe('kein-renderer')
    expect(kontext.uebersetzen).toHaveBeenCalledTimes(1)
    expect(kontext.freigeben).toHaveBeenCalledTimes(1)
  })

  it('ein Kontext ohne Shader und ohne Freigabe ist »kein-renderer«', () => {
    const kontext = einKontext()
    const gl = {
      ...kontext.gl,
      createShader: () => null,
      getExtension: () => null,
    }

    expect(grafikprobe(dokument(() => gl))).toBe('kein-renderer')
  })

  it('ein verlorener Kontext ist »kein-renderer« und wird nicht übersetzt', () => {
    const kontext = einKontext({ verloren: true })

    expect(grafikprobe(dokument(() => kontext.gl))).toBe('kein-renderer')
    expect(kontext.uebersetzen).not.toHaveBeenCalled()
  })
})

describe('tischmalerei', () => {
  afterEach(() => {
    vi.restoreAllMocks()
    document.body.replaceChildren()
  })

  it('sagt ohne WebGL, woran es liegt, statt leer zu bleiben', async () => {
    const naht = vi.fn()
    const ort = einOrt()

    expect(await tischmalerei(ort, naht)).toEqual({
      art: 'ohne',
      grund: 'kein-webgl',
    })
    expect(naht).not.toHaveBeenCalled()
    expect(ort.children).toHaveLength(0)
  })

  it('gibt den Maler her, wo er hochkommt', async () => {
    mitWebGL()
    const malerei = await tischmalerei(einOrt(), () => Promise.resolve(maler))

    expect(malerei.art).toBe('da')
  })

  it('jeder Anlauf legt eine eigene Leinwand in den Ort, und das Abräumen nimmt sie heraus', async () => {
    // **UI-N01** (#2089): bis zum 2026-10-06 bekam PixiJS die Leinwand, die
    // React hält. Sein Abräumen nahm sie aus dem Dokument und verlor ihren
    // Kontext; der nächste Anlauf bekam sie wieder — mit dem verlorenen
    // Kontext —, und die Webansicht stand.
    mitWebGL()
    const ort = einOrt()
    const innen = einMaler()
    const naht = vi.fn<(leinwand: HTMLCanvasElement) => Promise<typeof innen>>(
      () => Promise.resolve(innen),
    )

    const erste = await tischmalerei(ort, naht)
    const eins = naht.mock.calls[0]?.[0]
    expect(eins?.parentElement).toBe(ort)
    if (erste.art !== 'da') throw new Error('kein Maler')

    erste.maler.raeume()
    expect(innen.raeume).toHaveBeenCalledTimes(1)
    expect(ort.children).toHaveLength(0)
    expect(ort.isConnected).toBe(true)

    await tischmalerei(ort, naht)
    const zwei = naht.mock.calls[1]?.[0]
    expect(zwei).not.toBe(eins)
    expect(zwei?.parentElement).toBe(ort)
  })

  it('ein gescheiterter Renderer lässt keine Leinwand zurück', async () => {
    mitWebGL()
    const ort = einOrt()
    const malerei = await tischmalerei(ort, () =>
      Promise.reject(new Error('WebGL unsupported')),
    )

    expect(malerei).toEqual({ art: 'ohne', grund: 'kein-renderer' })
    expect(ort.children).toHaveLength(0)
  })

  it('ein Kontext, der nicht übersetzt, ruft die Naht gar nicht erst', async () => {
    mitWebGL(einKontext({ uebersetzt: false }))
    const naht = vi.fn()
    const ort = einOrt()

    expect(await tischmalerei(ort, naht)).toEqual({
      art: 'ohne',
      grund: 'kein-renderer',
    })
    expect(naht).not.toHaveBeenCalled()
    expect(ort.children).toHaveLength(0)
  })
})
