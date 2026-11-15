// **Was am Bestand hängt, liest nach, wenn ein anderer schreibt** — Befunde
// 3 bis 5 des Nachlaufs zum Gesamtreview über 0.6.0 (#1660).
//
// #2062 gab den fremden Stand an die Listen der Seiten. Drei Lesewege hingen
// weiter nur an ihrem eigenen Anlass: die Nutzer und Profile, die Paketliste
// der Vorlagenwerkstatt und die Feldformen eines Sets. Legte die
// Kommandozeile einen Nutzer an oder tauschte ein Paket, stand im offenen
// Fenster der alte Stand — und im Karteneditor neue Werte neben alten Formen.
import { act, renderHook, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { deckprobe } from '../seiten/deckprobe'
import { useDecks } from './useDecks'
import { useKartendaten } from './useKartendaten'
import { useNutzer } from './useNutzer'
import { useVorlagenwerkstatt } from './useVorlagenwerkstatt'

import type { ReactNode } from 'react'

function Rahmen({ children }: { children: ReactNode }) {
  return <Sprachraum sprache="de">{children}</Sprachraum>
}

afterEach(() => {
  raeumeApi()
})

// nachladen prüft, dass ein höherer Stand die Bindung noch einmal fragt.
//
// **Erst die erste Antwort, dann der Wechsel** (§3): wer vorher wechselte,
// sähe zwei Rufe aus dem ersten Laden.
async function nachladen(
  gefragt: { mock: { calls: unknown[] } },
  wechsle: (stand: number) => void,
) {
  await waitFor(() => {
    expect(gefragt.mock.calls.length).toBeGreaterThan(0)
  })
  const vorher = gefragt.mock.calls.length

  wechsle(1)

  await waitFor(() => {
    expect(gefragt.mock.calls.length).toBeGreaterThan(vorher)
  })
}

describe('Ein fremder Stand', () => {
  it('lädt die Nutzer und Profile nach', async () => {
    const Nutzer = vi.fn(() => Promise.resolve({ nutzer: [], fehler: null }))

    belegeApi({
      Nutzer,
      GezeigtesProfil: () => Promise.resolve({ profil: null, fehler: null }),
      WerArbeitet: () => Promise.resolve({ nutzerID: '', fehler: null }),
      Anbieter: () =>
        Promise.resolve({ anbieter: [], staende: [], fehler: null }),
      Nutzerbilder: () =>
        Promise.resolve({ bilder: {}, badges: {}, fehler: null }),
    })

    const { rerender } = renderHook(({ stand }) => useNutzer(stand), {
      wrapper: Rahmen,
      initialProps: { stand: 0 },
    })

    await nachladen(Nutzer, (stand) => {
      rerender({ stand })
    })
  })

  it('lädt die Paketliste der Vorlagenwerkstatt nach', async () => {
    const Pakete = vi.fn(() =>
      Promise.resolve({ pakete: [], befunde: [], fehler: null }),
    )

    belegeApi({ Pakete })

    const { rerender } = renderHook(
      ({ stand }) => useVorlagenwerkstatt(stand),
      { wrapper: Rahmen, initialProps: { stand: 0 } },
    )

    await nachladen(Pakete, (stand) => {
      rerender({ stand })
    })
  })

  it('lädt die Feldformen eines Sets nach', async () => {
    const Kartenfelder = vi.fn(() =>
      Promise.resolve({ felder: [], fehler: null }),
    )

    belegeApi({
      Kartenzeilen: () => Promise.resolve({ zeilen: [], fehler: null }),
      Kartenfelder,
    })

    const { rerender } = renderHook(
      ({ stand }) =>
        useKartendaten({
          setID: '1',
          verlaufsstand: stand,
          gewaehlt: undefined,
        }),
      { wrapper: Rahmen, initialProps: { stand: 0 } },
    )

    await nachladen(Kartenfelder, (stand) => {
      rerender({ stand })
    })
  })
})

// **Ein Nachlesen räumt keine Meldung weg** (§7, Regel 3) — die Nebenwirkung,
// nach der der Nachlauf fragte. Wer nach einem gescheiterten Schreiben die
// Meldung liest, verlöre sie sonst, sobald irgendwer anders etwas schreibt.
describe('Das Nachlesen des offenen Decks', () => {
  it('lässt eine stehende Meldung stehen', async () => {
    let name = 'Aggro'
    const Deck = vi.fn(() =>
      Promise.resolve({ deck: deckprobe({ name }), fehler: null }),
    )

    belegeApi({
      Decks: () => Promise.resolve({ decks: [deckprobe()], fehler: null }),
      Formate: () => Promise.resolve({ formate: [], fehler: null }),
      Deck,
      DeckBenutzt: () => Promise.resolve({ fehler: null }),
    })

    const { result, rerender } = renderHook(({ stand }) => useDecks(stand), {
      wrapper: Rahmen,
      initialProps: { stand: 0 },
    })

    await act(() => result.current.waehle('1'))
    act(() => {
      result.current.melde(new Error('Das Schreiben ist gescheitert.'))
    })
    const meldung = result.current.fehler
    expect(meldung).toBeDefined()

    name = 'Kontrolle'
    rerender({ stand: 1 })

    // **Auf die Wirkung gewartet**: das nachgelesene Deck steht da.
    await waitFor(() => {
      expect(result.current.offen?.name).toBe('Kontrolle')
    })
    expect(result.current.fehler).toBe(meldung)
  })
})
