// **Abschluss und Protokoll in der Werkbank** — DOM-04.AK-04
// (#1699),
// Vermerk 0139: ein abgeschlossenes Turnier zeigt, dass es feststeht, bietet
// nichts zum Ändern an und lässt sich mit Rückfrage wiedereröffnen; das
// Protokoll sagt in Sätzen, was geschah.
import { screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

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
import type { Api } from '../api'
import type { TurnierDTO } from '../apiturniere'

const texte = kataloge.de.seiten.turniere

afterEach(() => {
  raeumeApi()
})

// fertig ist ein Turnier mit einer gespielten Runde.
function fertig(teil: Partial<TurnierDTO> = {}): TurnierDTO {
  return turnierprobe({
    teilnehmer: [
      teilnehmerprobe({ id: '3', name: 'Anna' }),
      teilnehmerprobe({ id: '4', name: 'Bert', stelle: 2 }),
    ],
    runden: [
      rundenprobe({
        paarungen: [
          paarungsprobe({
            id: '12',
            ergebnis: { spieleA: 2, spieleB: 0, unentschieden: 0 },
          }),
        ],
      }),
    ],
    ...teil,
  })
}

const zu = fertig({ abgeschlossen: '2026-09-27T12:00:00Z' })

describe('der Abschluss eines Turniers', () => {
  it('schließt ein offenes Turnier gegen seine Fassung ab', async () => {
    const abschliessen = vi.fn<Api['TurnierAbschliessen']>(() =>
      Promise.resolve({ fehler: null }),
    )
    const werkbank = await aufgeschlagen(fertig(), {
      TurnierAbschliessen: abschliessen,
    })

    await userEvent.click(
      within(werkbank).getByRole('button', { name: texte.abschliessen }),
    )

    await waitFor(() => {
      expect(abschliessen).toHaveBeenCalledWith('7', '1')
    })
  })

  it('meldet, warum sich ein Turnier nicht abschließen lässt', async () => {
    const werkbank = await aufgeschlagen(fertig(), {
      TurnierAbschliessen: () =>
        Promise.resolve({
          fehler: {
            code: 'validierung' as const,
            schluessel: 'turnierNichtFertig',
            meldung: 'In Runde 1 hat eine Paarung noch kein Ergebnis.',
          },
        }),
    })

    await userEvent.click(
      within(werkbank).getByRole('button', { name: texte.abschliessen }),
    )

    expect(
      await screen.findByText(kataloge.de.fehler.turnierNichtFertig),
    ).toBeInTheDocument()
  })

  it('zeigt ein abgeschlossenes Turnier ohne etwas zum Ändern', async () => {
    const werkbank = await aufgeschlagen(zu)

    expect(
      within(werkbank).getByText(
        new RegExp(texte.abgeschlossenAm.split('{datum}')[0] ?? ''),
      ),
    ).toBeInTheDocument()
    expect(
      within(werkbank).queryByRole('button', { name: texte.abschliessen }),
    ).toBeNull()

    // **Keine Anmeldung, keine Runde, kein Eintragen.**
    expect(
      within(werkbank).queryByRole('button', { name: texte.anmelden }),
    ).toBeNull()

    await zumTurniergebiet('runden')
    expect(
      within(werkbank).queryByRole('button', { name: texte.rundeLegen }),
    ).toBeNull()
    expect(
      within(werkbank).queryByRole('button', {
        name: `${texte.eintragen}: Anna ${texte.gegen} Bert`,
      }),
    ).toBeNull()
  })

  it('eröffnet erst nach der Rückfrage wieder', async () => {
    const wiedereroeffnen = vi.fn<Api['TurnierWiedereroeffnen']>(() =>
      Promise.resolve({ fehler: null }),
    )
    const werkbank = await aufgeschlagen(zu, {
      TurnierWiedereroeffnen: wiedereroeffnen,
    })

    await userEvent.click(
      within(werkbank).getByRole('button', { name: texte.wiedereroeffnen }),
    )

    const frage = within(werkbank).getByRole('group', {
      name: texte.wiedereroeffnenFrage,
    })
    expect(wiedereroeffnen).not.toHaveBeenCalled()

    await userEvent.click(
      within(frage).getByRole('button', { name: texte.wiedereroeffnen }),
    )

    await waitFor(() => {
      expect(wiedereroeffnen).toHaveBeenCalledWith('7', '1')
    })
  })

  it('lässt die Rückfrage ohne Folgen fallen', async () => {
    const wiedereroeffnen = vi.fn<Api['TurnierWiedereroeffnen']>()
    const werkbank = await aufgeschlagen(zu, {
      TurnierWiedereroeffnen: wiedereroeffnen,
    })

    await userEvent.click(
      within(werkbank).getByRole('button', { name: texte.wiedereroeffnen }),
    )
    await userEvent.click(
      within(werkbank).getByRole('button', {
        name: texte.nichtWiedereroeffnen,
      }),
    )

    expect(
      within(werkbank).queryByRole('group', {
        name: texte.wiedereroeffnenFrage,
      }),
    ).toBeNull()
    expect(wiedereroeffnen).not.toHaveBeenCalled()
  })
})

describe('das Protokoll', () => {
  it('sagt in Sätzen, was geschah', async () => {
    await aufgeschlagen(fertig(), {
      TurnierProtokoll: () =>
        Promise.resolve({
          eintraege: [
            {
              zeit: '2026-09-27T11:00:00Z',
              vorgang: 'angelegt',
              angaben: 'Küchentisch',
              paarungID: '',
            },
            {
              zeit: '2026-09-27T11:05:00Z',
              vorgang: 'ergebnis',
              angaben: '2:0:0',
              paarungID: '12',
            },
            {
              zeit: '2026-09-27T11:06:00Z',
              vorgang: 'zurueckgenommen',
              angaben: '',
              paarungID: '99',
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
        einsetzen(texte.protokollAngelegt, { angaben: 'Küchentisch' }),
      ),
    ).toBeInTheDocument()
    expect(
      within(liste).getByText(
        einsetzen(texte.protokollErgebnis, {
          paarung: `Anna ${texte.gegen} Bert`,
          angaben: '2:0:0',
        }),
      ),
    ).toBeInTheDocument()

    // **Eine Paarung, die es nicht mehr gibt, heißt nach ihrer Kennung.**
    expect(
      within(liste).getByText(
        einsetzen(texte.protokollZurueckgenommen, { paarung: '99' }),
      ),
    ).toBeInTheDocument()
  })

  it('sagt, wenn noch nichts geschah', async () => {
    await aufgeschlagen(fertig(), {
      TurnierProtokoll: () => Promise.resolve({ eintraege: [], fehler: null }),
    })

    await zumTurniergebiet('protokoll')

    expect(await screen.findByText(texte.protokollLeer)).toBeInTheDocument()
  })
})
