// **Ein Mausgriff am Markttisch endet sauber** — Gesamtreview über 0.9.0,
// Befund 4 (#2382).
//
// Jeder linke `pointerdown` hängte einen einmaligen `pointerup`-Horcher an das
// Dokument: ohne `pointercancel`, ohne Zeigerkennung, ohne Abbau. Nach einem
// abgebrochenen Griff legte ein späteres Loslassen über einer Ablage die alte
// Karte. Und der Horcher hielt die Positionen vom Zugreifen fest — was sich
// während des Griffs änderte, überschrieb das Loslassen.
import { fireEvent, render, screen, within } from '@testing-library/react'
import { useState } from 'react'
import { afterEach, expect, it, vi } from 'vitest'

import { Anschlussraum } from '../anschlussraum'
import { keineHerkunft } from '../apiherkunft'
import { Markttisch } from './Markttisch'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'
import { belegeApi, raeumeApi } from '../test/api'
import { einAbgleich } from '../test/tauschabgleich'

import type { ExemplarDTO } from '../api'
import type { Handelsposition } from '../apihandelsleitung'

const texte = kataloge.de.seiten.handeln

afterEach(() => {
  raeumeApi()
})

const exemplar = (id: number, name: string): ExemplarDTO => ({
  id,
  fassung: 1,
  printing: `p-${String(id)}`,
  druckID: '',
  vorlaeufig: false,
  menge: 2,
  zustand: 'nearMint',
  sprache: 'de',
  finish: 'nonfoil',
  erwerbsdatum: '',
  erwerbspreis: '',
  erwerbswaehrung: '',
  notiz: '',
  name,
  setcode: 'p12',
  nummer: String(id),
  lagerort: '',
  wunsch: false,
  bezeichnung: name,
  herkunft: keineHerkunft,
})

// vonAussen ist eine Position, die während eines Griffs anderswo entsteht.
const vonAussen: Handelsposition = {
  von: 'ich-1',
  printing: 'p-anders',
  druckId: 0,
  menge: 1,
  guete: 'nearMint',
  sprache: 'de',
  finish: 'nonfoil',
  abbild: '',
  bewertung: 0,
}

function Buehne({
  gesetzt,
}: {
  gesetzt: (positionen: Handelsposition[]) => void
}) {
  const [positionen, setzePositionen] = useState<Handelsposition[]>([])
  const [da, setzeDa] = useState(true)

  return (
    <Sprachraum sprache="de">
      <Anschlussraum
        aktuell="handeln"
        wechseln={() => {
          /* hier wird nicht gewechselt */
        }}
      >
        <button
          type="button"
          onClick={() => {
            setzePositionen([vonAussen])
          }}
        >
          von außen
        </button>
        <button
          type="button"
          onClick={() => {
            setzeDa(false)
          }}
        >
          abbauen
        </button>
        {da && (
          <Markttisch
            ich="ich-1"
            gegenueber="du-1"
            name="Bert"
            positionen={positionen}
            setzePositionen={(neu) => {
              gesetzt(neu)
              setzePositionen(neu)
            }}
          />
        )}
        <div data-ablage="gebe">Ablage nach dem Abbau</div>
      </Anschlussraum>
    </Sprachraum>
  )
}

async function zeige() {
  belegeApi({
    Sammlung: () =>
      Promise.resolve({
        exemplare: [exemplar(1, 'Glutläufer')],
        fehler: null,
      }),
    HandelSichtAnfragen: () => Promise.resolve({ zeilen: [], fehler: null }),
    Tauschabgleich: () => Promise.resolve(einAbgleich()),
    HandelPreisanhalt: () =>
      Promise.resolve({
        anhalt: { zeilen: [], seiten: [], differenzen: [], hauptwaehrung: '' },
        fehler: null,
      }),
  })
  const gesetzt = vi.fn()
  render(<Buehne gesetzt={gesetzt} />)

  const fach = await screen.findByRole('region', {
    name: texte.markttisch.meinFach,
  })
  const glut = await within(fach).findByText('Glutläufer')
  const gebe = screen.getByRole('region', { name: texte.positionVonIch })

  return { gesetzt, glut, gebe }
}

it('legt nach einem abgebrochenen Griff beim späteren Loslassen nichts', async () => {
  const { gesetzt, glut, gebe } = await zeige()

  fireEvent.pointerDown(glut, { button: 0, pointerId: 1 })
  fireEvent.pointerCancel(glut, { pointerId: 1 })
  fireEvent.pointerUp(gebe, { pointerId: 1 })

  expect(gesetzt).not.toHaveBeenCalled()
})

it('lässt sich von einem fremden Zeiger nicht abbrechen', async () => {
  const { gesetzt, glut, gebe } = await zeige()

  fireEvent.pointerDown(glut, { button: 0, pointerId: 1 })
  fireEvent.pointerCancel(glut, { pointerId: 2 })
  fireEvent.pointerUp(gebe, { pointerId: 1 })

  expect(gesetzt).toHaveBeenLastCalledWith([
    expect.objectContaining({ von: 'ich-1', printing: 'p-1' }),
  ])
})

it('lässt sich von einem fremden Zeiger nicht loslassen', async () => {
  const { gesetzt, glut, gebe } = await zeige()

  fireEvent.pointerDown(glut, { button: 0, pointerId: 1 })
  fireEvent.pointerUp(gebe, { pointerId: 2 })
  expect(gesetzt).not.toHaveBeenCalled()

  fireEvent.pointerUp(gebe, { pointerId: 1 })
  expect(gesetzt).toHaveBeenLastCalledWith([
    expect.objectContaining({ von: 'ich-1', printing: 'p-1' }),
  ])
})

it('legt auf den Stand, der beim Loslassen gilt', async () => {
  const { gesetzt, glut, gebe } = await zeige()

  fireEvent.pointerDown(glut, { button: 0, pointerId: 1 })
  fireEvent.click(screen.getByRole('button', { name: 'von außen' }))
  fireEvent.pointerUp(gebe, { pointerId: 1 })

  expect(gesetzt).toHaveBeenLastCalledWith([
    vonAussen,
    expect.objectContaining({ printing: 'p-1' }),
  ])
})

it('legt nach dem Abbau des Tischs nichts mehr', async () => {
  const { gesetzt, glut } = await zeige()

  fireEvent.pointerDown(glut, { button: 0, pointerId: 1 })
  fireEvent.click(screen.getByRole('button', { name: 'abbauen' }))
  fireEvent.pointerUp(screen.getByText('Ablage nach dem Abbau'), {
    pointerId: 1,
  })

  expect(gesetzt).not.toHaveBeenCalled()
})
