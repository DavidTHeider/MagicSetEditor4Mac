// **Die Kartenliste ist eine Tabstelle, nicht hundert** —
// #1290.
//
// # Der Befund, gemessen am 2026-08-30 an der laufenden Brücke
//
// Bei 1280 × 800, `Vanguard.mse-set`, aufgeschlagene Karte: **85**
// Tabulatorschritte bis zum ersten Kartenfeld. Aufgeteilt:
//
//	Kartenliste             54
//	Laschen                 13
//	Kopfzeile und Werkzeuge 11
//	Seitenleiste             7
//
// Die Liste ist also **zwei Drittel** des Weges — 27 gezeichnete Zeilen mal
// zwei Bedienelemente. Und das sind nur die gezeichneten: die Liste rollt
// virtuell, bei 3.000 Karten stünden dort dieselben 54.
//
// # Wie eine Liste sich verhält
//
// MSE2 zeigt seine Karten in einem `wxListCtrl` (`src/gui/control/card_list.cpp`,
// Fassung f9d9356 vom 2021-07-06). Ein solches Steuerelement ist **eine**
// Tabstelle; innerhalb bewegen die Pfeiltasten. Das ist keine Eigenheit von wx,
// sondern wie eine Liste auf jeder Plattform bedient wird — und es ist die
// ARIA-Gestalt `grid` mit wanderndem `tabindex`.
//
// # Was hier zugesichert wird
//
//   - **Aus der Liste heraus führt ein Tabulatorschritt.** Sie hat genau ein
//     Element mit `tabIndex >= 0`.
//   - ↑ und ↓ bewegen die Auswahl und schlagen die Karte auf — wie MSEs
//     `selectPrevious`/`selectNext`, und mit denselben Grenzen.
//   - Pos1 und Ende führen zur ersten und zur letzten.
//   - **Das Ankreuzfeld bleibt erreichbar** — über → und ←, nicht über den
//     Tabulator.

import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import type { KartenzeileDTO } from '../api'
import { Sprachraum } from '../sprache/kontext'
import { Kartenliste } from './Kartenliste'

function karte(id: string, nummer: number, name: string): KartenzeileDTO {
  return { id, nummer, name, felder: 1 }
}

const drei = [
  karte('1', 0, 'Erste'),
  karte('2', 1, 'Zweite'),
  karte('3', 2, 'Dritte'),
]

function zeige(
  gewaehlt: string | undefined,
  waehlen = vi.fn(),
  anhaken = vi.fn(),
  mitKaestchen = false,
) {
  render(
    <Sprachraum sprache="de">
      <Kartenliste
        spalten={[]}
        marken={{}}
        setID="7"
        stand={0}
        karten={drei}
        gewaehlt={gewaehlt}
        waehlen={waehlen}
        angehakt={[]}
        anhaken={anhaken}
        mitKaestchen={mitKaestchen}
      />
    </Sprachraum>,
  )

  return { waehlen, anhaken }
}

// tabstellen zählt, was der Tabulator in der Liste erreicht.
function tabstellen() {
  return [
    ...screen
      .getByRole('grid')
      .querySelectorAll(
        '[tabindex="0"], a[href], button:not([tabindex]), input:not([tabindex])',
      ),
  ]
}

describe('Die Kartenliste als eine Tabstelle', () => {
  it('hat genau eine Tabstelle, gleich wie viele Karten sie zeigt', () => {
    zeige('2')

    expect(tabstellen()).toHaveLength(1)
  })

  it('hält die Tabstelle auf der aufgeschlagenen Karte', () => {
    zeige('2')

    expect(tabstellen()[0]).toHaveAccessibleName(/Zweite/)
  })

  it('nimmt ohne Auswahl die erste Zeile als Tabstelle', () => {
    zeige(undefined)

    expect(tabstellen()[0]).toHaveAccessibleName(/Erste/)
  })

  it('schlägt mit ↓ die nächste Karte auf', async () => {
    const { waehlen } = zeige('2')

    screen.getByRole('row', { name: /Zweite/ }).focus()
    await userEvent.keyboard('{ArrowDown}')

    expect(waehlen).toHaveBeenCalledWith('3')
  })

  it('schlägt mit ↑ die vorige auf', async () => {
    const { waehlen } = zeige('2')

    screen.getByRole('row', { name: /Zweite/ }).focus()
    await userEvent.keyboard('{ArrowUp}')

    expect(waehlen).toHaveBeenCalledWith('1')
  })

  it('bleibt an den Rändern stehen — es gibt keinen Umlauf', async () => {
    const { waehlen } = zeige('3')

    screen.getByRole('row', { name: /Dritte/ }).focus()
    await userEvent.keyboard('{ArrowDown}')

    // **Erst der Beleg, dass jemand zuhört** (§3).
    await userEvent.keyboard('{ArrowUp}')
    expect(waehlen).toHaveBeenCalledWith('2')
    expect(waehlen).toHaveBeenCalledTimes(1)
  })

  it('führt mit Pos1 zur ersten und mit Ende zur letzten', async () => {
    const { waehlen } = zeige('2')

    screen.getByRole('row', { name: /Zweite/ }).focus()
    await userEvent.keyboard('{Home}')
    expect(waehlen).toHaveBeenCalledWith('1')

    await userEvent.keyboard('{End}')
    expect(waehlen).toHaveBeenCalledWith('3')
  })

  // **Das Ankreuzfeld steht seit UXR-05 (#1913) nur im Auswahlmodus** —
  // dort ist es mit → zu erreichen, wie vorher.
  it('lässt das Ankreuzfeld im Auswahlmodus über → erreichen', async () => {
    zeige('2', vi.fn(), vi.fn(), true)

    const zeile = screen.getByRole('row', { name: /Zweite/ })
    zeile.focus()
    await userEvent.keyboard('{ArrowRight}')

    expect(document.activeElement).toHaveAccessibleName(
      /Zweite — Für die Massenänderung anhaken/,
    )
  })

  // **Ohne ihn führt → gleich auf den Knopf der Zeile** — es gibt kein
  // Kästchen, über das der Weg ginge.
  it('führt ohne Auswahlmodus mit → auf den Knopf der Zeile', async () => {
    zeige('2')

    const zeile = screen.getByRole('row', { name: /Zweite/ })
    zeile.focus()
    await userEvent.keyboard('{ArrowRight}')

    expect(document.activeElement).toHaveAccessibleName(/Zweite — Bearbeiten/)
  })
})
