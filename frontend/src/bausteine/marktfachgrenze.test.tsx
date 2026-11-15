// **Ein großes Fach baut nicht alles auf** — #2538, gefunden mit dem Messlauf
// aus #2532: 25.000 eigene Exemplare als Knöpfe hielten die Seite beim
// Aufschlagen 23 bis 45 Sekunden an. Das Fach zeigt eine begrenzte Zahl,
// lässt sich nach Namen durchsuchen, und „Mehr zeigen“ holt die nächsten.
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { expect, it, vi } from 'vitest'

import { FACHGRENZE, Fachliste, type Fachgriff } from './Marktfach'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'

import type { Fachkarte } from './marktlegen'

const tisch = kataloge.de.seiten.handeln.markttisch

const griff: Fachgriff = {
  auswahl: [],
  klicken: vi.fn(),
  taste: () => false,
  ziehen: vi.fn(),
  gegriffen: () => false,
}

function karten(zahl: number): Fachkarte[] {
  return Array.from({ length: zahl }, (_, i) => ({
    kennung: `mein:${String(i)}`,
    quelle: 'mein',
    printing: `p-${String(i)}`,
    druckId: 0,
    name: `Sammelkarte ${String(i).padStart(5, '0')}`,
    zustand: 'nearMint',
    sprache: 'de',
    finish: 'nonfoil',
    menge: 1,
  }))
}

function zeige(zahl: number) {
  render(
    <Sprachraum sprache="de">
      <Fachliste karten={karten(zahl)} leer="" griff={griff} />
    </Sprachraum>,
  )

  return userEvent.setup()
}

const kartenknoepfe = () =>
  within(screen.getByRole('list')).getAllByRole('button')

it('zeigt höchstens die Grenze und sagt, wie viele es sind', async () => {
  const hand = zeige(FACHGRENZE + 50)

  expect(kartenknoepfe()).toHaveLength(FACHGRENZE)
  expect(
    screen.getByText(
      einsetzen(tisch.fachZaehlung, {
        gezeigt: String(FACHGRENZE),
        gesamt: String(FACHGRENZE + 50),
      }),
    ),
  ).toBeInTheDocument()

  await hand.click(screen.getByRole('button', { name: tisch.fachMehr }))

  expect(kartenknoepfe()).toHaveLength(FACHGRENZE + 50)
  expect(screen.queryByRole('button', { name: tisch.fachMehr })).toBeNull()
})

it('findet eine Karte nach ihrem Namen', async () => {
  const hand = zeige(FACHGRENZE + 50)

  await hand.type(
    screen.getByRole('searchbox', { name: tisch.fachSuchen }),
    'sammelkarte 00240',
  )

  expect(
    kartenknoepfe().map((knopf) => knopf.getAttribute('aria-label')),
  ).toEqual(['Sammelkarte 00240'])

  await hand.clear(screen.getByRole('searchbox', { name: tisch.fachSuchen }))
  await hand.type(
    screen.getByRole('searchbox', { name: tisch.fachSuchen }),
    'gibt es nicht',
  )

  expect(screen.getByText(tisch.fachKeinTreffer)).toBeInTheDocument()
})

// **Ein kleines Fach bleibt, wie es war** — ohne Suche und ohne Zählzeile.
it('zeigt ein kleines Fach ganz und ohne Suche', () => {
  zeige(3)

  expect(kartenknoepfe()).toHaveLength(3)
  expect(screen.queryByRole('searchbox')).toBeNull()
})
