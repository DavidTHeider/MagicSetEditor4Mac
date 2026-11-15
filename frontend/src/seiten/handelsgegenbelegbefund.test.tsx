// **Ein Gegenbeleg, der nicht passt, steht am Beleg** —
// #2500.
//
// **Bis dahin sagte das Fenster dauerhaft „Nur du hast unterschrieben"**, als
// wäre nichts gekommen. Jetzt nennt es, von wem einer kam, woran er
// scheiterte und was als Nächstes zu tun ist — und ein späterer, passender
// Beleg verdrängt die Meldung.
import { screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'
import { eineListe } from './peerkontaktprobe'
import { zeigeHandel } from './tauschprobe'
import {
  bestaetigt,
  dieVorgaenge,
  einJournal,
  einStand,
  eineFreigabe,
  imAbschluss,
  trageEin,
  mitBelegnachlauf,
} from './buchungsprobe'
import { kataloge } from '../sprache/kataloge'

import type { GegenbelegbefundDTO } from '../apihandelsverbuchung'

const eigene = kataloge.de.seiten.handeln

const einBefund: GegenbelegbefundDTO = {
  von: 'peer-bert',
  abdruck: 'abdruck-drueben',
  mangel: 'zeilen',
  grund: 'der Beleg trägt 1 Zeilen und meiner 2',
  empfangen: '2026-10-19T19:30:00Z',
}

// zeige stellt einen verbuchten Handel mit diesem Beleg hin.
async function zeige(beidseitig: boolean, befund: GegenbelegbefundDTO) {
  belegeApi(
    mitBelegnachlauf({
      Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
      HandelVorgaenge: dieVorgaenge(imAbschluss()),
      Handelsverbuchung: () =>
        Promise.resolve(
          einStand({
            zustand: beidseitig ? 'beidseitigVerbucht' : 'lokalVerbucht',
            beleg: {
              vorgang: 'v-1',
              abdruck: 'abdruck-1',
              unterschrift: 'deadbeef',
              beidseitig,
              gegenzeichner: beidseitig ? 'peer-bert' : '',
              gegengezeichnet: beidseitig ? '2026-10-19T20:00:00Z' : '',
              gebucht: '2026-10-19T18:00:00Z',
              befund,
            },
          }),
        ),
      HandelFreigabeErteilen: eineFreigabe,
      Handelsjournal: () => Promise.resolve(einJournal()),
    }),
  )

  await trageEin(await zeigeHandel(), eigene.buchungVerbuchen)
}

describe('Ein abgewiesener Gegenbeleg', () => {
  afterEach(() => {
    raeumeApi()
  })

  it('nennt Absender, Mangel und den nächsten Schritt', async () => {
    await zeige(false, einBefund)

    const satz = await screen.findByText(eigene.gegenbelegMaengel.zeilen, {
      exact: false,
    })

    expect(satz).toHaveTextContent('peer-bert')
    expect(satz).toHaveTextContent(
      eigene.buchungGegenbelegRat.replace('{wer}', 'peer-bert'),
    )
    expect(
      screen.queryByText(eigene.buchungOhneGegenbeleg),
    ).not.toBeInTheDocument()
  })

  // **Ein Mangel, den dieses Fenster nicht kennt, wird genannt** (§16):
  // eine neuere Gegenseite darf einen sechsten haben.
  it('nennt einen unbekannten Mangel roh', async () => {
    await zeige(false, { ...einBefund, mangel: 'farbe' })

    expect(
      await screen.findByText(/farbe/, { exact: false }),
    ).toBeInTheDocument()
  })

  it('weicht einem geprüften Beleg, der danach kam', async () => {
    await zeige(true, einBefund)

    expect(
      await screen.findByText(/Geprüfter Beleg von peer-bert/),
    ).toBeInTheDocument()
    expect(
      screen.queryByText(eigene.gegenbelegMaengel.zeilen, { exact: false }),
    ).not.toBeInTheDocument()
  })
})
