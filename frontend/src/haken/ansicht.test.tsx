// **Die gemerkte Ansicht eines Bereichs** — UX-11.AK-04
// (#1683).
import { act, renderHook, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { ReactNode } from 'react'

import { belegeApi, raeumeApi } from '../test/api'
import { useAnsicht } from './useAnsicht'
import { useEinstellungen } from './useEinstellungen'
import { Sprachraum } from '../sprache/kontext'

const schluessel = 'arbeitsplatz.ansicht.decks'
const formen = ['liste', 'kacheln'] as const

function einstellungen(wirksam: string) {
  return () =>
    Promise.resolve({
      einstellungen: [
        {
          schluessel,
          wert: 'erben',
          wirksam,
          werte: ['erben', 'liste', 'kacheln'],
          bereich: 'arbeitsplatz',
        },
      ],
    })
}

function Rahmen({ children }: { children: ReactNode }) {
  return <Sprachraum sprache="de">{children}</Sprachraum>
}

afterEach(() => {
  raeumeApi()
})

describe('useAnsicht', () => {
  it('liest, was gilt — auch geerbt', async () => {
    belegeApi({ Einstellungen: einstellungen('kacheln') })

    const { result } = renderHook(() => useAnsicht(schluessel, formen), {
      wrapper: Rahmen,
    })

    await waitFor(() => {
      expect(result.current.form).toBe('kacheln')
    })
  })

  // **Auf den Ladeschluss gewartet und nicht auf den Wert**: die erste Form
  // ist auch der Stand vor der Antwort (§3).
  it('nimmt für einen fremden Wert die erste Form', async () => {
    belegeApi({ Einstellungen: einstellungen('gitter') })

    const { result } = renderHook(
      () => ({
        ansicht: useAnsicht(schluessel, formen),
        geladen: useEinstellungen().geladen,
      }),
      { wrapper: Rahmen },
    )

    await waitFor(() => {
      expect(result.current.geladen).toBe(true)
    })
    expect(result.current.ansicht.form).toBe('liste')
  })

  it('schreibt die Wahl und zeigt sie sofort', async () => {
    const EinstellungSetzen = vi.fn(() => Promise.resolve({ fehler: null }))
    belegeApi({ Einstellungen: einstellungen('liste'), EinstellungSetzen })

    const { result } = renderHook(() => useAnsicht(schluessel, formen), {
      wrapper: Rahmen,
    })
    await waitFor(() => {
      expect(result.current.form).toBe('liste')
    })

    act(() => {
      result.current.setzeForm('kacheln')
    })

    expect(result.current.form).toBe('kacheln')
    await waitFor(() => {
      expect(EinstellungSetzen).toHaveBeenCalledWith(schluessel, 'kacheln')
    })
  })
})
