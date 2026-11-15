// **»Zum Tausch« beginnt einen neuen Entwurf** — #2391.
//
// Gefunden bei der Ausgangsmessung P0 (#2336, H2): war mit demselben
// Gegenüber schon ein angebotener Handel aufgeschlagen, endete Sammlung →
// Exemplar → »Zum Tausch« → »Neuer Tausch …« → das Gegenüber in diesem
// Handel. Das vorbelegte Formular erschien nicht, und das Exemplar wartete
// unsichtbar. Der Knopf in der Schublade hieß dabei »Aufgeschlagen« — er
// sagte, was schon war, und nicht, was er tut.
//
// **Entschieden**: wer in der Schublade ein Gegenüber wählt, beginnt einen
// neuen Entwurf mit ihm. Ein offener Handel mit ihm bleibt in der Liste und
// unberührt — ein angebotener nimmt ohnehin keine Zeilen, nur ein
// Gegenangebot.
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { afterEach, expect, it } from 'vitest'

import { Anschlussraum, useAnschluss } from '../anschlussraum'
import { Handeln } from './Handeln'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { eineListe } from './peerkontaktprobe'
import {
  bestaetigt,
  keinAnhalt,
  oeffneHandel,
  waehleGegenueber,
} from './tauschprobe'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'

import type { Seitenschluessel } from '../navigation'

const eigene = kataloge.de.seiten.handeln

const meineKennung = '12D3KooWIch'

function Buehne() {
  const [aktuell, setzeAktuell] = useState<Seitenschluessel>('handeln')

  return (
    <Sprachraum sprache="de">
      <Anschlussraum aktuell={aktuell} wechseln={setzeAktuell}>
        <Abgeber />
        <Handeln />
      </Anschlussraum>
    </Sprachraum>
  )
}

function Abgeber() {
  const anschluss = useAnschluss()

  return (
    <button
      type="button"
      onClick={() => {
        anschluss.uebergib({
          art: 'handel',
          name: 'Glutschmiede',
          printing: 'p-9',
          guete: 'nearMint',
          sprache: 'de',
          finish: 'nonfoil',
        })
      }}
    >
      abgeben
    </button>
  )
}

afterEach(() => {
  raeumeApi()
})

it('beginnt einen neuen Entwurf, auch wenn mit dem Gegenüber ein Handel aufgeschlagen ist', async () => {
  belegeApi({
    ...keinAnhalt,
    Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
    Peeridentitaet: () =>
      Promise.resolve({
        stand: 'bereit',
        kennung: meineKennung,
        fingerabdruck: 'AB-CD',
        seit: '2026-09-19T08:00:00Z',
        vorher: '',
        vorherFingerabdruck: '',
        befunde: [],
        fehler: null,
      }),
    HandelVorgaenge: () =>
      Promise.resolve({
        vorgaenge: [
          {
            vorgang: 'v-1',
            revision: 1,
            eltern: '',
            abdruck: 'abcdef0123456789',
            art: 'tausch',
            initiator: meineKennung,
            parteien: [meineKennung, 'bgegenueber'],
            ich: meineKennung,
            zustand: 'angeboten',
            positionen: [],
            bewertungGeteilt: false,
            zustimmungen: [],
            ende: false,
            beschluss: '',
            lage: 'warten',
            naechster: '',
            moegliche: ['abbrechen'],
            phase: 'angebot',
          },
        ],
        fehler: null,
      }),
  })

  render(<Buehne />)

  const hand = userEvent.setup()

  // **Das Gegenüber gewählt und darin den angebotenen Handel aufgeschlagen**
  // — so stand es, nachdem ein Entwurf angeboten worden war.
  await waehleGegenueber(hand, 'Fremder Rechner')
  await oeffneHandel(hand, 'v-1')

  // **Ein Exemplar »Zum Tausch«** und dasselbe Gegenüber noch einmal.
  await hand.click(screen.getByRole('button', { name: 'abgeben' }))
  await waehleGegenueber(hand, 'Fremder Rechner')

  expect(
    await screen.findByText(
      einsetzen(eigene.anschlussWartet, { name: 'Glutschmiede' }),
    ),
  ).toBeInTheDocument()
})

// **»Neuer Tausch mit …« beginnt je Klick einen neuen Entwurf** — P12
// (#2380, H1): derselbe Weg wie die Schublade, nur ohne sie. Was am Tisch
// lag, bleibt beim zweiten Klick nicht liegen.
it('beginnt mit »Neuer Tausch mit …« je Klick einen neuen Entwurf am Tisch', async () => {
  belegeApi({
    ...keinAnhalt,
    Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
    Peeridentitaet: () =>
      Promise.resolve({
        stand: 'bereit',
        kennung: meineKennung,
        fingerabdruck: 'AB-CD',
        seit: '2026-09-19T08:00:00Z',
        vorher: '',
        vorherFingerabdruck: '',
        befunde: [],
        fehler: null,
      }),
    HandelVorgaenge: () => Promise.resolve({ vorgaenge: [], fehler: null }),
    HandelSichtAnfragen: () => Promise.resolve({ zeilen: [], fehler: null }),
    Sammlung: () =>
      Promise.resolve({
        exemplare: [
          {
            id: 7,
            fassung: 1,
            printing: 'p-7',
            druckID: '',
            vorlaeufig: false,
            menge: 1,
            zustand: 'nearMint',
            sprache: 'de',
            finish: 'nonfoil',
            erwerbsdatum: '',
            erwerbspreis: '',
            erwerbswaehrung: '',
            notiz: '',
            name: 'Glutschmiede',
            setcode: 'p12',
            nummer: '7',
            lagerort: '',
            wunsch: false,
            bezeichnung: 'Glutschmiede',
            herkunft: { art: '', quelle: '', quellProfil: 0, uebernommen: '' },
          },
        ],
        fehler: null,
      }),
  } as never)

  render(<Buehne />)

  const hand = userEvent.setup()
  const neuMit = einsetzen(eigene.markttisch.neuMit, {
    name: 'Fremder Rechner',
  })

  await hand.click(await screen.findByRole('button', { name: neuMit }))
  const mein = await screen.findByRole('region', {
    name: eigene.markttisch.meinFach,
  })
  ;(await within(mein).findByRole('button', { name: 'Glutschmiede' })).focus()
  await hand.keyboard(' ')
  await hand.keyboard('{Enter}')
  const gebe = screen.getByRole('region', { name: eigene.positionVonIch })
  expect(await within(gebe).findByText(/Glutschmiede/)).toBeVisible()

  await hand.click(screen.getByRole('button', { name: neuMit }))

  // **Ein neuer Entwurf**: der Tisch steht wieder, und leer.
  const wieder = await screen.findByRole('region', {
    name: eigene.positionVonIch,
  })
  await within(
    await screen.findByRole('region', { name: eigene.markttisch.meinFach }),
  ).findByRole('button', { name: 'Glutschmiede' })
  expect(within(wieder).queryByText(/Glutschmiede/)).toBeNull()
})
