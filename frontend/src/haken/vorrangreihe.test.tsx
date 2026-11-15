// **Die Reihe des Vorrangs, einmal geladen und nicht ewig** — #2054. Was
// `kifeldvorrang.test.tsx` am Feld zeigt, hier am Haken selbst: die Frist,
// ein Fehler und ein Abbau, bevor die Antwort da ist.
import { act, renderHook, waitFor } from '@testing-library/react'
import { afterEach, expect, it, vi } from 'vitest'

import { useEinstellungen } from './useEinstellungen'
import { useVorrangreihe, vergissVorrangreihe } from './useVorrangreihe'
import { belegeApi, raeumeApi } from '../test/api'
import { effekteNachholen } from '../test/effekte'

import type { KIKatalogantwort } from '../antwortenmarkt'

afterEach(() => {
  raeumeApi()
  vergissVorrangreihe()
  vi.useRealTimers()
})

const katalog = (vorrang: string[]): KIKatalogantwort => ({
  nurLokal: false,
  einrichtung: 'bereit',
  anbieter: [],
  cloud: { stand: 'ungewiss', beleg: '' },
  befunde: [],
  modelle: [],
  vorrang,
  vorrangbefunde: [],
  fehler: null,
})

// beruehrt baut den Haken und berührt die Zeile einmal.
function beruehrt() {
  const lauf = renderHook(() => useVorrangreihe())
  act(() => {
    lauf.result.current.beruehrt()
  })

  return lauf
}

it('fragt nicht, solange niemand fragt', async () => {
  const gefragt = vi.fn(() => Promise.resolve(katalog(['ollama/a'])))
  belegeApi({ KIModelle: gefragt })

  const { result } = renderHook(() => useVorrangreihe())
  await effekteNachholen()

  expect(gefragt).not.toHaveBeenCalled()
  expect(result.current.reihe).toBeUndefined()
})

it('fragt nach einer Minute neu', async () => {
  vi.useFakeTimers({ toFake: ['Date'] })
  const gefragt = vi
    .fn()
    .mockResolvedValueOnce(katalog(['ollama/a']))
    .mockResolvedValueOnce(katalog(['ollama/b']))
  belegeApi({ KIModelle: gefragt })

  const erster = beruehrt()
  await waitFor(() => {
    expect(erster.result.current.reihe).toEqual(['ollama/a'])
  })

  // **Derselbe Haken, wieder berührt** (Z5) — vorher fragte erst ein neuer.
  vi.setSystemTime(Date.now() + 61_000)
  act(() => {
    erster.result.current.beruehrt()
  })
  await waitFor(() => {
    expect(erster.result.current.reihe).toEqual(['ollama/b'])
  })
  expect(gefragt).toHaveBeenCalledTimes(2)
})

it('merkt sich keinen Fehler und zeigt nichts', async () => {
  const gefragt = vi
    .fn<() => Promise<KIKatalogantwort>>()
    .mockRejectedValueOnce(new Error('weg'))
    .mockResolvedValueOnce(katalog(['ollama/a']))
  belegeApi({ KIModelle: gefragt })

  const erster = beruehrt()
  await waitFor(() => {
    expect(gefragt).toHaveBeenCalledOnce()
  })
  await expect(gefragt.mock.results[0]?.value).rejects.toThrow('weg')
  await effekteNachholen()
  expect(erster.result.current.reihe).toBeUndefined()

  const zweiter = beruehrt()
  await waitFor(() => {
    expect(zweiter.result.current.reihe).toEqual(['ollama/a'])
  })
})

it('trägt nichts mehr ein, wenn der Haken vor der Antwort abgebaut ist', async () => {
  let antworten: (wert: KIKatalogantwort) => void = () => undefined
  const gefragt = vi.fn(
    () =>
      new Promise<KIKatalogantwort>((los) => {
        antworten = los
      }),
  )
  belegeApi({ KIModelle: gefragt })

  const { result, unmount } = beruehrt()
  await waitFor(() => {
    expect(gefragt).toHaveBeenCalledOnce()
  })
  unmount()
  antworten(katalog(['ollama/a']))
  await effekteNachholen()

  expect(result.current.reihe).toBeUndefined()
})

// **Eine KI-Einstellung verwirft die gemerkte Reihe** — wer das Modell des
// Vorrangs wählt, soll es am Formular sofort sehen und nicht erst nach einer
// Minute. Eine andere Einstellung lässt sie stehen.
it('fragt nach einer geänderten KI-Einstellung neu, nach einer anderen nicht', async () => {
  const gefragt = vi.fn(() => Promise.resolve(katalog(['ollama/a'])))
  const geschrieben: (() => void)[] = []
  belegeApi({
    KIModelle: gefragt,
    Einstellungen: () => Promise.resolve({ einstellungen: [] }),
    EinstellungSetzen: () =>
      new Promise((los) => {
        geschrieben.push(() => {
          los({})
        })
      }),
  })

  const einstellungen = renderHook(() => useEinstellungen())

  // schreibe setzt einen Wert und lässt das Schreiben erst dann antworten,
  // wenn es angekommen ist — gezählt wird danach, nicht davor.
  async function schreibe(schluessel: string) {
    const vorher = geschrieben.length
    einstellungen.result.current.setze(schluessel, 'x')
    await waitFor(() => {
      expect(geschrieben).toHaveLength(vorher + 1)
    })
    geschrieben[vorher]?.()
    await effekteNachholen()
  }

  async function reihe() {
    const gelesen = beruehrt()
    await waitFor(() => {
      expect(gelesen.result.current.reihe).toEqual(['ollama/a'])
    })
  }

  await reihe()
  await schreibe('schalter.netz')
  await reihe()
  expect(gefragt).toHaveBeenCalledOnce()

  await schreibe('ki.vorrangmodell')
  await reihe()
  expect(gefragt).toHaveBeenCalledTimes(2)
})

// **Ein Fehler nach einem Erfolg zeigt nichts** — N5 aus dem Nachlauf über
// Phase 6.11: die Reihe von vorhin blieb stehen, obwohl sie nicht mehr galt.
it('zeigt nach einer gescheiterten Neuabfrage nicht die alte Reihe', async () => {
  const gefragt = vi
    .fn<() => Promise<KIKatalogantwort>>()
    .mockResolvedValueOnce(katalog(['ollama/a']))
    .mockRejectedValueOnce(new Error('weg'))
  belegeApi({ KIModelle: gefragt })

  const lauf = beruehrt()
  await waitFor(() => {
    expect(lauf.result.current.reihe).toEqual(['ollama/a'])
  })

  act(() => {
    vergissVorrangreihe()
  })

  await waitFor(() => {
    expect(gefragt).toHaveBeenCalledTimes(2)
  })
  await expect(gefragt.mock.results[1]?.value).rejects.toThrow('weg')
  await effekteNachholen()

  expect(lauf.result.current.reihe).toBeUndefined()
  expect(lauf.result.current.befunde).toEqual([])
})

it('trägt auch einen Fehler nicht ein, der nach dem Abbau kommt', async () => {
  let scheitern: (grund: Error) => void = () => undefined
  const gefragt = vi.fn(
    () =>
      new Promise<KIKatalogantwort>((_, nein) => {
        scheitern = nein
      }),
  )
  belegeApi({ KIModelle: gefragt })

  const { result, unmount } = beruehrt()
  await waitFor(() => {
    expect(gefragt).toHaveBeenCalledOnce()
  })
  unmount()
  scheitern(new Error('weg'))
  await effekteNachholen()

  expect(result.current.reihe).toBeUndefined()
})
