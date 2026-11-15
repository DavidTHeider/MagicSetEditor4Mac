// **Der Weg zur nächsten Karte** —
// #1290.
//
// # Der Befund, gemessen am 2026-08-30 im laufenden Fenster
//
// Bei 1280 × 800, `Vanguard.mse-set`, aufgeschlagene Karte: 67 Schaltflächen,
// 108 fokussierbare Elemente, 2152 px Seitenhöhe — und die Suche nach
// „nächste / weiter / next" über **alle** Schaltflächen und Verweise war
// **leer**. Wer hundert Karten baut, musste für jede zurück in die Liste.
//
// # Was hier geprüft wird
//
//   - Beide Knöpfe sind da und tun, was sie sagen.
//   - **Am Rand sind sie abgeschaltet**, nicht unsichtbar: ein Knopf, der
//     verschwindet, verschiebt den daneben.
//   - Die Stelle steht dabei — „12 von 107". Ohne sie sagt „nächste" nicht,
//     wo man ist.

import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import { Kartenblaettern } from './Kartenblaettern'
import { Sprachraum } from '../sprache/kontext'
import type { KartenzeileDTO } from '../api'

const zeile = (id: string): KartenzeileDTO =>
  ({ id, nummer: Number(id), aufschriften: [] }) as unknown as KartenzeileDTO

const drei = [zeile('1'), zeile('2'), zeile('3')]

function zeichne(gewaehlt: string, waehlen = vi.fn()) {
  render(
    <Sprachraum sprache="de">
      <Kartenblaettern karten={drei} gewaehlt={gewaehlt} waehlen={waehlen} />
    </Sprachraum>,
  )

  return waehlen
}

describe('Kartenblaettern', () => {
  it('führt zur nächsten Karte', async () => {
    const waehlen = zeichne('2')

    await userEvent.click(screen.getByRole('button', { name: /Nächste Karte/ }))

    expect(waehlen).toHaveBeenCalledWith('3')
  })

  it('führt zur vorigen Karte', async () => {
    const waehlen = zeichne('2')

    await userEvent.click(screen.getByRole('button', { name: /Vorige Karte/ }))

    expect(waehlen).toHaveBeenCalledWith('1')
  })

  it('schaltet am Ende die nächste ab und lässt sie stehen', () => {
    zeichne('3')

    expect(screen.getByRole('button', { name: /Nächste Karte/ })).toBeDisabled()
    expect(screen.getByRole('button', { name: /Vorige Karte/ })).toBeEnabled()
  })

  it('schaltet am Anfang die vorige ab', () => {
    zeichne('1')

    expect(screen.getByRole('button', { name: /Vorige Karte/ })).toBeDisabled()
    expect(screen.getByRole('button', { name: /Nächste Karte/ })).toBeEnabled()
  })

  it('sagt nichts, solange die Karte nicht in der Liste steht', () => {
    // **Ein Zustand von Dauer eines Herzschlags**, aber einer: die Liste lädt
    // nach, und die aufgeschlagene Karte steht kurz in keiner. „0 von 3" wäre
    // dann eine Auskunft, die nicht stimmt.
    zeichne('gibtsnicht')

    expect(screen.queryByText(/von 3/)).not.toBeInTheDocument()
    expect(screen.getByRole('button', { name: /Nächste Karte/ })).toBeDisabled()
  })

  it('sagt, an welcher Stelle die Karte steht', () => {
    zeichne('2')

    expect(screen.getByText('2 von 3')).toBeInTheDocument()
  })
})
