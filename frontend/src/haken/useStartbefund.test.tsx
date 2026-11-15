// **Der Startbefund und die überholte Antwort** —
// #962.
//
// Der Weg durch die Anwendung steht in `src/wiederherstellung.test.tsx`.
// Hier steht die eine Zusage, die dort nicht herzustellen ist: **eine Antwort
// gehört zu dem Stand, aus dem sie stammt** (§7, Regel 2).
import { act, renderHook, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'
import { useStartbefund } from './useStartbefund'

afterEach(() => {
  raeumeApi()
})

describe('Ein Startbefund, den ein neuerer überholt', () => {
  it('trägt sich nicht mehr ein', async () => {
    let ersteLoesen: (wert: { offen: boolean; grund: string }) => void = () =>
      undefined

    const erste = new Promise<{ offen: boolean; grund: string }>((ja) => {
      ersteLoesen = ja
    })

    const Startbefund = vi
      .fn()
      .mockReturnValueOnce(erste)
      .mockResolvedValue({ offen: true, grund: '' })

    belegeApi({ Startbefund })

    const { result, rerender } = renderHook(
      ({ stand }: { stand: number }) => useStartbefund(stand),
      { initialProps: { stand: 0 } },
    )

    // **Die zweite Frage geht hinaus, bevor die erste beantwortet ist** —
    // genau das geschieht nach einem geglückten Zurückspielen.
    rerender({ stand: 1 })

    await waitFor(() => {
      expect(result.current.geladen).toBe(true)
    })

    expect(result.current.offen).toBe(true)

    // **Und jetzt kommt die erste Antwort — verspätet und mit einem anderen
    // Befund.** Ohne den Zähler stünde danach „nicht offen" da, und die
    // Anwendung spränge in die Wiederherstellung zurück.
    await act(async () => {
      ersteLoesen({ offen: false, grund: 'der Stand von vorhin' })
      await erste
    })

    expect(result.current.offen).toBe(true)
    expect(result.current.grund).toBe('')
  })
})
