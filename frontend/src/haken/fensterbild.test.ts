// **Das Fensterbild gehört dem Menschen, der es gewählt hat** — Nachlauf über
// Phase 6.3, N6.
//
// Bis dahin hielt nur der `key` am `Gestaltungsraum` es fest: wer ihn
// entfernte, ließ jeden Test grün, und der nächste Mensch sah das Bild, das
// der vorige in dieser Sitzung gewählt hatte. Jetzt hält der Haken selbst,
// wem ein gewähltes Bild gehört — und derselbe Haken trägt das Tischbild.
import { act, renderHook, waitFor } from '@testing-library/react'
import { afterEach, expect, it } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'
import { useFensterbild } from './useFensterbild'

afterEach(() => {
  raeumeApi()
})

it('vergisst beim Wechsel des Menschen, was der vorige gewählt hat', async () => {
  belegeApi({
    Fensterbild: () => Promise.resolve({ adresse: '' }),
    FensterbildWaehlen: () =>
      Promise.resolve({ adresse: '/fensterbild/a.png' }),
  })

  const lauf = renderHook(({ wem }) => useFensterbild(wem), {
    initialProps: { wem: 'a' },
  })

  await waitFor(() => {
    expect(lauf.result.current.bereit).toBe(true)
  })

  act(() => {
    lauf.result.current.waehlen()
  })

  await waitFor(() => {
    expect(lauf.result.current.adresse).toBe('/fensterbild/a.png')
  })

  lauf.rerender({ wem: 'b' })

  await waitFor(() => {
    expect(lauf.result.current.bereit).toBe(true)
  })
  expect(lauf.result.current.adresse).toBe('')
})
