// **Die Ränder von Protokoll und abgeschlossener Werkbank** — DOM-04.AK-04
// (#1699):
// ein Protokoll, das sich nicht laden lässt, ein Vorgang, den dieses Fenster
// nicht kennt, ein Bye, ein Name, den es nicht mehr gibt — und eine Paarung
// ohne Stand in einem abgeschlossenen Turnier, die es nur in einem
// Bestand geben kann, den jemand an der Leitung vorbei gefüllt hat.
import { screen, within } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'

import { aufgeschlagen, zumTurniergebiet } from './turnierwerkbank'
import { raeumeApi } from '../test/api'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'
import {
  paarungsprobe,
  rundenprobe,
  teilnehmerprobe,
  turnierprobe,
} from './turnierprobe'

const texte = kataloge.de.seiten.turniere

afterEach(() => {
  raeumeApi()
})

// mitBye ist ein Turnier mit einem Bye und einer Paarung gegen jemanden, den
// die Liste der Teilnehmenden nicht mehr führt.
const mitBye = turnierprobe({
  teilnehmer: [teilnehmerprobe({ id: '3', name: 'Anna' })],
  runden: [
    rundenprobe({
      paarungen: [
        paarungsprobe({ id: '13', a: '3', b: '' }),
        paarungsprobe({ id: '14', a: '8', b: '3' }),
      ],
    }),
  ],
})

describe('das Protokoll an seinen Rändern', () => {
  it('nennt einen Bye, einen verschwundenen Namen und einen fremden Vorgang', async () => {
    await aufgeschlagen(mitBye, {
      TurnierProtokoll: () =>
        Promise.resolve({
          eintraege: [
            {
              zeit: '2026-09-27T11:00:00Z',
              vorgang: 'ergebnis',
              angaben: '2:0:0',
              paarungID: '13',
            },
            {
              zeit: '2026-09-27T11:01:00Z',
              vorgang: 'ergebnis',
              angaben: '0:2:0',
              paarungID: '14',
            },
            {
              zeit: '2026-09-27T11:02:00Z',
              vorgang: 'zukunft',
              angaben: '',
              paarungID: '',
            },
          ],
          fehler: null,
        }),
    })

    await zumTurniergebiet('protokoll')

    const liste = await screen.findByRole('list', {
      name: texte.gebietProtokoll,
    })
    expect(
      within(liste).getByText(
        einsetzen(texte.protokollErgebnis, {
          paarung: `Anna ${texte.gegen} ${texte.bye}`,
          angaben: '2:0:0',
        }),
      ),
    ).toBeInTheDocument()
    expect(
      within(liste).getByText(
        einsetzen(texte.protokollErgebnis, {
          paarung: `8 ${texte.gegen} Anna`,
          angaben: '0:2:0',
        }),
      ),
    ).toBeInTheDocument()

    // **Ein Vorgang, den dieses Fenster nicht kennt, steht beim Namen** —
    // eine neuere Fassung darf mehr protokollieren, und nichts verschwindet.
    expect(within(liste).getByText('zukunft')).toBeInTheDocument()
  })

  it('sagt, wenn es sich nicht laden lässt', async () => {
    await aufgeschlagen(mitBye, {
      TurnierProtokoll: () =>
        Promise.resolve({
          eintraege: [],
          fehler: {
            code: 'unbekannt' as const,
            schluessel: 'unbekannt',
            meldung: 'Die Platte klemmt.',
          },
        }),
    })

    await zumTurniergebiet('protokoll')

    expect(await screen.findByRole('alert')).toBeInTheDocument()
  })
})

describe('eine abgeschlossene Werkbank', () => {
  it('zeigt eine Paarung ohne Stand als offen und bietet kein Feld an', async () => {
    const werkbank = await aufgeschlagen(
      turnierprobe({
        abgeschlossen: '2026-09-27T12:00:00Z',
        teilnehmer: [
          teilnehmerprobe({ id: '3', name: 'Anna' }),
          teilnehmerprobe({ id: '4', name: 'Bert', stelle: 2 }),
        ],
        runden: [rundenprobe()],
      }),
    )

    await zumTurniergebiet('runden')

    expect(within(werkbank).getByText(texte.offen)).toBeInTheDocument()
    expect(within(werkbank).queryByRole('spinbutton')).toBeNull()
  })
})
