// **Ein Lesefehler setzt die Einstellungen nicht still zurück** — U08
// (#2124).
//
// Schlug das Nachlesen fehl — etwa weil ein anderes Programm den Bestand
// gerade sperrte —, räumte der Haken alle Einträge ab: die Sprache fiel auf
// die des Systems zurück, ein abgeschalteter Bereich erschien wieder. Bei
// **demselben** Bestand und Profil bleibt jetzt der bestätigte Stand, und der
// Fehler steht daneben. **Ein echter Wechsel** — anderes Profil, ersetzter
// Bestand — bekommt keine fremden Einstellungen.
import { renderHook, waitFor } from '@testing-library/react'
import { afterEach, expect, it } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'
import { useEinstellungen } from './useEinstellungen'

afterEach(() => {
  raeumeApi()
})

const sprache = {
  schluessel: 'sprache',
  wert: 'en',
  wirksam: 'en',
  werte: ['erben', 'de', 'en'],
  bereich: 'anwendung',
}

function mitFehlerAbDemZweitenLesen() {
  let gelesen = 0

  belegeApi({
    Einstellungen: () => {
      gelesen += 1

      return gelesen === 1
        ? Promise.resolve({ einstellungen: [sprache] })
        : Promise.reject(new Error('der Bestand ist gesperrt'))
    },
  })
}

interface Eigenschaften {
  stand: number
  bezug: string
}

const useProbe = ({ stand, bezug }: Eigenschaften) =>
  useEinstellungen(stand, 'fragen', bezug)

it('behält bei demselben Bezug den bestätigten Stand und sagt den Fehler', async () => {
  mitFehlerAbDemZweitenLesen()

  const { result, rerender } = renderHook(useProbe, {
    initialProps: { stand: 0, bezug: 'bestand-1/profil-1' },
  })

  await waitFor(() => {
    expect(result.current.wert('sprache')).toBe('en')
  })

  rerender({ stand: 1, bezug: 'bestand-1/profil-1' })

  await waitFor(() => {
    expect(result.current.ursache).toBeDefined()
  })

  expect(result.current.wert('sprache')).toBe('en')
})

it('räumt bei einem echten Wechsel, auch wenn das Lesen scheitert', async () => {
  mitFehlerAbDemZweitenLesen()

  const { result, rerender } = renderHook(useProbe, {
    initialProps: { stand: 0, bezug: 'bestand-1/profil-1' },
  })

  await waitFor(() => {
    expect(result.current.wert('sprache')).toBe('en')
  })

  rerender({ stand: 1, bezug: 'bestand-1/profil-2' })

  await waitFor(() => {
    expect(result.current.ursache).toBeDefined()
  })

  expect(result.current.wert('sprache')).toBe('')
})
