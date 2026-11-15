// **Ein Schreiben gehört zu dem Profil, in dem es begann** — U08
// (#2124),
// Zweitblick über C und D, Befund 4.
//
// Der Abschluss eines Schreibens holte sich seine Laufnummer erst **nach**
// der Antwort. Lag dazwischen ein Profilwechsel, galt er im neuen Profil —
// und scheiterte das Nachlesen dort, trug er den Wert aus A in die Anzeige
// von B ein.
import { act, cleanup, renderHook, waitFor } from '@testing-library/react'
import { afterEach, expect, it } from 'vitest'

import type { Umschlag } from '../api'
import { belegeApi, raeumeApi } from '../test/api'
import { effekteNachholen } from '../test/effekte'
import { useEinstellungen } from './useEinstellungen'

afterEach(() => {
  cleanup()
  raeumeApi()
})

const sprache = (wert: string) => ({
  schluessel: 'sprache',
  wert,
  wirksam: wert,
  werte: ['erben', 'de', 'en', 'fr'],
  bereich: 'anwendung',
})

it('trägt einen alten Schreibabschluss nicht in das neue Profil ein', async () => {
  let gelesen = 0
  let fertig: ((antwort: Umschlag) => void) | undefined
  let geschrieben = 0

  belegeApi({
    // A liest `en`, B liest `de`; danach scheitert jedes Lesen.
    Einstellungen: () => {
      gelesen++

      return gelesen <= 2
        ? Promise.resolve({
            einstellungen: [sprache(gelesen === 1 ? 'en' : 'de')],
          })
        : Promise.reject(new Error('planmäßig: das Lesen scheitert'))
    },
    EinstellungSetzen: () => {
      geschrieben++

      return geschrieben === 1
        ? new Promise<Umschlag>((erfuellen) => {
            fertig = erfuellen
          })
        : Promise.resolve({ fehler: null })
    },
  })

  const { result, rerender } = renderHook(
    ({ bezug }) => useEinstellungen(0, 'fragen', bezug),
    { initialProps: { bezug: 'profil-A' } },
  )
  await waitFor(() => {
    expect(result.current.wert('sprache')).toBe('en')
  })

  act(() => {
    result.current.setze('sprache', 'fr')
  })
  await waitFor(() => {
    expect(fertig).toBeDefined()
  })

  rerender({ bezug: 'profil-B' })
  await waitFor(() => {
    expect(result.current.wert('sprache')).toBe('de')
  })

  act(() => {
    fertig?.({ fehler: null })
  })
  await effekteNachholen()

  // **Jemand hört zu**: ein Schreiben in B kommt an — sein Nachlesen
  // scheitert, und der angenommene Wert gilt. Stünde `fr` aus A unter den
  // angenommenen, stünde es jetzt ebenso da.
  act(() => {
    result.current.setze('anderes', 'x')
  })
  await waitFor(() => {
    expect(result.current.ursache).toBeDefined()
  })

  expect(result.current.wert('sprache')).toBe('de')
})

it('zeigt den Fehlschlag eines alten Schreibens nicht im neuen Profil', async () => {
  let ablehnen: ((ursache: Error) => void) | undefined
  let gelesen = 0

  belegeApi({
    Einstellungen: () => {
      gelesen++

      return Promise.resolve({ einstellungen: [sprache('de')] })
    },
    EinstellungSetzen: () =>
      new Promise<Umschlag>((_, nein) => {
        ablehnen = nein
      }),
  })

  const { result, rerender } = renderHook(
    ({ bezug, stand }) => useEinstellungen(stand, 'fragen', bezug),
    { initialProps: { bezug: 'profil-A', stand: 0 } },
  )
  await waitFor(() => {
    expect(result.current.geladen).toBe(true)
  })

  act(() => {
    result.current.setze('sprache', 'fr')
  })
  await waitFor(() => {
    expect(ablehnen).toBeDefined()
  })

  rerender({ bezug: 'profil-B', stand: 0 })
  await waitFor(() => {
    expect(gelesen).toBe(2)
  })

  act(() => {
    ablehnen?.(new Error('planmäßig: das alte Schreiben scheitert'))
  })
  await effekteNachholen()

  // **Jemand hört zu**: ein neuer Stand in B wird gelesen — danach ist auch
  // die Ablehnung längst verarbeitet.
  rerender({ bezug: 'profil-B', stand: 1 })
  await waitFor(() => {
    expect(gelesen).toBe(3)
  })

  expect(result.current.fehlschlag).toBeUndefined()
})
