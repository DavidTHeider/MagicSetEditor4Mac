// **Womit das Kartenformular anfängt** — UX-05.AK-01
// (#1677).
//
// # Wogegen hier geprüft wird
//
// **Dass die Tiefe eine Einstellung ist und kein Zustand der Seite.** Das
// Kriterium sagt „ein je Nutzer oder Profil änderbarer Standardmodus"; wer
// sie in einem `useState` hielte, fände beim nächsten Aufschlagen wieder
// `normal` vor.
import { act, renderHook, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { ReactNode } from 'react'

import { belegeApi, raeumeApi } from '../test/api'
import { Feldtiefeschluessel, useFeldtiefe } from './useFeldtiefe'
import { Sprachraum } from '../sprache/kontext'

function einstellungen(wert: string) {
  return () =>
    Promise.resolve({
      einstellungen: [
        {
          schluessel: Feldtiefeschluessel,
          wert,
          wirksam: wert,
          werte: ['normal', 'erweitert', 'experte'],
          bereich: 'anwendung',
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

describe('useFeldtiefe', () => {
  it('liest die gemerkte Tiefe', async () => {
    belegeApi({ Einstellungen: einstellungen('experte') })

    const { result } = renderHook(() => useFeldtiefe(), { wrapper: Rahmen })

    await waitFor(() => {
      expect(result.current.tiefe).toBe('experte')
    })
  })

  // **Ein fremder Wert ist `normal`** (§8): ein von Hand geänderter Bestand
  // kann dort alles stehen haben, und ein leeres Formular wäre schlechter
  // als das flachste.
  it('nimmt für einen fremden Wert die flachste', async () => {
    belegeApi({ Einstellungen: einstellungen('irgendwas') })

    const { result } = renderHook(() => useFeldtiefe(), { wrapper: Rahmen })

    // **Auf den Ladeschluss gewartet und nicht auf den Wert**: `normal` ist
    // auch der Stand vor der Antwort, und eine Zusicherung, die schon
    // vorher zutrifft, prüft nichts (§3).
    await waitFor(() => {
      expect(result.current.tiefe).toBe('normal')
    })

    expect(result.current.tiefe).toBe('normal')
  })

  // **Ein Wechsel gilt sofort und wird gemerkt** (§7, Regel 3): was der
  // Schreibweg zurückgibt, ist die Vorgabe des nächsten Mals; bis dahin
  // steht die gewählte Tiefe — sonst spränge das Formular kurz zurück.
  it('schreibt die gewählte Tiefe und zeigt sie sofort', async () => {
    const EinstellungSetzen = vi.fn(() => Promise.resolve({ fehler: null }))

    belegeApi({
      Einstellungen: einstellungen('normal'),
      EinstellungSetzen,
    })

    const { result } = renderHook(() => useFeldtiefe(), { wrapper: Rahmen })

    await waitFor(() => {
      expect(result.current.tiefe).toBe('normal')
    })

    act(() => {
      result.current.aufTiefe('erweitert')
    })

    expect(result.current.tiefe).toBe('erweitert')

    await waitFor(() => {
      expect(EinstellungSetzen).toHaveBeenCalledWith(
        Feldtiefeschluessel,
        'erweitert',
      )
    })
  })
})
