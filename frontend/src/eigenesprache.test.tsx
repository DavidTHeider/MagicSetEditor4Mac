// **Das Fenster spricht eine eigene Übersetzung** — LOC-02.AK-03
// (#1688).
//
// Die Rückfallkette: eigene Sätze, dann ihre Basis, dann die gewählte
// Sprache, wenn es das Paket nicht gibt. **Eine leere Beschriftung gibt es
// nicht** — und ein Satz, den der Editor übernimmt, steht ohne Neuwahl da.
import { act, render, screen, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { App } from './App'
import { eigeneSpracheGeaendert } from './sprache/eigenesprache'
import { kataloge } from './sprache/kataloge'
import { belegeApi, raeumeApi } from './test/api'
import { sprachantwort, sprachpaketprobe } from './test/sprachprobe'

import type { SprachpaketAntwort } from './apisprache'

afterEach(() => {
  raeumeApi()
  document.documentElement.lang = 'de'
})

function starte(
  eigene: string,
  Sprachpaket: (kennung: string) => Promise<SprachpaketAntwort>,
) {
  belegeApi({
    Einstellungen: () =>
      Promise.resolve({
        einstellungen: [
          {
            schluessel: 'sprache',
            wert: 'de',
            wirksam: 'de',
            werte: ['de', 'en'],
            bereich: 'anwendung',
          },
          {
            schluessel: 'sprache.eigene',
            wert: eigene,
            wirksam: eigene,
            werte: [],
            bereich: 'anwendung',
          },
        ],
      }),
    SpracheSetzen: () => Promise.resolve({}),
    Sprachpaket,
  })

  render(<App />)
}

function paket(ueberschrift: string, basis = 'en') {
  return sprachantwort({
    paket: sprachpaketprobe({ kennung: 'pt-BR', basis }),
    texte: { 'seiten.start.ueberschrift': { text: ueberschrift, quelle: '' } },
  })
}

describe('Eine eigene Übersetzung im Fenster', () => {
  it('legt ihre Sätze über die Basis und nennt sich im Dokument', async () => {
    const Sprachpaket = vi
      .fn()
      .mockResolvedValueOnce(paket('Olá'))
      .mockResolvedValue(paket('Olá de novo'))
    starte('pt-BR', Sprachpaket)

    expect(
      await screen.findByRole('heading', { name: 'Olá' }),
    ).toBeInTheDocument()
    await waitFor(() => {
      expect(document.documentElement.lang).toBe('pt-BR')
    })

    // **Eine andere Übersetzung ändert diese nicht** — und lädt nichts.
    act(() => {
      eigeneSpracheGeaendert('da')
    })
    expect(Sprachpaket).toHaveBeenCalledTimes(1)

    act(() => {
      eigeneSpracheGeaendert('pt-BR')
    })
    expect(
      await screen.findByRole('heading', { name: 'Olá de novo' }),
    ).toBeInTheDocument()
  })

  // **Erst geladen, dann fort** — so ist der Rückfall eine Wirkung, die man
  // sieht. Bis zum 2026-09-30 wartete der Lauf auf die deutsche Überschrift,
  // die schon vor der Antwort dastand, und betrat den Fehlerweg nur, wenn die
  // Antwort zufällig früh genug kam (§3).
  it('fällt auf die gewählte Sprache zurück, wenn es das Paket nicht mehr gibt', async () => {
    const Sprachpaket = vi
      .fn()
      .mockResolvedValueOnce(paket('Olá'))
      .mockResolvedValue(
        sprachantwort({
          fehler: {
            code: 'nichtGefunden',
            schluessel: 'nichtGefunden',
            meldung: 'fort',
            feld: '',
            werte: {},
          },
        }),
      )
    starte('pt-BR', Sprachpaket)

    expect(
      await screen.findByRole('heading', { name: 'Olá' }),
    ).toBeInTheDocument()

    act(() => {
      eigeneSpracheGeaendert('pt-BR')
    })

    expect(
      await screen.findByRole('heading', {
        name: kataloge.de.seiten.start.ueberschrift,
      }),
    ).toBeInTheDocument()
  })

  it('fällt zurück, wenn die Basis keine Fenstersprache ist', async () => {
    starte('pt-BR', () => Promise.resolve(paket('Olá', 'xx')))

    expect(
      await screen.findByRole('heading', {
        name: kataloge.de.seiten.start.ueberschrift,
      }),
    ).toBeInTheDocument()
    expect(
      screen.queryByRole('heading', { name: 'Olá' }),
    ).not.toBeInTheDocument()
  })
})
