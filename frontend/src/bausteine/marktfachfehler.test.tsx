// **Ein Fehler des eigenen Fachs steht da, mit dem Weg zurück** —
// Gesamtreview über 0.9.0, Befund 3 (#2382).
//
// `useLaden` gab bei einer abgewiesenen `Sammlung()` den Satz samt
// Vorgangskennung her, und der Markttisch las vom eigenen Fach nur die
// Daten: die blieben leer, und das Fach zeigte weiter, dass es lädt — ohne
// Grund und ohne Rückweg. Das Fach des Gegenübers hatte beides.
import { render, screen, within } from '@testing-library/react'
import nutzerereignis from '@testing-library/user-event'
import { afterEach, expect, it } from 'vitest'

import { Anschlussraum } from '../anschlussraum'
import { keineHerkunft } from '../apiherkunft'
import { Markttisch } from './Markttisch'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'
import { belegeApi, raeumeApi } from '../test/api'
import { einAbgleich } from '../test/tauschabgleich'

import type { ExemplarDTO } from '../api'

const tisch = kataloge.de.seiten.handeln.markttisch
const zustaende = kataloge.de.bausteine.zustaende

afterEach(() => {
  raeumeApi()
})

const exemplar: ExemplarDTO = {
  id: 7,
  fassung: 3,
  printing: 'bolt-m10',
  druckID: '',
  vorlaeufig: false,
  menge: 4,
  zustand: 'nearMint',
  sprache: 'de',
  finish: 'nonfoil',
  erwerbsdatum: '',
  erwerbspreis: '',
  erwerbswaehrung: '',
  notiz: '',
  name: 'Blitzschlag',
  setcode: 'm10',
  nummer: '146',
  lagerort: '',
  wunsch: false,
  bezeichnung: 'Blitzschlag',
  herkunft: keineHerkunft,
}

it('nennt den Fehler des eigenen Fachs und lädt es erneut', async () => {
  let rufe = 0
  belegeApi({
    Sammlung: () => {
      rufe += 1

      return Promise.resolve(
        rufe === 1
          ? {
              exemplare: [],
              fehler: {
                code: 'unbekannt',
                meldung: 'Die Sammlung ließ sich nicht lesen.',
                schluessel: '',
                feld: '',
              },
            }
          : { exemplare: [exemplar], fehler: null },
      )
    },
    HandelSichtAnfragen: () => Promise.resolve({ zeilen: [], fehler: null }),
    Tauschabgleich: () => Promise.resolve(einAbgleich([])),
    HandelPreisanhalt: () =>
      Promise.resolve({
        anhalt: { zeilen: [], seiten: [], differenzen: [], hauptwaehrung: '' },
        fehler: null,
      }),
  })
  render(
    <Sprachraum sprache="de">
      <Anschlussraum
        aktuell="handeln"
        wechseln={() => {
          /* hier wird nicht gewechselt */
        }}
      >
        <Markttisch
          ich="ich-1"
          gegenueber="du-1"
          name="Bert"
          positionen={[]}
          setzePositionen={() => {
            /* hier wird nichts gezogen */
          }}
        />
      </Anschlussraum>
    </Sprachraum>,
  )

  const fach = await screen.findByRole('region', { name: tisch.meinFach })
  expect(await within(fach).findByRole('alert')).toHaveTextContent(
    'Die Sammlung ließ sich nicht lesen.',
  )

  await nutzerereignis
    .setup()
    .click(within(fach).getByRole('button', { name: zustaende.wiederholen }))

  expect(await within(fach).findByText('Blitzschlag')).toBeInTheDocument()
  expect(within(fach).queryByRole('alert')).toBeNull()
  expect(rufe).toBe(2)
})
