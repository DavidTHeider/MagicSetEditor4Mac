// **Was über die Leitung kam, steht im offenen Handel** — gefunden mit zwei
// Fenstern (Hüllenprobe `nat04handel`, S12.NAT-04,
// #1672).
//
// Bert hatte zugestimmt, Anna auch, und ihre Zustimmung stand in seinem
// Bestand — sein Fenster sagte weiter „Wartet auf Anna" und bot „Vorbereiten"
// nicht an. Die Liste links lud nach, der offene Handel rechts nicht, und
// was er nach dem eigenen Schritt zuletzt gehört hatte, stand vor dem Neuen.
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'
import { mitStandnachlauf } from '../test/standnachlauf'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { Handeln } from './Handeln'
import { eineListe } from './peerkontaktprobe'
import { einVorgang } from './handelslisteprobe'
import { bestaetigt, waehleGegenueber, zuPhase } from './tauschprobe'

import type { Handelsvorgang } from '../apihandelsleitung'

const eigene = kataloge.de.seiten.handeln

afterEach(() => {
  raeumeApi()
})

function seite(bestandsstand: number) {
  return (
    <Sprachraum sprache="de">
      <Handeln bestandsstand={bestandsstand} />
    </Sprachraum>
  )
}

// mitBestand stellt einen Handel bereit, dessen Stand der Lauf auswechselt —
// wie drüben die Leitung.
function mitBestand(anfang: Handelsvorgang, mehr = {}) {
  const bestand = { vorgang: anfang }

  belegeApi(
    mitStandnachlauf({
      Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
      HandelVorgaenge: () =>
        Promise.resolve({ vorgaenge: [bestand.vorgang], fehler: null }),
      // **Ein Zug bringt hinüber, was er ging** (P13, Vermerk 0200).
      HandelSchicken: () =>
        Promise.resolve({ nachrichtId: 'n-1', fehler: null }),
      HandelZustellen: () =>
        Promise.resolve({
          gegenueber: 'bgegenueber',
          getragen: 1,
          fehler: null,
        }),
      ...mehr,
    }),
  )

  return bestand
}

// oeffne schlägt den Handel auf, wie es ein Mensch täte.
async function oeffne(hand: ReturnType<typeof userEvent.setup>) {
  await waehleGegenueber(hand, 'Fremder Rechner')
  await zuPhase(hand, 'gegenueber')
  await hand.click(await screen.findByRole('radio', { name: /Revision 1/ }))
}

// vereinbart ist der Stand, nachdem beide zugestimmt haben.
const vereinbart = einVorgang({
  zustand: 'vereinbart',
  zustimmungen: ['bich', 'bgegenueber'],
  lage: 'dran',
  naechster: 'vorbereiten',
  moegliche: ['vorbereiten', 'abbrechen'],
})

it('zeigt den Schritt, der über die Leitung kam, ohne neu zu öffnen', async () => {
  const bestand = mitBestand(
    einVorgang({
      zustimmungen: ['bich'],
      lage: 'wartet',
      naechster: '',
      moegliche: ['abbrechen'],
    }),
  )
  const hand = userEvent.setup()
  const { rerender } = render(seite(0))

  await oeffne(hand)
  await zuPhase(hand, 'angebot')
  const schritte = await screen.findByRole('group', {
    name: eigene.assistent.schritte,
  })
  expect(
    within(schritte).queryByRole('button', {
      name: eigene.zug.vorbereitenUebergeben,
    }),
  ).toBeNull()

  bestand.vorgang = vereinbart
  rerender(seite(1))

  expect(
    await screen.findByRole('button', {
      name: eigene.zug.vorbereitenUebergeben,
    }),
  ).toBeVisible()
})

it('lässt den eigenen letzten Schritt nicht vor dem stehen, was danach kam', async () => {
  const gewartet = einVorgang({
    zustimmungen: ['bich'],
    lage: 'wartet',
    naechster: '',
    moegliche: ['abbrechen'],
  })
  const bestand = mitBestand(einVorgang(), {
    HandelSchritt: () => {
      bestand.vorgang = gewartet

      return Promise.resolve({ vorgang: gewartet, fehler: null })
    },
  })
  const hand = userEvent.setup()
  const { rerender } = render(seite(0))

  await oeffne(hand)
  await zuPhase(hand, 'angebot')
  await hand.click(
    await screen.findByRole('button', { name: eigene.schrittZustimmen }),
  )

  // **Der eigene Schritt ist angekommen** — erst dann zählt, was folgt.
  await waitFor(() => {
    expect(
      screen.queryByRole('button', { name: eigene.schrittZustimmen }),
    ).toBeNull()
  })

  bestand.vorgang = vereinbart
  rerender(seite(1))

  expect(
    await screen.findByRole('button', {
      name: eigene.zug.vorbereitenUebergeben,
    }),
  ).toBeVisible()
})
