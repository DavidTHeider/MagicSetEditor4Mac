// **Der Markenfilter fragt Go** — UX-12.AK-03
// (#1684).
import { act, renderHook, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { nachMarken, useMarkenfilter } from './useMarkenfilter'
import { belegeApi, raeumeApi } from '../test/api'

afterEach(() => {
  raeumeApi()
})

describe('useMarkenfilter', () => {
  it('liest die Marken je Objekt und fragt erst mit einer Wahl', async () => {
    const je = vi
      .fn()
      .mockResolvedValue({ marken: { '31': ['tausch'] }, fehler: null })
    const filtern = vi
      .fn()
      .mockResolvedValue({ kennungen: ['31'], fehler: null })
    belegeApi({ MarkenJeObjekt: je, MarkenFiltern: filtern })

    const { result } = renderHook(() => useMarkenfilter('exemplar'))

    await waitFor(() => {
      expect(result.current.je).toEqual({ '31': ['tausch'] })
    })
    expect(result.current.getroffen).toBeUndefined()
    expect(filtern).not.toHaveBeenCalled()

    act(() => {
      result.current.setzeFilter({ marken: ['tausch'], verknuepfung: 'oder' })
    })

    await waitFor(() => {
      expect(result.current.getroffen).toEqual(new Set(['31']))
    })
    expect(filtern).toHaveBeenCalledWith('exemplar', ['tausch'], 'oder')

    // **Nachladen fragt beides neu** — nach einer gesetzten Marke.
    act(() => {
      result.current.nachladen()
    })
    await waitFor(() => {
      expect(je).toHaveBeenCalledTimes(2)
    })
  })

  it('meldet einen Fehler aus Go', async () => {
    belegeApi({
      MarkenJeObjekt: vi.fn().mockResolvedValue({
        marken: {},
        fehler: { code: 'validierung', meldung: 'kaputt', schluessel: '' },
      }),
    })

    const { result } = renderHook(() => useMarkenfilter('deck'))

    await waitFor(() => {
      expect(result.current.fehler).toBe('kaputt')
    })
    expect(result.current.je).toEqual({})
  })
})

describe('nachMarken', () => {
  const zeilen = [{ id: 'a' }, { id: 'b' }, { id: undefined }]

  it('lässt ohne Filter alles stehen, sonst nur die Treffer und was keine Kennung hat', () => {
    expect(nachMarken(zeilen, (zeile) => zeile.id, undefined)).toHaveLength(3)
    expect(nachMarken(zeilen, (zeile) => zeile.id, new Set(['b']))).toEqual([
      { id: 'b' },
      { id: undefined },
    ])
  })
})
