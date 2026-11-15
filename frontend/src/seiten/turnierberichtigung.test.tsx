// **Eine Berichtigung sagt vorher, was sie verschiebt** — DOM-04.AK-02
// (#1699):
// *„Ergebniskorrekturen zeigen Auswirkungen auf folgende Runden und
// Ranglisten."*
//
// **Nur eine Berichtigung fragt**: das erste Ergebnis der laufenden Runde
// geht wie bisher sofort hinein.
import { screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { aufgeschlagen, ergebnisOeffnen } from './turnierwerkbank'
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
import type { KorrekturfolgenDTO } from '../apiturniere'

const texte = kataloge.de.seiten.turniere

afterEach(() => {
  raeumeApi()
})

const sieg = { spieleA: 2, spieleB: 0, unentschieden: 0 }

// zweiRunden ist ein Turnier mit Anna gegen Bert in Runde 1 (gespielt) und
// Anna gegen Cleo in Runde 2 (offen).
const zweiRunden = turnierprobe({
  teilnehmer: [
    teilnehmerprobe({ id: '3', name: 'Anna' }),
    teilnehmerprobe({ id: '4', name: 'Bert', stelle: 2 }),
    teilnehmerprobe({ id: '5', name: 'Cleo', stelle: 3 }),
  ],
  runden: [
    rundenprobe({
      id: '9',
      nummer: 1,
      paarungen: [paarungsprobe({ id: '12', ergebnis: sieg })],
    }),
    rundenprobe({
      id: '10',
      nummer: 2,
      paarungen: [paarungsprobe({ id: '13', a: '3', b: '5' })],
    }),
  ],
})

const folgen: KorrekturfolgenDTO = {
  runde: 1,
  spaetereRunden: [2],
  verschoben: [
    {
      teilnehmerID: '4',
      name: 'Bert',
      platzVorher: 2,
      platzNachher: 1,
      punkteVorher: 0,
      punkteNachher: 3,
    },
  ],
}

// eintragen klappt das Eintragen der Paarung auf und drückt *Eintragen* —
// seit UXR-12 (#1920) steht es hinter *Ergebnis*, und Runde 1 ist
// eingeklappt, sobald es eine zweite gibt.
async function eintragen(
  werkbank: HTMLElement,
  paarung: string,
  runde?: number,
) {
  await ergebnisOeffnen(werkbank, paarung, runde)
  await klickeEintragen(werkbank, paarung)
}

async function klickeEintragen(werkbank: HTMLElement, paarung: string) {
  await userEvent.click(
    within(werkbank).getByRole('button', {
      name: `${texte.eintragen}: ${paarung}`,
    }),
  )
}

describe('eine Berichtigung', () => {
  it('nennt ihre Folgen und schreibt erst nach dem Ja', async () => {
    const setzen = vi.fn<Api['TurnierErgebnisSetzen']>(() =>
      Promise.resolve({ fehler: null }),
    )
    const fragen = vi.fn<Api['TurnierKorrekturfolgen']>(() =>
      Promise.resolve({ folgen, fehler: null }),
    )
    const werkbank = await aufgeschlagen(
      zweiRunden,
      { TurnierErgebnisSetzen: setzen, TurnierKorrekturfolgen: fragen },
      'runden',
    )

    await eintragen(werkbank, `Anna ${texte.gegen} Bert`, 1)

    const frage = await within(werkbank).findByRole('group', {
      name: texte.folgenFrage,
    })
    expect(fragen).toHaveBeenCalledWith('7', '12', sieg)
    expect(setzen).not.toHaveBeenCalled()
    expect(
      within(frage).getByText(
        einsetzen(texte.folgenPlatz, {
          name: 'Bert',
          vorher: '2',
          nachher: '1',
          punkteVorher: '0',
          punkteNachher: '3',
        }),
      ),
    ).toBeInTheDocument()
    expect(
      within(frage).getByText(einsetzen(texte.folgenSpaeter, { runden: '2' })),
    ).toBeInTheDocument()

    await userEvent.click(
      within(frage).getByRole('button', { name: texte.folgenEintragen }),
    )

    await waitFor(() => {
      expect(setzen).toHaveBeenCalledWith('7', '1', '12', sieg)
    })
  })

  it('schreibt nichts, wenn niemand bestätigt', async () => {
    const setzen = vi.fn<Api['TurnierErgebnisSetzen']>(() =>
      Promise.resolve({ fehler: null }),
    )
    const werkbank = await aufgeschlagen(
      zweiRunden,
      {
        TurnierErgebnisSetzen: setzen,
        TurnierKorrekturfolgen: () => Promise.resolve({ folgen, fehler: null }),
      },
      'runden',
    )

    await eintragen(werkbank, `Anna ${texte.gegen} Bert`, 1)
    await userEvent.click(
      await within(werkbank).findByRole('button', { name: texte.folgenNicht }),
    )

    expect(
      within(werkbank).queryByRole('group', { name: texte.folgenFrage }),
    ).toBeNull()
    expect(setzen).not.toHaveBeenCalled()
  })

  // **Auch das Zurücknehmen ist eine Berichtigung.**
  it('fragt auch vor dem Zurücknehmen', async () => {
    const fragen = vi.fn<Api['TurnierKorrekturfolgen']>(() =>
      Promise.resolve({ folgen, fehler: null }),
    )
    const werkbank = await aufgeschlagen(
      zweiRunden,
      { TurnierKorrekturfolgen: fragen },
      'runden',
    )

    await ergebnisOeffnen(werkbank, `Anna ${texte.gegen} Bert`, 1)
    await userEvent.click(
      within(werkbank).getByRole('button', {
        name: `${texte.zuruecknehmen}: Anna ${texte.gegen} Bert`,
      }),
    )

    await waitFor(() => {
      expect(fragen).toHaveBeenCalledWith('7', '12', null)
    })
  })

  // **Verschiebt sie nichts, geht sie ohne Frage hinein.**
  it('schreibt sofort, wenn sich nichts verschiebt', async () => {
    const setzen = vi.fn<Api['TurnierErgebnisSetzen']>(() =>
      Promise.resolve({ fehler: null }),
    )
    const werkbank = await aufgeschlagen(
      zweiRunden,
      {
        TurnierErgebnisSetzen: setzen,
        TurnierKorrekturfolgen: () =>
          Promise.resolve({
            folgen: { runde: 1, spaetereRunden: [], verschoben: [] },
            fehler: null,
          }),
      },
      'runden',
    )

    await eintragen(werkbank, `Anna ${texte.gegen} Bert`, 1)

    await waitFor(() => {
      expect(setzen).toHaveBeenCalled()
    })
  })

  // **Das erste Ergebnis der laufenden Runde fragt nicht.**
  it('fragt nicht beim ersten Ergebnis der laufenden Runde', async () => {
    const setzen = vi.fn<Api['TurnierErgebnisSetzen']>(() =>
      Promise.resolve({ fehler: null }),
    )
    const fragen = vi.fn<Api['TurnierKorrekturfolgen']>()
    const werkbank = await aufgeschlagen(
      zweiRunden,
      { TurnierErgebnisSetzen: setzen, TurnierKorrekturfolgen: fragen },
      'runden',
    )

    // **Erst ein Spielstand** — drei Nullen sind keiner.
    await ergebnisOeffnen(werkbank, `Anna ${texte.gegen} Cleo`)
    const feld = within(werkbank).getByLabelText(
      `${texte.spieleA} — Anna ${texte.gegen} Cleo`,
    )
    await userEvent.clear(feld)
    await userEvent.type(feld, '2')
    await klickeEintragen(werkbank, `Anna ${texte.gegen} Cleo`)

    await waitFor(() => {
      expect(setzen).toHaveBeenCalled()
    })
    expect(fragen).not.toHaveBeenCalled()
  })

  it('meldet, wenn die Frage scheitert, und schreibt nicht', async () => {
    const setzen = vi.fn<Api['TurnierErgebnisSetzen']>()
    const werkbank = await aufgeschlagen(
      zweiRunden,
      {
        TurnierErgebnisSetzen: setzen,
        TurnierKorrekturfolgen: () =>
          Promise.resolve({
            folgen: null,
            fehler: {
              code: 'unbekannt',
              schluessel: 'fehler.unbekannt',
              meldung: 'Der Bestand schweigt.',
            },
          }),
      },
      'runden',
    )

    await eintragen(werkbank, `Anna ${texte.gegen} Bert`, 1)

    // **Die Meldung steht über der Seite** — wie jede aus `lies`.
    expect(await screen.findByText('Der Bestand schweigt.')).toHaveAttribute(
      'role',
      'alert',
    )
    expect(setzen).not.toHaveBeenCalled()
  })
})
