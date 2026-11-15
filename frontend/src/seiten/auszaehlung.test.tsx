// **Auszählen nach dem, was das Spielpaket erklärt** — Werkstück C von Phase
// 4.11 (#1338).
//
// # Was MSE2 hier tut
//
// Es zeigt links die Dimensionen als Liste und rechts ein Diagramm
// (`src/gui/set/stats_panel.cpp`, Fassung f9d9356 vom 2021-07-06). Hier stehen
// sie in der Wahl „Worüber" über demselben Bild wie die Belegung —
// `magic.mse-game` erklärt fünfzig Dimensionen, und fünfzig Zeilen neben
// einem Diagramm wären die halbe Seite.
//
// # Warum die Fälle über die ganze Seite gehen
//
// **Seit UXR-08 hat der Abschnitt kein eigenes Auswahlfeld mehr**
// (#1916):
// die Dimensionen sind eine Gruppe der einen Wahl der Statistik. Gewählt wird
// deshalb dort, und die Zusagen sind dieselben wie vorher.
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { Statistik } from './Statistik'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { einsetzen } from '../sprache/einsetzen'
import { belegeApi, raeumeApi } from '../test/api'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de.seiten.statistik

// zwei sind zwei Dimensionen, wie ein Spielpaket sie erklärt.
const zwei = [
  { name: 'color category', beschreibung: 'je Rahmenfarbe', numerisch: false },
  { name: 'mana value', beschreibung: 'je Manawert', numerisch: true },
]

// kennzahlen sind die Zahlen eines Sets mit einem Feld.
const kennzahlen = () =>
  Promise.resolve({
    statistik: {
      karten: 12,
      felder: 1,
      leere: 0,
      belegung: [{ name: 'name', karten: 12 }],
    },
  })

function zeige(zusatz = {}) {
  belegeApi({
    Statistik: kennzahlen,
    Auszaehldimensionen: () =>
      Promise.resolve({ dimensionen: zwei, fehler: null }),
    ...zusatz,
  })

  render(
    <Sprachraum sprache="de">
      <Statistik setID="1" />
    </Sprachraum>,
  )
}

// wahl ist die Wahl „Worüber", sobald die Dimensionen darin stehen.
async function wahl() {
  await screen.findByRole('group', { name: texte.gruppeSpielpaket })

  return screen.getByRole('combobox', { name: texte.worueber })
}

// **Die Dimensionen stehen in der Wahl, als eigene Gruppe.**
it('zeigt die Dimensionen des Spielpakets', async () => {
  zeige()

  const gruppe = await screen.findByRole('group', {
    name: texte.gruppeSpielpaket,
  })

  expect(
    within(gruppe).getByRole('option', { name: 'color category' }),
  ).toBeInTheDocument()
  expect(within(gruppe).getAllByRole('option')).toHaveLength(2)
})

// **Erst die Wahl rechnet.** Vorher wird nicht gefragt — eine Auszählung
// schickt jede Karte durch ein Skript.
it('zählt erst aus, wenn eine Dimension gewählt ist', async () => {
  const Auszaehlen = vi.fn(() =>
    Promise.resolve({
      dimension: 'color category',
      beschreibung: '',
      numerisch: false,
      gruppen: [
        { name: 'White', karten: 8, farbe: 'rgb(255,231,185)', kartenIDs: [] },
        { name: 'Land', karten: 4, farbe: '', kartenIDs: [] },
      ],
      befunde: [],
      fehler: null,
    }),
  )

  zeige({ Auszaehlen })

  const feld = await wahl()

  // **Vor der Wahl kein einziger Ruf.** Das ist die Zusage, und sie steht
  // hier als Zahl: ein Ruf über 3.761 Karten kostete gemessen 25 Sekunden.
  expect(Auszaehlen).not.toHaveBeenCalled()

  await userEvent.selectOptions(feld, 'dimension:color category')

  await waitFor(() => {
    expect(Auszaehlen).toHaveBeenCalledWith('1', 'color category')
  })

  // **Der Anteil bezieht sich auf die gezählten Karten**, nicht auf den
  // größten Balken: 8 von 12 sind 67 %.
  expect(await screen.findByText('8 (67 %)')).toBeInTheDocument()
  expect(screen.getByText('White')).toBeInTheDocument()

  // **Und das Bild sagt, wonach es zählt** — und warum es dauert.
  expect(
    screen.getByRole('heading', {
      name: einsetzen(texte.nachDimension, { dimension: 'color category' }),
    }),
  ).toBeInTheDocument()
})

