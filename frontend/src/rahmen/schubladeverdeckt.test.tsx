// **Unter einer Schublade blättert keine Karte** — UX-08
// (#1680).
//
// # Warum durch die ganze Anwendung
//
// **Weil die Zusage eine Naht ist.** Der Haken hört auf `Verdeckung`
// (`haken/useKartenblaettern.test.tsx`); ob der Arbeitsplatz sie der
// Kartenseite **unter** der Schublade auch gibt, sieht nur ein Lauf durch
// `Seiteninhalt`. Ohne sie wechselte ⇟ in der Konsole die Karte darunter,
// die niemand sieht.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { App } from '../App'
import { kataloge } from '../sprache/kataloge'
import { belegeApi, belegeFenster, raeumeApi } from '../test/api'
import { effekteNachholen } from '../test/effekte'
import { ausKarten } from '../test/kartenproben'
import { leisteneintrag } from '../test/leistenweg'

const texte = kataloge.de
const worte = texte.allgemein.arbeitsplatz.schubladen

afterEach(() => {
  raeumeApi()
})

describe('Die Kartenseite unter einer Schublade', () => {
  it('blättert nicht, solange die Schublade steht, und danach wieder', async () => {
    const wege = ausKarten(
      () =>
        Promise.resolve({
          karten: [
            { id: '11', nummer: 1, werte: [{ feld: 'name', wert: 'Erste' }] },
            { id: '22', nummer: 2, werte: [{ feld: 'name', wert: 'Zweite' }] },
          ],
        }),
      '11',
    )
    const Karte = vi.fn(wege.Karte)

    belegeApi({
      ...belegeFenster('de'),
      ...wege,
      Karte,
      Sets: () =>
        Promise.resolve({
          sets: [
            {
              id: '7',
              name: 'Probeset',
              spiel: 'magic',
              stil: 'magic-m15',
              karten: 2,
              herkunft: '/a.mse-set',
              herkunftFehlt: false,
            },
          ],
        }),
      Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
      Statistik: () =>
        Promise.resolve({
          statistik: { karten: 2, felder: 1, leere: 0, belegung: [] },
        }),
      Verteilung: () => Promise.resolve({ verteilung: [] }),
    })

    render(<App />)

    await userEvent.click(await leisteneintrag('sets'))
    await userEvent.click(
      await screen.findByRole('button', {
        name: `Probeset — ${texte.seiten.sets.oeffnen}`,
      }),
    )

    // **Die erste Karte ist aufgeschlagen** — beim Öffnen die Vorgabe (#668).
    await screen.findByRole('button', {
      name: `Zweite — ${texte.seiten.karten.bearbeiten}`,
    })

    const knopf = screen.getByRole('button', { name: worte.werkzeuge.knopf })
    await userEvent.click(knopf)

    const schublade = await screen.findByRole('dialog', {
      name: worte.werkzeuge.titel,
    })
    await waitFor(() => {
      expect(schublade).toContainElement(document.activeElement as HTMLElement)
    })

    await userEvent.keyboard('{PageDown}')
    await effekteNachholen()

    expect(Karte).not.toHaveBeenCalledWith('22')

    // **Der Beleg, dass jemand zuhört** (§3): ohne Schublade blättert
    // dieselbe Taste zur zweiten Karte.
    await userEvent.keyboard('{Escape}')
    await waitFor(() => {
      expect(
        screen.queryByRole('dialog', { name: worte.werkzeuge.titel }),
      ).toBeNull()
    })
    await effekteNachholen()

    await userEvent.keyboard('{PageDown}')

    await waitFor(() => {
      expect(Karte).toHaveBeenCalledWith('22')
    })
  })
})
