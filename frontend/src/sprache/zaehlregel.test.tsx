// **Zwei Sprachräume, zwei Zählregeln** — Befund N5 des Nachlaufs über
// Phase 6.5 (LOC-02, #1688): die Regel stand in einem Zustand des Moduls,
// den jeder Sprachraum beim Zeichnen setzte — der zuletzt gezeichnete galt
// für alle. Jetzt trägt jede Zählform ihre eigene.
import { render, screen } from '@testing-library/react'
import { expect, it } from 'vitest'

import { flach, ueberlagere } from './flach'
import { kataloge } from './kataloge'
import { Sprachraum, useTexte } from './kontext'
import { zaehlwort } from './zaehlwort'

function Zahl() {
  return <p>{zaehlwort(21, useTexte().allgemein.zaehlwoerter.karten)}</p>
}

const polnisch = ueberlagere(
  kataloge.ru,
  {
    'allgemein.zaehlwoerter.karten.eine': 'karta',
    'allgemein.zaehlwoerter.karten.wenige': 'karty',
    'allgemein.zaehlwoerter.karten.viele': 'kart',
  },
  'pl',
)

it('zählt jeden Sprachraum nach seiner Regel', () => {
  render(
    <>
      <Sprachraum sprache="ru" katalog={polnisch}>
        <Zahl />
      </Sprachraum>
      <Sprachraum sprache="ru">
        <Zahl />
      </Sprachraum>
    </>,
  )

  expect(screen.getByText('21 kart')).toBeInTheDocument()
  expect(screen.getByText('21 карта')).toBeInTheDocument()
})

it('hält die Regel aus der flachen Form heraus', () => {
  expect(flach(polnisch)).not.toHaveProperty(
    'allgemein.zaehlwoerter.karten.regel',
  )
  expect(Object.keys(flach(polnisch))).toEqual(Object.keys(flach(kataloge.ru)))
})
