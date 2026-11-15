// **Mit wem man im Netz spielen kann** — P9c
// (#2377):
// bestätigte Gegenüber, denen „Eine Partie spielen“ freigegeben ist, mit dem
// Namen, den ein Mensch liest. Kein anderer — weder ein unbestätigter noch
// einer ohne das Recht.
import { renderHook, waitFor } from '@testing-library/react'
import { afterEach, expect, it } from 'vitest'

import { einKontakt, eineListe } from '../seiten/peerkontaktprobe'
import { belegeApi, raeumeApi } from '../test/api'
import { usePartiekontakte } from './usePartiekontakte'

afterEach(raeumeApi)

it('nennt die bestätigten Gegenüber mit dem Recht am Tisch', async () => {
  belegeApi({
    Peerkontakte: () =>
      Promise.resolve(
        eineListe([
          einKontakt({
            kennung: 'p-bo',
            eigenerName: 'Bo',
            stand: 'bestaetigt',
            rechte: ['partie'],
          }),
          einKontakt({
            kennung: 'p-cy',
            selbstgewaehlterName: 'Cy',
            stand: 'bestaetigt',
            rechte: ['partie', 'handel'],
          }),
          einKontakt({
            kennung: 'p-un',
            eigenerName: 'Un',
            stand: 'unbestaetigt',
            rechte: ['partie'],
          }),
          einKontakt({
            kennung: 'p-ha',
            eigenerName: 'Ha',
            stand: 'bestaetigt',
            rechte: ['handel'],
          }),
        ]),
      ),
  })

  const { result } = renderHook(() => usePartiekontakte())

  expect(result.current).toEqual([])
  await waitFor(() => {
    expect(result.current).toEqual([
      { kennung: 'p-bo', name: 'Bo' },
      { kennung: 'p-cy', name: 'Cy' },
    ])
  })
})
