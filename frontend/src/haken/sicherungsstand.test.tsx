// **Das Alter der letzten Sicherung und der Plan** — OPS01
// (#2347).
import { act, renderHook, waitFor } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import type {
  Sicherungsstand,
  SicherungsstandAntwort,
} from '../antwortenbestand'
import { belegeApi, raeumeApi } from '../test/api'
import { standMit } from '../test/sicherungsstand'
import { Sprachraum } from '../sprache/kontext'
import { abstand } from './useBestandsstand'
import { useSicherungsstand } from './useSicherungsstand'
import { useVorgang } from './useVorgang'

beforeEach(() => {
  vi.useFakeTimers({ shouldAdvanceTime: true })
})

afterEach(() => {
  vi.useRealTimers()
  raeumeApi()
})

function rahmen({ children }: { children: React.ReactNode }) {
  return <Sprachraum sprache="de">{children}</Sprachraum>
}

const antwort = (
  stand: Sicherungsstand,
  teile: Partial<SicherungsstandAntwort> = {},
): SicherungsstandAntwort => ({ stand, fehler: null, ...teile })

function zeige(nachlesen: unknown = 0) {
  return renderHook(
    ({ stand }: { stand: unknown }) => useSicherungsstand(useVorgang(), stand),
    { wrapper: rahmen, initialProps: { stand: nachlesen } },
  )
}

describe('useSicherungsstand', () => {
  it('liest den Stand beim Aufbau und nach einer eigenen Sicherung', async () => {
    const lesen = vi
      .fn()
      .mockResolvedValueOnce(antwort(standMit()))
      .mockResolvedValue(
        antwort(
          standMit({ stufe: 'frisch', gesichert: '2026-10-15T09:30:00Z' }),
        ),
      )

    belegeApi({ Sicherungsstand: lesen })

    const { result, rerender } = zeige()

    await waitFor(() => {
      expect(result.current.stand?.stufe).toBe('unbekannt')
    })

    rerender({ stand: '/x/b.msesicherung' })

    await waitFor(() => {
      expect(result.current.stand?.stufe).toBe('frisch')
    })
  })

  // **Nach dem Einrichten sichert der Plan im eigenen Prozess** — das bemerkt
  // kein fremder Stand; gefragt wird, solange er fällig ist.
  it('richtet ein und fragt nach, bis der Plan gesichert hat', async () => {
    const faellig = standMit({
      plan: { ordner: '/Stick', letzter: '', faellig: true },
    })
    const gesichert = standMit({
      stufe: 'frisch',
      anlass: 'plan',
      gesichert: '2026-10-15T09:31:00Z',
      plan: {
        ordner: '/Stick',
        letzter: '2026-10-15T09:31:00Z',
        faellig: false,
      },
    })

    const lesen = vi
      .fn()
      .mockResolvedValueOnce(antwort(standMit()))
      .mockResolvedValueOnce(antwort(faellig))
      .mockResolvedValue(antwort(gesichert))
    const einrichten = vi.fn().mockResolvedValue(antwort(faellig))

    belegeApi({ Sicherungsstand: lesen, SicherungsplanEinrichten: einrichten })

    const { result } = zeige()

    await waitFor(() => {
      expect(result.current.stand?.plan.ordner).toBe('')
    })

    act(() => {
      result.current.einrichten()
    })

    await waitFor(() => {
      expect(result.current.stand?.plan.faellig).toBe(true)
    })

    await act(async () => {
      await vi.advanceTimersByTimeAsync(abstand)
    })

    await waitFor(() => {
      expect(result.current.stand?.plan.letzter).toBe('2026-10-15T09:31:00Z')
    })

    expect(einrichten).toHaveBeenCalledTimes(1)
  })

  it('beendet den Plan und nennt einen Fehler des Plans und der Bindung', async () => {
    const beenden = vi.fn().mockResolvedValue({
      stand: standMit(),
      fehler: { code: 'belegt', schluessel: 'belegt', meldung: 'belegt' },
    })

    belegeApi({
      Sicherungsstand: vi.fn().mockResolvedValue(
        antwort(
          standMit({ plan: { ordner: '/Stick', letzter: '', faellig: true } }),
          {
            planfehler: {
              code: 'nichtGefunden',
              schluessel: 'nichtGefunden',
              meldung: 'fehlt',
            },
          },
        ),
      ),
      SicherungsplanBeenden: beenden,
    })

    const { result } = zeige()

    await waitFor(() => {
      expect(result.current.planfehler).toBeDefined()
    })

    act(() => {
      result.current.beenden()
    })

    await waitFor(() => {
      expect(result.current.fehler).toBeDefined()
    })

    expect(beenden).toHaveBeenCalledTimes(1)
  })
})
