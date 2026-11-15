// **Die Vorschau bleibt stehen, bis das neue Bild da ist** — U12
// (#2126).
//
// Maßstab und getippter Text waren Teil der Ladefrage, und eine neue Frage
// räumt: beim Zoomen verschwand die Karte, bis neu gerastert war, und
// zwischen zwei Tastendrücken sprang sie auf den gespeicherten Stand zurück.
// **Eine andere Karte** räumt weiter — ihr Bild ist ein anderes.
import { renderHook, waitFor } from '@testing-library/react'
import { afterEach, expect, it } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'
import { useKartenvorschau } from './useKartenvorschau'

import type { VorschauDTO } from '../formen'

afterEach(() => {
  raeumeApi()
})

const bild = (name: string): VorschauDTO => ({
  stil: '',
  gezeichnet: '',
  bild: name,
  breite: 10,
  hoehe: 14,
  dpi: 0,
  befunde: [],
  flaechen: [],
  vorgaben: [],
  nennungen: [],
})

interface Eigenschaften {
  karte: string
  getippt: { feld: string; wert: string } | undefined
  faktor: number
}

const useProbe = ({ karte, getippt, faktor }: Eigenschaften) =>
  useKartenvorschau(karte, getippt, faktor)

// haengend lässt jeden Ruf ab dem `ab`-ten hängen — das neue Bild kommt nie.
function haengend(ab: number) {
  let vorschauen = 0
  let proben = 0

  belegeApi({
    Vorschau: (karte: string, faktor: number) => {
      vorschauen += 1

      return vorschauen >= ab
        ? new Promise(() => undefined)
        : Promise.resolve({
            vorschau: bild(`${karte}@${String(faktor)}`),
            fehler: null,
          })
    },
    Probevorschau: (karte: string, _feld: string, wert: string) => {
      proben += 1

      return proben >= ab
        ? new Promise(() => undefined)
        : Promise.resolve({ vorschau: bild(`${karte}:${wert}`), fehler: null })
    },
  })
}

it('behält beim Zoomen das Bild derselben Karte, bis das neue gerastert ist', async () => {
  haengend(2)

  const { result, rerender } = renderHook(useProbe, {
    initialProps: { karte: 'k1', getippt: undefined, faktor: 1 },
  })

  await waitFor(() => {
    expect(result.current.gezeigt?.bild).toBe('k1@1')
  })

  rerender({ karte: 'k1', getippt: undefined, faktor: 2 })

  expect(result.current.gezeigt?.bild).toBe('k1@1')
})

it('springt zwischen zwei Tastendrücken nicht auf den gespeicherten Stand zurück', async () => {
  haengend(2)

  const { result, rerender } = renderHook(useProbe, {
    initialProps: {
      karte: 'k1',
      getippt: { feld: 'name', wert: 'A' },
      faktor: 1,
    },
  })

  await waitFor(() => {
    expect(result.current.gezeigt?.bild).toBe('k1:A')
  })

  rerender({ karte: 'k1', getippt: { feld: 'name', wert: 'AB' }, faktor: 1 })

  expect(result.current.gezeigt?.bild).toBe('k1:A')
})

it('zeigt bei einer anderen Karte das alte Bild nicht', async () => {
  haengend(2)

  const { result, rerender } = renderHook(useProbe, {
    initialProps: { karte: 'k1', getippt: undefined, faktor: 1 },
  })

  await waitFor(() => {
    expect(result.current.gezeigt?.bild).toBe('k1@1')
  })

  rerender({ karte: 'k2', getippt: undefined, faktor: 1 })

  expect(result.current.gezeigt).toBeUndefined()
})