// **Eine Gruppe ohne Karten steht trotzdem da** — das ist MSEs erklärte
// Reihenfolge, und eine Farbe ohne Karten ist eine Aussage über das Set.
it('zeigt auch eine Gruppe mit null Karten', async () => {
  zeige({
    Auszaehlen: () =>
      Promise.resolve({
        dimension: 'color category',
        beschreibung: '',
        numerisch: false,
        gruppen: [{ name: 'Land', karten: 0, farbe: '', kartenIDs: [] }],
        befunde: [],
        fehler: null,
      }),
  })

  await userEvent.selectOptions(await wahl(), 'dimension:color category')

  // **Und „0 (0 %)" und nicht „NaN %".** Das Ganze ist hier null, und eine
  // Division dadurch wäre keine Zahl.
  expect(await screen.findByText('0 (0 %)')).toBeInTheDocument()
})

it('erklärt eine Dimension, nach der keine Karte einen Wert trägt', async () => {
  zeige({
    Auszaehlen: () =>
      Promise.resolve({
        dimension: 'color category',
        beschreibung: '',
        numerisch: false,
        gruppen: [],
        befunde: [],
        fehler: null,
      }),
  })

  await userEvent.selectOptions(await wahl(), 'dimension:color category')

  expect(await screen.findByText(texte.dimensionLeer)).toBeInTheDocument()
})

// **Ein Befund steht am Bild und hält die Seite nicht an.**
it('zeigt die Befunde eines Laufs', async () => {
  zeige({
    Auszaehlen: () =>
      Promise.resolve({
        dimension: 'color category',
        beschreibung: '',
        numerisch: false,
        gruppen: [{ name: 'White', karten: 1, farbe: '', kartenIDs: [] }],
        befunde: ['die Dimension ließ sich für 3 Karten nicht rechnen'],
        fehler: null,
      }),
  })

  await userEvent.selectOptions(await wahl(), 'dimension:color category')

  expect(
    await screen.findByText(/für 3 Karten nicht rechnen/),
  ).toBeInTheDocument()
})

// **Ein Fehler des Laufs steht da, und die Wahl bleibt.**
it('zeigt einen Fehler des Laufs', async () => {
  zeige({
    Auszaehlen: () =>
      Promise.resolve({
        dimension: '',
        beschreibung: '',
        numerisch: false,
        gruppen: [],
        befunde: [],
        fehler: {
          code: 'nichtGefunden' as const,
          schluessel: '',
          meldung: 'dieses Spiel erklärt keine Dimension "color category"',
        },
      }),
  })

  await userEvent.selectOptions(await wahl(), 'dimension:color category')

  expect(await screen.findByText(/erklärt keine Dimension/)).toBeInTheDocument()
  expect(
    screen.getByRole('combobox', { name: texte.worueber }),
  ).toBeInTheDocument()
})

// **Ohne Dimensionen steht die Gruppe nicht da.**
//
// §6: leere Zustände erklären den nächsten Schritt — und für ein Set ohne
// Spielpaket gibt es in dieser Gruppe keinen.
it('zeigt keine Gruppe, wenn das Spielpaket keine Dimension erklärt', async () => {
  zeige({
    Auszaehldimensionen: () =>
      Promise.resolve({ dimensionen: [], fehler: null }),
  })

  await screen.findByRole('group', { name: texte.gruppeFelder })

  expect(
    screen.queryByRole('group', { name: texte.gruppeSpielpaket }),
  ).not.toBeInTheDocument()
})

// **Ein Wechsel des Sets räumt die Wahl mit** (§7, Regel 1).
//
// Ohne das fragte die Seite unmittelbar nach einer Dimension, die das neue
// Spiel gar nicht erklärt — und die leere Antwort sähe aus, als trüge dort
// keine Karte einen Wert.
it('räumt die gewählte Dimension, wenn das Set wechselt', async () => {
  const Auszaehlen = vi.fn(() =>
    Promise.resolve({
      dimension: 'color category',
      beschreibung: '',
      numerisch: false,
      gruppen: [{ name: 'White', karten: 1, farbe: '', kartenIDs: [] }],
      befunde: [],
      fehler: null,
    }),
  )

  belegeApi({
    Statistik: kennzahlen,
    Auszaehldimensionen: () =>
      Promise.resolve({ dimensionen: zwei, fehler: null }),
    Auszaehlen,
  })

  const { rerender } = render(
    <Sprachraum sprache="de">
      <Statistik setID="1" />
    </Sprachraum>,
  )

  await userEvent.selectOptions(await wahl(), 'dimension:color category')

  await waitFor(() => {
    expect(Auszaehlen).toHaveBeenCalledWith('1', 'color category')
  })

  rerender(
    <Sprachraum sprache="de">
      <Statistik setID="2" />
    </Sprachraum>,
  )

  // **Die Wahl steht wieder auf der Belegung** — und es wird nicht für Set 2
  // nach derselben Dimension gefragt.
  await waitFor(() => {
    expect(screen.getByRole('combobox', { name: texte.worueber })).toHaveValue(
      'belegung',
    )
  })

  expect(Auszaehlen).not.toHaveBeenCalledWith('2', 'color category')
})
