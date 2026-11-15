// **Die Seite zeichnet nicht fünftausend Zeilen** — Werkstück 1 von Phase 4.5
// (#1292),
// seit U8 (#1419) durch die **Fensterliste** eingelöst (Vermerk 0085).
//
// # Der Befund, der diese Tests verlangt
//
// Gemessen am 2026-08-30 an der laufenden Brücke, 1280 × 800, gegen einen
// Bestand von 5.000 Exemplaren (`scripts/sammlungsmass.mjs`): 5.000 Zeilen im
// DOM, 1,2 bis 1,5 s bis die Seite stand — davon rund eine Sekunde React, das
// die Zeilen zeichnet.
//
// # Von der Grenze zur Fensterung
//
// Bis U8 kappte die Seite bei 200 Zeilen und bot „mehr zeigen" an. Das hielt
// die Zahl im DOM klein, aber eine Auswahl über nicht gezeichnete Zeilen und
// eine Tastatur, die jede Zeile erreicht, gingen damit nicht — genau die
// Abnahme von U8. Die Fensterliste zeichnet, was ins Fenster passt, und
// nennt trotzdem alle (`aria-rowcount`); §10 gilt weiter: die Zahl steht da.
import { keineHerkunft } from '../apiherkunft'
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, expect, it, vi } from 'vitest'

import type { ExemplarDTO } from '../api'
import { Sammlung } from './Sammlung'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'

afterEach(() => {
  raeumeApi()
})

// viele gibt so viele Exemplare, wie verlangt — jedes mit eigenem Namen.
function viele(anzahl: number): ExemplarDTO[] {
  return Array.from({ length: anzahl }, (_, nummer) => ({
    id: nummer + 1,
    fassung: 1,
    printing: `druck-${String(nummer)}`,
    druckID: '',
    vorlaeufig: false,
    menge: 1,
    zustand: 'nearMint',
    sprache: 'en',
    finish: 'nonfoil',
    erwerbsdatum: '',
    erwerbspreis: '',
    erwerbswaehrung: '',
    notiz: '',
    name: `Karte ${String(nummer).padStart(4, '0')}`,
    setcode: 'p01',
    nummer: String(nummer),
    lagerort: '',
    wunsch: false,
    bezeichnung: `Karte ${String(nummer).padStart(4, '0')}`,
    herkunft: keineHerkunft,
  }))
}

function zeige(exemplare: ExemplarDTO[]) {
  belegeApi({
    Sammlung: () => Promise.resolve({ exemplare, fehler: null }),

    // **Die Sicht wirkt in Go** (#1500) — der Beleg bildet die Suche nach, so
    // weit dieser Fall sie braucht, und gibt den verlangten Ausschnitt.
    Sammlungsfenster: (
      suche: string,
      _ausfuehrung: string,
      _sortierung: string,
      ab: number,
      anzahl: number,
    ) => {
      const gesehen = exemplare.filter(
        (eines) =>
          suche === '' ||
          eines.bezeichnung.toLowerCase().includes(suche.toLowerCase()),
      )

      return Promise.resolve({
        exemplare: gesehen.slice(ab, anzahl > 0 ? ab + anzahl : undefined),
        gesamt: gesehen.length,
        insgesamt: exemplare.length,
        stueck: exemplare.length,
        fehler: null,
      })
    },
    Portfolio: () =>
      Promise.resolve({ zeilen: [], ohnePreis: 0, fehler: null }),
    Preise: () => Promise.resolve({ staende: [], befunde: [], fehler: null }),
  })

  render(
    <Sprachraum sprache="de">
      <Sammlung bestandGeaendert={vi.fn()} />
    </Sprachraum>,
  )
}

// **Über der Grenze wird gekappt — und die Kappung steht da.**
beforeEach(() => {
  Object.defineProperty(HTMLElement.prototype, 'clientHeight', {
    configurable: true,
    get: () => 600,
  })
})

it('zeichnet nur das Fenster und nennt trotzdem alle Zeilen', async () => {
  zeige(viele(5000))

  const gitter = await screen.findByRole('grid', { name: 'Exemplare' })
  expect(gitter).toHaveAttribute('aria-rowcount', '5001')

  await waitFor(() => {
    expect(screen.getAllByRole('row').length).toBeGreaterThan(10)
  })
  // Deutlich unter fünftausend — und die letzte Zeile steht nicht im DOM.
  expect(screen.getAllByRole('row').length).toBeLessThan(60)
  expect(screen.queryByText('Karte 4999')).not.toBeInTheDocument()
})

it('erreicht die letzte Zeile mit End und zeichnet sie dann', async () => {
  zeige(viele(5000))

  await screen.findByText('Karte 0000')
  const erste = screen.getByText('Karte 0000').closest('[role="row"]')
  if (!(erste instanceof HTMLElement)) {
    throw new Error('keine Zeile')
  }
  erste.focus()
  await userEvent.keyboard('{End}')

  const gitter = within(screen.getByRole('grid', { name: 'Exemplare' }))
  expect(await gitter.findByText('Karte 4999')).toBeInTheDocument()
  expect(gitter.getByText('Karte 4999').closest('[role="row"]')).toHaveFocus()
})

it('lässt eine kurze Liste ganz stehen', async () => {
  zeige(viele(5))

  const gitter = await screen.findByRole('grid', { name: 'Exemplare' })
  expect(gitter).toHaveAttribute('aria-rowcount', '6')
  await waitFor(() => {
    expect(screen.getAllByRole('row')).toHaveLength(6)
  })
})

it('zeigt nach einer Suche die Treffer und nennt ihre Zahl', async () => {
  zeige(viele(5000))

  await screen.findByText('Karte 0000')
  await userEvent.type(
    screen.getByLabelText(kataloge.de.seiten.sammlung.suche),
    'Karte 49',
  )

  await waitFor(() => {
    expect(screen.getByRole('grid', { name: 'Exemplare' })).toHaveAttribute(
      'aria-rowcount',
      '101',
    )
  })
  expect(screen.getByText('Karte 4900')).toBeInTheDocument()
})
